package inttestutils

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"

	"github.com/buger/jsonparser"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/distribution/services/utils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

const (
	distributionGpgKeyCreatePattern = `{"public_key":"%s","private_key":"%s"}`
	ArtifactoryGpgKeyCreatePattern  = `{"alias":"cli tests distribution key","public_key":"%s"}`
)

type distributableDistributionStatus string
type receivedDistributionStatus string

const (
	open                  distributableDistributionStatus = "OPEN"
	signed                distributableDistributionStatus = "SIGNED"
	stored                distributableDistributionStatus = "STORED"
	readyForDistribution  distributableDistributionStatus = "READY_FOR_DISTRIBUTION"

	NotDistributed receivedDistributionStatus = "Not distributed"
	InProgress     receivedDistributionStatus = "In progress"
	Completed      receivedDistributionStatus = "Completed"
	Failed         receivedDistributionStatus = "Failed"
)

// GET api/v1/release_bundle/:name/:version
type distributableResponse struct {
	utils.ReleaseBundleBody
	Name    string                          `json:"name,omitempty"`
	Version string                          `json:"version,omitempty"`
	State   distributableDistributionStatus `json:"state,omitempty"`
}

// Get api/v1/release_bundle/:name/:version/distribution
type receivedResponse struct {
	Id     string                     `json:"id,omitempty"`
	Status receivedDistributionStatus `json:"status,omitempty"`
}

type ReceivedResponses struct {
	receivedResponses []receivedResponse
}

// Send GPG keys to Distribution and Artifactory to allow signing of release bundles
func SendGpgKeys(artHttpDetails httputils.HttpClientDetails, distHttpDetails httputils.HttpClientDetails) {
	keysDir := filepath.Join(tests.GetTestResourcesPath(), "distribution")
	publicKey, err := os.ReadFile(filepath.Join(keysDir, "public.key.1"))
	coreutils.ExitOnErr(err)
	privateKey, err := os.ReadFile(filepath.Join(keysDir, "private.key"))
	coreutils.ExitOnErr(err)

	client, err := httpclient.ClientBuilder().Build()
	coreutils.ExitOnErr(err)

	// PUT → Distribution
	content := fmt.Sprintf(distributionGpgKeyCreatePattern, publicKey, privateKey)
	resp, body, err := client.SendPut(*tests.JfrogUrl+"distribution/api/v1/keys/pgp", []byte(content), distHttpDetails, "")
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	coreutils.ExitOnErr(err)
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// POST → Artifactory
	content = fmt.Sprintf(ArtifactoryGpgKeyCreatePattern, publicKey)
	resp, body, err = client.SendPost(*tests.JfrogUrl+"artifactory/api/security/keys/trusted", []byte(content), artHttpDetails, "")
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	coreutils.ExitOnErr(err)
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusCreated, http.StatusConflict); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func GetLocalBundle(t *testing.T, bundleName, bundleVersion string, distHttpDetails httputils.HttpClientDetails) *distributableResponse {
	resp, body := getLocalBundle(t, bundleName, bundleVersion, distHttpDetails)
	if err := errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK); err != nil {
		t.Error(err.Error())
		return nil
	}
	response := &distributableResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Error(err)
		return nil
	}
	return response
}

func VerifyLocalBundleExistence(t *testing.T, bundleName, bundleVersion string, expectExist bool, distHttpDetails httputils.HttpClientDetails) {
	for i := 0; i < 120; i++ {
		resp, body := getLocalBundle(t, bundleName, bundleVersion, distHttpDetails)

		switch resp.StatusCode {
		case http.StatusOK:
			if expectExist { return }
		case http.StatusNotFound:
			if !expectExist { return }
		default:
			t.Error(resp.Status)
			t.Error(string(body))
			return
		}

		t.Log("Waiting for " + bundleName + "/" + bundleVersion + "...")
		time.Sleep(time.Second)
	}
	t.Errorf("Release bundle %s/%s exist: %v unlike expected", bundleName, bundleVersion, expectExist)
}

func AssertReleaseBundleOpen(t *testing.T, distributableResponse *distributableResponse) {
	assert.NotNil(t, distributableResponse)
	assert.Equal(t, open, distributableResponse.State)
}

func AssertReleaseBundleSigned(t *testing.T, distributableResponse *distributableResponse) {
	assert.NotNil(t, distributableResponse)
	assert.Contains(t, []distributableDistributionStatus{signed, stored, readyForDistribution}, distributableResponse.State)
}

func WaitForDistribution(t *testing.T, bundleName, bundleVersion string, distHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	for i := 0; i < 120; i++ {
		resp, body, _, err := client.SendGet(*tests.JfrogUrl+"distribution/api/v1/release_bundle/"+bundleName+"/"+bundleVersion+"/distribution", true, distHttpDetails, "")
		assert.NoError(t, err)
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK); err != nil {
			t.Error(err.Error())
			return
		}

		response := &ReceivedResponses{}
		if err = json.Unmarshal(body, &response.receivedResponses); err != nil {
			t.Error(err)
			return
		}
		if len(response.receivedResponses) == 0 {
			t.Error("Release bundle \"" + bundleName + "/" + bundleVersion + "\" not found")
			return
		}

		switch response.receivedResponses[0].Status {
		case Completed:
			return
		case Failed:
			t.Error("Distribution failed for " + bundleName + "/" + bundleVersion)
			return
		}

		t.Log("Waiting for " + bundleName + "/" + bundleVersion + "...")
		time.Sleep(time.Second)
	}
	t.Error("Timeout for release bundle distribution " + bundleName + "/" + bundleVersion)
}

func WaitForDeletion(t *testing.T, bundleName, bundleVersion string, distHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	for i := 0; i < 120; i++ {
		resp, body, _, err := client.SendGet(*tests.JfrogUrl+"distribution/api/v1/release_bundle/"+bundleName+"/"+bundleVersion+"/distribution", true, distHttpDetails, "")
		assert.NoError(t, err)
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		if resp.StatusCode == http.StatusNotFound {
			return
		}

		if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK); err != nil {
			t.Error(err.Error())
			return
		}

		t.Log("Waiting for distribution deletion " + bundleName + "/" + bundleVersion + "...")
		time.Sleep(time.Second)
	}
	t.Error("Timeout for release bundle deletion " + bundleName + "/" + bundleVersion)
}

func getLocalBundle(t *testing.T, bundleName, bundleVersion string, distHttpDetails httputils.HttpClientDetails) (*http.Response, []byte) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	resp, body, _, err := client.SendGet(*tests.JfrogUrl+"distribution/api/v1/release_bundle/"+bundleName+"/"+bundleVersion, true, distHttpDetails, "")
	assert.NoError(t, err)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return resp, body
}

func CleanUpOldBundles(distHttpDetails httputils.HttpClientDetails, bundleVersion string, distributionCli *coreTests.JfrogCli) {
	getActualItems := func() ([]string, error) { return ListAllBundlesNames(distHttpDetails) }
	deleteItem := func(bundleName string) {
		err := distributionCli.Exec("rbdel", bundleName, bundleVersion, "--site=*", "--delete-from-dist", "--quiet")
		if err != nil {
			log.Error(err)
		} else {
			log.Info("Bundle", bundleName, "deleted.")
		}
	}
	tests.CleanUpOldItems([]string{tests.BundleName}, getActualItems, deleteItem)
}

func ListAllBundlesNames(distHttpDetails httputils.HttpClientDetails) ([]string, error) {
	var bundlesNames []string

	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return nil, err
	}

	resp, body, _, err := client.SendGet(*tests.JfrogUrl+"distribution/api/v1/release_bundle/distribution", true, distHttpDetails, "")
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK, http.StatusCreated); err != nil {
		return nil, err
	}

	var keyError error
	_, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil || keyError != nil {
			return
		}
		bundleName, err := jsonparser.GetString(value, "release_bundle_name")
		if err != nil {
			keyError = err
			return
		}
		bundlesNames = append(bundlesNames, bundleName)
	})
	if keyError != nil {
		return nil, keyError
	}

	return bundlesNames, err
}

func CleanDistributionRepositories(t *testing.T, distributionDetails *config.ServerDetails) {
	deleteSpec := spec.NewBuilder().Pattern(tests.DistRepo1).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, distributionDetails)
	assert.NoError(t, err)

	deleteSpec = spec.NewBuilder().Pattern(tests.DistRepo2).BuildSpec()
	_, _, err = tests.DeleteFiles(deleteSpec, distributionDetails)
	assert.NoError(t, err)
}

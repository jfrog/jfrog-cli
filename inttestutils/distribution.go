package inttestutils

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-core/v2/common/spec"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

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
	artifactoryGpgKeyCreatePattern  = `{"alias":"cli tests distribution key","public_key":"%s"}`
)

type distributableDistributionStatus string
type receivedDistributionStatus string

const (
	// Release bundle created and open for changes:
	open distributableDistributionStatus = "OPEN"
	// Release bundle is signed, but not stored:
	signed distributableDistributionStatus = "SIGNED"
	// Release bundle is signed and stored, but not scanned by Xray:
	stored distributableDistributionStatus = "STORED"
	// Release bundle is signed, stored and scanned by Xray:
	readyForDistribution distributableDistributionStatus = "READY_FOR_DISTRIBUTION"

	NotDistributed receivedDistributionStatus = "Not distributed"
	InProgress     receivedDistributionStatus = "In progress"
	Completed      receivedDistributionStatus = "Completed"
	Failed         receivedDistributionStatus = "Failed"
)

// GET api/v1/release_bundle/:name/:version
// Retrieve the status of a release bundle before distribution.
type distributableResponse struct {
	utils.ReleaseBundleBody
	Name    string                          `json:"name,omitempty"`
	Version string                          `json:"version,omitempty"`
	State   distributableDistributionStatus `json:"state,omitempty"`
}

// Get api/v1/release_bundle/:name/:version/distribution
// Retrieve the status of a release bundle after distribution.
type receivedResponse struct {
	Id     string                     `json:"id,omitempty"`
	Status receivedDistributionStatus `json:"status,omitempty"`
}

type ReceivedResponses struct {
	receivedResponses []receivedResponse
}

// Send GPG keys to Distribution and Artifactory to allow signing of release bundles
func SendGpgKeys(artHttpDetails httputils.HttpClientDetails, distHttpDetails httputils.HttpClientDetails) {
	// Read gpg public and private keys
	keysDir := filepath.Join(tests.GetTestResourcesPath(), "distribution")
	publicKey, err := ioutil.ReadFile(filepath.Join(keysDir, "public.key.1"))
	coreutils.ExitOnErr(err)
	privateKey, err := ioutil.ReadFile(filepath.Join(keysDir, "private.key"))
	coreutils.ExitOnErr(err)

	// Create http client
	client, err := httpclient.ClientBuilder().Build()
	coreutils.ExitOnErr(err)

	// Send public and private keys to Distribution
	content := fmt.Sprintf(distributionGpgKeyCreatePattern, publicKey, privateKey)
	resp, body, err := client.SendPut(*tests.JfrogUrl+"distribution/api/v1/keys/pgp", []byte(content), distHttpDetails, "")
	coreutils.ExitOnErr(err)
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Send public key to Artifactory
	content = fmt.Sprintf(artifactoryGpgKeyCreatePattern, publicKey)
	resp, body, err = client.SendPost(*tests.JfrogUrl+"artifactory/api/security/keys/trusted", []byte(content), artHttpDetails, "")
	coreutils.ExitOnErr(err)
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusCreated, http.StatusConflict); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

// Get a local release bundle
func GetLocalBundle(t *testing.T, bundleName, bundleVersion string, distHttpDetails httputils.HttpClientDetails) *distributableResponse {
	resp, body := getLocalBundle(t, bundleName, bundleVersion, distHttpDetails)
	if err := errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK); err != nil {
		t.Error(err.Error())
		return nil
	}
	response := &distributableResponse{}
	err := json.Unmarshal(body, &response)
	if err != nil {
		t.Error(err)
		return nil
	}
	return response
}

// Return true if the release bundle exists locally on distribution
func VerifyLocalBundleExistence(t *testing.T, bundleName, bundleVersion string, expectExist bool, distHttpDetails httputils.HttpClientDetails) {
	for i := 0; i < 120; i++ {
		resp, body := getLocalBundle(t, bundleName, bundleVersion, distHttpDetails)
		switch resp.StatusCode {
		case http.StatusOK:
			if expectExist {
				return
			}
		case http.StatusNotFound:
			if !expectExist {
				return
			}
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

// Assert release bundle status is OPEN
func AssertReleaseBundleOpen(t *testing.T, distributableResponse *distributableResponse) {
	assert.NotNil(t, distributableResponse)
	assert.Equal(t, open, distributableResponse.State)
}

// Assert release bundle status is SIGNED, STORED or READY_FOR_DISTRIBUTION
func AssertReleaseBundleSigned(t *testing.T, distributableResponse *distributableResponse) {
	assert.NotNil(t, distributableResponse)
	assert.Contains(t, []distributableDistributionStatus{signed, stored, readyForDistribution}, distributableResponse.State)
}

// Wait for distribution of a release bundle
func WaitForDistribution(t *testing.T, bundleName, bundleVersion string, distHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	for i := 0; i < 120; i++ {
		resp, body, _, err := client.SendGet(*tests.JfrogUrl+"distribution/api/v1/release_bundle/"+bundleName+"/"+bundleVersion+"/distribution", true, distHttpDetails, "")
		assert.NoError(t, err)
		if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK); err != nil {
			t.Error(err.Error())
			return
		}
		response := &ReceivedResponses{}
		err = json.Unmarshal(body, &response.receivedResponses)
		if err != nil {
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
		case InProgress, NotDistributed:
			// Wait
		}
		t.Log("Waiting for " + bundleName + "/" + bundleVersion + "...")
		time.Sleep(time.Second)
	}
	t.Error("Timeout for release bundle distribution " + bundleName + "/" + bundleVersion)
}

// Wait for deletion of a release bundle
func WaitForDeletion(t *testing.T, bundleName, bundleVersion string, distHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	for i := 0; i < 120; i++ {
		resp, body, _, err := client.SendGet(*tests.JfrogUrl+"distribution/api/v1/release_bundle/"+bundleName+"/"+bundleVersion+"/distribution", true, distHttpDetails, "")
		assert.NoError(t, err)
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
	return resp, body
}

func CleanUpOldBundles(distHttpDetails httputils.HttpClientDetails, bundleVersion string, distributionCli *tests.JfrogCli) {
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

	// Build http client
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return nil, err
	}

	// Send get request
	resp, body, _, err := client.SendGet(*tests.JfrogUrl+"distribution/api/v1/release_bundle/distribution", true, distHttpDetails, "")
	if err != nil {
		return nil, err
	}
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK, http.StatusCreated); err != nil {
		return nil, err
	}

	// Extract release bundle names from the json response
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
		return nil, err
	}

	return bundlesNames, err
}

// Clean up 'cli-dist1-<timestamp>' and 'cli-dist2-<timestamp>' after running a distribution test
func CleanDistributionRepositories(t *testing.T, distributionDetails *config.ServerDetails) {
	deleteSpec := spec.NewBuilder().Pattern(tests.DistRepo1).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, distributionDetails)
	assert.NoError(t, err)
	deleteSpec = spec.NewBuilder().Pattern(tests.DistRepo2).BuildSpec()
	_, _, err = tests.DeleteFiles(deleteSpec, distributionDetails)
	assert.NoError(t, err)
}

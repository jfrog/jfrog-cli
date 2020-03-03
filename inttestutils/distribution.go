package inttestutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

const (
	gpgKeyId                        = "234503"
	distributionGpgKeyCreatePattern = `{"public_key":"%s","private_key":"%s"}`
	artifactoryGpgkeyCreatePattern  = `{"alias":"cli tests distribution key","public_key":"%s"}`
)

type distributableDistributionStatus string
type receivedDistributionStatus string

const (
	Open                 distributableDistributionStatus = "OPEN"
	ReadyForDistribution                                 = "READY_FOR_DISTRIBUTION"
	Signed                                               = "SIGNED"
	NotDistributed       receivedDistributionStatus      = "Not distributed"
	InProgress                                           = "In progress"
	Completed                                            = "Completed"
	Failed                                               = "Failed"
)

// GET api/v1/release_bundle/:name/:version
// Retreive the status of a release bundle before distribution.
type distributableResponse struct {
	Name    string                          `json:"name,omitempty"`
	Version string                          `json:"version,omitempty"`
	State   distributableDistributionStatus `json:"state,omitempty"`
}

// Get api/v1/release_bundle/:name/:version/distribution
// Retreive the status of a release bundle after distribution.
type receivedResponse struct {
	Id     string                     `json:"id,omitempty"`
	Status receivedDistributionStatus `json:"status,omitempty"`
}

type ReceivedResponses struct {
	receivedResponses []receivedResponse
}

// Send GPG keys to Distribution and Artifactory to allow signing of release bundles
func SendGpgKeys(artHttpDetails httputils.HttpClientDetails) {
	// Read gpg public and private keys
	keysDir := filepath.Join(tests.GetTestResourcesPath(), "distribution")
	publicKey, err := ioutil.ReadFile(filepath.Join(keysDir, "public.key"))
	cliutils.ExitOnErr(err)
	privateKey, err := ioutil.ReadFile(filepath.Join(keysDir, "private.key"))
	cliutils.ExitOnErr(err)

	// Create http client
	client, err := httpclient.ClientBuilder().Build()
	cliutils.ExitOnErr(err)

	// Send public and private keys to Distribution
	content := fmt.Sprintf(distributionGpgKeyCreatePattern, publicKey, privateKey)
	resp, body, err := client.SendPut(*tests.RtDistributionUrl+"api/v1/keys/pgp", []byte(content), artHttpDetails)
	cliutils.ExitOnErr(err)
	if resp.StatusCode != http.StatusOK {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}

	// Send public key to Artifactory
	content = fmt.Sprintf(artifactoryGpgkeyCreatePattern, publicKey)
	resp, body, err = client.SendPost(*tests.RtUrl+"api/security/keys/trusted", []byte(content), artHttpDetails)
	cliutils.ExitOnErr(err)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}
}

// Delete GPG key from Artifactory to clean up the test environment
func DeleteGpgKeys(artHttpDetails httputils.HttpClientDetails) {
	// Create http client
	client, err := httpclient.ClientBuilder().Build()
	cliutils.ExitOnErr(err)

	// Send public key to Artifactory
	resp, body, err := client.SendDelete(*tests.RtUrl+"api/security/keys/trusted/"+gpgKeyId, nil, artHttpDetails)
	cliutils.ExitOnErr(err)
	if resp.StatusCode != http.StatusNoContent {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}
}

// Return true if the release bundle is signed
func IsBundleSigned(t *testing.T, bundleName, bundleVersion string, artHttpDetails httputils.HttpClientDetails) bool {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	resp, body, _, err := client.SendGet(*tests.RtDistributionUrl+"api/v1/release_bundle/"+bundleName+"/"+bundleVersion, true, artHttpDetails)
	assert.NoError(t, err)
	if resp.StatusCode != http.StatusOK {
		t.Error(resp.Status)
		t.Error(string(body))
		return false
	}
	response := &distributableResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Error(err)
		return false
	}
	return response.State == Signed
}

// Wait for distribution of a release bundle
func WaitForDistribution(t *testing.T, bundleName, bundleVersion string, artHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	for i := 0; i < 120; i++ {
		resp, body, _, err := client.SendGet(*tests.RtDistributionUrl+"api/v1/release_bundle/"+bundleName+"/"+bundleVersion+"/distribution", true, artHttpDetails)
		assert.NoError(t, err)
		if resp.StatusCode != http.StatusOK {
			t.Error(resp.Status)
			t.Error(string(body))
			return
		}
		response := &ReceivedResponses{}
		err = json.Unmarshal(body, &response.receivedResponses)
		if err != nil {
			t.Error(err)
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
func WaitForDeletion(t *testing.T, bundleName, bundleVersion string, artHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	for i := 0; i < 120; i++ {
		resp, body, _, err := client.SendGet(*tests.RtDistributionUrl+"api/v1/release_bundle/"+bundleName+"/"+bundleVersion+"/distribution", true, artHttpDetails)
		assert.NoError(t, err)
		if resp.StatusCode == http.StatusNotFound {
			return
		}
		if resp.StatusCode != http.StatusOK {
			t.Error(resp.Status)
			t.Error(string(body))
			return
		}
		t.Log("Waiting for distribution deletion " + bundleName + "/" + bundleVersion + "...")
		time.Sleep(time.Second)
	}
	t.Error("Timeout for release bundle deletion " + bundleName + "/" + bundleVersion)
}

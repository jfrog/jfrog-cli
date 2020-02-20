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
	gpgKeyId                = "234503"
	releaseBundleAqlPattern = `{"name":"%s",` +
		`"version":"%s",` +
		`"dry_run":false,` +
		`"sign_immediately":true,` +
		`"spec":{"queries":[{"aql":"items.find(%s)"}]}}`
	repoPathNameAqlPattern = `{\"$and\":[` +
		`{\"repo\":{\"$match\":\"%s\"}},` +
		`{\"path\":{\"$match\":\"%s\"}},` +
		`{\"name\":{\"$match\":\"%s\"}}` +
		`]}`
	distributionGpgKeyCreatePattern = `{"public_key":"%s","private_key":"%s"}`
	artifactoryGpgkeyCreatePattern  = `{"alias":"cli tests distribution key","public_key":"%s"}`
	distributionPattern             = `{"dry_run":false,"distribution_rules":[{"site_name":"*"}]}`
)

type RepoPathName struct {
	Repo string
	Path string
	Name string
}

type DistributionStatus string

const (
	NotDistributed DistributionStatus = "Not distributed"
	InProgress                        = "In progress"
	Completed                         = "Completed"
	Failed                            = "Failed"
)

type DistributionResponse struct {
	Id     string             `json:"id,omitempty"`
	Status DistributionStatus `json:"status,omitempty"`
}

type DistributionResponses struct {
	distributionResponse []DistributionResponse
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

func CreateAndDistributeBundle(t *testing.T, bundleName, bundleVersion string, triples []RepoPathName, artHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	aql := createAqlForCreateBundle(bundleName, bundleVersion, triples)
	resp, body, err := client.SendPost(*tests.RtDistributionUrl+"api/v1/release_bundle", []byte(aql), artHttpDetails)
	assert.NoError(t, err)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNotFound {
		t.Error(resp.Status)
		t.Error(string(body))
	}
	distribute(t, bundleName, bundleVersion, artHttpDetails)
}

func distribute(t *testing.T, bundleName, bundleVersion string, artHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	url := *tests.RtDistributionUrl + "api/v1/distribution/" + bundleName + "/" + bundleVersion
	resp, body, err := client.SendPost(url, []byte(distributionPattern), artHttpDetails)
	assert.NoError(t, err)
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNotFound {
		t.Error(resp.Status)
		t.Error(string(body))
	}
	waitForDistribution(t, bundleName, bundleVersion, artHttpDetails)
}

func DeleteBundle(t *testing.T, bundleName, bundleVersion string, artHttpDetails httputils.HttpClientDetails) {
	// Delete distributable bundle on Distribution
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	resp, body, err := client.SendDelete(*tests.RtDistributionUrl+"api/v1/release_bundle/"+bundleName, nil, artHttpDetails)
	assert.NoError(t, err)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		t.Error(resp.Status)
		t.Error(string(body))
	}

	// Delete received bundle in Artifactory
	resp, body, err = client.SendDelete(*tests.RtUrl+"api/release/bundles/"+bundleName+"/"+bundleVersion, nil, artHttpDetails)
	assert.NoError(t, err)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Error(resp.Status)
		t.Error(string(body))
	}
}

// Create the AQL for the release bundle creation
func createAqlForCreateBundle(bundleName, bundleVersion string, triples []RepoPathName) string {
	innerQueryPattern := "{\\\"$or\\\":["
	for i, triple := range triples {
		innerQueryPattern += fmt.Sprintf(repoPathNameAqlPattern, triple.Repo, triple.Path, triple.Name)
		if i+1 < len(triples) {
			innerQueryPattern += ","
		}
	}
	return fmt.Sprintf(releaseBundleAqlPattern, bundleName, bundleVersion, innerQueryPattern+"]}")
}

// Wait for distribution of a release bundle
func waitForDistribution(t *testing.T, bundleName, bundleVersion string, artHttpDetails httputils.HttpClientDetails) {
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
		response := &DistributionResponses{}
		err = json.Unmarshal(body, &response.distributionResponse)
		if err != nil {
			t.Error(err)
			return
		}

		switch response.distributionResponse[0].Status {
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

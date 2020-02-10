package inttestutils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

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
	distributionGpgKeyCreatePattern = `{"public_key":"%s", "private_key":"%s"}`
	artifactoryGpgkeyCreatePattern  = `{"alias":"cli tests distribution key", "public_key":"%s"}`
)

type RepoPathName struct {
	Repo string
	Path string
	Name string
}

// Send GPG keys to Distribution and Artifactory to allow signing of release bundles
func SendGpgKeys(artHttpDetails httputils.HttpClientDetails) {
	// Read gpg public and private keys
	keysDir := filepath.Join(tests.GetTestResourcesPath(), "releasebundles")
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

// Create a release bundle
func CreateBundle(t *testing.T, bundleName, bundleVersion string, triples []RepoPathName, artHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	aql := createAqlForCreateBundle(bundleName, bundleVersion, triples)
	resp, body, err := client.SendPost(*tests.RtDistributionUrl+"api/v1/release_bundle", []byte(aql), artHttpDetails)
	assert.NoError(t, err)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNotFound {
		t.Error(resp.Status)
		t.Error(string(body))
	}
}

// Delete a release bundle
func DeleteBundle(t *testing.T, bundleName string, artHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	resp, body, err := client.SendDelete(*tests.RtDistributionUrl+"api/v1/release_bundle/"+bundleName, nil, artHttpDetails)
	assert.NoError(t, err)

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		t.Error(resp.Status)
		t.Error(string(body))
	}
}

// Create the AQL for the release bundle creation
func createAqlForCreateBundle(bundleName, bundleVersion string, triples []RepoPathName) string {
	innerQueryPattern := ""
	for i, triple := range triples {
		innerQueryPattern += fmt.Sprintf(repoPathNameAqlPattern, triple.Repo, triple.Path, triple.Name)
		if i+1 < len(triples) {
			innerQueryPattern += ","
		}
	}
	return fmt.Sprintf(releaseBundleAqlPattern, bundleName, bundleVersion, innerQueryPattern)
}

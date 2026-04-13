package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var apiCli *coreTests.JfrogCli

func InitApiTests() {
	initApiCli()
}

func initApiTest(t *testing.T) {
	if !*tests.TestApi {
		t.Skip("Skipping api command test. To run add the '-test.api=true' option.")
	}
}

func initApiCli() {
	if apiCli != nil {
		return
	}
	*tests.JfrogUrl = clientUtils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	apiCli = coreTests.NewJfrogCli(execMain, "jfrog", authenticateApiCmd())
}

func authenticateApiCmd() string {
	cred := fmt.Sprintf("--url=%s", *tests.JfrogUrl)
	if *tests.JfrogAccessToken != "" {
		cred += fmt.Sprintf(" --access-token=%s", *tests.JfrogAccessToken)
	} else {
		cred += fmt.Sprintf(" --user=%s --password=%s", *tests.JfrogUser, *tests.JfrogPassword)
	}
	return cred
}

// TestApiGetArtifactoryPing verifies a GET request to the Artifactory ping endpoint returns 200 with "OK".
func TestApiGetArtifactoryPing(t *testing.T) {
	initApiTest(t)
	stdout, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "artifactory/api/system/ping")
	require.NoError(t, err)
	assert.Equal(t, "200", strings.TrimSpace(string(stderr)))
	assert.Contains(t, strings.TrimSpace(string(stdout)), "OK")
}

// TestApiGetAccessPing verifies a GET request to the Access ping endpoint returns 200 with JSON {"status":"UP"}.
func TestApiGetAccessPing(t *testing.T) {
	initApiTest(t)
	stdout, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "access/api/v1/system/ping")
	require.NoError(t, err)
	assert.Equal(t, "200", strings.TrimSpace(string(stderr)))
	var result map[string]string
	require.NoError(t, json.Unmarshal(stdout, &result))
	assert.Equal(t, "UP", result["status"])
}

// TestApiVerbose verifies that --verbose writes request/response diagnostic lines to stderr.
func TestApiVerbose(t *testing.T) {
	initApiTest(t)
	stdout, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "--verbose", "artifactory/api/system/ping")
	require.NoError(t, err)
	stderrStr := string(stderr)
	assert.Contains(t, stderrStr, "* Request to")
	assert.Contains(t, stderrStr, "> GET")
	assert.Contains(t, stderrStr, "* Response")
	// Status code is still written as the last stderr line.
	lines := strings.Split(strings.TrimSpace(stderrStr), "\n")
	assert.Equal(t, "200", lines[len(lines)-1])
	assert.Contains(t, strings.TrimSpace(string(stdout)), "OK")
}

// TestApiNotFound verifies that a 404 response causes a non-zero exit.
func TestApiNotFound(t *testing.T) {
	initApiTest(t)
	err := apiCli.Exec("api", "artifactory/api/nosuchendpointxxx")
	assert.Error(t, err)
}

// TestApiMethodPost verifies that --method=POST is forwarded to the server.
// Uses the AQL search endpoint, which requires POST.
func TestApiMethodPost(t *testing.T) {
	initApiTest(t)
	_, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "--method=POST", `--data=items.find({})`, "artifactory/api/search/aql")
	require.NoError(t, err)
	assert.Equal(t, "200", strings.TrimSpace(string(stderr)))
}

// TestApiPostWithData verifies that --data sends a request body with a POST.
// Uses the AQL search endpoint, which requires a POST body.
func TestApiPostWithData(t *testing.T) {
	initApiTest(t)
	stdout, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "--method=POST", `--data=items.find({})`, "artifactory/api/search/aql")
	require.NoError(t, err)
	assert.Equal(t, "200", strings.TrimSpace(string(stderr)))
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(stdout, &result))
	_, hasResults := result["results"]
	assert.True(t, hasResults)
}

// TestApiCustomHeader verifies that a custom --header value is accepted and the request succeeds.
func TestApiCustomHeader(t *testing.T) {
	initApiTest(t)
	_, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "--header=X-Jfrog-Test: integration-test", "access/api/v1/system/ping")
	require.NoError(t, err)
	assert.Equal(t, "200", strings.TrimSpace(string(stderr)))
}

// TestApiMethodCaseInsensitive verifies that method names are normalized to uppercase.
func TestApiMethodCaseInsensitive(t *testing.T) {
	initApiTest(t)
	stdout, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "--method=get", "artifactory/api/system/ping")
	require.NoError(t, err)
	assert.Equal(t, "200", strings.TrimSpace(string(stderr)))
	assert.Contains(t, strings.TrimSpace(string(stdout)), "OK")
}

// TestApiLeadingSlashInPath verifies that a path without a leading slash is handled correctly.
func TestApiLeadingSlashInPath(t *testing.T) {
	initApiTest(t)
	// Without leading slash — the command should normalise it internally.
	stdout, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "artifactory/api/system/ping")
	require.NoError(t, err)
	assert.Equal(t, "200", strings.TrimSpace(string(stderr)))
	assert.Contains(t, strings.TrimSpace(string(stdout)), "OK")
}

// TestApiWithInputFile verifies that --input reads a request body from a file.
func TestApiWithInputFile(t *testing.T) {
	initApiTest(t)
	// Write an AQL query to a temp file and POST it to the AQL search endpoint.
	payloadFile := tests.CreateTempFile(t, "items.find({})")
	_, stderr, err := tests.GetCmdOutput(t, apiCli, "api", "--method=POST", "--input="+payloadFile, "artifactory/api/search/aql")
	require.NoError(t, err)
	assert.Equal(t, "200", strings.TrimSpace(string(stderr)))
}

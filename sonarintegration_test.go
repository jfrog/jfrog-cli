package main

import (
	"encoding/json"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"testing"
)

var (
	sonarIntegrationCLI *coreTests.JfrogCli
)

func initSonarCli() {
	if sonarIntegrationCLI != nil {
		return
	}
	sonarIntegrationCLI = coreTests.NewJfrogCli(execMain, "jfrog", authenticateAccess())
}

func initSonarIntegrationTest(t *testing.T) {
	if !*tests.TestSonar {
		t.Skip("Skipping Access test. To run Access test add the '-test.access=true' option.")
	}
}

func TestSonarIntegration(t *testing.T) {
	initSonarCli()
	initSonarIntegrationTest(t)
	// Generate access token from sonarqube
	getSonarAccessToken(t)
	// Create the project and settings on sonarqube
	createAndConfigureSonarProject(t)
	// Create build info/artifact/package from sample projects
	// trigger sonar scan from either plugins of maven/gradle
	// trigger sonar scan using sonar-cli
	// Fetch the sonar evidence and attach against the artifacts using jfrog-cli
}

func getSonarAccessToken(t *testing.T) string {
	client := createHttpClient(t, "")
	req, err := createFetchSonarAccessTokenRequest(t)
	resp, err := client.Do(req)
	if err != nil {
		assert.NoError(t, err)
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		assert.NoError(t, err)
	}
	var result struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	return result.Token
}

func createFetchSonarAccessTokenRequest(t *testing.T) (*http.Request, error) {
	req, err := http.NewRequest("POST", "http://localhost:9000/api/user_tokens/generate", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetBasicAuth("admin", "admin")
	q := req.URL.Query()
	q.Add("name", "jfrog-cli-token")
	req.URL.RawQuery = q.Encode()
	return req, err
}

func createAndConfigureSonarProject(t *testing.T) {
	// This function should create a SonarQube project and configure it as needed.
	// It can include API calls to SonarQube to set up the project, quality gates, etc.
	req, err := http.NewRequest("POST", "http://localhost:9000/api/projects/create", nil)
	if err != nil {
		assert.NoError(t, err)
	}
	req.Header.Set("Authorization", "Bearer "+getSonarAccessToken(t))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}

func createHttpClient(t *testing.T, proxy string) *http.Client {
	// Create a custom HTTP client with proxy settings if needed
	client := &http.Client{}
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			assert.NoError(t, err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	return client
}

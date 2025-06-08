package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	configUtils "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

var (
	sonarIntegrationCLI *coreTests.JfrogCli
	evidenceDetails     *configUtils.ServerDetails
)

func initSonarCli() {
	if sonarIntegrationCLI != nil {
		return
	}
	sonarIntegrationCLI = coreTests.NewJfrogCli(execMain, "jfrog", authenticateEvidence())
}

func initSonarIntegrationTest(t *testing.T) {
	if !*tests.TestSonar {
		t.Skip("Skipping Access test. To run Access test add the '-test.access=true' option.")
	}
}

func authenticateEvidence() string {
	*tests.JfrogUrl = clientUtils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	evidenceDetails = &configUtils.ServerDetails{
		Url: *tests.JfrogUrl}
	evidenceDetails.EvidenceUrl = clientUtils.AddTrailingSlashIfNeeded(evidenceDetails.Url) + "evidence/"

	cred := fmt.Sprintf("--url=%s", *tests.JfrogUrl)
	if *tests.JfrogAccessToken != "" {
		evidenceDetails.AccessToken = *tests.JfrogAccessToken
		cred += fmt.Sprintf(" --access-token=%s", evidenceDetails.AccessToken)
	} else {
		evidenceDetails.User = *tests.JfrogUser
		evidenceDetails.Password = *tests.JfrogPassword
		cred += fmt.Sprintf(" --user=%s --password=%s", evidenceDetails.User, evidenceDetails.Password)
	}
	return cred
}

func TestSonarPrerequisites(t *testing.T) {
	initSonarIntegrationTest(t)
	// read the file called report-task.txt
	reportFilePath := "testdata/maven/mavenprojectwithsonar/target/sonar/report-task.txt"
	if _, err := os.Stat(reportFilePath); os.IsNotExist(err) {
		t.Fatalf("Failed to find file %s", reportFilePath)
	}
	// read file content
	fileContent, err := os.ReadFile(reportFilePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", reportFilePath, err)
	}
	found := false
	sonarURL := ""
	for _, line := range strings.Split(string(fileContent), "\n") {
		if strings.HasPrefix(line, "ceTaskUrl=") {
			found = true
			sonarURL = strings.TrimPrefix(line, "ceTaskUrl=")
			break
		}
	}
	if !found {
		t.Fatalf("File %s does not contain 'ceTaskUrl=' in any line", reportFilePath)
	}
	if sonarURL == "" {
		t.Fatalf("File %s does not contain a valid SonarQube URL", reportFilePath)
	}
	assert.True(t, strings.HasPrefix(sonarURL, "http://localhost:9000/api/ce/task?id="), "SonarQube URL is not valid: %s", sonarURL)
	taskID := strings.TrimPrefix(sonarURL, "http://localhost:9000/api/ce/task?id=")
	assert.NotEmpty(t, taskID, "Evidence successfully created and verified")
}

func TestSonarIntegrationAsEvidence(t *testing.T) {
	initSonarCli()
	initSonarIntegrationTest(t)

	// Get the SonarQube access token
	setSonarAccessTokenFromEnv(t)
	privateKeyFilePath, publicKeyName := KeyPairGenerationAndUpload(t)
	// Run the JFrog CLI command to collect evidence
	output := sonarIntegrationCLI.RunCliCmdWithOutput(t, "evidence", "create", "--predicate-type=\"https://jfrog.com/evidence/sonarqube/v1\"", "--package-name=demo-sonar", "--package-version=1.0", "--package-repo-name=dev-maven-local", "--key-alias="+publicKeyName, "--key-path="+privateKeyFilePath)
	assert.Contains(t, output, "Successfully created evidence for SonarQube analysis")
	_, err := utils.CreateEvidenceServiceManager(evidenceDetails, false)
	assert.NoError(t, err)
}

func KeyPairGenerationAndUpload(t *testing.T) (string, string) {
	artifactoryURL := os.Getenv("PLATFORM_URL")
	apiKey := os.Getenv("PLATFORM_API_KEY")
	publicKeyName := "test-evidence-key"
	privateKeyPath := "./test-evidence-private.pem"
	publicKeyPath := "./test-evidence-public.pem"

	// 1. Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// 2. Save private key to file
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	err = os.WriteFile(privateKeyPath, privPem, 0600)
	if err != nil {
		t.Fatalf("Failed to write private key: %v", err)
	}

	// 3. Save public key to file
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	err = os.WriteFile(publicKeyPath, pubPem, 0644)
	if err != nil {
		t.Fatalf("Failed to write public key: %v", err)
	}

	// 4. Upload public key to Artifactory
	uploadPublicKeyToArtifactory(t, artifactoryURL, apiKey, publicKeyName, publicKeyPath)
	return privateKeyPath, publicKeyName
}

func setSonarAccessTokenFromEnv(t *testing.T) {
	sonarToken := os.Getenv("SONAR_TOKEN")
	assert.NotEmpty(t, sonarToken, "SONAR_TOKEN should not be empty")
	err := os.Setenv("JF_SONAR_ACCESS_TOKEN", sonarToken)
	assert.NoError(t, err)
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

// uploadPublicKeyToArtifactory uploads a PEM-encoded public key to Artifactory Evidence with the given alias.
func uploadPublicKeyToArtifactory(t *testing.T, artifactoryURL, apiKey, alias, publicKeyPath string) {
	pubKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		t.Fatalf("Failed to read public key file: %v", err)
	}
	url := fmt.Sprintf("%s/api/v1/evidence/publicKey/%s", artifactoryURL, alias)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(pubKeyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload public key: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to upload public key, status: %s, body: %s", resp.Status, string(body))
	}
}

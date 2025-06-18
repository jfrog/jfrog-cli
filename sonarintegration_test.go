package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	configUtils "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	sonarIntegrationCLI *coreTests.JfrogCli
	evidenceDetails     *configUtils.ServerDetails
)

type KeyPair struct {
	PairName   string `json:"pairName"`
	PairType   string `json:"pairType"`
	Alias      string `json:"alias"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

const (
	KeyPairAlias = "evidence-local"
	keyPairName  = "test-signing-key"
)

func initSonarCli() {
	if sonarIntegrationCLI != nil {
		return
	}
	sonarIntegrationCLI = coreTests.NewJfrogCli(execMain, "jfrog", authenticateEvidence())
}

func initSonarIntegrationTest(t *testing.T) {
	if !*tests.TestSonar {
		t.Skip("Skipping Access test. To run Access test add the '-test.sonarIntegration=true' option.")
	}
	// check if JF_SONARQUBE_ACCESS_TOKEN env variable is empty then throw an error
	if os.Getenv("JF_SONARQUBE_ACCESS_TOKEN") == "" {
		t.Fatal("JF_SONARQUBE_ACCESS_TOKEN environment variable is not set. Please set it to run the SonarQube integration test.")
	}
}

func authenticateEvidence() string {
	*tests.JfrogUrl = clientUtils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	evidenceDetails = &configUtils.ServerDetails{
		Url: *tests.JfrogUrl}
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
	reportFilePath := "testdata/maven/mavenprojectwithsonar/target/sonar/report-task.txt"
	if _, err := os.Stat(reportFilePath); os.IsNotExist(err) {
		t.Fatalf("Failed to find file %s", reportFilePath)
	}
	fileContent, err := os.ReadFile(reportFilePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", reportFilePath, err)
	}
	isCeTaskUrlFound := false
	sonarURL := ""
	for _, line := range strings.Split(string(fileContent), "\n") {
		if strings.HasPrefix(line, "ceTaskUrl=") {
			isCeTaskUrlFound = true
			sonarURL = strings.TrimPrefix(line, "ceTaskUrl=")
			break
		}
	}
	assert.True(t, isCeTaskUrlFound, "File %s does not contain 'ceTaskUrl='", reportFilePath)
	assert.NotEmpty(t, "File %s does not contain a valid SonarQube URL", reportFilePath)
	assert.True(t, strings.HasPrefix(sonarURL, "http://localhost:9000/api/ce/task?id="), "SonarQube URL is not valid: %s", sonarURL)
	taskID := strings.TrimPrefix(sonarURL, "http://localhost:9000/api/ce/task?id=")
	assert.NotEmpty(t, taskID, "task ID should not be empty")
	resp, err := http.Get("http://localhost:9000/api/system/status")
	if err != nil {
		t.Fatalf("Failed to connect to SonarQube server: %v", err)
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK, "SonarQube server is not running or returned an unexpected status code")
	// Check if given sonar_access_token is valid
	sonarAccessToken := os.Getenv("JF_SONARQUBE_ACCESS_TOKEN")
	if sonarAccessToken == "" {
		t.Fatal("JF_SONARQUBE_ACCESS_TOKEN environment variable is not set. Please set it to run the SonarQube integration test.")
	}
	// use sonarAccessToken to authenticate with SonarQube
	req, err := http.NewRequest("GET", "http://localhost:9000/api/authentication/validate", nil)
	req.Header.Set("Authorization", "Bearer "+sonarAccessToken)
	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err, "Failed to validate SonarQube access token")
}

func TestSonarIntegrationAsEvidence(t *testing.T) {
	initSonarCli()
	initSonarIntegrationTest(t)
	privateKeyFilePath := KeyPairGenerationAndUpload(t)
	err := os.Setenv("JFROG_CLI_LOG_LEVEL", "DEBUG")
	assert.NoError(t, err)
	defer os.Unsetenv("JFROG_CLI_LOG_LEVEL")
	// Change to the directory containing the Maven project and execute cli command
	origDir, err := os.Getwd()
	assert.NoError(t, err)
	newDir := "testdata/maven/mavenprojectwithsonar"
	err = os.Chdir(newDir)
	assert.NoError(t, err)
	defer func() {
		err := os.Chdir(origDir)
		assert.NoError(t, err)
	}()
	output := sonarIntegrationCLI.RunCliCmdWithOutput(t, "evd", "create", "--predicate-type=sonar",
		"--package-name=demo-sonar", "--package-version=1.0", "--package-repo-name=dev-maven-local",
		"--key-alias="+keyPairName, "--key="+privateKeyFilePath)
	assert.Empty(t, output)
	evidenceResponse, err := FetchEvidenceFromArtifactory(t, *tests.JfrogUrl, *tests.JfrogAccessToken, "dev-maven-local", "demo-sonar", "1.0")
	assert.NoError(t, err)
	t.Logf("Evidence response: %s", evidenceResponse)
}

func TestSonarIntegrationAsEvidenceWithZeroConfig(t *testing.T) {
	initSonarCli()
	initSonarIntegrationTest(t)
	privateKeyFilePath := KeyPairGenerationAndUpload(t)
	err := os.Setenv("JFROG_CLI_LOG_LEVEL", "DEBUG")
	assert.NoError(t, err)
	defer os.Unsetenv("JFROG_CLI_LOG_LEVEL")
	// Change to the directory containing the Maven project and execute cli command
	origDir, err := os.Getwd()
	assert.NoError(t, err)
	newDir := "testdata/maven/mavenprojectwithsonar"
	err = os.Chdir(newDir)
	assert.NoError(t, err)
	defer func() {
		err := os.Chdir(origDir)
		assert.NoError(t, err)
	}()
	// Remove the directory .jfrog/evidence
	evidenceDir := filepath.Join(".jfrog", "evidence")
	err = os.RemoveAll(evidenceDir)
	assert.NoError(t, err)
	output := sonarIntegrationCLI.RunCliCmdWithOutput(t, "evd", "create", "--predicate-type=sonar",
		"--package-name=demo-sonar", "--package-version=1.0", "--package-repo-name=dev-maven-local",
		"--key-alias="+keyPairName, "--key="+privateKeyFilePath)
	assert.Empty(t, output)
	evidenceResponse, err := FetchEvidenceFromArtifactory(t, *tests.JfrogUrl, *tests.JfrogAccessToken, "dev-maven-local", "demo-sonar", "1.0")
	assert.NoError(t, err)
	t.Logf("Evidence response: %s", evidenceResponse)
}

// KeyPairGenerationAndUpload Deletes the existing signing key from Artifactory,
// generates a new RSA key pair, and uploads it to Artifactory.
func KeyPairGenerationAndUpload(t *testing.T) string {
	artifactoryURL := os.Getenv("PLATFORM_URL")
	apiKey := os.Getenv("PLATFORM_API_KEY")
	assert.NotEmpty(t, artifactoryURL)
	assert.NotEmpty(t, apiKey, "PLATFORM_API_KEY should not be empty")
	deleteSigningKeyFromArtifactory(t, artifactoryURL, apiKey, keyPairName)
	privateKeyFilePath, publicKeyFilePath, err := generateRSAKeyPair()
	assert.NoError(t, err)
	UploadSigningKeyPairToArtifactory(t, artifactoryURL, apiKey, privateKeyFilePath, publicKeyFilePath)
	return privateKeyFilePath
}

func generateRSAKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubPem := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	tempDir := os.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	pubPath := filepath.Join(tempDir, "public.pem")
	err = os.WriteFile(privateKeyPath, pem.EncodeToMemory(privateKeyPEM), 0600)
	if err != nil {
		return "", "", err
	}
	err = os.WriteFile(pubPath, pem.EncodeToMemory(pubPem), 0644)
	if err != nil {
		return "", "", err
	}
	return privateKeyPath, pubPath, nil
}

func deleteSigningKeyFromArtifactory(t *testing.T, artifactoryURL, apiKey, keyPairName string) {
	assert.NotEmpty(t, artifactoryURL)
	assert.NotEmpty(t, apiKey, "PLATFORM_API_KEY should not be empty")
	url := fmt.Sprintf("%sartifactory/api/security/keypair/%s", artifactoryURL, keyPairName)
	log.Debug(url)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	assert.NoError(t, err)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			assert.NoError(t, err, "Failed to close response body")
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to delete private key, status: %s, body: %s", resp.Status, string(body))
	}
}

// UploadSigningKeyPairToArtifactory reads private and public key files and uploads them to Artifactory.
func UploadSigningKeyPairToArtifactory(t *testing.T, artifactoryURL, apiKey, privateKeyPath, publicKeyPath string) {
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to read private key file: %v", err)
	}
	pubKeyBytes, err := os.ReadFile(publicKeyPath)
	assert.NoError(t, err)
	url := fmt.Sprintf("%sartifactory/api/security/keypair", artifactoryURL)
	t.Logf("Keypair create URL %s", url)
	reqBody := KeyPair{
		PairName:   keyPairName,
		PairType:   "RSA",
		Alias:      KeyPairAlias,
		PrivateKey: string(privateKeyBytes),
		PublicKey:  string(pubKeyBytes),
	}
	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			assert.NoError(t, err, "Failed to close response body")
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to upload private key, status: %s, body: %s", resp.Status, string(body))
	}
}

// FetchEvidenceFromArtifactory fetches evidence using GraphQL API
func FetchEvidenceFromArtifactory(t *testing.T, artifactoryURL, apiKey, packageRepo, packageName, packageVersion string) ([]byte, error) {
	// Construct the GraphQL API URL
	url := fmt.Sprintf("%sonemodel/api/v1/graphql", clientUtils.AddTrailingSlashIfNeeded(artifactoryURL))

	t.Logf("Fetching evidence from GraphQL API: %s", url)

	// Construct the GraphQL query using the working format
	query := fmt.Sprintf(`{
		evidence {
			searchEvidence(where:{hasSubjectWith:{repositoryKey:"%s"}}) {
				edges {
					node {
						name
						path
						repositoryKey
						downloadPath
						sha256
						predicateType
						createdAt
						createdBy
						verified
						predicateSlug
					}
				}
			}
		}
	}`, packageRepo)

	// Create request payload
	requestBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %v", err)
	}

	// Create the request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	} else {
		t.Fatal("API key is required to fetch evidence from Artifactory")
	}
	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			assert.NoError(t, err, "Failed to close response body")
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch evidence, status: %s, body: %s", resp.Status, string(body))
	}
	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return bodyBytes, nil
}

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

const KeyPairAlias = "evidence-local"

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
	assert.NotEmpty(t, taskID, "task ID should not be empty")
}

func TestSonarIntegrationAsEvidence(t *testing.T) {
	initSonarCli()
	initSonarIntegrationTest(t)

	// Get the SonarQube access token
	setSonarAccessTokenFromEnv(t)
	privateKeyFilePath := KeyPairGenerationAndUpload(t)
	// Run the JFrog CLI command to collect evidence
	output := sonarIntegrationCLI.RunCliCmdWithOutput(t, "evidence", "create", "--predicate-type=sonar", "--package-name=demo-sonar", "--package-version=1.0", "--package-repo-name=dev-maven-local", "--key-alias="+KeyPairAlias, "--key-path="+privateKeyFilePath)
	assert.Contains(t, output, "Successfully created evidence for SonarQube analysis")
	_, err := utils.CreateEvidenceServiceManager(evidenceDetails, false)
	assert.NoError(t, err)
}

func KeyPairGenerationAndUpload(t *testing.T) string {
	artifactoryURL := os.Getenv("PLATFORM_URL")
	apiKey := os.Getenv("PLATFORM_API_KEY")
	assert.NotEmpty(t, artifactoryURL)
	assert.NotEmpty(t, apiKey, "PLATFORM_API_KEY should not be empty")

	privateKeyFilePath, publicKeyFilePath, err := generateRSAKeyPair()
	assert.NoError(t, err)

	UploadSigningKeyPairToArtifactory(t, artifactoryURL, apiKey, privateKeyFilePath, publicKeyFilePath)
	return privateKeyFilePath
}

func generateRSAKeyPair() (string, string, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privPem := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubPem := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	tempDir := os.TempDir()
	privPath := filepath.Join(tempDir, "private.pem")
	pubPath := filepath.Join(tempDir, "public.pem")
	err = os.WriteFile(privPath, pem.EncodeToMemory(privPem), 0600)
	if err != nil {
		return "", "", err
	}
	err = os.WriteFile(pubPath, pem.EncodeToMemory(pubPem), 0644)
	if err != nil {
		return "", "", err
	}
	return privPath, pubPath, nil
}

func setSonarAccessTokenFromEnv(t *testing.T) {
	sonarToken := os.Getenv("SONAR_TOKEN")
	assert.NotEmpty(t, sonarToken, "SONAR_TOKEN should not be empty")
	err := os.Setenv("JF_SONAR_ACCESS_TOKEN", sonarToken)
	assert.NoError(t, err)
}

func UploadSigningKeyPairToArtifactory(t *testing.T, artifactoryURL, apiKey, privateKeyPath, publicKeyPath string) {
	// Read the private key file
	privKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to read private key file: %v", err)
	}
	pubKeyBytes, err := os.ReadFile(publicKeyPath)
	assert.NoError(t, err)
	// Upload the private key to Artifactory Evidence
	url := fmt.Sprintf("%sartifactory/api/security/keypair", artifactoryURL)
	log.Debug(url)
	reqBody := KeyPair{
		PairName:   "test-signing-key",
		PairType:   "RSA",
		Alias:      KeyPairAlias,
		PrivateKey: string(privKeyBytes),
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to upload private key, status: %s, body: %s", resp.Status, string(body))
	}
}

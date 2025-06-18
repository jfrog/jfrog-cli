package main

import (
	"bytes"
	"encoding/json"
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
	"time"
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

type EvidenceResponse struct {
	Data Data `json:"data"`
}
type Node struct {
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	RepositoryKey string    `json:"repositoryKey"`
	DownloadPath  string    `json:"downloadPath"`
	Sha256        string    `json:"sha256"`
	PredicateType string    `json:"predicateType"`
	CreatedAt     time.Time `json:"createdAt"`
	CreatedBy     string    `json:"createdBy"`
	Verified      bool      `json:"verified"`
	PredicateSlug string    `json:"predicateSlug"`
}
type Edges struct {
	Node Node `json:"node"`
}
type SearchEvidence struct {
	Edges []Edges `json:"edges"`
}
type Evidence struct {
	SearchEvidence SearchEvidence `json:"searchEvidence"`
}
type Data struct {
	Evidence Evidence `json:"evidence"`
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
	evidenceResponseBytes, err := FetchEvidenceFromArtifactory(t, *tests.JfrogUrl, *tests.JfrogAccessToken, "dev-maven-local", "demo-sonar", "1.0")
	assert.NoError(t, err)
	// Unmarshal the response into EvidenceResponse struct
	var evidenceResponse EvidenceResponse
	err = json.Unmarshal(evidenceResponseBytes, &evidenceResponse)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(evidenceResponse.Data.Evidence.SearchEvidence.Edges))
	assert.Equal(t, evidenceResponse.Data.Evidence.SearchEvidence.Edges[0].Node.Path, "com/example/demo-sonar/1.0/demo-sonar-1.0.pom")
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
	// Remove evidence configuration so that the zero config will be used
	evidenceDir := filepath.Join(".jfrog", "evidence")
	err = os.RemoveAll(evidenceDir)
	assert.NoError(t, err)
	output := sonarIntegrationCLI.RunCliCmdWithOutput(t, "evd", "create", "--predicate-type=sonar",
		"--package-name=demo-sonar", "--package-version=1.0", "--package-repo-name=dev-maven-local",
		"--key-alias="+keyPairName, "--key="+privateKeyFilePath)
	assert.Empty(t, output)
	evidenceResponseBytes, err := FetchEvidenceFromArtifactory(t, *tests.JfrogUrl, *tests.JfrogAccessToken, "dev-maven-local", "demo-sonar", "1.0")
	assert.NoError(t, err)
	// Unmarshal the response into EvidenceResponse struct
	var evidenceResponse EvidenceResponse
	err = json.Unmarshal(evidenceResponseBytes, &evidenceResponse)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(evidenceResponse.Data.Evidence.SearchEvidence.Edges))
	assert.Equal(t, evidenceResponse.Data.Evidence.SearchEvidence.Edges[1].Node.Path, "com/example/demo-sonar/1.0/demo-sonar-1.0.pom")
}

func TestSonarIntegrationEvidenceCollectionWithBuildPublish(t *testing.T) {
	initSonarCli()
	initSonarIntegrationTest(t)
	//privateKeyFilePath := KeyPairGenerationAndUpload(t)
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
	evidenceDetails.ArtifactoryUrl = *tests.JfrogUrl + "artifactory/"
	copyEvidenceYaml(t)
	output := sonarIntegrationCLI.RunCliCmdWithOutput(t, "rt", "bp", "test-sonar-jf-cli-integration", "1")
	assert.Empty(t, output)
	evidenceResponseBytes, err := FetchEvidenceFromArtifactory(t, *tests.JfrogUrl, *tests.JfrogAccessToken, "dev-maven-local", "demo-sonar", "1.0")
	assert.NoError(t, err)
	var evidenceResponse EvidenceResponse
	err = json.Unmarshal(evidenceResponseBytes, &evidenceResponse)
	assert.NoError(t, err)
	//assert.Equal(t, 2, len(evidenceResponse.Data.Evidence.SearchEvidence.Edges))
	//assert.Equal(t, evidenceResponse.Data.Evidence.SearchEvidence.Edges[1].Node.Path, "com/example/demo-sonar/1.0/demo-sonar-1.0.pom")
	t.Logf("Evidence created successfully with build info: %s", evidenceResponse.Data.Evidence.SearchEvidence.Edges[1].Node.Path)
}

func copyEvidenceYaml(t *testing.T) {
	src := "evidence.yaml"
	dstDir := filepath.Join(".jfrog", "evidence")
	dst := filepath.Join(dstDir, "evidence.yaml")

	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory %s: %v", dstDir, err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		t.Fatalf("Failed to open source file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}
}

// KeyPairGenerationAndUpload Deletes the existing signing key from Artifactory,
// generates a new RSA key pair, and uploads it to Artifactory.
func KeyPairGenerationAndUpload(t *testing.T) string {
	artifactoryURL := os.Getenv("PLATFORM_URL")
	apiKey := os.Getenv("PLATFORM_API_KEY")
	assert.NotEmpty(t, artifactoryURL)
	assert.NotEmpty(t, apiKey, "PLATFORM_API_KEY should not be empty")
	//deleteSigningKeyFromArtifactory(t, artifactoryURL, apiKey, keyPairName)
	privateKeyFilePath, publicKeyFilePath, err := generateRSAKeyPair()
	assert.NoError(t, err)
	if FetchSigningKeyPairFromArtifactory(t, artifactoryURL, apiKey) {
		return privateKeyFilePath
	}
	UploadSigningKeyPairToArtifactory(t, artifactoryURL, apiKey, privateKeyFilePath, publicKeyFilePath)
	return privateKeyFilePath
}

func generateRSAKeyPair() (string, string, error) {
	//privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	//if err != nil {
	//	return "", "", err
	//}
	//privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	//privateKeyPEM := &pem.Block{
	//	Type:  "RSA PRIVATE KEY",
	//	Bytes: privateKeyBytes,
	//}
	//pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	//if err != nil {
	//	return "", "", err
	//}
	//pubPem := &pem.Block{
	//	Type:  "PUBLIC KEY",
	//	Bytes: pubBytes,
	//}
	privateKeyString := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDFgJe3kRIYML2R
Kjjp70XbF+WVsUWdZLN6H3Hzm3FVhVcHcYpLKGxGhbTVN3yAtAA5CLqe4+BXOybM
ACV2NboEV0KhSXcx6MAShyMm/6Ze4POF07yMifewjOstsrxGg4FkL38n3MYQm7y3
bipFDXg93uGb8zVWG0wcqa5v1u0dD56xoTGSRrEtdjogFtkYVysXcyg7zKzzfQeH
zFwm3jZAG6wDwIlut00vTO62gVopnll+FZnTDSeYZy4nXh4Qo6v1F/gMmV0fIHNh
ZENjf2Y/TYROr0u67qH3XmgZqsi9hTi+OL2H14iRuwm6erKTenH4XhnvsTIOokcg
EdoE9LDFAgMBAAECggEAVEkvNjNOjg1K8UccF9W5sakunOYU5/kgURdXWZe2U8F+
ZRpS4wVCxAvuoum1k/V9fNmZTxK/3GpNgdT0J9EA7DZTJKLGIAIM6jtKyKtklGwa
8Ttt5WpBztIs0YlMKSmZECjm8puY2WClNoDowCRh8sGJ9bRiyDcJEdhmLauC8JnI
YmJC1c/vFp0FBw/jw5euEPKa559nIFN5Wbwxrl/6A6S7Lp7AuebLlHeLanu7X7e0
BNhT6sLOhnHjTFomex/z7eg9g5O577OjuYrw1a+81y6CkXTu6a35tnqg9RWtM7JX
WCjo/f/iO/ZE1F/qmu3x97b6Ljuv3yAFNeKfVEQatwKBgQDoH53rrunSCF+dXQTG
MZ9bTcUCw1a8saugC2guJ+xSt8HA0I6PYvqUZWfYgmMm6J8Vu/h1kj7kGIuugX0W
IX9OPIB86mQa/djTfPWaWmnPYwxRQ8DPkzxkdm2qcldY4UwrPo3nsFvGyD6Xfzkp
d7JlDv0cNtcE+rdIHMSTk/blswKBgQDZ0VCOP2sNAZ5uHeS+ksnmxAD00Jt6VukX
Sw9bsBNFeGP2G3m086xhCPMm0PlmuPitRCdzQypJcAJwQTaOFbf4KLBYpEIo2YJb
QXaiQaQZXeWRxzUWysBmsqcSfzAod4BLwimkSGXbHYC9ryanJ7iliNFzyWSpj/sV
ld9y9p1DpwKBgCHL6KxWDUk9Wt6ImpdYxkD+875RPqG+pKRqxMJjoa7xfk5aj0cl
PCK7GQGXCmSx3efGNIi5wFppkHzZ8aJ1QhncCUEmx2h+qUExonjUzS8a1sJGQR53
64UdER6OA1W3h+WL+BFRxisNIL/iECqPePPp2MRw36Gj92eSeLScCIitAoGAKzyK
YgIirM1CdpdGfbHDlCQaEH6MLkesMyx6Gvgjiymvpf2kNhAcipJtOapHp2VWL4aU
0iNl9HfgdAnt21xiTUc+YgoQ++zZHGYtN14SRdrGpB5H4oNSl9Akq95FX/MAq4ka
HPsmBM2hbYWkBZAz7d/vu60hZysmaw158mcTpocCgYEAkkLv5jtKEHOCJjrdyYl0
5Bv3Z22NTUdKaFY8wZnqmVBlJVsDG2D6Ypw3NEAQPKY5PJ44XSsM+nPjbBloyLpJ
k4UTtgRSG5/ZgMcDjJIDZIuIivah/g0I+ZkLBmyh8mOdEL/skGvj4iWH0It0V2l5
IAecx7gdLfPlyBAFZ5Jp9rc=
-----END PRIVATE KEY-----`
	publicKeyString := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxYCXt5ESGDC9kSo46e9F
2xfllbFFnWSzeh9x85txVYVXB3GKSyhsRoW01Td8gLQAOQi6nuPgVzsmzAAldjW6
BFdCoUl3MejAEocjJv+mXuDzhdO8jIn3sIzrLbK8RoOBZC9/J9zGEJu8t24qRQ14
Pd7hm/M1VhtMHKmub9btHQ+esaExkkaxLXY6IBbZGFcrF3MoO8ys830Hh8xcJt42
QBusA8CJbrdNL0zutoFaKZ5ZfhWZ0w0nmGcuJ14eEKOr9Rf4DJldHyBzYWRDY39m
P02ETq9Luu6h915oGarIvYU4vji9h9eIkbsJunqyk3px+F4Z77EyDqJHIBHaBPSw
xQIDAQAB
-----END PUBLIC KEY-----`
	tempDir := os.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	pubPath := filepath.Join(tempDir, "public.pem")
	err := os.WriteFile(privateKeyPath, []byte(privateKeyString), 0600)
	if err != nil {
		return "", "", err
	}
	err = os.WriteFile(pubPath, []byte(publicKeyString), 0644)
	if err != nil {
		return "", "", err
	}
	return privateKeyPath, pubPath, nil
}

func FetchSigningKeyPairFromArtifactory(t *testing.T, artifactoryURL, apiKey string) bool {
	url := fmt.Sprintf("%sartifactory/api/security/keypair/%s", artifactoryURL, keyPairName)
	t.Logf("Fetching key pair from Artifactory: %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		assert.NoError(t, err, "Failed to close response body")
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to fetch key pair, status: %s, body: %s", resp.Status, string(body))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "failed to read response body")
	var keyPair KeyPair
	err = json.Unmarshal(bodyBytes, &keyPair)
	assert.NoError(t, err)
	assert.Equal(t, keyPairName, keyPair.PairName)
	t.Logf("Successfully fetched and saved key pair: %s", keyPair.PairName)
	if keyPairName == keyPair.PairName {
		return true
	}
	return false
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

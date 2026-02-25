package main

// HuggingFace Integration Tests
// Run with: go test -v -test.huggingface -jfrog.url=http://localhost:8081/ -jfrog.user=admin -jfrog.password=password

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initHuggingFaceTest(t *testing.T) {
	if !*tests.TestHuggingFace {
		t.Skip("Skipping HuggingFace test. To run HuggingFace test add the '-test.huggingface=true' option.")
	}

	if artifactoryCli == nil {
		initArtifactoryCli()
	}

	// Set up home directory configuration so GetDefaultServerConf() can find the server
	createJfrogHomeConfig(t, true)

	// Initialize serverDetails for HuggingFace tests
	serverDetails = &config.ServerDetails{
		Url:            *tests.JfrogUrl,
		ArtifactoryUrl: *tests.JfrogUrl + tests.ArtifactoryEndpoint,
		SshKeyPath:     *tests.JfrogSshKeyPath,
		SshPassphrase:  *tests.JfrogSshPassphrase,
	}
	if *tests.JfrogAccessToken != "" {
		serverDetails.AccessToken = *tests.JfrogAccessToken
	} else {
		serverDetails.User = *tests.JfrogUser
		serverDetails.Password = *tests.JfrogPassword
	}
}

func cleanHuggingFaceTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	tests.CleanFileSystem()
}

// checkHuggingFaceHubAvailable checks if python3 and huggingface_hub library are available
func checkHuggingFaceHubAvailable(t *testing.T) {
	// Check if python3 is available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not found in PATH, skipping HuggingFace test")
	}

	// Check if huggingface_hub library is installed
	cmd := exec.Command("python3", "-c", "import huggingface_hub")
	if err := cmd.Run(); err != nil {
		t.Skip("huggingface_hub library not installed, skipping HuggingFace test. Install with: pip install huggingface_hub")
	}
}

// isExpectedUploadError checks if the error is an expected error for upload without credentials
// Returns true if the error is expected (authentication/authorization related), false otherwise
func isExpectedUploadError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// Expected errors when uploading without proper credentials
	expectedPatterns := []string{
		"401",
		"403",
		"unauthorized",
		"authentication",
		"permission",
		"access denied",
		"forbidden",
		"credentials",
		"token",
		"login",
	}
	for _, pattern := range expectedPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// TestHuggingFaceDownload tests the HuggingFace download command
func TestHuggingFaceDownload(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Test download with a small test model
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test basic download command structure
	// Using a well-known small model for testing
	args := []string{
		"hf", "d", "gpt2",
		"--repo-type=model",
	}

	// Execute and verify success
	err := jfrogCli.Exec(args...)
	assert.NoError(t, err, "HuggingFace download command should succeed")
}

// TestHuggingFaceDownloadWithRevision tests the HuggingFace download command with revision parameter
func TestHuggingFaceDownloadWithRevision(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test download with revision parameter
	args := []string{
		"hf", "d", "gpt2",
		"--repo-type=model",
		"--revision=main",
	}

	err := jfrogCli.Exec(args...)
	assert.NoError(t, err, "HuggingFace download with revision should succeed")
}

// TestHuggingFaceDownloadDataset tests the HuggingFace download command for datasets
func TestHuggingFaceDownloadDataset(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test download dataset
	args := []string{
		"hf", "d", "squad",
		"--repo-type=dataset",
	}

	err := jfrogCli.Exec(args...)
	assert.NoError(t, err, "HuggingFace download dataset should succeed")
}

// TestHuggingFaceDownloadWithEtagTimeout tests the HuggingFace download command with etag-timeout
func TestHuggingFaceDownloadWithEtagTimeout(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test download with etag-timeout parameter
	args := []string{
		"hf", "d", "gpt2",
		"--repo-type=model",
		"--etag-timeout=3600",
	}

	err := jfrogCli.Exec(args...)
	assert.NoError(t, err, "HuggingFace download with etag-timeout should succeed")
}

// TestHuggingFaceUpload tests the HuggingFace upload command
func TestHuggingFaceUpload(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Create a temporary directory with test files to upload
	tempDir, err := os.MkdirTemp("", "hf-upload-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	// Create a test file in the temp directory
	testFile := filepath.Join(tempDir, "test_model.txt")
	err = os.WriteFile(testFile, []byte("test model content"), 0644)
	require.NoError(t, err, "Failed to create test model file")

	// Create a model config file
	configFile := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configFile, []byte(`{"model_type": "test"}`), 0644)
	require.NoError(t, err, "Failed to create config file")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test upload command structure
	args := []string{
		"hf", "u", tempDir, "test-org/test-model",
		"--repo-type=model",
	}

	err = jfrogCli.Exec(args...)
	// Upload should either succeed (with credentials) or fail with an auth error (without credentials)
	if err != nil {
		assert.True(t, isExpectedUploadError(err),
			"Upload failed with unexpected error: %v. Expected either success or authentication-related error", err)
	}
}

// TestHuggingFaceUploadWithRevision tests the HuggingFace upload command with revision parameter
func TestHuggingFaceUploadWithRevision(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Create a temporary directory with test files to upload
	tempDir, err := os.MkdirTemp("", "hf-upload-revision-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	// Create a test file in the temp directory
	testFile := filepath.Join(tempDir, "test_model.txt")
	err = os.WriteFile(testFile, []byte("test model content for revision test"), 0644)
	require.NoError(t, err, "Failed to create test model file")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test upload with revision parameter
	args := []string{
		"hf", "u", tempDir, "test-org/test-model",
		"--repo-type=model",
		"--revision=test-branch",
	}

	err = jfrogCli.Exec(args...)
	// Upload should either succeed (with credentials) or fail with an auth error (without credentials)
	if err != nil {
		assert.True(t, isExpectedUploadError(err),
			"Upload with revision failed with unexpected error: %v. Expected either success or authentication-related error", err)
	}
}

// TestHuggingFaceUploadDataset tests the HuggingFace upload command for datasets
func TestHuggingFaceUploadDataset(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Create a temporary directory with test dataset files
	tempDir, err := os.MkdirTemp("", "hf-upload-dataset-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	// Create test dataset files
	trainFile := filepath.Join(tempDir, "train.json")
	err = os.WriteFile(trainFile, []byte(`[{"text": "sample training data"}]`), 0644)
	require.NoError(t, err, "Failed to create train file")

	testFileData := filepath.Join(tempDir, "test.json")
	err = os.WriteFile(testFileData, []byte(`[{"text": "sample test data"}]`), 0644)
	require.NoError(t, err, "Failed to create test file")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test upload dataset
	args := []string{
		"hf", "u", tempDir, "test-org/test-dataset",
		"--repo-type=dataset",
	}

	err = jfrogCli.Exec(args...)
	// Upload should either succeed (with credentials) or fail with an auth error (without credentials)
	if err != nil {
		assert.True(t, isExpectedUploadError(err),
			"Upload dataset failed with unexpected error: %v. Expected either success or authentication-related error", err)
	}
}

// TestHuggingFaceCommandValidation tests that the HuggingFace command properly validates arguments
func TestHuggingFaceCommandValidation(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test download without model name should fail
	args := []string{
		"hf", "d",
	}
	err := jfrogCli.Exec(args...)
	assert.Error(t, err, "Download without model name should fail")

	// Test upload without folder path and repo-id should fail
	args = []string{
		"hf", "u",
	}
	err = jfrogCli.Exec(args...)
	assert.Error(t, err, "Upload without folder path and repo-id should fail")

	// Test upload with only folder path should fail
	args = []string{
		"hf", "u", "/tmp/test-folder",
	}
	err = jfrogCli.Exec(args...)
	assert.Error(t, err, "Upload with only folder path should fail")

	// Test invalid subcommand should fail
	args = []string{
		"hf", "invalid",
	}
	err = jfrogCli.Exec(args...)
	assert.Error(t, err, "Invalid subcommand should fail")
}

// TestHuggingFaceHelp tests that the HuggingFace help is displayed correctly
func TestHuggingFaceHelp(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test help flag
	args := []string{
		"hf", "--help",
	}
	err := jfrogCli.Exec(args...)
	assert.NoError(t, err, "Help command should not return error")
}

// InitHuggingFaceTests initializes HuggingFace tests
func InitHuggingFaceTests() {
	initArtifactoryCli()
}

// CleanHuggingFaceTests cleans up after HuggingFace tests
func CleanHuggingFaceTests() {
	// Cleanup is handled per-test
}

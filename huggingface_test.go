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
	// Using sshleifer/tiny-gpt2 which is a very small model (~2MB) designed for testing
	args := []string{
		"hf", "d", "sshleifer/tiny-gpt2",
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
	// Using sshleifer/tiny-gpt2 which is a very small model (~2MB) designed for testing
	args := []string{
		"hf", "d", "sshleifer/tiny-gpt2",
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
	// Using stanfordnlp/imdb which is a well-known public dataset
	args := []string{
		"hf", "d", "stanfordnlp/imdb",
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
	// Using sshleifer/tiny-gpt2 which is a very small model (~2MB) designed for testing
	args := []string{
		"hf", "d", "sshleifer/tiny-gpt2",
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

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test help flag
	args := []string{
		"hf", "--help",
	}
	err := jfrogCli.Exec(args...)
	assert.NoError(t, err, "Help command should not return error")
}

// TestHuggingFaceDownloadInvalidRepoID tests that download with invalid repo ID returns appropriate error
func TestHuggingFaceDownloadInvalidRepoID(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test download with non-existent repository ID
	args := []string{
		"hf", "d", "non-existent-org/non-existent-model-12345xyz",
		"--repo-type=model",
	}

	err := jfrogCli.Exec(args...)
	assert.Error(t, err, "Download with invalid repo ID should fail")

	// Verify error message contains relevant information
	if err != nil {
		errStr := strings.ToLower(err.Error())
		hasRelevantError := strings.Contains(errStr, "404") ||
			strings.Contains(errStr, "not found") ||
			strings.Contains(errStr, "does not exist") ||
			strings.Contains(errStr, "repository") ||
			strings.Contains(errStr, "couldn't find")
		assert.True(t, hasRelevantError,
			"Error should indicate repository not found, got: %v", err)
	}
}

// TestHuggingFaceUploadEmptyDirectory tests that upload with empty directory returns appropriate error
func TestHuggingFaceUploadEmptyDirectory(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Create an empty temporary directory
	tempDir, err := os.MkdirTemp("", "hf-upload-empty-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test upload with empty directory
	args := []string{
		"hf", "u", tempDir, "test-org/test-empty-model",
		"--repo-type=model",
	}

	err = jfrogCli.Exec(args...)
	// Empty directory upload behavior depends on huggingface_hub - it may succeed or fail
	// If it fails, it should be with an appropriate error (not a crash)
	if err != nil {
		// Verify it's either an auth error or an empty/no files error
		errStr := strings.ToLower(err.Error())
		isExpected := isExpectedUploadError(err) ||
			strings.Contains(errStr, "empty") ||
			strings.Contains(errStr, "no files") ||
			strings.Contains(errStr, "nothing to upload")
		assert.True(t, isExpected,
			"Upload empty directory failed with unexpected error: %v", err)
	}
}

// TestHuggingFaceUploadNonExistentDirectory tests that upload with non-existent directory returns appropriate error
func TestHuggingFaceUploadNonExistentDirectory(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test upload with non-existent directory
	args := []string{
		"hf", "u", "/non/existent/path/to/model", "test-org/test-model",
		"--repo-type=model",
	}

	err := jfrogCli.Exec(args...)
	assert.Error(t, err, "Upload with non-existent directory should fail")

	// Verify error message indicates path issue
	if err != nil {
		errStr := strings.ToLower(err.Error())
		hasPathError := strings.Contains(errStr, "not found") ||
			strings.Contains(errStr, "no such file") ||
			strings.Contains(errStr, "does not exist") ||
			strings.Contains(errStr, "path") ||
			strings.Contains(errStr, "directory")
		assert.True(t, hasPathError,
			"Error should indicate path not found, got: %v", err)
	}
}

// TestHuggingFaceUploadWithSpecialCharactersInPath tests upload with special characters in folder path
func TestHuggingFaceUploadWithSpecialCharactersInPath(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Create a temporary directory with special characters in name
	baseDir, err := os.MkdirTemp("", "hf-upload-special-*")
	require.NoError(t, err, "Failed to create base temp directory")
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	// Create subdirectory with special characters (spaces and dashes)
	specialDir := filepath.Join(baseDir, "model with spaces-and-dashes")
	err = os.MkdirAll(specialDir, 0755)
	require.NoError(t, err, "Failed to create special character directory")

	// Create test files
	testFile := filepath.Join(specialDir, "config.json")
	err = os.WriteFile(testFile, []byte(`{"model_type": "test-special"}`), 0644)
	require.NoError(t, err, "Failed to create test file")

	modelFile := filepath.Join(specialDir, "model file with spaces.bin")
	err = os.WriteFile(modelFile, []byte("test model binary content"), 0644)
	require.NoError(t, err, "Failed to create model file with spaces")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Test upload with special characters in path
	args := []string{
		"hf", "u", specialDir, "test-org/test-special-chars-model",
		"--repo-type=model",
	}

	err = jfrogCli.Exec(args...)
	// Should either succeed or fail with auth error, not crash due to special characters
	if err != nil {
		assert.True(t, isExpectedUploadError(err),
			"Upload with special characters failed with unexpected error: %v. Expected either success or authentication-related error", err)
	}
}

// TestHuggingFaceUploadOverwrite tests uploading the same model twice to verify overwrite behavior
func TestHuggingFaceUploadOverwrite(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Create a temporary directory with test files
	tempDir, err := os.MkdirTemp("", "hf-upload-overwrite-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	// Create initial model file
	configFile := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configFile, []byte(`{"model_type": "test", "version": 1}`), 0644)
	require.NoError(t, err, "Failed to create config file")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	repoID := "test-org/test-overwrite-model"

	// First upload
	args := []string{
		"hf", "u", tempDir, repoID,
		"--repo-type=model",
	}

	err = jfrogCli.Exec(args...)
	firstUploadErr := err
	if err != nil && !isExpectedUploadError(err) {
		t.Fatalf("First upload failed with unexpected error: %v", err)
	}

	// Update the model file
	err = os.WriteFile(configFile, []byte(`{"model_type": "test", "version": 2}`), 0644)
	require.NoError(t, err, "Failed to update config file")

	// Second upload (overwrite)
	err = jfrogCli.Exec(args...)
	secondUploadErr := err

	// Both uploads should have same behavior (both succeed or both fail with auth)
	if firstUploadErr == nil {
		assert.NoError(t, secondUploadErr, "Second upload (overwrite) should also succeed")
	} else if isExpectedUploadError(firstUploadErr) {
		// If first failed with auth, second should too
		if secondUploadErr != nil {
			assert.True(t, isExpectedUploadError(secondUploadErr),
				"Second upload failed with unexpected error: %v", secondUploadErr)
		}
	}
}

// TestHuggingFaceDownloadWithBuildInfo tests download with build info collection
func TestHuggingFaceDownloadWithBuildInfo(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	buildName := "hf-download-build-test"
	buildNumber := "1"

	// Test download with build info flags
	// Using sshleifer/tiny-gpt2 which is a very small model (~2MB) designed for testing
	args := []string{
		"hf", "d", "sshleifer/tiny-gpt2",
		"--repo-type=model",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}

	err := jfrogCli.Exec(args...)
	assert.NoError(t, err, "HuggingFace download with build info should succeed")

	// Clean up build info
	t.Cleanup(func() {
		// Attempt to clean build info (may fail if not created, which is fine)
		_ = jfrogCli.Exec("rt", "build-discard", buildName, "--max-builds=0")
	})
}

// TestHuggingFaceUploadWithBuildInfo tests upload with build info collection
func TestHuggingFaceUploadWithBuildInfo(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Create a temporary directory with test files
	tempDir, err := os.MkdirTemp("", "hf-upload-buildinfo-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	// Create test files
	configFile := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configFile, []byte(`{"model_type": "test-buildinfo"}`), 0644)
	require.NoError(t, err, "Failed to create config file")

	modelFile := filepath.Join(tempDir, "model.bin")
	err = os.WriteFile(modelFile, []byte("test model content for build info"), 0644)
	require.NoError(t, err, "Failed to create model file")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	buildName := "hf-upload-build-test"
	buildNumber := "1"

	// Test upload with build info flags
	args := []string{
		"hf", "u", tempDir, "test-org/test-buildinfo-model",
		"--repo-type=model",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}

	err = jfrogCli.Exec(args...)
	// Upload should either succeed (with credentials) or fail with auth error
	if err != nil {
		assert.True(t, isExpectedUploadError(err),
			"Upload with build info failed with unexpected error: %v. Expected either success or authentication-related error", err)
	}

	// Clean up build info
	t.Cleanup(func() {
		// Attempt to clean build info (may fail if not created, which is fine)
		_ = jfrogCli.Exec("rt", "build-discard", buildName, "--max-builds=0")
	})
}

// TestHuggingFaceDownloadWithBuildInfoAndModule tests download with build info and module
func TestHuggingFaceDownloadWithBuildInfoAndModule(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	buildName := "hf-download-module-build-test"
	buildNumber := "1"
	moduleName := "tiny-bert-model-module"

	// Test download with build info and module flags
	// Using sshleifer/tiny-gpt2 which is a very small model (~2MB) designed for testing
	args := []string{
		"hf", "d", "sshleifer/tiny-gpt2",
		"--repo-type=model",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
		"--module=" + moduleName,
	}

	err := jfrogCli.Exec(args...)
	assert.NoError(t, err, "HuggingFace download with build info and module should succeed")

	// Clean up build info
	t.Cleanup(func() {
		_ = jfrogCli.Exec("rt", "build-discard", buildName, "--max-builds=0")
	})
}

// TestHuggingFaceUploadWithBuildInfoAndProject tests upload with build info and project
func TestHuggingFaceUploadWithBuildInfoAndProject(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	// Create a temporary directory with test files
	tempDir, err := os.MkdirTemp("", "hf-upload-project-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	// Create test files
	configFile := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configFile, []byte(`{"model_type": "test-project"}`), 0644)
	require.NoError(t, err, "Failed to create config file")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	buildName := "hf-upload-project-build-test"
	buildNumber := "1"
	projectKey := "test-project"

	// Test upload with build info and project flags
	args := []string{
		"hf", "u", tempDir, "test-org/test-project-model",
		"--repo-type=model",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
		"--project=" + projectKey,
	}

	err = jfrogCli.Exec(args...)
	// Upload should either succeed (with credentials) or fail with auth/project error
	if err != nil {
		errStr := strings.ToLower(err.Error())
		isExpected := isExpectedUploadError(err) ||
			strings.Contains(errStr, "project") ||
			strings.Contains(errStr, "not found")
		assert.True(t, isExpected,
			"Upload with project failed with unexpected error: %v", err)
	}

	// Clean up build info
	t.Cleanup(func() {
		_ = jfrogCli.Exec("rt", "build-discard", buildName, "--max-builds=0", "--project="+projectKey)
	})
}

// TestHuggingFaceDownloadDatasetAndVerifyFiles tests downloading a dataset and verifying files exist
func TestHuggingFaceDownloadDatasetAndVerifyFiles(t *testing.T) {
	initHuggingFaceTest(t)
	defer cleanHuggingFaceTest(t)

	// Check if python3 and huggingface_hub are available
	checkHuggingFaceHubAvailable(t)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Download a small, well-known dataset
	args := []string{
		"hf", "d", "stanfordnlp/imdb",
		"--repo-type=dataset",
	}

	err := jfrogCli.Exec(args...)
	if err != nil {
		// Skip verification if download failed (might be network/auth issues)
		t.Skipf("Download failed, skipping file verification: %v", err)
	}

	// Get HuggingFace cache directory
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err, "Failed to get user home directory")

	// HuggingFace typically caches to ~/.cache/huggingface/hub/
	hfCacheDir := filepath.Join(homeDir, ".cache", "huggingface", "hub")

	// Check if cache directory exists
	if _, err := os.Stat(hfCacheDir); os.IsNotExist(err) {
		t.Log("HuggingFace cache directory not found at default location, skipping file verification")
		return
	}

	// Verify some files exist in cache (dataset files are cached with specific naming)
	found := false
	err = filepath.Walk(hfCacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.Contains(path, "imdb") {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	require.NoError(t, err, "Failed to walk cache directory")
	assert.True(t, found, "Downloaded dataset files should exist in HuggingFace cache")
}

// InitHuggingFaceTests initializes HuggingFace tests
func InitHuggingFaceTests() {
	initArtifactoryCli()
}

// CleanHuggingFaceTests cleans up after HuggingFace tests
func CleanHuggingFaceTests() {
	// Cleanup is handled per-test
}

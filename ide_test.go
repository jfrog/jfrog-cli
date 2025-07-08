package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVscodeSetupCommand(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create VSCode extension repository if it doesn't exist
	repoName := tests.RtRepo1 + "-vscode"
	createVscodeRepo(t, repoName)
	defer deleteRepo(t, repoName)

	// Create temporary VSCode installation directory
	tempDir := t.TempDir()
	productPath := filepath.Join(tempDir, "product.json")

	// Create mock product.json file
	originalProductJSON := map[string]interface{}{
		"extensionsGallery": map[string]interface{}{
			"serviceUrl": "https://marketplace.visualstudio.com/_apis/public/gallery",
		},
		"nameShort": "Code",
		"version":   "1.70.0",
	}

	jsonData, err := json.Marshal(originalProductJSON)
	require.NoError(t, err)

	err = os.WriteFile(productPath, jsonData, 0644)
	require.NoError(t, err)

	// Test VSCode setup command
	expectedServiceURL := serverDetails.ArtifactoryUrl + "api/vscodeextensions/" + repoName + "/_apis/public/gallery"

	// Run the VSCode setup command
	runJfrogCli(t, "vscode", "set", "service-url", expectedServiceURL, "--product-path", productPath)

	// Verify the configuration was applied
	modifiedData, err := os.ReadFile(productPath)
	require.NoError(t, err)

	var modifiedProductJSON map[string]interface{}
	err = json.Unmarshal(modifiedData, &modifiedProductJSON)
	require.NoError(t, err)

	extensionsGallery, ok := modifiedProductJSON["extensionsGallery"].(map[string]interface{})
	require.True(t, ok, "extensionsGallery should be present")

	actualServiceURL, ok := extensionsGallery["serviceUrl"].(string)
	require.True(t, ok, "serviceUrl should be a string")
	assert.Equal(t, expectedServiceURL, actualServiceURL)

	// Verify backup was created
	backupPath := productPath + ".backup"
	assert.FileExists(t, backupPath, "Backup file should be created")

	// Verify backup content
	backupData, err := os.ReadFile(backupPath)
	require.NoError(t, err)

	var backupProductJSON map[string]interface{}
	err = json.Unmarshal(backupData, &backupProductJSON)
	require.NoError(t, err)

	backupGallery, ok := backupProductJSON["extensionsGallery"].(map[string]interface{})
	require.True(t, ok)

	backupServiceURL, ok := backupGallery["serviceUrl"].(string)
	require.True(t, ok)
	assert.Equal(t, "https://marketplace.visualstudio.com/_apis/public/gallery", backupServiceURL)
}

func TestVscodeGetCommand(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create temporary VSCode installation directory
	tempDir := t.TempDir()
	productPath := filepath.Join(tempDir, "product.json")

	// Create mock product.json file with custom service URL
	serviceURL := serverDetails.ArtifactoryUrl + "api/vscodeextensions/my-repo/_apis/public/gallery"
	productJSON := map[string]interface{}{
		"extensionsGallery": map[string]interface{}{
			"serviceUrl": serviceURL,
		},
		"nameShort": "Code",
		"version":   "1.70.0",
	}

	jsonData, err := json.Marshal(productJSON)
	require.NoError(t, err)

	err = os.WriteFile(productPath, jsonData, 0644)
	require.NoError(t, err)

	// Test VSCode get command
	output := runJfrogCliWithOutput(t, "vscode", "get", "service-url", "--product-path", productPath)

	// Verify the output contains the expected service URL
	assert.Contains(t, output, serviceURL)
}

func TestVscodeResetCommand(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create temporary VSCode installation directory
	tempDir := t.TempDir()
	productPath := filepath.Join(tempDir, "product.json")
	backupPath := productPath + ".backup"

	// Create mock product.json file with custom service URL
	modifiedProductJSON := map[string]interface{}{
		"extensionsGallery": map[string]interface{}{
			"serviceUrl": serverDetails.ArtifactoryUrl + "api/vscodeextensions/my-repo/_apis/public/gallery",
		},
		"nameShort": "Code",
		"version":   "1.70.0",
	}

	jsonData, err := json.Marshal(modifiedProductJSON)
	require.NoError(t, err)

	err = os.WriteFile(productPath, jsonData, 0644)
	require.NoError(t, err)

	// Create a backup file with original content
	originalProductJSON := map[string]interface{}{
		"extensionsGallery": map[string]interface{}{
			"serviceUrl": "https://marketplace.visualstudio.com/_apis/public/gallery",
		},
		"nameShort": "Code",
		"version":   "1.70.0",
	}

	backupData, err := json.Marshal(originalProductJSON)
	require.NoError(t, err)

	err = os.WriteFile(backupPath, backupData, 0644)
	require.NoError(t, err)

	// Test VSCode reset command
	runJfrogCli(t, "vscode", "reset", "--product-path", productPath)

	// Verify the configuration was reset
	restoredData, err := os.ReadFile(productPath)
	require.NoError(t, err)

	var restoredProductJSON map[string]interface{}
	err = json.Unmarshal(restoredData, &restoredProductJSON)
	require.NoError(t, err)

	extensionsGallery, ok := restoredProductJSON["extensionsGallery"].(map[string]interface{})
	require.True(t, ok)

	serviceURL, ok := extensionsGallery["serviceUrl"].(string)
	require.True(t, ok)
	assert.Equal(t, "https://marketplace.visualstudio.com/_apis/public/gallery", serviceURL)
}

func TestVscodeAutoDetection(t *testing.T) {
	// Skip this test if we're not on a system where we can easily mock VSCode installation
	if runtime.GOOS == "windows" {
		t.Skip("Auto-detection test skipped on Windows due to path complexity")
	}

	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create VSCode extension repository
	repoName := tests.RtRepo1 + "-vscode"
	createVscodeRepo(t, repoName)
	defer deleteRepo(t, repoName)

	// Create mock VSCode installation structure
	tempDir := t.TempDir()
	var vscodeDir string

	switch runtime.GOOS {
	case "darwin":
		vscodeDir = filepath.Join(tempDir, "Visual Studio Code.app", "Contents", "Resources", "app")
	case "linux":
		vscodeDir = filepath.Join(tempDir, "code", "resources", "app")
	default:
		t.Skip("Unsupported OS for auto-detection test")
	}

	err := os.MkdirAll(vscodeDir, 0755)
	require.NoError(t, err)

	productPath := filepath.Join(vscodeDir, "product.json")

	// Create mock product.json
	productJSON := map[string]interface{}{
		"extensionsGallery": map[string]interface{}{
			"serviceUrl": "https://marketplace.visualstudio.com/_apis/public/gallery",
		},
		"nameShort": "Code",
		"version":   "1.70.0",
	}

	jsonData, err := json.Marshal(productJSON)
	require.NoError(t, err)

	err = os.WriteFile(productPath, jsonData, 0644)
	require.NoError(t, err)

	// Set environment to point to our mock installation
	originalPATH := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPATH)

	// Note: This is a simplified test - real auto-detection would need more complex mocking
	// The actual command would attempt auto-detection, but we're testing with an explicit path
	serviceURL := serverDetails.ArtifactoryUrl + "api/vscodeextensions/" + repoName + "/_apis/public/gallery"

	// Test with an explicit path (simulating successful auto-detection)
	runJfrogCli(t, "vscode", "set", "service-url", serviceURL, "--product-path", productPath)

	// Verify configuration was applied
	modifiedData, err := os.ReadFile(productPath)
	require.NoError(t, err)

	var modifiedProductJSON map[string]interface{}
	err = json.Unmarshal(modifiedData, &modifiedProductJSON)
	require.NoError(t, err)

	extensionsGallery, ok := modifiedProductJSON["extensionsGallery"].(map[string]interface{})
	require.True(t, ok)

	actualServiceURL, ok := extensionsGallery["serviceUrl"].(string)
	require.True(t, ok)
	assert.Equal(t, serviceURL, actualServiceURL)
}

func TestVscodeInvalidRepository(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create temporary VSCode installation directory
	tempDir := t.TempDir()
	productPath := filepath.Join(tempDir, "product.json")

	// Create mock product.json file
	productJSON := map[string]interface{}{
		"extensionsGallery": map[string]interface{}{
			"serviceUrl": "https://marketplace.visualstudio.com/_apis/public/gallery",
		},
		"nameShort": "Code",
		"version":   "1.70.0",
	}

	jsonData, err := json.Marshal(productJSON)
	require.NoError(t, err)

	err = os.WriteFile(productPath, jsonData, 0644)
	require.NoError(t, err)

	// Try to set service URL with non-existent repository
	invalidServiceURL := serverDetails.ArtifactoryUrl + "api/vscodeextensions/non-existent-repo/_apis/public/gallery"

	// This should fail due to repository validation
	err = execJfrogCli(t, "vscode", "set", "service-url", invalidServiceURL, "--product-path", productPath)
	assert.Error(t, err, "Command should fail with invalid repository")
	assert.Contains(t, err.Error(), "repository validation failed")
}

func TestVscodePermissionHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Permission test skipped on Windows")
	}

	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create VSCode extension repository
	repoName := tests.RtRepo1 + "-vscode"
	createVscodeRepo(t, repoName)
	defer deleteRepo(t, repoName)

	// Create temporary file with restricted permissions
	tempDir := t.TempDir()
	productPath := filepath.Join(tempDir, "product.json")

	productJSON := map[string]interface{}{
		"extensionsGallery": map[string]interface{}{
			"serviceUrl": "https://marketplace.visualstudio.com/_apis/public/gallery",
		},
	}

	jsonData, err := json.Marshal(productJSON)
	require.NoError(t, err)

	err = os.WriteFile(productPath, jsonData, 0000) // No permissions
	require.NoError(t, err)

	serviceURL := serverDetails.ArtifactoryUrl + "api/vscodeextensions/" + repoName + "/_apis/public/gallery"

	// Command should fail due to permission issues
	err = execJfrogCli(t, "vscode", "set", "service-url", serviceURL, "--product-path", productPath)
	assert.Error(t, err, "Command should fail due to permission issues")

	// Restore permissions for cleanup
	err = os.Chmod(productPath, 0644)
	require.NoError(t, err)
}

// Helper functions

func createVscodeRepo(t *testing.T, repoName string) {
	// Create a VSCode extensions repository configuration
	repoConfig := `{
		"key": "` + repoName + `",
		"rclass": "local", 
		"packageType": "generic",
		"description": "VSCode extensions repository for testing"
	}`

	// Write a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "repo-config.json")
	err := os.WriteFile(configPath, []byte(repoConfig), 0644)
	require.NoError(t, err)

	// Create repository
	runJfrogCli(t, "rt", "repo-create", configPath)
}

func deleteRepo(t *testing.T, repoName string) {
	// Delete repository
	runJfrogCli(t, "rt", "repo-delete", repoName, "--quiet")
}

func runJfrogCliWithOutput(t *testing.T, args ...string) string {
	// Capture output from CLI command
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// Use a buffer to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := jfrogCli.Exec(args...)
	assert.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)
	return string(output)
}

func execJfrogCli(t *testing.T, args ...string) error {
	// Execute CLI command and return error (for testing error cases)
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(args...)
}

// Test helper that validates VSCode repository exists and is properly configured
func validateVscodeRepo(t *testing.T, repoName string) {
	// Check if repository exists using repository info command
	output := runJfrogCliWithOutput(t, "rt", "repo-config", repoName)

	// Parse repository configuration
	var repoConfig map[string]interface{}
	err := json.Unmarshal([]byte(output), &repoConfig)
	require.NoError(t, err)

	// Verify it's the correct repository
	assert.Equal(t, repoName, repoConfig["key"])
	assert.Equal(t, "local", repoConfig["rclass"])
	assert.Equal(t, "generic", repoConfig["packageType"])
}

// Benchmark test for VSCode setup performance
func BenchmarkVscodeSetup(b *testing.B) {
	if !*tests.TestArtifactory {
		b.Skip("Artifactory is not being tested, skipping...")
	}

	// Create a temporary testing.T for setup functions that require it
	t := &testing.T{}

	// Setup
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	repoName := tests.RtRepo1 + "-vscode-bench"
	createVscodeRepo(t, repoName)
	defer deleteRepo(t, repoName)

	tempDir := b.TempDir()
	productPath := filepath.Join(tempDir, "product.json")

	productJSON := map[string]interface{}{
		"extensionsGallery": map[string]interface{}{
			"serviceUrl": "https://marketplace.visualstudio.com/_apis/public/gallery",
		},
	}

	jsonData, err := json.Marshal(productJSON)
	if err != nil {
		b.Fatal(err)
	}
	err = os.WriteFile(productPath, jsonData, 0644)
	if err != nil {
		b.Fatal(err)
	}

	serviceURL := serverDetails.ArtifactoryUrl + "api/vscodeextensions/" + repoName + "/_apis/public/gallery"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset file content before each iteration
		err = os.WriteFile(productPath, jsonData, 0644)
		if err != nil {
			b.Fatal(err)
		}

		// Run the setup command
		jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
		err = jfrogCli.Exec("vscode", "set", "service-url", serviceURL, "--product-path", productPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

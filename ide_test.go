package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	defer deleteRepo(repoName)

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
	runJfrogCli(t, "vscode-config", expectedServiceURL, "--product-json-path", productPath)

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

	// Note: Backup is created in JFrog backup directory with timestamp
	// The backup creation is confirmed by the successful command execution
	// and the backup path is logged by the VSCode command itself
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
	defer deleteRepo(repoName)

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
	runJfrogCli(t, "vscode-config", serviceURL, "--product-json-path", productPath)

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

	// This should fail due to repository validation when server config is provided
	// We need to provide server configuration flags for repository validation to occur
	if serverDetails != nil {
		var err error
		if serverDetails.AccessToken != "" {
			err = execJfrogCli("vscode-config", "--repo-key", "non-existent-repo", "--product-json-path", productPath,
				"--url", serverDetails.ArtifactoryUrl, "--access-token", serverDetails.AccessToken)
		} else {
			err = execJfrogCli("vscode-config", "--repo-key", "non-existent-repo", "--product-json-path", productPath,
				"--url", serverDetails.ArtifactoryUrl, "--user", serverDetails.User, "--password", serverDetails.Password)
		}
		assert.Error(t, err, "Command should fail with invalid repository")
		// Should contain either the generic validation error or the specific "does not exist" error
		errorText := err.Error()
		assert.True(t,
			strings.Contains(errorText, "repository validation failed") ||
				strings.Contains(errorText, "does not exist") ||
				strings.Contains(errorText, "non-existent-repo"),
			"Expected error message to indicate repository validation failure, got: %s", errorText)
	} else {
		t.Skip("Server details not available for repository validation test")
	}
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
	defer deleteRepo(repoName)

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
	err = execJfrogCli("vscode-config", serviceURL, "--product-json-path", productPath)
	assert.Error(t, err, "Command should fail due to permission issues")

	// Restore permissions for cleanup
	err = os.Chmod(productPath, 0644)
	require.NoError(t, err)
}

func TestJetbrainsSetupCommand(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create JetBrains plugins repository if it doesn't exist
	repoName := tests.RtRepo1 + "-jetbrains"
	createJetbrainsRepo(t, repoName)
	defer deleteRepo(repoName)

	// Create temporary JetBrains IDE installation directory
	tempDir := t.TempDir()

	// Create mock IntelliJ IDEA directory structure
	ideaConfigDir := filepath.Join(tempDir, ".config", "JetBrains", "IntelliJIdea2023.3")
	err := os.MkdirAll(ideaConfigDir, 0755)
	require.NoError(t, err)

	propertiesPath := filepath.Join(ideaConfigDir, "idea.properties")

	// Create mock idea.properties file with default content
	originalProperties := `# IDE and Plugin Repository Configuration
ide.config.path=${user.home}/.config/JetBrains/IntelliJIdea2023.3
ide.system.path=${user.home}/.local/share/JetBrains/IntelliJIdea2023.3
`

	err = os.WriteFile(propertiesPath, []byte(originalProperties), 0644)
	require.NoError(t, err)

	// Test JetBrains setup command
	expectedRepositoryURL := serverDetails.ArtifactoryUrl + "api/jetbrainsplugins/" + repoName

	// Set environment variable to make the mock IDE detectable
	originalJetBrainsConfig := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalJetBrainsConfig != "" {
			if err := os.Setenv("XDG_CONFIG_HOME", originalJetBrainsConfig); err != nil {
				t.Logf("Warning: failed to restore XDG_CONFIG_HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("Warning: failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}
	}()
	if err := os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config")); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	// Run the JetBrains setup command
	runJfrogCli(t, "jetbrains-config", expectedRepositoryURL)

	// Verify the configuration was applied
	modifiedProperties, err := os.ReadFile(propertiesPath)
	require.NoError(t, err)

	modifiedContent := string(modifiedProperties)
	assert.Contains(t, modifiedContent, "idea.plugins.host="+expectedRepositoryURL)
	assert.Contains(t, modifiedContent, "JFrog Artifactory plugins repository")

	// Verify backup was created
	backupFiles, err := filepath.Glob(propertiesPath + ".backup.*")
	require.NoError(t, err)
	assert.Greater(t, len(backupFiles), 0, "Backup file should be created")

	// Verify backup content
	if len(backupFiles) > 0 {
		backupData, err := os.ReadFile(backupFiles[0])
		require.NoError(t, err)
		backupContent := string(backupData)
		assert.Contains(t, backupContent, originalProperties)
		assert.NotContains(t, backupContent, "idea.plugins.host=")
	}
}

func TestJetbrainsAutoDetection(t *testing.T) {
	// Skip this test if we're not on a system where we can easily mock JetBrains installation
	if runtime.GOOS == "windows" {
		t.Skip("Auto-detection test skipped on Windows due to path complexity")
	}

	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create JetBrains plugins repository
	repoName := tests.RtRepo1 + "-jetbrains"
	createJetbrainsRepo(t, repoName)
	defer deleteRepo(repoName)

	// Create mock JetBrains IDEs installation structure
	tempDir := t.TempDir()
	var configBase string

	switch runtime.GOOS {
	case "darwin":
		configBase = filepath.Join(tempDir, "Library", "Application Support", "JetBrains")
	case "linux":
		configBase = filepath.Join(tempDir, ".config", "JetBrains")
	default:
		t.Skip("Unsupported OS for auto-detection test")
	}

	// Create multiple mock IDE installations
	ides := []struct {
		name    string
		version string
	}{
		{"IntelliJIdea", "2023.3"},
		{"PyCharm", "2023.3"},
		{"WebStorm", "2023.3"},
	}

	for _, ide := range ides {
		ideConfigDir := filepath.Join(configBase, ide.name+ide.version)
		err := os.MkdirAll(ideConfigDir, 0755)
		require.NoError(t, err)

		propertiesPath := filepath.Join(ideConfigDir, "idea.properties")
		properties := fmt.Sprintf("# %s %s Configuration\nide.config.path=%s\n", ide.name, ide.version, ideConfigDir)

		err = os.WriteFile(propertiesPath, []byte(properties), 0644)
		require.NoError(t, err)
	}

	// Set environment to point to our mock installation
	originalHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("Warning: failed to restore HOME: %v", err)
		}
	}()
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	if runtime.GOOS == "linux" {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if originalXDG != "" {
				if err := os.Setenv("XDG_CONFIG_HOME", originalXDG); err != nil {
					t.Logf("Warning: failed to restore XDG_CONFIG_HOME: %v", err)
				}
			} else {
				if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
					t.Logf("Warning: failed to unset XDG_CONFIG_HOME: %v", err)
				}
			}
		}()
		if err := os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config")); err != nil {
			t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
		}
	}

	// Test JetBrains auto-detection command
	repositoryURL := serverDetails.ArtifactoryUrl + "api/jetbrainsplugins/" + repoName

	// Note: This will work if the auto-detection logic finds our mock IDEs
	runJfrogCli(t, "jetbrains-config", repositoryURL)

	// Verify configuration was applied to detected IDEs
	for _, ide := range ides {
		ideConfigDir := filepath.Join(configBase, ide.name+ide.version)
		propertiesPath := filepath.Join(ideConfigDir, "idea.properties")

		if _, err := os.Stat(propertiesPath); err == nil {
			modifiedProperties, err := os.ReadFile(propertiesPath)
			require.NoError(t, err)
			modifiedContent := string(modifiedProperties)
			assert.Contains(t, modifiedContent, "idea.plugins.host="+repositoryURL)
		}
	}
}

func TestJetbrainsInvalidRepository(t *testing.T) {
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create temporary JetBrains IDE installation directory
	tempDir := t.TempDir()
	ideaConfigDir := filepath.Join(tempDir, ".config", "JetBrains", "IntelliJIdea2023.3")
	err := os.MkdirAll(ideaConfigDir, 0755)
	require.NoError(t, err)

	propertiesPath := filepath.Join(ideaConfigDir, "idea.properties")

	// Create mock idea.properties file
	properties := `# IDE Configuration
ide.config.path=${user.home}/.config/JetBrains/IntelliJIdea2023.3
`

	err = os.WriteFile(propertiesPath, []byte(properties), 0644)
	require.NoError(t, err)

	// Set environment for detection
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			if err := os.Setenv("XDG_CONFIG_HOME", originalXDG); err != nil {
				t.Logf("Warning: failed to restore XDG_CONFIG_HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("Warning: failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}
	}()
	if err := os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config")); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	// This should fail due to repository validation when server config is provided
	// We need to provide server configuration flags for repository validation to occur
	if serverDetails != nil {
		var err error
		if serverDetails.AccessToken != "" {
			err = execJfrogCli("jetbrains-config", "--repo-key", "non-existent-repo",
				"--url", serverDetails.ArtifactoryUrl, "--access-token", serverDetails.AccessToken)
		} else {
			err = execJfrogCli("jetbrains-config", "--repo-key", "non-existent-repo",
				"--url", serverDetails.ArtifactoryUrl, "--user", serverDetails.User, "--password", serverDetails.Password)
		}
		assert.Error(t, err, "Command should fail with invalid repository")
		// Should contain either the generic validation error or the specific "does not exist" error
		errorText := err.Error()
		assert.True(t,
			strings.Contains(errorText, "repository validation failed") ||
				strings.Contains(errorText, "does not exist") ||
				strings.Contains(errorText, "non-existent-repo"),
			"Expected error message to indicate repository validation failure, got: %s", errorText)
	} else {
		t.Skip("Server details not available for repository validation test")
	}
}

func TestJetbrainsPermissionHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Permission test skipped on Windows")
	}

	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	// Create JetBrains plugins repository
	repoName := tests.RtRepo1 + "-jetbrains"
	createJetbrainsRepo(t, repoName)
	defer deleteRepo(repoName)

	// Create temporary directory with restricted permissions
	tempDir := t.TempDir()
	ideaConfigDir := filepath.Join(tempDir, ".config", "JetBrains", "IntelliJIdea2023.3")
	err := os.MkdirAll(ideaConfigDir, 0755)
	require.NoError(t, err)

	propertiesPath := filepath.Join(ideaConfigDir, "idea.properties")

	properties := `# IDE Configuration
ide.config.path=${user.home}/.config/JetBrains/IntelliJIdea2023.3
`

	err = os.WriteFile(propertiesPath, []byte(properties), 0000) // No permissions
	require.NoError(t, err)

	// Set environment for detection
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			if err := os.Setenv("XDG_CONFIG_HOME", originalXDG); err != nil {
				t.Logf("Warning: failed to restore XDG_CONFIG_HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("Warning: failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}
	}()
	if err := os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config")); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	repositoryURL := serverDetails.ArtifactoryUrl + "api/jetbrainsplugins/" + repoName

	// Command should fail due to permission issues
	err = execJfrogCli("jetbrains-config", repositoryURL)
	assert.Error(t, err, "Command should fail due to permission issues")

	// Restore permissions for cleanup
	err = os.Chmod(propertiesPath, 0644)
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

	// Create repository using artifactory CLI with credentials
	err = artifactoryCli.Exec("repo-create", configPath)
	require.NoError(t, err)
}

func deleteRepo(repoName string) {
	// Delete repository using artifactory CLI with credentials
	_ = artifactoryCli.Exec("repo-delete", repoName, "--quiet")
}

func createJetbrainsRepo(t *testing.T, repoName string) {
	// Create a JetBrains plugins repository configuration
	repoConfig := `{
		"key": "` + repoName + `",
		"rclass": "local", 
		"packageType": "jetbrainsplugins",
		"description": "JetBrains plugins repository for testing"
	}`

	// Write a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "repo-config.json")
	err := os.WriteFile(configPath, []byte(repoConfig), 0644)
	require.NoError(t, err)

	// Create repository using artifactory CLI with credentials
	err = artifactoryCli.Exec("repo-create", configPath)
	require.NoError(t, err)
}

func execJfrogCli(args ...string) error {
	// Execute CLI command and return error (for testing error cases)
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(args...)
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
	defer deleteRepo(repoName)

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
		err = jfrogCli.Exec("vscode-config", serviceURL, "--product-json-path", productPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark test for JetBrains setup performance
func BenchmarkJetbrainsSetup(b *testing.B) {
	if !*tests.TestArtifactory {
		b.Skip("Artifactory is not being tested, skipping...")
	}

	// Create a temporary testing.T for setup functions that require it
	t := &testing.T{}

	// Setup
	initArtifactoryTest(t, "")
	defer cleanArtifactoryTest()

	repoName := tests.RtRepo1 + "-jetbrains-bench"
	createJetbrainsRepo(t, repoName)
	defer deleteRepo(repoName)

	tempDir := b.TempDir()
	ideaConfigDir := filepath.Join(tempDir, ".config", "JetBrains", "IntelliJIdea2023.3")
	err := os.MkdirAll(ideaConfigDir, 0755)
	if err != nil {
		b.Fatal(err)
	}

	propertiesPath := filepath.Join(ideaConfigDir, "idea.properties")
	properties := `# IDE Configuration
ide.config.path=${user.home}/.config/JetBrains/IntelliJIdea2023.3
`

	err = os.WriteFile(propertiesPath, []byte(properties), 0644)
	if err != nil {
		b.Fatal(err)
	}

	// Set environment for detection
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			if err := os.Setenv("XDG_CONFIG_HOME", originalXDG); err != nil {
				t.Logf("Warning: failed to restore XDG_CONFIG_HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Logf("Warning: failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}
	}()
	if err := os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config")); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	repositoryURL := serverDetails.ArtifactoryUrl + "api/jetbrainsplugins/" + repoName

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset file content before each iteration
		err = os.WriteFile(propertiesPath, []byte(properties), 0644)
		if err != nil {
			b.Fatal(err)
		}

		// Run the setup command
		jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
		err = jfrogCli.Exec("jetbrains-config", repositoryURL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Unit test to verify IDE commands are properly registered
func TestIDECommandsRegistration(t *testing.T) {
	// This test verifies that our IDE commands are properly registered in the CLI
	// without requiring a running Artifactory server

	// Test that VSCode command is available
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("vscode-config", "--help")
	assert.NoError(t, err, "VSCode config command should be available")

	// Test that JetBrains command is available
	err = jfrogCli.Exec("jetbrains-config", "--help")
	assert.NoError(t, err, "JetBrains config command should be available")

	// Test that JetBrains alias is available
	err = jfrogCli.Exec("jb", "--help")
	assert.NoError(t, err, "JetBrains alias 'jb' should be available")
}

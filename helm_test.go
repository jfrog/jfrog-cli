package main

// Helm Integration Tests
//
// IMPORTANT: These tests require Artifactory to be running and accessible.
// The global test setup (InitBuildToolsTests) will attempt to connect to Artifactory
// and create required repositories. If Artifactory is not accessible, the test setup
// will fail with connection errors.
//
// Requirements:
// 1. Helm CLI installed and available in PATH
// 2. Artifactory instance running and accessible (default: http://localhost:8081/)
// 3. Test flags: -test.helm=true
// 4. Artifactory URL: -jfrog.url=http://localhost:8081/ (or your Artifactory URL)
// 5. Artifactory credentials: -jfrog.user=admin -jfrog.password=password (or use -jfrog.adminToken)
//
// Example command to run tests:
//   go test -v -test.helm=true -jfrog.url=http://localhost:8081/ -jfrog.user=admin -jfrog.password=password ./helm_test.go
//
// Note: If Artifactory is not running, the test setup will fail with connection errors.
// Individual tests include connectivity checks and will skip gracefully, but the global
// setup must complete successfully for tests to run.

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func initHelmTest(t *testing.T) {
	if !*tests.TestHelm {
		t.Skip("Skipping Helm test. To run Helm test add the '-test.helm=true' option.")
	}

	// Check if Helm is available
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("Helm not found in PATH, skipping Helm test")
	}

	// Check Artifactory connectivity first, before any operations that require it
	// This prevents test failures when Artifactory is not running
	// Note: Global setup (InitBuildToolsTests) may have already tried to connect,
	// but individual tests will skip gracefully if Artifactory is not accessible
	if !isArtifactoryAccessible(t) {
		t.Skip("Artifactory is not accessible. Please ensure Artifactory is running and accessible at the configured URL (default: http://localhost:8081/).")
	}

	// At this point, Artifactory should be accessible
	// The global setup (InitBuildToolsTests) should have already initialized everything
	// But we verify connectivity here to ensure the test can proceed
}

// isArtifactoryAccessible checks if Artifactory is accessible by attempting a simple API call
func isArtifactoryAccessible(t *testing.T) bool {
	// Try to ping Artifactory - if this fails, we'll skip the test
	// This is a best-effort check to avoid failing tests when Artifactory isn't running
	// We use a simple HTTP check instead of the CLI to avoid circular dependencies
	artifactoryUrl := *tests.JfrogUrl
	if artifactoryUrl == "" {
		return false
	}

	// Ensure URL has trailing slash and add Artifactory endpoint
	if !strings.HasSuffix(artifactoryUrl, "/") {
		artifactoryUrl += "/"
	}
	artifactoryUrl += "artifactory/api/system/ping"

	// Try a simple GET request to Artifactory's ping endpoint
	// This is a lightweight check that doesn't require authentication
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(artifactoryUrl)
	if err != nil {
		return false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log but don't fail on close errors
			t.Logf("Warning: Failed to close response body: %v", closeErr)
		}
	}()
	return resp.StatusCode == http.StatusOK
}

func cleanHelmTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	deleteFilesFromRepo(t, tests.HelmLocalRepo)
	tests.CleanFileSystem()
}

// TestHelmPushWithBuildInfo tests helm push command with build info collection
func TestHelmPushWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	// Check if Helm is available
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("Helm not found in PATH, skipping Helm test")
	}

	buildName := tests.HelmBuildName + "-push"
	buildNumber := "1"

	// Create a test Helm chart
	chartDir := createTestHelmChart(t, "test-chart", "0.1.0")
	defer func() {
		if err := os.RemoveAll(chartDir); err != nil {
			t.Logf("Warning: Failed to remove test chart directory %s: %v", chartDir, err)
		}
	}()

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	err = os.Chdir(chartDir)
	require.NoError(t, err)

	// First, package the chart
	helmCmd := exec.Command("helm", "package", ".")
	helmCmd.Dir = chartDir
	err = helmCmd.Run()
	require.NoError(t, err, "helm package should succeed")

	// Find the packaged chart file
	chartFiles, err := filepath.Glob(filepath.Join(chartDir, "*.tgz"))
	require.NoError(t, err)
	require.Greater(t, len(chartFiles), 0, "Chart package file should be created")
	chartFile := filepath.Base(chartFiles[0])

	// Build OCI registry URL - Helm requires oci:// scheme
	// Format: oci://<host>/artifactory/<repo>
	parsedURL, err := url.Parse(serverDetails.ArtifactoryUrl)
	require.NoError(t, err)
	registryHost := parsedURL.Host
	registryURL := fmt.Sprintf("oci://%s/artifactory/%s", registryHost, tests.HelmLocalRepo)

	// Login to Helm OCI registry before pushing
	// If login fails due to account lockout, skip the test with a helpful message
	err = loginHelmRegistry(t, registryHost)
	if err != nil && strings.Contains(err.Error(), "account temporarily locked") {
		t.Skip("Artifactory account is temporarily locked due to recurrent login failures. Please wait and try again, or verify credentials are correct.")
	}
	require.NoError(t, err, "helm registry login should succeed")

	// Run helm push with build info
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "push", chartFile,
		registryURL,
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm push should succeed")

	// Publish build info
	assert.NoError(t, runJfrogCliWithoutAssertion("rt", "bp", buildName, buildNumber))

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	// Validate build info structure
	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, true, false)
}

// TestHelmPackageWithBuildInfo tests helm package command with build info collection
func TestHelmPackageWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	// Check if Helm is available
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("Helm not found in PATH, skipping Helm test")
	}

	buildName := tests.HelmBuildName + "-package"
	buildNumber := "1"

	// Create a test Helm chart with dependencies
	chartDir := createTestHelmChartWithDependencies(t, "test-chart", "0.1.0")
	defer func() {
		if err := os.RemoveAll(chartDir); err != nil {
			t.Logf("Warning: Failed to remove test chart directory %s: %v", chartDir, err)
		}
	}()

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	err = os.Chdir(chartDir)
	require.NoError(t, err)

	// First, update dependencies to download them to charts/ directory
	helmCmd := exec.Command("helm", "dependency", "update")
	helmCmd.Dir = chartDir
	err = helmCmd.Run()
	require.NoError(t, err, "helm dependency update should succeed")

	// Run helm package with build info
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "package", ".",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm package should succeed")

	// Publish build info
	assert.NoError(t, runJfrogCliWithoutAssertion("rt", "bp", buildName, buildNumber))

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	// Validate build info structure - package should include dependencies
	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, true, true)
}

// TestHelmDependencyUpdateWithBuildInfo tests helm dependency update command with build info
func TestHelmDependencyUpdateWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	// Check if Helm is available
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("Helm not found in PATH, skipping Helm test")
	}

	buildName := tests.HelmBuildName + "-dep-update"
	buildNumber := "1"

	// Create a test Helm chart with dependencies
	chartDir := createTestHelmChartWithDependencies(t, "test-chart", "0.1.0")
	defer func() {
		if err := os.RemoveAll(chartDir); err != nil {
			t.Logf("Warning: Failed to remove test chart directory %s: %v", chartDir, err)
		}
	}()

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	err = os.Chdir(chartDir)
	require.NoError(t, err)

	// Run helm dependency update with build info
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "dependency", "update",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm dependency update should succeed")

	// Publish build info
	assert.NoError(t, runJfrogCliWithoutAssertion("rt", "bp", buildName, buildNumber))

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	// Validate build info structure - dependency update should include dependencies
	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, false, true)
}

// TestHelmInstallWithBuildInfo tests helm install command with build info collection
func TestHelmInstallWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	// Check if Helm is available
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("Helm not found in PATH, skipping Helm test")
	}

	buildName := tests.HelmBuildName + "-install"
	buildNumber := "1"

	// Create a test Helm chart
	chartDir := createTestHelmChart(t, "test-chart", "0.1.0")
	defer func() {
		if err := os.RemoveAll(chartDir); err != nil {
			t.Logf("Warning: Failed to remove test chart directory %s: %v", chartDir, err)
		}
	}()

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	err = os.Chdir(chartDir)
	require.NoError(t, err)

	// Run helm install with build info (dry-run to avoid requiring Kubernetes)
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "install", "test-release", ".",
		"--dry-run",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm install should succeed")

	// Publish build info
	assert.NoError(t, runJfrogCliWithoutAssertion("rt", "bp", buildName, buildNumber))

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	// Validate build info structure - install should include dependencies
	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, false, true)
}

// TestHelmTemplateWithBuildInfo tests helm template command with build info collection
func TestHelmTemplateWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	// Check if Helm is available
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("Helm not found in PATH, skipping Helm test")
	}

	buildName := tests.HelmBuildName + "-template"
	buildNumber := "1"

	// Create a test Helm chart
	chartDir := createTestHelmChart(t, "test-chart", "0.1.0")
	defer func() {
		if err := os.RemoveAll(chartDir); err != nil {
			t.Logf("Warning: Failed to remove test chart directory %s: %v", chartDir, err)
		}
	}()

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	err = os.Chdir(chartDir)
	require.NoError(t, err)

	// Run helm template with build info
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "template", "test-release", ".",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm template should succeed")

	// Publish build info
	assert.NoError(t, runJfrogCliWithoutAssertion("rt", "bp", buildName, buildNumber))

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	// Validate build info structure
	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, false, true)
}

// loginHelmRegistry logs into the Helm OCI registry using credentials from serverDetails
func loginHelmRegistry(t *testing.T, registryHost string) error {
	// Get credentials from serverDetails
	user := serverDetails.User
	pass := serverDetails.Password
	if serverDetails.AccessToken != "" {
		// If access token is provided, use it as password
		pass = serverDetails.AccessToken
		if user == "" {
			// Extract username from token if not provided
			// For simplicity, we'll use a default or extract from token
			user = "admin" // Default fallback, or extract from token
		}
	}

	if user == "" || pass == "" {
		return fmt.Errorf("credentials required for Helm registry login")
	}

	// Run helm registry login
	cmd := exec.Command("helm", "registry", "login", registryHost, "--username", user, "--password-stdin")
	cmd.Stdin = strings.NewReader(pass)
	// Capture output to check for specific errors
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		// Check if the error is due to account lockout
		errorOutput := stderr.String()
		if strings.Contains(errorOutput, "recurrent login failures") || strings.Contains(errorOutput, "blocked") {
			t.Logf("Helm registry login failed due to account lockout. Please wait and try again, or verify credentials are correct.")
			return fmt.Errorf("account temporarily locked: %w", err)
		}
		// For other errors, return as-is
		return fmt.Errorf("helm registry login failed: %w (stderr: %s)", err, errorOutput)
	}

	return nil
}

// Helper function to create a test Helm chart
func createTestHelmChart(t *testing.T, name, version string) string {
	tempDir, err := os.MkdirTemp("", "helm-test-*")
	require.NoError(t, err)

	chartDir := filepath.Join(tempDir, name)
	err = os.MkdirAll(chartDir, 0755)
	require.NoError(t, err)

	// Create Chart.yaml
	chartYaml := fmt.Sprintf(`apiVersion: v2
name: %s
description: A Helm chart for testing
type: application
version: %s
appVersion: "1.0.0"
`, name, version)

	err = os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte(chartYaml), 0644)
	require.NoError(t, err)

	// Create values.yaml
	valuesYaml := `replicaCount: 1
image:
  repository: nginx
  tag: "1.21"
`
	err = os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte(valuesYaml), 0644)
	require.NoError(t, err)

	// Create templates directory
	templatesDir := filepath.Join(chartDir, "templates")
	err = os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Create a simple deployment template
	deploymentYaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: {{ .Release.Name }}
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}
    spec:
      containers:
      - name: {{ .Release.Name }}
        image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
`
	err = os.WriteFile(filepath.Join(templatesDir, "deployment.yaml"), []byte(deploymentYaml), 0644)
	require.NoError(t, err)

	return chartDir
}

// Helper function to create a test Helm chart with dependencies
func createTestHelmChartWithDependencies(t *testing.T, name, version string) string {
	chartDir := createTestHelmChart(t, name, version)

	// Add dependencies to Chart.yaml
	chartYamlPath := filepath.Join(chartDir, "Chart.yaml")
	chartYamlContent, err := os.ReadFile(chartYamlPath)
	require.NoError(t, err)

	var chartData map[string]interface{}
	err = yaml.Unmarshal(chartYamlContent, &chartData)
	require.NoError(t, err)

	// Add dependencies (using common Helm charts for testing)
	chartData["dependencies"] = []map[string]interface{}{
		{
			"name":       "postgresql",
			"version":    "14.x.x",
			"repository": "https://charts.bitnami.com/bitnami",
			"condition":  "postgresql.enabled",
		},
		{
			"name":       "redis",
			"version":    "18.x.x",
			"repository": "https://charts.bitnami.com/bitnami",
			"condition":  "redis.enabled",
		},
	}

	updatedChartYaml, err := yaml.Marshal(chartData)
	require.NoError(t, err)

	err = os.WriteFile(chartYamlPath, updatedChartYaml, 0644)
	require.NoError(t, err)

	return chartDir
}

// Helper function to validate Helm build info structure
func validateHelmBuildInfo(t *testing.T, buildInfo buildinfo.BuildInfo, buildName string, expectArtifacts, expectDependencies bool) {
	assert.Equal(t, buildName, buildInfo.Name, "Build name should match")
	assert.Equal(t, "1", buildInfo.Number, "Build number should match")
	assert.NotNil(t, buildInfo.Agent, "Build info should have agent")
	assert.NotNil(t, buildInfo.BuildAgent, "Build info should have build agent")
	assert.NotEmpty(t, buildInfo.Started, "Build info should have start time")

	require.Greater(t, len(buildInfo.Modules), 0, "Build info should have at least one module")

	module := buildInfo.Modules[0]
	assert.Equal(t, buildinfo.Helm, module.Type, "Module type should be helm")
	assert.NotEmpty(t, module.Id, "Module should have ID")

	if expectArtifacts {
		assert.Greater(t, len(module.Artifacts), 0, "Module should have artifacts for push/package commands")
		for _, artifact := range module.Artifacts {
			assert.NotEmpty(t, artifact.Name, "Artifact should have name")
			assert.NotEmpty(t, artifact.Sha256, "Artifact should have SHA256 checksum")
			if strings.Contains(artifact.Name, "manifest.json") || strings.Contains(artifact.Name, "config") {
				assert.NotEmpty(t, artifact.Path, "OCI artifact should have path")
			}
		}
	} else {
		assert.LessOrEqual(t, len(module.Artifacts), 0, "Module should not have artifacts for install/template commands")
	}

	if expectDependencies {
		assert.Greater(t, len(module.Dependencies), 0, "Module should have dependencies")
		for _, dep := range module.Dependencies {
			assert.NotEmpty(t, dep.Id, "Dependency should have ID")
			hasChecksum := dep.Sha1 != "" || dep.Sha256 != "" || dep.Md5 != ""
			assert.True(t, hasChecksum, "Dependency %s should have at least one checksum", dep.Id)
			assert.NotContains(t, dep.Id, "x.x", "Dependency ID should not contain version ranges")
		}
	} else {
		assert.Equal(t, 0, len(module.Dependencies), "Module should not have dependencies for push command")
	}

	if expectArtifacts && len(module.Artifacts) > 0 {
		for _, artifact := range module.Artifacts {
			assert.NotEmpty(t, artifact.Name, "Artifact should have name")
		}
	}
}


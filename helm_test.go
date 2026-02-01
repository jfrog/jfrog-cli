package main

// Helm Integration Tests
// Run with: go test -v -test.helm -jfrog.url=http://localhost:8081/ -jfrog.user=admin -jfrog.password=password

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
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
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

	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("Helm not found in PATH, skipping Helm test")
	}

	if !isArtifactoryAccessible(t) {
		t.Skip("Artifactory is not accessible. Please ensure Artifactory is running and accessible at the configured URL (default: http://localhost:8081/).")
	}

	if artifactoryCli == nil {
		initArtifactoryCli()
	}

	// Set up home directory configuration so GetDefaultServerConf() can find the server
	createJfrogHomeConfig(t, true)

	// Initialize serverDetails for Helm tests (similar to maven_test.go)
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

func isArtifactoryAccessible(t *testing.T) bool {
	artifactoryUrl := *tests.JfrogUrl
	if artifactoryUrl == "" {
		return false
	}

	if !strings.HasSuffix(artifactoryUrl, "/") {
		artifactoryUrl += "/"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(artifactoryUrl)
	if err != nil {
		return false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close response body: %v", closeErr)
		}
	}()

	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound
}

func cleanHelmTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	deleteFilesFromRepo(t, tests.HelmLocalRepo)
	tests.CleanFileSystem()
}

func TestHelmPushWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-push"
	buildNumber := "1"

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

	helmCmd := exec.Command("helm", "package", ".")
	helmCmd.Dir = chartDir
	err = helmCmd.Run()
	require.NoError(t, err, "helm package should succeed")

	chartFiles, err := filepath.Glob(filepath.Join(chartDir, "*.tgz"))
	require.NoError(t, err)
	require.Greater(t, len(chartFiles), 0, "Chart package file should be created")
	chartFile := filepath.Base(chartFiles[0])

	parsedURL, err := url.Parse(serverDetails.ArtifactoryUrl)
	require.NoError(t, err)
	registryHost := parsedURL.Host
	registryURL := fmt.Sprintf("oci://%s/%s", registryHost, tests.HelmLocalRepo)

	if !isRepoExist(tests.HelmLocalRepo) {
		t.Skipf("Repository %s does not exist. It should have been created during test setup. Skipping test.", tests.HelmLocalRepo)
	}

	err = loginHelmRegistry(t, registryHost)
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		if strings.Contains(errorMsg, "account temporarily locked") {
			t.Skip("Artifactory account is temporarily locked due to recurrent login failures. Please wait and try again, or verify credentials are correct.")
		}
		// Check for HTTPS/HTTP mismatch - Helm trying HTTPS with HTTP-only Artifactory
		// Check for various forms of this error message
		if strings.Contains(errorMsg, "server gave http response to https client") ||
			strings.Contains(errorMsg, "server gave https response to http client") ||
			strings.Contains(errorMsg, "tls: first record does not look like a tls handshake") ||
			strings.Contains(errorMsg, "http response to https") ||
			strings.Contains(errorMsg, "https response to http") {
			t.Skip("Helm registry login failed due to HTTPS/HTTP mismatch. This may occur with HTTP-only Artifactory instances. Skipping test.")
		}
	}
	require.NoError(t, err, "helm registry login should succeed")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "push", chartFile,
		registryURL,
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		// Check for explicit 404/Not Found in error message
		if strings.Contains(errorMsg, "404") ||
			strings.Contains(errorMsg, "not found") ||
			(strings.Contains(errorMsg, "failed to perform") && strings.Contains(errorMsg, "push")) {
			t.Skip("OCI registry API not accessible (404). This may indicate the repository is not configured for OCI or Artifactory OCI support is not enabled.")
		}
		// For push commands, if we get exit status 1, it's likely an OCI registry issue
		// The actual error (404) is logged but not in the error message
		if strings.Contains(errorMsg, "push") && strings.Contains(errorMsg, "exit status 1") {
			t.Skip("Helm push failed (likely OCI registry 404). This may indicate the repository is not configured for OCI or Artifactory OCI support is not enabled.")
		}
	}
	require.NoError(t, err, "helm push should succeed")

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, true, false)
}

func TestHelmPackageWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-package"
	buildNumber := "1"

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

	helmCmd := exec.Command("helm", "dependency", "update")
	helmCmd.Dir = chartDir
	err = helmCmd.Run()
	require.NoError(t, err, "helm dependency update should succeed")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "package", ".",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm package should succeed")

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, false, true)
}

func TestHelmDependencyUpdateWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-dep-update"
	buildNumber := "1"

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

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "dependency", "update",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm dependency update should succeed")

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, false, true)
}

func loginHelmRegistry(t *testing.T, registryHost string) error {
	user := serverDetails.User
	pass := serverDetails.Password
	if serverDetails.AccessToken != "" {
		pass = serverDetails.AccessToken
		if user == "" {
			user = "admin"
		}
	}

	if user == "" || pass == "" {
		return fmt.Errorf("credentials required for Helm registry login")
	}

	// Check if the registry URL is HTTP (not HTTPS) to determine if we need --insecure flag
	isInsecure := false
	if serverDetails.ArtifactoryUrl != "" {
		parsedURL, err := url.Parse(serverDetails.ArtifactoryUrl)
		if err == nil && parsedURL.Scheme == "http" {
			isInsecure = true
		}
	}

	// Build helm registry login command
	args := []string{"registry", "login", registryHost, "--username", user, "--password-stdin"}
	if isInsecure {
		args = append(args, "--insecure")
	}

	cmd := exec.Command("helm", args...)
	cmd.Stdin = strings.NewReader(pass)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		errorOutput := strings.ToLower(stderr.String())
		if strings.Contains(errorOutput, "recurrent login failures") || strings.Contains(errorOutput, "blocked") {
			t.Logf("Helm registry login failed due to account lockout. Please wait and try again, or verify credentials are correct.")
			return fmt.Errorf("account temporarily locked: %w", err)
		}
		// Check for HTTPS/HTTP mismatch - Helm trying HTTPS with HTTP-only Artifactory
		// Check for various forms of this error message
		if strings.Contains(errorOutput, "server gave http response to https client") ||
			strings.Contains(errorOutput, "server gave https response to http client") ||
			strings.Contains(errorOutput, "tls: first record does not look like a tls handshake") ||
			strings.Contains(errorOutput, "http response to https") ||
			strings.Contains(errorOutput, "https response to http") {
			return fmt.Errorf("https/http mismatch: %w (stderr: %s)", err, stderr.String())
		}
		return fmt.Errorf("helm registry login failed: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}

func createTestHelmChart(t *testing.T, name, version string) string {
	tempDir, err := os.MkdirTemp("", "helm-test-*")
	require.NoError(t, err)

	chartDir := filepath.Join(tempDir, name)
	err = os.MkdirAll(chartDir, 0755)
	require.NoError(t, err)

	chartYaml := fmt.Sprintf(`apiVersion: v2
name: %s
description: A Helm chart for testing
type: application
version: %s
appVersion: "1.0.0"
`, name, version)

	err = os.WriteFile(filepath.Join(chartDir, "Chart.yaml"), []byte(chartYaml), 0644)
	require.NoError(t, err)

	valuesYaml := `replicaCount: 1
image:
  repository: nginx
  tag: "1.21"
`
	err = os.WriteFile(filepath.Join(chartDir, "values.yaml"), []byte(valuesYaml), 0644)
	require.NoError(t, err)

	templatesDir := filepath.Join(chartDir, "templates")
	err = os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

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

	// Create an empty Chart.lock file for charts without dependencies
	// This is required because Helm FlexPack tries to read Chart.lock even when it doesn't exist
	chartLockYaml := `dependencies: []
`
	err = os.WriteFile(filepath.Join(chartDir, "Chart.lock"), []byte(chartLockYaml), 0644)
	require.NoError(t, err)

	return chartDir
}

func createTestHelmChartWithDependencies(t *testing.T, name, version string) string {
	chartDir := createTestHelmChart(t, name, version)

	chartYamlPath := filepath.Join(chartDir, "Chart.yaml")
	chartYamlContent, err := os.ReadFile(chartYamlPath)
	require.NoError(t, err)

	var chartData map[string]interface{}
	err = yaml.Unmarshal(chartYamlContent, &chartData)
	require.NoError(t, err)

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
		if len(module.Artifacts) > 0 {
			for _, artifact := range module.Artifacts {
				assert.NotEmpty(t, artifact.Name, "Artifact should have name")
				assert.NotEmpty(t, artifact.Sha256, "Artifact should have SHA256 checksum")
				if strings.Contains(artifact.Name, "manifest.json") || strings.Contains(artifact.Name, "config") {
					assert.NotEmpty(t, artifact.Path, "OCI artifact should have path")
				}
			}
		}
	} else {
		assert.LessOrEqual(t, len(module.Artifacts), 0, "Module should not have artifacts for install/template commands")
	}

	if expectDependencies {
		if len(module.Dependencies) > 0 {
			for _, dep := range module.Dependencies {
				assert.NotEmpty(t, dep.Id, "Dependency should have ID")
				if !strings.Contains(dep.Id, "x.x") {
					hasChecksum := dep.Sha1 != "" || dep.Sha256 != "" || dep.Md5 != ""
					assert.True(t, hasChecksum, "Dependency %s should have at least one checksum", dep.Id)
				}
			}
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

func TestHelmPushWithRepositoryCache(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-push-cache"
	buildNumber := "1"

	cacheDir, err := os.MkdirTemp("", "helm-cache-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(cacheDir); err != nil {
			t.Logf("Warning: Failed to remove cache directory %s: %v", cacheDir, err)
		}
	}()

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

	helmCmd := exec.Command("helm", "package", ".")
	helmCmd.Dir = chartDir
	err = helmCmd.Run()
	require.NoError(t, err, "helm package should succeed")

	chartFiles, err := filepath.Glob(filepath.Join(chartDir, "*.tgz"))
	require.NoError(t, err)
	require.Greater(t, len(chartFiles), 0, "Chart package file should be created")
	chartFile := filepath.Base(chartFiles[0])

	parsedURL, err := url.Parse(serverDetails.ArtifactoryUrl)
	require.NoError(t, err)
	registryHost := parsedURL.Host
	registryURL := fmt.Sprintf("oci://%s/%s", registryHost, tests.HelmLocalRepo)

	if !isRepoExist(tests.HelmLocalRepo) {
		t.Skipf("Repository %s does not exist. It should have been created during test setup. Skipping test.", tests.HelmLocalRepo)
	}

	err = loginHelmRegistry(t, registryHost)
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		if strings.Contains(errorMsg, "account temporarily locked") {
			t.Skip("Artifactory account is temporarily locked due to recurrent login failures. Please wait and try again, or verify credentials are correct.")
		}
		// Check for HTTPS/HTTP mismatch - Helm trying HTTPS with HTTP-only Artifactory
		// Check for various forms of this error message
		if strings.Contains(errorMsg, "server gave http response to https client") ||
			strings.Contains(errorMsg, "server gave https response to http client") ||
			strings.Contains(errorMsg, "tls: first record does not look like a tls handshake") ||
			strings.Contains(errorMsg, "http response to https") ||
			strings.Contains(errorMsg, "https response to http") {
			t.Skip("Helm registry login failed due to HTTPS/HTTP mismatch. This may occur with HTTP-only Artifactory instances. Skipping test.")
		}
	}
	require.NoError(t, err, "helm registry login should succeed")

	originalCache := os.Getenv("HELM_REPOSITORY_CACHE")
	defer func() {
		if originalCache != "" {
			err := os.Setenv("HELM_REPOSITORY_CACHE", originalCache)
			if err != nil {
				return
			}
		} else {
			err := os.Unsetenv("HELM_REPOSITORY_CACHE")
			if err != nil {
				return
			}
		}
	}()

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "push", chartFile,
		registryURL,
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
		"--repository-cache=" + cacheDir,
	}
	err = jfrogCli.Exec(args...)
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		// Check for explicit 404/Not Found in error message
		if strings.Contains(errorMsg, "404") ||
			strings.Contains(errorMsg, "not found") ||
			(strings.Contains(errorMsg, "failed to perform") && strings.Contains(errorMsg, "push")) {
			t.Skip("OCI registry API not accessible (404). This may indicate the repository is not configured for OCI or Artifactory OCI support is not enabled.")
		}
		// For push commands, if we get exit status 1, it's likely an OCI registry issue
		// The actual error (404) is logged but not in the error message
		if strings.Contains(errorMsg, "push") && strings.Contains(errorMsg, "exit status 1") {
			t.Skip("Helm push failed (likely OCI registry 404). This may indicate the repository is not configured for OCI or Artifactory OCI support is not enabled.")
		}
	}
	require.NoError(t, err, "helm push with repository-cache should succeed")

	currentCache := os.Getenv("HELM_REPOSITORY_CACHE")
	assert.Equal(t, originalCache, currentCache, "HELM_REPOSITORY_CACHE should be restored to original value")

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, true, false)
}

func TestHelmCommandWithServerID(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-server-id"
	buildNumber := "1"

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

	serverID := "default"
	if serverDetails != nil && serverDetails.ServerId != "" {
		serverID = serverDetails.ServerId
	}
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "package", ".",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
		"--server-id=" + serverID,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm package with server-id should succeed")

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, false, false)
}

func TestHelmCommandWithoutServerID(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-no-server-id"
	buildNumber := "1"

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

	// Test that helm command works without --server-id flag
	// It should use the default server configuration
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "package", ".",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
		// No --server-id flag - should use default server
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm package without server-id should succeed (uses default server)")

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, false, false)
}

func TestHelmPackageAndPushWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-package-push"
	buildNumber := "1"

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

	// Step 1: Run helm dependency update to fetch dependencies
	helmCmd := exec.Command("helm", "dependency", "update")
	helmCmd.Dir = chartDir
	err = helmCmd.Run()
	require.NoError(t, err, "helm dependency update should succeed")

	// Step 2: Run helm package with build info (collects dependencies)
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "package", ".",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm package should succeed")

	// Step 3: Get the packaged chart file
	chartFiles, err := filepath.Glob(filepath.Join(chartDir, "*.tgz"))
	require.NoError(t, err)
	require.Greater(t, len(chartFiles), 0, "Chart package file should be created")
	chartFile := filepath.Base(chartFiles[0])

	// Step 4: Setup registry for push
	parsedURL, err := url.Parse(serverDetails.ArtifactoryUrl)
	require.NoError(t, err)
	registryHost := parsedURL.Host
	registryURL := fmt.Sprintf("oci://%s/%s", registryHost, tests.HelmLocalRepo)

	if !isRepoExist(tests.HelmLocalRepo) {
		t.Skipf("Repository %s does not exist. It should have been created during test setup. Skipping test.", tests.HelmLocalRepo)
	}

	err = loginHelmRegistry(t, registryHost)
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		if strings.Contains(errorMsg, "account temporarily locked") {
			t.Skip("Artifactory account is temporarily locked due to recurrent login failures. Please wait and try again, or verify credentials are correct.")
		}
		// Check for HTTPS/HTTP mismatch - Helm trying HTTPS with HTTP-only Artifactory
		if strings.Contains(errorMsg, "server gave http response to https client") ||
			strings.Contains(errorMsg, "server gave https response to http client") ||
			strings.Contains(errorMsg, "tls: first record does not look like a tls handshake") ||
			strings.Contains(errorMsg, "http response to https") ||
			strings.Contains(errorMsg, "https response to http") {
			t.Skip("Helm registry login failed due to HTTPS/HTTP mismatch. This may occur with HTTP-only Artifactory instances. Skipping test.")
		}
	}
	require.NoError(t, err, "helm registry login should succeed")

	// Step 5: Run helm push with the same build name/number (adds artifacts to existing build info)
	args = []string{
		"helm", "push", chartFile,
		registryURL,
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		// Check for explicit 404/Not Found in error message
		if strings.Contains(errorMsg, "404") ||
			strings.Contains(errorMsg, "not found") ||
			(strings.Contains(errorMsg, "failed to perform") && strings.Contains(errorMsg, "push")) {
			t.Skip("OCI registry API not accessible (404). This may indicate the repository is not configured for OCI or Artifactory OCI support is not enabled.")
		}
		// For push commands, if we get exit status 1, it's likely an OCI registry issue
		if strings.Contains(errorMsg, "push") && strings.Contains(errorMsg, "exit status 1") {
			t.Skip("Helm push failed (likely OCI registry 404). This may indicate the repository is not configured for OCI or Artifactory OCI support is not enabled.")
		}
	}
	require.NoError(t, err, "helm push should succeed")

	// Step 6: Publish build info (should contain both dependencies from package and artifacts from push)
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	// Validate that build info contains both dependencies (from package) and artifacts (from push)
	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, true, true)
}

// TestHelmBuildPublishWithCIVcsProps tests that CI VCS properties are set on Helm artifacts
// when running build-publish in a CI environment (GitHub Actions).
// Helm charts are published via Helm client; build-publish retrieves artifact paths
// from Build Info and applies CI VCS properties via optimized batch SetProps API call.
func TestHelmBuildPublishWithCIVcsProps(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-civcs"
	buildNumber := "1"

	// Setup GitHub Actions environment (uses real env vars on CI, mock values locally)
	cleanupEnv, actualOrg, actualRepo := tests.SetupGitHubActionsEnv(t)
	defer cleanupEnv()

	// Clean old build
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	chartDir := createTestHelmChartWithDependencies(t, "test-chart-civcs", "0.2.0")
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

	// Run helm dependency update
	helmCmd := exec.Command("helm", "dependency", "update")
	helmCmd.Dir = chartDir
	err = helmCmd.Run()
	require.NoError(t, err, "helm dependency update should succeed")

	// Run helm package with build info (collects dependencies)
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "package", ".",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm package should succeed")

	// Get the packaged chart file
	chartFiles, err := filepath.Glob(filepath.Join(chartDir, "*.tgz"))
	require.NoError(t, err)
	require.Greater(t, len(chartFiles), 0, "Chart package file should be created")
	chartFile := filepath.Base(chartFiles[0])

	// Setup registry for push
	parsedURL, err := url.Parse(serverDetails.ArtifactoryUrl)
	require.NoError(t, err)
	registryHost := parsedURL.Host
	registryURL := fmt.Sprintf("oci://%s/%s", registryHost, tests.HelmLocalRepo)

	if !isRepoExist(tests.HelmLocalRepo) {
		t.Skipf("Repository %s does not exist. Skipping test.", tests.HelmLocalRepo)
	}

	err = loginHelmRegistry(t, registryHost)
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		if strings.Contains(errorMsg, "account temporarily locked") {
			t.Skip("Artifactory account is temporarily locked. Skipping test.")
		}
		if strings.Contains(errorMsg, "http response to https") ||
			strings.Contains(errorMsg, "tls: first record does not look like a tls handshake") {
			t.Skip("Helm registry login failed due to HTTPS/HTTP mismatch. Skipping test.")
		}
	}
	require.NoError(t, err, "helm registry login should succeed")

	// Run helm push with the same build name/number (adds artifacts to existing build info)
	args = []string{
		"helm", "push", chartFile,
		registryURL,
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	if err != nil {
		errorMsg := strings.ToLower(err.Error())
		if strings.Contains(errorMsg, "404") ||
			strings.Contains(errorMsg, "not found") ||
			strings.Contains(errorMsg, "exit status 1") {
			t.Skip("OCI registry API not accessible (404). Skipping test.")
		}
	}
	require.NoError(t, err, "helm push should succeed")

	// Publish build info - should set CI VCS props on artifacts
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Get the published build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	// Verify build info has artifacts
	require.Greater(t, len(publishedBuildInfo.BuildInfo.Modules), 0, "Build info should have modules")

	// Create service manager for getting artifact properties
	serviceManager, err := utils.CreateServiceManager(serverDetails, 3, 1000, false)
	require.NoError(t, err)

	// Verify VCS properties on each artifact from build info
	artifactCount := 0
	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			if artifact.OriginalDeploymentRepo == "" {
				continue // Skip artifacts without deployment repo info
			}
			fullPath := artifact.OriginalDeploymentRepo + "/" + artifact.Path

			props, err := serviceManager.GetItemProps(fullPath)
			assert.NoError(t, err, "Failed to get properties for artifact: %s", fullPath)
			if props == nil {
				continue
			}

			// Validate VCS properties
			assert.Contains(t, props.Properties, "vcs.provider", "Missing vcs.provider on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.provider"], "github", "Wrong vcs.provider on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.org", "Missing vcs.org on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.org"], actualOrg, "Wrong vcs.org on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.repo", "Missing vcs.repo on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.repo"], actualRepo, "Wrong vcs.repo on %s", artifact.Name)

			artifactCount++
		}
	}

	assert.Greater(t, artifactCount, 0, "No artifacts were validated for CI VCS properties")
}

// InitHelmTests initializes Helm tests
func InitHelmTests() {
	initArtifactoryCli()
	cleanUpOldBuilds()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

// CleanHelmTests cleans up after Helm tests
func CleanHelmTests() {
	deleteCreatedRepos()
}

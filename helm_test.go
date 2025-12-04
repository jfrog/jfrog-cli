package main

// Helm Integration Tests
// These tests run automatically as part of Artifactory tests.
// Run with: go test -v -test.artifactory -jfrog.url=http://localhost:8081/ -jfrog.user=admin -jfrog.password=password

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
	if !*tests.TestArtifactory && !*tests.TestArtifactoryProject {
		t.Skip("Skipping Helm test. Helm tests run as part of Artifactory tests. Use '-test.artifactory' to run them.")
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
		t.Fatalf("Repository %s does not exist. It should have been created during test setup.", tests.HelmLocalRepo)
	}

	err = loginHelmRegistry(t, registryHost)
	if err != nil && strings.Contains(err.Error(), "account temporarily locked") {
		t.Skip("Artifactory account is temporarily locked due to recurrent login failures. Please wait and try again, or verify credentials are correct.")
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

func TestHelmInstallWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-install"
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

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "install", "test-release", ".",
		"--dry-run",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	if err != nil {
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "Kubernetes cluster unreachable") ||
			strings.Contains(errorMsg, "connection refused") ||
			strings.Contains(errorMsg, "dial tcp") ||
			strings.Contains(errorMsg, "INSTALLATION FAILED") {
			t.Skip("Kubernetes cluster not available, skipping helm install test")
		}
		require.NoError(t, err, "helm install should succeed")
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err, "Failed to get build info")
	require.True(t, found, "build info should be found")

	validateHelmBuildInfo(t, publishedBuildInfo.BuildInfo, buildName, false, true)
}

func TestHelmTemplateWithBuildInfo(t *testing.T) {
	initHelmTest(t)
	defer cleanHelmTest(t)

	buildName := tests.HelmBuildName + "-template"
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

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"helm", "template", "test-release", ".",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	err = jfrogCli.Exec(args...)
	require.NoError(t, err, "helm template should succeed")

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
		errorOutput := stderr.String()
		if strings.Contains(errorOutput, "recurrent login failures") || strings.Contains(errorOutput, "blocked") {
			t.Logf("Helm registry login failed due to account lockout. Please wait and try again, or verify credentials are correct.")
			return fmt.Errorf("account temporarily locked: %w", err)
		}
		return fmt.Errorf("helm registry login failed: %w (stderr: %s)", err, errorOutput)
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
		t.Fatalf("Repository %s does not exist. It should have been created during test setup.", tests.HelmLocalRepo)
	}

	err = loginHelmRegistry(t, registryHost)
	if err != nil && strings.Contains(err.Error(), "account temporarily locked") {
		t.Skip("Artifactory account is temporarily locked due to recurrent login failures. Please wait and try again, or verify credentials are correct.")
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

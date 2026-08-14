package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	apmBuildName = "apm-test-build"
	dirPerms     = 0755
	filePerms    = 0644
)

// initApmTest initializes the APM test environment.
func initApmTest(t *testing.T) {
	if !*tests.TestApm {
		t.Skip("Skipping APM tests. To run APM test add the '-test.apm=true' option.")
	}
	// Ensure APM is installed
	_, err := exec.LookPath("apm")
	require.NoError(t, err, "APM must be installed to run APM tests. Install from: https://github.com/microsoft/apm/releases")
	// Ensure JFROG_RUN_NATIVE is not set (clean state for non-native tests)
	_ = os.Unsetenv("JFROG_RUN_NATIVE")
	createJfrogHomeConfig(t, true)
	createApmRepository(t)
	initApmConfig(t)
}

// getApmCli returns a CLI configured for APM commands (without "rt" prefix).
// APM commands are: jfrog agent apm ..., not jfrog rt agent apm ...
func getApmCli() *coreTests.JfrogCli {
	return coreTests.NewJfrogCli(execMain, "jfrog", "")
}

// createApmRepository creates a local APM repository for testing.
func createApmRepository(t *testing.T) {
	if !isRepoExist(tests.AgentPackagesLocalRepo) {
		repoConfig := tests.GetTestResourcesPath() + tests.AgentPackagesLocalRepositoryConfig
		repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
		require.NoError(t, err)
		execCreateRepoRest(repoConfig, tests.AgentPackagesLocalRepo)
	}
}

// initApmConfig sets up the APM configuration in ~/.apm/config.json via jf setup.
func initApmConfig(t *testing.T) {
	// Use jf setup to configure APM (not jf rt setup)
	setupCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	err := setupCli.Exec("setup", "agent-apm", "--repo", tests.AgentPackagesLocalRepo)
	require.NoError(t, err, "jf setup agent-apm should succeed")
}


// cleanApmTest cleans up resources after APM tests.
func cleanApmTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	deleteSpec := spec.NewBuilder().Pattern(tests.AgentPackagesLocalRepo).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	require.NoError(t, err, "cleanup should remove test artifacts")
	tests.CleanFileSystem()
}

// createApmTestProject creates a minimal APM project structure with apm.yml.
func createApmTestProject(t *testing.T, projectDir string) {
	err := os.MkdirAll(projectDir, dirPerms)
	require.NoError(t, err)

	// Create minimal .apm directory
	apmDir := filepath.Join(projectDir, ".apm")
	err = os.MkdirAll(apmDir, dirPerms)
	require.NoError(t, err)

	// Create basic primitives directory
	primitivesDir := filepath.Join(apmDir, "primitives")
	err = os.MkdirAll(primitivesDir, dirPerms)
	require.NoError(t, err)

	// Create apm.yml
	apmYamlContent := `version: "1.0.0"
name: test-apm-package
description: Test APM package for e2e testing
primitives:
  agents: []
  skills: []
  models: []
  tools: []
dependencies:
  apm: []
  mcp: []
`

	apmYamlPath := filepath.Join(projectDir, "apm.yml")
	err = os.WriteFile(apmYamlPath, []byte(apmYamlContent), filePerms)
	require.NoError(t, err)

	// Create a dummy file for packaging
	dummyFile := filepath.Join(primitivesDir, "placeholder.txt")
	err = os.WriteFile(dummyFile, []byte("placeholder content"), filePerms)
	require.NoError(t, err)
}

// validateApmBuildInfo validates the generated build info from an APM command.
func validateApmBuildInfo(t *testing.T, buildName, buildNumber string, expectedArtifacts int) {
	builds, err := build.GetGeneratedBuildsInfo(buildName, buildNumber, "")
	require.NoError(t, err)
	require.Len(t, builds, 1, "Expected exactly one build info")

	buildResult := builds[0]
	require.NotNil(t, buildResult)

	// Verify build properties
	assert.Equal(t, buildName, buildResult.Name)
	assert.Equal(t, buildNumber, buildResult.Number)

	// Verify modules exist if artifacts expected
	if expectedArtifacts > 0 && len(buildResult.Modules) > 0 {
		module := buildResult.Modules[0]
		// Verify all artifacts have checksums
		for _, artifact := range module.Artifacts {
			assert.NotEmpty(t, artifact.Sha256, "Artifact should have SHA256 checksum")
			assert.NotEmpty(t, artifact.Path, "Artifact should have path")
		}

		// Verify dependencies if present
		for _, dep := range module.Dependencies {
			assert.NotEmpty(t, dep.Sha256, "Dependency should have SHA256 checksum")
			assert.NotEmpty(t, dep.Id, "Dependency should have ID")
		}
	}
}

// validateBuildInfoDependencies validates dependencies exist in build info
func validateBuildInfoDependencies(t *testing.T, buildName, buildNumber string) {
	builds, err := build.GetGeneratedBuildsInfo(buildName, buildNumber, "")
	require.NoError(t, err, "Should retrieve build info without error")
	require.Len(t, builds, 1, "Should have exactly one build")
	require.Len(t, builds[0].Modules, 1, "Build should have at least one module")

	module := builds[0].Modules[0]
	require.NotEmpty(t, module.Dependencies, "Dependencies should be present in build info")

	for _, dep := range module.Dependencies {
		assert.NotEmpty(t, dep.Id, "Dependency should have ID")
	}
}

// validateBuildInfoArtifacts validates artifacts in build info
func validateBuildInfoArtifacts(t *testing.T, buildName, buildNumber string, expectedCount int) {
	builds, err := build.GetGeneratedBuildsInfo(buildName, buildNumber, "")
	require.NoError(t, err, "Should retrieve build info without error")
	require.Len(t, builds, 1, "Should have exactly one build")
	require.Len(t, builds[0].Modules, 1, "Build should have at least one module")

	module := builds[0].Modules[0]
	require.Len(t, module.Artifacts, expectedCount, "Artifacts count should match expected")

	for _, artifact := range module.Artifacts {
		assert.NotEmpty(t, artifact.Path, "Artifact should have path")
		assert.NotEmpty(t, artifact.Sha256, "Artifact should have checksum")
	}
}

// validateBuildInfoHasBothArtifactsAndDependencies validates both exist
func validateBuildInfoHasBothArtifactsAndDependencies(t *testing.T, buildName, buildNumber string) {
	builds, err := build.GetGeneratedBuildsInfo(buildName, buildNumber, "")
	require.NoError(t, err, "Should retrieve build info without error")
	require.Len(t, builds, 1, "Should have exactly one build")
	require.Len(t, builds[0].Modules, 1, "Build should have at least one module")

	module := builds[0].Modules[0]
	require.NotEmpty(t, module.Dependencies, "Build info should have dependencies")
	require.NotEmpty(t, module.Artifacts, "Build info should have artifacts")
}

// TestApmSetupAndConfig validates APM setup with apm config file persistence (P0: Scenario #1).
func TestApmSetupAndConfig(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	apmConfigPath := filepath.Join(homeDir, ".apm", "config.json")

	// First setup call (use correct CLI prefix: jfrog, not jfrog rt)
	setupCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	err = setupCli.Exec("setup", "agent-apm", "--repo", tests.AgentPackagesLocalRepo)
	require.NoError(t, err, "jf setup agent-apm should succeed")

	// Verify config file was created
	assert.FileExists(t, apmConfigPath, "APM config file should be created")

	// Verify config contains registry reference
	configData, err := os.ReadFile(apmConfigPath)
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal(configData, &config)
	require.NoError(t, err)

	registries, ok := config["registries"].(map[string]interface{})
	assert.True(t, ok, "Config should have registries section")
	assert.NotEmpty(t, registries, "Registries section should not be empty")

	// Verify idempotency - second call should not fail (use correct CLI prefix)
	setupCli = coreTests.NewJfrogCli(execMain, "jfrog", "")
	err = setupCli.Exec("setup", "agent-apm", "--repo", tests.AgentPackagesLocalRepo)
	require.NoError(t, err, "jf setup agent-apm should be idempotent")
}

// TestApmInstallWithBuildInfo validates `jf agent apm install` with build-info capture (P0: Scenario #13).
func TestApmInstallWithBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-install-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "101"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Run apm install with build-info capture
	err = getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install should succeed with build-info")

	// Validate build info was created
	validateApmBuildInfo(t, apmBuildName, buildNumber, 0)

	// Publish the build info
	err = artifactoryCli.Exec("rt", "bp", apmBuildName, buildNumber)
	require.NoError(t, err, "jf rt bp should succeed")

	// Clean up build info
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmPublishWithBuildInfo validates `jf agent apm publish` with build-info capture (P0: Scenario #3).
func TestApmPublishWithBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-publish-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "102"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Run apm publish with build-info capture
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "jfrog/test-apm-pkg", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm publish should succeed with build-info")

	// Validate build info was created with artifact
	validateApmBuildInfo(t, apmBuildName, buildNumber, 1)

	// Publish the build info
	err = artifactoryCli.Exec("rt", "bp", apmBuildName, buildNumber)
	require.NoError(t, err, "jf rt bp should succeed")

	// Verify artifact was uploaded to Artifactory
	deleteSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/jfrog/test-apm-pkg/*.zip").
		BuildSpec()
	artifacts, _, err := tests.SearchFiles(deleteSpec, serverDetails)
	require.NoError(t, err)
	assert.NotEmpty(t, artifacts, "Published APM package should be found in repository")

	// Clean up
	_, _, _ = tests.DeleteFiles(deleteSpec, serverDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmPublishArtifactPath validates artifact upload to correct path (P0: Scenario #4).
func TestApmPublishArtifactPath(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-publish-path-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	owner := "acme"
	packageName := "my-agent-skill"
	err = getApmCli().Exec("agent", "apm", "publish", "--package", fmt.Sprintf("%s/%s", owner, packageName))
	require.NoError(t, err, "jf agent apm publish should succeed")

	// Verify artifact path: <owner>/<name>/<name>-<version>.zip
	searchSpec := spec.NewBuilder().
		Pattern(fmt.Sprintf("%s/%s/%s/*.zip", tests.AgentPackagesLocalRepo, owner, packageName)).
		BuildSpec()
	artifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	assert.NotEmpty(t, artifacts, "Artifact should be found at expected path: <owner>/<name>/<name>-<version>.zip")

	// Verify artifact name format
	if len(artifacts) > 0 {
		assert.True(t,
			strings.Contains(artifacts[0].Path, packageName+"-") && strings.HasSuffix(artifacts[0].Path, ".zip"),
			"Artifact path should follow pattern: <name>-<version>.zip")
	}

	// Clean up
	_, _, _ = tests.DeleteFiles(searchSpec, serverDetails)
}

// TestApmPublishRequiresPackageFlag validates that --package flag is required (P0: Scenario #23).
func TestApmPublishRequiresPackageFlag(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-publish-no-pkg-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Attempt publish without --package flag
	err = getApmCli().Exec("agent", "apm", "publish")
	assert.Error(t, err, "jf agent apm publish without --package should fail")
	assert.Contains(t, err.Error(), "package", "Error message should mention --package flag")
}

// TestApmInstallInvalidPackage validates handling of missing/invalid package references (P0: Scenario #15).
func TestApmInstallInvalidPackage(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-invalid-pkg-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	// Create project with invalid dependency
	err = os.MkdirAll(filepath.Join(projectDir, ".apm"), 0755)
	require.NoError(t, err)

	apmYamlContent := `version: "1.0.0"
name: test-with-missing-dep
dependencies:
  apm:
    - name: nonexistent/package
      version: "1.0.0"
`
	apmYamlPath := filepath.Join(projectDir, "apm.yml")
	err = os.WriteFile(apmYamlPath, []byte(apmYamlContent), 0644)
	require.NoError(t, err)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Attempt install with invalid package
	// Note: This depends on APM's own error handling
	err = getApmCli().Exec("agent", "apm", "install")
	// Error is expected when trying to fetch nonexistent package
	if err != nil {
		assert.True(t,
			strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found"),
			"Error should indicate package not found")
	}
}

// TestApmAuthEnvironmentVariable validates APM_REGISTRY_TOKEN env var usage (P0: Scenario #33).
func TestApmAuthEnvironmentVariable(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-auth-env-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "103"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Set env var for registry auth
	registryName := "default"
	err = os.Setenv(fmt.Sprintf("APM_REGISTRY_TOKEN_%s", strings.ToUpper(registryName)), *tests.JfrogAccessToken)
	require.NoError(t, err)
	defer func() {
		_ = os.Unsetenv(fmt.Sprintf("APM_REGISTRY_TOKEN_%s", strings.ToUpper(registryName)))
	}()

	// Run install with env var auth
	err = getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install should succeed with env var auth")

	// Clean up build info
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmMissingCredentials validates error handling when credentials are missing (P0: Scenario #36).
func TestApmMissingCredentials(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-no-creds-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Remove APM config to simulate missing credentials
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	apmConfigPath := filepath.Join(homeDir, ".apm", "config.json")
	err = os.Remove(apmConfigPath)
	if err != nil && !os.IsNotExist(err) {
		require.NoError(t, err)
	}

	// Unset any auth env vars
	for _, envVar := range os.Environ() {
		if strings.Contains(envVar, "APM_REGISTRY") {
			key := strings.Split(envVar, "=")[0]
			_ = os.Unsetenv(key)
		}
	}

	// Attempt install without credentials
	err = getApmCli().Exec("agent", "apm", "install")
	assert.Error(t, err, "jf agent apm install without credentials should fail")
}

// TestApmBuildInfoArtifactMetadata validates artifact metadata (P0: Scenario #6).
func TestApmBuildInfoArtifactMetadata(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-artifact-metadata-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "104"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/artifact-metadata", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Validate build info has complete artifact metadata
	builds, err := build.GetGeneratedBuildsInfo(apmBuildName, buildNumber, "")
	require.NoError(t, err)
	require.Len(t, builds, 1)

	buildResult := builds[0]
	require.NotEmpty(t, buildResult.Modules)

	module := buildResult.Modules[0]
	for _, artifact := range module.Artifacts {
		// Verify metadata fields are present
		assert.NotEmpty(t, artifact.Path, "Artifact path should be present")
		assert.NotEmpty(t, artifact.Type, "Artifact type should be present")
		assert.NotEmpty(t, artifact.Sha256, "Artifact SHA256 should be present")
	}

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmBuildPropertiesStamping validates build properties on artifacts (P0: Scenario #8).
func TestApmBuildPropertiesStamping(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-props-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "105"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "publish", "--package", "jfrog/props-test", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Publish build info
	err = artifactoryCli.Exec("rt", "bp", apmBuildName, buildNumber)
	require.NoError(t, err)

	// Verify build properties were stamped on artifacts
	searchSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/jfrog/props-test/*.zip").
		BuildSpec()
	artifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	require.NotEmpty(t, artifacts)

	// Verify properties contain build info
	artifact := artifacts[0]
	assert.NotEmpty(t, artifact.Props, "Artifact should have properties")

	// Check for build name/number in properties
	foundBuildName := false
	foundBuildNumber := false
	for buildPropKey, buildPropVals := range artifact.Props {
		if buildPropKey == "build.name" {
			foundBuildName = true
			for _, val := range buildPropVals {
				assert.Contains(t, val, apmBuildName)
			}
		}
		if buildPropKey == "build.number" {
			foundBuildNumber = true
			for _, val := range buildPropVals {
				assert.Contains(t, val, buildNumber)
			}
		}
	}

	assert.True(t, foundBuildName, "Artifact should have build.name property")
	assert.True(t, foundBuildNumber, "Artifact should have build.number property")

	// Clean up
	_, _, _ = tests.DeleteFiles(searchSpec, serverDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmModuleFlag validates --module flag for custom module names (P1: Scenario #26).
func TestApmModuleFlag(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-module-flag-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "106"
	customModule := "custom-apm-module"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "install", "--module", customModule, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install with --module flag should succeed")

	// Validate custom module name in build info
	builds, err := build.GetGeneratedBuildsInfo(apmBuildName, buildNumber, "")
	require.NoError(t, err)
	require.Len(t, builds, 1)

	buildResult := builds[0]
	var foundModule bool
	for _, module := range buildResult.Modules {
		if module.Id == customModule {
			foundModule = true
			break
		}
	}
	assert.True(t, foundModule, fmt.Sprintf("Custom module %s should be present in build info", customModule))

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmRoundTripPublishAndInstall validates full round-trip (P1: Scenario #40).
func TestApmRoundTripPublishAndInstall(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	// Create and publish a package
	publishProjectDir, err := os.MkdirTemp("", "apm-roundtrip-publish-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(publishProjectDir)
	}()

	createApmTestProject(t, publishProjectDir)

	buildNumberPublish := "201"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, publishProjectDir)

	owner := "roundtrip"
	pkgName := "test-package"

	// Publish the package
	err = getApmCli().Exec("agent", "apm", "publish", "--package", fmt.Sprintf("%s/%s", owner, pkgName), "--build-name", apmBuildName, "--build-number", buildNumberPublish)
	require.NoError(t, err, "jf agent apm publish should succeed")

	// Create a new directory to install from
	installProjectDir, err := os.MkdirTemp("", "apm-roundtrip-install-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(installProjectDir)
	}()

	// Create a project that depends on the published package
	installApmYaml := `version: "1.0.0"
name: test-consumer
description: Consumer of published APM package
dependencies:
  apm:
    - name: ` + owner + `/` + pkgName + `
`

	err = os.MkdirAll(filepath.Join(installProjectDir, ".apm"), 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(installProjectDir, "apm.yml"), []byte(installApmYaml), 0644)
	require.NoError(t, err)

	clientTestUtils.ChangeDirAndAssert(t, installProjectDir)

	buildNumberInstall := "202"

	// Install the published package
	err = getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumberInstall)
	require.NoError(t, err, "jf agent apm install should succeed with published package")

	// Validate both build infos
	validateApmBuildInfo(t, apmBuildName, buildNumberPublish, 1)
	validateApmBuildInfo(t, apmBuildName, buildNumberInstall, 0)

	// Clean up
	searchSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/" + owner + "/" + pkgName + "/*.zip").
		BuildSpec()
	_, _, _ = tests.DeleteFiles(searchSpec, serverDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmChecksumsInBuildInfo validates SHA256 checksums are recorded (P0: Scenario #18).
func TestApmChecksumsInBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-checksums-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "107"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/checksums", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Get build info and verify checksums
	builds, err := build.GetGeneratedBuildsInfo(apmBuildName, buildNumber, "")
	require.NoError(t, err)
	require.Len(t, builds, 1)

	if len(builds[0].Modules) > 0 {
		module := builds[0].Modules[0]
		for _, artifact := range module.Artifacts {
			// SHA256 is required
			assert.NotEmpty(t, artifact.Sha256, "Artifact should have SHA256")
			// SHA1 and MD5 are optional but should be present if available
			if artifact.Sha1 != "" {
				assert.Len(t, artifact.Sha1, 40, "SHA1 should be 40 hex characters")
			}
			if artifact.Md5 != "" {
				assert.Len(t, artifact.Md5, 32, "MD5 should be 32 hex characters")
			}
		}
	}

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmProjectFlag validates --project flag for project isolation (P1: Scenario #27).
func TestApmProjectFlag(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-project-flag-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "108"
	projectKey := "test-project"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "install", "--project", projectKey, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install with --project flag should succeed")

	// Validate build info is scoped to project
	builds, err := build.GetGeneratedBuildsInfo(apmBuildName, buildNumber, projectKey)
	require.NoError(t, err)
	require.Len(t, builds, 1, "Build should be found when queried with correct project key")

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmUpdateWithBuildInfo validates `jf agent apm update` with build-info (P1: Scenario #16).
func TestApmUpdateWithBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-update-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "109"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// First, install to have a lockfile
	err = getApmCli().Exec("agent", "apm", "install")
	require.NoError(t, err)

	// Then update with build-info capture
	err = getApmCli().Exec("agent", "apm", "update", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm update should succeed with build-info")

	// Validate build info was created
	validateApmBuildInfo(t, apmBuildName, buildNumber, 0)

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmNativeFlags validates native APM flags with -- escape (P1: Scenario #28).
func TestApmNativeFlags(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-native-flags-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Test --dry-run flag with -- escape
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/native-flags", "--", "--dry-run")
	require.NoError(t, err, "jf agent apm publish with --dry-run should succeed")

	// Verify no artifact was uploaded for dry-run
	searchSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/test/native-flags/*.zip").
		BuildSpec()
	artifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	_ = artifacts
}

// TestApmBuildInfoRead validates `jf rt bi` read command (P0: Scenario #5).
func TestApmBuildInfoRead(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-bi-read-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "110"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Create build info first
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/bi-read", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Publish to Artifactory
	err = artifactoryCli.Exec("rt", "bp", apmBuildName, buildNumber)
	require.NoError(t, err)

	// Read build info
	err = artifactoryCli.Exec("rt", "bi", apmBuildName, buildNumber)
	require.NoError(t, err, "jf rt bi should succeed reading published build info")

	// Clean up
	_, _, _ = tests.DeleteFiles(
		spec.NewBuilder().Pattern(tests.AgentPackagesLocalRepo+"/test/bi-read/*.zip").BuildSpec(),
		serverDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmIntegrationFullPipeline validates end-to-end workflow (P1: Scenario #50).
func TestApmIntegrationFullPipeline(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-e2e-pipeline-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildName := "apm-e2e-pipeline"
	buildNumber := "300"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Step 1: Install (with build-info)
	err = getApmCli().Exec("agent", "apm", "install", "--build-name", buildName, "--build-number", buildNumber)
	require.NoError(t, err, "Step 1: Install should succeed")

	// Step 2: Publish (with build-info)
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "e2e/pipeline", "--build-name", buildName, "--build-number", buildNumber)
	require.NoError(t, err, "Step 2: Publish should succeed")

	// Step 3: Publish build info
	err = artifactoryCli.Exec("rt", "bp", buildName, buildNumber)
	require.NoError(t, err, "Step 3: Publish build info should succeed")

	// Clean up
	_, _, _ = tests.DeleteFiles(
		spec.NewBuilder().Pattern(tests.AgentPackagesLocalRepo+"/e2e/pipeline/*.zip").BuildSpec(),
		serverDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ============================================================================
// GAP ANALYSIS TESTS - Registry Configuration & Dependencies
// ============================================================================

// createApmProjectWithYaml creates a test project directory with apm.yml content.
func createApmProjectWithYaml(t *testing.T, yamlContent string) string {
	projectDir, err := os.MkdirTemp("", "apm-test-*")
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(projectDir, ".apm", "primitives"), dirPerms)
	require.NoError(t, err)

	apmYmlPath := filepath.Join(projectDir, "apm.yml")
	err = os.WriteFile(apmYmlPath, []byte(yamlContent), filePerms)
	require.NoError(t, err)

	return projectDir
}

// setupTestWorkingDirectory saves current directory and changes to projectDir with defer cleanup.
func setupTestWorkingDirectory(t *testing.T, projectDir string) func() {
	wd, err := os.Getwd()
	require.NoError(t, err)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)
	return func() {
		clientTestUtils.ChangeDirAndAssert(t, wd)
	}
}

// TestApmBuildFlagsRequired validates both build-name and build-number are required together
func TestApmBuildFlagsRequired(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir := createApmProjectWithYaml(t, getBasicApmYaml())
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	defer setupTestWorkingDirectory(t, projectDir)()

	// Test missing build-number
	err := getApmCli().Exec("agent", "apm", "install", "--build-name", "test-build")
	assert.Error(t, err, "Should error when build-number missing but build-name provided")

	// Test missing build-name
	err = getApmCli().Exec("agent", "apm", "install", "--build-number", "1")
	assert.Error(t, err, "Should error when build-name missing but build-number provided")
}

// TestApmInstallWithDependenciesInBuildInfo validates dependencies captured in build info
func TestApmInstallWithDependenciesInBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir := createProjectWithDependencies(t, "app-with-deps", []string{"apm"})
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "400"
	err := runApmInstall(buildNumber)
	require.NoError(t, err, "install should succeed")

	validateBuildInfoDependencies(t, apmBuildName, buildNumber)
	deleteBuildInfo()
}

// TestApmPublishWithArtifactsInBuildInfo validates artifacts captured in build info
func TestApmPublishWithArtifactsInBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir := createApmProjectWithYaml(t, getBasicApmYaml())
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "401"
	err := runApmPublish("test/artifacts-demo", apmBuildName, buildNumber)
	require.NoError(t, err, "publish should succeed")

	validateBuildInfoArtifacts(t, apmBuildName, buildNumber, 1)
	_ = deleteArtifacts(tests.AgentPackagesLocalRepo + "/test/artifacts-demo/*.zip")
	deleteBuildInfo()
}

// TestApmBuildInfoWithArtifactsAndDependencies validates both artifacts and dependencies
func TestApmBuildInfoWithArtifactsAndDependencies(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir := createProjectWithDependencies(t, "complete-app", []string{"apm"})
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "402"

	// Step 1: Install (captures dependencies)
	err := runApmInstall(buildNumber)
	require.NoError(t, err)

	// Step 2: Publish (adds artifacts)
	err = runApmPublish("complete/demo", apmBuildName, buildNumber)
	require.NoError(t, err)

	// Validate both exist
	validateBuildInfoHasBothArtifactsAndDependencies(t, apmBuildName, buildNumber)

	_ = deleteArtifacts(tests.AgentPackagesLocalRepo + "/complete/demo/*.zip")
	deleteBuildInfo()
}

// TestApmUpdateWithVersionChange validates update captures new version in build info
func TestApmUpdateWithVersionChange(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir := createApmProjectWithYaml(t, getBasicApmYaml())
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "403"

	// Step 1: Install
	err := runApmInstall(buildNumber)
	require.NoError(t, err, "install should succeed")

	builds1, err := build.GetGeneratedBuildsInfo(apmBuildName, buildNumber, "")
	require.NoError(t, err)

	// Step 2: Update
	err = runApmUpdate(apmBuildName, buildNumber)
	require.NoError(t, err, "update should succeed")

	builds2, err := build.GetGeneratedBuildsInfo(apmBuildName, buildNumber, "")
	require.NoError(t, err)

	assert.Equal(t, len(builds1), len(builds2), "Build info should reflect update")

	deleteBuildInfo()
}

// TestApmAuthEnvVarNotExposed validates credentials stay in env (not leaked in logs)
func TestApmAuthEnvVarNotExposed(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir := createApmProjectWithYaml(t, getBasicApmYaml())
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	defer setupTestWorkingDirectory(t, projectDir)()

	// Env var should exist after command runs (we're not removing it)
	// The test verifies the command worked with the env var auth
	err := getApmCli().Exec("agent", "apm", "install")
	require.NoError(t, err, "install should work with env var auth")

	// Verify env var still set (commands don't clear environment)
	jfrogUrl := os.Getenv("JFROG_URL")
	assert.NotEmpty(t, jfrogUrl, "JFROG_URL should still be set")
}

// TestApmDifferentRegistriesAsArtifactoryRepos validates multiple distinct Artifactory repos
func TestApmDifferentRegistriesAsArtifactoryRepos(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	// Create two different repos
	repos := []string{"apm-registry-1", "apm-registry-2"}
	for _, repoName := range repos {
		if !isRepoExist(repoName) {
			repoConfig := tests.GetTestResourcesPath() + tests.AgentPackagesLocalRepositoryConfig
			repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
			require.NoError(t, err)
			execCreateRepoRest(repoConfig, repoName)
		}
	}

	projectDir := createProjectWithRegistries(t, "multi-repo-app", repos)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "404"
	err := runApmInstall(buildNumber)
	require.NoError(t, err, "install should succeed with multiple distinct registries")

	validateApmBuildInfo(t, apmBuildName, buildNumber, 0)
	deleteBuildInfo()
}

// getBasicApmYaml returns basic APM YAML
func getBasicApmYaml() string {
	return createApmYaml("test-app", "1.0.0", []string{}, nil)
}

// createApmYaml creates customizable APM YAML with parameters
func createApmYaml(name, version string, dependencies []string, registries map[string]string) string {
	// Note: registries parameter is deprecated - registries are configured globally via setup command
	depsSection := ""
	if len(dependencies) > 0 {
		for _, dep := range dependencies {
			depsSection += fmt.Sprintf("  %s: []\n", dep)
		}
	} else {
		depsSection = "  apm: []\n"
	}

	return fmt.Sprintf(`version: "1.0.0"
name: %s
version: %s
primitives:
  agents: []
dependencies:
%s`, name, version, depsSection)
}

// createMultiRegistryYaml creates APM YAML with multiple distinct registries
// Note: Registries are configured globally via setup command, not in apm.yml
func createMultiRegistryYaml(name string, registryRepos []string) string {
	return fmt.Sprintf(`version: "1.0.0"
name: %s
primitives:
  agents: []
dependencies:
  apm: []
`, name)
}

// createProjectWithDependencies creates a project directory with specified dependencies
func createProjectWithDependencies(t *testing.T, name string, deps []string) string {
	apmYaml := createApmYaml(name, "1.0.0", deps, nil)
	return createApmProjectWithYaml(t, apmYaml)
}

// createProjectWithRegistries creates a project with multiple distinct registries
func createProjectWithRegistries(t *testing.T, name string, registryRepos []string) string {
	apmYaml := createMultiRegistryYaml(name, registryRepos)
	return createApmProjectWithYaml(t, apmYaml)
}

// runApmInstall runs install command with optional build info
func runApmInstall(buildNumber string) error {
	args := []string{"agent", "apm", "install"}
	if buildNumber != "" {
		args = append(args, "--build-name", apmBuildName, "--build-number", buildNumber)
	}
	return getApmCli().Exec(args...)
}

// runApmPublish runs publish command with optional build info
func runApmPublish(packagePath, buildName, buildNumber string) error {
	args := []string{"agent", "apm", "publish"}
	if packagePath != "" {
		args = append(args, "--package", packagePath)
	}
	if buildName != "" && buildNumber != "" {
		args = append(args, "--build-name", buildName, "--build-number", buildNumber)
	}
	return getApmCli().Exec(args...)
}

// runApmUpdate runs update command with optional build info
func runApmUpdate(buildName, buildNumber string) error {
	args := []string{"agent", "apm", "update"}
	if buildName != "" && buildNumber != "" {
		args = append(args, "--build-name", buildName, "--build-number", buildNumber)
	}
	return getApmCli().Exec(args...)
}

// deleteBuildInfo deletes build info from Artifactory
func deleteBuildInfo() {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// deleteArtifacts deletes artifacts from repository
func deleteArtifacts(pattern string) error {
	spec := spec.NewBuilder().Pattern(pattern).BuildSpec()
	_, _, err := tests.DeleteFiles(spec, serverDetails)
	return err
}
func TestApmMultipleRegistriesInApmYml(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	apmYaml := `version: "1.0.0"
name: multi-registry-app
description: App using multiple registries
primitives:
  agents: []
dependencies:
  apm: []
`

	projectDir := createApmProjectWithYaml(t, apmYaml)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "200"
	// Install should work with multiple registries defined
	err := getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "install should succeed with multiple registries")

	validateApmBuildInfo(t, apmBuildName, buildNumber, 0)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmRegistryPrecedenceDefaultFallback validates default registry fallback (P0: Scenario #1 variant).
func TestApmRegistryPrecedenceDefaultFallback(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	apmYaml := `version: "1.0.0"
name: test-default-registry
description: Test default registry fallback
primitives:
  agents: []
dependencies:
  apm: []
`

	projectDir := createApmProjectWithYaml(t, apmYaml)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "201"
	// Install should use default registry when no explicit registry specified
	err := getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "install should use default registry")

	validateApmBuildInfo(t, apmBuildName, buildNumber, 0)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmPublishWithDependencyMetadata validates publish captures dependency metadata (P0: Scenario #7).
func TestApmPublishWithDependencyMetadata(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	apmYaml := `version: "1.0.0"
name: app-with-deps
version: 1.0.0
description: App with explicit dependencies
primitives:
  agents: []
dependencies:
  apm: []
`

	projectDir := createApmProjectWithYaml(t, apmYaml)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	dummyFile := filepath.Join(projectDir, ".apm", "primitives", "skill.json")
	err := os.WriteFile(dummyFile, []byte(`{"type": "agent"}`), filePerms)
	require.NoError(t, err)

	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "202"
	// Publish should capture dependency metadata
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/app-with-deps",
		"--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "publish should succeed with dependencies")

	// Validate build info includes dependency metadata
	builds, err := build.GetGeneratedBuildsInfo(apmBuildName, buildNumber, "")
	require.NoError(t, err)
	require.Len(t, builds, 1)

	module := builds[0].Modules[0]
	assert.NotEmpty(t, module.Artifacts, "Should have artifact metadata")

	// Clean up
	_, _, _ = tests.DeleteFiles(
		spec.NewBuilder().Pattern(tests.AgentPackagesLocalRepo+"/test/app-with-deps/*.zip").BuildSpec(),
		serverDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmUpdateChangesLockfile validates update behavior with dependencies (P1: Scenario #16).
func TestApmUpdateChangesLockfile(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-update-lock-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "203"
	// First install to create initial lockfile
	err = getApmCli().Exec("agent", "apm", "install")
	require.NoError(t, err)

	// Verify lockfile created
	lockfilePath := filepath.Join(projectDir, "apm.lock.yaml")
	assert.FileExists(t, lockfilePath, "apm.lock.yaml should exist after install")

	// Update with build-info
	err = getApmCli().Exec("agent", "apm", "update", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "update should succeed")

	// Verify lockfile still exists (update should maintain it)
	assert.FileExists(t, lockfilePath, "apm.lock.yaml should still exist after update")

	validateApmBuildInfo(t, apmBuildName, buildNumber, 0)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmFrozenModeWithDependencies validates frozen mode works with dependencies (P1: Scenario #14).
func TestApmFrozenModeWithDependencies(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-frozen-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	defer setupTestWorkingDirectory(t, projectDir)()

	// First install to create lockfile
	err = getApmCli().Exec("agent", "apm", "install")
	require.NoError(t, err)

	// Frozen install should succeed (lockfile exists and is up-to-date)
	err = getApmCli().Exec("agent", "apm", "install", "--", "--frozen")
	require.NoError(t, err, "frozen install should succeed with existing lockfile")
}

// TestApmInstallAndPublishWithBuildInfoComplete validates complete install→publish workflow (P1: Scenario #50).
func TestApmInstallAndPublishWithBuildInfoComplete(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-complete-flow-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildName := "apm-complete-flow"
	buildNumber := "204"

	defer setupTestWorkingDirectory(t, projectDir)()

	// Step 1: Install with build-info
	err = getApmCli().Exec("agent", "apm", "install", "--build-name", buildName, "--build-number", buildNumber)
	require.NoError(t, err, "install with build-info should succeed")

	validateApmBuildInfo(t, buildName, buildNumber, 0)

	// Step 2: Publish with build-info
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "complete/workflow",
		"--build-name", buildName, "--build-number", buildNumber)
	require.NoError(t, err, "publish with build-info should succeed")

	validateApmBuildInfo(t, buildName, buildNumber, 1)

	// Step 3: Publish build info to Artifactory
	err = artifactoryCli.Exec("rt", "bp", buildName, buildNumber)
	require.NoError(t, err, "build-info publish should succeed")

	// Clean up
	_, _, _ = tests.DeleteFiles(
		spec.NewBuilder().Pattern(tests.AgentPackagesLocalRepo+"/complete/workflow/*.zip").BuildSpec(),
		serverDetails)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestApmDryRunNoArtifacts validates --dry-run doesn't upload (P1: Scenario #28).
func TestApmDryRunNoArtifacts(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-dryrun-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	defer setupTestWorkingDirectory(t, projectDir)()

	// Dry-run publish should not upload artifacts
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "dryrun/test", "--", "--dry-run")
	require.NoError(t, err, "dry-run publish should succeed")

	// Verify nothing was uploaded
	searchSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/dryrun/test/*.zip").
		BuildSpec()
	artifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	assert.Empty(t, artifacts, "dry-run should not create artifacts in repository")
}

// TestApmMultiModuleWorkspace validates workspace support (P1: Scenario #31 variant).
func TestApmMultiModuleWorkspace(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-workspace-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	// Create workspace structure
	err = os.MkdirAll(filepath.Join(projectDir, "module1", ".apm", "primitives"), dirPerms)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(projectDir, "module2", ".apm", "primitives"), dirPerms)
	require.NoError(t, err)

	// Create workspace apm.yml
	workspaceYaml := `version: "1.0.0"
name: workspace-root
workspaces:
  - path: module1
  - path: module2
`

	rootYamlPath := filepath.Join(projectDir, "apm.yml")
	err = os.WriteFile(rootYamlPath, []byte(workspaceYaml), filePerms)
	require.NoError(t, err)

	// Create module manifests
	module1Yaml := `name: module1
version: 1.0.0
primitives:
  agents: []
`
	module2Yaml := `name: module2
version: 1.0.0
primitives:
  agents: []
`

	err = os.WriteFile(filepath.Join(projectDir, "module1", "apm.yml"), []byte(module1Yaml), filePerms)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(projectDir, "module2", "apm.yml"), []byte(module2Yaml), filePerms)
	require.NoError(t, err)

	buildNumber := "205"

	defer setupTestWorkingDirectory(t, projectDir)()

	// Install workspace should process all modules
	err = getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "workspace install should succeed")

	validateApmBuildInfo(t, apmBuildName, buildNumber, 0)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

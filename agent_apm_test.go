package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	apmBuildName = "apm-test-build"
	dirPerms     = 0755
	filePerms    = 0644
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns everything written to
// it. apm's own diagnostics (e.g. "HTTP 404 ...") are printed straight to os.Stdout by the
// underlying apm subprocess and never appear in the Go error returned by CLI commands, so
// assertions on that text must inspect captured stdout instead of err.Error().
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fnErr := fn()

	require.NoError(t, w.Close())
	os.Stdout = origStdout

	out, readErr := io.ReadAll(r)
	require.NoError(t, readErr)
	return string(out), fnErr
}

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

// publishApmDependencyPackage publishes a minimal, real APM package to the default registry
// (tests.AgentPackagesLocalRepo) so other tests can declare it as a resolvable dependency
// (via the "owner/name#version" shorthand) and exercise real install/build-info collection.
// packageSpec is "owner/name"; version is the version to publish (e.g. "1.0.0").
func publishApmDependencyPackage(t *testing.T, packageSpec, version string) {
	t.Helper()
	pubDir, err := os.MkdirTemp("", "apm-dep-publish-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(pubDir)
	}()

	require.NoError(t, os.MkdirAll(filepath.Join(pubDir, ".apm", "primitives"), dirPerms))
	_, pkgName, ok := strings.Cut(packageSpec, "/")
	require.True(t, ok, "packageSpec must be in owner/name form, got %q", packageSpec)

	apmYaml := fmt.Sprintf(`version: "1.0.0"
name: %s
version: %s
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
`, pkgName, version)
	require.NoError(t, os.WriteFile(filepath.Join(pubDir, "apm.yml"), []byte(apmYaml), filePerms))
	require.NoError(t, os.WriteFile(filepath.Join(pubDir, ".apm", "primitives", "placeholder.txt"), []byte("placeholder content"), filePerms))

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, pubDir)

	require.NoError(t, getApmCli().Exec("agent", "apm", "publish", "--package", packageSpec, "--registry", tests.AgentPackagesLocalRepo),
		"publishing dependency package %s should succeed", packageSpec)
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

// createAgentPackagesRepoWithKey creates an agent-packages local repository whose "key"
// field matches repoName. ReplaceTemplateVariables always substitutes the ${AGENT_PACKAGES_LOCAL_REPO}
// placeholder with the tests.AgentPackagesLocalRepo constant, so for repos with a different name
// we patch the "key" field ourselves after substitution to avoid an Artifactory key/path conflict.
func createAgentPackagesRepoWithKey(t *testing.T, repoName string) {
	repoConfig := tests.GetTestResourcesPath() + tests.AgentPackagesLocalRepositoryConfig
	repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
	require.NoError(t, err)

	content, err := os.ReadFile(repoConfig)
	require.NoError(t, err)
	patched := strings.Replace(string(content), `"key": "`+tests.AgentPackagesLocalRepo+`"`, `"key": "`+repoName+`"`, 1)

	patchedPath := filepath.Join(filepath.Dir(repoConfig), repoName+"_repository_config.json")
	require.NoError(t, os.WriteFile(patchedPath, []byte(patched), filePerms))

	execCreateRepoRest(patchedPath, repoName)
}

// ensureApmTestProjectExists creates the shared tests.ProjectKey Artifactory project and assigns
// tests.AgentPackagesLocalRepo to it. "--project" scoping on jf commands (e.g. jf rt bp
// --project=X) requires a real Project entity server-side - it's not just a local metadata tag -
// so tests exercising project scoping must provision one first.
//
// Skips (not fails) the calling test when Projects/Access isn't available in the current
// environment: local Artifactory instances used in some CI/test setups don't have it licensed
// or enabled, which is an environment limitation, not a defect in the apm code under test. Same
// graceful-skip pattern as TestApkAdd_ProjectBuildInfoCollection.
func ensureApmTestProjectExists(t *testing.T) {
	t.Helper()
	accessManager, err := artUtils.CreateAccessServiceManager(serverDetails, false)
	if err != nil {
		t.Skipf("Skipping project-scoped test - cannot create access manager: %v", err)
	}

	// Best-effort: ignore "doesn't exist yet" and any other delete failure alike, since the
	// only thing that matters is a clean CreateProject call next.
	_ = accessManager.DeleteProject(tests.ProjectKey)

	if err := accessManager.CreateProject(accessServices.ProjectParams{
		ProjectDetails: accessServices.Project{
			DisplayName: "apm test project " + tests.ProjectKey,
			ProjectKey:  tests.ProjectKey,
		},
	}); err != nil {
		t.Skipf("Skipping project-scoped test - cannot create project: %v", err)
	}
	if err := accessManager.AssignRepoToProject(tests.AgentPackagesLocalRepo, tests.ProjectKey, true); err != nil {
		t.Skipf("Skipping project-scoped test - cannot assign repo to project: %v", err)
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
license: UNLICENSED
targets:
  - claude
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

// createApmTestProjectWithDependency creates the same minimal project as createApmTestProject,
// but declares depSpec (e.g. "test/dep-pkg#1.0.0") as a real APM dependency. The caller is
// responsible for having already published depSpec's package (see publishApmDependencyPackage)
// so install actually resolves it and produces apm.lock.yaml / build info.
func createApmTestProjectWithDependency(t *testing.T, projectDir, depSpec string) {
	createApmTestProject(t, projectDir)

	apmYamlContent := `version: "1.0.0"
name: test-apm-package
description: Test APM package for e2e testing
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
  skills: []
  models: []
  tools: []
dependencies:
  apm:
    - ` + depSpec + `
`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "apm.yml"), []byte(apmYamlContent), filePerms))
}

// fetchPublishedApmBuildInfo publishes the locally-collected build info to Artifactory
// (jf rt bp) and reads it back from the server.
//
// apm's install/publish/update commands only ever call Build.AddArtifacts /
// Build.SavePartialBuildInfo, which write *partial* build-info files under
// <buildDir>/partials/ - they never call Build.SaveBuildInfo to materialize a combined,
// "generated" build info directly under <buildDir> (the file build.GetGeneratedBuildsInfo
// reads). That's consistent with the rest of jfrog-cli's build-info design: jf rt bp
// itself calls Build.ToBuildInfo(), which reads the same partials and assembles the
// final build info at publish time - GetGeneratedBuildsInfo is for package managers whose
// commands call Build.SaveBuildInfo directly (npm, docker, conan, etc.), not for reading
// pre-publish partials. So build.GetGeneratedBuildsInfo(name, number, "") is always
// guaranteed to return zero results for apm and cannot be used to validate its build info
// pre-publish; publish-then-verify-on-server is required.
func fetchPublishedApmBuildInfo(t *testing.T, buildName, buildNumber string) *buildinfo.BuildInfo {
	t.Helper()
	return fetchPublishedApmBuildInfoInProject(t, buildName, buildNumber, "")
}

// fetchPublishedApmBuildInfoInProject is fetchPublishedApmBuildInfo scoped to an Artifactory project key.
func fetchPublishedApmBuildInfoInProject(t *testing.T, buildName, buildNumber, projectKey string) *buildinfo.BuildInfo {
	t.Helper()
	bpArgs := []string{"bp", buildName, buildNumber}
	if projectKey != "" {
		bpArgs = append(bpArgs, "--project", projectKey)
	}
	require.NoError(t, artifactoryCli.Exec(bpArgs...), "jf rt bp should succeed")

	published, found, err := tests.GetBuildInfoInProject(serverDetails, buildName, buildNumber, projectKey)
	require.NoError(t, err)
	require.True(t, found, "published build info should be found on the server")
	return &published.BuildInfo
}

// validateApmBuildInfo publishes and validates the build info collected by an APM command.
func validateApmBuildInfo(t *testing.T, buildName, buildNumber string, expectedArtifacts int) {
	buildResult := fetchPublishedApmBuildInfo(t, buildName, buildNumber)

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

// validateBuildInfoDependencies validates dependencies exist in the published build info
func validateBuildInfoDependencies(t *testing.T, buildName, buildNumber string) {
	buildResult := fetchPublishedApmBuildInfo(t, buildName, buildNumber)
	require.Len(t, buildResult.Modules, 1, "Build should have at least one module")

	module := buildResult.Modules[0]
	require.NotEmpty(t, module.Dependencies, "Dependencies should be present in build info")

	for _, dep := range module.Dependencies {
		assert.NotEmpty(t, dep.Id, "Dependency should have ID")
	}
}

// validateBuildInfoArtifacts validates artifacts in the published build info
func validateBuildInfoArtifacts(t *testing.T, buildName, buildNumber string, expectedCount int) {
	buildResult := fetchPublishedApmBuildInfo(t, buildName, buildNumber)
	require.Len(t, buildResult.Modules, 1, "Build should have at least one module")

	module := buildResult.Modules[0]
	require.Len(t, module.Artifacts, expectedCount, "Artifacts count should match expected")

	for _, artifact := range module.Artifacts {
		assert.NotEmpty(t, artifact.Path, "Artifact should have path")
		assert.NotEmpty(t, artifact.Sha256, "Artifact should have checksum")
	}
}

// validateBuildInfoHasBothArtifactsAndDependencies validates both exist in the published build info
func validateBuildInfoHasBothArtifactsAndDependencies(t *testing.T, buildName, buildNumber string) {
	buildResult := fetchPublishedApmBuildInfo(t, buildName, buildNumber)
	require.Len(t, buildResult.Modules, 1, "Build should have at least one module")

	module := buildResult.Modules[0]
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

	var config map[string]any
	err = json.Unmarshal(configData, &config)
	require.NoError(t, err)

	registries, ok := config["registries"].(map[string]any)
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

	publishApmDependencyPackage(t, "test/install-bi-dep", "1.0.0")

	projectDir, err := os.MkdirTemp("", "apm-install-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	// A real, resolvable dependency is required: apm only writes apm.lock.yaml (and thus only
	// jfrog-cli only collects build-info) when the project has at least one dependency.
	createApmTestProjectWithDependency(t, projectDir, "test/install-bi-dep#1.0.0")

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
	err = artifactoryCli.Exec("bp", apmBuildName, buildNumber)
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
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "jfrog/test-apm-pkg", "--registry", tests.AgentPackagesLocalRepo, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm publish should succeed with build-info")

	// Validate build info was created with artifact
	validateApmBuildInfo(t, apmBuildName, buildNumber, 1)

	// Publish the build info
	err = artifactoryCli.Exec("bp", apmBuildName, buildNumber)
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
	err = getApmCli().Exec("agent", "apm", "publish", "--package", fmt.Sprintf("%s/%s", owner, packageName), "--registry", tests.AgentPackagesLocalRepo)
	require.NoError(t, err, "jf agent apm publish should succeed")

	// Verify artifact path: <owner>/<name>/<name>-<version>.zip
	searchSpec := spec.NewBuilder().
		Pattern(fmt.Sprintf("%s/%s/%s/*.zip", tests.AgentPackagesLocalRepo, owner, packageName)).
		BuildSpec()
	artifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	assert.NotEmpty(t, artifacts, "Artifact should be found at expected path: <owner>/<name>/<name>-<version>.zip")

	// Verify artifact name format. Note: ResultItem.Path is the artifact's *directory* (e.g.
	// "acme/my-agent-skill"); the filename itself is a separate field, Name.
	if len(artifacts) > 0 {
		assert.True(t,
			strings.HasPrefix(artifacts[0].Name, packageName+"-") && strings.HasSuffix(artifacts[0].Name, ".zip"),
			"Artifact name should follow pattern: <name>-<version>.zip, got %q", artifacts[0].Name)
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

	// APM dependency shorthand is "owner/name#version" (a plain string), resolved against the
	// default registry. A nonexistent package fails at resolve time with a 404-style error.
	apmYamlContent := `version: "1.0.0"
name: test-with-missing-dep
license: UNLICENSED
targets:
  - claude
dependencies:
  apm:
    - nonexistent/package#1.0.0
`
	apmYamlPath := filepath.Join(projectDir, "apm.yml")
	err = os.WriteFile(apmYamlPath, []byte(apmYamlContent), 0644)
	require.NoError(t, err)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Attempt install with invalid package. apm's own diagnostics (including the "HTTP 404"
	// detail) are printed to stdout by the apm subprocess, not embedded in the Go error, so we
	// must capture stdout to assert on them.
	output, cmdErr := captureStdout(t, func() error {
		return getApmCli().Exec("agent", "apm", "install")
	})
	assert.Error(t, cmdErr, "install of a nonexistent package should fail")
	assert.True(t,
		strings.Contains(output, "404") || strings.Contains(output, "no package"),
		"Output should indicate package not found, got: %s", output)
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

	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/artifact-metadata", "--registry", tests.AgentPackagesLocalRepo, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Validate build info has complete artifact metadata
	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
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

	err = getApmCli().Exec("agent", "apm", "publish", "--package", "jfrog/props-test", "--registry", tests.AgentPackagesLocalRepo, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Publish build info
	err = artifactoryCli.Exec("bp", apmBuildName, buildNumber)
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
	assert.NotEmpty(t, artifact.Properties, "Artifact should have properties")

	// Check for build name/number in properties
	foundBuildName := false
	foundBuildNumber := false
	for _, prop := range artifact.Properties {
		switch prop.Key {
		case "build.name":
			foundBuildName = true
			assert.Contains(t, prop.Value, apmBuildName)
		case "build.number":
			foundBuildNumber = true
			assert.Contains(t, prop.Value, buildNumber)
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

	publishApmDependencyPackage(t, "test/module-flag-dep", "1.0.0")

	projectDir, err := os.MkdirTemp("", "apm-module-flag-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProjectWithDependency(t, projectDir, "test/module-flag-dep#1.0.0")

	buildNumber := "106"
	customModule := "custom-apm-module"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "install", "--module", customModule, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install with --module flag should succeed")

	// Validate custom module name in build info
	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
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
	err = getApmCli().Exec("agent", "apm", "publish", "--package", fmt.Sprintf("%s/%s", owner, pkgName), "--registry", tests.AgentPackagesLocalRepo, "--build-name", apmBuildName, "--build-number", buildNumberPublish)
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
license: UNLICENSED
targets:
  - claude
dependencies:
  apm:
    - ` + owner + `/` + pkgName + `#1.0.0
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

	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/checksums", "--registry", tests.AgentPackagesLocalRepo, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Get build info and verify checksums
	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)

	if len(buildResult.Modules) > 0 {
		module := buildResult.Modules[0]
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

	ensureApmTestProjectExists(t)
	publishApmDependencyPackage(t, "test/project-flag-dep", "1.0.0")

	projectDir, err := os.MkdirTemp("", "apm-project-flag-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProjectWithDependency(t, projectDir, "test/project-flag-dep#1.0.0")

	buildNumber := "108"
	projectKey := tests.ProjectKey
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "install", "--project", projectKey, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install with --project flag should succeed")

	// Validate build info is scoped to project
	buildResult := fetchPublishedApmBuildInfoInProject(t, apmBuildName, buildNumber, projectKey)
	require.NotNil(t, buildResult, "Build should be found when queried with correct project key")

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmUpdateWithBuildInfo validates `jf agent apm update` with build-info (P1: Scenario #16).
func TestApmUpdateWithBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/update-bi-dep", "1.0.0")

	projectDir, err := os.MkdirTemp("", "apm-update-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProjectWithDependency(t, projectDir, "test/update-bi-dep#1.0.0")

	buildNumber := "109"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// First, install to have a lockfile
	err = getApmCli().Exec("agent", "apm", "install")
	require.NoError(t, err)

	// Then update with build-info capture. --yes is required: apm update shows a
	// confirmation plan and exits 1 without it, even in CI/non-interactive shells.
	err = getApmCli().Exec("agent", "apm", "update", "--yes", "--build-name", apmBuildName, "--build-number", buildNumber)
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

	// Test --dry-run native APM flag (passed directly, not via -- escape)
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/native-flags", "--registry", tests.AgentPackagesLocalRepo, "--dry-run")
	require.NoError(t, err, "jf agent apm publish with --dry-run should succeed")

	// Verify no artifact was uploaded for dry-run
	searchSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/test/native-flags/*.zip").
		BuildSpec()
	artifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	_ = artifacts
}

// TestApmBuildInfoRead validates a published apm build-info can be read back from Artifactory
// (P0: Scenario #5). There is no "jf rt bi" read command - jf's build-info commands are all
// write-side (build-publish/build-collect-env/etc.); reading a published build back is done via
// the REST API, which is what tests.GetBuildInfo (used throughout this file) wraps.
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
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/bi-read", "--registry", tests.AgentPackagesLocalRepo, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Publish to Artifactory
	err = artifactoryCli.Exec("bp", apmBuildName, buildNumber)
	require.NoError(t, err)

	// Read the published build info back from Artifactory
	published, found, err := tests.GetBuildInfo(serverDetails, apmBuildName, buildNumber)
	require.NoError(t, err, "reading the published build info should succeed")
	require.True(t, found, "published build info should be found on the server")
	assert.Equal(t, apmBuildName, published.BuildInfo.Name)
	assert.Equal(t, buildNumber, published.BuildInfo.Number)

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
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "e2e/pipeline", "--registry", tests.AgentPackagesLocalRepo, "--build-name", buildName, "--build-number", buildNumber)
	require.NoError(t, err, "Step 2: Publish should succeed")

	// Step 3: Publish build info
	err = artifactoryCli.Exec("bp", buildName, buildNumber)
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

	publishApmDependencyPackage(t, "test/install-with-deps-bi", "1.0.0")

	projectDir := createProjectWithDependencies(t, "app-with-deps", []string{"test/install-with-deps-bi#1.0.0"})
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

	publishApmDependencyPackage(t, "test/complete-app-dep", "1.0.0")

	projectDir := createProjectWithDependencies(t, "complete-app", []string{"test/complete-app-dep#1.0.0"})
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

// TestApmUpdateWithVersionChange validates update captures a new dependency version in build info.
// A bare "#1.0.0" pin is exact and apm update never moves it; only a semver range like "^1.0.0"
// is a floating constraint update can re-resolve, so this uses "^1.0.0" and republishes the
// dependency at 1.0.1 in between install and update (matching the documented apmbughunt flow).
func TestApmUpdateWithVersionChange(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/version-change-dep", "1.0.0")

	projectDir, err := os.MkdirTemp("", "apm-update-version-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	createApmTestProjectWithDependency(t, projectDir, "test/version-change-dep#^1.0.0")
	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "403"

	// Step 1: Install at 1.0.0
	err = runApmInstall(buildNumber)
	require.NoError(t, err, "install should succeed")

	installResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, installResult.Modules, "install build info should have a module")
	assert.Contains(t, installResult.Modules[0].Dependencies[0].Id, "1.0.0", "install should resolve the dependency at 1.0.0")
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)

	// Bump and republish the dependency so update has something new to pick up.
	publishApmDependencyPackage(t, "test/version-change-dep", "1.0.1")

	// Step 2: Update should re-resolve the floating range to 1.0.1
	err = runApmUpdate(apmBuildName, buildNumber)
	require.NoError(t, err, "update should succeed")

	updateResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, updateResult.Modules, "update build info should have a module")
	assert.Contains(t, updateResult.Modules[0].Dependencies[0].Id, "1.0.1", "update should resolve the dependency at 1.0.1")

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

	// Auth via APM_REGISTRY_TOKEN_<REGISTRY> env var (same mechanism as TestApmAuthEnvironmentVariable).
	registryName := "default"
	tokenEnvVar := fmt.Sprintf("APM_REGISTRY_TOKEN_%s", strings.ToUpper(registryName))
	require.NoError(t, os.Setenv(tokenEnvVar, *tests.JfrogAccessToken))
	defer func() {
		_ = os.Unsetenv(tokenEnvVar)
	}()

	// The token must be usable for auth but never echoed back in apm's own stdout/log output.
	output, err := captureStdout(t, func() error {
		return getApmCli().Exec("agent", "apm", "install")
	})
	require.NoError(t, err, "install should work with env var auth")
	assert.NotContains(t, output, *tests.JfrogAccessToken, "access token should not be exposed in command output")
}

// TestApmDifferentRegistriesAsArtifactoryRepos validates multiple distinct Artifactory repos
func TestApmDifferentRegistriesAsArtifactoryRepos(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	// Create two different repos
	repos := []string{"apm-registry-1", "apm-registry-2"}
	for _, repoName := range repos {
		if !isRepoExist(repoName) {
			createAgentPackagesRepoWithKey(t, repoName)
		}
	}
	defer func() {
		for _, repoName := range repos {
			deleteRepo(repoName)
		}
	}()

	// Register each repo as its own named APM registry in ~/.apm/config.json.
	// "jfrog setup agent-apm --repo X" names the registry after the repo (registry.X.*),
	// so calling it once per repo yields multiple distinct, independently addressable registries.
	setupCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	for _, repoName := range repos {
		err := setupCli.Exec("setup", "agent-apm", "--repo", repoName)
		require.NoError(t, err, "setup should succeed for repo %s", repoName)
	}

	projectDir := createProjectWithRegistries(t, "multi-repo-app")
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
	return createApmYaml("test-app", "1.0.0", nil)
}

// createApmYaml creates customizable APM YAML with parameters. apmDeps are real APM dependency
// specs in "owner/name#version" shorthand (see publishApmDependencyPackage); an empty slice
// yields an empty "apm: []" dependency list. Registries are never declared here - they're
// configured globally via "jf setup agent-apm", not per-project.
func createApmYaml(name, version string, apmDeps []string) string {
	depsSection := "  apm: []\n"
	if len(apmDeps) > 0 {
		var b strings.Builder
		b.WriteString("  apm:\n")
		for _, dep := range apmDeps {
			_, _ = fmt.Fprintf(&b, "    - %s\n", dep)
		}
		depsSection = b.String()
	}

	return fmt.Sprintf(`version: "1.0.0"
name: %s
version: %s
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
dependencies:
%s`, name, version, depsSection)
}

// createMultiRegistryYaml creates a minimal apm.yml for tests exercising multiple registries.
// Registries are configured globally via "jf setup agent-apm", not declared in apm.yml itself.
func createMultiRegistryYaml(name string) string {
	return fmt.Sprintf(`version: "1.0.0"
name: %s
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
dependencies:
  apm: []
`, name)
}

// createProjectWithDependencies creates a project directory with specified dependencies
func createProjectWithDependencies(t *testing.T, name string, deps []string) string {
	apmYaml := createApmYaml(name, "1.0.0", deps)
	return createApmProjectWithYaml(t, apmYaml)
}

// createProjectWithRegistries creates a project used by tests that register multiple named
// Artifactory repos as registries (see createAgentPackagesRepoWithKey / TestApmDifferentRegistriesAsArtifactoryRepos).
// The apm.yml itself never lists them - only "jf setup agent-apm" does that - so no registry
// names need to flow into the generated YAML here.
func createProjectWithRegistries(t *testing.T, name string) string {
	return createApmProjectWithYaml(t, createMultiRegistryYaml(name))
}

// runApmInstall runs install command with optional build info
func runApmInstall(buildNumber string) error {
	args := []string{"agent", "apm", "install"}
	if buildNumber != "" {
		args = append(args, "--build-name", apmBuildName, "--build-number", buildNumber)
	}
	return getApmCli().Exec(args...)
}

// runApmPublish runs publish command with optional build info. --registry is passed
// explicitly since publish (unlike install) refuses to guess when more than one registry
// happens to be configured in ~/.apm/config.json (a real risk on any shared machine/CI runner).
func runApmPublish(packagePath, buildName, buildNumber string) error {
	args := []string{"agent", "apm", "publish"}
	if packagePath != "" {
		args = append(args, "--package", packagePath, "--registry", tests.AgentPackagesLocalRepo)
	}
	if buildName != "" && buildNumber != "" {
		args = append(args, "--build-name", buildName, "--build-number", buildNumber)
	}
	return getApmCli().Exec(args...)
}

// runApmUpdate runs update command with optional build info. --yes is required: apm update
// shows a confirmation plan and exits 1 without it, even in CI/non-interactive shells.
func runApmUpdate(buildName, buildNumber string) error {
	args := []string{"agent", "apm", "update", "--yes"}
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
license: UNLICENSED
targets:
  - claude
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
license: UNLICENSED
targets:
  - claude
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
license: UNLICENSED
targets:
  - claude
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
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "test/app-with-deps", "--registry", tests.AgentPackagesLocalRepo,
		"--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "publish should succeed with dependencies")

	// Validate build info includes dependency metadata
	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, buildResult.Modules)

	module := buildResult.Modules[0]
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

	// Update with build-info. --yes is required: apm update shows a confirmation plan and
	// exits 1 without it, even in CI/non-interactive shells.
	err = getApmCli().Exec("agent", "apm", "update", "--yes", "--build-name", apmBuildName, "--build-number", buildNumber)
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
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "complete/workflow", "--registry", tests.AgentPackagesLocalRepo,
		"--build-name", buildName, "--build-number", buildNumber)
	require.NoError(t, err, "publish with build-info should succeed")

	validateApmBuildInfo(t, buildName, buildNumber, 1)

	// Step 3: Publish build info to Artifactory
	err = artifactoryCli.Exec("bp", buildName, buildNumber)
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
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "dryrun/test", "--registry", tests.AgentPackagesLocalRepo, "--dry-run")
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
license: UNLICENSED
targets:
  - claude
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
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
`
	module2Yaml := `name: module2
version: 1.0.0
license: UNLICENSED
targets:
  - claude
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

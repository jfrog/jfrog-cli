package main

import (
	"archive/zip"
	"crypto/md5"  // #nosec G501 -- checksum verification against Artifactory's own reported MD5, not security-sensitive
	"crypto/sha1" // #nosec G505 -- checksum verification against Artifactory's own reported SHA1, not security-sensitive
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	buildUtils "github.com/jfrog/jfrog-cli-core/v2/common/build"
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

// computeFileSHA1 and computeFileMD5 mirror apk_test.go's computeFileSHA256 (same package) for
// the other two checksums build-info round-trip tests need to independently verify.
func computeFileSHA1(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path) // #nosec G304 -- path is always a test-controlled temp download destination
	require.NoError(t, err, "open file for SHA1: %s", path)
	defer func() { require.NoError(t, f.Close()) }()
	h := sha1.New() // #nosec G401 -- checksum verification, not a security-relevant crypto use
	_, err = io.Copy(h, f)
	require.NoError(t, err, "compute SHA1 for: %s", path)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func computeFileMD5(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path) // #nosec G304 -- path is always a test-controlled temp download destination
	require.NoError(t, err, "open file for MD5: %s", path)
	defer func() { require.NoError(t, f.Close()) }()
	h := md5.New() // #nosec G401 -- checksum verification, not a security-relevant crypto use
	_, err = io.Copy(h, f)
	require.NoError(t, err, "compute MD5 for: %s", path)
	return fmt.Sprintf("%x", h.Sum(nil))
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

// publishApmDependencyPackage publishes a minimal, real APM package, always at version 1.0.0, to
// the default registry (tests.AgentPackagesLocalRepo) so other tests can declare it as a
// resolvable dependency (via the "owner/name#1.0.0" shorthand) and exercise real
// install/build-info collection. packageSpec is "owner/name".
func publishApmDependencyPackage(t *testing.T, packageSpec string) {
	t.Helper()
	publishApmDependencyPackageToRegistry(t, packageSpec, tests.AgentPackagesLocalRepo)
}

// publishApmDependencyPackageToRegistry is publishApmDependencyPackage targeting a specific,
// already-configured registry name (e.g. one of several distinct repos set up via
// "jf setup apm --repo <name>"), instead of always the default tests.AgentPackagesLocalRepo.
// Always publishes at version 1.0.0 - no caller has ever needed a different one.
func publishApmDependencyPackageToRegistry(t *testing.T, packageSpec, registryName string) {
	t.Helper()
	pubDir, err := os.MkdirTemp("", "apm-dep-publish-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(pubDir)
	}()

	require.NoError(t, os.MkdirAll(filepath.Join(pubDir, ".apm", "primitives"), dirPerms))
	_, pkgName, ok := strings.Cut(packageSpec, "/")
	require.True(t, ok, "packageSpec must be in owner/name form, got %q", packageSpec)

	apmYaml := fmt.Sprintf(`name: %s
version: 1.0.0
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
`, pkgName)
	require.NoError(t, os.WriteFile(filepath.Join(pubDir, "apm.yml"), []byte(apmYaml), filePerms))
	require.NoError(t, os.WriteFile(filepath.Join(pubDir, ".apm", "primitives", "placeholder.txt"), []byte("placeholder content"), filePerms))

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, pubDir)

	require.NoError(t, getApmCli().Exec("agent", "apm", "publish", "--package", packageSpec, "--registry", registryName),
		"publishing dependency package %s to registry %s should succeed", packageSpec, registryName)
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
	require.NoError(t, os.WriteFile(patchedPath, []byte(patched), filePerms)) // #nosec G703 -- repoName is always one of this test's own hardcoded literals, not external input

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
	err := setupCli.Exec("setup", "apm", "--repo", tests.AgentPackagesLocalRepo)
	require.NoError(t, err, "jf setup apm should succeed")
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
// apm's install/publish commands only ever call Build.AddArtifacts /
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

// readLocalApmPartialBuildInfo reads locally-collected build info directly via
// Build.ToBuildInfo(), which assembles it from partial files without touching the server or
// clearing anything - the same mechanism pnpm_test.go/npm_test.go use to validate build info
// collection. Unlike fetchPublishedApmBuildInfo, this must be used for INTERMEDIATE checks
// within a multi-step test: "jf rt bp" calls Build.Clean() after a successful publish, wiping
// local partials for that exact build name/number. Calling fetchPublishedApmBuildInfo (or any
// validate* built on it) more than once for the same build/number silently loses whatever an
// earlier step wrote - confirmed live: a dependency captured after install disappeared from a
// later "has both artifacts and dependencies" check, once an intervening bp call for that same
// build/number had already run and cleared it. Reserve the server round-trip for a single,
// final check per build/number.
func readLocalApmPartialBuildInfo(t *testing.T, buildName, buildNumber string) *buildinfo.BuildInfo {
	t.Helper()
	buildInfoService := buildUtils.CreateBuildInfoService()
	apmBuild, err := buildInfoService.GetOrCreateBuildWithProject(buildName, buildNumber, "")
	require.NoError(t, err)
	bi, err := apmBuild.ToBuildInfo()
	require.NoError(t, err)
	return bi
}

// apmRegistryURL builds the real registry URL for repoName, matching exactly what
// AgentPackagesBaseURL in jfrog-cli-artifactory constructs from serverDetails
// (<ArtifactoryUrl>/api/agentpackages/<repo>/). A registry declared in apm.yml's own
// registries: block is used by apm as its literal API base URL for that registry - not merely
// matched by host for credential discovery - so it must be this exact form, not just any URL on
// the right host, or apm's own HTTP requests 404/403 against the wrong path.
func apmRegistryURL(repoName string) string {
	return strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/agentpackages/" + repoName + "/"
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

	// Dependency checksums come from the same HEAD-based resolution as artifact checksums (see
	// resolveChecksumsByHead in jfrog-cli-artifactory), so all three are required here too, not
	// merely the ID.
	for _, dep := range module.Dependencies {
		assert.NotEmpty(t, dep.Id, "Dependency should have ID")
		assert.NotEmpty(t, dep.Sha256, "Dependency should have SHA256")
		assert.Len(t, dep.Sha1, 40, "Dependency should have a 40 hex-character SHA1")
		assert.Len(t, dep.Md5, 32, "Dependency should have a 32 hex-character MD5")
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
		assert.NotEmpty(t, artifact.Sha256, "Artifact should have SHA256")
		assert.Len(t, artifact.Sha1, 40, "Artifact should have a 40 hex-character SHA1")
		assert.Len(t, artifact.Md5, 32, "Artifact should have a 32 hex-character MD5")
	}
}

// validateBuildInfoHasBothArtifactsAndDependencies validates both exist in the published build
// info, with checksums on each - not just presence of the two lists.
func validateBuildInfoHasBothArtifactsAndDependencies(t *testing.T, buildName, buildNumber string) {
	buildResult := fetchPublishedApmBuildInfo(t, buildName, buildNumber)
	require.Len(t, buildResult.Modules, 1, "Build should have at least one module")

	module := buildResult.Modules[0]
	require.NotEmpty(t, module.Dependencies, "Build info should have dependencies")
	require.NotEmpty(t, module.Artifacts, "Build info should have artifacts")

	for _, dep := range module.Dependencies {
		assert.NotEmpty(t, dep.Sha256, "Dependency should have SHA256")
	}
	for _, artifact := range module.Artifacts {
		assert.NotEmpty(t, artifact.Sha256, "Artifact should have SHA256")
	}
}

// TestApmSetupAndConfig validates APM setup with apm config file persistence (P0: Scenario #1).
// apmRegistryEntry mirrors one entry under ~/.apm/config.json's "registries" map. Default is
// only ever present (and true) on whichever registry "jf setup apm" most recently
// configured - apm's own config command clears it from any previously-default entry, so at most
// one registry has Default == true at a time.
type apmRegistryEntry struct {
	URL     string `json:"url"`
	Token   string `json:"token"`
	Default bool   `json:"default"`
}

// readApmRegistries parses ~/.apm/config.json's registries map.
func readApmRegistries(t *testing.T) map[string]apmRegistryEntry {
	t.Helper()
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	configData, err := os.ReadFile(filepath.Join(homeDir, ".apm", "config.json"))
	require.NoError(t, err)

	var config struct {
		Registries map[string]apmRegistryEntry `json:"registries"`
	}
	require.NoError(t, json.Unmarshal(configData, &config))
	return config.Registries
}

func TestApmSetupAndConfig(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	apmConfigPath := filepath.Join(homeDir, ".apm", "config.json")

	// First setup call (use correct CLI prefix: jfrog, not jfrog rt)
	setupCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	err = setupCli.Exec("setup", "apm", "--repo", tests.AgentPackagesLocalRepo)
	require.NoError(t, err, "jf setup apm should succeed")

	assert.FileExists(t, apmConfigPath, "APM config file should be created")

	// Verify --repo maps to an actual registry entry (not just "some registries exist"), with a
	// URL that references the repo and a token, and that it's the default registry.
	registries := readApmRegistries(t)
	primary, ok := registries[tests.AgentPackagesLocalRepo]
	require.True(t, ok, "registries should contain an entry named after --repo (%s)", tests.AgentPackagesLocalRepo)
	assert.Contains(t, primary.URL, tests.AgentPackagesLocalRepo, "registry URL should reference the configured repo")
	assert.NotEmpty(t, primary.Token, "registry entry should have a token")
	assert.True(t, primary.Default, "the just-configured repo should be the default registry")

	// Second setup call against a DIFFERENT repo should flip the default to it, and clear
	// Default from the previously-default entry - proving "default" tracks the most recently
	// configured repo, not just whichever was configured first.
	//
	// ~/.apm/config.json is a real user-global file, not scoped per test, and several other
	// tests in this file install without an explicit --registry (relying on default
	// resolution) - so restore tests.AgentPackagesLocalRepo as the default before returning,
	// regardless of how this test's own assertions turn out.
	secondRepo := "apm-setup-config-test-repo"
	if !isRepoExist(secondRepo) {
		createAgentPackagesRepoWithKey(t, secondRepo)
	}
	defer deleteRepo(secondRepo)
	defer func() {
		_ = setupCli.Exec("setup", "apm", "--repo", tests.AgentPackagesLocalRepo)
	}()

	err = setupCli.Exec("setup", "apm", "--repo", secondRepo)
	require.NoError(t, err, "jf setup apm should succeed against a second, different repo")

	registries = readApmRegistries(t)
	second, ok := registries[secondRepo]
	require.True(t, ok, "registries should now contain an entry named after the second --repo (%s)", secondRepo)
	assert.Contains(t, second.URL, secondRepo, "second registry URL should reference the second repo")
	assert.True(t, second.Default, "the most recently configured repo should be the default registry")

	if first, ok := registries[tests.AgentPackagesLocalRepo]; ok {
		assert.False(t, first.Default, "the previously-default registry should no longer be marked default")
	}

	// Verify idempotency - re-running setup for the same (now non-default) repo should still
	// succeed and flip default back to it.
	err = setupCli.Exec("setup", "apm", "--repo", tests.AgentPackagesLocalRepo)
	require.NoError(t, err, "jf setup apm should be idempotent")

	registries = readApmRegistries(t)
	assert.True(t, registries[tests.AgentPackagesLocalRepo].Default, "re-running setup for the primary repo should make it the default again")
}

// TestApmInstallWithBuildInfo validates `jf agent apm install` with build-info capture (P0: Scenario #13).
func TestApmInstallWithBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/install-bi-dep")

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

// TestApmAuthEnvVarBehavior validates two distinct env-var-auth scenarios (both via
// APM_REGISTRY_TOKEN_<REGISTRY>, agent/apm/common/apmenv.go) as subtests sharing one
// initApmTest/cleanApmTest cycle instead of two separate test functions.
func TestApmAuthEnvVarBehavior(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	// The registry name apm actually knows about is the repo key itself ("cli-agent-packages-local"),
	// not a literal name "default" - jf setup apm calls ConfigureApmRegistryPersistent(repoName),
	// which writes registry.<repoName>.{url,token,default} into ~/.apm/config.json using repoName
	// verbatim. apm sanitizes that name into its env var form the same way jf does
	// (sanitizeApmEnvName in apmenv.go: uppercase, "-"/"." -> "_"). Using any other name here (e.g.
	// the earlier "default") produces an env var apm never looks at for this registry, so a "wrong
	// token" set under that name is silently never consulted - confirmed live, this is exactly why
	// the wrong-token subtest below kept passing for the wrong reason before this fix.
	registryName := tests.AgentPackagesLocalRepo
	tokenEnvVar := fmt.Sprintf("APM_REGISTRY_TOKEN_%s", strings.ToUpper(strings.ReplaceAll(registryName, "-", "_")))

	t.Run("wrong token is honored instead of silently overridden", func(t *testing.T) {
		// jf's own BuildApmEnv (agent/apm/common/apmenv.go) always auto-injects
		// APM_REGISTRY_TOKEN_<NAME> from the configured server before running apm - a plain
		// install with the CORRECT token set ourselves would succeed identically whether or not
		// jf actually reads our value or silently substitutes its own, so that alone wouldn't
		// prove anything. Setting an intentionally WRONG token instead only fails if jf genuinely
		// leaves our value alone (injectRegistryCredentialEnv's "respecting existing value" branch)
		// instead of overriding it with the correct one - which is exactly what this proves.
		//
		// (A debug-log assertion on "credential env var already set" was tried here first, but
		// log.SetDefaultLogger() - which reads JFROG_CLI_LOG_LEVEL - is only called from
		// main(), not execMain(); this test harness invokes execMain() directly in-process, so
		// the log level set via os.Setenv here is never actually picked up. Confirmed live: the
		// log line never appeared no matter what level was set.)
		publishApmDependencyPackage(t, "test/auth-env-wrong-token-dep")

		// Belt and braces: apm's own docs say an env var token outranks ~/.apm/config.json's
		// stored one, but remove the stored token for this registry anyway so there is no valid
		// fallback credential at all - the only credential apm can possibly use is the wrong one
		// set below. Restored afterward by re-running jf setup apm (initApmConfig), which
		// every other test in this file also depends on having a correctly configured registry.
		require.NoError(t, exec.Command("apm", "config", "unset", fmt.Sprintf("registry.%s.token", registryName)).Run(), // #nosec G204 -- fixed argv, no user input
			"removing the stored registry token should succeed")
		defer initApmConfig(t)

		projectDir, err := os.MkdirTemp("", "apm-auth-env-wrong-*")
		require.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(projectDir)
		}()
		createApmTestProjectWithDependency(t, projectDir, "test/auth-env-wrong-token-dep#1.0.0")
		defer setupTestWorkingDirectory(t, projectDir)()

		require.NoError(t, os.Setenv(tokenEnvVar, "definitely-not-a-real-token"))
		defer func() {
			_ = os.Unsetenv(tokenEnvVar)
		}()

		err = getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", "103")
		assert.Error(t, err, "install should fail when the pre-set (invalid) token env var is honored instead of silently overridden")
	})

	t.Run("correct token is not exposed in output", func(t *testing.T) {
		projectDir := createApmProjectWithYaml(t, getBasicApmYaml())
		defer func() {
			_ = os.RemoveAll(projectDir)
		}()
		defer setupTestWorkingDirectory(t, projectDir)()

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
	})
}

// TestApmMissingCredentials validates that install fails when there is no registry to discover
// at all - not, as the name might suggest, because credentials are generically "missing". apm
// always gets its actual token from jf's own configured server (BuildApmEnv in
// jfrog-cli-artifactory), regardless of ~/.apm/config.json; that file (and apm.yml's own
// registries: block) only supply the registry NAME+URL to route that token through. With
// neither source present, BuildApmEnv fails before credentials are ever considered - confirmed
// here by asserting on its exact error text ("no APM registry found"), not just a non-nil error,
// so this test can't silently start passing for an unrelated reason.
// See TestApmInstallSucceedsWithRegistryDeclaredInApmYml for the complementary case: apm.yml's
// own registries: block is sufficient on its own, even with ~/.apm/config.json absent.
func TestApmMissingCredentials(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-no-creds-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir) // apm.yml here declares no registries: block

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Remove ~/.apm/config.json - with apm.yml declaring no registries: block either, this
	// leaves BuildApmEnv nothing to discover a registry from.
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	apmConfigPath := filepath.Join(homeDir, ".apm", "config.json")
	err = os.Remove(apmConfigPath)
	if err != nil && !os.IsNotExist(err) {
		require.NoError(t, err)
	}
	defer initApmConfig(t) // restore ~/.apm/config.json for later tests regardless of outcome

	// Unset any auth env vars
	for _, envVar := range os.Environ() {
		if strings.Contains(envVar, "APM_REGISTRY") {
			key := strings.Split(envVar, "=")[0]
			_ = os.Unsetenv(key)
		}
	}

	// Attempt install with no registry source available.
	err = getApmCli().Exec("agent", "apm", "install")
	require.Error(t, err, "jf agent apm install without a discoverable registry should fail")
	assert.Contains(t, err.Error(), "no APM registry found",
		"the failure should specifically be 'no registry found', not some unrelated error")
}

// apmRegistryToken reads registry.<repoName>.token out of ~/.apm/config.json, the same file
// ConfigureApmRegistryPersistent writes to via `apm config set`. Returns "" if the file or the
// registry entry doesn't exist.
func apmRegistryToken(t *testing.T, repoName string) string {
	t.Helper()
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(homeDir, ".apm", "config.json")) // #nosec G304 -- fixed, test-controlled path
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		require.NoError(t, err)
	}
	var cfg struct {
		Registries map[string]struct {
			Token string `json:"token"`
		} `json:"registries"`
	}
	require.NoError(t, json.Unmarshal(data, &cfg))
	return cfg.Registries[repoName].Token
}

// TestApmAuthWithUsernamePassword validates BuildRegistryEntry's Priority-2 path
// (agent/apm/common/apmenv.go): when the configured jf server has no AccessToken - only
// User+Password - jf must mint a brand-new Artifactory access token itself and write THAT into
// ~/.apm/config.json, rather than ever embedding the raw password. This is the one auth path in
// the whole APM registry flow with no other e2e coverage: every other test in this file
// configures the "default" server via --access-token and so only ever exercises Priority-1
// (use the existing token as-is).
//
// To force Priority-2 for real (not just in a mocked unit test) without needing every
// environment this suite runs in to hand out a plaintext platform password, this reconfigures
// "default" with --user/--password derived from the current access token itself (Artifactory
// accepts a token as a Basic Auth password) via tests.SetBasicAuthFromAccessToken. `jf setup`
// always calls config.GetSpecificConfig with excludeRefreshableTokens=true (buildtools/cli.go),
// which - per excludeRefreshableTokensFromDetails - strips the AccessToken jf's own config layer
// auto-mints alongside User+Password back out again before BuildRegistryEntry ever sees
// serverDetails. So this hits the real Priority-2 branch, not a contrived edge case.
func TestApmAuthWithUsernamePassword(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	repoName := tests.AgentPackagesLocalRepo
	tokenBeforeSwitch := apmRegistryToken(t, repoName)
	require.NotEmpty(t, tokenBeforeSwitch, "initApmTest/initApmConfig should have already written a token for %s", repoName)

	// Switch "default" to User+Password only, remembering the original mode to restore it.
	origAccessToken := *tests.JfrogAccessToken
	origUser, origPassword := tests.SetBasicAuthFromAccessToken()
	defer func() {
		*tests.JfrogUser, *tests.JfrogPassword, *tests.JfrogAccessToken = origUser, origPassword, origAccessToken
		createJfrogHomeConfig(t, true) // restore "default" to its original access-token mode
		initApmConfig(t)               // re-run `jf setup apm` so later tests get a valid token again
	}()
	// Recreate "default" via the same add-with-explicit-URL helper used everywhere else in this
	// file, rather than `config edit` without --url: edit's URL-preservation behavior isn't
	// guaranteed, and re-adding keeps this test's setup identical to every other server-config
	// switch in this file.
	*tests.JfrogAccessToken = ""
	createJfrogHomeConfig(t, true)

	// Force a fresh mint: without this, BuildRegistryEntry's own Priority-1 check
	// (serverDetails.AccessToken != "") would never even run, since the OLD token is still
	// sitting in ~/.apm/config.json from initApmConfig - but that's a stale value ConfigureApmRegistryPersistent
	// is about to overwrite anyway, not something BuildRegistryEntry reads back to decide its
	// own priority. Removing it first just makes the "did a fresh token actually get minted"
	// assertion below unambiguous.
	require.NoError(t, exec.Command("apm", "config", "unset", fmt.Sprintf("registry.%s.token", repoName)).Run()) // #nosec G204 -- fixed argv

	require.NoError(t,
		coreTests.NewJfrogCli(execMain, "jfrog setup", "").Exec("apm", "--repo", repoName),
		"jf setup apm should succeed when the server only has User+Password configured")

	mintedToken := apmRegistryToken(t, repoName)
	require.NotEmpty(t, mintedToken, "jf setup apm should have written a freshly minted token")
	assert.NotEqual(t, tokenBeforeSwitch, mintedToken, "the token should be freshly minted, not the stale one left over from access-token mode")
	assert.Equal(t, 2, strings.Count(mintedToken, "."), "a real Artifactory access token is a JWT (header.payload.signature); a base64(user:pass) blob would not have this shape")
	assert.NotEqual(t, basicAuthBase64(*tests.JfrogUser, *tests.JfrogPassword), mintedToken, "the minted token must not just be the raw credentials re-encoded")

	// Prove the minted token isn't just well-formed but actually authenticates: publish and
	// install a real package through it, end to end.
	publishApmDependencyPackage(t, "test/auth-username-password-dep")
	projectDir, err := os.MkdirTemp("", "apm-auth-userpass-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(projectDir) }()
	createApmTestProjectWithDependency(t, projectDir, "test/auth-username-password-dep#1.0.0")
	defer setupTestWorkingDirectory(t, projectDir)()

	require.NoError(t, getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", "104"),
		"install should succeed authenticating with the freshly minted access token")
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

func basicAuthBase64(user, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
}

// TestApmRegistriesDeclaredInApmYml validates that apm.yml's own registries: block (a url: only -
// see manifest.go's ManifestRegistry - matched to jf's configured server by host, via
// discoverMatchingRegistries) is sufficient on its own for registry discovery: with a single
// entry and ~/.apm/config.json entirely absent, and with multiple entries declared at once
// alongside a present config.json. jf still injects the actual token from its own configured
// server (serverDetails) in both cases; apm.yml never carries a token itself, only the name->URL
// mapping that tells jf which registry name to inject that token under. Both cases share one
// initApmTest/cleanApmTest cycle as subtests rather than two separate test functions.
func TestApmRegistriesDeclaredInApmYml(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	cases := []struct {
		name            string
		registryNames   []string
		removeApmConfig bool
	}{
		{
			name:            "single registry, config.json absent",
			registryNames:   []string{tests.AgentPackagesLocalRepo},
			removeApmConfig: true,
		},
		{
			name:          "multiple registries, config.json present",
			registryNames: []string{"registry-one", "registry-two"},
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var registriesYaml strings.Builder
			for _, name := range tc.registryNames {
				_, _ = fmt.Fprintf(&registriesYaml, "  %s:\n    url: \"%s\"\n", name, apmRegistryURL(tests.AgentPackagesLocalRepo))
			}
			apmYaml := fmt.Sprintf(`name: registry-in-yaml-project
version: 1.0.0
license: UNLICENSED
targets:
  - claude
registries:
%sdependencies:
  apm: []
`, registriesYaml.String())

			projectDir := createApmProjectWithYaml(t, apmYaml)
			defer func() {
				_ = os.RemoveAll(projectDir)
			}()
			defer setupTestWorkingDirectory(t, projectDir)()

			if tc.removeApmConfig {
				// The registry above must be discoverable purely from apm.yml's own registries:
				// block, matched by host to the configured jf server.
				homeDir, err := os.UserHomeDir()
				require.NoError(t, err)
				apmConfigPath := filepath.Join(homeDir, ".apm", "config.json")
				removeErr := os.Remove(apmConfigPath)
				if removeErr != nil && !os.IsNotExist(removeErr) {
					require.NoError(t, removeErr)
				}
				defer initApmConfig(t) // restore for the next subtest/test regardless of outcome
			}

			buildNumber := fmt.Sprintf("20%d", i)
			err := getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumber)
			require.NoError(t, err, "install should succeed using apm.yml's own registries: block")

			validateApmBuildInfo(t, apmBuildName, buildNumber, 0)
			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
		})
	}
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
	require.NotEmpty(t, module.Artifacts, "build info should have an artifact")
	for _, artifact := range module.Artifacts {
		// Verify metadata fields are present. Checksum correctness (not just presence) is
		// covered separately by TestApmChecksumsInBuildInfo's download-and-recompute round trip.
		assert.NotEmpty(t, artifact.Path, "Artifact path should be present")
		assert.NotEmpty(t, artifact.Type, "Artifact type should be present")
		assert.NotEmpty(t, artifact.Sha256, "Artifact SHA256 should be present")
		assert.Len(t, artifact.Sha1, 40, "Artifact should have a 40 hex-character SHA1")
		assert.Len(t, artifact.Md5, 32, "Artifact should have a 32 hex-character MD5")
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

	publishApmDependencyPackage(t, "test/module-flag-dep")

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
// TestApmChecksumsInBuildInfo validates that build info's checksums are not merely present with
// the right format, but actually correct. A well-formed-but-wrong checksum (e.g. from a bug that
// happens to produce a same-shaped value) would pass a presence/length-only check, so this
// downloads the published artifact back from Artifactory and independently recomputes SHA256,
// SHA1, and MD5 locally, then asserts build info's reported values match exactly - the same
// round-trip principle as apk_test.go's TestApkUpload_ChecksumRoundTrip, extended to cross-check
// against build info's own claims rather than only comparing two local files.
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

	owner, pkgName := "test", "checksums"
	err = getApmCli().Exec("agent", "apm", "publish", "--package", owner+"/"+pkgName, "--registry", tests.AgentPackagesLocalRepo, "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err)

	// Get build info. Artifactory always computes and returns all three checksums together for
	// a stored artifact (the HEAD lookup apm's checksum resolution uses - see
	// resolveChecksumsByHead in jfrog-cli-artifactory - reads X-Checksum-Sha1/Sha256/Md5 off the
	// same response), so all three are required, not merely "present if available".
	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, buildResult.Modules, "build info should have a module")

	module := buildResult.Modules[0]
	require.NotEmpty(t, module.Artifacts, "build info should have an artifact")

	// Download the published artifact and independently recompute its checksums.
	downloadDir, err := os.MkdirTemp("", "apm-checksums-download-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(downloadDir)
	}()
	artifactPattern := fmt.Sprintf("%s/%s/%s/*.zip", tests.AgentPackagesLocalRepo, owner, pkgName)
	require.NoError(t, artifactoryCli.Exec("dl", artifactPattern, downloadDir+"/", "--flat"))

	downloadedFiles, err := os.ReadDir(downloadDir)
	require.NoError(t, err)
	require.Len(t, downloadedFiles, 1, "exactly one artifact should have been downloaded")
	downloadedPath := filepath.Join(downloadDir, downloadedFiles[0].Name())

	actualSha256 := computeFileSHA256(t, downloadedPath)
	actualSha1 := computeFileSHA1(t, downloadedPath)
	actualMd5 := computeFileMD5(t, downloadedPath)

	for _, artifact := range module.Artifacts {
		assert.Equal(t, actualSha256, artifact.Sha256, "build info SHA256 should match the actual downloaded artifact")
		assert.Equal(t, actualSha1, artifact.Sha1, "build info SHA1 should match the actual downloaded artifact")
		assert.Equal(t, actualMd5, artifact.Md5, "build info MD5 should match the actual downloaded artifact")
	}

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmProjectFlag validates --project flag for project isolation (P1: Scenario #27).
func TestApmProjectFlag(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	ensureApmTestProjectExists(t)
	publishApmDependencyPackage(t, "test/project-flag-dep")

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

	// --dry-run passed directly (not via a "--" escape). Captures stdout to confirm apm's own
	// output actually acknowledges dry-run mode - proving the flag reached apm as a real,
	// recognized flag rather than being silently swallowed or misinterpreted - which
	// TestApmDryRunNoArtifacts (server-side non-upload only) doesn't check.
	output, err := captureStdout(t, func() error {
		return getApmCli().Exec("agent", "apm", "publish", "--package", "test/native-flags", "--registry", tests.AgentPackagesLocalRepo, "--dry-run")
	})
	require.NoError(t, err, "jf agent apm publish with --dry-run should succeed")
	assert.True(t, strings.Contains(strings.ToLower(output), "dry-run") || strings.Contains(strings.ToLower(output), "would publish"),
		"apm's own output should confirm dry-run mode was engaged, got: %s", output)

	// Verify no artifact was uploaded for dry-run
	searchSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/test/native-flags/*.zip").
		BuildSpec()
	artifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	assert.Empty(t, artifacts, "dry-run should not create artifacts in repository")
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

// TestApmIntegrationFullPipeline validates end-to-end install->publish->build-publish, checking
// real state after each step rather than only exit codes.
// TestApmInstallAndPublishWithBuildInfoComplete covers the same shape without a dependency; this
// is the one with both a real dependency AND a publish in a single pipeline.
func TestApmIntegrationFullPipeline(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/e2e-pipeline-dep")

	projectDir, err := os.MkdirTemp("", "apm-e2e-pipeline-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProjectWithDependency(t, projectDir, "test/e2e-pipeline-dep#1.0.0")

	buildName := "apm-e2e-pipeline"
	buildNumber := "300"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Step 1: Install (with build-info) - verify the dependency was captured LOCALLY (no server
	// round-trip yet). "jf rt bp" clears local partials for this exact build name/number after a
	// successful publish (Build.Clean() in build-info-go), so checking via the server here would
	// erase this step's dependency before Step 2's artifact could join it in one combined check.
	err = getApmCli().Exec("agent", "apm", "install", "--build-name", buildName, "--build-number", buildNumber)
	require.NoError(t, err, "Step 1: Install should succeed")
	localAfterInstall := readLocalApmPartialBuildInfo(t, buildName, buildNumber)
	require.NotEmpty(t, localAfterInstall.Modules, "locally-collected build info should have a module after install")
	assert.NotEmpty(t, localAfterInstall.Modules[0].Dependencies, "locally-collected build info should have the dependency after install")

	// Step 2: Publish (with build-info) - the dependency (Step 1) and the new artifact are both
	// still in local partials at this point (no bp call has run yet for this build/number), so
	// this first server round-trip sees both together.
	err = getApmCli().Exec("agent", "apm", "publish", "--package", "e2e/pipeline", "--registry", tests.AgentPackagesLocalRepo, "--build-name", buildName, "--build-number", buildNumber)
	require.NoError(t, err, "Step 2: Publish should succeed")
	validateBuildInfoHasBothArtifactsAndDependencies(t, buildName, buildNumber)

	// Step 3: Verify the published package actually landed in Artifactory. (Build info was
	// already published as part of Step 2's check above; a second "jf rt bp" here would just
	// republish an empty build, since Clean() already cleared local partials.)
	searchSpec := spec.NewBuilder().Pattern(tests.AgentPackagesLocalRepo + "/e2e/pipeline/*.zip").BuildSpec()
	artifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	require.NotEmpty(t, artifacts, "published package should be found in Artifactory")

	// Clean up
	_, _, _ = tests.DeleteFiles(searchSpec, serverDetails)
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

	publishApmDependencyPackage(t, "test/install-with-deps-bi")

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

// TestApmDependencyChecksumsInBuildInfo validates that dependency checksums (SHA1, MD5, SHA256)
// are all present and correct in build-info. This test covers the checksum resolution pipeline
// for dependencies (as opposed to TestApmChecksumsInBuildInfo which covers artifacts).
// The three-tier checksum resolution (cache → batched AQL → lockfile) must populate all
// three checksum types (SHA1, MD5, SHA256) for each dependency.
func TestApmDependencyChecksumsInBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	// Create and publish a test dependency package
	publishApmDependencyPackage(t, "test/checksum-validation-dep")

	// Create a project that installs that dependency
	projectDir := createProjectWithDependencies(t, "app-with-checksum-deps", []string{"test/checksum-validation-dep#1.0.0"})
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "501"
	err := runApmInstall(buildNumber)
	require.NoError(t, err, "install should succeed")

	// Get build info and validate dependency checksums
	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, buildResult.Modules, "build info should have a module")

	module := buildResult.Modules[0]
	require.NotEmpty(t, module.Dependencies, "build info should have dependencies")

	// Validate that EVERY dependency has all three checksum types (SHA1, MD5, SHA256)
	for _, dep := range module.Dependencies {
		assert.NotEmpty(t, dep.Sha256, "dependency %s should have SHA256 checksum", dep.Id)
		assert.NotEmpty(t, dep.Sha1, "dependency %s should have SHA1 checksum", dep.Id)
		assert.NotEmpty(t, dep.Md5, "dependency %s should have MD5 checksum", dep.Id)

		// Validate checksum format (hex string)
		assert.Regexp(t, "^[a-f0-9]{64}$", dep.Sha256, "dependency %s SHA256 should be valid hex", dep.Id)
		assert.Regexp(t, "^[a-f0-9]{40}$", dep.Sha1, "dependency %s SHA1 should be valid hex", dep.Id)
		assert.Regexp(t, "^[a-f0-9]{32}$", dep.Md5, "dependency %s MD5 should be valid hex", dep.Id)
	}

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

	publishApmDependencyPackage(t, "test/complete-app-dep")

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

// TestApmAuthWithoutEnvVarSucceeds validates the common, default case every other auth test in
// this file deliberately sets an env var to test around: install and publish must both succeed
// with NO APM_REGISTRY_* env var set at all, relying purely on jf's own automatic credential
// injection (BuildApmEnv/injectRegistryCredentialEnv in jfrog-cli-artifactory) from its
// configured server.
func TestApmAuthWithoutEnvVarSucceeds(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	// Ensure no leftover APM_REGISTRY_* env var from a prior test in this process interferes.
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "APM_REGISTRY_") {
			_ = os.Unsetenv(strings.SplitN(envVar, "=", 2)[0])
		}
	}

	publishApmDependencyPackage(t, "test/no-env-var-auth-dep")

	projectDir, err := os.MkdirTemp("", "apm-no-env-var-auth-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProjectWithDependency(t, projectDir, "test/no-env-var-auth-dep#1.0.0")

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	buildNumber := "112"
	require.NoError(t, getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumber),
		"install should succeed with no APM_REGISTRY_* env var set")
	require.NoError(t, getApmCli().Exec("agent", "apm", "publish", "--package", "test/no-env-var-auth-pkg", "--registry", tests.AgentPackagesLocalRepo),
		"publish should succeed with no APM_REGISTRY_* env var set")

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmCommandsFailWithoutJfServerConfig validates that removing jf's own server
// configuration entirely (not just APM_REGISTRY_* env vars or ~/.apm/config.json) causes
// install/publish to fail, since jf itself has nothing to build credentials from -
// confirming BuildApmEnv's credential injection genuinely depends on jf's own configured server,
// not some other fallback. Restores the "default" server config afterward unconditionally
// (regardless of how this test's own assertions turn out): every other test in this file, and
// the whole test binary, depends on it existing.
func TestApmCommandsFailWithoutJfServerConfig(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/no-server-config-dep")

	projectDir, err := os.MkdirTemp("", "apm-no-server-config-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProjectWithDependency(t, projectDir, "test/no-server-config-dep#1.0.0")

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Remove the "default" jf server config entirely. Restore it unconditionally afterward -
	// every other test in this file depends on it existing.
	configCli := coreTests.NewJfrogCli(execMain, "jfrog config", "")
	require.NoError(t, configCli.Exec("rm", "default", "--quiet"), "removing the default server config should succeed")
	defer createJfrogHomeConfig(t, true)

	assert.Error(t, getApmCli().Exec("agent", "apm", "install"), "install should fail without a configured jf server")
	assert.Error(t, getApmCli().Exec("agent", "apm", "publish", "--package", "test/no-server-config-pkg"), "publish should fail without a configured jf server")
}

// TestApmMixedRegistryDependenciesInOneInstall validates that a SINGLE install can resolve
// dependencies from two DIFFERENT registries at once - one dependency from apm-registry-1
// (non-default at install time), another from apm-registry-2 (the default, since
// "jf setup apm --repo X" makes the most-recently-configured repo the default and this
// test configures repos[1] last) - both declared in the same apm.yml via the object-form
// dependency's explicit "registry:" field, confirmed live (a local, parse-only apm install
// dry-run against unreachable ports) to be the real schema: "id: owner/name" +
// "registry: <name>" resolves against exactly the named registry. Covers both the
// non-default-registry and default-registry cases via its two dependencies, so a separate
// single-dependency "different registry" test would only be a strict subset of this one.
func TestApmMixedRegistryDependenciesInOneInstall(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

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

	setupCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	for _, repoName := range repos {
		err := setupCli.Exec("setup", "apm", "--repo", repoName)
		require.NoError(t, err, "setup should succeed for repo %s", repoName)
	}
	defer func() {
		_ = setupCli.Exec("setup", "apm", "--repo", tests.AgentPackagesLocalRepo)
	}()

	// Publish one dependency to each registry.
	owner := "test"
	pkgA, pkgB := "mixed-registry-dep-a", "mixed-registry-dep-b"
	publishApmDependencyPackageToRegistry(t, owner+"/"+pkgA, repos[0])
	publishApmDependencyPackageToRegistry(t, owner+"/"+pkgB, repos[1])

	// Consumer project depends on both, each via the object-form dependency's explicit registry:
	// field naming its own, different registry.
	projectDir, err := os.MkdirTemp("", "apm-mixed-registry-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".apm", "primitives"), dirPerms))
	apmYaml := fmt.Sprintf(`name: mixed-registry-consumer
version: 1.0.0
license: UNLICENSED
targets:
  - claude
dependencies:
  apm:
    - id: %s/%s
      version: "1.0.0"
      registry: %s
    - id: %s/%s
      version: "1.0.0"
      registry: %s
`, owner, pkgA, repos[0], owner, pkgB, repos[1])
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "apm.yml"), []byte(apmYaml), filePerms))

	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "405"
	err = runApmInstall(buildNumber)
	require.NoError(t, err, "install should resolve both dependencies, each from its own distinct registry")

	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, buildResult.Modules, "build info should have a module")
	module := buildResult.Modules[0]
	require.Len(t, module.Dependencies, 2, "both mixed-registry dependencies should be captured")

	var foundA, foundB bool
	for _, dep := range module.Dependencies {
		switch {
		case strings.Contains(dep.Id, pkgA):
			foundA = true
			assert.NotEmpty(t, dep.Sha256, "dependency from registry 1 should have a SHA256 checksum")
		case strings.Contains(dep.Id, pkgB):
			foundB = true
			assert.NotEmpty(t, dep.Sha256, "dependency from registry 2 should have a SHA256 checksum")
		}
	}
	assert.True(t, foundA, "dependency published to %s should be present in build info", repos[0])
	assert.True(t, foundB, "dependency published to %s should be present in build info", repos[1])

	deleteBuildInfo()
}

// getBasicApmYaml returns basic APM YAML
func getBasicApmYaml() string {
	return createApmYaml("test-app", "1.0.0", nil)
}

// createApmYaml creates customizable APM YAML with parameters. apmDeps are real APM dependency
// specs in "owner/name#version" shorthand (see publishApmDependencyPackage); an empty slice
// yields an empty "apm: []" dependency list. Registries are never declared here - they're
// configured globally via "jf setup apm", not per-project.
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

	return fmt.Sprintf(`name: %s
version: %s
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
dependencies:
%s`, name, version, depsSection)
}

// createProjectWithDependencies creates a project directory with specified dependencies
func createProjectWithDependencies(t *testing.T, name string, deps []string) string {
	apmYaml := createApmYaml(name, "1.0.0", deps)
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

// TestApmRegistryPrecedenceDefaultFallback validates that apm.yml's own "registries: default:
// <name>" sibling key (a real, distinct field from any per-registry "default" flag in
// ~/.apm/config.json - see manifest.go's ManifestRegistries.Default / its custom UnmarshalYAML)
// controls which registry a BARE, no-explicit-registry dependency ("owner/name#version")
// resolves against. Confirmed live with a local, parse/resolve-only apm install against two
// unreachable ports: the bare dependency routed to whichever port apm.yml's own default: key
// named, not the first-declared entry - so this publishes a real package to only the SECOND of
// two repos and asserts the bare-shorthand dependency still resolves successfully, which is only
// possible if apm.yml's default: is actually being honored.
func TestApmRegistryPrecedenceDefaultFallback(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

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

	setupCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	for _, repoName := range repos {
		err := setupCli.Exec("setup", "apm", "--repo", repoName)
		require.NoError(t, err, "setup should succeed for repo %s", repoName)
	}
	defer func() {
		_ = setupCli.Exec("setup", "apm", "--repo", tests.AgentPackagesLocalRepo)
	}()

	// Publish a real package to the SECOND repo only.
	owner, pkgName := "test", "default-fallback-dep"
	publishApmDependencyPackageToRegistry(t, owner+"/"+pkgName, repos[1])

	// apm.yml declares both repos as named registries, with its own default: pointing at the
	// second one. The dependency below uses the bare shorthand (no explicit registry: field), so
	// it can only resolve correctly if apm.yml's own default: is actually being honored.
	projectDir, err := os.MkdirTemp("", "apm-registry-precedence-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".apm", "primitives"), dirPerms))
	apmYaml := fmt.Sprintf(`name: registry-precedence-consumer
version: 1.0.0
license: UNLICENSED
targets:
  - claude
registries:
  %s:
    url: "%s"
  %s:
    url: "%s"
  default: %s
dependencies:
  apm:
    - %s/%s#1.0.0
`, repos[0], apmRegistryURL(repos[0]), repos[1], apmRegistryURL(repos[1]), repos[1], owner, pkgName)
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "apm.yml"), []byte(apmYaml), filePerms))

	defer setupTestWorkingDirectory(t, projectDir)()

	buildNumber := "201"
	err = getApmCli().Exec("agent", "apm", "install", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "install should resolve the bare-shorthand dependency via apm.yml's own registries.default: precedence")

	validateBuildInfoDependencies(t, apmBuildName, buildNumber)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmPublishWithDependencyMetadata validates publish captures dependency metadata (P0: Scenario #7).
func TestApmPublishWithDependencyMetadata(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	apmYaml := `name: app-with-deps
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

// TestApmFrozenModeWithDependencies validates frozen mode works with dependencies (P1: Scenario #14).
func TestApmFrozenModeWithDependencies(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/frozen-mode-dep")

	projectDir, err := os.MkdirTemp("", "apm-frozen-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProjectWithDependency(t, projectDir, "test/frozen-mode-dep#1.0.0")

	defer setupTestWorkingDirectory(t, projectDir)()

	// First install to create lockfile
	err = getApmCli().Exec("agent", "apm", "install")
	require.NoError(t, err)

	// Frozen install should succeed (lockfile exists and is up-to-date). --frozen must be
	// passed directly, not after a "--" escape: apm parses anything after "--" as a
	// positional package argument, not a flag (see TestApmNativeFlags for the same bug).
	err = getApmCli().Exec("agent", "apm", "install", "--frozen")
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

// TestApmNativeCliWorksWithJfSetupCredentials validates that once "jf setup apm" has run,
// the native apm binary can be invoked directly - bypassing "jf agent apm ..." entirely, with no
// build-name/build-number, no build-info collection at all - and still authenticate
// successfully. jf setup apm persists credentials into ~/.apm/config.json; that's a
// different mechanism from BuildApmEnv's APM_REGISTRY_TOKEN_<NAME> env-var injection, which only
// happens when jf itself invokes apm as a subprocess. A user running the plain "apm" command in
// their own shell gets none of that env-var wiring, so this test strips any leftover
// APM_REGISTRY_* env vars first, to prove config.json alone is sufficient.
//
// Exit code from apm is not enough evidence either way, so both halves check real server state:
// publish is verified by searching Artifactory for the uploaded artifact (not just that apm
// returned 0), and install is verified by asserting the resulting apm.lock.yaml actually
// references the package and version that was just published (not just that a lockfile exists).
func TestApmNativeCliWorksWithJfSetupCredentials(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	// initApmTest already ran "jf setup apm --repo <repo>" (via initApmConfig), writing
	// credentials into ~/.apm/config.json. Strip any APM_REGISTRY_* env vars a prior test in
	// this process may have left behind, so a passing result here can only be explained by that
	// config file.
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "APM_REGISTRY_") {
			_ = os.Unsetenv(strings.SplitN(envVar, "=", 2)[0])
		}
	}

	owner, pkgName := "test", "native-cli-pkg"

	// Step 1: publish using the native apm binary directly (no "jf agent apm publish", no
	// --build-name/--build-number).
	publishDir, err := os.MkdirTemp("", "apm-native-publish-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(publishDir)
	}()
	require.NoError(t, os.MkdirAll(filepath.Join(publishDir, ".apm", "primitives"), dirPerms))
	apmYaml := fmt.Sprintf(`name: %s
version: 1.0.0
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
`, pkgName)
	require.NoError(t, os.WriteFile(filepath.Join(publishDir, "apm.yml"), []byte(apmYaml), filePerms))
	require.NoError(t, os.WriteFile(filepath.Join(publishDir, ".apm", "primitives", "placeholder.txt"), []byte("placeholder content"), filePerms))

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, publishDir)

	nativePublish := exec.Command("apm", "publish", "--package", owner+"/"+pkgName, "--registry", tests.AgentPackagesLocalRepo) // #nosec G204 -- fixed argv, no shell, no user input
	nativePublish.Stdout = os.Stdout
	nativePublish.Stderr = os.Stderr
	require.NoError(t, nativePublish.Run(), "native apm publish (no jf wrapper) should succeed using jf setup apm's persisted credentials")

	// Verify the package was genuinely uploaded to Artifactory - not just that apm exited 0.
	searchSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/" + owner + "/" + pkgName + "/*.zip").
		BuildSpec()
	publishedArtifacts, _, err := tests.SearchFiles(searchSpec, serverDetails)
	require.NoError(t, err)
	require.NotEmpty(t, publishedArtifacts, "native apm publish should have uploaded the package to Artifactory")
	assert.True(t,
		strings.HasPrefix(publishedArtifacts[0].Name, pkgName+"-") && strings.HasSuffix(publishedArtifacts[0].Name, ".zip"),
		"published artifact name should follow <name>-<version>.zip, got %q", publishedArtifacts[0].Name)

	// Step 2: install that same package using the native apm binary, from a separate consumer
	// project (no "jf agent apm install").
	installDir, err := os.MkdirTemp("", "apm-native-install-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(installDir)
	}()
	consumerYaml := fmt.Sprintf(`version: "1.0.0"
name: native-cli-consumer
license: UNLICENSED
targets:
  - claude
dependencies:
  apm:
    - %s/%s#1.0.0
`, owner, pkgName)
	require.NoError(t, os.MkdirAll(filepath.Join(installDir, ".apm"), dirPerms))
	require.NoError(t, os.WriteFile(filepath.Join(installDir, "apm.yml"), []byte(consumerYaml), filePerms))

	clientTestUtils.ChangeDirAndAssert(t, installDir)

	nativeInstall := exec.Command("apm", "install") // #nosec G204 -- fixed argv, no shell, no user input
	nativeInstall.Stdout = os.Stdout
	nativeInstall.Stderr = os.Stderr
	require.NoError(t, nativeInstall.Run(), "native apm install (no jf wrapper) should succeed using jf setup apm's persisted credentials")

	// Verify the package was genuinely resolved from Artifactory - not just that apm exited 0.
	lockfilePath := filepath.Join(installDir, "apm.lock.yaml")
	require.FileExists(t, lockfilePath, "apm.lock.yaml should exist after native apm install")
	lockfileContent, err := os.ReadFile(lockfilePath)
	require.NoError(t, err)
	assert.Contains(t, string(lockfileContent), pkgName, "lockfile should reference the installed package")
	assert.Contains(t, string(lockfileContent), "1.0.0", "lockfile should record the resolved version")

	// Clean up the published artifact from Artifactory.
	_, _, _ = tests.DeleteFiles(searchSpec, serverDetails)
}

// TestApmInstallPositionalPackageWithBuildInfo validates
// "jf agent apm install <owner>/<name>#<version>" - naming the dependency directly on the
// command line, which both adds it to apm.yml and installs it in one step. Every other
// install-with-dependency test in this file pre-declares the dependency in apm.yml's
// dependencies: block first and calls plain "install"; this is the one CLI-driven path.
func TestApmInstallPositionalPackageWithBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/positional-install-dep")

	projectDir, err := os.MkdirTemp("", "apm-positional-install-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	createApmTestProject(t, projectDir)

	buildNumber := "111"
	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// Install by naming the package directly on the command line, not by pre-declaring it in
	// apm.yml first.
	err = getApmCli().Exec("agent", "apm", "install", "test/positional-install-dep#1.0.0", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install <owner>/<name>#<version> should succeed")

	// apm.yml should have been updated with the new dependency as a side effect.
	apmYamlContent, err := os.ReadFile(filepath.Join(projectDir, "apm.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(apmYamlContent), "test/positional-install-dep", "apm.yml should be updated with the positionally-installed dependency")

	// Verify the dependency is captured in build info, with a real checksum - not just that the
	// install command itself succeeded.
	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, buildResult.Modules, "build info should have a module")
	module := buildResult.Modules[0]
	require.NotEmpty(t, module.Dependencies, "build info should have a dependency")

	var found bool
	for _, dep := range module.Dependencies {
		if strings.Contains(dep.Id, "positional-install-dep") {
			found = true
			assert.NotEmpty(t, dep.Sha256, "positionally-installed dependency should have a SHA256 checksum")
		}
	}
	assert.True(t, found, "build info dependency list should include the positionally-installed package")

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmPassthroughServerIdFlag validates that "jf agent apm <native-subcommand>" (any
// subcommand other than install/publish) honors --server-id for server selection, and
// never lets that flag leak through as a raw, unrecognized argument to the native apm binary
// (which has no --server-id option of its own).
//
// Regression test for a passthrough bug where RunApmPassthroughDefault resolved auth from the
// default configured server unconditionally, ignoring --server-id entirely, while the raw
// "--server-id <value>" tokens still rode along in the forwarded args and broke apm itself
// with "Error: No such option: --server-id". Fixed by extracting --server-id via
// coreutils.ExtractServerIdFromCommand before resolving the server or building the native
// command, mirroring jf nix's own passthrough dispatcher.
func TestApmPassthroughServerIdFlag(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/passthrough-server-id-dep")

	projectDir, err := os.MkdirTemp("", "apm-passthrough-server-id-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	createApmTestProjectWithDependency(t, projectDir, "test/passthrough-server-id-dep#1.0.0")

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	// "outdated" (like most native apm subcommands) needs an existing apm.lock.yaml to have
	// anything to check - install first so a real lockfile exists before exercising passthrough.
	require.NoError(t, getApmCli().Exec("agent", "apm", "install"), "setup install for the passthrough test should succeed")

	// Explicit --server-id naming the real, configured server ("default", set up by
	// createJfrogHomeConfig) should succeed exactly like omitting it entirely.
	err = getApmCli().Exec("agent", "apm", "outdated", "--server-id", "default")
	require.NoError(t, err, "jf agent apm outdated --server-id <valid server> should succeed, not leak the flag to apm")

	// A bad --server-id must fail at jf's OWN server resolution, with jf's own "does not
	// exist" error - proving --server-id was actually extracted and consumed by jf, not
	// silently ignored and then forwarded to apm. Before the fix, this case did not surface
	// "does not exist" at all: --server-id's value was never inspected for server selection
	// (silently falling back to the default server), so the only failure was apm's own
	// generic "unrecognized option" exit further downstream.
	err = getApmCli().Exec("agent", "apm", "outdated", "--server-id", "definitely-not-a-configured-server")
	require.Error(t, err, "jf agent apm outdated --server-id <bad server> should fail")
	assert.Contains(t, err.Error(), "does not exist",
		"error should come from jf's own server resolution (proving --server-id was consumed), not from a bare exec failure after the flag leaked through to apm")
}

// NOTE: a real e2e test for "install's default module ID falls back to the build name when
// apm.yml has no version" was attempted here and removed - the native apm binary itself hard-
// requires the version field ("Missing required field 'version' in apm.yml") and refuses to run
// at all without it, so that scenario is unreachable through a genuine "jf agent apm install"
// call; apm never lets jf's own build-info code see an incomplete manifest in practice. The
// behavior is already covered at the unit level by TestDerivedModuleID in
// jfrog-cli-artifactory/agent/apm/common/build_info_test.go, which exercises derivedModuleID
// directly without going through apm's own stricter validation.

// TestApmInstallDevDependencyScope validates that a dependency installed with --dev is recorded
// with scope "dev" in build info, matching npm's own dev-dependency scope convention (see
// finalScope in agent/apm/common/dependency_resolver.go).
func TestApmInstallDevDependencyScope(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/dev-scope-dep")

	projectDir, err := os.MkdirTemp("", "apm-dev-scope-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	createApmTestProject(t, projectDir)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	buildNumber := "114"
	err = getApmCli().Exec("agent", "apm", "install", "--dev", "test/dev-scope-dep#1.0.0", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install --dev <pkg> should succeed")

	// apm.lock.yaml should record is_dev: true for this dependency - the flag
	// dependency_resolver.go's finalScope trusts to classify the scope.
	lockfileContent, err := os.ReadFile(filepath.Join(projectDir, "apm.lock.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(lockfileContent), "is_dev: true",
		"apm.lock.yaml should mark the --dev-installed dependency as a dev dependency")

	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, buildResult.Modules, "build info should have a module")
	module := buildResult.Modules[0]

	var found bool
	for _, dep := range module.Dependencies {
		if strings.Contains(dep.Id, "dev-scope-dep") {
			found = true
			assert.Contains(t, dep.Scopes, "dev",
				"a --dev-installed dependency should be scoped 'dev' in build info, matching npm's own dev-dependency scope convention")
		}
	}
	assert.True(t, found, "build info dependency list should include the --dev-installed package")

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmInstallRootFlagBuildInfo validates that --root redirects apm_modules/ and
// apm.lock.yaml under the given directory, and - the actual jf-owned regression risk - that
// build-info collection reads the lockfile from that redirected location rather than silently
// looking in the default project root and finding nothing.
func TestApmInstallRootFlagBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	publishApmDependencyPackage(t, "test/root-flag-dep")

	projectDir, err := os.MkdirTemp("", "apm-root-flag-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	createApmTestProjectWithDependency(t, projectDir, "test/root-flag-dep#1.0.0")
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "out"), dirPerms))

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	buildNumber := "115"
	err = getApmCli().Exec("agent", "apm", "install", "--root", "out", "--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm install --root out should succeed")

	require.FileExists(t, filepath.Join(projectDir, "out", "apm.lock.yaml"),
		"apm.lock.yaml should be written under --root's target directory")
	_, statErr := os.Stat(filepath.Join(projectDir, "apm.lock.yaml"))
	assert.True(t, os.IsNotExist(statErr),
		"apm.lock.yaml should NOT be written at the project root when --root redirects it")

	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, buildResult.Modules, "build info should have a module")
	module := buildResult.Modules[0]
	require.NotEmpty(t, module.Dependencies,
		"build info should have picked up dependencies from the --root-redirected lockfile, not found nothing")

	var found bool
	for _, dep := range module.Dependencies {
		if strings.Contains(dep.Id, "root-flag-dep") {
			found = true
		}
	}
	assert.True(t, found,
		"build info dependency list should include the dependency, proving build-info collection read the --root-redirected apm.lock.yaml")

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmPublishZipFlagBuildInfo validates that --zip's explicit archive path is what
// build-info's checksum is actually derived from, not a default-named file. The zip here is
// deliberately named differently from apm's own auto-pack convention (<name>-<version>.zip),
// so build-info can only get the right checksum by genuinely reading --zip's value - a broken
// extraction would fall back to a default-named file that doesn't exist, leaving no checksum at
// all rather than merely a wrong one.
func TestApmPublishZipFlagBuildInfo(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	owner, pkgName := "test", "zip-flag-pkg"
	projectDir, err := os.MkdirTemp("", "apm-zip-flag-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()

	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, ".apm", "primitives"), dirPerms))
	apmYaml := fmt.Sprintf(`name: %s
version: 1.0.0
license: UNLICENSED
targets:
  - claude
primitives:
  agents: []
`, pkgName)
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "apm.yml"), []byte(apmYaml), filePerms))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".apm", "primitives", "placeholder.txt"), []byte("placeholder content"), filePerms))

	customZipPath := filepath.Join(projectDir, "custom-prebuilt.zip")
	zipFile, err := os.Create(customZipPath) // #nosec G304 -- test-controlled temp path
	require.NoError(t, err)
	zipWriter := zip.NewWriter(zipFile)
	fileWriter, err := zipWriter.Create("apm.yml")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte(apmYaml))
	require.NoError(t, err)
	require.NoError(t, zipWriter.Close())
	require.NoError(t, zipFile.Close())
	expectedChecksum := computeFileSHA256(t, customZipPath)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	buildNumber := "116"
	err = getApmCli().Exec("agent", "apm", "publish", "--package", owner+"/"+pkgName, "--registry", tests.AgentPackagesLocalRepo, "--zip", customZipPath,
		"--build-name", apmBuildName, "--build-number", buildNumber)
	require.NoError(t, err, "jf agent apm publish --zip <custom path> should succeed")

	buildResult := fetchPublishedApmBuildInfo(t, apmBuildName, buildNumber)
	require.NotEmpty(t, buildResult.Modules, "build info should have a module")
	module := buildResult.Modules[0]
	require.NotEmpty(t, module.Artifacts, "build info should have an artifact")

	assert.Equal(t, expectedChecksum, module.Artifacts[0].Sha256,
		"published artifact's build-info checksum should match the explicit --zip file's own hash, proving --zip's value was actually used rather than a default-named file")

	searchSpec := spec.NewBuilder().
		Pattern(tests.AgentPackagesLocalRepo + "/" + owner + "/" + pkgName + "/*.zip").
		BuildSpec()
	_, _, _ = tests.DeleteFiles(searchSpec, serverDetails)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, apmBuildName, artHttpDetails)
}

// TestApmPublishPackageFlagNotMisPromotedFromOtherFlagValues is a regression test for the exact
// scenario reviewers flagged on PR #518: a value-taking apm-native flag's own value ("foo.zip")
// must never be mistaken for --package, even when immediately followed by something that looks
// like a bare positional package spec ("acme/pkg").
func TestApmPublishPackageFlagNotMisPromotedFromOtherFlagValues(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	projectDir, err := os.MkdirTemp("", "apm-mispromote-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	createApmTestProject(t, projectDir)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "publish", "--zip", "foo.zip", "acme/pkg")
	require.Error(t, err, "publish without an explicit --package should fail, even with --zip foo.zip acme/pkg present")
	assert.Contains(t, err.Error(), "requires --package",
		"error should be the explicit --package requirement, not a downstream failure from silently promoting acme/pkg into --package")
}

// writeFakeApmVersionScript writes a fake "apm" executable to a fresh directory that only
// understands "--version", printing versionOutput and exiting 0. Returns the directory to
// prepend to PATH.
func writeFakeApmVersionScript(t *testing.T, versionOutput string) string {
	t.Helper()
	binDir := t.TempDir()
	apmPath := filepath.Join(binDir, "apm")
	script := "#!/bin/sh\necho \"" + versionOutput + "\"\nexit 0\n"
	if runtime.GOOS == "windows" {
		apmPath += ".bat"
		script = "@echo " + versionOutput + "\r\n@exit /b 0\r\n"
	}
	// 0755 (not 0644) is required here: this file is exec'd directly as a command, and the
	// executable bit is what makes that possible - a stricter mode would make PATH resolution
	// find it but fail to run it. Still safe: the stub lives under t.TempDir(), never a
	// shared or persistent location.
	require.NoError(t, os.WriteFile(apmPath, []byte(script), 0755)) // #nosec G306 -- executable stub under t.TempDir, not a shared path
	return binDir
}

// TestApmMinVersionGate validates that ValidateApmPrerequisites rejects an installed apm below
// minSupportedApmVersion (agent/apm/common/utils.go) before ever touching Artifactory or running
// the real install/publish flow, naming both the required and the actual version in the
// error.
func TestApmMinVersionGate(t *testing.T) {
	initApmTest(t)
	defer cleanApmTest(t)

	binDir := writeFakeApmVersionScript(t, "Agent Package Manager (APM) CLI version 0.10.0 (fake)")
	prevPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+prevPath)

	projectDir, err := os.MkdirTemp("", "apm-min-version-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(projectDir)
	}()
	createApmTestProject(t, projectDir)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	clientTestUtils.ChangeDirAndAssert(t, projectDir)

	err = getApmCli().Exec("agent", "apm", "install")
	require.Error(t, err, "install should be rejected when the installed apm version is below the minimum supported version")
	assert.Contains(t, err.Error(), "0.23.0", "error should name the minimum supported version")
	assert.Contains(t, err.Error(), "0.10.0", "error should name the actual (too-old) installed version")
}

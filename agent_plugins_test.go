package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	biutils "github.com/jfrog/build-info-go/utils"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	coreBuild "github.com/jfrog/jfrog-cli-core/v2/common/build"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

// ---------------------------------------------------------------------------
// Init / cleanup
// ---------------------------------------------------------------------------

func InitAgentPluginsTests() {
	initArtifactoryCli()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

func CleanAgentPluginsTests() {
	deleteCreatedRepos()
}

func initAgentPluginsTest(t *testing.T) {
	if !*tests.TestAgentPlugins {
		t.Skip("Skipping agent plugins tests. To run add '--test.agentPlugins'.")
	}
	createJfrogHomeConfig(t, false)
	require.True(t, isRepoExist(tests.AgentPluginsLocalRepo), "agent plugins local repo does not exist: "+tests.AgentPluginsLocalRepo)
}

func cleanAgentPluginsTest() {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.AgentPluginsBuildName, artHttpDetails)
	_ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, "1", "")
	tests.CleanFileSystem()
}

// runAgentPluginsCmd executes `jf agent plugins <args...>`.
func runAgentPluginsCmd(t *testing.T, args ...string) error {
	t.Helper()
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(append([]string{"agent", "plugins"}, args...)...)
}

// createTestPlugin copies the test-plugin fixture to a fresh temp dir and patches
// plugin.json with the given slug and version so tests don't conflict.
func createTestPlugin(t *testing.T, slug, version string) string {
	t.Helper()
	pluginSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "agent_plugins", "test-plugin")
	pluginPath, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	assert.NoError(t, biutils.CopyDir(pluginSrc, pluginPath, true, nil))

	manifest := map[string]string{
		"name":        slug,
		"version":     version,
		"description": "Integration test plugin",
	}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(pluginPath, "plugin.json"), data, 0644)) // #nosec G306 -- test fixture
	return pluginPath
}

// assertPluginExists verifies the zip for slug/version is present in the local repo.
func assertPluginExists(t *testing.T, slug, version string) {
	t.Helper()
	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)
	_, err = sm.GetItemProps(pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version))
	require.NoError(t, err, "artifact should exist: %s v%s", slug, version)
}

// assertPluginAbsent verifies the zip for slug/version is gone from the local repo.
func assertPluginAbsent(t *testing.T, slug, version string) {
	t.Helper()
	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)
	_, err = sm.GetItemProps(pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version))
	assert.Error(t, err, "artifact should not exist: %s v%s", slug, version)
}

// pluginArtifactPath returns the Artifactory path for a published plugin zip:
// <repo>/<slug>/<version>/<slug>-<version>.zip
func pluginArtifactPath(repo, slug, version string) string {
	return repo + "/" + slug + "/" + version + "/" + slug + "-" + version + ".zip"
}

// ---------------------------------------------------------------------------
// P0 — Publish
// ---------------------------------------------------------------------------

// TestAgentPluginsPublish verifies that publishing a plugin directory uploads
// the zip to the correct path in the agentplugins local repository.
// Covers scenarios #9 and #74.
func TestAgentPluginsPublish(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "publish-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	assertPluginExists(t, slug, version)
}

// TestAgentPluginsVersionCollisionCI verifies that publishing the same version
// twice in CI/non-interactive mode fails with a clear "already exists" error.
// Covers scenario #19.
func TestAgentPluginsVersionCollisionCI(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "collision-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// Force non-interactive mode so the collision check fails immediately.
	t.Setenv("CI", "true")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	require.Error(t, err, "second publish of the same version in CI mode should fail")
	assert.Contains(t, strings.ToLower(err.Error()), "already exists",
		"error should mention 'already exists'")
}

// TestAgentPluginsPublishWithVersion verifies that --version overrides the
// manifest version on publish.
// Covers scenario #11.
func TestAgentPluginsPublishWithVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "version-override-plugin"
	overrideVersion := "2.0.0"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+overrideVersion,
	))

	assertPluginExists(t, slug, overrideVersion)
}

// TestAgentPluginsPublishMissingPluginJson verifies that publishing a directory
// without plugin.json returns a clear error.
// Covers scenario #14.
func TestAgentPluginsPublishMissingPluginJson(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	emptyDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"publish", emptyDir,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	assert.Error(t, err, "publish of directory without plugin.json should fail")
}

// TestAgentPluginsPublishToNonExistentRepo verifies that publishing to a
// nonexistent repository returns a clear error.
// Covers scenario #23.
func TestAgentPluginsPublishToNonExistentRepo(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	pluginPath := createTestPlugin(t, "invalid-repo-plugin", "1.0.0")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo=nonexistent-agent-plugins-repo-xyz",
	)
	assert.Error(t, err, "publish to nonexistent repo should fail")
}

// ---------------------------------------------------------------------------
// P0 — Install
// ---------------------------------------------------------------------------

// TestAgentPluginsInstallLatest verifies that installing a plugin without
// --version picks up the latest published version and places files at
// <installPath>/<slug>/. Uses --path to bypass harness resolution.
// Covers scenarios #25 and #33.
func TestAgentPluginsInstallLatest(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "install-latest-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	// Plugin files should be at <installDir>/<slug>/
	assert.FileExists(t, filepath.Join(installDir, slug, "plugin.json"),
		"plugin.json should exist after install")
}

// TestAgentPluginsInstallSpecificVersion verifies that --version installs the
// requested version rather than latest.
// Covers scenario #32.
func TestAgentPluginsInstallSpecificVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "install-version-plugin"

	// Publish two versions.
	v1Path := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, "2.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	installDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version=1.0.0",
	))

	installedManifest := filepath.Join(installDir, slug, "plugin.json")
	require.FileExists(t, installedManifest)
	data, err := os.ReadFile(installedManifest) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]string
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, "1.0.0", manifest["version"], "installed version should be 1.0.0, not latest 2.0.0")
}

// TestAgentPluginsInstallNotFound verifies that installing an unknown slug
// returns a clear not-found error.
// Covers scenario #34.
func TestAgentPluginsInstallNotFound(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	installDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", "nonexistent-slug-xyzzy",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	)
	assert.Error(t, err, "installing an unknown slug should fail with a not-found error")
}

// ---------------------------------------------------------------------------
// P0 — Delete
// ---------------------------------------------------------------------------

// TestAgentPluginsDelete verifies that deleting a specific version removes
// the version folder from Artifactory. --version is always required by the command.
// Covers scenarios #48 and #49.
func TestAgentPluginsDelete(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))
	assertPluginExists(t, slug, version)

	assert.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+version,
	))
	assertPluginAbsent(t, slug, version)
}

// TestAgentPluginsDeleteDryRun verifies that --dry-run does not remove the
// artifact from Artifactory.
// Covers scenario #50.
func TestAgentPluginsDeleteDryRun(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-dryrun-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	assert.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+version,
		"--dry-run",
	))

	assertPluginExists(t, slug, version)
}

// ---------------------------------------------------------------------------
// P0 — Search
// ---------------------------------------------------------------------------

// TestAgentPluginsSearch verifies that `jf agent plugins search <query>`
// returns matches by the agentplugins.name property without error.
// Covers scenario #56.
func TestAgentPluginsSearch(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "search-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	assert.NoError(t, runAgentPluginsCmd(t,
		"search", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "search should succeed after publish")
}

// TestAgentPluginsSearchNoMatches verifies that searching with a query that
// matches nothing returns an empty result — not an error.
// Covers scenario #59.
func TestAgentPluginsSearchNoMatches(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	assert.NoError(t, runAgentPluginsCmd(t,
		"search", "nonexistent-plugin-xyzzy-abc123",
		"--repo="+tests.AgentPluginsLocalRepo,
	), "search with no matches should return empty result, not an error")
}

// ---------------------------------------------------------------------------
// P0 — Checksum integrity
// ---------------------------------------------------------------------------

// TestAgentPluginsChecksumIntegrity verifies that after publish the artifact
// in build info has a non-empty, non-"untrusted" SHA256 checksum, confirming
// Artifactory computed the checksum correctly on upload.
// Covers scenario #69.
func TestAgentPluginsChecksumIntegrity(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "checksum-plugin"
	version := "1.0.0"
	buildNumber := "1"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	require.True(t, found, "build info not found — was jf rt bp successful?")
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts,
		"expected at least one artifact in build info")

	for _, a := range publishedBuildInfo.BuildInfo.Modules[0].Artifacts {
		assert.NotEmpty(t, a.Sha256, "artifact %s: sha256 must not be empty", a.Name)
		assert.NotEqual(t, "untrusted", strings.ToLower(a.Sha256),
			"artifact %s: sha256 must not be 'untrusted'", a.Name)
	}
}

// ---------------------------------------------------------------------------
// P0 — Round-trip (publish → install)
// ---------------------------------------------------------------------------

// TestAgentPluginsRoundTrip publishes a plugin then installs it and verifies
// the installed manifest matches slug and version.
// Covers scenario #75.
func TestAgentPluginsRoundTrip(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "roundtrip-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	installedManifest := filepath.Join(installDir, slug, "plugin.json")
	require.FileExists(t, installedManifest, "plugin.json should exist after install")
	data, err := os.ReadFile(installedManifest) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]string
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, slug, manifest["name"], "installed plugin name should match published slug")
	assert.Equal(t, version, manifest["version"], "installed plugin version should match published version")
}

// ---------------------------------------------------------------------------
// P1 — List
// ---------------------------------------------------------------------------

// TestAgentPluginsListRemote verifies that `jf agent plugins list` returns
// without error after a publish.
// Covers scenario #52.
func TestAgentPluginsListRemote(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "list-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	assert.NoError(t, runAgentPluginsCmd(t,
		"list",
		"--repo="+tests.AgentPluginsLocalRepo,
	), "list should succeed after publish")
}

// ---------------------------------------------------------------------------
// P1 — Build info
// ---------------------------------------------------------------------------

// TestAgentPluginsPublishWithBuildInfo verifies that --build-name and
// --build-number cause the published zip to appear as an artifact in build info
// with a valid SHA256 checksum.
// Covers scenario #60.
func TestAgentPluginsPublishWithBuildInfo(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "buildinfo-plugin"
	version := "1.0.0"
	buildNumber := "1"
	pluginPath := createTestPlugin(t, slug, version)

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	require.True(t, found, "build info not found — was 'jf rt bp' successful?")
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1, "expected 1 build info module")

	module := publishedBuildInfo.BuildInfo.Modules[0]
	require.NotEmpty(t, module.Artifacts, "published zip should appear as an artifact in build info")
	assert.NotEmpty(t, module.Artifacts[0].Sha256, "artifact sha256 should be non-empty in build info")
}

// TestAgentPluginsNoBuildInfoWithoutFlags verifies that publishing without
// --build-name and --build-number does not create a build info entry.
// Covers scenarios #61–63.
func TestAgentPluginsNoBuildInfoWithoutFlags(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "no-buildinfo-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	localBuilds, err := coreBuild.GetGeneratedBuildsInfo(tests.AgentPluginsBuildName, "1", "")
	assert.NoError(t, err)
	assert.Empty(t, localBuilds, "no local build info should be stored when --build-name/--build-number are absent")
}

// TestAgentPluginsModuleOverride verifies that --module overrides the default
// module ID (slug) in build info.
// Covers scenario #64.
func TestAgentPluginsModuleOverride(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "module-override-plugin"
	buildNumber := "1"
	customModule := "my-custom-agent-module"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
		"--module="+customModule,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	require.True(t, found, "build info not found")
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)
	assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id,
		"--module flag should override the default module ID in build info")
}

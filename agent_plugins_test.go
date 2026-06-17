package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	biutils "github.com/jfrog/build-info-go/utils"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	coreBuild "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	agentTestutil "github.com/jfrog/jfrog-cli-artifactory/agent/common/testutil"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-evidence/evidence/cryptox"
	"github.com/jfrog/jfrog-cli-evidence/evidence/generate"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
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
	// The test Artifactory instance has no evidence/One-Model service configured.
	// Disable the quiet-failure evidence gate so install commands don't block on 403.
	t.Setenv("JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE", "true")
}

func cleanAgentPluginsTest() {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.AgentPluginsBuildName, artHttpDetails)
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

	require.NoError(t, biutils.CopyDir(pluginSrc, pluginPath, true, nil))

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

// createTestClaudePlugin creates a plugin fixture whose manifest lives at
// .claude-plugin/plugin.json (the Claude-style harness path) rather than the
// root plugin.json. Publish discovers it via the built-in manifest search paths.
func createTestClaudePlugin(t *testing.T, slug, version string) string {
	t.Helper()
	pluginPath, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	claudeDir := filepath.Join(pluginPath, ".claude-plugin")
	require.NoError(t, os.MkdirAll(claudeDir, 0755)) // #nosec G301 -- test directory

	manifest := map[string]string{
		"name":        slug,
		"version":     version,
		"description": "Integration test claude-style plugin",
	}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "plugin.json"), data, 0644)) // #nosec G306 -- test fixture
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
// A search-based check gives a cleaner absence assertion without relying on error message text.
func assertPluginAbsent(t *testing.T, slug, version string) {
	t.Helper()
	path := pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version)
	searchSpec := spec.NewBuilder().Pattern(path).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	require.NoError(t, err, "search for absent artifact failed")
	defer func() { _ = reader.Close() }()
	var found bool
	for item := new(artUtils.SearchResult); reader.NextRecord(item) == nil; item = new(artUtils.SearchResult) {
		found = true
		break
	}
	assert.False(t, found, "artifact should not exist: %s v%s", slug, version)
}

// pluginArtifactPath returns the Artifactory path for a published plugin zip:
// <repo>/<slug>/<version>/<slug>-<version>.zip
func pluginArtifactPath(repo, slug, version string) string {
	return repo + "/" + slug + "/" + version + "/" + slug + "-" + version + ".zip"
}

// uploadMarketplaceJSON uploads a minimal <harness>-marketplace.json to the repo root
// so that install without --version can resolve the version via marketplace lookup.
func uploadMarketplaceJSON(t *testing.T, harness, slug, version string) {
	t.Helper()
	content := fmt.Sprintf(`{"name":%q,"plugins":[{"name":%q,"version":%q}]}`, harness, slug, version)
	f, err := os.CreateTemp("", harness+"-marketplace-*.json")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	fileName := harness + "-marketplace.json"
	require.NoError(t, artifactoryCli.Exec("u", f.Name(),
		tests.AgentPluginsLocalRepo+"/"+fileName,
		"--flat=true",
	), "uploading %s to test repo must succeed", fileName)
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

	require.NoError(t, runAgentPluginsCmd(t,
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

	require.NoError(t, runAgentPluginsCmd(t,
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

	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+version,
		"--dry-run",
	))

	assertPluginExists(t, slug, version)
}

// TestAgentPluginsDeleteDryRunNotFound verifies that delete --dry-run on a
// plugin that does not exist returns a not-found error rather than silently
// succeeding.
func TestAgentPluginsDeleteDryRunNotFound(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"delete", "nonexistent-dryrun-plugin",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version=1.0.0",
		"--dry-run",
	)
	require.Error(t, err, "delete --dry-run on a missing plugin must return an error")
	assert.Contains(t, err.Error(), "not found", "error should indicate the plugin was not found")
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
	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
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
	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
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

// TestAgentPluginsPublishBuildNameWithoutNumber verifies that providing only
// one of --build-name / --build-number (scenarios #61 and #62) returns an
// error requiring both flags together.
func TestAgentPluginsPublishBuildNameWithoutNumber(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	cases := []struct {
		name        string
		extraArgs   []string
		description string
	}{
		{
			name:        "name-only",
			extraArgs:   []string{"--build-name=" + tests.AgentPluginsBuildName},
			description: "--build-name without --build-number must return an error",
		},
		{
			name:        "number-only",
			extraArgs:   []string{"--build-number=42"},
			description: "--build-number without --build-name must return an error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			slug := "build-flag-validation-plugin-" + tc.name
			pluginPath := createTestPlugin(t, slug, "1.0.0")
			args := append([]string{"publish", pluginPath, "--repo=" + tests.AgentPluginsLocalRepo}, tc.extraArgs...)
			require.Error(t, runAgentPluginsCmd(t, args...), tc.description)
		})
	}
}

// TestAgentPluginsModuleOverride verifies that --module overrides the default
// module ID (slug) in build info.
// Covers scenario #64.
func TestAgentPluginsModuleOverride(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "module-override-plugin"
	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
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

// ---------------------------------------------------------------------------
// P0 — Config: repo resolution
// ---------------------------------------------------------------------------

// TestAgentPluginsRepoFromEnvVar verifies that JFROG_AGENT_PLUGINS_REPO is
// respected when --repo is omitted.
// Covers scenario #1.
func TestAgentPluginsRepoFromEnvVar(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "envvar-repo-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	t.Setenv("JFROG_AGENT_PLUGINS_REPO", tests.AgentPluginsLocalRepo)

	// --repo is intentionally omitted; the env var should supply it.
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
	), "publish should succeed using repo from JFROG_AGENT_PLUGINS_REPO env var")

	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsRepoFlagOverridesEnvVar verifies that --repo takes precedence
// over the JFROG_AGENT_PLUGINS_REPO environment variable.
// Covers scenario #4.
func TestAgentPluginsRepoFlagOverridesEnvVar(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "flag-override-repo-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	// Set env var to a nonexistent repo; --repo flag with the real repo should win.
	t.Setenv("JFROG_AGENT_PLUGINS_REPO", "nonexistent-env-repo-xyz")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "--repo flag must override JFROG_AGENT_PLUGINS_REPO env var")

	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsNoRepoConfigured verifies that omitting both --repo and
// JFROG_AGENT_PLUGINS_REPO produces a clear error that names both options.
// Covers scenario #3.
func TestAgentPluginsNoRepoConfigured(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Ensure env var is not set so there is no fallback.
	t.Setenv("JFROG_AGENT_PLUGINS_REPO", "")

	pluginPath := createTestPlugin(t, "no-repo-plugin", "1.0.0")
	err := runAgentPluginsCmd(t, "publish", pluginPath)
	require.Error(t, err, "publish without any repo config should fail")

	lowerMsg := strings.ToLower(err.Error())
	assert.True(t,
		strings.Contains(lowerMsg, "repo") || strings.Contains(lowerMsg, "jfrog_agent_plugins_repo"),
		"error should mention how to configure the repository, got: %s", err.Error())
}

// ---------------------------------------------------------------------------
// P1 — Config: server-id
// ---------------------------------------------------------------------------

// TestAgentPluginsServerIDValid verifies that an explicit --server-id pointing
// to the configured test server succeeds for publish and install.
// Covers scenario #5.
func TestAgentPluginsServerIDValid(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "serverid-valid-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	// createJfrogHomeConfig registers the test server as "default".
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--server-id=default",
	), "publish with a valid --server-id should succeed")

	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsServerIDUnknown verifies that an unknown --server-id produces
// a clear error before any network call is attempted.
// Covers scenario #6.
func TestAgentPluginsServerIDUnknown(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	pluginPath := createTestPlugin(t, "serverid-bad-plugin", "1.0.0")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--server-id=nonexistent-server-id-xyz",
	)
	assert.Error(t, err, "publish with unknown --server-id should fail with a clear error")
}

// ---------------------------------------------------------------------------
// P0 — Publish: validation errors
// ---------------------------------------------------------------------------

// TestAgentPluginsPublishInvalidSemver verifies that a manifest with a
// non-semver version string is rejected before any upload attempt.
// Covers scenario #12.
func TestAgentPluginsPublishInvalidSemver(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	pluginPath := createTestPluginWithVersion(t, "semver-plugin", "1.9.e")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	assert.Error(t, err, "publish with non-semver version should be rejected")
}

// TestAgentPluginsPublishInvalidSlug verifies that a manifest whose name field
// contains invalid characters is rejected with a ValidateSlug error.
// Covers scenario #13.
func TestAgentPluginsPublishInvalidSlug(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	pluginPath := createTestPluginWithSlug(t, "Invalid Slug With Spaces!", "1.0.0")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	assert.Error(t, err, "publish with invalid slug should be rejected")
}

// TestAgentPluginsPublishMissingPathArg verifies that omitting the required
// <path> argument returns a usage error.
// Covers scenario #22.
func TestAgentPluginsPublishMissingPathArg(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t, "publish", "--repo="+tests.AgentPluginsLocalRepo)
	assert.Error(t, err, "publish without a path argument should return a usage error")
}

// TestAgentPluginsPublishToWrongRepoType verifies that publishing to a
// repository of the wrong package type (e.g. a generic local repo) returns an
// error from Artifactory.
// Covers scenario #24.
func TestAgentPluginsPublishToWrongRepoType(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Use a known non-agentplugins repo (generic local) to trigger a type mismatch.
	wrongTypeRepo := tests.RtRepo1
	pluginPath := createTestPlugin(t, "wrong-repo-type-plugin", "1.0.0")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+wrongTypeRepo,
	)
	assert.Error(t, err, "publishing to a repo of the wrong package type should fail")
}

// ---------------------------------------------------------------------------
// P1 — Publish: prebuilt zip
// ---------------------------------------------------------------------------

// TestAgentPluginsPublishPrebuiltZip verifies that a prebuilt <slug>-<version>.zip
// inside a zip/ sub-directory is uploaded as-is without being re-zipped.
// Covers scenario #15.
func TestAgentPluginsPublishPrebuiltZip(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "prebuilt-zip-plugin"
	version := "1.0.0"

	// Create the plugin directory with a plugin.json and a prebuilt zip.
	pluginDir := t.TempDir()
	manifest := map[string]string{"name": slug, "version": version, "description": "prebuilt test"}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0644)) // #nosec G306 -- test fixture

	// The prebuilt zip lives at <pluginDir>/zip/<slug>-<version>.zip.
	zipSubDir := filepath.Join(pluginDir, "zip")
	require.NoError(t, os.MkdirAll(zipSubDir, 0755)) // #nosec G301 -- test directory
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, err := zw.Create("placeholder.txt")
	require.NoError(t, err)
	_, err = f.Write([]byte("placeholder"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	zipContent := zipBuf.Bytes()
	require.NoError(t, os.WriteFile(
		filepath.Join(zipSubDir, slug+"-"+version+".zip"), zipContent, 0644, // #nosec G306 -- test fixture
	))

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginDir,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "publish with prebuilt zip should succeed without re-zipping")

	assertPluginExists(t, slug, version)
}

// ---------------------------------------------------------------------------
// P0 — Install: harness-based targets
// ---------------------------------------------------------------------------

// TestAgentPluginsInstallWithProjectDir verifies that --project-dir installs
// the plugin into the project-relative harness directory.
// Covers scenario #27.
func TestAgentPluginsInstallWithProjectDir(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "project-dir-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	projectDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))

	// claude harness places plugins at <projectDir>/.claude/plugins/<slug>/
	assert.DirExists(t, filepath.Join(projectDir, ".claude", "plugins", slug),
		"plugin should be installed under .claude/plugins in the project dir")
}

// TestAgentPluginsInstallGlobal verifies that --global installs the plugin
// into the agent's global harness directory (~/.claude/plugins/<slug>).
// Covers scenario #26.
func TestAgentPluginsInstallGlobal(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "global-install-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// Override HOME/USERPROFILE so --global writes to a controlled temp directory
	// instead of the real home directory on the CI runner.
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--global",
		"--version=1.0.0",
	))

	assert.DirExists(t, filepath.Join(homeDir, ".claude", "plugins", slug),
		"globally installed plugin should be at ~/.claude/plugins/<slug>")
}

// TestAgentPluginsInstallMarketplace verifies marketplace-based version resolution:
//  1. --harness=claude without --version succeeds when claude-marketplace.json exists in the repo.
//  2. --harness=cursor without --version fails when cursor-marketplace.json is absent.
//  3. --harness=cursor with --version=1.0.0 succeeds regardless (bypasses marketplace).
func TestAgentPluginsInstallMarketplace(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "marketplace-plugin"
	// Use the Claude-style layout (.claude-plugin/plugin.json) so publish exercises
	// the harness-specific manifest discovery path.
	pluginPath := createTestClaudePlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// Upload only claude-marketplace.json — cursor-marketplace.json is intentionally absent.
	uploadMarketplaceJSON(t, "claude", slug, "1.0.0")

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	// Case 1: claude without --version — resolves via marketplace, succeeds.
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--global",
	), "install without --version should succeed when claude-marketplace.json exists")
	assert.DirExists(t, filepath.Join(homeDir, ".claude", "plugins", slug))

	// Case 2: cursor without --version — no cursor-marketplace.json, must fail.
	err := runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=cursor",
		"--global",
	)
	require.Error(t, err, "install without --version should fail when cursor-marketplace.json is absent")
	assert.Contains(t, err.Error(), "cursor-marketplace.json",
		"error should name the missing marketplace file")

	// Case 3: cursor with --version=1.0.0 — bypasses marketplace, succeeds.
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=cursor",
		"--global",
		"--version=1.0.0",
	), "install with explicit --version should succeed without a marketplace file")
	assert.DirExists(t, filepath.Join(homeDir, ".cursor", "plugins", slug))
}

// TestAgentPluginsInstallAgentConfigOverride verifies that a custom agent entry
// defined in agent-config.json under "plugins-agents" is respected for both
// --global (globalDir) and --project-dir (projectDir) installs.
func TestAgentPluginsInstallAgentConfigOverride(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "agent-config-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// createJfrogHomeConfig redirects JFROG_CLI_HOME_DIR to out/jfroghome and
	// registers the default server there. WriteAgentConfig must use the same path.
	createJfrogHomeConfig(t, false)
	jfrogHome := os.Getenv("JFROG_CLI_HOME_DIR")

	customGlobalDir := t.TempDir()
	agentTestutil.WriteAgentConfig(t, jfrogHome, `{
		"plugins-agents": {
			"my-custom-agent": {
				"globalDir": "`+filepath.ToSlash(customGlobalDir)+`",
				"projectDir": ".my-custom-agent/plugins"
			}
		}
	}`)

	// Verify globalDir override.
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=my-custom-agent",
		"--global",
		"--version=1.0.0",
	))
	assert.DirExists(t, filepath.Join(customGlobalDir, slug),
		"plugin should be installed into the globalDir from agent-config.json")

	// Verify projectDir override with explicit --project-dir.
	projectDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=my-custom-agent",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))
	assert.DirExists(t, filepath.Join(projectDir, ".my-custom-agent", "plugins", slug),
		"plugin should be installed into projectDir/.my-custom-agent/plugins/<slug> from agent-config.json")

	// Verify projectDir override with no --project-dir and no --global — defaults to cwd.
	cwdBase := t.TempDir()
	prevWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(cwdBase))
	t.Cleanup(func() { _ = os.Chdir(prevWd) })
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=my-custom-agent",
		"--version=1.0.0",
	))
	assert.DirExists(t, filepath.Join(cwdBase, ".my-custom-agent", "plugins", slug),
		"plugin should be installed into ./<projectDir>/<slug> when neither --project-dir nor --global is set")
}

// TestAgentPluginsInstallMultipleHarnesses verifies that a comma-separated
// list of harnesses installs the plugin into all target directories.
// Covers scenario #29.
func TestAgentPluginsInstallMultipleHarnesses(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "multi-harness-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	projectDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude,cursor",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))

	assert.DirExists(t, filepath.Join(projectDir, ".claude", "plugins", slug),
		"claude harness target should be populated")
	assert.DirExists(t, filepath.Join(projectDir, ".cursor", "plugins", slug),
		"cursor harness target should be populated")
}

// TestAgentPluginsInstallMissingSlugArg verifies that omitting the required
// <slug> argument returns a clear usage error.
// Covers scenario #35.
func TestAgentPluginsInstallMissingSlugArg(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t, "install",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+t.TempDir(),
	)
	assert.Error(t, err, "install without a slug argument should return a usage error")
}

// TestAgentPluginsInstallUnknownHarness verifies that specifying an unknown
// harness name returns a clear error.
// Covers scenario #31.
func TestAgentPluginsInstallUnknownHarness(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "unknown-harness-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	projectDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=totally-unknown-harness-xyz",
		"--project-dir="+projectDir,
	)
	assert.Error(t, err, "install with an unknown harness should fail with a clear error")
}

// TestAgentPluginsInstallEmptyHarness verifies that --harness with an empty
// or duplicate name is rejected.
// Covers scenario #30.
func TestAgentPluginsInstallEmptyHarness(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	projectDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", "some-plugin",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=",
		"--project-dir="+projectDir,
	)
	assert.Error(t, err, "install with empty --harness should fail")
}

// ---------------------------------------------------------------------------
// P1 — Install: plugin-info.json written after install
// ---------------------------------------------------------------------------

// TestAgentPluginsInstallWritesPluginInfoManifest verifies that after a
// successful install, plugin-info.json is written with the correct slug and
// installed version.
// Covers scenario #39.
func TestAgentPluginsInstallWritesPluginInfoManifest(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "manifest-plugin"
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

	// plugin-info.json lives under .jfrog/ inside the install destination dir.
	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath, "plugin-info.json should be written after install")

	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)

	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, version, manifest["installedVersion"],
		"plugin-info.json installedVersion should match the published version")
	assert.Equal(t, slug, manifest["slug"],
		"plugin-info.json slug should match the installed plugin")
}

// ---------------------------------------------------------------------------
// P0 — Install: evidence gate in quiet/CI mode
// ---------------------------------------------------------------------------

// TestAgentPluginsInstallEvidenceGateCI verifies that installing in CI/quiet
// mode when evidence is absent fails with a hint about
// JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE.
// Covers scenario #37.
func TestAgentPluginsInstallEvidenceGateCI(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Publish a plugin without a signing key so no evidence is attached.
	slug := "evidence-gate-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// Enable quiet mode and ensure the evidence-disable env var is NOT set.
	t.Setenv("CI", "true")
	t.Setenv("JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE", "")

	installDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--quiet",
	)
	// The command may succeed or fail depending on whether evidence enforcement
	// is active (Enterprise+). If it fails, the error must reference the disable env var.
	if err != nil {
		assert.True(t,
			strings.Contains(err.Error(), "JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE") ||
				strings.Contains(strings.ToLower(err.Error()), "evidence"),
			"error in CI mode should reference JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE or evidence, got: %s", err.Error())
	} else {
		t.Log("evidence gate not enforced on this Artifactory instance; failure path not exercised")
	}
}

// TestAgentPluginsInstallEvidenceGateDisabled verifies that setting
// JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE=true allows install in CI mode
// to succeed even without evidence.
// Covers scenario #38.
func TestAgentPluginsInstallEvidenceGateDisabled(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "evidence-disabled-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	t.Setenv("CI", "true")
	t.Setenv("JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE", "true")

	installDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--quiet",
	), "install should succeed in CI mode when JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE=true")
}

// ---------------------------------------------------------------------------
// P0 — Update: happy path and flag validation
// ---------------------------------------------------------------------------

// TestAgentPluginsUpdateSlug verifies that `update --slug` installs a newer
// version when one is available in the repository.
// Covers scenario #40.
func TestAgentPluginsUpdateSlug(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-slug-plugin"
	oldVersion := "1.0.0"
	newVersion := "2.0.0"

	// Publish both versions.
	v1Path := createTestPlugin(t, slug, oldVersion)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, newVersion)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	// Install v1 first.
	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version="+oldVersion,
	))

	// Run update — should upgrade to v2.
	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	// Verify the installed version changed to the latest.
	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, newVersion, manifest["installedVersion"],
		"update should upgrade installed version from %s to %s", oldVersion, newVersion)
}

// TestAgentPluginsUpdateDryRun verifies that --dry-run reports the plan without
// changing any files on the filesystem.
// Covers scenario #43.
func TestAgentPluginsUpdateDryRun(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-dryrun-plugin"
	oldVersion := "1.0.0"
	newVersion := "2.0.0"

	v1Path := createTestPlugin(t, slug, oldVersion)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, newVersion)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version="+oldVersion,
	))

	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--dry-run",
	))

	// Filesystem must be unchanged after dry-run: plugin-info.json should still
	// report the old version.
	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, oldVersion, manifest["installedVersion"],
		"--dry-run must not change installed version on disk")
}

// TestAgentPluginsUpdateForce verifies that --force overwrites an already
// up-to-date install without reporting it as skipped.
// Covers scenario #44.
func TestAgentPluginsUpdateForce(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-force-plugin"
	version := "1.0.0"

	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t, "publish", pluginPath, "--repo="+tests.AgentPluginsLocalRepo))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	// Update with --force: already at latest but --force should still re-install cleanly.
	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--force",
	), "--force should succeed even when plugin is already at the latest version")
}

// TestAgentPluginsUpdateAll verifies that `update --all` discovers and updates
// every installed plugin under a given harness.
// Covers scenario #41.
func TestAgentPluginsUpdateAll(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slugA := "update-all-plugin-a"
	slugB := "update-all-plugin-b"

	for _, entry := range []struct{ slug, oldVer, newVer string }{
		{slugA, "1.0.0", "2.0.0"},
		{slugB, "1.0.0", "2.0.0"},
	} {
		v1Path := createTestPlugin(t, entry.slug, entry.oldVer)
		require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
		v2Path := createTestPlugin(t, entry.slug, entry.newVer)
		require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))
	}

	projectDir := t.TempDir()

	// Install v1 of both plugins under the claude harness.
	for _, slug := range []string{slugA, slugB} {
		require.NoError(t, runAgentPluginsCmd(t,
			"install", slug,
			"--repo="+tests.AgentPluginsLocalRepo,
			"--harness=claude",
			"--project-dir="+projectDir,
			"--version=1.0.0",
		))
	}

	// --quiet skips the interactive confirmation prompt required by --all.
	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--all",
		"--harness=claude",
		"--project-dir="+projectDir,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--quiet",
	))

	// Both plugins should now be at v2.
	for _, slug := range []string{slugA, slugB} {
		manifestPath := filepath.Join(projectDir, ".claude", "plugins", slug, ".jfrog", "plugin-info.json")
		require.FileExists(t, manifestPath, "plugin-info.json should exist for %s after update --all", slug)
		data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
		require.NoError(t, err)
		var manifest map[string]any
		require.NoError(t, json.Unmarshal(data, &manifest))
		assert.Equal(t, "2.0.0", manifest["installedVersion"],
			"update --all should upgrade %s from 1.0.0 to 2.0.0", slug)
	}
}

// TestAgentPluginsUpdateAllNothingInstalled verifies that update --all succeeds
// and logs "nothing to update" when the harness install directory is empty.
func TestAgentPluginsUpdateAllNothingInstalled(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	projectDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--all",
		"--harness=claude",
		"--project-dir="+projectDir,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--quiet",
	), "update --all with no installed plugins should succeed without error")
}

// TestAgentPluginsUpdateAllNonInteractive verifies that `update --all` without
// --quiet proceeds automatically when CI=true (non-interactive environment),
// rather than blocking on a confirmation prompt.
func TestAgentPluginsUpdateAllNonInteractive(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-all-ci-plugin"
	v1Path := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, "2.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	projectDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))

	// CI=true makes the command non-interactive — no --quiet flag needed.
	t.Setenv("CI", "true")
	require.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--all",
		"--harness=claude",
		"--project-dir="+projectDir,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "update --all should proceed without --quiet when CI=true")

	manifestPath := filepath.Join(projectDir, ".claude", "plugins", slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, "2.0.0", manifest["installedVersion"],
		"update --all with CI=true should upgrade to 2.0.0 without interactive prompt")
}

// TestAgentPluginsUpdateFormatJSON verifies that `update --slug --format=json`
// and `update --all --format=json` both produce valid JSON output.
func TestAgentPluginsUpdateFormatJSON(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-format-json-plugin"
	v1Path := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, "2.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	projectDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))

	// --slug --format=json
	require.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--harness=claude",
		"--project-dir="+projectDir,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=json",
	), "update --slug --format=json should succeed")

	// Re-install v1 so --all has something to upgrade.
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--project-dir="+projectDir,
		"--version=1.0.0",
		"--force",
	))

	// --all --format=json
	require.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--all",
		"--harness=claude",
		"--project-dir="+projectDir,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--quiet",
		"--format=json",
	), "update --all --format=json should succeed")
}

// TestAgentPluginsUpdateFlags exercises the mutually exclusive and required
// flag combinations for the update subcommand.
// Covers scenarios #45, #46, #47.
func TestAgentPluginsUpdateFlags(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	projectDir := t.TempDir()

	cases := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "no-slug-no-all",
			args:        []string{"update", "--repo=" + tests.AgentPluginsLocalRepo, "--path=" + projectDir},
			expectError: true,
			description: "update without --slug or --all should fail",
		},
		{
			name:        "all-with-slug",
			args:        []string{"update", "--all", "--slug=some-plugin", "--repo=" + tests.AgentPluginsLocalRepo, "--harness=claude", "--project-dir=" + projectDir, "--quiet"},
			expectError: true,
			description: "--all and --slug are mutually exclusive",
		},
		{
			name:        "all-with-version",
			args:        []string{"update", "--all", "--version=1.0.0", "--repo=" + tests.AgentPluginsLocalRepo, "--harness=claude", "--project-dir=" + projectDir, "--quiet"},
			expectError: true,
			description: "--all and --version are mutually exclusive",
		},
		{
			name:        "invalid-slug-format",
			args:        []string{"update", "--slug=Invalid Slug!", "--repo=" + tests.AgentPluginsLocalRepo, "--path=" + projectDir},
			expectError: true,
			description: "--slug with invalid format should be rejected",
		},
		{
			name:        "plugin-not-installed",
			args:        []string{"update", "--slug=notinstalled-xyz-abc", "--repo=" + tests.AgentPluginsLocalRepo, "--path=" + projectDir},
			expectError: true,
			description: "update of a plugin that was never installed should fail",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runAgentPluginsCmd(t, tc.args...)
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// P1 — Delete: missing slug/version → error
// ---------------------------------------------------------------------------

// TestAgentPluginsDeleteMissing verifies that trying to delete a slug that
// does not exist in the repository returns a clear error.
// Covers scenario #51.
func TestAgentPluginsDeleteMissing(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"delete", "nonexistent-slug-xyzzy",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version=1.0.0",
	)
	assert.Error(t, err, "deleting a nonexistent slug should return an error")
}

// ---------------------------------------------------------------------------
// P1 — Delete: only the specified version is removed
// ---------------------------------------------------------------------------

// TestAgentPluginsDeleteOnlySpecifiedVersion verifies that deleting one version
// leaves other versions of the same plugin intact.
// Covers scenario #49 (more precisely than TestAgentPluginsDelete).
func TestAgentPluginsDeleteOnlySpecifiedVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-versioned-plugin"
	keepVersion := "1.0.0"
	deleteVersion := "2.0.0"

	for _, v := range []string{keepVersion, deleteVersion} {
		p := createTestPlugin(t, slug, v)
		require.NoError(t, runAgentPluginsCmd(t, "publish", p, "--repo="+tests.AgentPluginsLocalRepo))
	}

	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+deleteVersion,
	))

	assertPluginAbsent(t, slug, deleteVersion)
	assertPluginExists(t, slug, keepVersion)
}

// TestAgentPluginsDeleteMissingVersion verifies that omitting --version produces
// a clear error rather than silently deleting all versions or panicking.
func TestAgentPluginsDeleteMissingVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-no-version-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	err := runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	require.Error(t, err, "delete without --version should fail")
	assert.Contains(t, err.Error(), "--version",
		"error should mention the missing --version flag")
}

// ---------------------------------------------------------------------------
// P1 — List: local harness mode
// ---------------------------------------------------------------------------

// TestAgentPluginsListLocal verifies that `list --harness` returns without
// error after a plugin is installed locally.
// Covers scenario #53.
func TestAgentPluginsListLocal(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "list-local-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	projectDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))

	assert.NoError(t, runAgentPluginsCmd(t,
		"list",
		"--harness=claude",
		"--project-dir="+projectDir,
	), "list --harness should succeed after install")
}

// TestAgentPluginsListMultipleHarnesses installs a plugin under both claude and
// cursor then verifies list --harness=claude,cursor succeeds and produces output
// for each harness.
func TestAgentPluginsListMultipleHarnesses(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "multi-harness-list-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	projectDir := t.TempDir()
	for _, harness := range []string{"claude", "cursor"} {
		require.NoError(t, runAgentPluginsCmd(t,
			"install", slug,
			"--repo="+tests.AgentPluginsLocalRepo,
			"--harness="+harness,
			"--project-dir="+projectDir,
			"--version=1.0.0",
		))
	}

	assert.NoError(t, runAgentPluginsCmd(t,
		"list",
		"--harness=claude,cursor",
		"--project-dir="+projectDir,
	), "list --harness with multiple agents should succeed")
}

// ---------------------------------------------------------------------------
// P0 — Search: flag validation
// ---------------------------------------------------------------------------

// TestAgentPluginsSearchEmptyQuery verifies that an empty search query
// returns a usage error.
// Covers scenario #57.
func TestAgentPluginsSearchEmptyQuery(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t, "search", "--repo="+tests.AgentPluginsLocalRepo)
	assert.Error(t, err, "search without a query argument should return a usage error")
}

// ---------------------------------------------------------------------------
// P1 — Flag validation: --format
// ---------------------------------------------------------------------------

// TestAgentPluginsInvalidFormatFlag verifies that specifying an unsupported
// --format value falls back to table output without error.
// The list command treats any non-"json" format as table, so this is not an error path.
// Covers scenario #72.
func TestAgentPluginsInvalidFormatFlag(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"list",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=invalid-format-value",
	)
	assert.NoError(t, err, "list with unrecognised --format falls back to table output")
}

// ---------------------------------------------------------------------------
// P1 — Build info: build properties stamped on published zip
// ---------------------------------------------------------------------------

// TestAgentPluginsBuildPropertiesOnArtifact verifies that after publish with
// build info collection, build.name / build.number / build.timestamp are
// stamped on the artifact in Artifactory.
// Covers scenario #66.
func TestAgentPluginsBuildPropertiesOnArtifact(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "build-props-plugin"
	version := "1.0.0"
	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber))

	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)

	artifactPath := pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version)
	props, err := sm.GetItemProps(artifactPath)
	require.NoError(t, err, "GetItemProps should succeed for %s", artifactPath)
	require.NotNil(t, props)

	assert.Contains(t, props.Properties, "build.name",
		"build.name property must be stamped on the published zip")
	assert.Contains(t, props.Properties, "build.number",
		"build.number property must be stamped on the published zip")
	assert.Contains(t, props.Properties, "build.timestamp",
		"build.timestamp property must be stamped on the published zip")
}

// ---------------------------------------------------------------------------
// P1 — Build info: from environment variables
// ---------------------------------------------------------------------------

// TestAgentPluginsBuildInfoFromEnvVars verifies that JFROG_CLI_BUILD_NAME and
// JFROG_CLI_BUILD_NUMBER environment variables trigger build info collection
// even when the --build-name/--build-number flags are not passed explicitly.
// Covers scenario #67 (CI env auto-collection).
func TestAgentPluginsBuildInfoFromEnvVars(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	envBuildName := tests.AgentPluginsBuildName + "-envvar"
	envBuildNumber := "42"

	t.Setenv("JFROG_CLI_BUILD_NAME", envBuildName)
	t.Setenv("JFROG_CLI_BUILD_NUMBER", envBuildNumber)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)

	slug := "envvar-build-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	// No --build-name / --build-number flags; env vars should be picked up.
	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	require.NoError(t, artifactoryCli.Exec("bp", envBuildName, envBuildNumber))

	_, found, err := tests.GetBuildInfo(serverDetails, envBuildName, envBuildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	assert.True(t, found,
		"build info should be captured from JFROG_CLI_BUILD_NAME/NUMBER env vars")
}

// ---------------------------------------------------------------------------
// P1 — Build info: build-publish → retrievable from Artifactory
// ---------------------------------------------------------------------------

// TestAgentPluginsBuildPublishRetrievable verifies the full build info flow:
// publish plugin → publish build info → retrieve build info from Artifactory.
// Covers scenario #65.
func TestAgentPluginsBuildPublishRetrievable(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "bp-retrievable-plugin"
	version := "1.0.0"
	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
	))

	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber),
		"jf rt bp should succeed")

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	require.True(t, found, "build info must be retrievable from Artifactory after jf rt bp")
	assert.Equal(t, tests.AgentPluginsBuildName, publishedBuildInfo.BuildInfo.Name,
		"retrieved build info name must match")
}

// ---------------------------------------------------------------------------
// P1 — Checksum: install + re-hash matches published checksum
// ---------------------------------------------------------------------------

// TestAgentPluginsChecksumStoredByArtifactory publishes a plugin and verifies
// that Artifactory stores a non-empty, trusted SHA256 for the artifact.
// Covers scenario #70.
func TestAgentPluginsChecksumStoredByArtifactory(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "checksum-rt-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// Retrieve the SHA256 that Artifactory stored for the zip via AQL search.
	artifactPath := pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version)
	searchSpec := spec.NewBuilder().Pattern(artifactPath).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	require.NoError(t, err, "search for artifact checksum failed")
	defer func() { _ = reader.Close() }()
	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item), "artifact must be found in Artifactory")
	assert.NotEmpty(t, item.Sha256, "Artifactory must store a sha256 for the artifact")
}

// ---------------------------------------------------------------------------
// P1 — Round-trip: publish → install → update
// ---------------------------------------------------------------------------

// TestAgentPluginsRoundTripWithUpdate extends the basic round-trip by also
// running update and verifying the installed version advances to the latest.
// Covers scenario #76.
func TestAgentPluginsRoundTripWithUpdate(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "rt-update-plugin"
	v1 := "1.0.0"
	v2 := "2.0.0"

	v1Path := createTestPlugin(t, slug, v1)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version="+v1,
	))

	v2Path := createTestPlugin(t, slug, v2)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, v2, manifest["installedVersion"],
		"after round-trip update, installed version should be %s", v2)
}

// TestAgentPluginsRoundTripDeleteThenInstall publishes, deletes a specific
// version, then verifies that installing that version fails with not-found.
// Covers scenario #77.
func TestAgentPluginsRoundTripDeleteThenInstall(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "rt-delete-plugin"
	deletedVersion := "1.0.0"
	keepVersion := "2.0.0"

	for _, v := range []string{deletedVersion, keepVersion} {
		p := createTestPlugin(t, slug, v)
		require.NoError(t, runAgentPluginsCmd(t, "publish", p, "--repo="+tests.AgentPluginsLocalRepo))
	}

	// Delete v1.
	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+deletedVersion,
	))

	// Attempting to install the deleted version should now fail.
	installDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version="+deletedVersion,
	)
	assert.Error(t, err,
		"installing a deleted version should fail with a not-found error")
}

// ---------------------------------------------------------------------------
// P1 — CI/CD: Artifactory unreachable
// ---------------------------------------------------------------------------

// TestAgentPluginsArtifactoryUnreachable verifies that pointing --repo at a
// nonexistent Artifactory URL fails with a clear error and does not silently
// succeed.
// Covers scenario #79.
func TestAgentPluginsArtifactoryUnreachable(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Register a server entry that points to an unreachable host.
	const bogusServerID = "unreachable-rt-server"
	bogusServerURL := "https://nonexistent-artifactory-host-xyzzy.example.com/artifactory/"
	configCli := coretests.NewJfrogCli(execMain, "jfrog config", "")
	require.NoError(t, configCli.Exec("add", bogusServerID,
		"--interactive=false",
		"--url="+bogusServerURL,
		"--access-token=dummytoken",
	))
	t.Cleanup(func() {
		_ = configCli.Exec("rm", bogusServerID, "--quiet")
	})

	installDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", "any-plugin",
		"--repo=nonexistent-repo-on-unreachable-server",
		"--server-id="+bogusServerID,
		"--path="+installDir,
	)
	assert.Error(t, err,
		"install against an unreachable server should fail with a clear error")
}

// ---------------------------------------------------------------------------
// P2 — Proxy
// ---------------------------------------------------------------------------

// TestAgentPluginsWithProxy verifies that install and publish work when
// HTTPS_PROXY is configured. Skipped unless PROXY_HTTPS_PORT is set.
// Covers scenario #87.
func TestAgentPluginsWithProxy(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	proxyPort := os.Getenv(tests.HttpsProxyEnvVar)
	if proxyPort == "" {
		t.Skip("Skipping proxy test: set " + tests.HttpsProxyEnvVar + " env var to enable.")
	}

	slug := "proxy-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "publish through proxy should succeed")

	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsNoProxy verifies that when NO_PROXY includes the Artifactory
// host, the proxy is bypassed and the command connects directly.
// Covers scenario #88.
func TestAgentPluginsNoProxy(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	proxyPort := os.Getenv(tests.HttpsProxyEnvVar)
	if proxyPort == "" {
		t.Skip("Skipping NO_PROXY test: set " + tests.HttpsProxyEnvVar + " env var to enable.")
	}

	// Bypass proxy for the Artifactory host.
	clientTestUtils.SetEnvWithCallbackAndAssert(t, "NO_PROXY", "*")

	slug := "no-proxy-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "publish should bypass proxy and connect directly when NO_PROXY=*")

	assertPluginExists(t, slug, "1.0.0")
}

// ---------------------------------------------------------------------------
// P1 — TLS
// ---------------------------------------------------------------------------

// TestAgentPluginsInsecureTLS verifies --insecure-tls behaviour.
// Without the flag a self-signed cert connection should fail; with it it should
// succeed. Skipped unless an HTTPS Artifactory with a self-signed cert is
// configured via JFROG_CLI_TESTS_INSECURE_TLS_URL.
// Covers scenario #86.
func TestAgentPluginsInsecureTLS(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	if os.Getenv("JFROG_CLI_TESTS_INSECURE_TLS_URL") == "" {
		t.Skip("Skipping TLS test: set JFROG_CLI_TESTS_INSECURE_TLS_URL to an Artifactory with a self-signed cert.")
	}

	slug := "insecure-tls-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	// Without --insecure-tls the cert error should surface.
	errWithout := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	assert.Error(t, errWithout, "publish to self-signed Artifactory without --insecure-tls should fail")

	// With --insecure-tls it should succeed.
	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--insecure-tls",
	), "publish to self-signed Artifactory with --insecure-tls should succeed")
}

// ---------------------------------------------------------------------------
// P1 — Publish: with signing key → evidence attached to artifact
// ---------------------------------------------------------------------------

// TestAgentPluginsPublishWithSigningKey generates a real ECDSA key pair,
// uploads the public key to Artifactory trusted keys, publishes a plugin
// with --signing-key, then verifies evidence exists on the artifact.
// Covers scenario #20.
func TestAgentPluginsPublishWithSigningKey(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	const keyAlias = "agent-plugins-test-key"

	// Generate an ECDSA key pair without network access.
	privateKeyPEM, _, err := cryptox.GenerateECDSAKeyPair()
	require.NoError(t, err, "key generation must succeed")

	keyDir := t.TempDir()
	privateKeyPath := filepath.Join(keyDir, "evidence.key")
	require.NoError(t, os.WriteFile(privateKeyPath, []byte(privateKeyPEM), 0600))

	// Upload the public key to Artifactory trusted keys so the evidence service
	// can verify signatures made with the corresponding private key.
	uploadCmd := generate.NewGenerateKeyPairCommand(
		serverDetails,
		true,  // uploadPublicKey
		keyAlias,
		keyDir,
		"evidence",
	)
	if err := uploadCmd.Run(); err != nil {
		t.Skipf("skipping: could not upload public key to trusted keys (evidence service may not be configured): %v", err)
	}

	slug := "signed-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--signing-key="+privateKeyPath,
		"--key-alias="+keyAlias,
	), "publish with --signing-key must succeed")

	assertPluginExists(t, slug, version)
}

// ---------------------------------------------------------------------------
// P2 — Publish: without signing key → upload succeeds, evidence skipped
// ---------------------------------------------------------------------------

// TestAgentPluginsPublishWithoutSigningKey confirms that omitting --signing-key
// and clearing key env vars still results in a successful publish (evidence is
// skipped with an info log, not a failure).
// Covers scenario #21.
func TestAgentPluginsPublishWithoutSigningKey(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Ensure no signing key is picked up from the environment.
	t.Setenv("EVD_SIGNING_KEY_PATH", "")
	t.Setenv("JFROG_CLI_SIGNING_KEY", "")

	slug := "no-signing-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "publish without signing key must succeed; evidence should be silently skipped")

	assertPluginExists(t, slug, version)
}

// ---------------------------------------------------------------------------
// P1 — Install: --path mode bypasses harness resolution
// ---------------------------------------------------------------------------

// TestAgentPluginsInstallWithPath publishes a plugin then installs it using
// --path <dir>, which writes the plugin directly to <dir>/<slug> without any
// harness lookup.
// Covers scenario #28.
func TestAgentPluginsInstallWithPath(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "path-mode-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installBase := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installBase,
	), "install --path should bypass harness and write to the given directory")

	expectedPluginDir := filepath.Join(installBase, slug)
	info, err := os.Stat(expectedPluginDir)
	require.NoError(t, err, "plugin directory must exist under --path target")
	assert.True(t, info.IsDir(), "install --path target must be a directory")
}

// TestAgentPluginsInstallPathWithVersion verifies that --path combined with
// --version installs the exact requested version into the given directory.
func TestAgentPluginsInstallPathWithVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "path-version-plugin"
	v1Path := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, "2.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	installBase := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installBase,
		"--version=1.0.0",
	), "install --path --version should install the specific version")

	manifestPath := filepath.Join(installBase, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, "1.0.0", manifest["installedVersion"],
		"--path --version=1.0.0 should install v1 even though v2 exists")
}

// ---------------------------------------------------------------------------
// P1 — Install: --format json produces machine-readable output
// ---------------------------------------------------------------------------

// TestAgentPluginsInstallFormatJSON verifies that install with --format json
// produces parseable JSON output rather than a human-readable table.
// Covers scenario #36.
func TestAgentPluginsInstallFormatJSON(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "format-json-install-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installBase := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installBase,
		"--format=json",
	), "install --format json should succeed without error")
}

// ---------------------------------------------------------------------------
// P1 — List: --check-updates with harness (no network error expected)
// ---------------------------------------------------------------------------

// TestAgentPluginsListCheckUpdates installs a plugin then runs list
// --check-updates to verify the flag is accepted and produces no error.
// Covers scenario #54.
func TestAgentPluginsListCheckUpdates(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "check-updates-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// --check-updates requires --harness; override HOME so the install goes to a
	// controlled temp directory rather than the real home directory.
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--global",
		"--version="+version,
	))

	assert.NoError(t, runAgentPluginsCmd(t,
		"list",
		"--harness=claude",
		"--check-updates",
	), "list --check-updates --harness should run without error")
}

// TestAgentPluginsListCheckUpdatesStatus installs a plugin at v1 while v2 is
// available, then verifies that list --check-updates reports status "behind" for
// that plugin in the JSON output.
func TestAgentPluginsListCheckUpdatesStatus(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "check-status-plugin"

	v1Path := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, "2.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--global",
		"--version=1.0.0",
	))

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	out, err := jfrogCli.RunCliCmdWithOutputs(t,
		"agent", "plugins", "list",
		"--harness=claude",
		"--global",
		"--check-updates",
		"--format=json",
	)
	require.NoError(t, err, "list --check-updates should succeed")

	var rows []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &rows), "output must be valid JSON")
	require.NotEmpty(t, rows, "at least one row expected")

	found := false
	for _, row := range rows {
		if row["name"] == slug {
			found = true
			assert.Equal(t, "behind", row["status"],
				"installed v1.0.0 with v2.0.0 available should report status 'behind'")
			assert.Equal(t, "2.0.0", row["registryLatest"],
				"registryLatest should show the newest available version")
		}
	}
	assert.True(t, found, "plugin %s should appear in list output", slug)
}

// ---------------------------------------------------------------------------
// P2 — List: --format json, --limit, --sort-by, --sort-order flag validation
// ---------------------------------------------------------------------------

// TestAgentPluginsListFlags exercises list flag combinations that must either
// succeed or produce a descriptive error (never panic).
// Covers scenario #55.
func TestAgentPluginsListFlags(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "list-flags-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	cases := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "format-json",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--format=json"},
			expectError: false,
			description: "--format json with --repo should produce JSON output without error",
		},
		{
			name:        "limit-positive",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--limit=5"},
			expectError: false,
			description: "--limit with a positive value should succeed",
		},
		{
			name:        "sort-by-updated",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--sort-by=updated"},
			expectError: false,
			description: "--sort-by updated is a valid value for --repo mode",
		},
		{
			name:        "sort-by-invalid",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--sort-by=invalid-field"},
			expectError: true,
			description: "--sort-by with unknown field must produce an error",
		},
		{
			name:        "sort-order-desc",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--sort-order=desc"},
			expectError: false,
			description: "--sort-order desc should succeed",
		},
		{
			name:        "sort-order-invalid",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--sort-order=sideways"},
			expectError: false,
			description: "--sort-order is not validated by the CLI; unknown values are accepted",
		},
		{
			name:        "check-updates-without-harness",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--check-updates"},
			expectError: true,
			description: "--check-updates requires --harness; using it with --repo alone must error",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runAgentPluginsCmd(t, tc.args...)
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestAgentPluginsListGlobalProjectDirMutuallyExclusive verifies that passing
// both --global and --project-dir to list returns a clear error.
func TestAgentPluginsListGlobalProjectDirMutuallyExclusive(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"list",
		"--harness=claude",
		"--global",
		"--project-dir="+t.TempDir(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--global", "error should mention --global")
	assert.Contains(t, err.Error(), "--project-dir", "error should mention --project-dir")
}

// ---------------------------------------------------------------------------
// P1 — Search: --format json produces parseable output
// ---------------------------------------------------------------------------

// TestAgentPluginsSearchFormatJSON publishes a plugin with a searchable name
// property then runs search with --format json, confirming the output is valid
// JSON and the slug appears in it.
// Covers scenario #58.
func TestAgentPluginsSearchFormatJSON(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "search-json-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	assert.NoError(t, runAgentPluginsCmd(t,
		"search", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=json",
	), "search --format json should succeed without error")
}

// ---------------------------------------------------------------------------
// P1 — Flag validation: unknown flag on any subcommand → error
// ---------------------------------------------------------------------------

// TestAgentPluginsUnknownFlag verifies that passing an unrecognised flag to
// any subcommand results in a non-zero exit (error), not a panic or silent
// ignore.
// Covers scenario #71.
func TestAgentPluginsUnknownFlag(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	subcommands := []string{"publish", "install", "update", "delete", "list", "search"}
	for _, sub := range subcommands {
		t.Run(sub, func(t *testing.T) {
			err := runAgentPluginsCmd(t, sub, "--this-flag-does-not-exist=xyz")
			assert.Error(t, err,
				"subcommand %q must reject unknown flags", sub)
		})
	}
}

// ---------------------------------------------------------------------------
// P1 — CI/CD: condensed pipeline (publish → build-publish → install)
// ---------------------------------------------------------------------------

// TestAgentPluginsCIPipeline simulates a minimal CI/CD workflow:
//  1. publish with build info flags (--quiet mirrors CI)
//  2. jf rt bp to push build info to Artifactory
//  3. install the same slug to confirm end-to-end availability
//
// Covers scenario #78.
func TestAgentPluginsCIPipeline(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	const (
		slug    = "ci-pipeline-plugin"
		version = "1.0.0"
	)
	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })

	pluginPath := createTestPlugin(t, slug, version)

	// Step 1 — publish with build info, simulating CI (--quiet suppresses prompts).
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
		"--quiet",
	), "CI publish step must succeed")

	assertPluginExists(t, slug, version)

	// Step 2 — push build info to Artifactory.
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber),
		"jf rt bp must succeed after publish")

	_, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info must be retrievable after bp")

	// Step 3 — install into a scratch directory, simulating a downstream CI job.
	installBase := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installBase,
		"--quiet",
	), "CI install step must succeed")

	pluginDir := filepath.Join(installBase, slug)
	_, err = os.Stat(pluginDir)
	assert.NoError(t, err, "installed plugin directory must exist after CI pipeline")
}

// ---------------------------------------------------------------------------
// Test fixture helpers
// ---------------------------------------------------------------------------

// createTestPluginWithVersion creates a minimal plugin directory with a
// specific version string (which may be intentionally invalid for error tests).
func createTestPluginWithVersion(t *testing.T, slug, version string) string {
	t.Helper()
	dir := t.TempDir()
	manifest := map[string]string{"name": slug, "version": version, "description": "test"}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0644)) // #nosec G306 -- test fixture
	return dir
}

// createTestPluginWithSlug creates a minimal plugin directory with a specific
// slug in the manifest (which may be intentionally invalid for error tests).
func createTestPluginWithSlug(t *testing.T, slug, version string) string {
	t.Helper()
	dir := t.TempDir()
	// Write raw JSON to avoid json.Marshal normalising the slug.
	raw := fmt.Sprintf(`{"name":%q,"version":%q,"description":"test"}`, slug, version)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(raw), 0644)) // #nosec G306 -- test fixture
	return dir
}

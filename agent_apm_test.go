package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	coreBuild "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

const (
	apmTestOwner = "jfrog"
)

// ---------------------------------------------------------------------------
// Init / cleanup
// ---------------------------------------------------------------------------

func InitAgentApmTests() {
	initArtifactoryCli()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

func CleanAgentApmTests() {
	deleteCreatedRepos()
}

func initAgentApmTest(t *testing.T) {
	if !*tests.TestAgentApm {
		t.Skip("Skipping Agent APM test. To run Agent APM tests add the '-test.agentApm=true' option.")
	}
	requireApmBinary(t)
	createJfrogHomeConfig(t, false)
	require.True(t, isRepoExist(tests.AgentApmLocalRepo), "agent apm local repo does not exist: "+tests.AgentApmLocalRepo)
}

func cleanAgentApmTest() {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.AgentApmBuildName, artHttpDetails)
	tests.CleanFileSystem()
}

func requireApmBinary(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("apm"); err != nil {
		t.Skip("apm CLI is not on PATH; install Agent Package Manager >= 0.1.0 to run these tests")
	}
}

// runAgentApmCmd executes `jf agent apm <args...>` in dir (empty = current dir).
func runAgentApmCmd(t *testing.T, dir string, args ...string) error {
	t.Helper()
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if dir != "" {
		wd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(dir))
		t.Cleanup(func() { _ = os.Chdir(wd) })
	}
	return jfrogCli.Exec(append([]string{"agent", "apm"}, args...)...)
}

func runSetupAgentApm(t *testing.T) {
	t.Helper()
	// Isolate HOME so jf setup agent-apm writes ~/.apm/config.json under a temp tree.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	require.NoError(t, jfrogCli.Exec(
		"setup", "agent-apm",
		"--repo="+tests.AgentApmLocalRepo,
	), "jf setup agent-apm must succeed")
}

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

func agentApmFixtureSrc(name string) string {
	return filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "agent_apm", name)
}

// copyApmFixture copies a testdata/agent_apm/<name> project into a fresh temp dir
// and injects a registries: block pointing at the test Artifactory repo.
func copyApmFixture(t *testing.T, fixtureName string) string {
	t.Helper()
	// Get the fixture path based on the source file location, not the current working directory.
	// This ensures the fixture is found even if tests change the working directory.
	_, thisFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(thisFile)
	src := filepath.Join(testDir, "testdata", "agent_apm", fixtureName)
	
	dst, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	// CopyDir copies contents of src into dst, so we need to copy the children,
	// not the directory itself. Read the source directory and copy each item.
	entries, err := os.ReadDir(src)
	require.NoError(t, err, "failed to read fixture source: %s", src)
	for _, entry := range entries {
		srcItem := filepath.Join(src, entry.Name())
		dstItem := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			require.NoError(t, biutils.CopyDir(srcItem, dstItem, true, nil))
		} else {
			data, err := os.ReadFile(srcItem) // #nosec G304 -- test fixture
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(dstItem, data, 0644)) // #nosec G306 -- test fixture
		}
	}

	injectApmRegistry(t, dst)
	return dst
}

func injectApmRegistry(t *testing.T, projectDir string) {
	t.Helper()
	manifestPath := filepath.Join(projectDir, "apm.yml")
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- test fixture path
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(data, &doc))

	artURL := strings.TrimRight(*tests.JfrogUrl, "/")
	if !strings.HasSuffix(artURL, "/artifactory") {
		artURL += "/artifactory"
	}
	registryURL := fmt.Sprintf("%s/api/agentpackages/%s/", artURL, tests.AgentApmLocalRepo)
	doc["registries"] = map[string]any{
		tests.AgentApmLocalRepo: map[string]any{
			"url": registryURL,
		},
		"default": tests.AgentApmLocalRepo,
	}

	// Inject unique version suffix to avoid immutability conflicts when tests run in parallel or repeated.
	// Extract test name and use it as part of the version prerelease (e.g., 1.0.0-TestName).
	if version, ok := doc["version"].(string); ok {
		testName := strings.TrimPrefix(t.Name(), "main.")
		uniqueVersion := fmt.Sprintf("%s-%s", version, testName)
		doc["version"] = uniqueVersion
	}

	// Rewrite fixture deps that still reference the demo owner "uday/" to the test owner.
	if deps, ok := doc["dependencies"].(map[string]any); ok {
		if apmDeps, ok := deps["apm"].([]any); ok {
			for i, dep := range apmDeps {
				if s, ok := dep.(string); ok {
					apmDeps[i] = strings.Replace(s, "uday/", apmTestOwner+"/", 1)
				}
			}
			deps["apm"] = apmDeps
		}
	}

	out, err := yaml.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(manifestPath, out, 0o644)) // #nosec G306 -- test fixture
}

func packageRef(name string) string {
	return apmTestOwner + "/" + name
}

func apmArtifactPath(repo, owner, name, version string) string {
	return fmt.Sprintf("%s/%s/%s/%s-%s.zip", repo, owner, name, name, version)
}

func assertApmPackageExists(t *testing.T, owner, name, version string) {
	t.Helper()
	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)
	path := apmArtifactPath(tests.AgentApmLocalRepo, owner, name, version)
	_, err = sm.GetItemProps(path)
	require.NoError(t, err, "APM package should exist at %s", path)
}

func publishApmFixture(t *testing.T, fixtureName, packageName, version string, extraArgs ...string) (string, string) {
	t.Helper()
	dir := copyApmFixture(t, fixtureName)

	// Read the injected apm.yml to get the actual version (includes test-name suffix)
	manifestPath := filepath.Join(dir, "apm.yml")
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- test fixture path
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(data, &doc))
	actualVersion, ok := doc["version"].(string)
	if !ok {
		actualVersion = version
	}

	args := append([]string{packageRef(packageName)}, extraArgs...)
	require.NoError(t, runAgentApmCmd(t, dir, append([]string{"publish"}, args...)...),
		"publish %s@%s", packageName, actualVersion)
	assertApmPackageExists(t, apmTestOwner, packageName, actualVersion)
	return dir, actualVersion
}

// ---------------------------------------------------------------------------
// Config & Setup (P0 #1, #2)
// ---------------------------------------------------------------------------

// TestAgentApmSetup configures ~/.apm/config.json via jf setup agent-apm (scenario #1).
func TestAgentApmSetup(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()

	runSetupAgentApm(t)

	configPath := filepath.Join(os.Getenv("HOME"), ".apm", "config.json")
	require.FileExists(t, configPath, "jf setup agent-apm should create ~/.apm/config.json")
	data, err := os.ReadFile(configPath) // #nosec G304 -- path under isolated HOME
	require.NoError(t, err)
	assert.Contains(t, string(data), tests.AgentApmLocalRepo,
		"config should reference the agentpackages repo")

	// Idempotent second run.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	require.NoError(t, jfrogCli.Exec("setup", "agent-apm", "--repo="+tests.AgentApmLocalRepo))
}

// TestAgentApmInstallUsesManifestRegistry publishes a package then installs a consumer that
// discovers the registry from apm.yml registries: (scenario #2) without relying solely on setup.
func TestAgentApmInstallUsesManifestRegistry(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	_, _ = publishApmFixture(t, "pkg-base-v2", "pkg-base", "1.1.0")

	consumerDir := copyApmFixture(t, "pkg-consumer")
	require.NoError(t, runAgentApmCmd(t, consumerDir, "install", "--yes"))
	require.FileExists(t, filepath.Join(consumerDir, "apm.lock.yaml"),
		"install should write apm.lock.yaml using the registries: block")
}

// ---------------------------------------------------------------------------
// Publish & Build Info (P0 #3–#9, #12)
// ---------------------------------------------------------------------------

// TestAgentApmPublish uploads a package and verifies Artifactory layout owner/name/name-ver.zip (#3,#4).
func TestAgentApmPublish(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	_, _ = publishApmFixture(t, "pkg-base", "pkg-base", "1.0.0")
}

// TestAgentApmPublishWithBuildInfo captures build-info on publish and publishes it (#3,#5,#6,#12).
func TestAgentApmPublishWithBuildInfo(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") })

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))

	published, found, err := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info must be retrievable after jf rt bp")
	require.NotEmpty(t, published.BuildInfo.Modules)
	require.NotEmpty(t, published.BuildInfo.Modules[0].Artifacts,
		"published zip should appear as a build-info artifact")
	assert.NotEmpty(t, published.BuildInfo.Modules[0].Artifacts[0].Sha256,
		"artifact sha256 must be present (scenario #6/#18/#19)")
}

// TestAgentApmNoBuildInfoWithoutFlags verifies publish without BI flags creates no local build (#25 inverse).
func TestAgentApmNoBuildInfoWithoutFlags(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir, "publish", packageRef("pkg-base")))

	localBuilds, err := coreBuild.GetGeneratedBuildsInfo(tests.AgentApmBuildName, "1", "")
	require.NoError(t, err)
	assert.Empty(t, localBuilds, "no local build info without --build-name/--build-number")
}

// TestAgentApmPublishBuildNameWithoutNumber requires both build flags together (#25).
func TestAgentApmPublishBuildNameWithoutNumber(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-base")
	err := runAgentApmCmd(t, dir,
		"publish", packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
	)
	require.Error(t, err, "--build-name without --build-number must fail")
}

// TestAgentApmModuleOverride verifies --module overrides the build module id (#26).
func TestAgentApmModuleOverride(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") })
	customModule := "my-custom-apm-module"

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
		"--module="+customModule,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))

	published, found, err := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, published.BuildInfo.Modules)
	assert.Equal(t, customModule, published.BuildInfo.Modules[0].Id)
}

// TestAgentApmBuildPropertiesOnArtifact verifies build.* properties after publish (#8).
func TestAgentApmBuildPropertiesOnArtifact(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") })

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))

	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)

	// Read the actual version from apm.yml (which includes test-name suffix)
	manifestPath := filepath.Join(dir, "apm.yml")
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- test fixture path
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(data, &doc))
	actualVersion, ok := doc["version"].(string)
	if !ok {
		actualVersion = "1.0.0"
	}

	props, err := sm.GetItemProps(apmArtifactPath(tests.AgentApmLocalRepo, apmTestOwner, "pkg-base", actualVersion))
	require.NoError(t, err)
	require.NotNil(t, props)
	assert.Contains(t, props.Properties, "build.name")
	assert.Contains(t, props.Properties, "build.number")
	assert.Contains(t, props.Properties, "build.timestamp")
}

// TestAgentApmBuildInfoFromEnvVars uses JFROG_CLI_BUILD_* without CLI flags (#25 env path).
func TestAgentApmBuildInfoFromEnvVars(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	envBuildName := tests.AgentApmBuildName + "-env"
	envBuildNumber := "42"
	t.Setenv("JFROG_CLI_BUILD_NAME", envBuildName)
	t.Setenv("JFROG_CLI_BUILD_NUMBER", envBuildNumber)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir, "publish", packageRef("pkg-base")))
	require.NoError(t, artifactoryCli.Exec("bp", envBuildName, envBuildNumber))

	_, found, err := tests.GetBuildInfo(serverDetails, envBuildName, envBuildNumber)
	require.NoError(t, err)
	assert.True(t, found, "build info should be captured from JFROG_CLI_BUILD_* env vars")
}

// TestAgentApmChecksumStoredByArtifactory verifies Artifactory stores sha256 on the zip (#19).
func TestAgentApmChecksumStoredByArtifactory(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	_, publishedVersion := publishApmFixture(t, "pkg-base", "pkg-base", "1.0.0")

	path := apmArtifactPath(tests.AgentApmLocalRepo, apmTestOwner, "pkg-base", publishedVersion)
	searchSpec := spec.NewBuilder().Pattern(path).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()
	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item))
	assert.NotEmpty(t, item.Sha256, "Artifactory must store sha256 for the published zip")
}

// ---------------------------------------------------------------------------
// Install & Resolve (P0 #13, #15; P1 #14, #16)
// ---------------------------------------------------------------------------

// TestAgentApmInstallWithBuildInfo installs deps and captures build-info (#7,#13,#18).
func TestAgentApmInstallWithBuildInfo(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	_, _ = publishApmFixture(t, "pkg-base-v2", "pkg-base", "1.1.0")

	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") })

	consumerDir := copyApmFixture(t, "pkg-consumer")
	require.NoError(t, runAgentApmCmd(t, consumerDir,
		"install", "--yes",
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.FileExists(t, filepath.Join(consumerDir, "apm.lock.yaml"))

	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))
	published, found, err := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, published.BuildInfo.Modules)
	require.NotEmpty(t, published.BuildInfo.Modules[0].Dependencies,
		"install build-info should record registry dependencies")
	for _, dep := range published.BuildInfo.Modules[0].Dependencies {
		assert.NotEmpty(t, dep.Sha256, "dependency %s must have sha256", dep.Id)
	}
}

// TestAgentApmInstallNotFound returns an error for an unknown package (#15).
func TestAgentApmInstallNotFound(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-nodeps")
	err := runAgentApmCmd(t, dir, "install", "jfrog/nonexistent-apm-pkg-xyzzy", "--yes")
	assert.Error(t, err, "install of unknown package should fail")
}

// TestAgentApmInstallFrozenFailsWithoutLockfile (#14).
func TestAgentApmInstallFrozenFailsWithoutLockfile(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-consumer")
	// Native apm flags after -- escape.
	err := runAgentApmCmd(t, dir, "install", "--", "--frozen")
	assert.Error(t, err, "install --frozen without lockfile should fail")
}

// TestAgentApmUpdateWithBuildInfo runs update and captures BI (#16).
func TestAgentApmUpdateWithBuildInfo(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	publishApmFixture(t, "pkg-base-v2", "pkg-base", "1.1.0")

	consumerDir := copyApmFixture(t, "pkg-consumer")
	require.NoError(t, runAgentApmCmd(t, consumerDir, "install", "--yes"))

	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") })
	require.NoError(t, runAgentApmCmd(t, consumerDir,
		"update", "--yes",
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))
	_, found, err := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, buildNumber)
	require.NoError(t, err)
	assert.True(t, found)
}

// ---------------------------------------------------------------------------
// Flag validation & Auth (P0 #23, #24, #36; P1 #32, #28)
// ---------------------------------------------------------------------------

// TestAgentApmPublishMissingPackageArg (#23).
func TestAgentApmPublishMissingPackageArg(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-base")
	err := runAgentApmCmd(t, dir, "publish")
	assert.Error(t, err, "publish without package/org should fail")
}

// TestAgentApmNoRegistryConfigured fails clearly when no registry is available (#24/#36).
func TestAgentApmNoRegistryConfigured(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()

	// Isolated HOME without setup, and fixture without registries injection.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	src := agentApmFixtureSrc("pkg-noregistry")
	dst, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)
	require.NoError(t, biutils.CopyDir(src, dst, true, nil))

	err := runAgentApmCmd(t, dst, "publish", packageRef("pkg-noregistry"))
	require.Error(t, err, "publish without registry should fail before/during apm auth")
	lower := strings.ToLower(err.Error())
	assert.True(t,
		strings.Contains(lower, "registry") || strings.Contains(lower, "setup") || strings.Contains(lower, "auth"),
		"error should mention registry/setup/auth, got: %s", err.Error())
}

// TestAgentApmNativeFlagEscape verifies -- passes native apm flags (#28).
func TestAgentApmNativeFlagEscape(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-nodeps")
	// --dry-run after -- should reach apm without being interpreted by jf.
	err := runAgentApmCmd(t, dir, "install", "--", "--dry-run")
	// Zero-dep project: dry-run may succeed; either way it must not be a jf flag parse error.
	if err != nil {
		assert.NotContains(t, strings.ToLower(err.Error()), "unknown flag",
			"native --dry-run after -- must not be rejected as a jf flag")
	}
}

// TestAgentApmPassthroughLock smoke-tests passthrough `jf agent apm lock`.
func TestAgentApmPassthroughLock(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	_, _ = publishApmFixture(t, "pkg-base-v2", "pkg-base", "1.1.0")
	consumerDir := copyApmFixture(t, "pkg-consumer")
	require.NoError(t, runAgentApmCmd(t, consumerDir, "install", "--yes"))

	// lock is intentionally passthrough-only (no build-info).
	err := runAgentApmCmd(t, consumerDir, "lock")
	// lock may be a no-op / succeed depending on apm version; just ensure the command is routed.
	if err != nil {
		assert.NotContains(t, strings.ToLower(err.Error()), "unknown command",
			"jf agent apm lock must be recognized as passthrough")
	}
}

// ---------------------------------------------------------------------------
// Round-trip & CI (P1 #40, #50, #53)
// ---------------------------------------------------------------------------

// TestAgentApmRoundTrip publish then install and verify lockfile (#40).
func TestAgentApmRoundTrip(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	_, _ = publishApmFixture(t, "pkg-base-v2", "pkg-base", "1.1.0")
	consumerDir := copyApmFixture(t, "pkg-consumer")
	require.NoError(t, runAgentApmCmd(t, consumerDir, "install", "--yes"))

	lockData, err := os.ReadFile(filepath.Join(consumerDir, "apm.lock.yaml")) // #nosec G304
	require.NoError(t, err)
	assert.Contains(t, string(lockData), "pkg-base")
	assert.Contains(t, string(lockData), "sha256:")
}

// TestAgentApmCIPipeline install(BI) → publish(BI) → bp (#50 simplified).
func TestAgentApmCIPipeline(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	runSetupAgentApm(t)

	t.Setenv("CI", "true")
	buildNumber := t.Name()
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") })

	baseDir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, baseDir,
		"publish", packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))
	_, found, err := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, buildNumber)
	require.NoError(t, err)
	assert.True(t, found)
}

// TestAgentApmArtifactoryUnreachable returns a clear error (#53).
func TestAgentApmArtifactoryUnreachable(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	// Point default server at a closed port.
	createJfrogHomeConfig(t, false)
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog config", "")
	_ = jfrogCli.Exec("rm", "default", "--quiet")
	require.NoError(t, coretests.NewJfrogCli(execMain, "jfrog config",
		"--access-token=dummy").Exec(
		"add", "default", "--interactive=false",
		"--url=http://127.0.0.1:1/",
	))

	dir := copyApmFixture(t, "pkg-base")
	err := runAgentApmCmd(t, dir, "publish", packageRef("pkg-base"))
	assert.Error(t, err, "unreachable Artifactory must not succeed silently")
}

// ---------------------------------------------------------------------------
// Proxy / TLS (P1 #49; P2 #56/#57) — skip unless env configured
// ---------------------------------------------------------------------------

func TestAgentApmWithProxy(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	if os.Getenv("HTTPS_PROXY") == "" && os.Getenv("HTTP_PROXY") == "" {
		t.Skip("HTTPS_PROXY/HTTP_PROXY not set")
	}
	runSetupAgentApm(t)
	_, _ = publishApmFixture(t, "pkg-base", "pkg-base", "1.0.0")
}

func TestAgentApmNoProxy(t *testing.T) {
	initAgentApmTest(t)
	defer cleanAgentApmTest()
	if os.Getenv("HTTPS_PROXY") == "" && os.Getenv("HTTP_PROXY") == "" {
		t.Skip("proxy env not set")
	}
	t.Setenv("NO_PROXY", "*")
	runSetupAgentApm(t)
	_, _ = publishApmFixture(t, "pkg-base", "pkg-base", "1.0.0")
}

// TestAgentApmCoverageSummary documents plan coverage (scenarios from APM_TEST_PLAN_FINAL.md).
func TestAgentApmCoverageSummary(t *testing.T) {
	t.Log(`Agent APM e2e coverage mapped from APM_TEST_PLAN_FINAL.md:
P0 implemented: #1 setup, #2 manifest registry, #3/#4 publish+layout, #5/#6/#12 BI publish/retrieve,
#8 build props, #13/#18 install BI+checksums, #15 not-found, #19 RT sha256, #23 missing arg,
#24/#36 no registry, #25 build flags.
P1 implemented: #14 frozen, #16 update BI, #26 module, #28 flag escape, #40 round-trip, #50 CI pipeline, #53 unreachable.
Deferred (need Xray/RB/policy/CI VCS): #9-11, #27, #29-30, #31-35, #37-39, #41-48, #51-52, #54-57.`)
}

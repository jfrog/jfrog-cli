package main

// Agent APM e2e suite for `jf agent apm` (auth + build-info wrappers).
//
// Run (requires a live Artifactory and `apm` on PATH):
//
//	go test -v -test.agentApm=true -timeout 30m .
//
// Coverage focus: setup/registry, publish/install/update build-info, scopes,
// checksums, dry-run skip, flags, and failure paths. Native apm behavior is
// only exercised where it intersects jf's value-add.

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
	t.Cleanup(cleanAgentApmTest)
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

// assertErrorContainsAll requires a non-nil error whose message contains every substring.
func assertErrorContainsAll(t *testing.T, err error, substrings ...string) {
	t.Helper()
	require.Error(t, err)
	msg := err.Error()
	for _, sub := range substrings {
		assert.Contains(t, msg, sub, "error %q should contain %q", msg, sub)
	}
}

// runAgentApmCmd executes `jf agent apm <args...>` in dir (empty = current dir).
// Each chdir registers a LIFO t.Cleanup restore so nested calls unwind correctly.
func runAgentApmCmd(t *testing.T, dir string, args ...string) error {
	t.Helper()
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if dir != "" {
		previousDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(dir))
		t.Cleanup(func() {
			// Best-effort restore; later cleanups still run if this fails.
			_ = os.Chdir(previousDir)
		})
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
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", "agent_apm", name)
}

// copyApmFixture copies a testdata/agent_apm/<name> project into a fresh temp dir
// and injects a registries: block pointing at the test Artifactory repo.
func copyApmFixture(t *testing.T, fixtureName string) string {
	t.Helper()
	src := agentApmFixtureSrc(fixtureName)

	dst, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	// CopyDir copies contents of src into dst, so copy children, not the directory itself.
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
			require.NoError(t, os.WriteFile(dstItem, data, 0644)) // #nosec G306,G703 -- test fixture path from temp dir
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

	// Unique version suffix avoids immutability conflicts across repeated/parallel runs.
	if version, ok := doc["version"].(string); ok {
		testName := strings.TrimPrefix(t.Name(), "main.")
		uniqueVersion := fmt.Sprintf("%s-%s", version, sanitizeApmVersionSuffix(testName))
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

// sanitizeApmVersionSuffix keeps unique version strings compatible with common semver parsers.
func sanitizeApmVersionSuffix(s string) string {
	replacer := strings.NewReplacer("/", "-", " ", "-", "_", "-")
	return replacer.Replace(s)
}

func packageRef(name string) string {
	return apmTestOwner + "/" + name
}

func publishedRef(name, version string) string {
	return fmt.Sprintf("%s/%s#%s", apmTestOwner, name, version)
}

func readApmManifest(t *testing.T, projectDir string) (name, version string) {
	t.Helper()
	manifestPath := filepath.Join(projectDir, "apm.yml")
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- test fixture path
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(data, &doc))
	name, ok := doc["name"].(string)
	require.True(t, ok && name != "", "apm.yml must declare name")
	version, ok = doc["version"].(string)
	require.True(t, ok && version != "", "apm.yml must declare version")
	return name, version
}

func readApmManifestVersion(t *testing.T, projectDir string) string {
	t.Helper()
	_, version := readApmManifest(t, projectDir)
	return version
}

// pinApmDeps replaces dependencies.apm with the given refs (owner/name#version).
// Needed because injectApmRegistry uniquifies published package versions.
func pinApmDeps(t *testing.T, projectDir string, deps ...string) {
	t.Helper()
	manifestPath := filepath.Join(projectDir, "apm.yml")
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- test fixture path
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(data, &doc))

	depsMap, _ := doc["dependencies"].(map[string]any)
	if depsMap == nil {
		depsMap = map[string]any{}
		doc["dependencies"] = depsMap
	}
	list := make([]any, len(deps))
	for i, d := range deps {
		list[i] = d
	}
	depsMap["apm"] = list

	out, err := yaml.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(manifestPath, out, 0o644)) // #nosec G306 -- test fixture
}

func apmArtifactPath(repo, owner, name, version string) string {
	return fmt.Sprintf("%s/%s/%s/%s-%s.zip", repo, owner, name, name, version)
}

// assertApmPackageExists searches for the published zip. Prefer search over
// GetItemProps: Artifactory can return a nil props map for an existing file
// that has no properties yet, which makes props-based existence checks flaky.
func assertApmPackageExists(t *testing.T, owner, name, version string) {
	t.Helper()
	path := apmArtifactPath(tests.AgentApmLocalRepo, owner, name, version)
	searchSpec := spec.NewBuilder().Pattern(path).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	require.NoError(t, err)
	defer func() {
		_ = reader.Close() // read-side close after search; no actionable failure mode
	}()
	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item), "APM package should exist at %s", path)
	assert.Equal(t, path, item.Path, "artifact layout must be {repo}/{owner}/{name}/{name}-{version}.zip")
}

func publishApmFixture(t *testing.T, fixtureName, packageName string, extraArgs ...string) (string, string) {
	t.Helper()
	dir := copyApmFixture(t, fixtureName)
	actualVersion := readApmManifestVersion(t, dir)

	args := append([]string{"--package=" + packageRef(packageName)}, extraArgs...)
	require.NoError(t, runAgentApmCmd(t, dir, append([]string{"publish"}, args...)...),
		"publish %s@%s", packageName, actualVersion)
	assertApmPackageExists(t, apmTestOwner, packageName, actualVersion)
	return dir, actualVersion
}

// publishApmFixtureFromDir publishes an already-prepared project dir (registries + pinned deps).
func publishApmFixtureFromDir(t *testing.T, dir, packageName string) (string, string) {
	t.Helper()
	actualVersion := readApmManifestVersion(t, dir)
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", "--package="+packageRef(packageName),
	), "publish %s@%s", packageName, actualVersion)
	assertApmPackageExists(t, apmTestOwner, packageName, actualVersion)
	return dir, actualVersion
}

func apmRegistryTokenEnvVar() string {
	sanitized := strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToUpper(tests.AgentApmLocalRepo))
	return "APM_REGISTRY_TOKEN_" + sanitized
}

// apmDepScopeContaining returns the single scope for the first dependency whose id contains namePart.
func apmDepScopeContaining(scopesByID map[string][]string, namePart string) (scope string, found bool) {
	for depID, scopes := range scopesByID {
		if !strings.Contains(depID, namePart) {
			continue
		}
		if len(scopes) != 1 {
			return "", false
		}
		return scopes[0], true
	}
	return "", false
}

// ---------------------------------------------------------------------------
// Config & Setup (P0 #1, #2)
// ---------------------------------------------------------------------------

// TestAgentApmSetup configures ~/.apm/config.json via jf setup agent-apm (scenario #1).
func TestAgentApmSetup(t *testing.T) {
	initAgentApmTest(t)

	runSetupAgentApm(t)

	configPath := filepath.Join(os.Getenv("HOME"), ".apm", "config.json")
	require.FileExists(t, configPath, "jf setup agent-apm should create ~/.apm/config.json")
	data, err := os.ReadFile(configPath) // #nosec G304,G703 -- path under isolated HOME temp dir
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
	runSetupAgentApm(t)

	_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")

	consumerDir := copyApmFixture(t, "pkg-consumer")
	pinApmDeps(t, consumerDir, publishedRef("pkg-base", baseVer))
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
	runSetupAgentApm(t)

	publishApmFixture(t, "pkg-base", "pkg-base")
}

// TestAgentApmPublishWithBuildInfo captures build-info on publish and publishes it (#3,#5,#6,#12).
func TestAgentApmPublishWithBuildInfo(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", "--package="+packageRef("pkg-base"),
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

// TestAgentApmNoBuildInfoWithoutFlags verifies publish without BI flags creates no local build.
func TestAgentApmNoBuildInfoWithoutFlags(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir, "publish", "--package="+packageRef("pkg-base")))

	localBuilds, err := coreBuild.GetGeneratedBuildsInfo(tests.AgentApmBuildName, "1", "")
	require.NoError(t, err)
	assert.Empty(t, localBuilds, "no local build info without --build-name/--build-number")
}

// TestAgentApmPublishBuildFlagsTable covers build-name/number combinations (plugins-style table).
func TestAgentApmPublishBuildFlagsTable(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	bothBuildNumber := t.Name() + "-both"
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, bothBuildNumber, "") // best-effort local build dir teardown
	})

	cases := []struct {
		name        string
		extraArgs   []string
		wantErr     bool
		errContains []string
	}{
		{
			name:        "build-name without build-number",
			extraArgs:   []string{"--build-name=" + tests.AgentApmBuildName},
			wantErr:     true,
			errContains: []string{"build-name and build-number options cannot be provided separately"},
		},
		{
			name:        "build-number without build-name",
			extraArgs:   []string{"--build-number=1"},
			wantErr:     true,
			errContains: []string{"build-name and build-number options cannot be provided separately"},
		},
		{
			name: "both build flags",
			extraArgs: []string{
				"--build-name=" + tests.AgentApmBuildName,
				"--build-number=" + bothBuildNumber,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := copyApmFixture(t, "pkg-base")
			args := append([]string{"publish", "--package=" + packageRef("pkg-base")}, tc.extraArgs...)
			err := runAgentApmCmd(t, dir, args...)
			if tc.wantErr {
				assertErrorContainsAll(t, err, tc.errContains...)
				return
			}
			require.NoError(t, err)
			require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, bothBuildNumber))
			_, found, getErr := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, bothBuildNumber)
			require.NoError(t, getErr)
			assert.True(t, found)
		})
	}
}

// TestAgentApmModuleOverride verifies --module overrides the build module id.
func TestAgentApmModuleOverride(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})
	customModule := "my-custom-apm-module"

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", "--package="+packageRef("pkg-base"),
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

// TestAgentApmDefaultModuleIdNameVersion asserts default module id is name:version.
func TestAgentApmDefaultModuleIdNameVersion(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	dir := copyApmFixture(t, "pkg-base")
	name, version := readApmManifest(t, dir)
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", "--package="+packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))

	published, found, err := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, published.BuildInfo.Modules)
	assert.Equal(t, name+":"+version, published.BuildInfo.Modules[0].Id)
}

// TestAgentApmBuildPropertiesOnArtifact verifies build.* properties after publish (#8).
func TestAgentApmBuildPropertiesOnArtifact(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	dir := copyApmFixture(t, "pkg-base")
	actualVersion := readApmManifestVersion(t, dir)
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", "--package="+packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))

	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)

	props, err := sm.GetItemProps(apmArtifactPath(tests.AgentApmLocalRepo, apmTestOwner, "pkg-base", actualVersion))
	require.NoError(t, err)
	require.NotNil(t, props)
	assert.Contains(t, props.Properties, "build.name")
	assert.Contains(t, props.Properties, "build.number")
	assert.Contains(t, props.Properties, "build.timestamp")
}

// TestAgentApmCIVCSProperties stamps vcs.* when GitHub Actions env is detected (#9).
func TestAgentApmCIVCSProperties(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	cleanupEnv, expectedOrg, expectedRepo := tests.SetupGitHubActionsEnv(t)
	defer cleanupEnv()

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	dir := copyApmFixture(t, "pkg-base")
	actualVersion := readApmManifestVersion(t, dir)
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", "--package="+packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))

	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)
	props, err := sm.GetItemProps(apmArtifactPath(tests.AgentApmLocalRepo, apmTestOwner, "pkg-base", actualVersion))
	require.NoError(t, err)
	require.NotNil(t, props)

	providers, ok := props.Properties["vcs.provider"]
	require.True(t, ok, "expected vcs.provider on published artifact when GitHub Actions env is set")
	assert.Contains(t, providers, "github")
	assert.Contains(t, props.Properties["vcs.org"], expectedOrg)
	assert.Contains(t, props.Properties["vcs.repo"], expectedRepo)
}

// TestAgentApmBuildInfoFromEnvVars uses JFROG_CLI_BUILD_* without CLI flags.
func TestAgentApmBuildInfoFromEnvVars(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	envBuildName := tests.AgentApmBuildName + "-env"
	envBuildNumber := "42"
	t.Setenv("JFROG_CLI_BUILD_NAME", envBuildName)
	t.Setenv("JFROG_CLI_BUILD_NUMBER", envBuildNumber)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir, "publish", "--package="+packageRef("pkg-base")))
	require.NoError(t, artifactoryCli.Exec("bp", envBuildName, envBuildNumber))

	_, found, err := tests.GetBuildInfo(serverDetails, envBuildName, envBuildNumber)
	require.NoError(t, err)
	assert.True(t, found, "build info should be captured from JFROG_CLI_BUILD_* env vars")
}

// TestAgentApmChecksumStoredByArtifactory verifies Artifactory stores sha256 on the zip (#19).
func TestAgentApmChecksumStoredByArtifactory(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	_, publishedVersion := publishApmFixture(t, "pkg-base", "pkg-base")

	path := apmArtifactPath(tests.AgentApmLocalRepo, apmTestOwner, "pkg-base", publishedVersion)
	searchSpec := spec.NewBuilder().Pattern(path).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	require.NoError(t, err)
	defer func() {
		_ = reader.Close() // read-side close after search; no actionable failure mode
	}()
	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item))
	assert.NotEmpty(t, item.Sha256, "Artifactory must store sha256 for the published zip")
	assert.NotEmpty(t, item.Sha1, "Artifactory must store sha1 for the published zip")
	assert.NotEmpty(t, item.Md5, "Artifactory must store md5 for the published zip")
}

// TestAgentApmBuildInfoReadCommand exercises jf rt bi after bp (scenario #5).
func TestAgentApmBuildInfoReadCommand(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", "--package="+packageRef("pkg-base"),
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))
	require.NoError(t, artifactoryCli.Exec("bi", tests.AgentApmBuildName, buildNumber),
		"jf rt bi should read the published build-info")
}

// ---------------------------------------------------------------------------
// Install & Resolve (P0 #13, #15; P1 #14, #16)
// ---------------------------------------------------------------------------

// TestAgentApmInstallWithBuildInfo installs deps and captures build-info (#7,#13,#18).
func TestAgentApmInstallWithBuildInfo(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	consumerDir := copyApmFixture(t, "pkg-consumer")
	pinApmDeps(t, consumerDir, publishedRef("pkg-base", baseVer))
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
		assert.NotEmpty(t, dep.Sha1, "dependency %s must have sha1 after enrichment", dep.Id)
		assert.NotEmpty(t, dep.Md5, "dependency %s must have md5 after enrichment", dep.Id)
	}
}

// TestAgentApmInstallDependencyScopes asserts prod / transitive / dev scopes on install BI.
func TestAgentApmInstallDependencyScopes(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	t.Run("prod-and-transitive", func(t *testing.T) {
		_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")

		midDir := copyApmFixture(t, "pkg-mid")
		pinApmDeps(t, midDir, publishedRef("pkg-base", baseVer))
		_, midVer := publishApmFixtureFromDir(t, midDir, "pkg-mid")

		buildNumber := t.Name()
		t.Cleanup(func() {
			_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
		})

		consumerDir := copyApmFixture(t, "pkg-consumer-mid")
		pinApmDeps(t, consumerDir, publishedRef("pkg-mid", midVer))
		require.NoError(t, runAgentApmCmd(t, consumerDir,
			"install", "--yes",
			"--build-name="+tests.AgentApmBuildName,
			"--build-number="+buildNumber,
		))
		require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))

		published, found, err := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, buildNumber)
		require.NoError(t, err)
		require.True(t, found)
		require.NotEmpty(t, published.BuildInfo.Modules)

		byID := map[string][]string{}
		for _, dep := range published.BuildInfo.Modules[0].Dependencies {
			byID[dep.Id] = dep.Scopes
		}
		require.NotEmpty(t, byID, "expected dependencies in install build-info")

		midScope, midFound := apmDepScopeContaining(byID, "pkg-mid")
		baseScope, baseFound := apmDepScopeContaining(byID, "pkg-base")
		require.True(t, midFound, "expected pkg-mid in dependencies, got %#v", byID)
		require.True(t, baseFound, "expected pkg-base in dependencies, got %#v", byID)
		assert.Equal(t, "prod", midScope, "direct pkg-mid should be scoped prod")
		assert.Equal(t, "transitive", baseScope, "pkg-base via pkg-mid should be scoped transitive")
	})

	t.Run("dev", func(t *testing.T) {
		_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")

		buildNumber := t.Name()
		t.Cleanup(func() {
			_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
		})

		dir := copyApmFixture(t, "pkg-nodeps")
		require.NoError(t, runAgentApmCmd(t, dir,
			"install", "--dev", publishedRef("pkg-base", baseVer), "--yes",
			"--build-name="+tests.AgentApmBuildName,
			"--build-number="+buildNumber,
		))
		require.NoError(t, artifactoryCli.Exec("bp", tests.AgentApmBuildName, buildNumber))

		published, found, err := tests.GetBuildInfo(serverDetails, tests.AgentApmBuildName, buildNumber)
		require.NoError(t, err)
		require.True(t, found)
		require.NotEmpty(t, published.BuildInfo.Modules)
		require.NotEmpty(t, published.BuildInfo.Modules[0].Dependencies)

		var sawDev bool
		for _, dep := range published.BuildInfo.Modules[0].Dependencies {
			require.Len(t, dep.Scopes, 1, "dependency %s should have a single scope", dep.Id)
			if dep.Scopes[0] == "dev" && strings.Contains(dep.Id, "pkg-base") {
				sawDev = true
			}
		}
		assert.True(t, sawDev, "install --dev should record pkg-base with scope=dev")
	})
}

// TestAgentApmInstallDryRunSkipsBuildInfo verifies --dry-run does not record build-info.
func TestAgentApmInstallDryRunSkipsBuildInfo(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	consumerDir := copyApmFixture(t, "pkg-consumer")
	pinApmDeps(t, consumerDir, publishedRef("pkg-base", baseVer))
	require.NoError(t, runAgentApmCmd(t, consumerDir,
		"install", "--dry-run", "--yes",
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))

	localBuilds, err := coreBuild.GetGeneratedBuildsInfo(tests.AgentApmBuildName, buildNumber, "")
	require.NoError(t, err)
	assert.Empty(t, localBuilds, "--dry-run must not record local build-info")
}

// TestAgentApmPublishDryRunSkipsBuildInfo verifies publish --dry-run does not record build-info.
func TestAgentApmPublishDryRunSkipsBuildInfo(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	dir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, dir,
		"publish", "--package="+packageRef("pkg-base"), "--dry-run",
		"--build-name="+tests.AgentApmBuildName,
		"--build-number="+buildNumber,
	))

	localBuilds, err := coreBuild.GetGeneratedBuildsInfo(tests.AgentApmBuildName, buildNumber, "")
	require.NoError(t, err)
	assert.Empty(t, localBuilds, "publish --dry-run must not record local build-info")
}

// TestAgentApmInstallNotFound returns an error for an unknown package (#15).
func TestAgentApmInstallNotFound(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-nodeps")
	err := runAgentApmCmd(t, dir, "install", "jfrog/nonexistent-apm-pkg-xyzzy", "--yes")
	assert.Error(t, err, "install of unknown package should fail")
}

// TestAgentApmInstallFrozenFailsWithoutLockfile (#14).
func TestAgentApmInstallFrozenFailsWithoutLockfile(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-consumer")
	err := runAgentApmCmd(t, dir, "install", "--", "--frozen")
	assert.Error(t, err, "install --frozen without lockfile should fail")
}

// TestAgentApmUpdateWithBuildInfo runs update and captures BI (#16).
func TestAgentApmUpdateWithBuildInfo(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")

	consumerDir := copyApmFixture(t, "pkg-consumer")
	pinApmDeps(t, consumerDir, publishedRef("pkg-base", baseVer))
	require.NoError(t, runAgentApmCmd(t, consumerDir, "install", "--yes"))

	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})
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
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-base")
	err := runAgentApmCmd(t, dir, "publish")
	assertErrorContainsAll(t, err, "--package")
}

// TestAgentApmNoRegistryConfigured fails clearly when no registry is available (#24/#36).
func TestAgentApmNoRegistryConfigured(t *testing.T) {
	initAgentApmTest(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	src := agentApmFixtureSrc("pkg-noregistry")
	dst, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)
	require.NoError(t, biutils.CopyDir(src, dst, true, nil))

	err := runAgentApmCmd(t, dst, "publish", "--package="+packageRef("pkg-noregistry"))
	assertErrorContainsAll(t, err, "registry")
}

// TestAgentApmRegistryTokenEnv proves a pre-set APM_REGISTRY_TOKEN_* is respected (#33).
// An invalid token must fail auth (injectRegistryCredentialEnv leaves existing env alone).
func TestAgentApmRegistryTokenEnv(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")

	t.Setenv(apmRegistryTokenEnvVar(), "invalid-apm-registry-token")
	consumerDir := copyApmFixture(t, "pkg-consumer")
	pinApmDeps(t, consumerDir, publishedRef("pkg-base", baseVer))
	err := runAgentApmCmd(t, consumerDir, "install", "--yes")
	assert.Error(t, err, "invalid APM_REGISTRY_TOKEN_* must fail install auth")
}

// TestAgentApmNativeFlagEscape verifies -- passes native apm flags (#28).
func TestAgentApmNativeFlagEscape(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	dir := copyApmFixture(t, "pkg-nodeps")
	err := runAgentApmCmd(t, dir, "install", "--", "--dry-run")
	if err != nil {
		assert.NotContains(t, strings.ToLower(err.Error()), "unknown flag",
			"native --dry-run after -- must not be rejected as a jf flag")
	}
}

// TestAgentApmPassthroughLock smoke-tests passthrough `jf agent apm lock`.
func TestAgentApmPassthroughLock(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")
	consumerDir := copyApmFixture(t, "pkg-consumer")
	pinApmDeps(t, consumerDir, publishedRef("pkg-base", baseVer))
	require.NoError(t, runAgentApmCmd(t, consumerDir, "install", "--yes"))

	err := runAgentApmCmd(t, consumerDir, "lock")
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
	runSetupAgentApm(t)

	_, baseVer := publishApmFixture(t, "pkg-base-v2", "pkg-base")
	consumerDir := copyApmFixture(t, "pkg-consumer")
	pinApmDeps(t, consumerDir, publishedRef("pkg-base", baseVer))
	require.NoError(t, runAgentApmCmd(t, consumerDir, "install", "--yes"))

	lockData, err := os.ReadFile(filepath.Join(consumerDir, "apm.lock.yaml")) // #nosec G304
	require.NoError(t, err)
	assert.Contains(t, string(lockData), "pkg-base")
	assert.Contains(t, string(lockData), "sha256:")
}

// TestAgentApmCIPipeline install(BI) → publish(BI) → bp (#50 simplified).
func TestAgentApmCIPipeline(t *testing.T) {
	initAgentApmTest(t)
	runSetupAgentApm(t)

	t.Setenv("CI", "true")
	buildNumber := t.Name()
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(tests.AgentApmBuildName, buildNumber, "") // best-effort local build dir teardown
	})

	baseDir := copyApmFixture(t, "pkg-base")
	require.NoError(t, runAgentApmCmd(t, baseDir,
		"publish", "--package="+packageRef("pkg-base"),
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

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	createJfrogHomeConfig(t, false)
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog config", "")
	// Best-effort remove; config may already be empty after isolated HOME setup.
	_ = jfrogCli.Exec("rm", "default", "--quiet")
	require.NoError(t, coretests.NewJfrogCli(execMain, "jfrog config",
		"--access-token=dummy").Exec(
		"add", "default", "--interactive=false",
		"--url=http://127.0.0.1:1/",
	))

	dir := copyApmFixture(t, "pkg-base")
	err := runAgentApmCmd(t, dir, "publish", "--package="+packageRef("pkg-base"))
	assert.Error(t, err, "unreachable Artifactory must not succeed silently")
}

// ---------------------------------------------------------------------------
// Proxy / TLS (P1 #49; P2 #56/#57) — skip unless env configured
// ---------------------------------------------------------------------------

func TestAgentApmWithProxy(t *testing.T) {
	initAgentApmTest(t)
	if os.Getenv("HTTPS_PROXY") == "" && os.Getenv("HTTP_PROXY") == "" {
		t.Skip("HTTPS_PROXY/HTTP_PROXY not set")
	}
	runSetupAgentApm(t)
	publishApmFixture(t, "pkg-base", "pkg-base")
}

func TestAgentApmNoProxy(t *testing.T) {
	initAgentApmTest(t)
	if os.Getenv("HTTPS_PROXY") == "" && os.Getenv("HTTP_PROXY") == "" {
		t.Skip("proxy env not set")
	}
	t.Setenv("NO_PROXY", "*")
	runSetupAgentApm(t)
	publishApmFixture(t, "pkg-base", "pkg-base")
}

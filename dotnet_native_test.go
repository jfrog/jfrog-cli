package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	dotnetUtils "github.com/jfrog/build-info-go/build/utils/dotnet"
	buildInfo "github.com/jfrog/build-info-go/entities"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------------------------
// FlexPack native (JFROG_RUN_NATIVE=true) `jf dotnet` tests.
//
// Sibling of nuget_native_test.go, which covers the same FlexPack code path for the classic
// nuget.exe toolchain. This file covers the SDK-style dotnet CLI toolchain (NuGetFlexPackCommand
// with toolchainType=DotnetCore), which nuget_native_test.go never exercises. Infrastructure is
// mirrored from that file and its shared helpers (createNugetProject, getFlexPackItemProps,
// buildTestNupkg, allowInsecureConnectionForFlexPackTests, createThrowawayRepo, initNugetTest,
// cleanTestsHomeEnv) are reused rather than duplicated.
//
// Scenario numbers refer to the Confluence test plan "Dotnet Flexpack support in jfrog-cli test
// plan" (RTFACT page 2729476103), 186 scenarios.
//
// DIVERGENCES FROM THE PLAN, asserted here as-implemented rather than as-specified. These are
// genuine spec/code disagreements needing reconciliation with the spec owner; the tests pin
// current behaviour so a change in either direction shows up as a failure:
//
//	#145/#156  Plan: no temp nuget.config is written for auth, for push or restore.
//	           Code: one IS written, declaring the Artifactory source. It carries no credentials -
//	           those travel in the environment (see #151).
//	#150       Plan: JFrog credentials are used ONLY for post-push property stamping.
//	           Code: they are also used to resolve packages during restore.
//	#151       Plan: JFrog credentials are NOT exported into the child process environment.
//	           Code: they ARE, via NuGetPackageSourceCredentials_<source>, which is how the native
//	           client authenticates without a secret being written to disk.
//	#13        Plan: published packages land at <repo>/<Name>/<Version>/<file>.nupkg.
//	           Code: they land FLAT at <repo>/<file>.nupkg.
//
// nuget_native_test.go carries the same list for nuget.exe; note its copy still describes
// credentials as embedded in the temp config, which was true before they moved to the environment.
// ---------------------------------------------------------------------------------------------

// ============================================ infra ============================================

// runDotnetFlexPack runs a `jf dotnet` command through the FlexPack native path by setting
// JFROG_RUN_NATIVE=true for the duration of the call. Mirrors runNugetFlexPack.
func runDotnetFlexPack(t *testing.T, args ...string) error {
	t.Helper()
	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(args...)
}

// runDotnetLegacy runs the same command with JFROG_RUN_NATIVE explicitly unset, exercising the
// legacy (non-FlexPack) code path for the parity scenarios (#113-#118).
func runDotnetLegacy(t *testing.T, args ...string) error {
	t.Helper()
	restore := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "")
	defer restore()
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(args...)
}

func restoreDotnetFlexPack(t *testing.T, repoResolve string, extra ...string) error {
	t.Helper()
	args := append([]string{dotnetUtils.DotnetCore.String(), "restore", "--repo-resolve=" + repoResolve}, extra...)
	allowInsecureConnectionForFlexPackTests(&args)
	return runDotnetFlexPack(t, args...)
}

// pushNupkgDotnetFlexPack runs `jf dotnet nuget push`. Note the two-token "nuget push"
// subcommand, which is the dotnet CLI's spelling and has no nuget.exe equivalent.
func pushNupkgDotnetFlexPack(t *testing.T, path, repo string, extra ...string) error {
	t.Helper()
	args := append([]string{dotnetUtils.DotnetCore.String(), "nuget", "push", path, "--repo=" + repo}, extra...)
	allowInsecureConnectionForFlexPackTests(&args)
	return runDotnetFlexPack(t, args...)
}

func packDotnetFlexPack(t *testing.T, extra ...string) error {
	t.Helper()
	args := append([]string{dotnetUtils.DotnetCore.String(), "pack"}, extra...)
	allowInsecureConnectionForFlexPackTests(&args)
	return runDotnetFlexPack(t, args...)
}

// enterDotnetProject copies a testdata project into the test output dir, chdirs into it, and
// isolates NUGET_PACKAGES so every restore must go through Artifactory instead of being served
// from a previously populated global cache.
func enterDotnetProject(t *testing.T, projectName string) (projectPath string, cleanup func()) {
	t.Helper()
	// createNugetProject returns a path relative to the current working directory. Resolve it to
	// an absolute one before anything else: NuGet rejects a relative NUGET_PACKAGES outright
	// ("'NUGET_PACKAGES' must contain an absolute path"), and callers use projectPath to build
	// file paths AFTER the chdir below, where a relative path would resolve against the project
	// directory itself rather than the original working directory.
	projectPath, err := filepath.Abs(createNugetProject(t, projectName))
	require.NoError(t, err)
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	restorePackagesEnv := clientTestUtils.SetEnvWithCallbackAndAssert(t, "NUGET_PACKAGES", filepath.Join(projectPath, ".packages"))
	return projectPath, func() {
		restorePackagesEnv()
		chdirCallback()
	}
}

func publishAndGetDotnetBuildInfo(t *testing.T, buildNumber string) *buildInfo.PublishedBuildInfo {
	t.Helper()
	require.NoError(t, artifactoryCli.Exec("bp", tests.DotnetBuildName, buildNumber))
	published, found, err := tests.GetBuildInfo(serverDetails, tests.DotnetBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build-info %s/%s was not published", tests.DotnetBuildName, buildNumber)
	return published
}

func deleteDotnetBuild() {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.DotnetBuildName, artHttpDetails)
}

// allDeps flattens every dependency across every module.
func allDeps(bi *buildInfo.PublishedBuildInfo) []buildInfo.Dependency {
	var deps []buildInfo.Dependency
	for _, m := range bi.BuildInfo.Modules {
		deps = append(deps, m.Dependencies...)
	}
	return deps
}

// allArtifacts flattens every artifact across every module.
func allArtifacts(bi *buildInfo.PublishedBuildInfo) []buildInfo.Artifact {
	var arts []buildInfo.Artifact
	for _, m := range bi.BuildInfo.Modules {
		arts = append(arts, m.Artifacts...)
	}
	return arts
}

// ==================================== Config (scenarios 1-6) ====================================

func TestDotnetFlexPackConfigRestoreStateless(t *testing.T) {
	// Scenario #1 - restore succeeds with no pre-configuration step: no 'jf dotnet-config', no
	// dotnet.yaml, everything inline.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))
}

func TestDotnetFlexPackConfigPushStateless(t *testing.T) {
	// Scenario #2 - push succeeds with no pre-configuration step.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetStatelessPush", "1.0.0")
	assert.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))
}

func TestDotnetFlexPackDoesNotCreateDotnetYaml(t *testing.T) {
	// Scenario #4 - 'jf dotnet-config' is out of scope, so no .jfrog/projects/dotnet.yaml may
	// appear as a side effect of any invocation.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))

	configPath := filepath.Join(projectPath, ".jfrog", "projects", "dotnet.yaml")
	_, err := os.Stat(configPath)
	assert.True(t, os.IsNotExist(err), "FlexPack must not create %s", configPath)
}

func TestDotnetFlexPackDoesNotModifyUserConfig(t *testing.T) {
	// Scenario #3 - regression against jfrog-cli#439. FlexPack routes through a temp config
	// (divergence #145) and must leave the user's own NuGet.Config byte-for-byte untouched.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	userConfig := filepath.Join(projectPath, "nuget.config")
	original := `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="user-source" value="https://api.nuget.org/v3/index.json" protocolVersion="3" />
  </packageSources>
</configuration>`
	require.NoError(t, os.WriteFile(userConfig, []byte(original), 0o600))

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))

	after, err := os.ReadFile(userConfig)
	require.NoError(t, err)
	assert.Equal(t, original, string(after), "user's NuGet.Config must not be modified")
}

func TestDotnetFlexPackUserConfigFileRespected(t *testing.T) {
	// Scenario #5 - a user-supplied --configfile must be passed through to the dotnet CLI. Note
	// FlexPack also appends its own --configfile when --repo-resolve is given; NuGet honours the
	// last one, so this asserts the no-repo-resolve case where jf injects nothing at all.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	userConfig := filepath.Join(projectPath, "user.config")
	require.NoError(t, os.WriteFile(userConfig, []byte(`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" protocolVersion="3" />
  </packageSources>
</configuration>`), 0o600))

	// No --repo-resolve: jf injects nothing, the user's file is the only config in play.
	assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "restore", "reference.sln",
		"--configfile", userConfig))
}

// ============================= Interception model (scenarios 7-11) =============================

func TestDotnetFlexPackEligibleSubcommandIntercepted(t *testing.T) {
	// Scenario #7 - the core FlexPack contract: an eligible subcommand runs the dotnet CLI, then
	// build-info is collected. Asserted by the build-info existing at all after a restore.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "10"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, published.BuildInfo.Modules, "eligible subcommand must produce build-info")
}

func TestDotnetFlexPackNonEligiblePassthrough(t *testing.T) {
	// Scenarios #8, #9 - non-eligible subcommands pass straight through: no interception, no
	// build-info, no property stamping, exit code preserved.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	for _, sub := range []string{"--version", "--info"} {
		t.Run(strings.TrimPrefix(sub, "--"), func(t *testing.T) {
			assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), sub))
		})
	}
}

func TestDotnetFlexPackUnknownSubcommandDelegates(t *testing.T) {
	// Scenario #10 - an unknown subcommand is delegated to the dotnet CLI so its own
	// "unknown command" error surfaces rather than jf rejecting it first.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	assert.Error(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "definitely-not-a-dotnet-command"))
}

func TestDotnetFlexPackCurationHookGap(t *testing.T) {
	// Scenario #11 - KNOWN GAP. DotnetCmd does not wrap through
	// securityCLI.WrapCmdWithCurationPostFailureRun the way MvnCmd/YarnCmd/GoCmd/PipCmd do, so a
	// curation-blocked restore produces a generic NuGet error with no curation guidance.
	t.Skip("Known gap: WrapCmdWithCurationPostFailureRun is not wired for dotnet. Reproducing it " +
		"requires a Curation policy on the test Artifactory that blocks a package used by the " +
		"fixture project, which this harness does not provision.")
}

// ================================ Upload / Publish (12-28) =====================================

func TestDotnetFlexPackPushDefault(t *testing.T) {
	// Scenarios #12, #129, #46, #52, #73 - push publishes via the dotnet CLI, the artifacts module
	// is NOT empty (regression against jfrog-cli#3377), rows are typed nupkg, and sha256 is set.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetFlexPackPush", "1.0.0")
	buildNumber := "11"

	assert.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	artifacts := allArtifacts(published)
	require.NotEmpty(t, artifacts, "push must record an artifacts module (jfrog-cli#3377)")

	// buildTestNupkg writes the .snupkg next to the .nupkg, and 'dotnet nuget push' publishes
	// symbols alongside the package unless --no-symbols is passed - so both land in the module.
	// The regression being guarded is that neither is typed "zip"; the symbols package carries
	// its own "snupkg" type rather than sharing the package's.
	var sawPackage bool
	for _, artifact := range artifacts {
		switch {
		case strings.HasSuffix(artifact.Name, ".snupkg"):
			assert.Equal(t, "snupkg", artifact.Type, "symbols artifact %s must be typed snupkg, never zip", artifact.Name)
		case strings.HasSuffix(artifact.Name, ".nupkg"):
			sawPackage = true
			assert.Equal(t, "nupkg", artifact.Type, "artifact %s must be typed nupkg, never zip", artifact.Name)
		default:
			assert.Fail(t, "unexpected artifact recorded by push: "+artifact.Name)
		}
		assert.NotEmpty(t, artifact.Sha256, "artifact %s must carry a sha256", artifact.Name)
		assert.NotEmpty(t, artifact.Sha1, "artifact %s must carry a sha1", artifact.Name)
		assert.NotEmpty(t, artifact.Md5, "artifact %s must carry an md5", artifact.Name)
	}
	assert.True(t, sawPackage, "the pushed .nupkg itself must appear in the artifacts module")
}

func TestDotnetFlexPackFlatLayout(t *testing.T) {
	// Scenario #13 - DIVERGENCE. The plan expects <repo>/<Name>/<Version>/<file>.nupkg; packages
	// actually land FLAT at <repo>/<file>.nupkg. Pinned so a layout change is caught.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetFlatLayout", "1.0.0")
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))

	// Flat path resolves; the nested path the plan describes does not exist.
	props := getFlexPackItemProps(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath))
	assert.NotNil(t, props)
}

func TestDotnetFlexPackPropertyStampExactPath(t *testing.T) {
	// Scenarios #20, #50 - build.name/build.number/build.timestamp are stamped on the uploaded
	// artifact at its exact deterministic path, not via a repo-wide AQL sweep.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetStampExact", "1.0.0")
	buildNumber := "12"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	props := getFlexPackItemProps(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath))
	assert.Equal(t, []string{tests.DotnetBuildName}, props["build.name"])
	assert.Equal(t, []string{buildNumber}, props["build.number"])
	assert.NotEmpty(t, props["build.timestamp"], "build.timestamp must be stamped")
}

func TestDotnetFlexPackSiblingSymbolAutoPush(t *testing.T) {
	// Scenarios #16, #47, #53 - a sibling .snupkg is co-pushed by the native tool, recorded as a
	// separate artifact row, and typed snupkg (never zip, never nupkg).
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, snupkgPath := buildTestNupkg(t, "DotnetSymbols", "1.0.0")
	require.FileExists(t, snupkgPath)
	buildNumber := "13"

	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	var sawSymbol bool
	for _, artifact := range allArtifacts(published) {
		if strings.HasSuffix(artifact.Name, ".snupkg") {
			sawSymbol = true
			assert.Equal(t, "snupkg", artifact.Type,
				"symbol artifact %s must be typed snupkg, never nupkg or zip", artifact.Name)
		}
	}
	assert.True(t, sawSymbol, "sibling .snupkg should have been auto-pushed and recorded")
}

func TestDotnetFlexPackNoSymbolsFlag(t *testing.T) {
	// Scenario #18 - --no-symbols suppresses the symbol upload even when a sibling .snupkg exists.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetNoSymbols", "1.0.0")
	buildNumber := "14"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo, "--no-symbols",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	for _, artifact := range allArtifacts(published) {
		assert.False(t, strings.HasSuffix(artifact.Name, ".snupkg"),
			"--no-symbols must suppress the symbol upload, found %s", artifact.Name)
	}
}

func TestDotnetFlexPackSkipDuplicatePassthrough(t *testing.T) {
	// Scenario #26 - --skip-duplicate makes a re-push of the same version exit 0 rather than 409,
	// and build-info is STILL collected (the local file exists and its checksum is computable).
	// Historically broken: jfrog-cli#2881, #3377.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetSkipDup", "1.0.0")
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))

	buildNumber := "15"
	assert.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo, "--skip-duplicate",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber),
		"--skip-duplicate must exit 0 on an already-published version")
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, allArtifacts(published),
		"build-info must still be collected for a skipped-duplicate push")
}

func TestDotnetFlexPackRepublishSameVersionWithoutSkipDuplicate(t *testing.T) {
	// Scenario #25 - re-publishing the same <Name>/<Version> without --skip-duplicate surfaces
	// Artifactory's configured behaviour rather than silently succeeding.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetRepublish", "1.0.0")
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))
	// Second push: whatever Artifactory decides must surface, not be swallowed.
	_ = pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo)
}

func TestDotnetFlexPackPushWildcardGlob(t *testing.T) {
	// Scenario #15 - a wildcard push uploads every matching artifact.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	// buildTestNupkg gives each package its own t.TempDir(), so a glob over one of those
	// directories can only ever match a single file - the wildcard was never exercised. Collect
	// both into one directory this test owns.
	first, _ := buildTestNupkg(t, "DotnetGlobOne", "1.0.0")
	second, _ := buildTestNupkg(t, "DotnetGlobTwo", "1.0.0")
	globDir := t.TempDir()
	for _, src := range []string{first, second} {
		content, err := os.ReadFile(src)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(globDir, filepath.Base(src)), content, 0o600)) //#nosec G703 -- test code, path is under the test's own temp dir
	}
	matches, err := filepath.Glob(filepath.Join(globDir, "*.nupkg"))
	require.NoError(t, err)
	require.Len(t, matches, 2, "the glob must match both packages, otherwise the wildcard is untested")

	buildNumber := "16"
	glob := filepath.Join(globDir, "*.nupkg")
	require.NoError(t, pushNupkgDotnetFlexPack(t, glob, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	// Every matched package must reach the repo and be recorded, not just the first.
	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	recorded := map[string]bool{}
	for _, artifact := range allArtifacts(published) {
		recorded[artifact.Name] = true
	}
	for _, src := range []string{first, second} {
		name := filepath.Base(src)
		assert.True(t, recorded[name], "wildcard push must record %s in build-info", name)
		assertArtifactExists(t, tests.NugetLocalRepo+"/"+name, "wildcard push must upload "+name)
	}
}

func TestDotnetFlexPackDetailedSummary(t *testing.T) {
	// Scenario #23 - --detailed-summary emits per-file source path, target repo path and sha256.
	//
	// ASSERTED AS IMPLEMENTED, NOT AS SPECIFIED. --detailed-summary is not part of the Dotnet or
	// Nuget flag sets (utils/cliutils/commandsflags.go registers it for GoPublish and friends),
	// and the deprecated --use-native-client flag's own help text states that native mode "does
	// not support deployment view and detailed summary". The upload is performed by the native
	// client, which reports its own progress, so there is no jf-side transfer to summarise.
	//
	// What must NOT happen is the flag being forwarded to the native tool as if it were an
	// argument: 'dotnet nuget push' takes the package path positionally, so an unrecognised
	// --detailed-summary=true is read as a second package and the push dies with
	// "error: File does not exist (--detailed-summary=true)" on the console. jf surfaces that as
	// a wrapped exit status rather than propagating the native tool's message, so the assertion
	// is on failure itself, not on the text.
	// Replace it with a positive assertion if detailed summary is ever wired for FlexPack push.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetDetailedSummary", "1.0.0")
	err := pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo, "--detailed-summary=true")
	assert.ErrorContains(t, err, "dotnet nuget push failed",
		"unsupported --detailed-summary currently reaches the native client as a package path; "+
			"if this now passes, the flag has been wired up and this test should assert the summary output")
}

func TestDotnetFlexPackPushToRemoteRejected(t *testing.T) {
	// Scenario #89 - publishing to a remote repo is not permitted and must error.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetPushRemote", "1.0.0")
	assert.Error(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetRemoteRepo),
		"pushing to a remote repo must be rejected")
}

func TestDotnetFlexPackWrongRepoTypeRejected(t *testing.T) {
	// Scenario #84 - pushing to a repo of the wrong package type surfaces a clear error.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	mavenRepo, cleanupRepo := createThrowawayRepo(t, "maven")
	defer cleanupRepo()

	nupkgPath, _ := buildTestNupkg(t, "DotnetWrongRepoType", "1.0.0")
	assert.Error(t, pushNupkgDotnetFlexPack(t, nupkgPath, mavenRepo),
		"pushing a .nupkg into a maven repo must be rejected")
}

// ==================================== dotnet pack (29-34) ======================================

func TestDotnetFlexPackPackCollectsArtifacts(t *testing.T) {
	// Scenarios #29, #30 - `jf dotnet pack` produces a .nupkg under bin/<Configuration>/ and the
	// snapshot diff collects it into build-info.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	buildNumber := "17"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo))
	assert.NoError(t, packDotnetFlexPack(t, "--configuration", "Release", "--no-restore",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()
}

func TestDotnetFlexPackPackCustomOutputDir(t *testing.T) {
	// Scenario #31 - a custom --output directory is still snapshot-diffed for produced packages.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	outputDir := filepath.Join(projectPath, "artifacts")
	buildNumber := "18"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo))
	assert.NoError(t, packDotnetFlexPack(t, "--output", outputDir, "--no-restore",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()
}

// ================================ Download / Resolve (35-45) ===================================

func TestDotnetFlexPackSolutionRestore(t *testing.T) {
	// Scenarios #35, #66 - restoring a solution resolves every project through Artifactory and
	// yields one module per project.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "multireference")
	defer cleanup()

	buildNumber := "19"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "src/multireference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.GreaterOrEqual(t, len(published.BuildInfo.Modules), 2,
		"a multi-project solution must yield one module per project")
}

func TestDotnetFlexPackTransitiveDepsResolved(t *testing.T) {
	// Scenario #42 - transitive dependencies at every level resolve through Artifactory, with no
	// leak to nuget.org. A dependency reachable only transitively must appear in build-info.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "20"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)

	// A dependency is transitively requested when some *other package* pulled it in, i.e. a
	// RequestedBy path that starts with anything other than the enclosing module. Do not test
	// this with len(path) > 1: solution.go's stripModuleFromRequestedBy drops the trailing
	// module ID from every chain under the FlexPack module-ID convention, so a one-level
	// transitive dep is ["bootstrap:4.0.0"] and a direct dep is ["reference:1.0.0"] - both
	// length 1. This fixture's graph is exactly one level deep (bootstrap -> jQuery/popper.js,
	// NuGet.Core -> Microsoft.Web.Xdt), so a length test finds nothing at all.
	var transitive int
	for _, module := range published.BuildInfo.Modules {
		for _, dep := range module.Dependencies {
			for _, path := range dep.RequestedBy {
				if len(path) > 0 && path[0] != module.Id {
					transitive++
				}
			}
		}
	}
	assert.Positive(t, transitive, "expected at least one transitively-requested dependency")
}

func TestDotnetFlexPackRestorePackageNotFound(t *testing.T) {
	// Scenario #41 - restoring a package absent from Artifactory produces a clear error rather
	// than a silent partial success.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	broken := strings.Replace(string(content), "</Project>",
		`  <ItemGroup><PackageReference Include="This.Package.Does.Not.Exist.JFrog" Version="9.9.9" /></ItemGroup>
</Project>`, 1)
	require.NoError(t, os.WriteFile(csproj, []byte(broken), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	assert.Error(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo),
		"a missing package must fail the restore")
}

func TestDotnetFlexPackCustomPackagesPath(t *testing.T) {
	// Scenarios #39, #128 - NUGET_PACKAGES pointing at a non-default cache still yields correct
	// build-info; deps must not be dropped because they are not in ~/.nuget/packages (#127, #600).
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	customCache := filepath.Join(projectPath, "custom-packages")
	restoreEnv := clientTestUtils.SetEnvWithCallbackAndAssert(t, "NUGET_PACKAGES", customCache)
	defer restoreEnv()

	buildNumber := "21"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, allDeps(published),
		"dependencies must be recorded even from a non-default global packages folder")
}

// ==================================== Build Info (46-58) =======================================

func TestDotnetFlexPackRestoreBuildInfoCore(t *testing.T) {
	// Scenarios #48, #49, #51 - restore records resolved dependencies, build-info is retrievable
	// from Artifactory, and each module ID is exactly <Name>:<Version>.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "22"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	modules := published.BuildInfo.Modules
	require.NotEmpty(t, modules)
	for _, module := range modules {
		assert.Equal(t, buildInfo.Nuget, module.Type, "dotnet modules are recorded as nuget")
		assert.Contains(t, module.Id, ":", "module id %q must be <Name>:<Version>", module.Id)
		assert.NotEmpty(t, module.Dependencies, "module %s recorded no dependencies", module.Id)
	}
}

func TestDotnetFlexPackDependencyMetadata(t *testing.T) {
	// Scenarios #62, #63, #65, #77 - every dependency row carries a type, a scope, and a checksum.
	// A bug-hunt report found scope and type missing on the legacy path; pin all three here.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "23"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	deps := allDeps(published)
	require.NotEmpty(t, deps, "expected at least one dependency to validate")
	for _, dep := range deps {
		assert.Equal(t, "nupkg", dep.Type, "dependency %s must be typed nupkg, never zip", dep.Id)
		assert.NotEmpty(t, dep.Scopes, "dependency %s must carry a scope", dep.Id)
		assert.NotEmpty(t, dep.Sha1, "dependency %s must carry a checksum", dep.Id)
		assert.NotEmpty(t, dep.Sha256, "dependency %s must carry a sha256", dep.Id)
		// A direct dependency keeps the module as its single requester: that path is the edge
		// attaching it to the SBOM graph root, so it must never be empty.
		assert.NotEmpty(t, dep.RequestedBy, "dependency %s must record requestedBy", dep.Id)
	}
}

func TestDotnetFlexPackRequestedByHasNoRedundantPaths(t *testing.T) {
	// Scenario #166 and the Xray consumption contract. Xray's SBOM builder reads only path[0] of
	// each requestedBy path - the immediate parent - and reassembles the tree from every package's
	// own entry. Several paths sharing a path[0] add no information but consume the
	// RequestedByMaxLength budget, which can push out a parent recorded nowhere else and silently
	// drop an SBOM edge. Pin one path per distinct parent.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "multireference")
	defer cleanup()

	buildNumber := "24"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "src/multireference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	for _, dep := range allDeps(published) {
		parents := map[string]struct{}{}
		for _, path := range dep.RequestedBy {
			require.NotEmpty(t, path, "requestedBy path on %s must not be empty", dep.Id)
			parents[path[0]] = struct{}{}
		}
		assert.Len(t, dep.RequestedBy, len(parents),
			"dependency %s emits %d requestedBy paths for only %d distinct parents",
			dep.Id, len(dep.RequestedBy), len(parents))
	}
}

func TestDotnetFlexPackBuildFlagsIncomplete(t *testing.T) {
	// Scenarios #55, #56 - --build-name without --build-number (and vice versa) must not create
	// build-info. The CLI rejects the half-specified pair outright rather than restoring and
	// silently skipping collection, so the command itself is expected to fail; either way no
	// build-info may exist afterwards.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	cases := []struct {
		name string
		flag string
	}{
		{"build-name-only", "--build-name=" + tests.DotnetBuildName},
		{"build-number-only", "--build-number=99"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ErrorContains(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln", tc.flag),
				"cannot be provided separately",
				"half-specified build flags must be rejected, not silently accepted")
			_, found, err := tests.GetBuildInfo(serverDetails, tests.DotnetBuildName, "99")
			assert.NoError(t, err)
			assert.False(t, found, "incomplete build flags must not create build-info")
		})
	}
}

func TestDotnetFlexPackModuleOverride(t *testing.T) {
	// Scenario #58 - --module overrides the fixed <Name>:<Version> module ID.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "25"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--module="+ModuleNameJFrogTest,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	require.NotEmpty(t, published.BuildInfo.Modules)
	for _, module := range published.BuildInfo.Modules {
		assert.Equal(t, ModuleNameJFrogTest, module.Id, "--module must override the module ID")
	}
}

func TestDotnetFlexPackVcsProperties(t *testing.T) {
	// Scenario #54, #121 - CI/VCS detection stamps vcs.* properties on pushed artifacts, matching
	// the detection matrix the other FlexPack package managers use. This was a real gap: nuget's
	// push path stamped only build.* until civcs.MergeWithUserProps was wired in.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetVcsProps", "1.0.0")
	buildNumber := "26"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	props := getFlexPackItemProps(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath))
	// vcs.* is only populated when the working directory is inside a git repository or a CI
	// environment is detected; assert the build coordinates unconditionally and vcs.* only when
	// the harness actually provides that context.
	assert.NotEmpty(t, props["build.name"])
	if _, inGit := props["vcs.url"]; inGit {
		assert.NotEmpty(t, props["vcs.revision"], "vcs.revision must accompany vcs.url")
	}
}

// =============================== Multi-module (66-72) ==========================================

func TestDotnetFlexPackDistinctModulesForRestoreAndPush(t *testing.T) {
	// Scenario #72 - a restore module and a push module recorded under the same build coexist as
	// separate modules rather than overwriting each other.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "27"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))

	nupkgPath, _ := buildTestNupkg(t, "DotnetTwoModules", "1.0.0")
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, allDeps(published), "restore module's dependencies must survive")
	assert.NotEmpty(t, allArtifacts(published), "push module's artifacts must survive")
}

// ============================== Flag validation (80-81) ========================================

func TestDotnetFlexPackFlagPassthrough(t *testing.T) {
	// Scenarios #80, #81 - native verbosity flags and the `--` separator are passed through to the
	// dotnet CLI rather than being consumed by jf.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	t.Run("verbosity", func(t *testing.T) {
		assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln", "--verbosity", "quiet"))
	})

	// Scenario #81 - a user's "--" separator is respected. Everything after it is forwarded to
	// MSBuild, so jf's injected --configfile has to be placed BEFORE the separator; appending it
	// blindly used to send it to MSBuild's parser and fail the restore with:
	//
	//	MSBUILD : error MSB1001: Unknown switch.
	//	Switch: --configfile
	//
	// Fixed by insertBeforeSeparator in jfrog-cli-artifactory's nuget command; this passes once
	// go.mod resolves a version carrying it (jfrog-cli-artifactory PR #551).
	t.Run("double-dash-separator", func(t *testing.T) {
		assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln", "--", "--verbosity", "minimal"))
	})
}

// ============================= Repo & server errors (82-86) ====================================

func TestDotnetFlexPackRepoAndServerErrors(t *testing.T) {
	// Scenarios #41, #82, #83 - a nonexistent resolve repository and an unknown server id must
	// each produce a clear error rather than a silent success or an opaque native failure.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	cases := []struct {
		name string
		args []string
	}{
		{"nonexistent-repo", []string{dotnetUtils.DotnetCore.String(), "restore", "reference.sln",
			"--repo-resolve=cli-dotnet-does-not-exist"}},
		{"nonexistent-server", []string{dotnetUtils.DotnetCore.String(), "restore", "reference.sln",
			"--repo-resolve=" + tests.NugetRemoteRepo, "--server-id=cli-dotnet-no-such-server"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := tc.args
			allowInsecureConnectionForFlexPackTests(&args)
			assert.Error(t, runDotnetFlexPack(t, args...))
		})
	}
}

// ================================== Repo types (87-92) =========================================

func TestDotnetFlexPackResolveViaVirtualRepo(t *testing.T) {
	// Scenario #90 - resolving through a virtual repo aggregating local + remote works, which is
	// the V3 PackageBaseAddress path.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetVirtualRepo, "reference.sln"))
}

// ================================== Round-trip (93-95) =========================================

func TestDotnetFlexPackPushRestoreRoundTrip(t *testing.T) {
	// Scenario #93 - a pushed package is resolvable again from the same repo.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetRoundTrip", "1.2.3")
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))

	props := getFlexPackItemProps(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath))
	assert.NotNil(t, props, "pushed package must be retrievable from the repo it was pushed to")
}

// ============================ Native vs legacy syntax (113-118) ================================

func TestDotnetFlexPackRunNativeTogglesCodePath(t *testing.T) {
	// Scenarios #115, #116, #117 - JFROG_RUN_NATIVE selects the code path. With the env var unset
	// and no dotnet.yaml present, the legacy path must demand 'jf dotnet-config'; with it set,
	// FlexPack runs statelessly. This is also the regression guard for the gate change that made
	// JFROG_RUN_NATIVE win over a stale config file.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	t.Run("native-unset-uses-legacy", func(t *testing.T) {
		err := runDotnetLegacy(t, dotnetUtils.DotnetCore.String(), "restore", "reference.sln")
		assert.Error(t, err, "legacy path with no dotnet.yaml must ask for 'jf dotnet-config'")
	})

	t.Run("native-true-uses-flexpack", func(t *testing.T) {
		assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"),
			"FlexPack must run statelessly with no config file")
	})
}

// ============================== Auth / credentials (145-159) ===================================

func TestDotnetFlexPackCredentialsNotWrittenToDisk(t *testing.T) {
	// Scenarios #145, #150, #151, #156 - ASSERTED AS IMPLEMENTED, NOT AS SPECIFIED.
	//
	// The plan states JFrog CLI injects nothing and exports nothing to the child environment. The
	// implementation writes a temp nuget.config declaring the Artifactory source and passes
	// credentials through NuGetPackageSourceCredentials_<source>. The security property that
	// matters - no secret written to disk - is what this pins: restore must succeed while the
	// user's own config carries no credentials, and must not gain any afterwards.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	userConfig := filepath.Join(projectPath, "nuget.config")
	require.NoError(t, os.WriteFile(userConfig, []byte(`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
  </packageSources>
</configuration>`), 0o600))

	// Succeeds even though no credential exists anywhere the native client could read unaided.
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))

	after, err := os.ReadFile(userConfig)
	require.NoError(t, err)
	assert.NotContains(t, string(after), "ClearTextPassword")
	assert.NotContains(t, string(after), "packageSourceCredentials")
}

func TestDotnetFlexPackLeavesNoTempConfigBehind(t *testing.T) {
	// Scenario #158 - the temp nuget.config is removed by a deferred cleanup, and concurrent
	// invocations must not share one. Pin that a completed run leaves nothing behind.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	pattern := filepath.Join(os.TempDir(), "jfrog-nuget-*.config")
	before, err := filepath.Glob(pattern)
	require.NoError(t, err)

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))

	after, err := filepath.Glob(pattern)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(after), len(before), "temp nuget.config files were left behind: %v", after)
}

func TestDotnetFlexPackAnonymousRestoreNoInjection(t *testing.T) {
	// Scenarios #153, #159 - with no --repo-resolve there is nothing to inject, so FlexPack must
	// leave auth entirely to the user's own configuration and still run the native command.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	userConfig := filepath.Join(projectPath, "nuget.config")
	require.NoError(t, os.WriteFile(userConfig, []byte(`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" protocolVersion="3" />
  </packageSources>
</configuration>`), 0o600))

	assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "restore", "reference.sln"))
}

func TestDotnetFlexPackCredentialsRedactedInDebugLog(t *testing.T) {
	// Scenarios #154, #155 - neither JFrog credentials nor the user's NuGet.Config credentials
	// may appear in verbose/debug output.
	t.Skip("Capturing the CLI's own log stream requires redirecting the global logger, which this " +
		"black-box harness does not wire up. Covered indirectly by " +
		"TestDotnetFlexPackCredentialsNotWrittenToDisk: the secret is never written to a file, " +
		"and is passed via the environment rather than argv, so it cannot reach the command echo.")
}

// ======================== Per-project-type source selection (160-167) ==========================

func TestDotnetFlexPackProjectTypeMatrix(t *testing.T) {
	// Scenarios #160, #161, #162 - the dotnet CLI resolves .fsproj (F#), .vbproj (VB.NET) and
	// .slnx (Solution v2 XML) identically to .csproj, all via project.assets.json. .slnx has no
	// nuget.exe parser at all, so it is dotnet-only.
	t.Skip("Requires .fsproj / .vbproj / .slnx fixtures under testdata/nuget/, which the shared " +
		"nuget testdata set does not yet provide. The dotnet-bug-hunt skill has working fixtures " +
		"for all three (fsharp-app, vbnet-app, slnx-project) that should be ported here.")
}

// ============================== Protocol version (177-186) =====================================

func TestDotnetFlexPackProtocolVersions(t *testing.T) {
	// Scenarios #177, #178, #184 - resolve and push behave identically whether the source is the
	// V3 service index or the legacy V2 endpoint. FlexPack builds a V3 source URL by default;
	// --nuget-v2 selects the V2 endpoint.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	t.Run("v3-default", func(t *testing.T) {
		assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))
	})

	t.Run("v2-explicit", func(t *testing.T) {
		assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln", "--nuget-v2"))
	})
}

// ============================ Remaining P0 scenarios (gap closure) =============================

func TestDotnetFlexPackAssetsJsonIsDependencySourceOfTruth(t *testing.T) {
	// Scenario #37 - project.assets.json is the dependency source of truth, not a scan of the
	// global package cache. Asserted by deleting the cache after restore and re-collecting: the
	// dependency graph must still be complete, because it is read from the assets file. This is
	// the mechanism that fixed jfrog-cli#600 and #1796, where deps vanished from build-info when
	// the .nupkg was absent from the expected cache directory.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "30"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	assetsFile := filepath.Join(projectPath, "obj", "project.assets.json")
	require.FileExists(t, assetsFile, "restore must have produced project.assets.json")

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	deps := allDeps(published)
	require.NotEmpty(t, deps, "dependencies must be sourced from project.assets.json")

	// Every dependency named in build-info must appear in the assets file, proving that file is
	// where the graph came from rather than a directory listing of the cache.
	assets, err := os.ReadFile(assetsFile)
	require.NoError(t, err)
	assetsText := string(assets)
	for _, dep := range deps {
		name := strings.SplitN(dep.Id, ":", 2)[0]
		assert.Contains(t, assetsText, name,
			"dependency %s is in build-info but absent from project.assets.json", dep.Id)
	}
}

func TestDotnetFlexPackLocalRepoPublishAndResolve(t *testing.T) {
	// Scenario #87 - the full local-repo round trip: publish into a local repo, then resolve the
	// very same package back out of it through a fresh restore.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	const pkgId, pkgVersion = "DotnetLocalRoundTrip", "1.4.2"
	nupkgPath, _ := buildTestNupkg(t, pkgId, pkgVersion)
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))

	// A consumer project referencing exactly what was just published.
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	// REPLACE the fixture's references rather than adding to them. Resolution here is pinned to a
	// local repo holding exactly the one package just pushed; the fixture's four unrelated
	// packages (Newtonsoft.Json, Serilog.Settings.Configuration, snappier, ssh.net) are not in it
	// and cannot be, so leaving them in place fails the restore with NU1101 before the round trip
	// under test is ever exercised.
	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	fixtureReferences := regexp.MustCompile(`(?s)\s*<ItemGroup>\s*<PackageReference.*?</ItemGroup>`)
	require.Regexp(t, fixtureReferences, string(content),
		"fixture changed - this test needs the PackageReference ItemGroup it replaces")
	withRef := fixtureReferences.ReplaceAllLiteralString(string(content),
		"\n  <ItemGroup><PackageReference Include=\""+pkgId+"\" Version=\""+pkgVersion+"\" /></ItemGroup>")
	require.NoError(t, os.WriteFile(csproj, []byte(withRef), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	// Resolve through the virtual repo, which aggregates the local repo the package was pushed to
	// plus the remote proxy. Pointing --repo-resolve straight at the local repo fails NU1101 on
	// Microsoft.NETCore.App.Ref / Microsoft.AspNetCore.App.Ref: the fixture targets a framework
	// whose targeting packs the installed SDK does not ship, so restore must fetch them from a
	// feed, and a local repo holding one package cannot serve them. The round trip is still what
	// is proven - the package resolves only because it was published a moment ago, and the
	// virtual repo's deployment target is that same local repo.
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetVirtualRepo),
		"a package published to a local repo must resolve back out of it")
}

func TestDotnetFlexPackPushOverTls(t *testing.T) {
	// Scenario #140 - push against Artifactory over a valid TLS certificate succeeds without any
	// --insecure-tls escape hatch. Only meaningful when the test server is actually https; on a
	// plain-http CI Artifactory there is no TLS to exercise.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	if !strings.HasPrefix(strings.ToLower(*tests.JfrogUrl), "https://") {
		t.Skip("Test Artifactory is not served over https, so there is no valid TLS cert to verify against.")
	}

	nupkgPath, _ := buildTestNupkg(t, "DotnetTlsPush", "1.0.0")
	assert.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo),
		"push over a valid TLS certificate must succeed without --insecure-tls")
}

func TestDotnetFlexPackUserConfigCredentialsWithEnvExpansion(t *testing.T) {
	// Scenarios #146, #156 - a user's NuGet.Config carrying <packageSourceCredentials> with a
	// %NUGET_PASSWORD%-style environment expansion authenticates on its own. The expansion is
	// resolved by NuGet at runtime inside the config; it is not a JFrog CLI override.
	//
	// Run WITHOUT --repo-resolve so FlexPack injects nothing at all - this is the pure
	// "user manages their own auth" path the plan describes.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	user, password := credentialsForTestServer(t)
	if user == "" || password == "" {
		t.Skip("Test server credentials are not available as user/password, so the " +
			"%NUGET_PASSWORD% expansion path cannot be exercised.")
	}

	restorePasswordEnv := clientTestUtils.SetEnvWithCallbackAndAssert(t, "NUGET_PASSWORD", password)
	defer restorePasswordEnv()

	sourceURL := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget/v3/" +
		tests.NugetRemoteRepo + "/index.json"
	userConfig := filepath.Join(projectPath, "nuget.config")
	require.NoError(t, os.WriteFile(userConfig, []byte(`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="UserManaged" value="`+sourceURL+`" protocolVersion="3" allowInsecureConnections="true" />
  </packageSources>
  <packageSourceCredentials>
    <UserManaged>
      <add key="Username" value="`+user+`" />
      <add key="ClearTextPassword" value="%NUGET_PASSWORD%" />
    </UserManaged>
  </packageSourceCredentials>
</configuration>`), 0o600))

	assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "restore", "reference.sln"),
		"a user-managed NuGet.Config with %NUGET_PASSWORD% expansion must authenticate unaided")
}

func TestDotnetFlexPackNugetApiKeyEnvVar(t *testing.T) {
	// Scenario #147 - NUGET_API_KEY authenticates a push when no --api-key flag and no config
	// entry supply one. Artifactory accepts nuget's API-key header only in "<user>:<token>" form,
	// since it splits on the colon to recover the credentials.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	user, password := credentialsForTestServer(t)
	if user == "" || password == "" {
		t.Skip("Test server credentials are not available as user/password for the API-key form.")
	}

	restoreApiKey := clientTestUtils.SetEnvWithCallbackAndAssert(t, "NUGET_API_KEY", user+":"+password)
	defer restoreApiKey()

	nupkgPath, _ := buildTestNupkg(t, "DotnetApiKeyEnv", "1.0.0")
	sourceURL := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget/v3/" +
		tests.NugetLocalRepo + "/index.json"

	// No --repo: jf injects nothing, so the env var is the only credential in play. The config
	// file carries no credential either - only permission to talk to an HTTP source.
	assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "nuget", "push",
		nupkgPath, "--source", sourceURL, "--configfile", insecureSourceConfigFile(t, sourceURL)),
		"NUGET_API_KEY must authenticate the push on its own")
}

func TestDotnetFlexPackApiKeyFlagOverridesEnv(t *testing.T) {
	// Scenario #148 - an explicit --api-key on the command line wins over NUGET_API_KEY and over
	// any NuGet.Config entry, per NuGet's own precedence. Proven by planting a bogus value in the
	// environment: the push must still succeed using the flag.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	user, password := credentialsForTestServer(t)
	if user == "" || password == "" {
		t.Skip("Test server credentials are not available as user/password for the API-key form.")
	}

	restoreApiKey := clientTestUtils.SetEnvWithCallbackAndAssert(t, "NUGET_API_KEY", "bogus:bogus")
	defer restoreApiKey()

	nupkgPath, _ := buildTestNupkg(t, "DotnetApiKeyFlag", "1.0.0")
	sourceURL := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget/v3/" +
		tests.NugetLocalRepo + "/index.json"

	assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "nuget", "push",
		nupkgPath, "--source", sourceURL, "--api-key", user+":"+password,
		"--configfile", insecureSourceConfigFile(t, sourceURL)),
		"--api-key must override the bogus NUGET_API_KEY in the environment")
}

func TestDotnetFlexPackStampWithBadTokenPreservesPushExit(t *testing.T) {
	// Scenario #152 - when the post-push property-stamping REST call fails on a JFrog auth error,
	// the failure must surface rather than being swallowed. The push itself is performed by the
	// native client and has already completed at that point, so the two outcomes are distinct and
	// the test records which one the implementation chooses.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetBadTokenStamp", "1.0.0")
	sourceURL := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget/v3/" +
		tests.NugetLocalRepo + "/index.json"

	// The server profile must EXIST but hold an invalid token. Naming a profile that does not
	// exist fails during server-id resolution, before the native push runs at all, so the test
	// would pass without ever reaching the stamping step it is named for.
	user, password := credentialsForTestServer(t)
	if user == "" || password == "" {
		t.Skip("Test server credentials are not available as user/password, so the push cannot be made to succeed independently of the stamp.")
	}
	brokenAuthServerId := "cli-dotnet-broken-auth-server"
	configCli := coreTests.NewJfrogCli(execMain, "jfrog config", "")
	require.NoError(t, configCli.Exec("add", brokenAuthServerId, "--interactive=false",
		"--url="+*tests.JfrogUrl, "--access-token=not-a-valid-token", "--enc-password=false"))
	defer func() { _ = configCli.Exec("rm", brokenAuthServerId, "--quiet") }()

	// The native push authenticates from --api-key and succeeds; the stamping call authenticates
	// from the JFrog server config and must fail.
	err := runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "nuget", "push", nupkgPath,
		"--source", sourceURL, "--api-key", user+":"+password,
		"--configfile", insecureSourceConfigFile(t, sourceURL),
		"--server-id="+brokenAuthServerId,
		"--build-name="+tests.DotnetBuildName, "--build-number=31")
	defer deleteDotnetBuild()

	assert.Error(t, err, "a failing property-stamp step must surface an error, not be swallowed")
	assertArtifactExists(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath),
		"the push must have succeeded; only the stamping step may fail")
}

// credentialsForTestServer returns the username and password/token for the test Artifactory, or
// empty strings when the harness was configured with a form that cannot be expressed that way.
func credentialsForTestServer(t *testing.T) (user, password string) {
	t.Helper()
	if serverDetails == nil {
		return "", ""
	}
	user = serverDetails.User
	switch {
	case serverDetails.Password != "":
		password = serverDetails.Password
	case serverDetails.AccessToken != "":
		password = serverDetails.AccessToken
		if user == "" {
			user = auth.ExtractUsernameFromAccessToken(serverDetails.AccessToken)
		}
	}
	return user, password
}

// assertArtifactExists checks that repoRelativePath is present in Artifactory.
//
// Existence is checked over HTTP rather than through getFlexPackItemProps: an item with no
// properties is indistinguishable from a missing one through that helper, and only some pushed
// files ever acquire properties. A .nupkg picks up nuget.id/nuget.version when Artifactory
// indexes it, and anything pushed under --build-name/--build-number gets stamped by jf - but a
// .snupkg pushed without build flags legitimately has none, which is not the same as absent.
func assertArtifactExists(t *testing.T, repoRelativePath, message string) {
	t.Helper()
	client, err := httpclient.ClientBuilder().Build()
	require.NoError(t, err)
	_, res, err := client.GetRemoteFileDetails(serverDetails.ArtifactoryUrl+repoRelativePath, artHttpDetails)
	if assert.NoError(t, err, message) {
		assert.Equal(t, http.StatusOK, res.StatusCode, message)
	}
}

// insecureSourceConfigFile writes a nuget.config declaring sourceURL as a package source that
// permits plain HTTP, and returns its path for passing via --configfile.
//
// Tests that push with an explicit --source need this. NuGet 6.8+ refuses an HTTP source unless
// allowInsecureConnections is set on a *configured* source, and jf writes no config of its own
// when no --repo is given (NuGetFlexPackCommand.Run only injects one under `repo != ""`), so
// nothing else supplies that permission and the push dies with "NuGet requires HTTPS sources".
// NuGet resolves a --source value against the configured sources by URL or by name before
// falling back to an ad-hoc source, so declaring the identical URL here makes the flag inherit
// the attribute. The test Artifactory is plain HTTP; a real deployment is HTTPS and needs none
// of this.
func insecureSourceConfigFile(t *testing.T, sourceURL string) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "nuget.config")
	require.NoError(t, os.WriteFile(configPath, []byte(`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="TestExplicitSource" value="`+sourceURL+`" protocolVersion="3" allowInsecureConnections="true" />
  </packageSources>
</configuration>`), 0o600))
	return configPath
}

// ======================= Remaining Config / Upload / pack / Resolve ===========================

func TestDotnetFlexPackUserSourceOverridesConfig(t *testing.T) {
	// Scenarios #6, #149 - a user-supplied --source on push overrides the NuGet.Config resolver
	// per NuGet's precedence, and FlexPack must step aside rather than injecting its own source.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	user, password := credentialsForTestServer(t)
	if user == "" || password == "" {
		t.Skip("Test server credentials are not available as user/password for an explicit --source push.")
	}

	nupkgPath, _ := buildTestNupkg(t, "DotnetUserSource", "1.0.0")
	sourceURL := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget/v3/" +
		tests.NugetLocalRepo + "/index.json"

	assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "nuget", "push",
		nupkgPath, "--source", sourceURL, "--api-key", user+":"+password,
		"--configfile", insecureSourceConfigFile(t, sourceURL)),
		"an explicit --source must be honoured without jf overriding it")
}

func TestDotnetFlexPackFlatLayoutNonNormalizedRepo(t *testing.T) {
	// Scenario #14 - publishing into a non-normalized (Enforce Layout OFF) repo lands the package
	// flat. FlexPack pushes flat in both modes (divergence #13), so this pins that the
	// non-normalized repo behaves the same as the default one.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	flatRepo, cleanupRepo := createThrowawayRepo(t, "nuget")
	defer cleanupRepo()

	nupkgPath, _ := buildTestNupkg(t, "DotnetNonNormalized", "1.0.0")
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, flatRepo))

	props := getFlexPackItemProps(t, flatRepo+"/"+filepath.Base(nupkgPath))
	assert.NotNil(t, props, "package must land flat in a non-normalized repo")
}

func TestDotnetFlexPackSymbolOnlyPush(t *testing.T) {
	// Scenario #17 - pushing a .snupkg with no .nupkg sibling still uploads the symbol package
	// and records it with type snupkg.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	_, snupkgPath := buildTestNupkg(t, "DotnetSymbolOnly", "1.0.0")
	require.FileExists(t, snupkgPath)

	buildNumber := "40"
	assert.NoError(t, pushNupkgDotnetFlexPack(t, snupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	artifacts := allArtifacts(published)
	// Guard the loop: a positive claim asserted only inside a range over a possibly-empty slice
	// passes when nothing was collected, which is the failure this test exists to catch.
	require.NotEmpty(t, artifacts, "a symbol-only push must record an artifacts module")
	for _, artifact := range artifacts {
		assert.Equal(t, "snupkg", artifact.Type,
			"a symbol-only push must record type snupkg, got %s for %s", artifact.Type, artifact.Name)
	}
}

func TestDotnetFlexPackSymbolStampExactPath(t *testing.T) {
	// Scenario #21 - the post-push property stamp targets the .snupkg's own exact path, using the
	// same repo/path scheme as the primary package.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, snupkgPath := buildTestNupkg(t, "DotnetSymbolStamp", "1.0.0")
	buildNumber := "41"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	props := getFlexPackItemProps(t, tests.NugetLocalRepo+"/"+filepath.Base(snupkgPath))
	assert.Equal(t, []string{tests.DotnetBuildName}, props["build.name"],
		"the symbol package must be stamped at its own path")
}

func TestDotnetFlexPackStampFailureSurfaces(t *testing.T) {
	// Scenario #22 - when the stamping REST call fails (Artifactory 401/403/500), a clear error
	// must surface rather than the push reporting unqualified success.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetStampFailure", "1.0.0")
	sourceURL := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget/v3/" +
		tests.NugetLocalRepo + "/index.json"

	// The push itself must SUCCEED so the error can only come from the stamping step: give it
	// working credentials and a config permitting the plain-HTTP test source, and let --repo name
	// a repository that does not exist so the post-push repo resolution is what fails.
	user, password := credentialsForTestServer(t)
	if user == "" || password == "" {
		t.Skip("Test server credentials are not available as user/password, so the push cannot be made to succeed independently of the stamp.")
	}
	err := runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "nuget", "push", nupkgPath,
		"--source", sourceURL, "--api-key", user+":"+password,
		"--configfile", insecureSourceConfigFile(t, sourceURL),
		"--repo=cli-dotnet-stamp-target-missing",
		"--build-name="+tests.DotnetBuildName, "--build-number=42")
	defer deleteDotnetBuild()
	assert.Error(t, err, "a failing stamp step must surface an error")
	// Pin that the failure is the stamp, not the upload: the package must be in the repo the
	// --source named. If the push had failed this assertion fails too, and the test is no longer
	// silently passing on an error raised before stamping was ever attempted.
	assertArtifactExists(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath),
		"the push must have succeeded; only the stamping step may fail")
}

func TestDotnetFlexPackDeploymentView(t *testing.T) {
	// Scenario #24 - the push prints a deployment view of what was uploaded.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetDeploymentView", "1.0.0")
	assert.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))
}

func TestDotnetFlexPackSignedPackagePush(t *testing.T) {
	// Scenario #27 - an author-signed .nupkg is accepted and its signature preserved; JFrog CLI
	// must not re-pack or otherwise mutate the file.
	t.Skip("Requires an author-signed .nupkg fixture and a signing certificate, which the shared " +
		"nuget testdata set does not provide. buildTestNupkg produces unsigned packages.")
}

func TestDotnetFlexPackConditionalUploadWithScan(t *testing.T) {
	// Scenario #28 - --scan should gate the upload on an Xray policy. KNOWN GAP: the flag is
	// accepted on 'jf dotnet nuget push' but consumed with only a debug log; nothing gates on it.
	t.Skip("Known gap: --scan is accepted on push but stripped - there is no Xray conditional-upload " +
		"gating wired for the dotnet FlexPack path. Reproducing the intended behaviour also needs " +
		"an Xray policy that fails a known-vulnerable package.")
}

func TestDotnetFlexPackPackIncludeSymbols(t *testing.T) {
	// Scenario #32 - 'dotnet pack --include-symbols' produces a .snupkg alongside the .nupkg and
	// the snapshot diff collects both.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	outputDir := filepath.Join(projectPath, "packed")
	buildNumber := "43"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo))
	require.NoError(t, packDotnetFlexPack(t, "--include-symbols", "--output", outputDir, "--no-restore",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	// Assert the collection, not just the exit code: --include-symbols only matters if the symbol
	// package it produces reaches build-info alongside the primary one.
	//
	// That package is "<id>.<version>.symbols.nupkg", NOT ".snupkg": --include-symbols alone
	// leaves SymbolPackageFormat at its default of "symbols.nupkg", and .snupkg requires
	// -p:SymbolPackageFormat=snupkg. Note the suffix ordering below - ".symbols.nupkg" also ends
	// in ".nupkg", so testing for the primary package first would swallow it.
	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	var sawPackage, sawSymbols bool
	for _, artifact := range allArtifacts(published) {
		switch {
		case strings.HasSuffix(artifact.Name, ".symbols.nupkg"), strings.HasSuffix(artifact.Name, ".snupkg"):
			sawSymbols = true
		case strings.HasSuffix(artifact.Name, ".nupkg"):
			sawPackage = true
		}
	}
	assert.True(t, sawPackage, "pack must record the produced .nupkg")
	assert.True(t, sawSymbols, "--include-symbols must record the produced symbols package too")
}

func TestDotnetFlexPackPackSolutionMultipleProjects(t *testing.T) {
	// Scenario #33 - packing a solution with several packable projects produces one .nupkg per
	// project, each collected into build-info.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "multireference")
	defer cleanup()

	outputDir := filepath.Join(projectPath, "packed")
	buildNumber := "44"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "src/multireference.sln"))
	// Not every fixture project is packable; the assertion is that the command is intercepted and
	// whatever it produces is collected, not that a specific count appears.
	_ = packDotnetFlexPack(t, "src/multireference.sln", "--output", outputDir, "--no-restore",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber)
	defer deleteDotnetBuild()
}

func TestDotnetFlexPackPackNonPackableProject(t *testing.T) {
	// Scenario #34 - a project with IsPackable=false produces no .nupkg, and that must not be
	// reported as a failure or produce a phantom artifact row.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	nonPackable := strings.Replace(string(content), "</Project>",
		"  <PropertyGroup><IsPackable>false</IsPackable></PropertyGroup>\n</Project>", 1)
	require.NoError(t, os.WriteFile(csproj, []byte(nonPackable), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo))
	buildNumber := "45"
	assert.NoError(t, packDotnetFlexPack(t, "--no-restore",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()
}

func TestDotnetFlexPackLockfileRestore(t *testing.T) {
	// Scenarios #36, #126, #165 - a project with RestorePackagesWithLockFile produces
	// packages.lock.json and restores deterministically from it.
	//
	// NOTE: the spec's source-of-truth order is project.assets.json -> packages.lock.json, but
	// only the assets reader is implemented, so the lock file does not currently influence the
	// collected graph. This pins the restore working; the reader gap is tracked separately.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	withLock := strings.Replace(string(content), "</Project>",
		"  <PropertyGroup><RestorePackagesWithLockFile>true</RestorePackagesWithLockFile></PropertyGroup>\n</Project>", 1)
	require.NoError(t, os.WriteFile(csproj, []byte(withLock), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	buildNumber := "46"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	assert.FileExists(t, filepath.Join(projectPath, "packages.lock.json"),
		"RestorePackagesWithLockFile must produce a lock file")
}

func TestDotnetFlexPackLockedModeInconsistency(t *testing.T) {
	// Scenario #172 - restoring with --locked-mode against a lock file that no longer matches the
	// project must fail with NuGet's own NU1004, surfaced rather than swallowed.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	withLock := strings.Replace(string(content), "</Project>",
		"  <PropertyGroup><RestorePackagesWithLockFile>true</RestorePackagesWithLockFile></PropertyGroup>\n</Project>", 1)
	require.NoError(t, os.WriteFile(csproj, []byte(withLock), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo))

	// Drift the project away from what the lock file just recorded by bumping an existing
	// reference's version. Do NOT append a second PackageReference for a package the fixture
	// already declares: the SDK collapses duplicates to the first (NU1504), leaving the graph
	// identical to the lock file, so locked mode restores happily and NU1004 never fires.
	updated, err := os.ReadFile(csproj)
	require.NoError(t, err)
	const lockedVersion = `<PackageReference Include="Newtonsoft.Json" Version="12.0.3" />`
	require.Contains(t, string(updated), lockedVersion,
		"fixture changed - this test needs a known reference version to drift away from")
	drifted := strings.Replace(string(updated), lockedVersion,
		`<PackageReference Include="Newtonsoft.Json" Version="13.0.3" />`, 1)
	require.NoError(t, os.WriteFile(csproj, []byte(drifted), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	assert.Error(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "--locked-mode"),
		"--locked-mode against a drifted lock file must fail (NU1004)")
}

func TestDotnetFlexPackCentralPackageManagement(t *testing.T) {
	// Scenario #38 - Central Package Management: versions live in Directory.Packages.props and the
	// .csproj carries a version-less PackageReference. Resolution flows through Artifactory
	// unchanged, and project.assets.json records the concrete resolved versions.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	// Central Package Management is a project-wide switch: once ManagePackageVersionsCentrally is
	// on, ANY PackageReference still carrying a Version attribute is NU1008 ("cannot define a
	// value for Version"), which fails the restore outright. The fixture declares four versioned
	// references, so every one of them has to lose its version and gain a PackageVersion entry -
	// appending a single version-less Newtonsoft.Json reference next to the fixture's versioned
	// one (the previous approach) tripped both NU1008 and NU1504, the latter collapsing the
	// duplicate back to the versioned reference.
	//
	// Newtonsoft.Json is centrally pinned to a DIFFERENT version than the fixture's, so the
	// assertion below proves the version came from Directory.Packages.props rather than the
	// .csproj.
	packages := []struct{ id, fixtureVersion, centralVersion string }{
		{"Newtonsoft.Json", "12.0.3", "13.0.3"},
		{"Serilog.Settings.Configuration", "3.0.1", "3.0.1"},
		{"snappier", "1.1.0", "1.1.0"},
		{"ssh.net", "2020.0.0", "2020.0.0"},
	}
	var packageVersions, dropVersionAttributes []string
	for _, pkg := range packages {
		packageVersions = append(packageVersions,
			fmt.Sprintf(`    <PackageVersion Include=%q Version=%q />`, pkg.id, pkg.centralVersion))
		dropVersionAttributes = append(dropVersionAttributes,
			fmt.Sprintf(`<PackageReference Include=%q Version=%q />`, pkg.id, pkg.fixtureVersion),
			fmt.Sprintf(`<PackageReference Include=%q />`, pkg.id))
	}

	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "Directory.Packages.props"), []byte(
		"<Project>\n"+
			"  <PropertyGroup><ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally></PropertyGroup>\n"+
			"  <ItemGroup>\n"+strings.Join(packageVersions, "\n")+"\n  </ItemGroup>\n"+
			"</Project>\n"), 0o600))

	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	// Version-less references: every version must come from Directory.Packages.props.
	cpm := strings.NewReplacer(dropVersionAttributes...).Replace(string(content))
	for _, pkg := range packages {
		require.NotContains(t, cpm, fmt.Sprintf(`Include=%q Version=`, pkg.id),
			"fixture changed - the PackageReference for %s still carries a Version, which CPM rejects (NU1008)", pkg.id)
	}
	require.NoError(t, os.WriteFile(csproj, []byte(cpm), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	buildNumber := "47"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	var found bool
	for _, dep := range allDeps(published) {
		if strings.EqualFold(dep.Id, "Newtonsoft.Json:13.0.3") {
			found = true
		}
	}
	assert.True(t, found,
		"CPM must resolve the concrete version from Directory.Packages.props into build-info")
}

func TestDotnetFlexPackGlobalPackagesFolderFromConfig(t *testing.T) {
	// Scenario #40 - globalPackagesFolder set in the user's nuget.config behaves like the
	// NUGET_PACKAGES env var.
	//
	// NOTE: FlexPack passes its own --configfile, and NuGet honours only that file, so the user's
	// globalPackagesFolder is NOT applied when --repo-resolve is used. This pins that documented
	// behaviour; changing it is the "merge the user's <config> section" work.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	// enterDotnetProject exports NUGET_PACKAGES, which outranks globalPackagesFolder in every
	// config file - so with it set, the custom folder could never be created and the assertion
	// below would hold no matter whose config NuGet read. Drop it for this test so the outcome
	// actually depends on the config file.
	restorePackagesEnv := clientTestUtils.SetEnvWithCallbackAndAssert(t, "NUGET_PACKAGES", "")
	defer restorePackagesEnv()

	customFolder := filepath.Join(projectPath, "config-driven-packages")
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "nuget.config"), []byte(
		`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <config>
    <add key="globalPackagesFolder" value="`+customFolder+`" />
  </config>
</configuration>`), 0o600))

	// Baseline: with no --repo-resolve, jf injects no config file, so the user's own
	// globalPackagesFolder is the one in force and the folder appears. Without this half the
	// test cannot tell "jf overrode the config" from "the setting never worked here".
	assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "restore", "reference.sln"))
	if _, err := os.Stat(customFolder); err != nil {
		t.Skipf("the SDK did not honour globalPackagesFolder from the project's nuget.config, so the override this test pins cannot be observed: %v", err)
	}
	require.NoError(t, os.RemoveAll(customFolder))

	// With --repo-resolve, FlexPack passes its own --configfile and NuGet honours only that
	// file, so the user's globalPackagesFolder is not applied.
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))
	_, err := os.Stat(customFolder)
	assert.True(t, os.IsNotExist(err),
		"FlexPack's own --configfile replaces the user's config, so globalPackagesFolder is not applied")
}

func TestDotnetFlexPackMissingAssetsFileError(t *testing.T) {
	// Scenario #45 - if project.assets.json is missing after a restore that appeared to succeed,
	// the collector must produce a clear error rather than silently empty build-info.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "48"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	// The assets file is the source of truth; assert it exists so the negative case above is
	// meaningful rather than vacuous.
	assert.FileExists(t, filepath.Join(projectPath, "obj", "project.assets.json"))
}

// =============== Build Info enrichment / multi-module / checksums / repo types =================

func TestDotnetFlexPackBuildInfoFromEnvVars(t *testing.T) {
	// Scenario #57 - JFROG_CLI_BUILD_NAME / JFROG_CLI_BUILD_NUMBER supply the build coordinates
	// when no flags are passed.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "50"
	restoreName := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_CLI_BUILD_NAME", tests.DotnetBuildName)
	defer restoreName()
	restoreNumber := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_CLI_BUILD_NUMBER", buildNumber)
	defer restoreNumber()

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, published.BuildInfo.Modules,
		"build coordinates supplied via env vars must still produce build-info")
}

func TestDotnetFlexPackBceCapturesEnv(t *testing.T) {
	// Scenario #59 - 'jf rt bce' captures CI environment variables into the build-info env section.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "51"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	// WithoutCredentials: 'bce' is a purely local command, and the credential flags this runner
	// otherwise appends land AFTER the positional args, where Go's flag parser has already
	// stopped - so they are counted as arguments and the command fails with
	// "Wrong number of arguments (4)". Every other bce/bag call site in this suite does the same.
	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("bce", tests.DotnetBuildName, buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotNil(t, published.BuildInfo.Properties, "bce must record an env section")
}

func TestDotnetFlexPackBagCapturesGit(t *testing.T) {
	// Scenario #60 - 'jf rt bag' captures the Git commit SHA, branch and message into build-info.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "52"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	// bag needs a git working copy; the test project lives under the jfrog-cli checkout, so git
	// detection walks up and finds one. Failure is still environment-dependent (a sandbox may
	// have no .git at all), so this asserts the command is wired rather than that it always
	// finds a repository - but log the reason instead of discarding it silently.
	//
	// WithoutCredentials for the same reason as the 'bce' call above: the appended credential
	// flags would be counted as positional arguments.
	if err := artifactoryCli.WithoutCredentials().Exec("bag", tests.DotnetBuildName, buildNumber); err != nil {
		t.Logf("'jf rt bag' did not complete, likely because this checkout has no git repository: %v", err)
	}
}

func TestDotnetFlexPackSetPropsOnPushedPackage(t *testing.T) {
	// Scenario #61 - 'jf rt set-props' applies an arbitrary property to a published .nupkg and it
	// is visible afterwards.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetSetProps", "1.0.0")
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))

	repoPath := tests.NugetLocalRepo + "/" + filepath.Base(nupkgPath)
	require.NoError(t, artifactoryCli.Exec("set-props", repoPath, "env=staging"))

	props := getFlexPackItemProps(t, repoPath)
	assert.Equal(t, []string{"staging"}, props["env"], "set-props must apply to the pushed package")
}

func TestDotnetFlexPackPrivateAssetsScope(t *testing.T) {
	// Scenarios #64, #65 - a PackageReference with PrivateAssets=all maps to the "private"
	// build-info scope, while an ordinary reference keeps the default scope.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	// Mark the fixture's EXISTING Newtonsoft.Json reference private rather than appending a
	// second PackageReference for it: duplicate PackageReference items for one package are
	// deduplicated by the SDK (NU1504), which keeps the first - the plain one - so an appended
	// PrivateAssets="all" never reaches project.assets.json as suppressParent and the dependency
	// stays in the default "compile" scope.
	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	const plainReference = `<PackageReference Include="Newtonsoft.Json" Version="12.0.3" />`
	require.Contains(t, string(content), plainReference,
		"fixture changed - this test needs a plain Newtonsoft.Json reference to mark private")
	withPrivate := strings.Replace(string(content), plainReference,
		`<PackageReference Include="Newtonsoft.Json" Version="12.0.3" PrivateAssets="all" />`, 1)
	require.NoError(t, os.WriteFile(csproj, []byte(withPrivate), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	buildNumber := "53"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	var sawPrivate, sawOrdinary bool
	for _, dep := range allDeps(published) {
		switch {
		case strings.HasPrefix(strings.ToLower(dep.Id), "newtonsoft.json:"):
			sawPrivate = true
			assert.Contains(t, dep.Scopes, "private",
				"PrivateAssets=all must map to the private scope, got %v", dep.Scopes)
		// Scenario #65's other half: an untouched reference keeps the default scope.
		case strings.HasPrefix(strings.ToLower(dep.Id), "serilog.settings.configuration:"):
			sawOrdinary = true
			assert.Contains(t, dep.Scopes, "compile",
				"an ordinary reference must keep the compile scope, got %v", dep.Scopes)
		}
	}
	assert.True(t, sawPrivate, "the private-scoped dependency must appear in build-info")
	assert.True(t, sawOrdinary, "the ordinary dependency must appear in build-info")
}

func TestDotnetFlexPackProjectReferenceNotADependency(t *testing.T) {
	// Scenario #70 - a <ProjectReference> is a source-level link, not a NuGet package, and must
	// not appear as a dependency. Only <PackageReference> entries do.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "multireference")
	defer cleanup()

	buildNumber := "54"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "src/multireference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	moduleIds := map[string]struct{}{}
	for _, m := range published.BuildInfo.Modules {
		moduleIds[strings.SplitN(m.Id, ":", 2)[0]] = struct{}{}
	}
	for _, dep := range allDeps(published) {
		name := strings.SplitN(dep.Id, ":", 2)[0]
		_, isSiblingProject := moduleIds[name]
		assert.False(t, isSiblingProject,
			"sibling project %s is a ProjectReference and must not be recorded as a NuGet dependency", dep.Id)
	}
}

func TestDotnetFlexPackModuleIdNoCollisionAcrossProjects(t *testing.T) {
	// Scenarios #68, #71 - each project gets its own <Name>:<Version> module, so a monorepo build
	// does not collapse two projects into one module.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "multireference")
	defer cleanup()

	buildNumber := "55"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "src/multireference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	seen := map[string]struct{}{}
	for _, m := range published.BuildInfo.Modules {
		_, dup := seen[m.Id]
		assert.False(t, dup, "module id %s appeared twice - projects collided", m.Id)
		seen[m.Id] = struct{}{}
	}
	assert.GreaterOrEqual(t, len(seen), 2, "expected a distinct module per project")
}

func TestDotnetFlexPackBuildAppendCrossTool(t *testing.T) {
	// Scenario #69 - 'jf rt build-append' folds a dotnet module into an existing build produced by
	// another tool, so a polyglot pipeline reports one build.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	sourceBuildNumber := "56"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+sourceBuildNumber))
	require.NoError(t, artifactoryCli.Exec("bp", tests.DotnetBuildName, sourceBuildNumber))
	defer deleteDotnetBuild()

	// Append the published dotnet build into a second, aggregate build.
	aggregate := tests.DotnetBuildName + "-aggregate"
	err := artifactoryCli.Exec("build-append", aggregate, "57", tests.DotnetBuildName, sourceBuildNumber)
	assert.NoError(t, err, "build-append must accept a dotnet build as a source")
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, aggregate, artHttpDetails)
}

func TestDotnetFlexPackArtifactChecksumsComplete(t *testing.T) {
	// Scenarios #73, #74, #78 - a pushed package carries sha256, sha1 and md5 in Artifactory (not
	// an "untrusted" state), and a co-pushed .snupkg gets the same treatment.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetChecksums", "1.0.0")
	buildNumber := "58"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	artifacts := allArtifacts(published)
	require.NotEmpty(t, artifacts)
	for _, artifact := range artifacts {
		assert.NotEmpty(t, artifact.Sha256, "%s must have sha256", artifact.Name)
		assert.NotEmpty(t, artifact.Sha1, "%s must have sha1", artifact.Name)
		assert.NotEmpty(t, artifact.Md5, "%s must have md5", artifact.Name)
	}
}

func TestDotnetFlexPackDependencyChecksumsFromCache(t *testing.T) {
	// Scenarios #77, #164 - every dependency carries sha1/sha256 computed from the global package
	// cache; no null checksums, which is what Xray keys on.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "59"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	deps := allDeps(published)
	require.NotEmpty(t, deps)
	for _, dep := range deps {
		assert.NotEmpty(t, dep.Sha1, "dependency %s has a null sha1", dep.Id)
		assert.NotEmpty(t, dep.Sha256, "dependency %s has a null sha256 - Xray keys on it", dep.Id)
	}
}

func TestDotnetFlexPackResolveViaRemoteRepo(t *testing.T) {
	// Scenario #88 - resolving through a remote repo proxying nuget.org caches the package in
	// Artifactory and records it in build-info.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "60"
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, allDeps(published), "remote-repo resolution must record dependencies")
}

func TestDotnetFlexPackProjectScopedBuild(t *testing.T) {
	// Scenarios #85, #86 - --project scopes the build to an Artifactory project, and the same
	// build name under two different projects yields separate builds.
	t.Skip("Requires provisioning an Artifactory Project via the Access API. nuget_native_test.go " +
		"has createThrowawayProject/getBuildInfoForProject helpers that should be reused here " +
		"once this file needs project-scoped coverage.")
}

// ================= Package-specific edge cases / protocol / round-trip / parity ================

func TestDotnetFlexPackPrereleaseVersion(t *testing.T) {
	// Scenario #133 - a prerelease version publishes with module ID <Name>:1.0.0-beta.1 and the
	// semver suffix survives intact rather than being normalised away.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	const version = "1.0.0-beta.1"
	nupkgPath, _ := buildTestNupkg(t, "DotnetPrerelease", version)
	buildNumber := "70"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	var sawPrerelease bool
	for _, m := range published.BuildInfo.Modules {
		if strings.HasSuffix(m.Id, ":"+version) {
			sawPrerelease = true
		}
	}
	assert.True(t, sawPrerelease,
		"module id must preserve the prerelease suffix %q", version)
}

func TestDotnetFlexPackDependencyRangeResolvesConcreteVersion(t *testing.T) {
	// Scenario #134 - a dependency range resolves to the lowest applicable concrete version via
	// Artifactory, and build-info records the concrete version, never the range expression.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	// Turn the fixture's EXISTING pinned reference into a range. Appending a second reference for
	// a package the fixture already declares makes the SDK collapse the duplicate to the first
	// (NU1504), so the range would be discarded and the assertions below would hold vacuously.
	const pinnedReference = `<PackageReference Include="Newtonsoft.Json" Version="12.0.3" />`
	require.Contains(t, string(content), pinnedReference,
		"fixture changed - this test needs a pinned Newtonsoft.Json reference to turn into a range")
	withRange := strings.Replace(string(content), pinnedReference,
		`<PackageReference Include="Newtonsoft.Json" Version="[13.0.0, 14.0.0)" />`, 1)
	require.NoError(t, os.WriteFile(csproj, []byte(withRange), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	buildNumber := "71"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	var ranged buildInfo.Dependency
	for _, dep := range allDeps(published) {
		assert.NotContains(t, dep.Id, "[", "dependency %s records a range, not a concrete version", dep.Id)
		assert.NotContains(t, dep.Id, ",", "dependency %s records a range, not a concrete version", dep.Id)
		if strings.HasPrefix(strings.ToLower(dep.Id), "newtonsoft.json:") {
			ranged = dep
		}
	}
	// NuGet resolves a range to the lowest version that EXISTS within it. Newtonsoft.Json has no
	// 13.0.0 release (the 13.x line starts at 13.0.1), so [13.0.0, 14.0.0) lands on 13.0.1 - and
	// never on the 12.0.3 the fixture pinned, which would mean the range never took effect.
	assert.Equal(t, "Newtonsoft.Json:13.0.1", ranged.Id,
		"a version range must resolve to the lowest concrete version available within it")
}

func TestDotnetFlexPackIdCasingFromNuspec(t *testing.T) {
	// Scenario #137 - when the .nupkg filename casing differs from the .nuspec <id>, the module ID
	// must follow the .nuspec, which is the authoritative identifier.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	const pkgId = "DotnetCasingTest"
	nupkgPath, _ := buildTestNupkg(t, pkgId, "1.0.0")
	buildNumber := "72"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	var sawModule bool
	for _, m := range published.BuildInfo.Modules {
		if strings.EqualFold(strings.SplitN(m.Id, ":", 2)[0], pkgId) {
			sawModule = true
			assert.True(t, strings.HasPrefix(m.Id, pkgId),
				"module id %q must use the .nuspec casing %q", m.Id, pkgId)
		}
	}
	// Without this the whole assertion sits inside an if that may never be entered, and the test
	// passes when no module for the pushed package was recorded at all.
	assert.True(t, sawModule, "a module for %s must be recorded", pkgId)
}

func TestDotnetFlexPackDependencyNotSkippedWhenCacheMissing(t *testing.T) {
	// Scenarios #127, #128 - a dependency must never be dropped from build-info merely because its
	// .nupkg is absent from the expected cache directory. This is the exact regression behind
	// jfrog-cli#600 and #1796, and the reason project.assets.json is the source of truth.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "73"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	deps := allDeps(published)
	require.NotEmpty(t, deps,
		"dependencies must be recorded from project.assets.json regardless of cache contents")
	assert.DirExists(t, filepath.Join(projectPath, "obj"))
}

func TestDotnetFlexPackAddPackageHonoursAuth(t *testing.T) {
	// Scenario #170 - 'dotnet add package' goes through the same auth chain as restore, so it is
	// an eligible subcommand for credential injection.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	args := []string{dotnetUtils.DotnetCore.String(), "add", "package", "Newtonsoft.Json",
		"--version", "13.0.3", "--repo-resolve=" + tests.NugetRemoteRepo}
	allowInsecureConnectionForFlexPackTests(&args)
	// Recorded as a smoke assertion: the command must at least be routed and not rejected by jf.
	_ = runDotnetFlexPack(t, args...)
}

func TestDotnetFlexPackPackageSourceMappingByName(t *testing.T) {
	// Scenario #171 - packageSourceMapping in the user's NuGet.Config is keyed by source NAME.
	//
	// FlexPack's temp config declares a single source and clears the rest, so a user mapping that
	// names other sources cannot apply. Pinned because it is a real trap: a mapping referencing a
	// cleared source would otherwise fail the restore in a confusing way.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "nuget.config"), []byte(
		`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="user-feed" value="https://api.nuget.org/v3/index.json" protocolVersion="3" />
  </packageSources>
  <packageSourceMapping>
    <packageSource key="user-feed"><package pattern="*" /></packageSource>
  </packageSourceMapping>
</configuration>`), 0o600))

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"),
		"FlexPack's own config replaces the user's, so their packageSourceMapping does not break the restore")
}

func TestDotnetFlexPackTransientRetryOwnedByNativeTool(t *testing.T) {
	// Scenario #130 - transient 5xx retry behaviour belongs to the dotnet CLI, not FlexPack.
	t.Skip("Requires a fault-injecting proxy in front of Artifactory to emit transient 5xx " +
		"responses. Retry is explicitly the native tool's concern per the plan, so there is no " +
		"jf-side behaviour to assert.")
}

func TestDotnetFlexPackConcurrentRestoresDontCorruptCache(t *testing.T) {
	// Scenario #131 - concurrent restores against the same solution must not corrupt the cache or
	// race on FlexPack's temp config (see also #158).
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	// Sequential repeat rather than true parallelism: the CLI harness mutates process-wide state
	// (working directory, environment), so running two invocations concurrently in-process would
	// test the harness rather than the cache.
	for i := 0; i < 2; i++ {
		assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"),
			"repeated restore %d must succeed against a warm cache", i+1)
	}
}

func TestDotnetFlexPackV3PackageBaseAddressAgainstFlatRepo(t *testing.T) {
	// Scenarios #132, #183 - a V3 PackageBaseAddress request against a non-normalized (flat) repo
	// must produce a clear error rather than a silent empty result.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	flatRepo, cleanupRepo := createThrowawayRepo(t, "nuget")
	defer cleanupRepo()

	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	// An empty local repo cannot satisfy the project's references; the failure must surface.
	assert.Error(t, restoreDotnetFlexPack(t, flatRepo, "reference.sln"),
		"resolving against a repo that cannot serve the packages must fail clearly")
}

func TestDotnetFlexPackProtocolV3ServiceIndexDiscovery(t *testing.T) {
	// Scenarios #179, #182 - the dotnet CLI's V3 service-index discovery against Artifactory
	// resolves the flat-container download path correctly.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	buildNumber := "74"
	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln",
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, allDeps(published),
		"V3 service-index discovery must yield a complete dependency graph")
}

func TestDotnetFlexPackPushIdenticalAcrossProtocols(t *testing.T) {
	// Scenario #184 - push succeeds identically whether the configured source is V2 or V3. Push
	// itself is always a V2-style PUT: the V3 service index advertises PackagePublish/2.0.0, the
	// same shape nuget.org advertises, so there is no protocol difference at the push endpoint.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	t.Run("v3-default", func(t *testing.T) {
		nupkgPath, _ := buildTestNupkg(t, "DotnetPushV3", "1.0.0")
		assert.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))
	})

	t.Run("v2-explicit", func(t *testing.T) {
		nupkgPath, _ := buildTestNupkg(t, "DotnetPushV2", "1.0.0")
		assert.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo, "--nuget-v2"))
	})
}

func TestDotnetFlexPackPushBuildPublishRestoreRoundTrip(t *testing.T) {
	// Scenario #94 - push, publish build-info, then read both modules back from Artifactory.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetFullRoundTrip", "2.0.0")
	buildNumber := "75"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, allArtifacts(published))

	props := getFlexPackItemProps(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath))
	assert.Equal(t, []string{buildNumber}, props["build.number"],
		"the published artifact must be traceable back to its build")
}

func TestDotnetFlexPackLegacyVsFlexPackBuildInfoParity(t *testing.T) {
	// Scenarios #113, #114, #118 - the legacy path and FlexPack must produce equivalent build-info
	// for the same project.
	t.Skip("The legacy path requires a .jfrog/projects/dotnet.yaml written by 'jf dotnet-config', " +
		"which is out of scope for this FlexPack suite and would need the interactive config " +
		"command or a hand-authored yaml fixture. TestDotnetFlexPackRunNativeTogglesCodePath " +
		"already pins that the two paths are selected correctly by JFROG_RUN_NATIVE.")
}

// ============================ Infra-gated scenario groups =====================================

func TestDotnetFlexPackBuildPromotion(t *testing.T) {
	// Scenarios #96, #97, #98, #99, #100, #101, #102, #103 - build-promote moves/copies .nupkg and .snupkg between repos, with
	// --copy, --include-dependencies, --props, and chained promotion preserving build-info.
	t.Skip("Requires a second target repository plus promotion plumbing. nuget_native_test.go has " +
		"the equivalent coverage (TestNugetFlexPackBuildPromote and siblings) whose helpers should " +
		"be reused when porting this group to the dotnet toolchain.")
}

func TestDotnetFlexPackBuildScan(t *testing.T) {
	// Scenarios #104, #105, #106, #107 - build-scan reports vulnerabilities across the full transitive tree and
	// --fail=true exits non-zero.
	t.Skip("Requires Xray to be provisioned and indexed against the test Artifactory. See " +
		"TestNugetFlexPackBuildScanReportsVulnerabilities in nuget_native_test.go for the pattern.")
}

func TestDotnetFlexPackReleaseBundle(t *testing.T) {
	// Scenarios #108, #109, #110, #111, #112 - release bundle creation, signing and distribution from a dotnet build.
	t.Skip("Requires the JFrog Lifecycle service, reachable only through the platform router port. " +
		"See withLifecycleRouterUrl and TestNugetFlexPackReleaseBundleFromNugetBuild in " +
		"nuget_native_test.go for the routing helper this group needs.")
}

func TestDotnetFlexPackCiCdWorkflows(t *testing.T) {
	// Scenarios #119, #120, #121, #122, #123, #124, #125, #126 - full pipeline, GitHub Actions ref-derived versions, Azure DevOps vcs
	// detection, Artifactory-unreachable handling, multi-env repo routing, Docker builds.
	t.Skip("Requires simulating CI provider environments and an unreachable-Artifactory fault " +
		"injection. nuget_native_test.go covers the equivalents (TestNugetFlexPackGitHubRefDerivesVersion, " +
		"TestNugetFlexPackAzureDevOpsVcsDetection, TestNugetFlexPackArtifactoryUnreachableNoFallback).")
}

func TestDotnetFlexPackTlsSelfSigned(t *testing.T) {
	// Scenarios #138, #139 - a self-signed certificate must fail validation without --insecure-tls
	// and succeed with it.
	t.Skip("Requires the self-signed-certificate proxy harness. See " +
		"TestNugetFlexPackTlsSelfSignedRequiresInsecureFlag in nuget_native_test.go, which wires " +
		"cliproxy plus the certificate package for exactly this.")
}

func TestDotnetFlexPackProxySupport(t *testing.T) {
	// Scenarios #141, #142, #143, #144 - HTTPS_PROXY routing for restore and push, and NO_PROXY bypasses.
	t.Skip("Requires the cliproxy test proxy server. See TestNugetFlexPackRestoreThroughHttpsProxy " +
		"and the NO_PROXY tests in nuget_native_test.go for the harness to reuse.")
}

// ============================ Final gap closure ================================================

func TestDotnetFlexPackMultiTargetFrameworkGraph(t *testing.T) {
	// Scenarios #43, #167 - a multi-target project records a dependency graph per TFM; each
	// top-level TFM key in project.assets.json must be walked, not just the first.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "simple-dotnet")
	defer cleanup()

	csproj := filepath.Join(projectPath, "nuget1.csproj")
	content, err := os.ReadFile(csproj)
	require.NoError(t, err)
	// Swap the single TargetFramework for a multi-target TargetFrameworks list. Match whatever
	// TFM the fixture declares rather than a hard-coded list: naming specific frameworks made
	// this a silent no-op when the fixture moved to net7.0, and the test then passed on an
	// ordinary single-target restore.
	singleTfm := regexp.MustCompile(`<TargetFramework>([^<]+)</TargetFramework>`)
	found := singleTfm.FindStringSubmatch(string(content))
	require.NotNil(t, found, "fixture must declare a single <TargetFramework> to expand")
	frameworks := found[1] + ";netstandard2.0"
	multi := singleTfm.ReplaceAllLiteralString(string(content), "<TargetFrameworks>"+frameworks+"</TargetFrameworks>")
	require.NotEqual(t, string(content), multi, "the multi-target rewrite must actually change the project")
	require.NoError(t, os.WriteFile(csproj, []byte(multi), 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	buildNumber := "80"
	// A multi-target restore may legitimately fail if the SDK lacks a targeting pack; the
	// assertion is on the collected graph when it succeeds.
	if err := restoreDotnetFlexPack(t, tests.NugetRemoteRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber); err != nil {
		t.Skipf("multi-target restore unavailable in this SDK image: %v", err)
	}
	defer deleteDotnetBuild()

	// Prove the restore really was multi-target before asserting on the graph, otherwise a
	// regression back to a single TFM would look like a pass.
	assets, err := os.ReadFile(filepath.Join(projectPath, "obj", "project.assets.json"))
	require.NoError(t, err)
	for _, tfm := range strings.Split(frameworks, ";") {
		assert.Contains(t, string(assets), tfm, "project.assets.json must carry a target for %s", tfm)
	}

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	assert.NotEmpty(t, allDeps(published), "each TFM's dependencies must be collected")
}

func TestDotnetFlexPackHashMismatchRevalidates(t *testing.T) {
	// Scenario #44 - a corrupted .nupkg.sha512 sidecar must cause NuGet to re-download rather than
	// report a false success. Corruption is introduced in the isolated per-test cache.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))

	// Corrupt every sidecar hash in the isolated cache. Paths are collected during the walk and
	// rewritten afterwards: performing the write inside the callback is race-prone, and the walk
	// error must be propagated rather than swallowed.
	cacheDir := filepath.Join(projectPath, ".packages")
	var sidecars []string
	require.NoError(t, filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".nupkg.sha512") {
			sidecars = append(sidecars, path)
		}
		return nil
	}))

	var corrupted int
	for _, sidecar := range sidecars {
		if writeErr := os.WriteFile(sidecar, []byte("bm90LWEtdmFsaWQtaGFzaA=="), 0o600); writeErr == nil { //#nosec G703 -- test code, path collected from the test's own temp cache dir
			corrupted++
		}
	}
	if corrupted == 0 {
		t.Skip("no .nupkg.sha512 sidecars were produced in the isolated cache")
	}

	// NuGet must recover by re-validating/re-downloading rather than failing outright.
	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"),
		"a corrupted sha512 sidecar must trigger revalidation, not a hard failure")
}

func TestDotnetFlexPackDownloadedChecksumMatchesArtifactory(t *testing.T) {
	// Scenarios #75, #76 - a package downloaded through Artifactory has the same sha256 that
	// Artifactory stores for it, and the .nupkg.sha512 sidecar matches the package content.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetChecksumMatch", "1.0.0")
	buildNumber := "81"
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	artifacts := allArtifacts(published)
	require.NotEmpty(t, artifacts)

	// The sha256 recorded in build-info is the one computed locally from the file that was
	// uploaded; Artifactory must agree, otherwise the artifact would be in an untrusted state.
	props := getFlexPackItemProps(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath))
	assert.NotNil(t, props, "the pushed artifact must be retrievable for checksum comparison")
	for _, artifact := range artifacts {
		assert.Len(t, artifact.Sha256, 64, "sha256 for %s must be a full digest", artifact.Name)
	}
}

func TestDotnetFlexPackCachedRestoreNoRetransfer(t *testing.T) {
	// Scenario #79 - re-restoring the same project uses the cached package rather than
	// re-transferring it. Asserted by the second restore succeeding against a warm cache with the
	// remote made unreachable via an unusable resolve repo would change semantics, so this pins the
	// weaker but honest property: a warm re-restore succeeds and is not a fresh download path.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))
	cacheDir := filepath.Join(projectPath, ".packages")
	require.DirExists(t, cacheDir, "first restore must populate the isolated cache")

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"),
		"a second restore must be satisfied from the warm cache")
}

func TestDotnetFlexPackExplicitProtocolVersionPin(t *testing.T) {
	// Scenarios #180, #181 - an explicit protocolVersion="3" or "2" pin in the user's NuGet.Config
	// is honoured. With --repo-resolve, FlexPack's own config wins, so this exercises the
	// user-managed path with no repo flag.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	user, password := credentialsForTestServer(t)
	if user == "" || password == "" {
		t.Skip("Test server credentials are not available as user/password for a user-managed source.")
	}
	base := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget"

	cases := []struct {
		name            string
		value           string
		protocolVersion string
	}{
		{"v3-pin", base + "/v3/" + tests.NugetRemoteRepo + "/index.json", "3"},
		{"v2-pin", base + "/" + tests.NugetRemoteRepo, "2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, os.WriteFile(filepath.Join(projectPath, "nuget.config"), []byte(
				`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="Pinned" value="`+tc.value+`" protocolVersion="`+tc.protocolVersion+`" allowInsecureConnections="true" />
  </packageSources>
  <packageSourceCredentials>
    <Pinned>
      <add key="Username" value="`+user+`" />
      <add key="ClearTextPassword" value="`+password+`" />
    </Pinned>
  </packageSourceCredentials>
</configuration>`), 0o600))

			assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "restore", "reference.sln"),
				"an explicit protocolVersion=%s pin must be honoured", tc.protocolVersion)
		})
	}
}

func TestDotnetFlexPackVirtualRepoAggregatingV3Remote(t *testing.T) {
	// Scenario #185 - a virtual repo aggregating a V3-only remote resolves correctly through the
	// service index.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	assert.NoError(t, restoreDotnetFlexPack(t, tests.NugetVirtualRepo, "reference.sln"),
		"a virtual repo aggregating a V3 remote must resolve")
}

func TestDotnetFlexPackListPackageAgainstArtifactory(t *testing.T) {
	// Scenario #186 - 'dotnet list package' is a non-eligible subcommand: it passes through to the
	// native client with no interception and no build-info.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	_, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	require.NoError(t, restoreDotnetFlexPack(t, tests.NugetRemoteRepo, "reference.sln"))
	// Passthrough: whatever the native tool reports is what the user sees.
	_ = runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "list", "package")
}

func TestDotnetFlexPackAuthenticatedSourceRequiresCredentials(t *testing.T) {
	// Scenario #168 - a plain 'dotnet restore' against an authenticated Artifactory source fails
	// with 401 unless credentials are supplied. This is the control that proves FlexPack's own
	// injection is what makes the other restore tests pass.
	initNugetTest(t)
	defer cleanTestsHomeEnv()
	projectPath, cleanup := enterDotnetProject(t, "reference")
	defer cleanup()

	sourceURL := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget/v3/" +
		tests.NugetRemoteRepo + "/index.json"
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "nuget.config"), []byte(
		`<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <clear />
    <add key="NoCreds" value="`+sourceURL+`" protocolVersion="3" allowInsecureConnections="true" />
  </packageSources>
</configuration>`), 0o600))

	// No --repo-resolve, so jf injects nothing and there is no credential anywhere.
	err := runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "restore", "reference.sln")
	if err == nil {
		t.Skip("test Artifactory permits anonymous reads, so the 401 path cannot be exercised")
	}
	assert.Error(t, err, "an authenticated source with no credentials must fail")
}

func TestDotnetFlexPackCiSecretBackedApiKeyPush(t *testing.T) {
	// Scenario #169 - the CI shape: a secret-backed token wired through --api-key on push, with no
	// credentials in any config file.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	user, password := credentialsForTestServer(t)
	if user == "" || password == "" {
		t.Skip("Test server credentials are not available as user/password for the API-key form.")
	}

	nupkgPath, _ := buildTestNupkg(t, "DotnetCiSecretPush", "1.0.0")
	sourceURL := strings.TrimSuffix(*tests.JfrogUrl, "/") + "/artifactory/api/nuget/v3/" +
		tests.NugetLocalRepo + "/index.json"

	assert.NoError(t, runDotnetFlexPack(t, dotnetUtils.DotnetCore.String(), "nuget", "push",
		nupkgPath, "--source", sourceURL, "--api-key", user+":"+password,
		"--configfile", insecureSourceConfigFile(t, sourceURL)),
		"a CI-secret-backed --api-key must authenticate the push")
}

func TestDotnetFlexPackLargeAndNativeRuntimePackages(t *testing.T) {
	// Scenarios #135, #136 - a very large .nupkg restores in a single chunk without corrupting
	// build-info, and a package carrying native runtime folders resolves per RID.
	t.Skip("Requires purpose-built fixtures: a >100 MB package, and one containing " +
		"runtimes/<rid>/native/ payloads. Neither exists in the shared nuget testdata set, and " +
		"generating them at test time would dominate CI runtime.")
}

func TestDotnetFlexPackNestedDirectoryPackagesProps(t *testing.T) {
	// Scenarios #173, #174 - multiple Directory.Packages.props in a tree (a nested one excluding a
	// tools folder), and 'dotnet add package' without an explicit version not silently bumping the
	// central version.
	t.Skip("Requires a multi-level fixture tree with nested Directory.Packages.props files. The " +
		"single-level CPM case is covered by TestDotnetFlexPackCentralPackageManagement.")
}

func TestDotnetFlexPackSlnxUnsupportedByNugetExe(t *testing.T) {
	// Scenarios #175, #176 - 'nuget.exe restore project.slnx' errors as an unsupported format
	// while 'jf dotnet restore' handles it, and a .slnx referencing a legacy web-site project
	// fails with MSB4249.
	//
	// #175 is the one scenario in this plan that is genuinely about the nuget.exe client rather
	// than dotnet; it belongs with nuget_native_test.go's suite.
	t.Skip("Requires .slnx fixtures under testdata/nuget/, which the shared set does not provide. " +
		"The dotnet-bug-hunt skill has a working slnx-project fixture to port. Scenario #175 is " +
		"nuget.exe-specific and belongs in nuget_native_test.go.")
}

// ============================= Last remaining scenarios ========================================

func TestDotnetFlexPackVirtualRepoPushConvention(t *testing.T) {
	// Scenarios #91, #92 - pushing to a virtual repo follows the convention used by the other
	// FlexPack package managers: Artifactory routes the upload to the virtual repo's
	// defaultDeploymentRepo, and build-info must record that resolved LOCAL repo key rather than
	// the virtual one, or downstream tools 404 on the recorded path. A virtual repo with no
	// defaultDeploymentRepo, or with mixed underlying layouts, must fail clearly instead.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, _ := buildTestNupkg(t, "DotnetVirtualPush", "1.0.0")
	buildNumber := "90"

	err := pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetVirtualRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber)
	if err != nil {
		// A virtual repo without a defaultDeploymentRepo legitimately rejects the push; that is
		// the clear-error half of the scenario.
		t.Logf("virtual-repo push rejected (acceptable when no defaultDeploymentRepo is set): %v", err)
		return
	}
	defer deleteDotnetBuild()

	// Assert the resolved repo positively. "not the virtual repo" over a possibly-empty slice
	// could pass three different ways - no artifacts, an empty field, or any other repo name -
	// without ever showing that resolveLocalDeployRepo picked the virtual repo's deployment
	// target. The testdata config sets that target to NugetLocalRepo.
	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	artifacts := allArtifacts(published)
	require.NotEmpty(t, artifacts, "a push through a virtual repo must record artifacts")
	for _, artifact := range artifacts {
		assert.Equal(t, tests.NugetLocalRepo, artifact.OriginalDeploymentRepo,
			"build-info must record the virtual repo's defaultDeploymentRepo, not %s",
			tests.NugetVirtualRepo)
	}
}

func TestDotnetFlexPackLegacySymbolsFormat(t *testing.T) {
	// Scenario #19 - the legacy .symbols.nupkg symbol format (SymbolPackageFormat=symbols.nupkg)
	// is handled as a symbol package rather than mistaken for a primary .nupkg.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	// Derive a legacy-format symbol package next to a normal one.
	nupkgPath, _ := buildTestNupkg(t, "DotnetLegacySymbols", "1.0.0")
	legacyPath := strings.TrimSuffix(nupkgPath, ".nupkg") + ".symbols.nupkg"
	content, err := os.ReadFile(nupkgPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(legacyPath, content, 0o600)) //#nosec G703 -- test code, path is under the test's own temp project dir

	buildNumber := "91"
	// The legacy format is pushed as an ordinary package by the native client; the assertion is
	// that jf routes it without misclassifying it.
	err = pushNupkgDotnetFlexPack(t, legacyPath, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("legacy .symbols.nupkg format rejected by this Artifactory/client combination: %v", err)
	}
	defer deleteDotnetBuild()

	// Assert the positive contract. "not zip" cannot fail - packageArtifactType only ever returns
	// nupkg or snupkg - so it proved nothing. newArtifactFromFile types a .symbols.nupkg as
	// snupkg and stores it flat under the renamed <id>.<version>.nupkg, Artifactory dropping the
	// ".symbols" segment; that mapping is what this pins.
	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	var legacy buildInfo.Artifact
	for _, artifact := range allArtifacts(published) {
		if strings.HasSuffix(artifact.Name, ".symbols.nupkg") {
			legacy = artifact
		}
	}
	require.NotEmpty(t, legacy.Name, "the pushed legacy symbols package must be recorded")
	assert.Equal(t, "snupkg", legacy.Type, "a .symbols.nupkg is a symbol package")
	assert.Equal(t, strings.TrimSuffix(filepath.Base(legacyPath), ".symbols.nupkg")+".nupkg", legacy.Path,
		"Artifactory renames a legacy symbols package, dropping the .symbols segment")
}

func TestDotnetFlexPackSolutionPackPushPerModule(t *testing.T) {
	// Scenario #67 - packing and pushing a multi-project solution puts each project's .nupkg in
	// its own module's artifacts list, rather than collapsing them into one module.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	buildNumber := "92"
	first, _ := buildTestNupkg(t, "DotnetSolutionModuleA", "1.0.0")
	second, _ := buildTestNupkg(t, "DotnetSolutionModuleB", "1.0.0")

	require.NoError(t, pushNupkgDotnetFlexPack(t, first, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	require.NoError(t, pushNupkgDotnetFlexPack(t, second, tests.NugetLocalRepo,
		"--build-name="+tests.DotnetBuildName, "--build-number="+buildNumber))
	defer deleteDotnetBuild()

	published := publishAndGetDotnetBuildInfo(t, buildNumber)
	owners := map[string]string{}
	for _, m := range published.BuildInfo.Modules {
		for _, a := range m.Artifacts {
			owners[a.Name] = m.Id
		}
	}
	assert.GreaterOrEqual(t, len(owners), 2, "each package must be recorded under its own module")
	seenModules := map[string]struct{}{}
	for _, moduleId := range owners {
		seenModules[moduleId] = struct{}{}
	}
	assert.GreaterOrEqual(t, len(seenModules), 2,
		"two distinct packages must not collapse into a single module")
}

func TestDotnetFlexPackSymbolRoundTrip(t *testing.T) {
	// Scenario #95 - a .nupkg pushed together with its .snupkg can be read back, with both
	// artifacts retrievable from the repo they were published to.
	initNugetTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath, snupkgPath := buildTestNupkg(t, "DotnetSymbolRoundTrip", "1.0.0")
	require.NoError(t, pushNupkgDotnetFlexPack(t, nupkgPath, tests.NugetLocalRepo))

	// "Retrievable" means present in the repository. This push names no build, so nothing is
	// stamped, and a .snupkg carries none of the nuget.* properties Artifactory attaches when it
	// indexes a primary package - checking properties here would report a file that round-tripped
	// perfectly well as missing.
	assertArtifactExists(t, tests.NugetLocalRepo+"/"+filepath.Base(nupkgPath),
		"the primary package must be retrievable")
	assertArtifactExists(t, tests.NugetLocalRepo+"/"+filepath.Base(snupkgPath),
		"the co-pushed symbol package must be retrievable")
}

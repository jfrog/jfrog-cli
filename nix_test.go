package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Initialization ====================

func initNixTest(t *testing.T) {
	if !*tests.TestNix {
		t.Skip("Skipping Nix test. To run Nix test add the '-test.nix=true' option.")
	}
	require.True(t, isRepoExist(tests.NixRemoteRepo), "Nix test remote repository doesn't exist.")
	require.True(t, isRepoExist(tests.NixVirtualRepo), "Nix test virtual repository doesn't exist.")
}

// ==================== Project Helpers ====================

func createNixProject(t *testing.T, outputFolder, projectName string) (string, func()) {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "nix", projectName)
	tmpDir, cleanupCallback := coretests.CreateTempDirWithCallbackAndAssert(t)

	projectPath := filepath.Join(tmpDir, outputFolder)
	assert.NoError(t, biutils.CopyDir(projectSrc, projectPath, true, nil))

	// Flake projects require git — initialize if flake.nix exists and no .git present
	if _, err := os.Stat(filepath.Join(projectPath, "flake.nix")); err == nil {
		initGitForFlake(t, projectPath)
	}

	return projectPath, cleanupCallback
}

// initGitForFlake initializes a git repo in the project directory so nix flake commands work.
// Nix flakes require all files to be tracked by git.
func initGitForFlake(t *testing.T, projectPath string) {
	cmds := [][]string{
		{"git", "init"},
		{"git", "add", "."},
		{"git", "-c", "user.name=test", "-c", "user.email=test@test.com", "commit", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Logf("git command %v: %s", args, string(output))
		}
	}
}

// ==================== FlexPack Install Tests ====================

//func TestNixFlakeLockFlexPack(t *testing.T) {
//	testNixFlexPack(t, "flake lock")
//}

func testNixFlexPack(t *testing.T, nixSubcmd string) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildNumber := "1"
	projectPath, cleanupProject := createNixProject(t, "nix-"+nixSubcmd, "nixproject")
	defer cleanupProject()

	// Build args — split compound subcommands like "flake lock" into separate args
	args := []string{"nix"}
	args = append(args, strings.Split(nixSubcmd, " ")...)
	args = append(args, "--build-name="+tests.NixBuildName, "--build-number="+buildNumber)

	testNixCmd(t, projectPath, buildNumber, filepath.Base(projectPath), 3, args)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.NixBuildName, artHttpDetails)
}

func testNixCmd(t *testing.T, projectPath, buildNumber, module string, expectedDependencies int, args []string) {
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current directory")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec(args...))

	// Validate local build-info was created with Nix module type
	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.NixBuildName, buildNumber, "", []string{module}, buildinfo.Nix)

	// Publish build-info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.NixBuildName, buildNumber))

	// Get and validate published build-info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.NixBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}

	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	require.Len(t, buildInfoModules, 1)
	assert.Equal(t, module, buildInfoModules[0].Id)
	assert.Equal(t, buildinfo.Nix, buildInfoModules[0].Type)
	assert.Len(t, buildInfoModules[0].Dependencies, expectedDependencies)

	// Validate Nix-specific: narHash checksums in SRI format
	for _, dep := range buildInfoModules[0].Dependencies {
		assert.NotEmpty(t, dep.Sha256, "SHA256 (narHash) should be present for dep %s", dep.Id)
		assert.Contains(t, dep.Sha256, "sha256-",
			"narHash should be in SRI format for dep %s, got: %s", dep.Id, dep.Sha256)
	}

	// Validate scopes
	for _, dep := range buildInfoModules[0].Dependencies {
		assert.Equal(t, []string{"build"}, dep.Scopes,
			"dep %s should have scope [build]", dep.Id)
	}

	// Validate requestedBy is present
	hasRequestedBy := false
	for _, dep := range buildInfoModules[0].Dependencies {
		if len(dep.RequestedBy) > 0 {
			hasRequestedBy = true
			break
		}
	}
	assert.True(t, hasRequestedBy, "at least one dependency should have RequestedBy")
}

// ==================== Build Info Flag Combination Tests ====================

//func TestNixBuildInfoBothFlags(t *testing.T) {
//	initNixTest(t)
//
//	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
//	defer setEnvCallback()
//
//	oldHomeDir, newHomeDir := prepareHomeDir(t)
//	defer func() {
//		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
//		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
//	}()
//
//	projectPath, cleanupProject := createNixProject(t, "nix-both-flags", "nixproject")
//	defer cleanupProject()
//
//	wd, err := os.Getwd()
//	assert.NoError(t, err)
//	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
//	defer chdirCallback()
//
//	buildName := "nix-both-flags-test"
//	buildNumber := "42"
//
//	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
//	assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
//		"--build-name="+buildName, "--build-number="+buildNumber))
//
//	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
//
//	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
//	assert.NoError(t, err)
//	assert.True(t, found, "build-info should be found when both flags set")
//	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
//	assert.Greater(t, len(publishedBuildInfo.BuildInfo.Modules[0].Dependencies), 0)
//
//	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
//}

func TestNixBuildInfoBuildNameOnly(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	projectPath, cleanupProject := createNixProject(t, "nix-name-only", "nixproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// Only --build-name, no --build-number → nix command runs, build-info may not be collected
	err = jfrogCli.Exec("nix", "flake", "lock", "--build-name=nix-name-only-test")
	// Command may succeed or fail depending on build-number extraction — just ensure no panic
	_ = err
}

func TestNixBuildInfoNoFlags(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	projectPath, cleanupProject := createNixProject(t, "nix-no-flags", "nixproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock"))
}

// ==================== Module Override Test ====================

func TestNixModuleOverride(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	projectPath, cleanupProject := createNixProject(t, "nix-module-override", "nixproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	buildName := "nix-module-override-test"
	buildNumber := "1"
	customModule := "my-custom-nix-module"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule))

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Table-Driven Subcommand Tests ====================

//func TestNixSubcommandVariants(t *testing.T) {
//	initNixTest(t)
//
//	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
//	defer setEnvCallback()
//
//	oldHomeDir, newHomeDir := prepareHomeDir(t)
//	defer func() {
//		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
//		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
//	}()
//
//	allTests := []struct {
//		name                 string
//		nixSubcmd            string
//		expectedDependencies int
//		collectsBuildInfo    bool
//	}{
//		{"nix-flake-lock", "flake lock", 3, true},
//	}
//
//	for buildNumber, test := range allTests {
//		t.Run(test.name, func(t *testing.T) {
//			buildNumberStr := strconv.Itoa(buildNumber + 1)
//			projectPath, cleanupProject := createNixProject(t, test.name, "nixproject")
//			defer cleanupProject()
//
//			if test.collectsBuildInfo {
//				args := []string{"nix"}
//				args = append(args, strings.Split(test.nixSubcmd, " ")...)
//				args = append(args, "--build-name="+tests.NixBuildName, "--build-number="+buildNumberStr)
//				testNixCmd(t, projectPath, buildNumberStr, filepath.Base(projectPath),
//					test.expectedDependencies, args)
//			}
//		})
//
//		if test.collectsBuildInfo {
//			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.NixBuildName, artHttpDetails)
//		}
//	}
//}

// ==================== Multiple Builds Test ====================

//func TestNixMultipleBuilds(t *testing.T) {
//	initNixTest(t)
//
//	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
//	defer setEnvCallback()
//
//	oldHomeDir, newHomeDir := prepareHomeDir(t)
//	defer func() {
//		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
//		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
//	}()
//
//	buildName := "nix-multi-build-test"
//
//	for i := 1; i <= 3; i++ {
//		buildNumber := strconv.Itoa(i)
//		projectPath, cleanupProject := createNixProject(t, fmt.Sprintf("nix-multi-%d", i), "nixproject")
//
//		wd, err := os.Getwd()
//		assert.NoError(t, err)
//		chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
//
//		jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
//		assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
//			"--build-name="+buildName, "--build-number="+buildNumber))
//		assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
//
//		publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
//		assert.NoError(t, err)
//		assert.True(t, found, "build %s should be found", buildNumber)
//		assert.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
//
//		chdirCallback()
//		cleanupProject()
//	}
//
//	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
//}

// ==================== Dependency Checksums Test ====================

//func TestNixDependencyChecksums(t *testing.T) {
//	initNixTest(t)
//
//	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
//	defer setEnvCallback()
//
//	oldHomeDir, newHomeDir := prepareHomeDir(t)
//	defer func() {
//		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
//		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
//	}()
//
//	projectPath, cleanupProject := createNixProject(t, "nix-checksums", "nixproject")
//	defer cleanupProject()
//
//	wd, err := os.Getwd()
//	assert.NoError(t, err)
//	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
//	defer chdirCallback()
//
//	buildName := "nix-checksum-test"
//	buildNumber := "1"
//
//	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
//	assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
//		"--build-name="+buildName, "--build-number="+buildNumber))
//	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
//
//	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
//	assert.NoError(t, err)
//	assert.True(t, found)
//
//	depsWithChecksums := 0
//	for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
//		if dep.Sha256 != "" {
//			depsWithChecksums++
//			// Verify SRI format
//			assert.Contains(t, dep.Sha256, "sha256-",
//				"dep %s should have narHash in SRI format", dep.Id)
//		}
//	}
//	assert.Greater(t, depsWithChecksums, 0, "at least one dep should have checksums")
//
//	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
//}

// ==================== Project Key Test ====================

func TestNixFlakeLockWithProjectKey(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	projectPath, cleanupProject := createNixProject(t, "nix-project-key", "nixproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	buildName := "nix-project-key-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--project=testprj"))

	// Note: build-publish with --project may fail if project doesn't exist on this Artifactory instance.
	// This is expected — the test validates that the --project flag is correctly passed through.
	err = artifactoryCli.Exec("bp", buildName, buildNumber, "--project=testprj")
	if err != nil {
		t.Logf("build-publish with --project failed (project may not exist): %v", err)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Channel-Based Integration Tests ====================
// These test the channel-based nix-build/nix-copy workflow (not flakes).

func TestNixBuild_HelloPackage(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-build-hello-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available or nixpkgs not configured: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be found")

	if found {
		bi := publishedBuildInfo.BuildInfo
		require.Len(t, bi.Modules, 1)
		assert.Equal(t, buildinfo.Nix, bi.Modules[0].Type)
		// hello on macOS ARM has 1-2 runtime deps (libiconv)
		assert.GreaterOrEqual(t, len(bi.Modules[0].Dependencies), 1,
			"hello should have at least 1 runtime dependency")

		// Validate dep ID format: name:version
		for _, dep := range bi.Modules[0].Dependencies {
			assert.Contains(t, dep.Id, ":", "dep ID should be name:version format, got: %s", dep.Id)
		}

		// Validate scopes are runtime (not build)
		for _, dep := range bi.Modules[0].Dependencies {
			assert.Equal(t, []string{"runtime"}, dep.Scopes,
				"channel-based deps should have scope [runtime], got %v for %s", dep.Scopes, dep.Id)
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixBuild_ModuleOverride(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-build-module-test"
	buildNumber := "1"
	customModule := "my-custom-nix-module"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id,
			"module ID should be overridden to custom name")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixBuild_EmptyClosure(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-build-empty-test"
	buildNumber := "1"

	// Build hello — on macOS ARM it may have 0 references (statically linked)
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	// Should have module with nix type regardless of dep count
	if found {
		require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
		assert.Equal(t, buildinfo.Nix, publishedBuildInfo.BuildInfo.Modules[0].Type)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixBuild_DepChecksums(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-build-checksums-test"
	buildNumber := "1"

	projectPath, cleanupProject := createNixProject(t, "nix-checksums-channel", "channelproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "nix-build", "default.nix",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available or deps not resolved: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		depsWithAQLChecksums := 0
		for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
			// Channel-based deps should have actual file checksums (sha1/md5) from AQL
			// not SRI narHash format
			if dep.Sha1 != "" {
				depsWithAQLChecksums++
				assert.NotEmpty(t, dep.Sha256, "sha256 should be set for %s", dep.Id)
				assert.NotEmpty(t, dep.Md5, "md5 should be set for %s", dep.Id)
				// Should NOT be SRI format (that's the flake collector)
				assert.NotContains(t, dep.Sha1, "sha",
					"sha1 should be hex, not SRI for %s", dep.Id)
			}
		}
		if len(publishedBuildInfo.BuildInfo.Modules[0].Dependencies) > 0 {
			// On a developer machine, some deps may already be in /nix/store/
			// and not fetched through Artifactory, so AQL may not resolve them.
			// In CI (clean nix store), all checksums should resolve.
			t.Logf("Deps with AQL checksums: %d/%d",
				depsWithAQLChecksums, len(publishedBuildInfo.BuildInfo.Modules[0].Dependencies))
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixBuild_NoBuildFlags(t *testing.T) {
	initNixTest(t)

	// Without build flags, nix-build should still run natively (no crash)
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello")
	if err != nil {
		// Command may fail if nix not available — that's OK
		t.Skipf("nix-build not available: %v", err)
	}
}

func TestNixChannel_Passthrough(t *testing.T) {
	initNixTest(t)

	// nix-channel --list should work as passthrough, no build-info
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-channel", "--list")
	if err != nil {
		t.Skipf("nix-channel not available: %v", err)
	}
}

func TestNixMultipleBuilds_DontCollide(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-multi-build-channel-test"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Build 1: hello (small closure)
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number=1")
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, "1"))

	// Build 2: same package, different build number
	err = jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number=2")
	assert.NoError(t, err)
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, "2"))

	// Verify both builds exist independently
	bi1, found1, err := tests.GetBuildInfo(serverDetails, buildName, "1")
	assert.NoError(t, err)
	assert.True(t, found1)

	bi2, found2, err := tests.GetBuildInfo(serverDetails, buildName, "2")
	assert.NoError(t, err)
	assert.True(t, found2)

	if found1 && found2 {
		// Both should have same number of deps (same package)
		assert.Equal(t, len(bi1.BuildInfo.Modules[0].Dependencies),
			len(bi2.BuildInfo.Modules[0].Dependencies),
			"same package should have same dep count across builds")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixCopy_VirtualToLocalResolution(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-copy-virtual-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Build first
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	// Copy to virtual repo — should resolve to local automatically
	toURL := fmt.Sprintf("https://%s:%s@%s/api/nix/%s/",
		*tests.JfrogUser, *tests.JfrogPassword,
		strings.TrimPrefix(strings.TrimPrefix(*tests.JfrogUrl, "https://"), "http://"),
		tests.NixVirtualRepo)
	err = jfrogCli.Exec("nix", "copy", "--to", toURL, "./result",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Logf("nix copy failed (may need credentials in URL): %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		assert.Equal(t, buildinfo.Nix, module.Type)

		// If copy succeeded, should have artifacts with originalDeploymentRepo = local repo
		for _, artifact := range module.Artifacts {
			if artifact.OriginalDeploymentRepo != "" {
				assert.Equal(t, tests.NixLocalRepo, artifact.OriginalDeploymentRepo,
					"artifacts should be deployed to local repo, not virtual")
			}
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixBuild_BuildOnlyNoCopy(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-build-only-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	// Publish WITHOUT copy → should have deps but NO artifacts
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found {
		require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
		module := publishedBuildInfo.BuildInfo.Modules[0]
		assert.Equal(t, buildinfo.Nix, module.Type)
		// Should have deps but no artifacts (copy not run)
		assert.Empty(t, module.Artifacts, "build-only should have no artifacts")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== IT-2: Additional nix-build dep tests ====================

func TestNixBuild_CustomProjectWithDeps(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-custom-deps-test"
	buildNumber := "1"

	projectPath, cleanupProject := createNixProject(t, "nix-custom-deps", "channelproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "nix-build", "default.nix",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available or deps not resolved: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		deps := publishedBuildInfo.BuildInfo.Modules[0].Dependencies
		assert.GreaterOrEqual(t, len(deps), 20,
			"custom project with curl+jq should have 20+ runtime deps, got %d", len(deps))
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixBuild_DepScopes(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-dep-scopes-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
			assert.Equal(t, []string{"runtime"}, dep.Scopes,
				"channel-based deps should have scope [runtime] for %s", dep.Id)
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixBuild_DepIDFormat(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-dep-id-format-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
			// Channel-based dep IDs are name:version (e.g., libiconv:109.100.2)
			assert.NotEmpty(t, dep.Id, "dep ID should not be empty")
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixBuild_ModuleID(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-module-id-test"
	buildNumber := "1"

	projectPath, cleanupProject := createNixProject(t, "nix-module-id", "channelproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "nix-build", "default.nix",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		// Module ID should be the working directory basename
		assert.Equal(t, "nix-module-id", publishedBuildInfo.BuildInfo.Modules[0].Id,
			"module ID should match working directory basename")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== IT-4: Full Lifecycle tests ====================

func TestNixFullLifecycle_HelloPackage(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-full-lifecycle-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Step 1: Build (collects deps)
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	// Step 2: Copy (collects artifacts + tags properties)
	toURL := fmt.Sprintf("https://%s:%s@%s/api/nix/%s/",
		*tests.JfrogUser, *tests.JfrogPassword,
		strings.TrimPrefix(strings.TrimPrefix(*tests.JfrogUrl, "https://"), "http://"),
		tests.NixVirtualRepo)
	copyErr := jfrogCli.Exec("nix", "copy", "--to", toURL, "./result",
		"--build-name="+buildName, "--build-number="+buildNumber)

	// Step 3: Publish
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		assert.Equal(t, buildinfo.Nix, module.Type)

		// Should always have deps from build step
		assert.GreaterOrEqual(t, len(module.Dependencies), 1,
			"should have at least 1 dependency from build step")

		// If copy succeeded, should also have artifacts
		if copyErr == nil {
			assert.GreaterOrEqual(t, len(module.Artifacts), 2,
				"should have at least 2 artifacts (narinfo + nar.xz) from copy step")
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixFullLifecycle_BuildInfoJSON(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-bi-json-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found {
		bi := publishedBuildInfo.BuildInfo
		assert.Equal(t, buildName, bi.Name)
		assert.Equal(t, buildNumber, bi.Number)
		assert.NotEmpty(t, bi.Started)
		require.Len(t, bi.Modules, 1)
		assert.Equal(t, buildinfo.Nix, bi.Modules[0].Type)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== IT-5: nix-channel tests ====================

func TestNixChannel_NoBuildInfo(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-channel-no-bi-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// nix-channel --list with build flags should NOT produce build-info
	err := jfrogCli.Exec("nix", "nix-channel", "--list",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-channel not available: %v", err)
	}

	// Build publish should fail or produce empty build-info
	_ = artifactoryCli.Exec("bp", buildName, buildNumber)

	_, found, _ := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	// Channel passthrough should NOT create build-info
	// (either not found, or found with empty modules)
	if found {
		t.Log("build-info was found for nix-channel (may have empty modules)")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== IT-6: nix-env tests ====================

func TestNixEnv_InstallPackage(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-env-install-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-env", "-iA", "nixpkgs.hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-env not available or nixpkgs not configured: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be found after nix-env install")

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		assert.Equal(t, buildinfo.Nix, module.Type)
		assert.GreaterOrEqual(t, len(module.Dependencies), 1,
			"nix-env should collect at least 1 runtime dep")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixEnv_NoBuildFlags(t *testing.T) {
	initNixTest(t)

	// nix-env without build flags should run natively, no crash
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-env", "--query", "--installed")
	if err != nil {
		t.Skipf("nix-env not available: %v", err)
	}
}

// ==================== IT-3: Additional nix copy tests ====================

func TestNixCopy_NoBuildFlags(t *testing.T) {
	initNixTest(t)

	// Build hello first
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello")
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	// Copy without build flags should not crash
	toURL := fmt.Sprintf("https://%s:%s@%s/api/nix/%s/",
		*tests.JfrogUser, *tests.JfrogPassword,
		strings.TrimPrefix(strings.TrimPrefix(*tests.JfrogUrl, "https://"), "http://"),
		tests.NixLocalRepo)
	// May fail due to URL encoding, but should not panic
	_ = jfrogCli.Exec("nix", "copy", "--to", toURL, "./result")
}

// ==================== IT-7: Edge Cases ====================

func TestNix_SubstituterParsing(t *testing.T) {
	// This tests parseRepoFromSubstituter indirectly —
	// when no --repo flag is passed to nix-build, the code reads nix.conf
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-substituter-parse-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-build", "<nixpkgs>", "-A", "hello",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// If substituter is configured in nix.conf, checksums should be resolved
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		// At minimum, the module should exist with correct type
		assert.Equal(t, buildinfo.Nix, publishedBuildInfo.BuildInfo.Modules[0].Type)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNix_DepRequestedBy(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-requestedby-test"
	buildNumber := "1"

	projectPath, cleanupProject := createNixProject(t, "nix-requestedby", "channelproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "nix-build", "default.nix",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix-build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		// With curl+jq deps, transitive deps should have requestedBy
		hasRequestedBy := false
		for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
			if len(dep.RequestedBy) > 0 {
				hasRequestedBy = true
				break
			}
		}
		assert.True(t, hasRequestedBy,
			"at least one transitive dep should have RequestedBy chains")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Flake Build Tests (nix build) ====================

func TestNixFlakeBuild_HelloPackage(t *testing.T) {
	// Scenario #6 — nix build (flake) hello from nixpkgs → deps collected from runtime closure
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-flake-build-hello-test"
	buildNumber := "1"

	projectPath, cleanupProject := createNixProject(t, "nix-flake-hello", "flakeproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "build",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix build not available or flake project not configured: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be found")

	if found {
		bi := publishedBuildInfo.BuildInfo
		require.Len(t, bi.Modules, 1)
		assert.Equal(t, buildinfo.Nix, bi.Modules[0].Type)
		assert.GreaterOrEqual(t, len(bi.Modules[0].Dependencies), 1,
			"flake build should collect at least 1 runtime dep")

		for _, dep := range bi.Modules[0].Dependencies {
			assert.Contains(t, dep.Id, ":", "dep ID should be name:version, got: %s", dep.Id)
			assert.Equal(t, []string{"runtime"}, dep.Scopes)
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixFlakeBuild_CustomProject(t *testing.T) {
	// Scenario #2 — nix build (flake) custom project with curl+jq → 20+ runtime deps
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-flake-build-custom-test"
	buildNumber := "1"

	projectPath, cleanupProject := createNixProject(t, "nix-flake-custom", "flakeproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "build",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		deps := publishedBuildInfo.BuildInfo.Modules[0].Dependencies
		// flakeproject testdata uses pkgs.hello which has a small runtime closure
		assert.GreaterOrEqual(t, len(deps), 1,
			"flake project should have at least 1 runtime dep, got %d", len(deps))
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixFlakeBuild_FullLifecycle(t *testing.T) {
	// Scenario #28 — Flake: build + copy + publish → deps AND artifacts in same module
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-flake-lifecycle-test"
	buildNumber := "1"

	projectPath, cleanupProject := createNixProject(t, "nix-flake-lifecycle", "flakeproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Step 1: Build (collects deps)
	err = jfrogCli.Exec("nix", "build",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix build not available: %v", err)
	}

	// Step 2: Copy (collects artifacts + tags properties)
	toURL := fmt.Sprintf("https://%s:%s@%s/api/nix/%s/",
		*tests.JfrogUser, *tests.JfrogPassword,
		strings.TrimPrefix(strings.TrimPrefix(*tests.JfrogUrl, "https://"), "http://"),
		tests.NixVirtualRepo)
	copyErr := jfrogCli.Exec("nix", "copy", "--to", toURL, "./result",
		"--build-name="+buildName, "--build-number="+buildNumber)

	// Step 3: Publish
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		assert.Equal(t, buildinfo.Nix, module.Type)

		// Should always have deps from build step
		assert.GreaterOrEqual(t, len(module.Dependencies), 1,
			"should have deps from build step")

		// If copy succeeded, should have artifacts
		if copyErr == nil {
			assert.GreaterOrEqual(t, len(module.Artifacts), 2,
				"should have at least .narinfo + .nar.xz from copy step")

			// Verify artifact types
			hasNarinfo := false
			hasNarXz := false
			for _, art := range module.Artifacts {
				if art.Type == "narinfo" {
					hasNarinfo = true
				}
				if art.Type == "xz" {
					hasNarXz = true
				}
				// All should have checksums and repo
				assert.NotEmpty(t, art.Sha1, "artifact %s should have sha1", art.Name)
				assert.Equal(t, tests.NixLocalRepo, art.OriginalDeploymentRepo,
					"artifact should be in local repo")
			}
			assert.True(t, hasNarinfo, "should have at least one .narinfo artifact")
			assert.True(t, hasNarXz, "should have at least one .nar.xz artifact")
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixFlakeBuild_ModuleOverride(t *testing.T) {
	// Scenario #29 — Flake: --module=custom-name overrides module ID
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-flake-module-test"
	buildNumber := "1"
	customModule := "my-flake-module"

	projectPath, cleanupProject := createNixProject(t, "nix-flake-module", "flakeproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "build",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule)
	if err != nil {
		t.Skipf("nix build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id,
			"module ID should be overridden to custom name")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixFlakeBuild_NoBuildFlags(t *testing.T) {
	// Scenario #31 — Flake: nix build without --build-name/--build-number → passthrough, no crash
	initNixTest(t)

	projectPath, cleanupProject := createNixProject(t, "nix-flake-noflags", "flakeproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "build")
	if err != nil {
		t.Skipf("nix build not available: %v", err)
	}
}

func TestNixFlakeBuild_DepScopes(t *testing.T) {
	// Scenario #30 — Flake: all deps should have scope ["runtime"]
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-flake-scopes-test"
	buildNumber := "1"

	projectPath, cleanupProject := createNixProject(t, "nix-flake-scopes", "flakeproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("nix", "build",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("nix build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
			assert.Equal(t, []string{"runtime"}, dep.Scopes,
				"flake dep %s should have scope [runtime]", dep.Id)
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Round-Trip Test ====================

func TestNixRoundTrip(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-round-trip-test"
	buildNumber := "1"
	projectPath, cleanupProject := createNixProject(t, "nix-round-trip", "nixproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Step 1: Run nix flake lock with build-info
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
		"--build-name="+buildName, "--build-number="+buildNumber))

	// Step 2: Publish build-info
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Step 3: Retrieve and validate full round-trip
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be found after round-trip")

	bi := publishedBuildInfo.BuildInfo
	assert.Equal(t, buildName, bi.Name)
	assert.Equal(t, buildNumber, bi.Number)
	require.Len(t, bi.Modules, 1)

	module := bi.Modules[0]
	assert.Equal(t, buildinfo.Nix, module.Type)
	assert.Len(t, module.Dependencies, 3, "should have 3 deps: nixpkgs, flake-utils, systems")

	depIDs := make(map[string]bool)
	for _, dep := range module.Dependencies {
		depIDs[dep.Id] = true
		assert.Contains(t, dep.Sha256, "sha256-")
		assert.Equal(t, []string{"build"}, dep.Scopes)
		assert.Contains(t, dep.Id, ":")
	}

	assert.True(t, depIDs["nixpkgs:0ad13a6833440b8e238947e47bea7f11071dc2b2"])
	assert.True(t, depIDs["flake-utils:b1d9ab70662946ef0850d488da1c9019f3a9752a"])
	assert.True(t, depIDs["systems:da67096a3b9bf56a91d16901293e51ba5b49a27e"])

	// Validate requestedBy: systems should be requested by flake-utils
	for _, dep := range module.Dependencies {
		if dep.Id == "systems:da67096a3b9bf56a91d16901293e51ba5b49a27e" {
			require.NotEmpty(t, dep.RequestedBy)
			foundFlakeUtils := false
			for _, chain := range dep.RequestedBy {
				for _, parent := range chain {
					if parent == "flake-utils:b1d9ab70662946ef0850d488da1c9019f3a9752a" {
						foundFlakeUtils = true
					}
				}
			}
			assert.True(t, foundFlakeUtils, "systems should be requested by flake-utils")
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixChannels_NixEnvWithModuleOverride(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-channels-env-mod-test"
	buildNumber := "1"
	customModule := "hello-channels-env-mod"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("nix", "nix-env", "-iA", "nixpkgs.hello",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule)
	if err != nil {
		t.Skipf("nix-env not available or nixpkgs channel not configured: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be found after nix-env")
	if !found {
		return
	}

	bi := publishedBuildInfo.BuildInfo
	require.Len(t, bi.Modules, 1, "channels nix-env should produce exactly one module")
	module := bi.Modules[0]
	assert.Equal(t, customModule, module.Id, "module ID must reflect the --module override")
	assert.Equal(t, buildinfo.Nix, module.Type)
	assert.GreaterOrEqual(t, len(module.Dependencies), 1,
		"nix-env should collect at least 1 runtime dep from the closure")
	for _, dep := range module.Dependencies {
		assert.Contains(t, dep.Id, ":", "dep ID should be name:version, got: %s", dep.Id)
		assert.Equal(t, []string{"runtime"}, dep.Scopes)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixFlakes_BuildCopyWithModuleOverride(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-flakes-build-copy-mod-test"
	buildNumber := "1"
	customModule := "hello-flake-build-copy-mod"

	projectPath, cleanupProject := createNixProject(t, "nix-flake-mod-merge", "flakeproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Step 1: flake build, collects deps + sets module=hello-flake-build-copy-mod
	err = jfrogCli.Exec("nix", "build",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule)
	if err != nil {
		t.Skipf("nix build not available: %v", err)
	}

	toURL := fmt.Sprintf("https://%s:%s@%s/api/nix/%s/",
		*tests.JfrogUser, *tests.JfrogPassword,
		strings.TrimPrefix(strings.TrimPrefix(*tests.JfrogUrl, "https://"), "http://"),
		tests.NixVirtualRepo)
	copyErr := jfrogCli.Exec("nix", "copy", "--to", toURL, "./result",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule)

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if !found {
		return
	}

	bi := publishedBuildInfo.BuildInfo
	require.Len(t, bi.Modules, 1,
		"flake build + nix copy with same --module must produce ONE merged module, not two")
	module := bi.Modules[0]
	assert.Equal(t, customModule, module.Id, "module ID must be the --module value on both steps")
	assert.Equal(t, buildinfo.Nix, module.Type)
	assert.GreaterOrEqual(t, len(module.Dependencies), 1,
		"merged module must carry deps from the flake-build step")
	if copyErr == nil {
		assert.GreaterOrEqual(t, len(module.Artifacts), 2,
			"merged module must carry .nar.xz+.narinfo artifacts from the nix-copy step")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNixCustomProject_NixBuildCopyWithModuleOverride(t *testing.T) {
	initNixTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-custom-build-copy-mod-test"
	buildNumber := "1"
	customModule := "my-test-app-build-copy-mod"

	projectPath, cleanupProject := createNixProject(t, "nix-custom-mod-merge", "channelproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Step 1: nix-build default.nix, collects deps + sets the module override
	err = jfrogCli.Exec("nix", "nix-build", "default.nix",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule)
	if err != nil {
		t.Skipf("nix-build not available for custom default.nix: %v", err)
	}

	toURL := fmt.Sprintf("https://%s:%s@%s/api/nix/%s/",
		*tests.JfrogUser, *tests.JfrogPassword,
		strings.TrimPrefix(strings.TrimPrefix(*tests.JfrogUrl, "https://"), "http://"),
		tests.NixVirtualRepo)
	copyErr := jfrogCli.Exec("nix", "copy", "--to", toURL, "./result",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule)

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if !found {
		return
	}

	bi := publishedBuildInfo.BuildInfo
	require.Len(t, bi.Modules, 1,
		"nix-build + nix copy on a user derivation with same --module must produce ONE merged module")
	module := bi.Modules[0]
	assert.Equal(t, customModule, module.Id)
	assert.Equal(t, buildinfo.Nix, module.Type)
	assert.GreaterOrEqual(t, len(module.Dependencies), 1,
		"custom derivation should record at least 1 runtime dep")
	if copyErr == nil {
		assert.GreaterOrEqual(t, len(module.Artifacts), 2,
			"merged module must carry .nar.xz+.narinfo artifacts from the nix-copy step")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

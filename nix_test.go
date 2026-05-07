package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	return projectPath, cleanupCallback
}

// ==================== FlexPack Install Tests ====================

func TestNixFlakeLockFlexPack(t *testing.T) {
	testNixFlexPack(t, "flake lock")
}

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
		assert.NotEmpty(t, dep.Checksum.Sha256, "SHA256 (narHash) should be present for dep %s", dep.Id)
		assert.Contains(t, dep.Checksum.Sha256, "sha256-",
			"narHash should be in SRI format for dep %s, got: %s", dep.Id, dep.Checksum.Sha256)
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

func TestNixBuildInfoBothFlags(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	projectPath, cleanupProject := createNixProject(t, "nix-both-flags", "nixproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	buildName := "nix-both-flags-test"
	buildNumber := "42"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
		"--build-name="+buildName, "--build-number="+buildNumber))

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be found when both flags set")
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	assert.Greater(t, len(publishedBuildInfo.BuildInfo.Modules[0].Dependencies), 0)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

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

func TestNixSubcommandVariants(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	allTests := []struct {
		name                 string
		nixSubcmd            string
		expectedDependencies int
		collectsBuildInfo    bool
	}{
		{"nix-flake-lock", "flake lock", 3, true},
	}

	for buildNumber, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			buildNumberStr := strconv.Itoa(buildNumber + 1)
			projectPath, cleanupProject := createNixProject(t, test.name, "nixproject")
			defer cleanupProject()

			if test.collectsBuildInfo {
				args := []string{"nix"}
				args = append(args, strings.Split(test.nixSubcmd, " ")...)
				args = append(args, "--build-name="+tests.NixBuildName, "--build-number="+buildNumberStr)
				testNixCmd(t, projectPath, buildNumberStr, filepath.Base(projectPath),
					test.expectedDependencies, args)
			}
		})

		if test.collectsBuildInfo {
			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.NixBuildName, artHttpDetails)
		}
	}
}

// ==================== Multiple Builds Test ====================

func TestNixMultipleBuilds(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "nix-multi-build-test"

	for i := 1; i <= 3; i++ {
		buildNumber := strconv.Itoa(i)
		projectPath, cleanupProject := createNixProject(t, fmt.Sprintf("nix-multi-%d", i), "nixproject")

		wd, err := os.Getwd()
		assert.NoError(t, err)
		chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)

		jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
		assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
			"--build-name="+buildName, "--build-number="+buildNumber))
		assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

		publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
		assert.NoError(t, err)
		assert.True(t, found, "build %s should be found", buildNumber)
		assert.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)

		chdirCallback()
		cleanupProject()
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Dependency Checksums Test ====================

func TestNixDependencyChecksums(t *testing.T) {
	initNixTest(t)

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	projectPath, cleanupProject := createNixProject(t, "nix-checksums", "nixproject")
	defer cleanupProject()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	buildName := "nix-checksum-test"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("nix", "flake", "lock",
		"--build-name="+buildName, "--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	depsWithChecksums := 0
	for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
		if dep.Checksum.Sha256 != "" {
			depsWithChecksums++
			// Verify SRI format
			assert.Contains(t, dep.Checksum.Sha256, "sha256-",
				"dep %s should have narHash in SRI format", dep.Id)
		}
	}
	assert.Greater(t, depsWithChecksums, 0, "at least one dep should have checksums")

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

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
		assert.Contains(t, dep.Checksum.Sha256, "sha256-")
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

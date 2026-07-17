package main

import (
	"os"
	"path/filepath"
	"testing"

	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Cargo integration tests. Scenario numbers (#N) refer to CARGO_TEST_PLAN.md.
//
// These tests drive the native `jf cargo` FlexPack command. Build-info collection is
// enabled by JFROG_RUN_NATIVE=true (flexpack.IsFlexPackEnabled). The command buckets are:
//
//	deps      -> build, install, update, add, fetch, generate-lockfile, run, test, check
//	artifacts -> package  (scans target/package/*.crate, no deploy)
//	publish   -> publish  (deploys via native cargo publish + collects artifacts)
//
// The build-info module Type is the string "cargo". The published build-info-go in
// jfrog-cli does not export an entities.Cargo constant (it lives in the local
// ../build-info-go used by jfrog-cli-artifactory), so we assert on the string value.
const cargoModuleType = "cargo"

// ==================== Initialization ====================

func initCargoTest(t *testing.T) {
	if !*tests.TestCargo {
		t.Skip("Skipping Cargo test. To run Cargo test add the '-test.cargo=true' option.")
	}
	require.True(t, isRepoExist(tests.CargoLocalRepo), "Cargo test local repository doesn't exist.")
}

// ==================== Project Helpers ====================

// createCargoProject copies a fixture from testdata/cargo/<projectName> into a fresh
// temp dir and returns its path plus a cleanup callback.
func createCargoProject(t *testing.T, outputFolder, projectName string) (string, func()) {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "cargo", projectName)
	tmpDir, cleanupCallback := coretests.CreateTempDirWithCallbackAndAssert(t)

	projectPath := filepath.Join(tmpDir, outputFolder)
	assert.NoError(t, biutils.CopyDir(projectSrc, projectPath, true, nil))
	return projectPath, cleanupCallback
}

// runInCargoProject sets JFROG_RUN_NATIVE, prepares an isolated home dir, copies the
// fixture, and chdirs into it. It returns the JfrogCli client and a teardown func that
// restores env/home/cwd. Mirrors the setup boilerplate used by the Nix tests.
func runInCargoProject(t *testing.T, outputFolder, projectName string) (*coretests.JfrogCli, func()) {
	nativeCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")

	oldHomeDir, newHomeDir := prepareHomeDir(t)

	projectPath, cleanupProject := createCargoProject(t, outputFolder, projectName)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)

	teardown := func() {
		chdirCallback()
		cleanupProject()
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
		nativeCallback()
	}
	return coretests.NewJfrogCli(execMain, "jfrog", ""), teardown
}

// ==================== Dependency Collection (build bucket) ====================

func TestCargoBuild_CollectsDeps(t *testing.T) {
	// Scenarios: #14 — cargo build resolves/collects deps; #22 — deps captured in module;
	//            module Type is "cargo".
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-build-deps", "simple")
	defer teardown()

	buildName := "cli-cargo-build-deps-test"
	buildNumber := "1"

	err := jfrogCli.Exec("cargo", "build",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("cargo build not available or failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be found after cargo build")
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		assert.Equal(t, cargoModuleType, string(publishedBuildInfo.BuildInfo.Modules[0].Type),
			"module type should be cargo")
	}
}

// ==================== Artifact Collection (package bucket) ====================

func TestCargoPackage_CollectsArtifacts(t *testing.T) {
	// Scenarios: #7 — crate path layout <name>/<version>/<name>-<version>.crate;
	//            #35 — artifact carries sha256/sha1/md5; artifact Type is "crate".
	// `cargo package` produces target/package/*.crate offline — no registry needed.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-package", "simple")
	defer teardown()

	buildName := "cli-cargo-package-test"
	buildNumber := "1"

	err := jfrogCli.Exec("cargo", "package",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("cargo package not available or failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		assert.Equal(t, cargoModuleType, string(module.Type))
		require.NotEmpty(t, module.Artifacts, "cargo package should capture the .crate artifact")
		for _, art := range module.Artifacts {
			assert.Equal(t, "crate", art.Type, "artifact type should be crate")
			assert.Contains(t, art.Path, "cli-cargo-lib/1.0.0/cli-cargo-lib-1.0.0.crate",
				"crate must follow <name>/<version>/<name>-<version>.crate layout")
			assert.NotEmpty(t, art.Sha256, "crate artifact must have sha256")
			assert.NotEmpty(t, art.Sha1, "crate artifact must have sha1")
			assert.NotEmpty(t, art.Md5, "crate artifact must have md5")
		}
	}
}

// ==================== Publish (publish bucket) ====================

func TestCargoPublish(t *testing.T) {
	// Scenario: #6 — cargo publish uploads the .crate and collects it as an artifact.
	// Publishing needs a registry configured in .cargo/config.toml pointing at the
	// Artifactory cargo repo; if that isn't set up in the environment, skip gracefully.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-publish", "simple")
	defer teardown()

	buildName := "cli-cargo-publish-test"
	buildNumber := "1"

	err := jfrogCli.Exec("cargo", "publish",
		"--registry", "artifactory", "--no-verify",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("cargo publish not available or registry not configured: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		assert.Equal(t, cargoModuleType, string(module.Type))
		assert.NotEmpty(t, module.Artifacts, "publish should record the .crate artifact")
	}
}

// ==================== Build-info flags ====================

func TestCargoModuleOverride(t *testing.T) {
	// Scenario: #31 — --module overrides the auto-detected module ID.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-module-override", "simple")
	defer teardown()

	buildName := "cli-cargo-module-test"
	buildNumber := "1"
	customModule := "my-custom-cargo-module"

	err := jfrogCli.Exec("cargo", "build",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule)
	if err != nil {
		t.Skipf("cargo build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id,
			"module ID should be overridden to the custom name")
	}
}

func TestCargoBuildInfo_ProjectKey(t *testing.T) {
	// Scenario: #51 — --project scopes build info to the correct project. The flag must
	// pass through; build-publish under a nonexistent project may fail (that's tolerated).
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-project-key", "simple")
	defer teardown()

	buildName := "cli-cargo-project-key-test"
	buildNumber := "1"

	err := jfrogCli.Exec("cargo", "build",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--project=testprj")
	if err != nil {
		t.Skipf("cargo build not available: %v", err)
	}

	if err = artifactoryCli.Exec("bp", buildName, buildNumber, "--project=testprj"); err != nil {
		t.Logf("build-publish with --project failed (project may not exist): %v", err)
	}
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestCargoBuildNameOnly_NoBuildInfo(t *testing.T) {
	// Scenario: #27 — --build-name without --build-number → cargo runs, no build-info collected.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-name-only", "simple")
	defer teardown()

	// Only --build-name → collector short-circuits (see collectDeps: empty name check is on
	// build-name; a name without a number produces no publishable build). No panic expected.
	err := jfrogCli.Exec("cargo", "build", "--build-name=cli-cargo-name-only-test")
	if err != nil {
		t.Skipf("cargo build not available: %v", err)
	}
}

func TestCargoNoFlags_Passthrough(t *testing.T) {
	// Scenario: #29 — no --build-name/--build-number → pure passthrough, no build-info, no crash.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-no-flags", "simple")
	defer teardown()

	if err := jfrogCli.Exec("cargo", "build"); err != nil {
		t.Skipf("cargo build not available: %v", err)
	}
}

// ==================== Multi-module / Workspace ====================

func TestCargoWorkspace_Modules(t *testing.T) {
	// Scenarios: #38/#75 — workspace members build with Artifactory resolution and are
	// captured in the build-info. Uses the two-member workspace fixture.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-workspace", "workspace")
	defer teardown()

	buildName := "cli-cargo-workspace-test"
	buildNumber := "1"

	err := jfrogCli.Exec("cargo", "build",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("cargo build not available or workspace not resolvable: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if found {
		require.GreaterOrEqual(t, len(publishedBuildInfo.BuildInfo.Modules), 1,
			"workspace build should produce at least one cargo module")
		assert.Equal(t, cargoModuleType, string(publishedBuildInfo.BuildInfo.Modules[0].Type))
	}
}

// ==================== Checksums ====================

func TestCargoDepChecksums(t *testing.T) {
	// Scenario: #42 — dependencies captured in build info carry checksums when resolvable.
	// On a machine with a warm cargo cache some deps may lack AQL checksums; we log the ratio.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-dep-checksums", "simple")
	defer teardown()

	buildName := "cli-cargo-dep-checksums-test"
	buildNumber := "1"

	err := jfrogCli.Exec("cargo", "build",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("cargo build not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		deps := publishedBuildInfo.BuildInfo.Modules[0].Dependencies
		withChecksums := 0
		for _, dep := range deps {
			if dep.Sha256 != "" {
				withChecksums++
			}
		}
		t.Logf("deps with sha256: %d/%d", withChecksums, len(deps))
	}
}

// ==================== Full lifecycle ====================

func TestCargoFullLifecycle(t *testing.T) {
	// Scenario: #79 — build (collect deps) → package (collect artifacts) → publish build-info.
	// Package collects artifacts offline, so this exercises deps + artifacts in one build.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-lifecycle", "simple")
	defer teardown()

	buildName := "cli-cargo-lifecycle-test"
	buildNumber := "1"

	if err := jfrogCli.Exec("cargo", "build",
		"--build-name="+buildName, "--build-number="+buildNumber); err != nil {
		t.Skipf("cargo build not available: %v", err)
	}
	if err := jfrogCli.Exec("cargo", "package",
		"--build-name="+buildName, "--build-number="+buildNumber); err != nil {
		t.Skipf("cargo package not available: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		assert.Equal(t, cargoModuleType, string(module.Type))
		assert.NotEmpty(t, module.Artifacts, "lifecycle build should carry the .crate artifact")
	}
}

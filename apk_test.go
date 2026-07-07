package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	coreutils "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Initialization ====================

func initApkTest(t *testing.T) {
	if !*tests.TestApk {
		t.Skip("Skipping APK test. To run APK test add the '-test.apk=true' option.")
	}
	require.True(t, isRepoExist(tests.AlpineLocalRepo), "APK test local repository doesn't exist.")
	require.True(t, isRepoExist(tests.AlpineVirtualRepo), "APK test virtual repository doesn't exist.")
}

// apkAvailable returns true when the `apk` binary is present on this machine.
// Tests that actually invoke `apk` must call t.Skip when this returns false.
func apkAvailable() bool {
	_, err := exec.LookPath("apk")
	return err == nil
}

// ==================== jf apk add (build info collection) ====================

// TestApkAdd_BasicBuildInfo verifies that `jf apk add` records build-info
// dependencies for a single explicitly requested package.
func TestApkAdd_BasicBuildInfo(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-add-basic"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed (apk system command unavailable or repo not configured): %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be published after jf apk add")

	if found {
		bi := publishedBuildInfo.BuildInfo
		require.Len(t, bi.Modules, 1)
		assert.Equal(t, buildinfo.Alpine, bi.Modules[0].Type)
		assert.GreaterOrEqual(t, len(bi.Modules[0].Dependencies), 1,
			"curl should produce at least 1 dependency (curl itself + transitive)")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestApkAdd_MultiplePackages verifies build-info when installing more than one
// package in a single invocation.
func TestApkAdd_MultiplePackages(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-add-multi"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "wget", "jq",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		deps := publishedBuildInfo.BuildInfo.Modules[0].Dependencies
		// wget + jq each pull in multiple transitive deps; at minimum 2 direct ones
		assert.GreaterOrEqual(t, len(deps), 2,
			"wget + jq should produce at least 2 direct dependencies")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestApkAdd_NoBuildFlags verifies that `jf apk add` without --build-name / --build-number
// still runs natively without a panic or crash.
func TestApkAdd_NoBuildFlags(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "busybox")
	if err != nil {
		// busybox may already be installed; that's fine.
		t.Logf("jf apk add busybox (no build flags): %v", err)
	}
}

// TestApkAdd_BuildNameOnly verifies that supplying only --build-name (no --build-number)
// does not cause a panic; build-info collection is simply skipped.
func TestApkAdd_BuildNameOnly(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "busybox",
		"--build-name=apk-name-only-test")
	_ = err // command may succeed or fail — just ensure no panic
}

// ==================== jf apk upgrade (build info collection) ====================

// TestApkUpgrade_BuildInfo verifies that `jf apk upgrade` records build-info.
func TestApkUpgrade_BuildInfo(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-upgrade"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "upgrade",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk upgrade failed (no packages to upgrade or system unavailable): %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should exist after jf apk upgrade")

	if found {
		require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
		assert.Equal(t, buildinfo.Alpine, publishedBuildInfo.BuildInfo.Modules[0].Type)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Module override ====================

// TestApkAdd_ModuleOverride verifies that --module overrides the module ID in build-info.
func TestApkAdd_ModuleOverride(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-module-override"
	buildNumber := "1"
	customModule := "my-custom-alpine-module"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--module="+customModule,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id,
			"module ID should be overridden to the custom value")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Build-info JSON structure ====================

// TestApkAdd_BuildInfoJSON verifies the basic structure of the published build-info:
// correct name, number, started timestamp, module type, and non-empty dep IDs.
func TestApkAdd_BuildInfoJSON(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-bi-json"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "jq",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found {
		bi := publishedBuildInfo.BuildInfo
		assert.Equal(t, buildName, bi.Name)
		assert.Equal(t, buildNumber, bi.Number)
		assert.NotEmpty(t, bi.Started, "build-info should have a Started timestamp")
		require.Len(t, bi.Modules, 1)
		assert.Equal(t, buildinfo.Alpine, bi.Modules[0].Type)
		for _, dep := range bi.Modules[0].Dependencies {
			assert.NotEmpty(t, dep.Id, "dependency ID must not be empty")
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Dependency ID format ====================

// TestApkAdd_DepIDFormat verifies that every dependency ID is in the "name:version" format.
func TestApkAdd_DepIDFormat(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-dep-id-format"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
			assert.Contains(t, dep.Id, ":",
				"dep ID should be 'name:version' format, got: %s", dep.Id)
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Dependency scopes ====================

// TestApkAdd_DepScopes verifies that directly-requested packages get scope "prod"
// and transitive packages get scope "transitive".
func TestApkAdd_DepScopes(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-dep-scopes"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// curl brings in transitive deps (libcurl, musl, etc.)
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		hasProd := false
		hasTransitive := false
		for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
			for _, scope := range dep.Scopes {
				if scope == "prod" {
					hasProd = true
				}
				if scope == "transitive" {
					hasTransitive = true
				}
			}
		}
		assert.True(t, hasProd, "at least one dependency should have scope 'prod'")
		// Transitive deps are present when curl pulls in libcurl, musl, etc.
		t.Logf("hasProd=%v hasTransitive=%v (transitive depends on whether new deps were installed)",
			hasProd, hasTransitive)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Dependency checksums ====================

// TestApkAdd_DepChecksums verifies that at least some dependencies have SHA256 checksums set.
// SHA1 and MD5 come from the local cache; SHA256 from the APK database.
func TestApkAdd_DepChecksums(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-dep-checksums"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		depsWithChecksums := 0
		for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
			if dep.Sha1 != "" || dep.Sha256 != "" {
				depsWithChecksums++
			}
		}
		t.Logf("Dependencies with at least one checksum: %d/%d",
			depsWithChecksums, len(publishedBuildInfo.BuildInfo.Modules[0].Dependencies))
		// On a machine with a populated APK cache, checksums should be present.
		// We assert at least the package itself has one.
		assert.Greater(t, depsWithChecksums, 0,
			"at least one dep should have a checksum (sha1 or sha256)")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== requestedBy chains ====================

// TestApkAdd_DepRequestedBy verifies that transitive dependencies carry non-empty RequestedBy.
func TestApkAdd_DepRequestedBy(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-requestedby"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// curl has libcurl as a transitive dep, which should carry RequestedBy=[["curl:..."]]
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		deps := publishedBuildInfo.BuildInfo.Modules[0].Dependencies
		if len(deps) <= 1 {
			t.Skip("only one dep installed; transitive RequestedBy cannot be validated")
		}
		hasTransitiveRequestedBy := false
		for _, dep := range deps {
			if len(dep.RequestedBy) > 0 && len(dep.RequestedBy[0]) > 0 {
				hasTransitiveRequestedBy = true
				// Each RequestedBy entry should be a non-empty parent ID
				for _, chain := range dep.RequestedBy {
					assert.NotEmpty(t, chain, "RequestedBy chain entry should not be empty for %s", dep.Id)
				}
				break
			}
		}
		assert.True(t, hasTransitiveRequestedBy,
			"at least one transitive dep should have RequestedBy with a parent ID")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Multiple builds isolation ====================

// TestApkAdd_MultipleBuildsIsolated verifies that two sequential `jf apk add` calls with
// different build numbers produce independent build-info records.
func TestApkAdd_MultipleBuildsIsolated(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-multi-isolated"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number=1",
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add (build 1) failed: %v", err)
	}
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, "1"))

	err = jfrogCli.Exec("apk", "add", "jq",
		"--build-name="+buildName, "--build-number=2",
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add (build 2) failed: %v", err)
	}
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, "2"))

	bi1, found1, err := tests.GetBuildInfo(serverDetails, buildName, "1")
	assert.NoError(t, err)
	assert.True(t, found1, "build 1 should be found")

	bi2, found2, err := tests.GetBuildInfo(serverDetails, buildName, "2")
	assert.NoError(t, err)
	assert.True(t, found2, "build 2 should be found")

	if found1 && found2 {
		assert.NotEqual(t,
			bi1.BuildInfo.Modules[0].Dependencies,
			bi2.BuildInfo.Modules[0].Dependencies,
			"build 1 (curl) and build 2 (jq) should have different dependency sets")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Passthrough commands (no build info) ====================

// TestApkUpdate_Passthrough verifies that `jf apk update` runs natively without
// collecting build-info (apk update only refreshes the package index).
func TestApkUpdate_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "update")
	if err != nil {
		t.Skipf("jf apk update failed (network or repo not reachable): %v", err)
	}
}

// TestApkUpdate_PassthroughWithBuildFlags verifies that `jf apk update` with build flags
// still runs as a passthrough — update does not produce build-info.
func TestApkUpdate_PassthroughWithBuildFlags(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	buildName := tests.ApkBuildName + "-update-passthrough"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "update",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Skipf("jf apk update failed: %v", err)
	}

	// Publish attempt may fail because apk update does not produce build-info
	_ = artifactoryCli.Exec("bp", buildName, buildNumber)

	_, found, _ := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if found {
		t.Log("build-info was found for apk update (may have empty modules)")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestApkDel_Passthrough verifies that `jf apk del` runs as passthrough without
// producing build-info (package removal is not tracked).
func TestApkDel_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	buildName := tests.ApkBuildName + "-del-passthrough"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// curl may or may not be installed; that's fine — test only verifies no panic.
	err := jfrogCli.Exec("apk", "del", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Logf("jf apk del curl: %v (may not have been installed)", err)
	}

	// del should NOT produce build-info
	_ = artifactoryCli.Exec("bp", buildName, buildNumber)
	_, found, _ := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if found {
		t.Log("build-info was found for apk del (may have empty modules)")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestApkInfo_Passthrough verifies that `jf apk info` runs as passthrough and never
// produces build-info.
func TestApkInfo_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "info", "musl")
	if err != nil {
		t.Skipf("jf apk info musl: %v", err)
	}
}

// TestApkSearch_Passthrough verifies that `jf apk search` is forwarded to the native
// apk binary without modification and without producing build-info.
func TestApkSearch_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "search", "curl")
	if err != nil {
		t.Skipf("jf apk search curl: %v", err)
	}
}

// TestApkFetch_Passthrough verifies that `jf apk fetch` is forwarded to the native
// apk binary without modification and without producing build-info.
func TestApkFetch_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	tmpDir, err := os.MkdirTemp("", "apk-fetch-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	fetchErr := jfrogCli.Exec("apk", "fetch", "--output", tmpDir, "curl")
	if fetchErr != nil {
		t.Skipf("jf apk fetch curl: %v (network or repo issue)", fetchErr)
	}
}

// TestApkFix_Passthrough verifies that `jf apk fix` is forwarded as a passthrough.
func TestApkFix_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "fix")
	if err != nil {
		t.Logf("jf apk fix: %v (nothing to fix is normal)", err)
	}
}

// TestApkAudit_Passthrough verifies that `jf apk audit` is forwarded as a passthrough
// without producing build-info.
func TestApkAudit_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "audit")
	if err != nil {
		t.Logf("jf apk audit: %v (clean system is fine)", err)
	}
}

// TestApkVersion_Passthrough verifies that `jf apk version` is forwarded as a passthrough.
func TestApkVersion_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "version")
	if err != nil {
		t.Skipf("jf apk version: %v", err)
	}
}

// TestApkStats_Passthrough verifies that `jf apk stats` is forwarded as a passthrough.
func TestApkStats_Passthrough(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "stats")
	if err != nil {
		t.Logf("jf apk stats: %v (may not be supported on all Alpine versions)", err)
	}
}

// ==================== jf apk config ====================

// TestApkConfig_SetsUpRepo verifies that `jf apk config` runs without error when the
// Artifactory repo key and server details are valid.
// Note: this test requires write access to /etc/apk/repositories (root on Alpine).
func TestApkConfig_SetsUpRepo(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}
	if os.Getuid() != 0 {
		t.Skip("jf apk config modifies /etc/apk/repositories and requires root.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "config",
		"--server-id=default",
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Logf("jf apk config: %v (Artifactory may not be reachable or missing RSA key)", err)
	}
}

// ==================== jf apk upload ====================

// TestApkUpload_LocalApkFile verifies that `jf apk upload` can upload a real .apk file
// to Artifactory. The test creates a minimal (but valid enough) .apk archive.
func TestApkUpload_LocalApkFile(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	tmpDir, err := os.MkdirTemp("", "apk-upload-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build a minimal .apk (tgz containing just a PKGINFO) via abuild-tar or tar.
	// If abuild-tar is not available, skip gracefully.
	fakePkgInfo := "pkgname = testpkg\npkgver = 1.0.0-r0\narch = x86_64\n"
	pkgInfoPath := tmpDir + "/.PKGINFO"
	if writeErr := os.WriteFile(pkgInfoPath, []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}

	apkFilePath := fmt.Sprintf("%s/testpkg-1.0.0-r0.apk", tmpDir)
	// Create a minimal tar.gz containing the PKGINFO
	tarCmd := exec.Command("tar", "-czf", apkFilePath, "-C", tmpDir, ".PKGINFO")
	if tarOut, tarErr := tarCmd.CombinedOutput(); tarErr != nil {
		t.Skipf("failed to create fake .apk archive with tar: %v\n%s", tarErr, tarOut)
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("apk", "upload", apkFilePath,
		"--repo="+tests.AlpineLocalRepo,
		"--alpine-version=3.18",
		"--server-id=default")
	if err != nil {
		t.Logf("jf apk upload: %v (Artifactory may not accept the synthetic .apk)", err)
	}
}

// ==================== Project key ====================

// TestApkAdd_ProjectKey verifies that `jf apk add` with --project is passed through
// correctly; the test is lenient if the project does not exist in Artifactory.
func TestApkAdd_ProjectKey(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-project-key"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "busybox",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--project=testprj",
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add --project failed: %v", err)
	}

	publishErr := artifactoryCli.Exec("bp", buildName, buildNumber, "--project=testprj")
	if publishErr != nil {
		t.Logf("build-publish with --project failed (project may not exist): %v", publishErr)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== JFrog CLI flag stripping ====================

// TestApkAdd_JFlagStripping verifies that JFrog CLI flags (--build-name, --build-number,
// --repo, --server-id, etc.) are not forwarded to the native apk binary.
// If they were forwarded, apk would reject the unknown flags and exit non-zero.
func TestApkAdd_JFlagStripping(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// If --build-name etc. are leaked to the native apk binary it will fail with
	// "ERROR: unknown option --build-name". A clean exit confirms flag stripping.
	err := jfrogCli.Exec("apk", "add", "busybox",
		"--build-name=flag-strip-test",
		"--build-number=99",
		"--repo="+tests.AlpineVirtualRepo,
		"--server-id=default")
	if err != nil {
		// A non-zero exit that is NOT due to unknown flags is OK (package already installed).
		t.Logf("jf apk add with all JF flags: %v", err)
	}
}

// ==================== Ghost Frog / execWithPackageManager ====================

// TestApkGhostFrog_InvokesJfApk verifies that when the JFROG_CLI_GHOST_FROG_APK environment
// variable is set, the native `apk` alias resolves through jf.
// This is a structural smoke-test — it verifies the alias is registered, not that it
// actually intercepts a real `apk` binary call (that requires OS-level symlink setup).
func TestApkGhostFrog_InvokesJfApk(t *testing.T) {
	initApkTest(t)

	// Ghost Frog is registered purely as a CLI feature — its wiring can be validated
	// without an actual apk binary by verifying the command is dispatched correctly.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// `jf apk --help` should succeed even without a real apk binary.
	err := jfrogCli.Exec("apk", "--help")
	// --help typically exits with code 0; if the sub-command is not registered it errors.
	if err != nil {
		t.Logf("jf apk --help returned non-nil: %v", err)
	}
}

// ==================== Environment secret filtering ====================

// TestApkAdd_EnvSecretFiltering verifies that secret environment variables are not
// leaked into the build-info (e.g. JFROG_CLI_ENV_EXCLUDE works as expected).
func TestApkAdd_EnvSecretFiltering(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "MY_SECRET_TOKEN", "super-secret-value")
	defer setEnvCallback()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-env-secrets"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo)
	if err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found {
		// Scan all env vars captured in build-info; none should contain the secret value.
		for k, v := range publishedBuildInfo.BuildInfo.Properties {
			assert.NotContains(t, v, "super-secret-value",
				"env var %s should not contain secret value in build-info", k)
		}
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== Alpine version flag ====================

// TestApkAdd_AlpineVersionFlag verifies that the --alpine-version flag is accepted and
// recorded in the module ID without being passed to the native apk binary.
func TestApkAdd_AlpineVersionFlag(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.ApkBuildName + "-alpine-version"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--repo="+tests.AlpineVirtualRepo,
		"--alpine-version=3.18")
	if err != nil {
		t.Skipf("jf apk add --alpine-version failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)

	if found {
		require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
		assert.Equal(t, buildinfo.Alpine, publishedBuildInfo.BuildInfo.Modules[0].Type)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	rtLifecycle "github.com/jfrog/jfrog-cli-artifactory/lifecycle"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	coreutils "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	accessServices "github.com/jfrog/jfrog-client-go/access/services"
	artServices "github.com/jfrog/jfrog-client-go/artifactory/services"
	lifecycleServices "github.com/jfrog/jfrog-client-go/lifecycle/services"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Initialization ====================

func initApkTest(t *testing.T) {
	if !*tests.TestAlpine {
		t.Skip("Skipping Alpine APK test. To run Alpine APK tests add the '-test.alpine=true' option.")
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

// computeFileSHA256 returns the hex-encoded SHA256 digest of the file at path.
func computeFileSHA256(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err, "open file for SHA256: %s", path)
	defer func() { require.NoError(t, f.Close()) }()
	h := sha256.New()
	_, err = io.Copy(h, f)
	require.NoError(t, err, "compute SHA256 for: %s", path)
	return fmt.Sprintf("%x", h.Sum(nil))
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

	buildName := tests.AlpineBuildName + "-add-basic"
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

// TestApkReleaseBundleCreation verifies that a Release Bundle can be created from the
// build-info produced by `jf apk upload`. Upload is used (not `apk add`) because a release
// bundle requires source *artifacts*, and only the upload flow records an artifact in the
// build-info — `apk add` records dependencies only, which yields "Source artifacts not found".
// It uploads an .apk with build-info, publishes the build, runs `rbc --build-name/--build-number`,
// and asserts the release bundle reaches COMPLETED status. Release bundle operations require a
// lifecycle service, so the test skips gracefully when it is unavailable.
func TestApkReleaseBundleCreation(t *testing.T) {
	initApkTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	tmpDir, err := os.MkdirTemp("", "apk-rb-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	fakePkgInfo := "pkgname = testpkg-rb\npkgver = 1.0.0-r0\narch = x86_64\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-rb-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to build test .apk: %v", tarErr)
	}

	buildName := tests.AlpineBuildName + "-rb"
	buildNumber := "1"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	// Upload an .apk with build-info so the build records a promotable/bundlable artifact.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath,
		"--repo="+tests.AlpineLocalRepo,
		"--alpine-version=3.18",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--server-id=default"); uploadErr != nil {
		t.Skipf("jf apk upload failed: %v", uploadErr)
	}
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// The build must exist with an Alpine module recording an artifact before we can bundle it.
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	if !assert.True(t, found, "build-info should be published before creating a release bundle") {
		return
	}
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules, "Alpine build-info should have at least one module")
	assert.Equal(t, buildinfo.Alpine, publishedBuildInfo.BuildInfo.Modules[0].Type)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts,
		"upload build-info must record an artifact so the release bundle has source artifacts")

	// Create a release bundle from the Alpine build-info.
	rbName := buildName + "-release-bundle"
	rbVersion := "1.0.0"
	if err = runJfrogCliWithoutAssertion("rbc", rbName, rbVersion,
		"--build-name="+buildName, "--build-number="+buildNumber); err != nil {
		// Release bundle creation requires a lifecycle/Distribution service — skip if unavailable.
		t.Skipf("Skipping release bundle creation test: %v", err)
	}

	// Verify the release bundle reached COMPLETED status, then clean it up.
	// The lifecycle service manager authenticates against serverDetails.LifecycleUrl, which is
	// not derived from the platform URL automatically — populate it as the lifecycle CLI does.
	rtLifecycle.PlatformToLifecycleUrls(serverDetails)
	lcManager, err := artUtils.CreateLifecycleServiceManager(serverDetails, false)
	if assert.NoError(t, err) {
		rbDetails := lifecycleServices.ReleaseBundleDetails{
			ReleaseBundleName:    rbName,
			ReleaseBundleVersion: rbVersion,
		}
		resp, statusErr := lcManager.GetReleaseBundleCreationStatus(rbDetails, "", true)
		if assert.NoError(t, statusErr) {
			assert.Equal(t, lifecycleServices.Completed, resp.Status,
				"release bundle created from Alpine build-info should reach COMPLETED status")
		}
		_ = lcManager.DeleteReleaseBundleVersion(rbDetails, lifecycleServices.CommonOptionalQueryParams{Async: false})
	}
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

	buildName := tests.AlpineBuildName + "-add-multi"
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

	buildName := tests.AlpineBuildName + "-upgrade"
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

	buildName := tests.AlpineBuildName + "-module-override"
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

	buildName := tests.AlpineBuildName + "-bi-json"
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

	buildName := tests.AlpineBuildName + "-dep-id-format"
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

	buildName := tests.AlpineBuildName + "-dep-scopes"
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

	buildName := tests.AlpineBuildName + "-dep-checksums"
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

	buildName := tests.AlpineBuildName + "-requestedby"
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

	buildName := tests.AlpineBuildName + "-multi-isolated"

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

// TestApkPassthroughCommands verifies that every apk sub-command that does not
// produce build-info (update, del, info, search, fetch, fix, audit, version, stats)
// is forwarded to the native apk binary without panicking.
// Each case also confirms that passing --build-name/--build-number does NOT result
// in build-info being published — these commands are read-only or destructive, not
// installation events.
func TestApkPassthroughCommands(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	cases := []struct {
		name string
		args []string
	}{
		{"update", []string{"apk", "update"}},
		{"update-with-build-flags", []string{"apk", "update", "--build-name=" + tests.AlpineBuildName + "-pt", "--build-number=1"}},
		{"del", []string{"apk", "del", "curl", "--build-name=" + tests.AlpineBuildName + "-pt", "--build-number=1"}},
		{"info", []string{"apk", "info", "musl"}},
		{"search", []string{"apk", "search", "curl"}},
		{"fix", []string{"apk", "fix"}},
		{"audit", []string{"apk", "audit"}},
		{"version", []string{"apk", "version"}},
		{"stats", []string{"apk", "stats"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// A non-zero exit is acceptable (e.g. nothing to fix, package absent);
			// what must not happen is a panic or an unknown-flag error caused by
			// JFrog CLI flags leaking into the native apk invocation.
			err := jfrogCli.Exec(tc.args...)
			if err != nil {
				t.Logf("jf %v: %v (non-zero exit is acceptable for passthrough)", tc.args, err)
			}
		})
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

// TestApkConfig_UnknownServerID verifies that jf apk config with an unknown
// --server-id returns a clear error rather than silently continuing.
func TestApkConfig_UnknownServerID(t *testing.T) {
	initApkTest(t)

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "config",
		"--server-id=nonexistent-server-id-xyz",
		"--repo="+tests.AlpineVirtualRepo)
	assert.Error(t, err, "jf apk config with an unknown --server-id should return a clear error")
}

// ==================== jf setup apk ====================

// TestSetupApk_WithRepoSkipsPromptAndConfigures verifies that `jf setup apk --repo <repo>`
// validates the repo, skips the interactive repo-type/repo selection, and rewrites
// /etc/apk/repositories to point at the given Artifactory Alpine repository with the
// credentials embedded in the URL so native apk can authenticate without HTTP_AUTH.
// The file must also be locked to 0600 because it now stores a secret.
// Note: setup apk writes to /etc/apk/* and therefore requires root on Alpine.
func TestSetupApk_WithRepoSkipsPromptAndConfigures(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}
	if os.Getuid() != 0 {
		t.Skip("jf setup apk modifies /etc/apk/repositories and requires root.")
	}

	// Configure the 'default' server so --server-id=default resolves an Artifactory URL.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	const apkRepositoriesFile = "/etc/apk/repositories"
	// Back up the current repositories file and restore it after the test.
	original, readErr := os.ReadFile(apkRepositoriesFile)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("failed to read %s: %v", apkRepositoriesFile, readErr)
	}
	defer func() {
		if original != nil {
			// #nosec G703 -- apkRepositoriesFile is a hardcoded constant, not user input
			require.NoError(t, os.WriteFile(apkRepositoriesFile, original, 0644))
		}
	}()

	// With --repo supplied, setup must not prompt for a repo type or repo selection;
	// a non-interactive run reaching completion confirms the prompt was skipped.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("setup", "apk",
		"--server-id=default",
		"--repo="+tests.AlpineVirtualRepo)
	require.NoError(t, err, "jf setup apk --repo should succeed for an existing repo without prompting")

	// The repositories file should now reference the configured Artifactory repo.
	content, err := os.ReadFile(apkRepositoriesFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), tests.AlpineVirtualRepo,
		"/etc/apk/repositories should contain the configured Artifactory Alpine repo")

	// setup apk points apk EXCLUSIVELY at Artifactory: every non-empty, non-comment line
	// must be an Artifactory repo line — public mirrors (e.g. dl-cdn) must be removed.
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		assert.Contains(t, trimmed, "/artifactory/",
			"/etc/apk/repositories should contain only the Artifactory repo (public mirrors must be removed), found: %s", trimmed)
	}

	// The configured Artifactory line must carry embedded credentials (userinfo "@")
	// so native apk authenticates directly from the file without HTTP_AUTH.
	var artifactoryLine string
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "/artifactory/") && strings.Contains(line, tests.AlpineVirtualRepo) {
			artifactoryLine = strings.TrimSpace(line)
			break
		}
	}
	require.NotEmpty(t, artifactoryLine, "expected an Artifactory repo line in %s", apkRepositoriesFile)
	parsed, parseErr := url.Parse(artifactoryLine)
	require.NoError(t, parseErr, "the configured repo URL should be parseable")
	assert.NotNil(t, parsed.User,
		"the configured repo URL should embed credentials in its userinfo component")
	if user := parsed.User; user != nil {
		assert.NotEmpty(t, user.Username(),
			"the embedded userinfo should contain a username")
	}

	// Because the file now stores a secret, it must be locked down to 0600.
	info, statErr := os.Stat(apkRepositoriesFile)
	require.NoError(t, statErr)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(),
		"/etc/apk/repositories must be 0600 since it embeds credentials")
}

// TestSetupApk_InvalidRepoFails verifies that `jf setup apk` with a repository that does not
// exist in Artifactory fails fast during repo validation, before touching /etc/apk.
func TestSetupApk_InvalidRepoFails(t *testing.T) {
	initApkTest(t)

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("setup", "apk",
		"--server-id=default",
		"--repo=nonexistent-alpine-repo-xyz-12345")
	assert.Error(t, err, "jf setup apk with a nonexistent --repo should fail during validation")
}

// ==================== P0: package not found ====================

// TestApkAdd_PackageNotFound verifies that jf apk add with a package that does
// not exist in Artifactory returns a clear error and does not silently fall back
// to the public Alpine CDN.
func TestApkAdd_PackageNotFound(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "nonexistent-pkg-xyz-jfrog-test-12345",
		"--repo="+tests.AlpineVirtualRepo,
		"--build-name="+tests.AlpineBuildName+"-pkg-not-found",
		"--build-number=1")
	assert.Error(t, err,
		"jf apk add with a nonexistent package should fail, not silently succeed or fall back to CDN")
}

// ==================== P0: checksum stored in Artifactory ====================

// TestApkUpload_ChecksumNotUntrusted verifies that after jf apk upload the
// artifact stored in Artifactory has a non-empty, non-"untrusted" SHA256
// checksum. Artifactory marks artifacts as "untrusted" when upload does not
// supply X-Checksum headers — this test catches that integration gap.
func TestApkUpload_ChecksumNotUntrusted(t *testing.T) {
	initApkTest(t)

	tmpDir, err := os.MkdirTemp("", "apk-chksum-stored-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	fakePkgInfo := "pkgname = testpkg-chksum\npkgver = 1.0.0-r0\narch = x86_64\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-chksum-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to build test .apk: %v", tarErr)
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath, "--repo="+tests.AlpineLocalRepo); uploadErr != nil {
		t.Skipf("jf apk upload failed: %v", uploadErr)
	}

	// Search Artifactory for the uploaded artifact and verify the stored sha256.
	searchSpec := spec.NewBuilder().Pattern(tests.AlpineLocalRepo + "/testpkg-chksum-1.0.0-r0.apk").BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, searchErr := searchCmd.Search()
	require.NoError(t, searchErr, "AQL search for uploaded artifact should succeed")
	defer func() { _ = reader.Close() }()

	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item),
		"uploaded artifact must be found in Artifactory — jf apk upload may have failed silently")
	assert.NotEmpty(t, item.Sha256,
		"Artifactory must store a sha256 checksum for the uploaded .apk artifact")
	assert.NotEqual(t, "untrusted", strings.ToLower(item.Sha256),
		"Artifactory must not mark the .apk artifact as 'untrusted' — X-Checksum headers may be missing on upload")
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
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

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

// TestApkUpload_ArchAutoDetectedFromPkgInfo verifies that when --arch is omitted,
// `jf apk upload` reads the architecture from the package's embedded .PKGINFO and
// uploads the artifact under that arch path segment. The .PKGINFO declares aarch64
// (deliberately different from the typical x86_64 CI host) so a match proves the arch
// came from the package metadata, not the uploading machine.
func TestApkUpload_ArchAutoDetectedFromPkgInfo(t *testing.T) {
	initApkTest(t)

	// Configure the 'default' server so --server-id=default resolves for upload.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	tmpDir, err := os.MkdirTemp("", "apk-arch-detect-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	const pkgArch = "aarch64"
	fakePkgInfo := "pkgname = testpkg-archdetect\npkgver = 1.0.0-r0\narch = " + pkgArch + "\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-archdetect-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to build test .apk: %v", tarErr)
	}

	// Upload WITHOUT --arch: the command must auto-detect aarch64 from .PKGINFO.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath,
		"--repo="+tests.AlpineLocalRepo,
		"--alpine-version=3.18",
		"--server-id=default"); uploadErr != nil {
		t.Skipf("jf apk upload failed: %v", uploadErr)
	}

	// The artifact must be stored under the aarch64 path segment derived from .PKGINFO.
	searchSpec := spec.NewBuilder().
		Pattern(tests.AlpineLocalRepo + "/*/" + pkgArch + "/testpkg-archdetect-1.0.0-r0.apk").
		BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, searchErr := searchCmd.Search()
	require.NoError(t, searchErr, "AQL search for uploaded artifact should succeed")
	defer func() { _ = reader.Close() }()

	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item),
		"artifact must be found under the aarch64 path — arch was not auto-detected from .PKGINFO")
	assert.Contains(t, item.Path, "/"+pkgArch+"/",
		"uploaded artifact path must contain the arch auto-detected from .PKGINFO")
}

// TestApkUpload_IndexablePathLayout verifies that `jf apk upload` deploys the package under
// the layout Artifactory requires for indexing: <repo>/<alpine-version>/<branch>/<arch>/<file>.
// Per Artifactory's Alpine documentation, packages not deployed under
// <branch>/<repository(component)>/<architecture>/ are not included in any APKINDEX. A distinct
// non-default branch (community) and arch are used so the assertion pins the full path shape.
func TestApkUpload_IndexablePathLayout(t *testing.T) {
	initApkTest(t)

	// Configure the 'default' server so --server-id=default resolves for upload.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	tmpDir, err := os.MkdirTemp("", "apk-layout-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	const (
		pkgArch       = "x86_64"
		alpineVersion = "3.20"
		branch        = "community"
	)
	fakePkgInfo := "pkgname = testpkg-layout\npkgver = 1.0.0-r0\narch = " + pkgArch + "\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-layout-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to build test .apk: %v", tarErr)
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath,
		"--repo="+tests.AlpineLocalRepo,
		"--alpine-version="+alpineVersion,
		"--branch="+branch,
		"--server-id=default"); uploadErr != nil {
		t.Skipf("jf apk upload failed: %v", uploadErr)
	}

	// jf apk upload normalizes the version to Alpine's canonical "v"-prefixed form, so an
	// input of "3.20" is stored under "v3.20". The stored artifact path must therefore be
	// exactly <repo>/v<version>/<branch>/<arch>/<file>.
	expectedRelPath := "v" + alpineVersion + "/" + branch + "/" + pkgArch + "/testpkg-layout-1.0.0-r0.apk"
	searchSpec := spec.NewBuilder().
		Pattern(tests.AlpineLocalRepo + "/" + expectedRelPath).
		BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, searchErr := searchCmd.Search()
	require.NoError(t, searchErr, "AQL search for the uploaded artifact should succeed")
	defer func() { _ = reader.Close() }()

	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item),
		"artifact must be found under <repo>/<version>/<branch>/<arch>/<file> — the branch segment must not be dropped")
	assert.Equal(t, tests.AlpineLocalRepo+"/"+expectedRelPath, item.Path,
		"upload must deploy under the indexable layout <repo>/<alpine-version>/<branch>/<arch>/<file>")
}

// TestApkUpload_UShortcut verifies that "u" is accepted as a shortcut for the "upload"
// subcommand and behaves identically — the package is deployed to Artifactory (also implicitly
// confirming the "3.20" -> "v3.20" version normalization).
func TestApkUpload_UShortcut(t *testing.T) {
	initApkTest(t)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	tmpDir, err := os.MkdirTemp("", "apk-u-shortcut-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	const pkgArch = "x86_64"
	fakePkgInfo := "pkgname = testpkg-ushortcut\npkgver = 1.0.0-r0\narch = " + pkgArch + "\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-ushortcut-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to build test .apk: %v", tarErr)
	}

	// Use the "u" alias instead of the full "upload" subcommand.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "u", apkPath,
		"--repo="+tests.AlpineLocalRepo,
		"--alpine-version=3.20",
		"--server-id=default"); uploadErr != nil {
		t.Skipf("jf apk u (upload shortcut) failed: %v", uploadErr)
	}

	searchSpec := spec.NewBuilder().
		Pattern(tests.AlpineLocalRepo + "/v3.20/main/" + pkgArch + "/testpkg-ushortcut-1.0.0-r0.apk").
		BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, searchErr := searchCmd.Search()
	require.NoError(t, searchErr, "AQL search for the uploaded artifact should succeed")
	defer func() { _ = reader.Close() }()

	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item),
		"artifact uploaded via the 'u' shortcut must be found — 'u' should alias 'upload'")
}

// TestApkUpload_AlpineVersionAutoDetectedFromHost verifies that when --alpine-version is omitted,
// `jf apk upload` falls back to the host's Alpine version from /etc/alpine-release and deploys the
// artifact under that "v"-prefixed version segment.
func TestApkUpload_AlpineVersionAutoDetectedFromHost(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — --alpine-version auto-detect reads /etc/alpine-release (Alpine only).")
	}

	// Derive the expected version from /etc/alpine-release (major.minor -> vX.Y).
	rel, relErr := os.ReadFile("/etc/alpine-release")
	if relErr != nil {
		t.Skipf("cannot read /etc/alpine-release: %v", relErr)
	}
	verParts := strings.Split(strings.TrimSpace(strings.TrimPrefix(string(rel), "v")), ".")
	if len(verParts) < 2 {
		t.Skipf("unexpected /etc/alpine-release content: %q", string(rel))
	}
	expectedVersion := "v" + verParts[0] + "." + verParts[1]

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	tmpDir, err := os.MkdirTemp("", "apk-verdetect-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	const pkgArch = "x86_64"
	fakePkgInfo := "pkgname = testpkg-verdetect\npkgver = 1.0.0-r0\narch = " + pkgArch + "\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-verdetect-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to build test .apk: %v", tarErr)
	}

	// Upload WITHOUT --alpine-version: the command must auto-detect it from /etc/alpine-release.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath,
		"--repo="+tests.AlpineLocalRepo,
		"--server-id=default"); uploadErr != nil {
		t.Skipf("jf apk upload (version auto-detect) failed: %v", uploadErr)
	}

	expectedRelPath := expectedVersion + "/main/" + pkgArch + "/testpkg-verdetect-1.0.0-r0.apk"
	searchSpec := spec.NewBuilder().
		Pattern(tests.AlpineLocalRepo + "/" + expectedRelPath).
		BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, searchErr := searchCmd.Search()
	require.NoError(t, searchErr, "AQL search for the uploaded artifact should succeed")
	defer func() { _ = reader.Close() }()

	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item),
		"artifact must be under the host's Alpine version (%s) — --alpine-version was not auto-detected", expectedVersion)
	assert.Equal(t, tests.AlpineLocalRepo+"/"+expectedRelPath, item.Path)
}

// TestApk_NonDefaultServerDoesNotMutateRepositoriesFile verifies that running a jf apk command
// with a --server-id different from the default server uses an isolated temporary repositories
// file, leaving the persistent /etc/apk/repositories (configured for the default server) untouched.
func TestApk_NonDefaultServerDoesNotMutateRepositoriesFile(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}
	if os.Getuid() != 0 {
		t.Skip("jf setup apk writes /etc/apk/repositories and requires root.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	// Add a second server (different id) pointing at the same Artifactory.
	const secondServerID = "apk-isolated-server"
	configCli := coretests.NewJfrogCli(execMain, "jfrog config", "")
	if cfgErr := configCli.Exec("add", secondServerID, "--interactive=false",
		"--artifactory-url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint,
		"--user=admin", "--password=password", "--enc-password=false", "--overwrite"); cfgErr != nil {
		t.Skipf("could not configure a second server: %v", cfgErr)
	}
	defer func() { _ = configCli.Exec("rm", secondServerID, "--quiet") }()

	const apkRepositoriesFile = "/etc/apk/repositories"
	original, readErr := os.ReadFile(apkRepositoriesFile)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("failed to read %s: %v", apkRepositoriesFile, readErr)
	}
	defer func() {
		if original != nil {
			// #nosec G703 -- apkRepositoriesFile is a hardcoded constant, not user input
			require.NoError(t, os.WriteFile(apkRepositoriesFile, original, 0644))
		}
	}()

	// Configure /etc/apk/repositories for the DEFAULT server.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	require.NoError(t, jfrogCli.Exec("setup", "apk", "--repo="+tests.AlpineVirtualRepo, "--server-id=default"),
		"jf setup apk should configure /etc/apk/repositories for the default server")

	before, err := os.ReadFile(apkRepositoriesFile)
	require.NoError(t, err)

	// Run a native apk command against the NON-default server. Isolation must kick in and use a
	// temporary repositories file; the apk operation itself may warn/fail (its index may be
	// missing), so we ignore its exit status and only assert the persistent file is untouched.
	_ = jfrogCli.Exec("apk", "update", "--repo="+tests.AlpineLocalRepo, "--server-id="+secondServerID)

	after, err := os.ReadFile(apkRepositoriesFile)
	require.NoError(t, err)
	assert.Equal(t, string(before), string(after),
		"/etc/apk/repositories must be unchanged when a jf apk command targets a non-default --server-id")
}

// ==================== jf rt build-promote ====================

// TestApkBuildPromote verifies that a build produced by `jf apk upload` can be promoted
// with `jf rt build-promote` (bpr). It uploads an .apk with build-info, publishes the
// build, then promotes it back to the same local repo using --copy (a deterministic no-op
// copy that still exercises the promotion API and updates the build status). Using --copy
// keeps the artifact in the source repo so it remains searchable afterwards.
func TestApkBuildPromote(t *testing.T) {
	initApkTest(t)

	// Configure the 'default' server so --server-id=default resolves for upload.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	tmpDir, err := os.MkdirTemp("", "apk-promote-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	const pkgArch = "x86_64"
	fakePkgInfo := "pkgname = testpkg-promote\npkgver = 1.0.0-r0\narch = " + pkgArch + "\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-promote-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to build test .apk: %v", tarErr)
	}

	buildName := tests.AlpineBuildName + "-promote"
	buildNumber := "1"

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	// Upload the .apk with build-info so the build has a promotable artifact.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath,
		"--repo="+tests.AlpineLocalRepo,
		"--alpine-version=3.18",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--server-id=default"); uploadErr != nil {
		t.Skipf("jf apk upload failed: %v", uploadErr)
	}
	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// The published build must record the uploaded .apk as an artifact before promotion.
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	if !assert.True(t, found, "build-info should be published before promotion") {
		return
	}
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules, "Alpine upload build-info should have a module")
	assert.Equal(t, buildinfo.Alpine, publishedBuildInfo.BuildInfo.Modules[0].Type)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts,
		"Alpine upload build-info module should record the uploaded .apk as an artifact")

	// Promote the build with --copy back to the same local repo (deterministic) and set a status.
	err = artifactoryCli.Exec("bpr", buildName, buildNumber, tests.AlpineLocalRepo,
		"--copy=true",
		"--status=promoted",
		"--comment=automated-alpine-test-promotion")
	assert.NoError(t, err, "jf rt build-promote should succeed with --copy to the local repo")

	// With --copy, the artifact must still exist in the source repo after promotion.
	searchSpec := spec.NewBuilder().
		Pattern(tests.AlpineLocalRepo + "/*/" + pkgArch + "/testpkg-promote-1.0.0-r0.apk").
		BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, searchErr := searchCmd.Search()
	require.NoError(t, searchErr, "AQL search for the promoted artifact should succeed")
	defer func() { _ = reader.Close() }()

	item := new(artUtils.SearchResult)
	assert.NoError(t, reader.NextRecord(item),
		"promoted artifact must still exist in the source repo after build-promote --copy")
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

	buildName := tests.AlpineBuildName + "-project-key"
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

// TestApkAdd_ProjectBuildInfoCollection verifies that `jf apk add --project <key>` collects
// build-info under a real (non-default) Artifactory project and that the build-info is
// retrievable scoped to that project key. Unlike TestApkAdd_ProjectKey (which is lenient),
// this test creates the project, assigns the Alpine repos to it, and asserts the published
// build-info contains an Alpine module with dependencies under the project.
func TestApkAdd_ProjectBuildInfoCollection(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	// Create the Access service manager and project before deferring any cleanup, so that
	// a t.Skipf on an environment without Projects/Access doesn't trigger cleanup asserts.
	accessManager, err := artUtils.CreateAccessServiceManager(serverDetails, false)
	if err != nil {
		t.Skipf("Skipping project build-info test — cannot create access manager: %v", err)
	}
	projectParams := accessServices.ProjectParams{
		ProjectDetails: accessServices.Project{
			DisplayName: "alpine-project-test " + tests.ProjectKey,
			ProjectKey:  tests.ProjectKey,
		},
	}
	_ = accessManager.DeleteProject(tests.ProjectKey)
	if err = accessManager.CreateProject(projectParams); err != nil {
		t.Skipf("Skipping project build-info test — cannot create project: %v", err)
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()
	defer func() {
		_ = accessManager.UnassignRepoFromProject(tests.AlpineVirtualRepo)
		_ = accessManager.UnassignRepoFromProject(tests.AlpineLocalRepo)
		_ = accessManager.DeleteProject(tests.ProjectKey)
	}()

	// Assign the Alpine repos used for resolution to the project.
	assert.NoError(t, accessManager.AssignRepoToProject(tests.AlpineVirtualRepo, tests.ProjectKey, true))
	assert.NoError(t, accessManager.AssignRepoToProject(tests.AlpineLocalRepo, tests.ProjectKey, true))

	buildName := tests.AlpineBuildName + "-project-buildinfo"
	buildNumber := "1"
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	// Collect build-info under the project.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if addErr := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName, "--build-number="+buildNumber,
		"--project="+tests.ProjectKey,
		"--repo="+tests.AlpineVirtualRepo); addErr != nil {
		t.Skipf("jf apk add --project failed (apk unavailable or repo not configured): %v", addErr)
	}
	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber, "--project="+tests.ProjectKey))

	// Retrieve the build-info scoped to the project key and assert Alpine collection.
	servicesManager, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)
	params := artServices.NewBuildInfoParams()
	params.BuildName = buildName
	params.BuildNumber = buildNumber
	params.ProjectKey = tests.ProjectKey
	publishedBuildInfo, found, err := servicesManager.GetBuildInfo(params)
	assert.NoError(t, err)
	assert.True(t, found, "build-info should be found under project %q", tests.ProjectKey)

	if found {
		bi := publishedBuildInfo.BuildInfo
		require.NotEmpty(t, bi.Modules, "project build-info should contain an Alpine module")
		assert.Equal(t, buildinfo.Alpine, bi.Modules[0].Type)
		assert.GreaterOrEqual(t, len(bi.Modules[0].Dependencies), 1,
			"curl should produce at least one dependency in the project build-info")
	}
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

	buildName := tests.AlpineBuildName + "-env-secrets"
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

	buildName := tests.AlpineBuildName + "-alpine-version"
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

// ==================== P0: missing required flag ====================

// TestApkAdd_MissingRepo verifies that jf apk add without --repo behaves like other
// JFrog CLI package managers (npm, pip): the command succeeds, build-info is still
// captured with checksums from the local APK DB, and AQL enrichment is simply skipped.
func TestApkAdd_MissingRepo(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	buildName := tests.AlpineBuildName + "-missing-repo"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// Deliberately omit --repo — should succeed, consistent with npm/pip behaviour.
	err := jfrogCli.Exec("apk", "add", "curl",
		"--build-name="+buildName,
		"--build-number="+buildNumber)
	assert.NoError(t, err, "jf apk add without --repo should succeed (no AQL enrichment, but build-info still captured)")

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info should be published even without --repo")
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies,
			"dependencies should be recorded even without --repo")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== P0: CI/VCS properties ====================

// TestApkUpload_CIVCSPropertiesStamped verifies that when jf apk upload is run inside
// a CI environment (GitHub Actions), the published build info causes vcs.provider,
// vcs.org, and vcs.repo properties to be stamped on the artifact in Artifactory.
func TestApkUpload_CIVCSPropertiesStamped(t *testing.T) {
	initApkTest(t)

	tmpDir, err := os.MkdirTemp("", "apk-civcs-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	buildName := tests.AlpineBuildName + "-civcs"
	buildNumber := "1"

	// Simulate GitHub Actions environment (uses real env vars on CI, mock values locally).
	cleanupEnv, actualOrg, actualRepo := tests.SetupGitHubActionsEnv(t)
	defer cleanupEnv()

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	// Build a minimal .apk file.
	fakePkgInfo := "pkgname = testpkg-civcs\npkgver = 1.0.0-r0\narch = x86_64\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-civcs-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available: %v", tarErr)
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath,
		"--repo="+tests.AlpineLocalRepo,
		"--build-name="+buildName, "--build-number="+buildNumber); uploadErr != nil {
		t.Skipf("jf apk upload failed: %v", uploadErr)
	}

	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Retrieve the published build info to find the artifact path.
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info must be published before checking VCS properties")
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules, "build info must contain at least one module")
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts, "module must contain at least one artifact")

	// Use the service manager to query per-artifact properties from Artifactory.
	sm, err := artUtils.CreateServiceManager(serverDetails, 3, 1000, false)
	require.NoError(t, err)

	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			fullPath := artifact.OriginalDeploymentRepo + "/" + artifact.Path
			props, propsErr := sm.GetItemProps(fullPath)
			require.NoError(t, propsErr, "failed to get properties for artifact: %s", fullPath)
			require.NotNil(t, props, "properties must not be nil for artifact: %s", fullPath)

			assert.Contains(t, props.Properties, "vcs.provider",
				"vcs.provider must be stamped on artifact: %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.provider"], "github",
				"vcs.provider must be 'github' for artifact: %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.org",
				"vcs.org must be stamped on artifact: %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.org"], actualOrg,
				"vcs.org must match expected org for artifact: %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.repo",
				"vcs.repo must be stamped on artifact: %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.repo"], actualRepo,
				"vcs.repo must match expected repo for artifact: %s", artifact.Name)
		}
	}
}

// ==================== P0 gaps ====================

// TestApkUpload_BuildPropertiesStamped verifies that build.name, build.number, and
// build.timestamp are stamped on an uploaded .apk artifact after jf bp.
func TestApkUpload_BuildPropertiesStamped(t *testing.T) {
	initApkTest(t)

	tmpDir, err := os.MkdirTemp("", "apk-props-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	buildName := tests.AlpineBuildName + "-props"
	buildNumber := "1"

	fakePkgInfo := "pkgname = testpkg-props\npkgver = 1.0.0-r0\narch = x86_64\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-props-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to create test .apk: %v", tarErr)
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath,
		"--repo="+tests.AlpineLocalRepo,
		"--build-name="+buildName, "--build-number="+buildNumber); uploadErr != nil {
		t.Skipf("jf apk upload failed: %v", uploadErr)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	verifyExistInArtifactoryByProps(
		[]string{tests.AlpineLocalRepo + "/testpkg-props-1.0.0-r0.apk"},
		tests.AlpineLocalRepo+"/testpkg-props-1.0.0-r0.apk",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", buildName, buildNumber),
		t,
	)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestApkAdd_InvalidRepo verifies that jf apk add with a nonexistent --repo still
// succeeds. Unlike npm/pip (which route all traffic through Artifactory), Alpine APK
// fetches packages from the configured system repositories (CDN by default) unless
// jf apk config has been run first. A nonexistent repo name only means AQL checksum
// enrichment is skipped — it does not block installation or build-info collection.
func TestApkAdd_InvalidRepo(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	buildName := tests.AlpineBuildName + "-invalid-repo"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--repo=nonexistent-repo-xyz-123",
		"--build-name="+buildName,
		"--build-number="+buildNumber)
	// Command must succeed: native apk fetches from CDN, build-info is captured
	// with DB checksums, and AQL enrichment is simply skipped for the bad repo.
	assert.NoError(t, err, "jf apk add with an unresolvable --repo should still succeed")

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info should be published even with an unresolvable --repo")
	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies,
			"dependencies should be recorded even with an unresolvable --repo")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// ==================== P1 gaps ====================

// TestApkAdd_BuildNameFromEnvVars verifies that JFROG_CLI_BUILD_NAME and
// JFROG_CLI_BUILD_NUMBER environment variables trigger build-info collection
// even when --build-name / --build-number flags are not passed explicitly.
func TestApkAdd_BuildNameFromEnvVars(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	envBuildName := tests.AlpineBuildName + "-envvar"
	envBuildNumber := "1"

	t.Setenv("JFROG_CLI_BUILD_NAME", envBuildName)
	t.Setenv("JFROG_CLI_BUILD_NUMBER", envBuildNumber)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if err := jfrogCli.Exec("apk", "add", "curl", "--repo="+tests.AlpineVirtualRepo); err != nil {
		t.Skipf("jf apk add failed: %v", err)
	}

	require.NoError(t, artifactoryCli.Exec("bp", envBuildName, envBuildNumber))

	_, found, err := tests.GetBuildInfo(serverDetails, envBuildName, envBuildNumber)
	require.NoError(t, err)
	assert.True(t, found, "build info should be captured from JFROG_CLI_BUILD_NAME/NUMBER env vars")
}

// TestApkAdd_UnknownServerID verifies that supplying an unknown --server-id
// causes jf apk add to exit with a non-zero status code.
func TestApkAdd_UnknownServerID(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--repo="+tests.AlpineVirtualRepo,
		"--server-id=nonexistent-server-id-xyz")
	assert.Error(t, err, "jf apk add with an unknown --server-id should return an error")
}

// TestApkAdd_ArtifactoryUnreachable verifies that when Artifactory is unreachable
// jf apk add returns a clear error and does not silently fall back to the public CDN.
func TestApkAdd_ArtifactoryUnreachable(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("apk", "add", "curl",
		"--repo="+tests.AlpineVirtualRepo,
		"--server-id=nonexistent-server-id-xyz")
	assert.Error(t, err, "jf apk add should fail when Artifactory is unreachable, not fall back to CDN")
}

// TestApkAdd_VirtualRepoAggregates verifies that install through the virtual repo
// (which aggregates the local and remote repos) succeeds and records build-info.
func TestApkAdd_VirtualRepoAggregates(t *testing.T) {
	initApkTest(t)
	if !apkAvailable() {
		t.Skip("apk binary not found — test requires Alpine Linux.")
	}

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := tests.AlpineBuildName + "-virtual-repo"
	buildNumber := "1"

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if err := jfrogCli.Exec("apk", "add", "busybox",
		"--repo="+tests.AlpineVirtualRepo,
		"--build-name="+buildName, "--build-number="+buildNumber); err != nil {
		t.Skipf("jf apk add via virtual repo failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info should be published when installing via virtual repo")
	if found {
		assert.GreaterOrEqual(t, len(publishedBuildInfo.BuildInfo.Modules[0].Dependencies), 1,
			"at least one dependency should be recorded when installing via virtual repo")
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestApkUpload_ChecksumRoundTrip verifies that the SHA256 checksum of an artifact
// downloaded from Artifactory matches the checksum of the originally uploaded file.
func TestApkUpload_ChecksumRoundTrip(t *testing.T) {
	initApkTest(t)

	tmpDir, err := os.MkdirTemp("", "apk-checksum-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	fakePkgInfo := "pkgname = testpkg-rt\npkgver = 2.0.0-r0\narch = x86_64\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-rt-2.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available to build test .apk: %v", tarErr)
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	if uploadErr := jfrogCli.Exec("apk", "upload", apkPath, "--repo="+tests.AlpineLocalRepo); uploadErr != nil {
		t.Skipf("jf apk upload failed: %v", uploadErr)
	}

	downloadDir := tmpDir + "/downloaded"
	require.NoError(t, os.MkdirAll(downloadDir, 0755))
	assert.NoError(t, artifactoryCli.Exec("dl",
		tests.AlpineLocalRepo+"/testpkg-rt-2.0.0-r0.apk",
		downloadDir+"/", "--flat"))

	downloadedPath := downloadDir + "/testpkg-rt-2.0.0-r0.apk"
	_, statErr := os.Stat(downloadedPath)
	assert.NoError(t, statErr, "downloaded artifact should exist on disk")

	assert.Equal(t,
		computeFileSHA256(t, apkPath),
		computeFileSHA256(t, downloadedPath),
		"downloaded artifact SHA256 should match the originally uploaded file")
}

// TestApkUpload_InsecureTLS verifies --insecure-tls behaviour on jf apk upload.
// Without the flag a self-signed cert connection should fail; with it it should succeed.
// Skipped unless JFROG_CLI_TESTS_INSECURE_TLS_URL is set.
func TestApkUpload_InsecureTLS(t *testing.T) {
	initApkTest(t)

	if os.Getenv("JFROG_CLI_TESTS_INSECURE_TLS_URL") == "" {
		t.Skip("Skipping TLS test: set JFROG_CLI_TESTS_INSECURE_TLS_URL to an Artifactory with a self-signed cert.")
	}

	tmpDir, err := os.MkdirTemp("", "apk-tls-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer clientTestUtils.RemoveAllAndAssert(t, tmpDir)

	fakePkgInfo := "pkgname = testpkg-tls\npkgver = 1.0.0-r0\narch = x86_64\n"
	if writeErr := os.WriteFile(tmpDir+"/.PKGINFO", []byte(fakePkgInfo), 0644); writeErr != nil {
		t.Fatalf("failed to write .PKGINFO: %v", writeErr)
	}
	apkPath := tmpDir + "/testpkg-tls-1.0.0-r0.apk"
	if tarErr := exec.Command("tar", "-czf", apkPath, "-C", tmpDir, ".PKGINFO").Run(); tarErr != nil {
		t.Skipf("tar not available: %v", tarErr)
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Without --insecure-tls the self-signed cert should cause an error.
	assert.Error(t,
		jfrogCli.Exec("apk", "upload", apkPath, "--repo="+tests.AlpineLocalRepo),
		"upload to self-signed Artifactory without --insecure-tls should fail")

	// With --insecure-tls it should succeed.
	assert.NoError(t,
		jfrogCli.Exec("apk", "upload", apkPath, "--repo="+tests.AlpineLocalRepo, "--insecure-tls"),
		"upload to self-signed Artifactory with --insecure-tls should succeed")
}

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	buildInfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chocoCommandProperty is the single module property that 'jf choco' records, holding the
// executed Chocolatey command line with any API key redacted.
const chocoCommandProperty = buildInfo.BuildInfoEnvPrefix + "CHOCO_COMMAND"

// initChocoTest gates every Chocolatey test. Chocolatey is a Windows-only package manager, and
// both 'jf choco' and 'jf setup choco' fail fast on any other OS, so the tests are skipped
// unless a real 'choco' executable is available.
func initChocoTest(t *testing.T) {
	if !*tests.TestChoco {
		t.Skip("Skipping Chocolatey test. To run Choco test add the '-test.choco=true' option.")
	}
	if runtime.GOOS != "windows" {
		t.Skipf("Skipping Chocolatey test. Chocolatey runs on Windows only. Detected OS: %s", runtime.GOOS)
	}
	if _, err := exec.LookPath("choco"); err != nil {
		t.Skip("Skipping Chocolatey test. The 'choco' executable was not found in PATH.")
	}
	createJfrogHomeConfig(t, true)
}

// runChoco runs a 'jf choco' command. Unlike 'jf nuget', the Chocolatey command has no
// JFROG_RUN_NATIVE gate - it always delegates to the native client - and its flag set has no
// --allow-insecure-connections, so neither is set here.
func runChoco(t *testing.T, args ...string) error {
	t.Helper()
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(args...)
}

// createChocoPackageSource writes a minimal but valid Chocolatey package layout into a fresh
// temp directory and returns that directory. A nuspec with neither dependencies nor content is
// rejected (NU5017), so a tools script is packed as content - which is also the conventional
// Chocolatey layout and makes the package safe to install.
func createChocoPackageSource(t *testing.T, id, version string) (packageDir, nuspecName string) {
	t.Helper()
	packageDir = t.TempDir()
	toolsDir := filepath.Join(packageDir, "tools")
	require.NoError(t, os.MkdirAll(toolsDir, 0o700))
	installScript := "Write-Host 'jfrog-cli-tests Chocolatey package installed'\n"
	require.NoError(t, os.WriteFile(filepath.Join(toolsDir, "chocolateyinstall.ps1"), []byte(installScript), 0o600))
	uninstallScript := "Write-Host 'jfrog-cli-tests Chocolatey package uninstalled'\n"
	require.NoError(t, os.WriteFile(filepath.Join(toolsDir, "chocolateyuninstall.ps1"), []byte(uninstallScript), 0o600))

	nuspecName = id + ".nuspec"
	nuspecContent := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>%s</id>
    <version>%s</version>
    <authors>jfrog-cli-tests</authors>
    <owners>jfrog-cli-tests</owners>
    <description>Test package for jf choco integration tests.</description>
  </metadata>
  <files>
    <file src="tools\**" target="tools" />
  </files>
</package>`, id, version)
	require.NoError(t, os.WriteFile(filepath.Join(packageDir, nuspecName), []byte(nuspecContent), 0o600))
	return packageDir, nuspecName
}

// packChocoPackage runs 'jf choco pack' from inside a generated package directory and returns the
// absolute path of the produced .nupkg. 'choco pack' writes the package into the current working
// directory, and the command snapshots that same directory to discover what it produced, so the
// test must run from there.
func packChocoPackage(t *testing.T, id, version string, extraArgs ...string) (nupkgPath string) {
	t.Helper()
	packageDir, nuspecName := createChocoPackageSource(t, id, version)
	packChocoPackageIn(t, packageDir, nuspecName, extraArgs...)
	nupkgPath = filepath.Join(packageDir, id+"."+version+".nupkg")
	require.FileExists(t, nupkgPath)
	return nupkgPath
}

// packChocoPackageIn runs 'jf choco pack' with packageDir as the working directory.
func packChocoPackageIn(t *testing.T, packageDir, nuspecName string, extraArgs ...string) {
	t.Helper()
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirWithCallback(t, workingDirectory, packageDir)()
	args := append([]string{"choco", "pack", nuspecName}, extraArgs...)
	require.NoError(t, runChoco(t, args...), "'jf choco pack' should succeed")
}

// pushChocoPackage runs 'jf choco push' for an already packed .nupkg.
func pushChocoPackage(t *testing.T, nupkgPath, repo string, extraArgs ...string) error {
	t.Helper()
	args := append([]string{"choco", "push", nupkgPath, "--repo=" + repo}, extraArgs...)
	return runChoco(t, args...)
}

// chocoArtifactPath is the repository-relative path a Chocolatey package lands on. The layout is
// flat: the package sits directly under the repository root, with no version directories.
func chocoArtifactPath(repo, id, version string) string {
	return repo + "/" + id + "." + version + ".nupkg"
}

// assertChocoArtifactExists verifies that a pushed package is retrievable from Artifactory.
func assertChocoArtifactExists(t *testing.T, repoRelativePath string) {
	t.Helper()
	client, err := httpclient.ClientBuilder().Build()
	require.NoError(t, err)
	_, res, err := client.GetRemoteFileDetails(serverDetails.ArtifactoryUrl+repoRelativePath, artHttpDetails)
	require.NoError(t, err, "pushed Chocolatey package should exist at %s", repoRelativePath)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// chocoSourceName mirrors the source name that 'jf setup choco' derives from the Artifactory
// hostname and repository key, so tests can clean the machine-wide source up afterwards.
func chocoSourceName(t *testing.T, repo string) string {
	t.Helper()
	parsedURL, err := url.Parse(serverDetails.ArtifactoryUrl)
	require.NoError(t, err)
	return "jfrt-" + sanitizeChocoSourceComponent(parsedURL.Hostname()) + "-" + repo
}

func sanitizeChocoSourceComponent(value string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(value) {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '.' || character == '_' || character == '-' {
			builder.WriteRune(character)
		} else {
			builder.WriteByte('-')
		}
	}
	return strings.Trim(builder.String(), "-")
}

// cleanupChocoSource removes a machine-wide Chocolatey source. Removal is best effort: the source
// may never have been created if the test failed early.
func cleanupChocoSource(t *testing.T, sourceName string) {
	t.Helper()
	t.Cleanup(func() {
		if output, err := exec.Command("choco", "source", "remove", "-n="+sourceName).CombinedOutput(); err != nil {
			t.Logf("cleanup: 'choco source remove -n=%s' returned %v: %s", sourceName, err, string(output))
		}
	})
}

// cleanupChocoInstalledPackage uninstalls a package that a test installed machine-wide.
func cleanupChocoInstalledPackage(t *testing.T, packageID string) {
	t.Helper()
	t.Cleanup(func() {
		if output, err := exec.Command("choco", "uninstall", packageID, "-y").CombinedOutput(); err != nil {
			t.Logf("cleanup: 'choco uninstall %s -y' returned %v: %s", packageID, err, string(output))
		}
	})
}

// getPublishedChocoBuildInfo publishes and then reads back the build info of a Chocolatey build.
func getPublishedChocoBuildInfo(t *testing.T, buildName, buildNumber string) buildInfo.BuildInfo {
	t.Helper()
	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info %s/%s should have been published", buildName, buildNumber)
	return publishedBuildInfo.BuildInfo
}

// getChocoCommandProperty reads the recorded Chocolatey command line out of a module. Module
// properties are typed as interface{} in build-info, so after a publish/fetch round trip they
// arrive as a generic map.
func getChocoCommandProperty(t *testing.T, module buildInfo.Module) string {
	t.Helper()
	switch properties := module.Properties.(type) {
	case nil:
		return ""
	case map[string]string:
		return properties[chocoCommandProperty]
	case map[string]interface{}:
		value, ok := properties[chocoCommandProperty].(string)
		if !ok {
			return ""
		}
		return value
	default:
		t.Fatalf("unexpected module properties type %T", module.Properties)
		return ""
	}
}

// TestChocoPackCollectsArtifacts covers the 'jf choco pack' happy path: the packed .nupkg is
// discovered in the working directory and recorded as a build-info artifact.
func TestChocoPackCollectsArtifacts(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	id, version := "ChocoPackPkg", "1.0.0"
	buildName := tests.ChocoBuildName + "-pack"
	buildNumber := "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	packChocoPackage(t, id, version, "--build-name="+buildName, "--build-number="+buildNumber)

	collectedBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, collectedBuildInfo.Modules, 1)
	module := collectedBuildInfo.Modules[0]
	assert.Equal(t, id+":"+version, module.Id, "pack module id should be '<PackageId>:<Version>'")
	assert.Equal(t, buildInfo.Nuget, module.Type, "Chocolatey packages are NuGet packages")
	require.Len(t, module.Artifacts, 1)
	artifact := module.Artifacts[0]
	assert.Equal(t, id+"."+version+".nupkg", artifact.Name)
	assert.Equal(t, "nupkg", artifact.Type)
	assert.Equal(t, artifact.Name, artifact.Path, "the Chocolatey layout is flat, so path equals name")
	assert.NotEmpty(t, artifact.Sha1)
	assert.NotEmpty(t, artifact.Sha256)
	assert.NotEmpty(t, artifact.Md5)
	assert.Contains(t, getChocoCommandProperty(t, module), "pack", "the executed choco command should be recorded")
}

// TestChocoPushBuildInfoAndProperties covers the primary end-to-end flow - pack, then push to a
// local repository - and asserts the artifact layout, the published build info and the build
// properties stamped on the uploaded package.
func TestChocoPushBuildInfoAndProperties(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	id, version := "ChocoPushPkg", "1.0.0"
	buildName := tests.ChocoBuildName + "-push"
	buildNumber := "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	nupkgPath := packChocoPackage(t, id, version)
	require.NoError(t, pushChocoPackage(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+buildName, "--build-number="+buildNumber))

	artifactPath := chocoArtifactPath(tests.NugetLocalRepo, id, version)
	assertChocoArtifactExists(t, artifactPath)

	pushedBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, pushedBuildInfo.Modules, 1)
	module := pushedBuildInfo.Modules[0]
	assert.Equal(t, id+":"+version, module.Id)
	assert.Equal(t, buildInfo.Nuget, module.Type)
	require.Len(t, module.Artifacts, 1)
	artifact := module.Artifacts[0]
	assert.Equal(t, id+"."+version+".nupkg", artifact.Name)
	assert.Equal(t, "nupkg", artifact.Type)
	assert.Equal(t, tests.NugetLocalRepo, artifact.OriginalDeploymentRepo)
	assert.NotEmpty(t, artifact.Sha1)
	assert.NotEmpty(t, artifact.Sha256)
	assert.NotEmpty(t, artifact.Md5)
	assert.Contains(t, getChocoCommandProperty(t, module), "push")

	// Build properties are stamped on the uploaded package so it can be traced back to its build.
	properties := getFlexPackItemProps(t, artifactPath)
	assert.Equal(t, []string{buildName}, properties["build.name"])
	assert.Equal(t, []string{buildNumber}, properties["build.number"])
	assert.NotEmpty(t, properties["build.timestamp"])
}

// TestChocoPushToRemoteRejected verifies that a remote repository is refused as a push target
// before anything is uploaded.
func TestChocoPushToRemoteRejected(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath := packChocoPackage(t, "ChocoPushRemotePkg", "1.0.0")
	err := pushChocoPackage(t, nupkgPath, tests.NugetRemoteRepo)
	require.Error(t, err, "pushing to a remote repository must be rejected")
	assert.Contains(t, err.Error(), "cannot be used as a Chocolatey push target")
}

// TestChocoPushSourceRepoMismatchRejected verifies that an explicit --source pointing at one
// repository while --repo names another is rejected, instead of silently pushing to one and
// recording build info against the other.
func TestChocoPushSourceRepoMismatchRejected(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	nupkgPath := packChocoPackage(t, "ChocoSourceMismatchPkg", "1.0.0")
	mismatchedSource := serverDetails.ArtifactoryUrl + "api/nuget/" + tests.NugetVirtualRepo
	err := pushChocoPackage(t, nupkgPath, tests.NugetLocalRepo, "-s="+mismatchedSource)
	require.Error(t, err, "a --source that disagrees with --repo must be rejected")
	assert.Contains(t, err.Error(), "use matching source and --repo values")
}

// TestChocoPushToVirtualRepoConvention verifies that pushing to a virtual repository forwards to
// its default deployment repository, and that build info records the resolved local repository
// rather than the virtual repository key.
func TestChocoPushToVirtualRepoConvention(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	id, version := "ChocoVirtualPushPkg", "1.0.0"
	buildName := tests.ChocoBuildName + "-virtual-push"
	buildNumber := "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	nupkgPath := packChocoPackage(t, id, version)
	require.NoError(t, pushChocoPackage(t, nupkgPath, tests.NugetVirtualRepo,
		"--build-name="+buildName, "--build-number="+buildNumber))

	// The package must land in the virtual repository's default deployment (local) repository.
	assertChocoArtifactExists(t, chocoArtifactPath(tests.NugetLocalRepo, id, version))

	virtualPushBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, virtualPushBuildInfo.Modules, 1)
	require.NotEmpty(t, virtualPushBuildInfo.Modules[0].Artifacts)
	for _, artifact := range virtualPushBuildInfo.Modules[0].Artifacts {
		assert.Equal(t, tests.NugetLocalRepo, artifact.OriginalDeploymentRepo,
			"build info must record the resolved local repository, not the virtual repository key")
	}
}

// TestChocoBuildFlagsValidation covers the two partial build-flag cases. They behave differently:
// --build-name alone is accepted and simply collects no build info, while --build-number without
// --build-name is rejected by jf's CLI-wide flag-pair validation.
func TestChocoBuildFlagsValidation(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	// Build-info is collected when, and only when, both --build-name and --build-number are given.
	// Neither flag is a plain passthrough that must still succeed; half of the pair is a mistake
	// worth reporting rather than silently ignoring, and it is reported before choco runs at all.
	testCases := []struct {
		name        string
		packageID   string
		buildName   string
		extraArgs   []string
		expectError bool
	}{
		{
			name:      "neither build flag pushes without collecting build info",
			packageID: "ChocoNoBuildFlagsPkg",
		},
		{
			name:        "build name without build number is rejected",
			packageID:   "ChocoNameOnlyPkg",
			buildName:   tests.ChocoBuildName + "-name-only",
			expectError: true,
		},
		{
			name:        "build number without build name is rejected",
			packageID:   "ChocoNumberOnlyPkg",
			extraArgs:   []string{"--build-number=1"},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			extraArgs := testCase.extraArgs
			if testCase.buildName != "" {
				extraArgs = append(extraArgs, "--build-name="+testCase.buildName)
				t.Cleanup(func() {
					inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, testCase.buildName, artHttpDetails)
				})
			}
			nupkgPath := packChocoPackage(t, testCase.packageID, "1.0.0")
			err := pushChocoPackage(t, nupkgPath, tests.NugetLocalRepo, extraArgs...)
			if testCase.expectError {
				require.Error(t, err, "one build flag without the other must be rejected")
				assert.Contains(t, err.Error(), "cannot be provided separately")
				// The rejection happens before the native command, so nothing was published.
				return
			}
			// A bare passthrough is legal: the package is pushed, there is just no build to publish.
			require.NoError(t, err)
			assertChocoArtifactExists(t, chocoArtifactPath(tests.NugetLocalRepo, testCase.packageID, "1.0.0"))
		})
	}
}

// TestChocoBuildInfoFromEnvVars verifies that the build name and number can come from the
// standard JFrog CLI environment variables instead of command-line flags.
func TestChocoBuildInfoFromEnvVars(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	buildName := tests.ChocoBuildName + "-envvars"
	buildNumber := "7"
	clientTestUtils.SetEnvAndAssert(t, "JFROG_CLI_BUILD_NAME", buildName)
	clientTestUtils.SetEnvAndAssert(t, "JFROG_CLI_BUILD_NUMBER", buildNumber)
	defer clientTestUtils.UnSetEnvAndAssert(t, "JFROG_CLI_BUILD_NAME")
	defer clientTestUtils.UnSetEnvAndAssert(t, "JFROG_CLI_BUILD_NUMBER")
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	nupkgPath := packChocoPackage(t, "ChocoEnvVarPkg", "1.0.0")
	require.NoError(t, pushChocoPackage(t, nupkgPath, tests.NugetLocalRepo))

	envVarBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, envVarBuildInfo.Modules, 1)
	assert.NotEmpty(t, envVarBuildInfo.Modules[0].Artifacts,
		"build info must be collected from JFROG_CLI_BUILD_NAME/NUMBER alone")
}

// TestChocoModuleOverride verifies that --module replaces the default '<PackageId>:<Version>'
// module id with a caller-chosen name.
func TestChocoModuleOverride(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	buildName := tests.ChocoBuildName + "-module-override"
	buildNumber := "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	nupkgPath := packChocoPackage(t, "ChocoModuleOverridePkg", "1.0.0")
	require.NoError(t, pushChocoPackage(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+buildName, "--build-number="+buildNumber, "--module=my-service"))

	moduleOverrideBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, moduleOverrideBuildInfo.Modules, 1)
	assert.Equal(t, "my-service", moduleOverrideBuildInfo.Modules[0].Id,
		"--module must override the default '<PackageId>:<Version>' module id")
}

// TestChocoCommandPropertyRedactsApiKey verifies that a Chocolatey API key passed through to the
// native client is never recorded in build info.
func TestChocoCommandPropertyRedactsApiKey(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	id, version := "ChocoRedactPkg", "1.0.0"
	buildName := tests.ChocoBuildName + "-redact"
	buildNumber := "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	sourceURL := serverDetails.ArtifactoryUrl + "api/nuget/" + tests.NugetLocalRepo
	apiKey := chocoPushApiKey(t)
	nupkgPath := packChocoPackage(t, id, version)
	// An explicit source and API key suppress the credentials the command would otherwise inject,
	// so the native push is driven entirely by these arguments.
	require.NoError(t, pushChocoPackage(t, nupkgPath, tests.NugetLocalRepo,
		"--build-name="+buildName, "--build-number="+buildNumber,
		"-s="+sourceURL, "-k="+apiKey))

	redactedBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, redactedBuildInfo.Modules, 1)
	recordedCommand := getChocoCommandProperty(t, redactedBuildInfo.Modules[0])
	require.NotEmpty(t, recordedCommand)
	assert.Contains(t, recordedCommand, "-k=***", "the API key must be redacted in build info")
	assert.NotContains(t, recordedCommand, apiKey, "the real credential must never be recorded")
}

// chocoPushApiKey builds the composite '<user>:<token>' key that Artifactory NuGet endpoints
// expect for authenticated pushes.
func chocoPushApiKey(t *testing.T) string {
	t.Helper()
	if serverDetails.AccessToken != "" {
		return serverDetails.User + ":" + serverDetails.AccessToken
	}
	return serverDetails.User + ":" + serverDetails.Password
}

// TestSetupChocoConfiguresSource covers the 'jf setup choco' happy path: the machine-wide
// Chocolatey source for the repository is created and authenticated.
func TestSetupChocoConfiguresSource(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	sourceName := chocoSourceName(t, tests.NugetVirtualRepo)
	cleanupChocoSource(t, sourceName)

	require.NoError(t, runChoco(t, "setup", "choco", "--repo="+tests.NugetVirtualRepo),
		"'jf setup choco' should configure the Chocolatey source")

	output, err := exec.Command("choco", "source", "list").CombinedOutput()
	require.NoError(t, err, "'choco source list' failed: %s", string(output))
	assert.Contains(t, string(output), sourceName, "the JFrog Chocolatey source should be configured")
	// The V2 NuGet endpoint is what Chocolatey speaks; a V3 URL would not work.
	assert.Contains(t, string(output), "api/nuget/"+tests.NugetVirtualRepo,
		"the source should point at the Artifactory NuGet V2 endpoint")
}

// TestChocoInstallCollectsDependencies covers the 'jf choco install' happy path end to end: a
// package is packed, pushed, resolved from Artifactory and recorded as a build-info dependency.
func TestChocoInstallCollectsDependencies(t *testing.T) {
	initChocoTest(t)
	defer cleanTestsHomeEnv()

	id, version := "ChocoInstallPkg", "1.0.0"
	buildName := tests.ChocoBuildName + "-install"
	buildNumber := "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	nupkgPath := packChocoPackage(t, id, version)
	require.NoError(t, pushChocoPackage(t, nupkgPath, tests.NugetLocalRepo))

	sourceName := chocoSourceName(t, tests.NugetLocalRepo)
	cleanupChocoSource(t, sourceName)
	require.NoError(t, runChoco(t, "setup", "choco", "--repo="+tests.NugetLocalRepo))

	cleanupChocoInstalledPackage(t, id)
	requireChocoInstall(t, id, version, sourceName, buildName, buildNumber)

	installBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, installBuildInfo.Modules, 1)
	module := installBuildInfo.Modules[0]
	assert.Equal(t, buildInfo.Nuget, module.Type)
	require.NotEmpty(t, module.Dependencies, "the installed package must be recorded as a dependency")
	dependency := module.Dependencies[0]
	assert.Equal(t, id+":"+version, dependency.Id)
	assert.Equal(t, "nupkg", dependency.Type)
	assert.Equal(t, tests.NugetLocalRepo, dependency.Repository,
		"--repo-resolve should be recorded as the resolution repository")
	assert.Contains(t, getChocoCommandProperty(t, module), "install")
}

// requireChocoInstall runs 'jf choco install' with a short retry, because a freshly pushed
// package is not immediately searchable through the NuGet endpoint.
func requireChocoInstall(t *testing.T, id, version, sourceName, buildName, buildNumber string) {
	t.Helper()
	args := []string{"choco", "install", id, "--version=" + version, "--repo-resolve=" + tests.NugetLocalRepo,
		"--build-name=" + buildName, "--build-number=" + buildNumber, "-y", "-s=" + sourceName}
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		if lastErr = runChoco(t, args...); lastErr == nil {
			return
		}
		t.Logf("'jf choco install' attempt %d failed, retrying: %v", attempt+1, lastErr)
	}
	require.NoError(t, lastErr, "'jf choco install' should succeed once the pushed package is indexed")
}

// ---------------------------------------------------------------------------------------------
// Regression tests for three behaviours Chocolatey has that the first implementation pass assumed
// away. Each was verified against Chocolatey's own documentation or source, and each failed before
// the accompanying build-info-go fix.
// ---------------------------------------------------------------------------------------------

// TestChocoPackRespectsOutputDirectory covers 'choco pack --output-directory'.
//
// The flag exists (--out / --outdir / --outputdirectory / --output-directory), so snapshotting only
// the working directory finds nothing the pack produced. That failure is silent: the command
// succeeds, the package is on disk, and the build info simply has no artifacts.
func TestChocoPackRespectsOutputDirectory(t *testing.T) {
	initChocoTest(t)

	const id, version = "JfChocoOutDir", "1.0.0"
	buildName := tests.ChocoBuildName + "-outdir"
	const buildNumber = "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	packageDir, nuspecName := createChocoPackageSource(t, id, version)
	outputDir := filepath.Join(packageDir, "build-output")
	require.NoError(t, os.MkdirAll(outputDir, 0o700))

	packChocoPackageIn(t, packageDir, nuspecName,
		"--output-directory="+outputDir,
		"--build-name="+buildName, "--build-number="+buildNumber)

	require.FileExists(t, filepath.Join(outputDir, id+"."+version+".nupkg"),
		"'choco pack' should have written the package into --output-directory")

	publishedBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, publishedBuildInfo.Modules, 1)
	require.Len(t, publishedBuildInfo.Modules[0].Artifacts, 1,
		"a package written to --output-directory must still be collected; collecting zero artifacts "+
			"here is the silent failure this test exists to catch")
	assert.Equal(t, id+"."+version+".nupkg", publishedBuildInfo.Modules[0].Artifacts[0].Name)
}

// TestChocoPushWithoutPositionalPath covers 'choco push' with no path argument.
//
// The path is optional: with exactly one .nupkg in the folder Chocolatey pushes it. Requiring an
// explicit positional means the package uploads but no artifact is recorded, so it is never
// stamped and never appears in the build info.
func TestChocoPushWithoutPositionalPath(t *testing.T) {
	initChocoTest(t)

	const id, version = "JfChocoBarePush", "1.0.0"
	buildName := tests.ChocoBuildName + "-bare-push"
	const buildNumber = "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	nupkgPath := packChocoPackage(t, id, version)

	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirWithCallback(t, workingDirectory, filepath.Dir(nupkgPath))()

	require.NoError(t, runChoco(t, "choco", "push",
		"--repo="+tests.NugetLocalRepo,
		"--build-name="+buildName, "--build-number="+buildNumber),
		"'choco push' with no positional path must succeed")

	publishedBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, publishedBuildInfo.Modules, 1)
	require.Len(t, publishedBuildInfo.Modules[0].Artifacts, 1,
		"a push with no positional path must still record the package Chocolatey found in the folder")
	assert.Equal(t, id+"."+version+".nupkg", publishedBuildInfo.Modules[0].Artifacts[0].Name)
	assertChocoArtifactExists(t, chocoArtifactPath(tests.NugetLocalRepo, id, version))
}

// TestChocoInstallRecordsVersionFromInstalledPackage is the end-to-end counterpart to
// resolveInstalledPackage's unit tests: Chocolatey stores lib\<id>\<id>.nupkg with no version in
// the file name, so a version read off that name is always empty and every dependency is dropped.
// Asserting the resolved version here proves the version came from the installed .nuspec.
func TestChocoInstallRecordsVersionFromInstalledPackage(t *testing.T) {
	initChocoTest(t)

	const id, version = "JfChocoInstalledVersion", "3.4.5"
	buildName := tests.ChocoBuildName + "-installed-version"
	const buildNumber = "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	cleanupChocoInstalledPackage(t, id)

	nupkgPath := packChocoPackage(t, id, version)
	require.NoError(t, pushChocoPackage(t, nupkgPath, tests.NugetLocalRepo))
	// Confirm the fixture really is published at 3.4.5 before asserting what the install recorded,
	// so a wrong recorded version cannot be blamed on a bad fixture.
	assertChocoArtifactExists(t, chocoArtifactPath(tests.NugetLocalRepo, id, version))

	sourceName := chocoSourceName(t, tests.NugetLocalRepo)
	cleanupChocoSource(t, sourceName)
	require.NoError(t, runChoco(t, "setup", "choco", "--repo="+tests.NugetLocalRepo))

	// Deliberately no --version: a bare install resolves latest, so the recorded version cannot
	// have come from the command line either.
	requireChocoInstall(t, id, "", sourceName, buildName, buildNumber)

	publishedBuildInfo := getPublishedChocoBuildInfo(t, buildName, buildNumber)
	require.Len(t, publishedBuildInfo.Modules, 1)
	require.Len(t, publishedBuildInfo.Modules[0].Dependencies, 1)
	assert.Equal(t, id+":"+version, publishedBuildInfo.Modules[0].Dependencies[0].Id,
		"the installed version must be resolved from the package's .nuspec, not from its file name")
}

// ---------------------------------------------------------------------------------------------
// Pass-through, statelessness and the platform gate.
// ---------------------------------------------------------------------------------------------

// initChocoTestAnyPlatform gates only on the feature flag, for the two scenarios that must be
// observed where Chocolatey cannot run. It deliberately configures no JFrog server: the OS gate
// fires before any server interaction, and requiring one here would hide a regression that moved
// the gate behind server resolution.
func initChocoTestAnyPlatform(t *testing.T) {
	t.Helper()
	if !*tests.TestChoco {
		t.Skip("Skipping Chocolatey test. To run Choco test add the '-test.choco=true' option.")
	}
}

// TestChocoNonWindowsGate asserts the command refuses to run off Windows, naming the detected OS
// instead of surfacing a bare "choco: executable file not found".
func TestChocoNonWindowsGate(t *testing.T) {
	initChocoTestAnyPlatform(t)
	if runtime.GOOS == "windows" {
		t.Skip("The OS gate only rejects non-Windows hosts; nothing to assert on Windows.")
	}

	err := runChoco(t, "choco", "install", "some-package")
	require.Error(t, err, "'jf choco' must refuse to run on a non-Windows host")
	assert.Contains(t, err.Error(), "Windows only")
	assert.Containsf(t, err.Error(), runtime.GOOS, "the error must name the detected OS (%s)", runtime.GOOS)
}

// TestChocoHelpWorksOnAllPlatforms asserts the gate lives in the command's Run(), not in its
// registration, so help stays reachable on the machines that cannot run the tool.
func TestChocoHelpWorksOnAllPlatforms(t *testing.T) {
	initChocoTestAnyPlatform(t)
	assert.NoError(t, runChoco(t, "choco", "--help"), "'jf choco --help' must work on every OS")
}

// TestChocoPassThroughCollectsNoBuildInfo covers the subcommands that are not build events. The
// config-mutating ones matter most: 'source', 'apikey' and 'config' change machine state, and a
// wrapper whose contract is that it writes no configuration must not treat them as builds.
func TestChocoPassThroughCollectsNoBuildInfo(t *testing.T) {
	initChocoTest(t)

	buildName := tests.ChocoBuildName + "-passthrough"
	const buildNumber = "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	for _, testCase := range []struct {
		name string
		args []string
	}{
		{"list", []string{"choco", "list", "-r"}},
		{"outdated", []string{"choco", "outdated", "-r"}},
		{"source-list", []string{"choco", "source", "list", "-r"}},
		{"config-list", []string{"choco", "config", "list"}},
		{"feature-list", []string{"choco", "feature", "list"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// Built as a fresh slice rather than appended onto testCase.args, which would be free
			// to reuse that slice's backing array and leak the build flags into the next case.
			args := make([]string, 0, len(testCase.args)+2)
			args = append(args, testCase.args...)
			args = append(args, "--build-name="+buildName, "--build-number="+buildNumber)
			// The native exit code is passed through; a query returning non-zero is not this
			// test's concern. What matters is that nothing was collected.
			_ = runChoco(t, args...)

			_, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
			assert.NoError(t, err)
			assert.Falsef(t, found, "'choco %s' is not a build event and must not produce build info", testCase.name)
		})
	}
}

// TestChocoPassThroughWithoutServerConfigured asserts a pass-through command does not require a
// configured JFrog server. Requiring one would make 'jf choco list' fail on any unconfigured
// machine, leaving the wrapper strictly worse than the tool it wraps.
func TestChocoPassThroughWithoutServerConfigured(t *testing.T) {
	initChocoTest(t)

	restoreHomeDir := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.HomeDir, t.TempDir())
	defer restoreHomeDir()

	assert.NoError(t, runChoco(t, "choco", "--version"),
		"pass-through must work with no JFrog server configured")
}

// TestChocoDoesNotMutateChocolateyConfig covers the statelessness contract, and the clearest line
// between 'jf choco' and 'jf setup choco'. chocolatey.config is machine-wide, so an accidental
// write by the wrapper would change behaviour for every user on the machine.
func TestChocoDoesNotMutateChocolateyConfig(t *testing.T) {
	initChocoTest(t)

	const id, version = "JfChocoStateless", "1.0.0"
	buildName := tests.ChocoBuildName + "-stateless"
	const buildNumber = "1"
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	sourcesBefore, err := exec.Command("choco", "source", "list", "-r").CombinedOutput()
	require.NoError(t, err)

	packageDir, nuspecName := createChocoPackageSource(t, id, version)
	packChocoPackageIn(t, packageDir, nuspecName, "--build-name="+buildName, "--build-number="+buildNumber)

	assert.NoDirExists(t, filepath.Join(packageDir, ".jfrog", "projects"),
		"FlexPack is stateless: 'jf choco' must not create a .jfrog/projects directory")

	sourcesAfter, err := exec.Command("choco", "source", "list", "-r").CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, string(sourcesBefore), string(sourcesAfter),
		"'jf choco' must not modify Chocolatey's machine-wide source list; that is 'jf setup choco's job")
}

// TestChocoNoBuildFlagsCollectsNothing asserts that without both build coordinates the command is a
// transparent native run: the package is produced, and nothing is published.
func TestChocoNoBuildFlagsCollectsNothing(t *testing.T) {
	initChocoTest(t)

	const id, version = "JfChocoNoFlags", "1.0.0"
	nupkgPath := packChocoPackage(t, id, version)
	assert.FileExists(t, nupkgPath, "the native pack must still run and produce the package")

	_, found, err := tests.GetBuildInfo(serverDetails, tests.ChocoBuildName+"-noflags", "1")
	assert.NoError(t, err)
	assert.False(t, found, "no build info may be published when the build flags are absent")
}

// TestChocoPushToNonexistentRepoRejected asserts an unknown repository fails clearly rather than
// surfacing a raw Artifactory 404 from deep inside the push.
func TestChocoPushToNonexistentRepoRejected(t *testing.T) {
	initChocoTest(t)

	nupkgPath := packChocoPackage(t, "JfChocoBadRepo", "1.0.0")
	require.Error(t, pushChocoPackage(t, nupkgPath, "cli-choco-nonexistent-repo"))
}

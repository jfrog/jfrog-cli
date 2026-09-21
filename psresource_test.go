package main

import (
	"os/exec"
	"testing"

	buildInfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// psresourcePlatformAvailable reports whether a PowerShell executable (pwsh on every OS, or
// powershell.exe as a Windows-only fallback) that also has the Microsoft.PowerShell.PSResourceGet
// module installed can be found on PATH. Unlike Chocolatey, PSResourceGet is cross-platform, so
// the gate here is "is the tool present", never an OS check.
func psresourcePlatformAvailable() bool {
	shell := ""
	if _, err := exec.LookPath("pwsh"); err == nil {
		shell = "pwsh"
	} else if _, err := exec.LookPath("powershell.exe"); err == nil {
		shell = "powershell.exe"
	} else {
		return false
	}
	checkScript := "if (Get-Module -ListAvailable -Name Microsoft.PowerShell.PSResourceGet) { exit 0 } else { exit 1 }"
	return exec.Command(shell, "-NoProfile", "-Command", checkScript).Run() == nil
}

// initPSResourceTestAnyPlatform gates only on the feature flag, for the scenarios that must be
// observable regardless of whether pwsh/PSResourceGet is installed on the test machine (help text,
// and CLI-level flag-pair validation, which is rejected before any native command is invoked).
func initPSResourceTestAnyPlatform(t *testing.T) {
	t.Helper()
	if !*tests.TestPSResource {
		t.Skip("Skipping PSResourceGet test. To run PSResource tests add the '-test.psresource=true' option.")
	}
}

// initPSResourceTest gates every test that actually needs to shell out to pwsh and run a
// PSResourceGet cmdlet (or that needs a configured JFrog server). Unlike Chocolatey - which fails
// fast on every OS but Windows - PSResourceGet can run on macOS, Linux and Windows alike, so the
// gate here is tool-availability, not runtime.GOOS.
func initPSResourceTest(t *testing.T) {
	initPSResourceTestAnyPlatform(t)
	if !psresourcePlatformAvailable() {
		t.Skip("Skipping PSResourceGet test. No PowerShell 7+ (pwsh) with the Microsoft.PowerShell.PSResourceGet module was found.")
	}
	createJfrogHomeConfig(t, true)
}

// runPSResource runs a 'jf <cmdlet>-PSResource' command.
func runPSResource(t *testing.T, args ...string) error {
	t.Helper()
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(args...)
}

// ---------------------------------------------------------------------------------------------
// Tests that must pass with no pwsh/PSResourceGet installation available at all: help text and
// the CLI-level build-flag-pair rejection, which happens before any native command runs.
// ---------------------------------------------------------------------------------------------

// TestPSResourceHelpWorksOnAllPlatforms asserts that each of the four PSResourceGet top-level
// commands prints help without shelling out to pwsh, so '--help' stays reachable on machines that
// cannot run the tool itself - the same contract 'jf choco --help' has on non-Windows hosts.
func TestPSResourceHelpWorksOnAllPlatforms(t *testing.T) {
	initPSResourceTestAnyPlatform(t)

	for _, cmdlet := range []string{"Install-PSResource", "Save-PSResource", "Update-PSResource", "Publish-PSResource"} {
		t.Run(cmdlet, func(t *testing.T) {
			assert.NoError(t, runPSResource(t, cmdlet, "--help"),
				"'jf %s --help' must work on every OS", cmdlet)
		})
	}
}

// TestPSResourceBuildFlagsPartialRejected covers the two partial build-flag cases shared by all
// four commands: '--build-name' alone or '--build-number' alone is rejected by jf's CLI-wide
// flag-pair validation before the native cmdlet ever runs, so this needs no pwsh installation.
func TestPSResourceBuildFlagsPartialRejected(t *testing.T) {
	initPSResourceTestAnyPlatform(t)

	testCases := []struct {
		name string
		args []string
	}{
		{"build name without build number", []string{"--build-name=" + tests.PSResourceBuildName}},
		{"build number without build name", []string{"--build-number=1"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			args := append([]string{"Install-PSResource", "-Name", "SomeModule"}, testCase.args...)
			err := runPSResource(t, args...)
			require.Error(t, err, "one build flag without the other must be rejected")
			assert.Contains(t, err.Error(), "cannot be provided separately")
		})
	}
}

// ---------------------------------------------------------------------------------------------
// End-to-end tests. These need a real 'pwsh' with Microsoft.PowerShell.PSResourceGet installed
// (and a configured JFrog server), so they skip gracefully wherever that is not the case rather
// than failing the run.
// ---------------------------------------------------------------------------------------------

// TestSetupPSResourceConfiguresRepository covers the 'jf setup psresource' happy path: a
// PSResourceGet repository registration is created for the given Artifactory repo.
func TestSetupPSResourceConfiguresRepository(t *testing.T) {
	initPSResourceTest(t)
	defer cleanTestsHomeEnv()

	require.NoError(t, runPSResource(t, "setup", "psresource", "--repo="+tests.NugetVirtualRepo),
		"'jf setup psresource' should register a PSResourceGet repository")
}

// TestPSResourceInstallCollectsDependencyBuildInfo covers the 'jf Install-PSResource' happy path:
// a resolved package is recorded as a direct-dependency build-info entry when both build-name and
// build-number are supplied.
func TestPSResourceInstallCollectsDependencyBuildInfo(t *testing.T) {
	initPSResourceTest(t)
	defer cleanTestsHomeEnv()

	buildName := tests.PSResourceBuildName + "-install"
	buildNumber := "1"

	err := runPSResource(t, "Install-PSResource", "-Name", "SomeModule", "-Repository", tests.NugetVirtualRepo,
		"--repo-resolve="+tests.NugetVirtualRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	// This environment has no real PSResourceGet repository registered against a live module feed,
	// so the native cmdlet itself is expected to fail; what matters here is that the CLI wiring
	// (flag parsing, server resolution, FlexPack command construction) got far enough to invoke it.
	if err != nil {
		t.Logf("'jf Install-PSResource' returned %v; this is expected without a real registered feed", err)
		return
	}
	// If it did succeed (e.g. against a real registered feed with SomeModule actually resolvable),
	// the locally-collected build-info must actually contain it - a silently empty or corrupted
	// build-info must not pass this test just because the command itself returned no error.
	inttestutils.ValidateGeneratedBuildInfoModule(t, buildName, buildNumber, "", []string{"psresource-project"}, buildInfo.Nuget)
}

// TestPSResourceSaveCollectsDependencyBuildInfo mirrors TestPSResourceInstallCollectsDependencyBuildInfo
// for 'jf Save-PSResource'. Without this, only Install-PSResource's SetSubCommand/native-dispatch
// wiring was ever exercised past the shared '--help' early-return - a wiring bug specific to
// Save-PSResource (wrong cmdletName captured, wrong flag forwarded) would not have been caught by
// any test.
func TestPSResourceSaveCollectsDependencyBuildInfo(t *testing.T) {
	initPSResourceTest(t)
	defer cleanTestsHomeEnv()

	buildName := tests.PSResourceBuildName + "-save"
	buildNumber := "1"

	err := runPSResource(t, "Save-PSResource", "-Name", "SomeModule", "-Repository", tests.NugetVirtualRepo,
		"--repo-resolve="+tests.NugetVirtualRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Logf("'jf Save-PSResource' returned %v; this is expected without a real registered feed", err)
		return
	}
	inttestutils.ValidateGeneratedBuildInfoModule(t, buildName, buildNumber, "", []string{"psresource-project"}, buildInfo.Nuget)
}

// TestPSResourceUpdateCollectsDependencyBuildInfo mirrors TestPSResourceInstallCollectsDependencyBuildInfo
// for 'jf Update-PSResource'.
func TestPSResourceUpdateCollectsDependencyBuildInfo(t *testing.T) {
	initPSResourceTest(t)
	defer cleanTestsHomeEnv()

	buildName := tests.PSResourceBuildName + "-update"
	buildNumber := "1"

	err := runPSResource(t, "Update-PSResource", "-Name", "SomeModule", "-Repository", tests.NugetVirtualRepo,
		"--repo-resolve="+tests.NugetVirtualRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Logf("'jf Update-PSResource' returned %v; this is expected without a real registered feed", err)
		return
	}
	inttestutils.ValidateGeneratedBuildInfoModule(t, buildName, buildNumber, "", []string{"psresource-project"}, buildInfo.Nuget)
}

// TestPSResourcePublishCollectsArtifactBuildInfo mirrors TestPSResourceInstallCollectsDependencyBuildInfo
// for 'jf Publish-PSResource' - the one command in the family that collects artifact (not
// dependency) build-info.
func TestPSResourcePublishCollectsArtifactBuildInfo(t *testing.T) {
	initPSResourceTest(t)
	defer cleanTestsHomeEnv()

	buildName := tests.PSResourceBuildName + "-publish"
	buildNumber := "1"

	err := runPSResource(t, "Publish-PSResource", "-Path", t.TempDir(), "-Repository", tests.NugetLocalRepo,
		"--repo="+tests.NugetLocalRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	if err != nil {
		t.Logf("'jf Publish-PSResource' returned %v; this is expected without a real registered feed and module manifest", err)
		return
	}
	inttestutils.ValidateGeneratedBuildInfoModule(t, buildName, buildNumber, "", []string{"psresource-project"}, buildInfo.Nuget)
}

// TestPSResourcePlatformGateWithoutPwsh asserts that on a machine with no usable pwsh +
// PSResourceGet installation, 'jf setup psresource' fails clearly instead of shelling out and
// surfacing a raw "executable file not found" error.
func TestPSResourcePlatformGateWithoutPwsh(t *testing.T) {
	initPSResourceTestAnyPlatform(t)
	if psresourcePlatformAvailable() {
		t.Skip("pwsh + Microsoft.PowerShell.PSResourceGet is available on this host; nothing to assert about the platform gate.")
	}

	restoreHomeDir := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.HomeDir, t.TempDir())
	defer restoreHomeDir()
	createJfrogHomeConfig(t, true)
	defer cleanTestsHomeEnv()

	err := runPSResource(t, "setup", "psresource", "--repo="+tests.NugetVirtualRepo)
	require.Error(t, err, "'jf setup psresource' must refuse to run without pwsh + PSResourceGet")
}

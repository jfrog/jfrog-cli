package main

import (
	"encoding/json"
	"net/url"
	"os"
	"testing"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/setup"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// otherHostArtifactoryUrl is a server that no setup test configures, so status against it
// must report the package manager as pointing elsewhere.
const otherHostArtifactoryUrl = "https://other-host.example.com/artifactory/"

// runSetupStatus runs `jf setup <packageManager> --status --format=json` and returns the report.
func runSetupStatus(t *testing.T, packageManager string, extraArgs ...string) setup.PackageManagerStatus {
	// --format=json makes execMain set this variable for the rest of the process.
	t.Setenv(cliutils.JfrogCliErrorOutputFormat, os.Getenv(cliutils.JfrogCliErrorOutputFormat))
	outputBuffer, _, previousLog := coreTests.RedirectLogOutputToBuffer()
	defer log.SetLogger(previousLog)

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	args := append([]string{"setup", packageManager, "--status", "--format=json"}, extraArgs...)
	require.NoError(t, jfrogCli.WithoutCredentials().Exec(args...))

	var status setup.PackageManagerStatus
	require.NoError(t, json.Unmarshal(outputBuffer.Bytes(), &status), outputBuffer.String())
	assert.Equal(t, setup.StatusSchemaVersion, status.SchemaVersion)
	assert.Equal(t, packageManager, status.PackageManager)
	return status
}

func assertSetupStatus(t *testing.T, packageManager string, expectedState setup.ConfigState, expectedRepoKey string, extraArgs ...string) setup.PackageManagerStatus {
	status := runSetupStatus(t, packageManager, extraArgs...)
	assert.Equal(t, expectedState, status.State, "location: %s", status.Location)
	if expectedRepoKey != "" {
		assert.Equal(t, expectedRepoKey, status.RepoKey)
	}
	return status
}

// assertSetupStatusNotConfigured is the check before `jf setup` runs. The maven, nuget and
// dotnet setup tests skip it: their setup writes the real user-level settings.xml or
// NuGet.Config, which the tests back up and restore but do not isolate, so what it holds
// before setup depends on the machine.
func assertSetupStatusNotConfigured(t *testing.T, packageManager string, extraArgs ...string) {
	assertSetupStatus(t, packageManager, setup.StateNotConfigured, "", extraArgs...)
}

// assertSetupStatusAfterSetup checks the configuration `jf setup` just wrote: it is configured
// for the test server with the expected repository and credentials, --verify reaches that
// repository with them, and it reads as another host's configuration when status is asked
// about a different server.
func assertSetupStatusAfterSetup(t *testing.T, packageManager, expectedRepoKey string) {
	status := assertSetupStatus(t, packageManager, setup.StateConfigured, expectedRepoKey)
	artifactoryUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	require.NoError(t, err)
	assert.Equal(t, artifactoryUrl.Host, status.Host)
	assert.Equal(t, setup.CredentialsPresent, status.Credentials)
	if assert.NotNil(t, status.BinaryFound) {
		assert.True(t, *status.BinaryFound)
	}
	assert.Nil(t, status.Verify, "status does not contact the server without --verify")

	verified := assertSetupStatus(t, packageManager, setup.StateConfigured, expectedRepoKey, "--verify")
	if assert.NotNil(t, verified.Verify) {
		// On Windows NuGet encrypts the stored password, which status cannot read, so the check
		// is sent anonymously and can only report that it could not authenticate.
		if coreutils.IsWindows() && (packageManager == project.Nuget.String() || packageManager == project.Dotnet.String()) {
			assert.NotEmpty(t, verified.Verify.Error)
		} else {
			assert.Equal(t, setup.VerifyStatus{RepoReachable: true, AuthOk: setup.ProbeTrue}, *verified.Verify)
		}
	}

	assertSetupStatus(t, packageManager, setup.StateOtherHost, "", "--url="+otherHostArtifactoryUrl)
}

// assertContainerSetupStatusAfterSetup is assertSetupStatusAfterSetup for docker, podman and
// helm. Registry logins carry no repository key, so --verify has nothing to check, and logins
// to several registries are normal, so a different server reads as not-configured rather
// than other-host.
func assertContainerSetupStatusAfterSetup(t *testing.T, packageManager string, extraArgs ...string) {
	status := assertSetupStatus(t, packageManager, setup.StateConfigured, "", extraArgs...)
	assert.Equal(t, setup.CredentialsPresent, status.Credentials)
	verified := assertSetupStatus(t, packageManager, setup.StateConfigured, "", append(extraArgs, "--verify")...)
	assert.Equal(t, &setup.VerifyStatus{AuthOk: setup.ProbeUnknown,
		Error: "not verified: " + packageManager + " logs in to the registry, not to a repository, so there is no repository to check"}, verified.Verify)
	assertSetupStatus(t, packageManager, setup.StateNotConfigured, "", "--url="+otherHostArtifactoryUrl)
}

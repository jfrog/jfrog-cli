package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/setup"
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

func assertSetupStatus(t *testing.T, packageManager string, expectedState setup.ConfigState, expectedRepoKey string, extraArgs ...string) {
	status := runSetupStatus(t, packageManager, extraArgs...)
	assert.Equal(t, expectedState, status.State, "location: %s", status.Location)
	if expectedRepoKey != "" {
		assert.Equal(t, expectedRepoKey, status.RepoKey)
	}
}

func assertSetupStatusNotConfigured(t *testing.T, packageManager string, extraArgs ...string) {
	assertSetupStatus(t, packageManager, setup.StateNotConfigured, "", extraArgs...)
}

// assertSetupStatusAfterSetup checks the configuration `jf setup` just wrote: it is configured
// for the test server with the expected repository, and reads as another host's configuration
// when status is asked about a different server.
func assertSetupStatusAfterSetup(t *testing.T, packageManager, expectedRepoKey string) {
	assertSetupStatus(t, packageManager, setup.StateConfigured, expectedRepoKey)
	assertSetupStatus(t, packageManager, setup.StateOtherHost, "", "--url="+otherHostArtifactoryUrl)
}

// assertContainerSetupStatusAfterSetup is assertSetupStatusAfterSetup for docker, podman and
// helm. Registry logins carry no repository key, and logins to several registries are normal,
// so a different server reads as not-configured rather than other-host.
func assertContainerSetupStatusAfterSetup(t *testing.T, packageManager string, extraArgs ...string) {
	assertSetupStatus(t, packageManager, setup.StateConfigured, "", extraArgs...)
	assertSetupStatus(t, packageManager, setup.StateNotConfigured, "", "--url="+otherHostArtifactoryUrl)
}

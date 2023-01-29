package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreenvsetup "github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func initAccessTest(t *testing.T) {
	if !*tests.TestAccess {
		t.Skip("Skipping Access test. To run Access test add the '-test.access=true' option.")
	}
}

func TestSetupInvitedUser(t *testing.T) {
	initAccessTest(t)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.HomeDir, tempDirPath)
	defer setEnvCallBack()
	serverDetails := &config.ServerDetails{Url: *tests.JfrogUrl, AccessToken: *tests.JfrogAccessToken}
	encodedCred := encodeConnectionDetails(serverDetails, t)
	setupCmd := coreenvsetup.NewEnvSetupCommand().SetEncodedConnectionDetails(encodedCred)
	suffix := setupCmd.SetupAndConfigServer()
	assert.Empty(t, suffix)
	configs, err := config.GetAllServersConfigs()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(configs))
	// Verify config values
	assert.Equal(t, configs[0].Url, *tests.JfrogUrl)
	assert.Equal(t, *tests.JfrogUrl+"artifactory/", configs[0].ArtifactoryUrl)
	// Verify token was refreshed
	assert.NotEqual(t, *tests.JfrogAccessToken, configs[0].AccessToken)
	assert.NotEmpty(t, configs[0].RefreshToken)
}

func encodeConnectionDetails(serverDetails *config.ServerDetails, t *testing.T) string {
	jsonConnectionDetails, err := json.Marshal(serverDetails)
	assert.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString(jsonConnectionDetails)
	return encoded
}

func TestRefreshableAccessTokens(t *testing.T) {
	initAccessTest(t)

	server := &config.ServerDetails{Url: *tests.JfrogUrl, AccessToken: *tests.JfrogAccessToken}
	err := coreenvsetup.GenerateNewLongTermRefreshableAccessToken(server)
	assert.NoError(t, err)
	assert.NotEmpty(t, server.RefreshToken)
	configCmd := commands.NewConfigCommand(commands.AddOrEdit, tests.ServerId).SetDetails(server).SetInteractive(false)
	assert.NoError(t, configCmd.Run())
	defer deleteServerConfig(t)

	// Upload a file and assert the refreshable tokens were generated.
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", "")
	uploadedFiles := 1
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/a1.in", uploadedFiles)
	if !assert.NoError(t, err) {
		return
	}
	curAccessToken, curRefreshToken, err := getAccessTokensFromConfig(t, tests.ServerId)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEmpty(t, curAccessToken)
	assert.NotEmpty(t, curRefreshToken)

	// Make the token always refresh.
	auth.InviteRefreshBeforeExpiryMinutes = 365 * 24 * 60

	// Upload a file and assert tokens were refreshed.
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/a2.in", uploadedFiles)
	if !assert.NoError(t, err) {
		return
	}
	curAccessToken, curRefreshToken, err = assertTokensChanged(t, tests.ServerId, curAccessToken, curRefreshToken)
	if !assert.NoError(t, err) {
		return
	}

	// Make the token not refresh. Verify Tokens did not refresh.
	auth.InviteRefreshBeforeExpiryMinutes = 0
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, tests.ServerId, "testdata/a/b/b2.in", uploadedFiles)
	if !assert.NoError(t, err) {
		return
	}
	newAccessToken, newRefreshToken, err := getArtifactoryTokensFromConfig(t, tests.ServerId)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, curAccessToken, newAccessToken)
	assert.Equal(t, curRefreshToken, newRefreshToken)

	// Cleanup
	cleanArtifactoryTest()
}

func getAccessTokensFromConfig(t *testing.T, serverId string) (accessToken, refreshToken string, err error) {
	details, err := config.GetSpecificConfig(serverId, false, false)
	if err != nil {
		assert.NoError(t, err)
		return "", "", err
	}
	return details.AccessToken, details.RefreshToken, nil
}

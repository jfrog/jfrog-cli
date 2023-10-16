package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreEnvSetup "github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var (
	accessDetails     *config.ServerDetails
	accessCli         *tests.JfrogCli
	accessHttpDetails httputils.HttpClientDetails
)

func initAccessTest(t *testing.T) {
	if !*tests.TestAccess {
		t.Skip("Skipping Access test. To run Access test add the '-test.access=true' option.")
	}
}

func initAccessCli() {
	if accessCli != nil {
		return
	}
	accessCli = tests.NewJfrogCli(execMain, "jfrog", authenticateAccess())
}

func InitAccessTests() {
	initArtifactoryCli()
	initAccessCli()
	cleanUpOldBuilds()
	cleanUpOldRepositories()
	cleanUpOldUsers()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
	cleanArtifactoryTest()
}

func authenticateAccess() string {
	*tests.JfrogUrl = clientUtils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	accessDetails = &config.ServerDetails{
		AccessUrl: *tests.JfrogUrl + tests.AccessEndpoint}

	cred := fmt.Sprintf("--url=%s", *tests.JfrogUrl)
	if *tests.JfrogAccessToken != "" {
		accessDetails.AccessToken = *tests.JfrogAccessToken
		cred += fmt.Sprintf(" --access-token=%s", accessDetails.AccessToken)
	} else {
		accessDetails.User = *tests.JfrogUser
		accessDetails.Password = *tests.JfrogPassword
		cred += fmt.Sprintf(" --user=%s --password=%s", accessDetails.User, accessDetails.Password)
	}

	accessAuth, err := accessDetails.CreateAccessAuthConfig()
	if err != nil {
		coreutils.ExitOnErr(err)
	}
	accessHttpDetails = accessAuth.CreateHttpClientDetails()
	return cred
}

func TestSetupInvitedUser(t *testing.T) {
	initAccessTest(t)
	tempDirPath, createTempDirCallback := coreTests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.HomeDir, tempDirPath)
	defer setEnvCallBack()
	setupServerDetails := &config.ServerDetails{Url: *tests.JfrogUrl, AccessToken: *tests.JfrogAccessToken}
	encodedCred := encodeConnectionDetails(setupServerDetails, t)
	setupCmd := coreEnvSetup.NewEnvSetupCommand().SetEncodedConnectionDetails(encodedCred)
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
	err := coreEnvSetup.GenerateNewLongTermRefreshableAccessToken(server)
	assert.NoError(t, err)
	assert.NotEmpty(t, server.RefreshToken)
	configCmd := commands.NewConfigCommand(commands.AddOrEdit, tests.ServerId).SetDetails(server).SetInteractive(false)
	assert.NoError(t, configCmd.Run())
	defer deleteServerConfig(t)

	// Upload a file and assert the refreshable tokens were generated.
	artifactoryCommandExecutor := tests.NewJfrogCli(execMain, "jfrog rt", "")
	uploadedFiles := 1
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/a1.in", uploadedFiles)
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
	auth.RefreshPlatformTokenBeforeExpiryMinutes = 365 * 24 * 60

	// Upload a file and assert tokens were refreshed.
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/a2.in", uploadedFiles)
	if !assert.NoError(t, err) {
		return
	}
	curAccessToken, curRefreshToken, err = assertTokensChanged(t, curAccessToken, curRefreshToken)
	if !assert.NoError(t, err) {
		return
	}

	// Make the token not refresh. Verify Tokens did not refresh.
	auth.RefreshPlatformTokenBeforeExpiryMinutes = 0
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/b/b2.in", uploadedFiles)
	if !assert.NoError(t, err) {
		return
	}
	newAccessToken, newRefreshToken, err := getArtifactoryTokensFromConfig(t)
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

const (
	userScope     = "applied-permissions/user"
	defaultExpiry = 31536000
)

var atcTestCases = []struct {
	name                string
	args                []string
	shouldExpire        bool
	expectedExpiry      uint
	expectedScope       string
	expectedRefreshable bool
	expectedReference   bool
}{
	{
		name:                "default",
		args:                []string{"atc"},
		shouldExpire:        true,
		expectedExpiry:      defaultExpiry,
		expectedScope:       userScope,
		expectedRefreshable: false,
		expectedReference:   false,
	},
	{
		name:                "explicit user, no expiry",
		args:                []string{"atc", auth.ExtractUsernameFromAccessToken(*tests.JfrogAccessToken), "--expiry=0"},
		shouldExpire:        false,
		expectedExpiry:      0,
		expectedScope:       userScope,
		expectedRefreshable: false,
		expectedReference:   false,
	},
	{
		name:                "refreshable, admin",
		args:                []string{"atc", "--refreshable", "--grant-admin"},
		shouldExpire:        true,
		expectedExpiry:      defaultExpiry,
		expectedScope:       "applied-permissions/admin",
		expectedRefreshable: true,
		expectedReference:   false,
	},
	{
		name:                "reference, custom scope, custom expiry",
		args:                []string{"atc", "--reference", "--scope=system:metrics:r", "--expiry=123456"},
		shouldExpire:        true,
		expectedExpiry:      123456,
		expectedScope:       "system:metrics:r",
		expectedRefreshable: false,
		expectedReference:   true,
	},
	{
		name:                "groups, description",
		args:                []string{"atc", "--groups=group1,group2", "--description=description"},
		shouldExpire:        true,
		expectedExpiry:      defaultExpiry,
		expectedScope:       "applied-permissions/groups:group1,group2",
		expectedRefreshable: false,
		expectedReference:   false,
	},
}

func TestAccessTokenCreate(t *testing.T) {
	initAccessTest(t)
	if *tests.JfrogAccessToken == "" {
		t.Skip("access token create command only supports authorization with access token, but a token is not provided. Skipping...")
	}

	for _, test := range atcTestCases {
		t.Run(test.name, func(t *testing.T) {
			var token auth.CreateTokenResponseData
			output := accessCli.RunCliCmdWithOutput(t, test.args...)
			assert.NoError(t, json.Unmarshal([]byte(output), &token))
			defer revokeToken(t, token.TokenId)

			if test.shouldExpire {
				assert.EqualValues(t, test.expectedExpiry, *token.ExpiresIn)
			} else {
				assert.Nil(t, token.ExpiresIn)
			}
			assert.NotEmpty(t, token.AccessToken)
			assert.Equal(t, test.expectedScope, token.Scope)
			assertNotEmptyIfExpected(t, test.expectedRefreshable, token.RefreshToken)
			assertNotEmptyIfExpected(t, test.expectedReference, token.ReferenceToken)

			// Try pinging Artifactory with the new token.
			assert.NoError(t, tests.NewJfrogCli(execMain, "jfrog rt",
				"--url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint+" --access-token="+token.AccessToken).Exec("ping"))
		})
	}
}

func assertNotEmptyIfExpected(t *testing.T, expected bool, output string) {
	if expected {
		assert.NotEmpty(t, output)
	} else {
		assert.Empty(t, output)
	}
}

func revokeToken(t *testing.T, tokenId string) {
	if tokenId == "" {
		return
	}

	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	resp, _, err := client.SendDelete(*tests.JfrogUrl+"access/api/v1/tokens/"+tokenId, nil, accessHttpDetails, "")
	assert.NoError(t, err)
	assert.NoError(t, errorutils.CheckResponseStatus(resp, http.StatusOK))
}

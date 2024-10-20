package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/access/services"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/stretchr/testify/assert"
)

var (
	accessDetails     *config.ServerDetails
	accessCli         *coreTests.JfrogCli
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
	accessCli = coreTests.NewJfrogCli(execMain, "jfrog", authenticateAccess())
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

func TestRefreshableAccessTokens(t *testing.T) {
	initAccessTest(t)

	server := &config.ServerDetails{Url: *tests.JfrogUrl, AccessToken: *tests.JfrogAccessToken}
	err := generateNewLongTermRefreshableAccessToken(server)
	assert.NoError(t, err)
	assert.NotEmpty(t, server.RefreshToken)
	configCmd := commands.NewConfigCommand(commands.AddOrEdit, tests.ServerId).SetDetails(server).SetInteractive(false)
	assert.NoError(t, configCmd.Run())
	defer deleteServerConfig(t)

	// Upload a file and assert the refreshable tokens were generated.
	artifactoryCommandExecutor := coreTests.NewJfrogCli(execMain, "jfrog rt", "")
	uploadedFiles := 1
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/a1.in", uploadedFiles)
	if !assert.NoError(t, err) {
		return
	}
	curAccessToken, curRefreshToken, curArtifactoryRefreshToken, err := getTokensFromConfig(t)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEmpty(t, curAccessToken)
	assert.NotEmpty(t, curRefreshToken)
	assert.Empty(t, curArtifactoryRefreshToken)

	// Make the token always refresh.
	auth.RefreshPlatformTokenBeforeExpiryMinutes = 365 * 24 * 60

	// Upload a file and assert tokens were refreshed.
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/a2.in", uploadedFiles)
	if !assert.NoError(t, err) {
		return
	}
	curAccessToken, curRefreshToken, err = assertAccessTokensChanged(t, curAccessToken, curRefreshToken)
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
	newAccessToken, newRefreshToken, newArtifactoryRefreshToken, err := getTokensFromConfig(t)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, curAccessToken, newAccessToken)
	assert.Equal(t, curRefreshToken, newRefreshToken)
	assert.Empty(t, newArtifactoryRefreshToken)

	// Cleanup
	cleanArtifactoryTest()
}

// Take the short-lived token and generate a long term (1 year expiry) refreshable accessToken.
func generateNewLongTermRefreshableAccessToken(server *config.ServerDetails) (err error) {
	accessManager, err := utils.CreateAccessServiceManager(server, false)
	if err != nil {
		return
	}
	// Create refreshable accessToken with 1 year expiry from the given short expiry token.
	params := createLongExpirationRefreshableTokenParams()
	token, err := accessManager.CreateAccessToken(*params)
	if err != nil {
		return
	}
	server.AccessToken = token.AccessToken
	server.RefreshToken = token.RefreshToken
	return
}

func createLongExpirationRefreshableTokenParams() *services.CreateTokenParams {
	params := services.CreateTokenParams{}
	// Using the platform's default expiration (1 year by default).
	params.ExpiresIn = nil
	params.Refreshable = clientUtils.Pointer(true)
	params.Audience = "*@*"
	return &params
}

// After refreshing an access token, assert that the access token and the refresh token were changed, and the Artifactory refresh token remained empty.
func assertAccessTokensChanged(t *testing.T, curAccessToken, curRefreshToken string) (newAccessToken, newRefreshToken string, err error) {
	var newArtifactoryRefreshToken string
	newAccessToken, newRefreshToken, newArtifactoryRefreshToken, err = getTokensFromConfig(t)
	if err != nil {
		assert.NoError(t, err)
		return "", "", err
	}
	assert.NotEqual(t, curAccessToken, newAccessToken)
	assert.NotEqual(t, curRefreshToken, newRefreshToken)
	assert.Empty(t, newArtifactoryRefreshToken)
	return newAccessToken, newRefreshToken, nil
}

const (
	userScope = "applied-permissions/user"
)

func TestAccessTokenCreate(t *testing.T) {
	initAccessTest(t)
	if *tests.JfrogAccessToken == "" {
		t.Skip("access token create command only supports authorization with access token, but a token is not provided. Skipping...")
	}

	var testCases = []struct {
		name         string
		args         []string
		shouldExpire bool
		// The expected expiry or -1 if we use the default expiry value
		expectedExpiry      int
		expectedScope       string
		expectedRefreshable bool
		expectedReference   bool
	}{
		{
			name:                "default",
			args:                []string{"atc"},
			shouldExpire:        true,
			expectedExpiry:      -1,
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
			expectedExpiry:      -1,
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
			expectedExpiry:      -1,
			expectedScope:       "applied-permissions/groups:group1,group2",
			expectedRefreshable: false,
			expectedReference:   false,
		},
	}

	// Discard output logging to prevent negative logs
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var token auth.CreateTokenResponseData
			output := accessCli.RunCliCmdWithOutput(t, testCase.args...)
			assert.NoError(t, json.Unmarshal([]byte(output), &token))
			defer revokeToken(t, token.TokenId)

			if testCase.shouldExpire {
				if testCase.expectedExpiry == -1 {
					// If expectedExpiry is -1, expect the default expiry
					assert.Positive(t, *token.ExpiresIn)
				} else {
					assert.EqualValues(t, testCase.expectedExpiry, *token.ExpiresIn)
				}
			} else {
				assert.Nil(t, token.ExpiresIn)
			}
			assert.NotEmpty(t, token.AccessToken)
			assert.Equal(t, testCase.expectedScope, token.Scope)
			assertNotEmptyIfExpected(t, testCase.expectedRefreshable, token.RefreshToken)
			assertNotEmptyIfExpected(t, testCase.expectedReference, token.ReferenceToken)

			// Try pinging Artifactory with the new token.
			assert.NoError(t, coreTests.NewJfrogCli(execMain, "jfrog rt",
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

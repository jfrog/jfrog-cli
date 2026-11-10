package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/general/token"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	"github.com/stretchr/testify/require"
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

	artifactoryCommandExecutor := coreTests.NewJfrogCli(execMain, "jfrog rt", "")
	uploadedFiles := 0

	// A token whose total lifetime falls under the fixed 30-minute refresh floor (see
	// auth.CalculateProportionalRefreshThreshold) is always due for refresh, regardless of
	// auth.RefreshPlatformTokenBeforeExpiryMinutes. Use one to verify a refresh actually happens.
	shortLivedServer := &config.ServerDetails{Url: *tests.JfrogUrl, AccessToken: *tests.JfrogAccessToken}
	err := generateShortLivedRefreshableAccessToken(shortLivedServer)
	assert.NoError(t, err)
	assert.NotEmpty(t, shortLivedServer.RefreshToken)
	configCmd := commands.NewConfigCommand(commands.AddOrEdit, tests.ServerId).SetDetails(shortLivedServer).SetInteractive(false)
	assert.NoError(t, configCmd.Run())

	// Upload a file and assert the refreshable tokens were generated.
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/a1.in", uploadedFiles)
	if !assert.NoError(t, err) {
		deleteServerConfig(t)
		return
	}
	curAccessToken, curRefreshToken, curArtifactoryRefreshToken, err := getTokensFromConfig(t)
	if !assert.NoError(t, err) {
		deleteServerConfig(t)
		return
	}
	assert.NotEmpty(t, curAccessToken)
	assert.NotEmpty(t, curRefreshToken)
	assert.Empty(t, curArtifactoryRefreshToken)

	// Upload another file and assert tokens were refreshed, since the short-lived token is always due.
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/a2.in", uploadedFiles)
	if !assert.NoError(t, err) {
		deleteServerConfig(t)
		return
	}
	_, _, err = assertAccessTokensChanged(t, curAccessToken, curRefreshToken)
	if !assert.NoError(t, err) {
		deleteServerConfig(t)
		return
	}

	// A long-lived token's proportional threshold (10% of its own lifetime) is always well above
	// RefreshPlatformTokenBeforeExpiryMinutes=0, so it's never due for refresh. Reuse the same
	// server ID with fresh, long-lived credentials to verify a refresh does NOT happen.
	longLivedServer := &config.ServerDetails{Url: *tests.JfrogUrl, AccessToken: *tests.JfrogAccessToken}
	err = generateNewLongTermRefreshableAccessToken(longLivedServer)
	assert.NoError(t, err)
	assert.NotEmpty(t, longLivedServer.RefreshToken)
	configCmd = commands.NewConfigCommand(commands.AddOrEdit, tests.ServerId).SetDetails(longLivedServer).SetInteractive(false)
	assert.NoError(t, configCmd.Run())
	defer deleteServerConfig(t)

	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/b/b2.in", uploadedFiles)
	if !assert.NoError(t, err) {
		return
	}
	curAccessToken, curRefreshToken, curArtifactoryRefreshToken, err = getTokensFromConfig(t)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEmpty(t, curAccessToken)
	assert.NotEmpty(t, curRefreshToken)
	assert.Empty(t, curArtifactoryRefreshToken)

	// Make the token not refresh. Verify Tokens did not refresh.
	auth.RefreshPlatformTokenBeforeExpiryMinutes = 0
	uploadedFiles++
	err = uploadWithSpecificServerAndVerify(t, artifactoryCommandExecutor, "testdata/a/b/b3.in", uploadedFiles)
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
	if err != nil && isInvalidFormattedNameError(err) {
		if username, ok := getCloudCompatibleUsername(*tests.JfrogAccessToken); ok {
			params.Username = username
			token, err = accessManager.CreateAccessToken(*params)
		}
	}
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

// Take the short-lived token and generate a refreshable accessToken whose own total lifetime is
// short enough (well under 300 minutes) to always fall under the fixed 30-minute refresh floor in
// auth.CalculateProportionalRefreshThreshold, making it unconditionally due for refresh.
func generateShortLivedRefreshableAccessToken(server *config.ServerDetails) (err error) {
	accessManager, err := utils.CreateAccessServiceManager(server, false)
	if err != nil {
		return
	}
	params := createShortExpirationRefreshableTokenParams()
	token, err := accessManager.CreateAccessToken(*params)
	if err != nil && isInvalidFormattedNameError(err) {
		if username, ok := getCloudCompatibleUsername(*tests.JfrogAccessToken); ok {
			params.Username = username
			token, err = accessManager.CreateAccessToken(*params)
		}
	}
	if err != nil {
		return
	}
	server.AccessToken = token.AccessToken
	server.RefreshToken = token.RefreshToken
	return
}

func createShortExpirationRefreshableTokenParams() *services.CreateTokenParams {
	params := services.CreateTokenParams{}
	// One minute — far below the 30-minute refresh floor, so the token is always due for refresh.
	params.ExpiresIn = clientUtils.Pointer(uint(60))
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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var token auth.CreateTokenResponseData
			output, err := runAtcWithCloudFallback(t, testCase.args...)
			if !assert.NoError(t, err) {
				return
			}
			if !assert.NoError(t, json.Unmarshal([]byte(output), &token)) {
				return
			}
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

type atvOutput struct {
	Valid   bool                   `json:"valid"`
	Error   string                 `json:"error"`
	Payload map[string]interface{} `json:"payload"`
	Types   []string               `json:"types"`
}

func TestAccessTokenVerify(t *testing.T) {
	initAccessTest(t)
	if *tests.JfrogAccessToken == "" {
		t.Skip("access token verify tests mint tokens with an admin access token, but a token is not provided. Skipping...")
	}
	createJfrogHomeConfig(t, true)
	defer cleanTestsHomeEnv()
	atvCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	mintWithoutCleanup := func(t *testing.T, args ...string) (auth.CreateTokenResponseData, string) {
		var created auth.CreateTokenResponseData
		output, err := runAtcWithCloudFallback(t, append([]string{"atc"}, args...)...)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal([]byte(output), &created))
		return created, output
	}
	mint := func(t *testing.T, args ...string) (auth.CreateTokenResponseData, string) {
		created, output := mintWithoutCleanup(t, args...)
		t.Cleanup(func() { revokeToken(t, created.TokenId) })
		return created, output
	}
	writeFile := func(t *testing.T, content string) string {
		path := filepath.Join(t.TempDir(), "token.txt")
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}
	verify := func(t *testing.T, secret string, args ...string) (atvOutput, error) {
		output, err := atvCli.RunCliCmdWithOutputs(t, append([]string{"atv", "--format=json"}, args...)...)
		assert.NotContains(t, output, secret)
		if err != nil {
			assert.NotContains(t, err.Error(), secret)
		}
		var result atvOutput
		if strings.TrimSpace(output) != "" {
			assert.NoError(t, json.Unmarshal([]byte(output), &result))
		}
		return result, err
	}
	assertValid := func(t *testing.T, result atvOutput, err error, types ...string) {
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Equal(t, types, result.Types)
	}

	t.Run("identity token as argument", func(t *testing.T) {
		created, _ := mint(t)
		result, err := verify(t, created.AccessToken, created.AccessToken)
		assertValid(t, result, err, "Identity Token")
	})
	t.Run("token file with the bare token", func(t *testing.T) {
		created, _ := mint(t)
		result, err := verify(t, created.AccessToken, "--token-file="+writeFile(t, created.AccessToken+"\n"))
		assertValid(t, result, err, "Identity Token")
	})
	t.Run("token file with jf atc JSON output", func(t *testing.T) {
		created, output := mint(t)
		result, err := verify(t, created.AccessToken, "--token-file="+writeFile(t, output))
		assertValid(t, result, err, "Identity Token")
	})
	t.Run("admin token", func(t *testing.T) {
		created, _ := mint(t, "--grant-admin")
		result, err := verify(t, created.AccessToken, created.AccessToken)
		assertValid(t, result, err, "Admin Token")
	})
	t.Run("scoped token", func(t *testing.T) {
		created, _ := mint(t, "--scope=system:metrics:r")
		result, err := verify(t, created.AccessToken, created.AccessToken)
		assertValid(t, result, err, "Scoped Token")
	})
	t.Run("groups token", func(t *testing.T) {
		created, _ := mint(t, "--groups=group1,group2")
		result, err := verify(t, created.AccessToken, created.AccessToken)
		assertValid(t, result, err, "Applied Permissions Groups Token")
	})
	t.Run("token without expiry", func(t *testing.T) {
		created, _ := mint(t, auth.ExtractUsernameFromAccessToken(*tests.JfrogAccessToken), "--expiry=0")
		result, err := verify(t, created.AccessToken, created.AccessToken)
		assertValid(t, result, err, "Identity Token")
		assert.NotContains(t, result.Payload, "exp")
	})
	t.Run("reference token", func(t *testing.T) {
		created, _ := mint(t, "--reference")
		require.NotEmpty(t, created.ReferenceToken)
		_, err := verify(t, created.ReferenceToken, created.ReferenceToken)
		assert.ErrorContains(t, err, "reference token")
	})
	t.Run("revoked token", func(t *testing.T) {
		created, _ := mintWithoutCleanup(t)
		revokeToken(t, created.TokenId)
		_, err := verify(t, created.AccessToken, created.AccessToken)
		assert.ErrorContains(t, err, "token rejected by Access server: status 40")
	})
	t.Run("tampered signature", func(t *testing.T) {
		created, _ := mint(t)
		i := len(created.AccessToken) - 10
		tampered := created.AccessToken[:i] + flipBase64Char(created.AccessToken[i]) + created.AccessToken[i+1:]
		_, err := verify(t, tampered, tampered)
		assert.ErrorContains(t, err, "invalid signature")
	})
	t.Run("locally signed token", func(t *testing.T) {
		forged := signLocalToken(t)
		_, err := verify(t, forged, forged)
		assert.ErrorContains(t, err, "invalid signature")
	})
	t.Run("unknown server id", func(t *testing.T) {
		created, _ := mint(t)
		_, err := verify(t, created.AccessToken, created.AccessToken, "--server-id=nope")
		assert.ErrorContains(t, err, "nope")
	})
}

func flipBase64Char(c byte) string {
	if c == 'A' {
		return "B"
	}
	return "A"
}

func signLocalToken(t *testing.T) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	encode := func(v interface{}) string {
		data, err := json.Marshal(v)
		require.NoError(t, err)
		return base64.RawURLEncoding.EncodeToString(data)
	}
	now := time.Now().Unix()
	signingInput := encode(map[string]string{"alg": "RS256", "typ": "JWT", "ver": "2", "kid": "forged"}) + "." +
		encode(map[string]interface{}{
			"jti": "forged", "sub": "jfac@01h/users/admin", "aud": "*@*", "scp": "applied-permissions/admin",
			"iss": "jfac@01h", "iat": now - 60, "exp": now + 3600,
		})
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	require.NoError(t, err)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func runAtcWithCloudFallback(t *testing.T, args ...string) (string, error) {
	output, err := accessCli.RunCliCmdWithOutputs(t, args...)
	if err == nil || !isInvalidFormattedNameError(err) {
		return output, err
	}
	username, ok := getCloudCompatibleUsername(*tests.JfrogAccessToken)
	if !ok {
		return output, err
	}
	fallbackArgs := replaceOrInsertUsernameArg(args, username)
	return accessCli.RunCliCmdWithOutputs(t, fallbackArgs...)
}

func replaceOrInsertUsernameArg(args []string, username string) []string {
	if len(args) <= 1 {
		return append(args, username)
	}
	newArgs := append([]string{}, args...)
	// Username is positional arg #1 for "jfrog atc". If absent, first arg after "atc" starts with '-'.
	if strings.HasPrefix(newArgs[1], "-") {
		return append([]string{newArgs[0], username}, newArgs[1:]...)
	}
	newArgs[1] = username
	return newArgs
}

func isInvalidFormattedNameError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Invalid formatted name")
}

func getCloudCompatibleUsername(accessToken string) (string, bool) {
	subject, err := auth.ExtractSubjectFromAccessToken(accessToken)
	if err != nil || subject == "" {
		return "", false
	}
	// Already cloud-formatted or service formatted subject.
	if strings.Contains(subject, "/users/") || strings.HasPrefix(subject, "jfrt@") {
		return subject, true
	}
	payload, err := extractTokenPayload(accessToken)
	if err != nil || payload.Issuer == "" {
		return "", false
	}
	return payload.Issuer + "/users/" + subject, true
}

func extractTokenPayload(tokenValue string) (*auth.TokenPayload, error) {
	parts := strings.Split(tokenValue, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid token format")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var payload auth.TokenPayload
	if err = json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func TestOidcExchangeToken(t *testing.T) {
	// If token ID was not provided by the CI, skip this test
	if os.Getenv(coreutils.OidcExchangeTokenId) == "" {
		t.Skip("No token ID available in environment, skipping test")
		return
	}
	accessCli = coreTests.NewJfrogCli(execMain, "jfrog", "")

	output := accessCli.RunCliCmdWithOutput(t, "eot", "setup-jfrog-cli-test", "--url=https://ecosysjfrog.jfrog.io")
	var result token.ExchangeCommandOutputStruct
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "Output should be valid JSON")
	assert.NotEmpty(t, result.AccessToken, "AccessToken should not be empty")
	assert.NotEmpty(t, result.Username, "Username should not be empty")

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

package commands

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli/utils/log"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func init() {
	log.SetDefaultLogger()
}

func TestBasicAuth(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "admin", Password: "password",
		ApiKey: "", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestApiKey(t *testing.T) {
	// API key is no longer allowed to be configured without providing a username.
	// This test is here to make sure that old configurations (with API key and no username) are still accepted.
	inputDetails := config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "", Password: "",
		ApiKey: "apiKey", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)

	inputDetails = config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "admin", Password: "",
		ApiKey: "apiKey", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestSshKey(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{
		Url:             "ssh://localhost:1339/",
		DistributionUrl: "http://localhost:1339/distribution",
		User:            "", Password: "",
		ApiKey: "", SshKeyPath: "/tmp/sshKey", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestAccessToken(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "", Password: "",
		ApiKey: "", SshKeyPath: "", AccessToken: "accessToken",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestRefreshToken(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "", Password: "",
		ApiKey: "", SshKeyPath: "", AccessToken: "accessToken", RefreshToken: "refreshToken",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)

	inputDetails = config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "user", Password: "pass",
		ApiKey: "", SshKeyPath: "", AccessToken: "", RefreshToken: "", TokenRefreshInterval: 10,
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestEmpty(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "", Password: "",
		ApiKey: "", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func configAndTest(t *testing.T, inputDetails *config.ArtifactoryDetails) {
	outputConfig, err := configAndGetTestServer(t, inputDetails, false)
	assert.NoError(t, err)
	assert.Equal(t, configStructToString(inputDetails), configStructToString(outputConfig), "unexpected configuration was saved to file")
	assert.NoError(t, DeleteConfig("test"))
	testExportImport(t, inputDetails)
}

func configAndGetTestServer(t *testing.T, inputDetails *config.ArtifactoryDetails, basicAuthOnly bool) (*config.ArtifactoryDetails, error) {
	configCmd := NewConfigCommand().SetDetails(inputDetails).SetServerId("test").SetUseBasicAuthOnly(basicAuthOnly)
	assert.NoError(t, configCmd.Config())
	return GetConfig("test")
}

func configStructToString(artConfig *config.ArtifactoryDetails) string {
	artConfig.IsDefault = false
	marshaledStruct, _ := json.Marshal(*artConfig)
	return string(marshaledStruct)
}

func TestEscapingUrlInConfigurationFromUser(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "admin", Password: "password",
		ApiKey: "", SshKeyPath: "", AccessToken: "",
		ServerId:  "test", ClientCertPath: "test/cert/path", ClientCertKeyPath: "test/cert/key/path",
		IsDefault: false}

	configCmd := NewConfigCommand().SetDetails(&inputDetails).SetDefaultDetails(&inputDetails).SetUseBasicAuthOnly(true)
	assert.NoError(t, configCmd.getConfigurationFromUser())
	assert.True(t, strings.HasSuffix(inputDetails.GetUrl(), "/"), "expected url to end with /")
}

func TestBasicAuthOnlyOption(t *testing.T) {
	// Test the functionality of the option
	inputDetails := config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		User:            "admin", Password: "password",
		ServerId:  "test", IsDefault: false}
	outputConfig, err := configAndGetTestServer(t, &inputDetails, true)
	assert.NoError(t, err)
	assert.Zero(t, outputConfig.TokenRefreshInterval, "expected refreshable token to be disabled")
	assert.NoError(t, DeleteConfig("test"))
}

func testExportImport(t *testing.T, inputDetails *config.ArtifactoryDetails) {
	serverToken, err := config.Export(inputDetails)
	assert.NoError(t, err)
	outputDetails, err := config.Import(serverToken)
	assert.NoError(t, err)
	assert.Equal(t, configStructToString(inputDetails), configStructToString(outputDetails), "unexpected configuration was saved to file")
}

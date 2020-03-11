package commands

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/log"
	"strings"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	log.SetDefaultLogger()
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
	configCmd := NewConfigCommand().SetDetails(inputDetails).SetServerId("test")
	err := configCmd.Config()
	if err != nil {
		t.Error(err.Error())
	}
	outputConfig, err := GetConfig("test")
	if err != nil {
		t.Error(err.Error())
	}
	if configStructToString(inputDetails) != configStructToString(outputConfig) {
		t.Error("Unexpected configuration was saved to file. Expected: " + configStructToString(inputDetails) + " Got " + configStructToString(outputConfig))
	}
	err = DeleteConfig("test")
	if err != nil {
		t.Error(err.Error())
	}
	testExportImport(t, inputDetails)
}

func configStructToString(artConfig *config.ArtifactoryDetails) string {
	artConfig.IsDefault = false
	marshaledStruct, _ := json.Marshal(*artConfig)
	return string(marshaledStruct)
}

func TestGetConfigurationFromUser(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{
		Url:             "http://localhost:8080/artifactory",
		DistributionUrl: "http://localhost:8080/distribution",
		User:            "admin", Password: "password",
		ApiKey: "", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}

	configCmd := NewConfigCommand().SetDetails(&inputDetails).SetDefaultDetails(&inputDetails)
	err := configCmd.getConfigurationFromUser()
	if err != nil {
		t.Error(err)
	}

	if !strings.HasSuffix(inputDetails.GetUrl(), "/") {
		t.Error("Expected url to end with /")
	}
}

func testExportImport(t *testing.T, inputDetails *config.ArtifactoryDetails) {
	serverToken, err := config.Export(inputDetails)
	if err != nil {
		t.Error(err.Error())
	}
	outputDetails, err := config.Import(serverToken)
	if err != nil {
		t.Error(err.Error())
	}
	if configStructToString(inputDetails) != configStructToString(outputDetails) {
		t.Error("Unexpected configuration was saved to file. Expected: " + configStructToString(inputDetails) + " Got " + configStructToString(outputDetails))
	}
}

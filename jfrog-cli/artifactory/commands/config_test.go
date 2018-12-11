package commands

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{Url: "http://localhost:8080/artifactory",
		User: "admin", Password: "password",
		ApiKey: "", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestApiKey(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{Url: "http://localhost:8080/artifactory",
		User: "", Password: "",
		ApiKey: "apiKey", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)

	inputDetails = config.ArtifactoryDetails{Url: "http://localhost:8080/artifactory",
		User: "admin", Password: "",
		ApiKey: "apiKey", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestSshKey(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{Url: "ssh://localhost:1339/",
		User: "", Password: "",
		ApiKey: "", SshKeyPath: "/tmp/sshKey", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestAccessToken(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{Url: "http://localhost:8080/artifactory",
		User: "", Password: "",
		ApiKey: "", SshKeyPath: "", AccessToken: "accessToken",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func TestEmpty(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{Url: "http://localhost:8080/artifactory",
		User: "", Password: "",
		ApiKey: "", SshKeyPath: "", AccessToken: "",
		ServerId:  "test",
		IsDefault: false}
	configAndTest(t, &inputDetails)
}

func configAndTest(t *testing.T, inputDetails *config.ArtifactoryDetails) {
	_, err := Config(inputDetails, nil, false, false, "test")
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
}

func configStructToString(artConfig *config.ArtifactoryDetails) string {
	artConfig.IsDefault = false
	marshaledStruct, _ := json.Marshal(*artConfig)
	return string(marshaledStruct)
}

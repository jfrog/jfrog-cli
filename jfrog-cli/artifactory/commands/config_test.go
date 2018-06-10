package commands

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"testing"
)

func TestConfig(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{Url: "http://localhost:8080/artifactory",
		User: "admin", Password: "password",
		ApiKey: "", SshKeyPath: "", ServerId: "test",
		IsDefault: false}
	_, err := Config(&inputDetails, nil, false, false, "test")
	if err != nil {
		t.Error(err.Error())
	}
	outputConfig, err := GetConfig("test")
	if err != nil {
		t.Error(err.Error())
	}
	if configStructToString(&inputDetails) != configStructToString(outputConfig) {
		t.Error("Unexpected configuration was saved to file. Expected: " + configStructToString(&inputDetails) + " Got " + configStructToString(outputConfig))
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

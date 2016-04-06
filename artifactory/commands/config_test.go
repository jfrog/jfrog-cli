package commands

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"testing"
)

func TestConfig(t *testing.T) {
	inputDetails := config.ArtifactoryDetails{"http://localhost:8080/artifactory", "admin", "password", "", nil, nil}
	Config(&inputDetails, nil, false, false)
	outputConfig := GetConfig()
	if configStructToString(&inputDetails) != configStructToString(outputConfig) {
		t.Error("Unexpected configuration was saved to file. Expected: " + configStructToString(&inputDetails) + " Got " + configStructToString(outputConfig))
	}
}

func configStructToString(artConfig *config.ArtifactoryDetails) string {
	marshaledStruct, _ := json.Marshal(*artConfig)
	return string(marshaledStruct)
}

package commands

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"testing"
)

func TestConfig(t *testing.T) {
	inputDetails := cliutils.ArtifactoryDetails{"http://localhost:8080/artifactory", "admin", "password", "", nil}
	Config(&inputDetails, nil, false, false)
	outputConfig := GetConfig()
	if configStructToString(&inputDetails) != configStructToString(outputConfig) {
		t.Error("Unexpected configuration was saved to file. Expected: " + configStructToString(&inputDetails) + " Got " + configStructToString(outputConfig))
	}
}

func configStructToString(artConfig *cliutils.ArtifactoryDetails) string {
	marshaledStruct, _ := json.Marshal(*artConfig)
	return string(marshaledStruct)
}

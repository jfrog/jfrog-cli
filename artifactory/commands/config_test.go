package commands

import (
	"encoding/json"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	inputDetails := cliutils.ArtifactoryDetails{"http://localhost:8080/artifactory", "admin", "password", "", nil}
	Config(&inputDetails, nil, false, false)
	outputConfig := GetConfig()
	printConfigStruct(&inputDetails)
	printConfigStruct(outputConfig)
	if !reflect.DeepEqual(inputDetails, *outputConfig) {
		t.Error("Unexpected configuration was saved to file. Expected: " + configStructToString(&inputDetails) + " Got " + configStructToString(outputConfig))
	}
}

func configStructToString(artConfig *cliutils.ArtifactoryDetails) string {
	marshaledStruct, _ := json.Marshal(*artConfig)
	return string(marshaledStruct)
}

func printConfigStruct(artConfig *cliutils.ArtifactoryDetails) {
	stringSturct := configStructToString(artConfig)
	fmt.Println(stringSturct)
}

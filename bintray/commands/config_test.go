package commands

import (
    "encoding/json"
	"github.com/JFrogDev/jfrog-cli-go/cliutils"
    "github.com/JFrogDev/jfrog-cli-go/bintray/tests"
	"testing"
)

func TestConfig(t *testing.T) {
    expected := tests.CreateBintrayDetails()
    Config(expected, nil, false)
    details := GetConfig()
	if configStructToString(expected) != configStructToString(details) {
		t.Error("Unexpected configuration was saved to file. Expected: " + configStructToString(expected) + " Got " + configStructToString(details))
	}
}

func configStructToString(config *cliutils.BintrayDetails) string {
	marshaledStruct, _ := json.Marshal(*config)
	return string(marshaledStruct)
}
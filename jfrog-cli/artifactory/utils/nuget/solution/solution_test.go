package solution

import (
	"encoding/json"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"reflect"
	"testing"
)

func TestEmptySolution(t *testing.T) {
	solution, err := Load(".")
	if err != nil {
		t.Error(err)
	}

	expected := &buildinfo.BuildInfo{}
	buildInfo, err := solution.BuildInfo()
	if err != nil {
		t.Error("An error occurred while creating the build info object", err.Error())
	}
	if !reflect.DeepEqual(buildInfo, expected) {
		expectedString, _ := json.Marshal(expected)
		buildInfoString, _ := json.Marshal(buildInfo)
		t.Errorf("Expecting: \n%s \nGot: \n%s", expectedString, buildInfoString)
	}
}

package solution

import (
	"testing"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	"reflect"
)

func TestEmptySolution(t *testing.T) {
	solution, err := Load(".")
	if err != nil {
		t.Error(err)
	}

	expected := &buildinfo.BuildInfo{}
	buildInfo := solution.BuildInfo()
	if !reflect.DeepEqual(buildInfo, expected) {
		t.Errorf("Expecting: \n%s \nGot: \n%s", expected, buildInfo)
	}
}

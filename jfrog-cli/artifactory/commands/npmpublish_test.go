package commands

import (
	"testing"
	"reflect"
)

func TestReadPackageInfo(t *testing.T) {
	var packages = map[string]packageInfo{
		`{ "name": "jfrog-cli-tests", "version": "1.0.0", "description": "test package"}`:
		{Name: "jfrog-cli-tests", Version: "1.0.0", scope: ""},
		`{ "name": "@jfrog/jfrog-cli-tests", "version": "1.0.0", "description": "test package"}`:
		{Name: "jfrog-cli-tests", Version: "1.0.0", scope: "@jfrog"}}

	for stringJson, packInfo := range packages {
		npmp := npmPublish{}
		npmp.readPackageInfo([]byte(stringJson))
		equals := reflect.DeepEqual(&packInfo, npmp.packageInfo)
		if !equals {
			t.Error("expeted:", packInfo, "got:", npmp.packageInfo)
		}
	}
}

func TestGetDeployPath(t *testing.T) {
	var packages = map[string]packageInfo{
		`jfrog-cli-tests/-/jfrog-cli-tests-1.0.0.tgz`:
		{Name: "jfrog-cli-tests", Version: "1.0.0", scope: ""},
		`@jfrog/jfrog-cli-tests/-/@jfrog/jfrog-cli-tests-1.0.0.tgz`:
		{Name: "jfrog-cli-tests", Version: "1.0.0", scope: "@jfrog"}}

	for expectedPath, packInfo := range packages {
		actualPath := getDeployPath(&packInfo)
		if actualPath != expectedPath {
			t.Error("expeted:", expectedPath, "got:", actualPath)
		}
	}
}

func TestGetExpectedPackedFileName(t *testing.T) {
	var packages = map[string]packageInfo{
		`jfrog-cli-tests-1.0.0.tgz`:
		{Name: "jfrog-cli-tests", Version: "1.0.0", scope: ""},
		`jfrog-jfrog-cli-tests-1.0.0.tgz`:
		{Name: "jfrog-cli-tests", Version: "1.0.0", scope: "@jfrog"}}

	for expectedFileName, packInfo := range packages {
		actualFileName := getExpectedPackedFileName(&packInfo)
		if actualFileName != expectedFileName {
			t.Error("expeted:", expectedFileName, "got:", actualFileName)
		}
	}
}

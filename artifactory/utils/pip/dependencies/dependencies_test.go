package dependencies

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var environmentPipDependencyMap = map[string]pipDependencyPackage{
	"pkg-1":     {packageType{"pkg-1", "pkg-1", "0.33.4"}, []dependency{}},
	"pkg-dep-7": {packageType{"pkg-dep-7", "pkg-dep-7", "19.2.1"}, []dependency{}},
	"pkg.dep-5": {packageType{"pkg.dep-5", "pkg.DEP-5", "1.11"}, []dependency{{"pkg6", "pkg6", "0.16.1"}}},
	"pkg2":      {packageType{"pkg2", "Pkg2", "41.0.1"}, []dependency{}},
	"pkg3":      {packageType{"pkg3", "pkg3", "3.5"}, []dependency{{"pkg2", "Pkg2", "41.0.1"}, {"pkg.dep-5", "pkg.DEP-5", "1.11"}, {"pkg6", "pkg6", "0.16.1"}}},
	"pkg4":      {packageType{"pkg4", "pkg4", "0.0.1"}, []dependency{}},
	"pkg6":      {packageType{"pkg6", "pkg6", "0.16.1"}, []dependency{{"pkg-dep-7", "pkg-dep-7", "19.2.1"}}},
	"pkg8":      {packageType{"pkg8", "pkg8", "0.13.2"}, []dependency{{"pkg-dep-7", "pkg-dep-7", "19.2.1"}, {"pkg8", "pkg8", "0.13.2"}}},
}

func TestParsePipDepTree(t *testing.T) {
	log.SetDefaultLogger()

	// Create file path.
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	pipdeptreeTestFilePath := filepath.Join(pwd, "testsdata/pipdeptree_output")

	// Read file.
	content, err := fileutils.ReadFile(pipdeptreeTestFilePath)
	if err != nil {
		t.Error(err)
	}

	// Parse content.
	actualDepTree, err := parsePipDependencyMapOutput(content)
	if err != nil {
		t.Error(err)
	}

	// Check identical.
	if !reflect.DeepEqual(environmentPipDependencyMap, actualDepTree) {
		t.Errorf("Expected: %v, Got: %v", environmentPipDependencyMap, actualDepTree)
	}
}

func TestExtractDependencies(t *testing.T) {
	log.SetDefaultLogger()

	tests := []struct {
		rootDependencies        []string
		expectedAllDependencies map[string]*buildinfo.Dependency
		expectedChildrenMap     map[string][]string
	}{
		{[]string{"pkg-1", "pkg-dep-7"}, map[string]*buildinfo.Dependency{"pkg-1": {"pkg-1:0.33.4", "", nil, nil}, "pkg-dep-7": {"pkg-dep-7:19.2.1", "", nil, nil}}, map[string][]string{"pkg-1": nil, "pkg-dep-7": nil}},
		{[]string{"pkg3"}, map[string]*buildinfo.Dependency{"pkg2": {"Pkg2:41.0.1", "", nil, nil}, "pkg.dep-5": {"pkg.DEP-5:1.11", "", nil, nil}, "pkg6": {"pkg6:0.16.1", "", nil, nil}, "pkg-dep-7": {"pkg-dep-7:19.2.1", "", nil, nil}, "pkg3": {"pkg3:3.5", "", nil, nil}}, map[string][]string{"pkg3": {"pkg2", "pkg.dep-5", "pkg6"}, "pkg-dep-7": nil, "pkg2": nil, "pkg.dep-5": {"pkg6"}, "pkg6": {"pkg-dep-7"}}},
		{[]string{"pkg8"}, map[string]*buildinfo.Dependency{"pkg8": {"pkg8:0.13.2", "", nil, nil}, "pkg-dep-7": {"pkg-dep-7:19.2.1", "", nil, nil}}, map[string][]string{"pkg8": {"pkg-dep-7", "pkg8"}, "pkg-dep-7": nil}},
	}

	for _, test := range tests {
		// Parse content.
		actualAllDependencies, actualChildrenMap, err := extractDependencies(test.rootDependencies, environmentPipDependencyMap)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(actualAllDependencies, test.expectedAllDependencies) {
			t.Error(fmt.Sprintf("Expected value: %v, got: %v.", test.expectedAllDependencies, actualAllDependencies))
		}
		if !reflect.DeepEqual(actualChildrenMap, test.expectedChildrenMap) {
			t.Error(fmt.Sprintf("Expected value: %v, got: %v.", test.expectedChildrenMap, actualChildrenMap))
		}
	}
}

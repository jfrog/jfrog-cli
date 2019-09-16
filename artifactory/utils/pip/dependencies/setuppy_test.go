package dependencies

import (
	"fmt"
	"reflect"
	"testing"
)

var testDependencyMap = map[string]pipDependencyPackage{
	"pkg1": {
		packageType{"pkg1", "Pkg1", "1.0.0"},
		[]dependency{
			{"dep11", "Dep11", "1.0.1"},
			{"dep12", "Dep12", "1.4.1"}}},
	"pkg2": {
		packageType{"pkg2", "PkG2", "2.2.2"},
		[]dependency{
			{"dep21", "DEP21", "2.0.1"},
		}},
	"pkg-.3": {
		packageType{"pkg-.3", "PkG-.3", "3.10.1"},
		[]dependency{
			{"de--p31", "dE--P31", "3.0.15"},
			{"dependency-of-3-2", "dEpenDenCy-oF-3-2", "3.3.1"},
			{"d.e.p.3.3", "d.E.P.3.3", "30.0.1"},
			{"d..ep34", "d..EP34", "343.334.443"},
		}},
	"pkg4": {
		packageType{"pkg4", "PkG4", "4.2.4"},
		[]dependency{},
	}}

func TestExtractRootDependencies(t *testing.T) {
	tests := []struct {
		packageName string
		rootDeps    []string
	}{
		{"pkg1", []string{"dep11", "dep12"}},
		{"pkg2", []string{"dep21"}},
		{"pkg-.3", []string{"de--p31", "dependency-of-3-2", "d.e.p.3.3", "d..ep34"}},
		{"pkg4", nil},
	}

	for _, test := range tests {
		actualValue, err := extractRootDependencies(testDependencyMap, test.packageName)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(actualValue, test.rootDeps) {
			t.Error(fmt.Sprintf("Expected value: %v, got: %v.", test.rootDeps, actualValue))
		}
	}
}

func TestGetProjectNameFromFileContent(t *testing.T) {
	tests := []struct {
		fileContent         string
		expectedProjectName string
	}{
		{"Metadata-Version: 1.0\nName: jfrog-python-example-1\nVersion: 1.0\nSummary: Project example for building Python project with JFrog products\nHome-page: https://github.com/jfrog/project-examples\nAuthor: JFrog\nAuthor-email: jfrog@jfrog.com\nLicense: UNKNOWN\nDescription: UNKNOWN\nPlatform: UNKNOWN", "jfrog-python-example-1"},
		{"Metadata-Version: Name: jfrog-python-example-2\nLicense: UNKNOWN\nDescription: UNKNOWN\nPlatform: UNKNOWN\nName: jfrog-python-example-2\nVersion: 1.0\nSummary: Project example for building Python project with JFrog products\nHome-page: https://github.com/jfrog/project-examples\nAuthor: JFrog\nAuthor-email: jfrog@jfrog.com", "jfrog-python-example-2"},
		{"Name:Metadata-Version: 3.0\nName: jfrog-python-example-3\nVersion: 1.0\nSummary: Project example for building Python project with JFrog products\nHome-page: https://github.com/jfrog/project-examples\nAuthor: JFrog\nAuthor-email: jfrog@jfrog.com\nName: jfrog-python-example-4", "jfrog-python-example-3"},
	}

	for _, test := range tests {
		actualValue, err := getProjectNameFromFileContent([]byte(test.fileContent))
		if err != nil {
			t.Error(err)
		}
		if actualValue != test.expectedProjectName {
			t.Errorf("Expected value: %s, got: %s.", test.expectedProjectName, actualValue)
		}
	}
}

func TestExtractPkginfoPathFromCommandOutput(t *testing.T) {
	tests := []struct {
		commandOutput       string
		expectedPkginfoPath string
		shouldFail	bool
	}{
		{"running egg_info\nwriting jfrog_python_example_unix.egg-info/PKG-INFO\nwriting dependency_links to jfrog_python_example.egg-info/dependency_links.txt\nwriting requirements to jfrog_python_example.egg-info/requires.txt\nwriting top-level names to jfrog_python_example.egg-info/top_level.txt\nreading manifest file 'jfrog_python_example.egg-info/SOURCES.txt'\nwriting manifest file 'jfrog_python_example.egg-info/SOURCES.txt'", "jfrog_python_example_unix.egg-info/PKG-INFO", false},
		{"running egg_info\nwriting jfrog_python_example_windows.egg-info\\PKG-INFO\nwriting dependency_links to jfrog_python_example.egg-info\\dependency_links.txt\nwriting requirements to jfrog_python_example.egg-info\\requires.txt\nwriting top-level names to jfrog_python_example.egg-info\\top_level.txt\nreading manifest file 'jfrog_python_example.egg-info\\SOURCES.txt'\nwriting manifest file 'jfrog_python_example.egg-info\\SOURCES.txt'", "jfrog_python_example_windows.egg-info\\PKG-INFO", false},
		{"running egg_info\nwriting dependency_links to jfrog_python_example.egg-info/dependency_links.txt\nwriting requirements to jfrog_python_example.egg-info/requires.txt\nwriting top-level names to jfrog_python_example.egg-info/top_level.txt\nreading manifest file 'jfrog_python_example.egg-info/SOURCES.txt'\nwriting manifest file 'jfrog_python_example.egg-info/SOURCES.txt'", "jfrog_python_example_unix.egg-info/PKG-INFO", true},
	}

	for i, test := range tests {
		actualValue, err := extractPkginfoPathFromCommandOutput(test.commandOutput)
		if err != nil {
			if !test.shouldFail {
				t.Errorf("Test case %d - %s", i, err)
			}
			continue
		}
		if actualValue != test.expectedPkginfoPath {
			t.Errorf("Test case %d - Expected value: %s, got: %s.", i, test.expectedPkginfoPath, actualValue)
		}
	}
}

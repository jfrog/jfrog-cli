package dependencies

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetPackageZipLocation(t *testing.T) {
	cachePath := filepath.Join(tests.GetBaseDir(true), "zip", "download")
	tests := []struct {
		dependencyName string
		version        string
		expectedPath   string
	}{
		{"rsc.io/quote", "v1.5.2", filepath.Join(filepath.Dir(cachePath), "rsc.io", "quote", "@v", "v1.5.2.zip")},
		{"rsc/quote", "v1.5.3", filepath.Join(cachePath, "rsc", "quote", "@v", "v1.5.3.zip")},
		{"rsc.io/quote", "v1.5.4", ""},
	}

	for _, test := range tests {
		t.Run(test.dependencyName+":"+test.version, func(t *testing.T) {
			actual, err := getPackageZipLocation(cachePath, test.dependencyName, test.version)
			if err != nil {
				t.Error(err.Error())
			}

			if test.expectedPath != actual {
				t.Errorf("Test name: %s:%s: Expected: %s, Got: %s", test.dependencyName, test.version, test.expectedPath, actual)
			}
		})
	}
}

func TestGetDependencyName(t *testing.T) {

	tests := []struct {
		dependencyName string
		expectedPath   string
	}{
		{"github.com/Sirupsen/logrus", "github.com/!sirupsen/logrus"},
		{"Rsc/quOte", "!rsc/qu!ote"},
		{"golang.org/x/crypto", "golang.org/x/crypto"},
		{"golang.org/X/crypto", "golang.org/!x/crypto"},
		{"rsc.io/quote", "rsc.io/quote"},
	}

	for _, test := range tests {
		t.Run(test.dependencyName, func(t *testing.T) {
			actual := getDependencyName(test.dependencyName)
			if test.expectedPath != actual {
				t.Errorf("Test name: %s: Expected: %s, Got: %s", test.dependencyName, test.expectedPath, actual)
			}
		})
	}
}

func TestCreateDependencyWithMod(t *testing.T) {
	err := fileutils.CreateTempDirPath()
	if err != nil {
		t.Error(err)
	}
	defer fileutils.RemoveTempDir()
	tempDir, err := fileutils.GetTempDirPath()
	if err != nil {
		t.Error(err)
	}
	baseDir := tests.GetBaseDir(true)
	cachePath := filepath.Join(baseDir, "zip")
	modContent := "module github.com/test"
	dep := Package{
		id:         "github.com/test:v1.2.3",
		modContent: []byte(modContent),
		zipPath:    filepath.Join(cachePath, "v1.2.3.zip"),
	}
	pathReturned, err := createDependencyWithMod(dep)
	if err != nil {
		t.Error(err)
	}
	path := filepath.Join(tempDir, "github.com", "test@v1.2.3", "go.mod")
	if path != pathReturned {
		t.Error(fmt.Sprintf("Expected %s, got %s", path, pathReturned))
	}

	mod, err := ioutil.ReadFile(pathReturned)
	if err != nil {
		t.Error(err)
	}

	if modContent != string(mod) {
		t.Error(fmt.Sprintf("Expected %s, got %s", modContent, string(mod)))
	}
	err = os.RemoveAll(filepath.Join(tempDir, "github.com"))
	if err != nil {
		t.Error(err)
	}
}

func TestMergeReplaceDependenciesWithGraphDependencies(t *testing.T) {
	tests := []struct {
		name              string
		replaceDeps       []string
		graphDependencies map[string]bool
		expectedMap       map[string]bool
	}{
		{"missingInGraphMap", []string{"replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v0.1.0", "replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-cli-go", "replace github.com/jfrog/jfrog-client-go => /path/to/mod/file"}, map[string]bool{},
			map[string]bool{"github.com/jfrog/jfrog-client-go@v0.1.0": true}},
		{"existsInGraphMap", []string{"replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v0.1.0", "replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-cli-go", "replace github.com/jfrog/jfrog-client-go => /path/to/mod/file"}, map[string]bool{"github.com/jfrog/jfrog-client-go@v0.1.0": true},
			map[string]bool{"github.com/jfrog/jfrog-client-go@v0.1.0": true}},
		{"addToGraphMap", []string{"replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v0.1.0", "replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-cli-go", "replace github.com/jfrog/jfrog-client-go => /path/to/mod/file"}, map[string]bool{"github.com/jfrog/jfrog-cli-go@v1.21.0": true},
			map[string]bool{"github.com/jfrog/jfrog-cli-go@v1.21.0": true, "github.com/jfrog/jfrog-client-go@v0.1.0": true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mergeReplaceDependenciesWithGraphDependencies(test.replaceDeps, test.graphDependencies)
			if !reflect.DeepEqual(test.expectedMap, test.graphDependencies) {
				t.Errorf("Test name: %s: Expected: %v, Got: %v", test.name, test.expectedMap, test.graphDependencies)
			}
		})
	}
}

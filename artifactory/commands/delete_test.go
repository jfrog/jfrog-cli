package commands

import (
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"strconv"
)

func TestReduceDirResult(t *testing.T) {
	paths := []utils.AqlSearchResultItem{}
	expected := []utils.AqlSearchResultItem{}

	paths = append(paths, utils.AqlSearchResultItem{Repo:"repo1", Path:"b", Name:"c"})
	expected = append(expected, utils.AqlSearchResultItem{Repo:"repo1", Path:"b", Name:"c/"})
	assertPackageFiles(expected, reduceDirResult(paths), t)

	paths = append(paths, utils.AqlSearchResultItem{Repo:"repo1", Path:"br", Name:"c"})
	expected = append(expected, utils.AqlSearchResultItem{Repo:"repo1", Path:"br", Name:"c/"})
	assertPackageFiles(expected, reduceDirResult(paths), t)

	paths = append(paths, utils.AqlSearchResultItem{Repo:"repo1", Path:"br/c/dont/care", Name:"somename"})
	assertPackageFiles(expected, reduceDirResult(paths), t)

	paths = append(paths, utils.AqlSearchResultItem{Repo:"repo1", Path:"bl", Name:"c1"})
	expected = append(expected, utils.AqlSearchResultItem{Repo:"repo1", Path:"bl", Name:"c1/"})
	assertPackageFiles(expected, reduceDirResult(paths), t)

	paths = append(paths, utils.AqlSearchResultItem{Repo:"repo1", Path:"bl/c1/you/dont/care", Name:"somename"})
	assertPackageFiles(expected, reduceDirResult(paths), t)

	paths = append(paths, utils.AqlSearchResultItem{Repo:"repo1", Path:"bl/c1/i/dont/care", Name:"somename"})
	assertPackageFiles(expected, reduceDirResult(paths), t)

	paths = append(paths, utils.AqlSearchResultItem{Repo:"repo1", Path:"b", Name:"."})
	assertPackageFiles(expected, reduceDirResult(paths), t)

	paths = append(paths, utils.AqlSearchResultItem{Repo:"repo2", Path:"bl", Name:"c1"})
	expected = append(expected, utils.AqlSearchResultItem{Repo:"repo2", Path:"bl", Name:"c1/"})
	assertPackageFiles(expected, reduceDirResult(paths), t)
}


func assertPackageFiles(expected, actual []utils.AqlSearchResultItem, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("Expected: " + strconv.Itoa(len(expected)) + ", Got: " + strconv.Itoa(len(actual)) + " files.")
	}

	expectedMap := make(map[string]utils.AqlSearchResultItem)
	for _, v := range expected {
		expectedMap[v.GetFullUrl()] = v
	}

	actualMap := make(map[string]utils.AqlSearchResultItem)
	for _, v := range actual {
		actualMap[v.GetFullUrl()] = v
	}

	for _, v := range actual {
		if _, ok := expectedMap[v.GetFullUrl()]; !ok {
			t.Error("Unexpected path:", v.GetFullUrl())
		}
	}
	for _, v := range expected {
		if _, ok := actualMap[v.GetFullUrl()]; !ok {
			t.Error("Path not found:", v.GetFullUrl())
		}
	}
}
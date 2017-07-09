package utils

import (
	"testing"
	"strconv"
)

func TestGetFullUrl(t *testing.T) {
	assertUrl("Repo", "some/path", "name", "Repo/some/path/name", t)
	assertUrl("", "some/path", "name", "some/path/name", t)
	assertUrl("Repo", "", "name", "Repo/name", t)
	assertUrl("Repo", "some/path", "", "Repo/some/path", t)
	assertUrl("", "some/path", "", "some/path", t)
	assertUrl("", "", "", "", t)
}

func assertUrl(repo, path, name, fullUrl string, t *testing.T) {
	testItem := AqlSearchResultItem{Repo:repo, Path: path, Name:name}
	if fullUrl != testItem.GetFullUrl() {
		t.Error("Unexpected URL built. Expected: `" + fullUrl + "` Got `" + testItem.GetFullUrl() + "`")
	}
}

func TestReduceDirResult(t *testing.T) {
	paths := []AqlSearchResultItem{}
	expected := []AqlSearchResultItem{}

	paths = append(paths, AqlSearchResultItem{Repo:"repo1", Path:"b", Name:"c/"})
	expected = append(expected, AqlSearchResultItem{Repo:"repo1", Path:"b", Name:"c/"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, AqlSearchResultItem{Repo:"repo1", Path:"br", Name:"c/"})
	expected = append(expected, AqlSearchResultItem{Repo:"repo1", Path:"br", Name:"c/"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, AqlSearchResultItem{Repo:"repo1", Path:"br/c/dont/care", Name:"somename"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, AqlSearchResultItem{Repo:"repo1", Path:"bl", Name:"c1/"})
	expected = append(expected, AqlSearchResultItem{Repo:"repo1", Path:"bl", Name:"c1/"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, AqlSearchResultItem{Repo:"repo1", Path:"bl/c1/you/dont/care", Name:"somename"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, AqlSearchResultItem{Repo:"repo1", Path:"bl/c1/i/dont/care", Name:"somename"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, AqlSearchResultItem{Repo:"repo1", Path:"b", Name:"."})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, AqlSearchResultItem{Repo:"repo2", Path:"bl", Name:"c1"})
	expected = append(expected, AqlSearchResultItem{Repo:"repo2", Path:"bl", Name:"c1"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)
}

func assertPackageFiles(expected, actual []AqlSearchResultItem, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("Expected: " + strconv.Itoa(len(expected)) + ", Got: " + strconv.Itoa(len(actual)) + " files.")
	}

	expectedMap := make(map[string]AqlSearchResultItem)
	for _, v := range expected {
		expectedMap[v.GetFullUrl()] = v
	}

	actualMap := make(map[string]AqlSearchResultItem)
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
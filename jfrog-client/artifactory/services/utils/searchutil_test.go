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
	testItem := ResultItem{Repo: repo, Path: path, Name: name}
	if fullUrl != testItem.GetItemRelativePath() {
		t.Error("Unexpected URL built. Expected: `" + fullUrl + "` Got `" + testItem.GetItemRelativePath() + "`")
	}
}

func TestReduceDirResult(t *testing.T) {
	paths := []ResultItem{}
	expected := []ResultItem{}

	paths = append(paths, ResultItem{Repo: "repo1", Path: "b", Name: "c/"})
	expected = append(expected, ResultItem{Repo: "repo1", Path: "b", Name: "c/"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, ResultItem{Repo: "repo1", Path: "br", Name: "c/"})
	expected = append(expected, ResultItem{Repo: "repo1", Path: "br", Name: "c/"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, ResultItem{Repo: "repo1", Path: "br/c/dont/care", Name: "somename"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, ResultItem{Repo: "repo1", Path: "bl", Name: "c1/"})
	expected = append(expected, ResultItem{Repo: "repo1", Path: "bl", Name: "c1/"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, ResultItem{Repo: "repo1", Path: "bl/c1/you/dont/care", Name: "somename"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, ResultItem{Repo: "repo1", Path: "bl/c1/i/dont/care", Name: "somename"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, ResultItem{Repo: "repo1", Path: "b", Name: "."})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)

	paths = append(paths, ResultItem{Repo: "repo2", Path: "bl", Name: "c1"})
	expected = append(expected, ResultItem{Repo: "repo2", Path: "bl", Name: "c1"})
	assertPackageFiles(expected, ReduceDirResult(paths, FilterTopChainResults), t)
}

func assertPackageFiles(expected, actual []ResultItem, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("Expected: " + strconv.Itoa(len(expected)) + ", Got: " + strconv.Itoa(len(actual)) + " files.")
	}

	expectedMap := make(map[string]ResultItem)
	for _, v := range expected {
		expectedMap[v.GetItemRelativePath()] = v
	}

	actualMap := make(map[string]ResultItem)
	for _, v := range actual {
		actualMap[v.GetItemRelativePath()] = v
	}

	for _, v := range actual {
		if _, ok := expectedMap[v.GetItemRelativePath()]; !ok {
			t.Error("Unexpected path:", v.GetItemRelativePath())
		}
	}
	for _, v := range expected {
		if _, ok := actualMap[v.GetItemRelativePath()]; !ok {
			t.Error("Path not found:", v.GetItemRelativePath())
		}
	}
}
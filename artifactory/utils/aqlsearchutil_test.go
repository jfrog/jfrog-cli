package utils

import (
	"testing"
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
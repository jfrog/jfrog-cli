package utils

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestBuildAqlSearchQueryRecursiveSimple(t *testing.T) {
	params := ArtifactoryCommonParams{Pattern: "repo-local", Target: "", Props: "", Build: "", Recursive: true, Regexp: false, IncludeDirs: false}
	aqlResult, _ := createAqlBodyForSpec(&params)
	expected := `{"repo": "repo-local","$or": [{"$and":[{"path": {"$match": "*"},"name": {"$match": "*"}}]}]}`

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryRecursiveWildcard(t *testing.T) {
	params := ArtifactoryCommonParams{Pattern: "repo-local2/a*b*c/dd/", Target: "", Props: "", Build: "", Recursive: true, Regexp: false, IncludeDirs: false}
	aqlResult, _ := createAqlBodyForSpec(&params)
	expected := `{"repo": "repo-local2","path": {"$ne": "."},"$or": [{"$and":[{"path": {"$match": "a*b*c/dd"},"name": {"$match": "*"}}]},{"$and":[{"path": {"$match": "a*b*c/dd/*"},"name": {"$match": "*"}}]}]}`
	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryNonRecursiveSimple(t *testing.T) {
	params := ArtifactoryCommonParams{Pattern: "repo-local", Target: "", Props: "", Build: "", Recursive: false, Regexp: false, IncludeDirs: false}
	aqlResult, _ := createAqlBodyForSpec(&params)
	expected := `{"repo": "repo-local","$or": [{"$and":[{"path": {"$match": "."},"name": {"$match": "*"}}]}]}`

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryNonRecursiveWildcard(t *testing.T) {
	specFile := ArtifactoryCommonParams{Pattern: "repo-local2/a*b*c/dd/", Target: "", Props: "", Build: "", Recursive: false, Regexp: false, IncludeDirs: false}
	aqlResult, _ := createAqlBodyForSpec(&specFile)
	expected := `{"repo": "repo-local2","path": {"$ne": "."},"$or": [{"$and":[{"path": {"$match": "a*b*c/dd"},"name": {"$match": "*"}}]}]}`

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestCreatePathFilePairs(t *testing.T) {
	pairs := []PathFilePair{{".", "a"}}
	validatePathPairs(createPathFilePairs("a", true), pairs, "a", t)
	pairs = []PathFilePair{{"a", "*"}, {"a/*", "*"}}
	validatePathPairs(createPathFilePairs("a/*", true), pairs, "a/*", t)
	pairs = []PathFilePair{{"a", "a*b"}, {"a/a*", "*b"}}
	validatePathPairs(createPathFilePairs("a/a*b", true), pairs, "a/a*b", t)
	pairs = []PathFilePair{{"a", "a*b*"}, {"a/a*", "*b*"}, {"a/a*b*", "*"}}
	validatePathPairs(createPathFilePairs("a/a*b*", true), pairs, "a/a*b*", t)
	pairs = []PathFilePair{{"a/a*b*/a", "b"}}
	validatePathPairs(createPathFilePairs("a/a*b*/a/b", true), pairs, "a/a*b*/a/b", t)
	pairs = []PathFilePair{{"*/a*", "*b*a*"}, {"*/a*/*", "*b*a*"}, {"*/a*/*b*", "*a*"}, {"*/a*/*b*a*", "*"}}
	validatePathPairs(createPathFilePairs("*/a*/*b*a*", true), pairs, "*/a*/*b*a*", t)
	pairs = []PathFilePair{{"*", "*"}}
	validatePathPairs(createPathFilePairs("*", true), pairs, "*", t)
	pairs = []PathFilePair{{"*", "*"}, {"*/*", "*"}}
	validatePathPairs(createPathFilePairs("*/*", true), pairs, "*/*", t)
	pairs = []PathFilePair{{"*", "a.z"}}
	validatePathPairs(createPathFilePairs("*/a.z", true), pairs, "*/a.z", t)
	pairs = []PathFilePair{{".", "a"}}
	validatePathPairs(createPathFilePairs("a", false), pairs, "a", t)
	pairs = []PathFilePair{{"", "*"}}
	validatePathPairs(createPathFilePairs("/*", false), pairs, "/*", t)
	pairs = []PathFilePair{{"", "a*b"}}
	validatePathPairs(createPathFilePairs("/a*b", false), pairs, "a*b", t)
}

func TestCreatePathFolderPairs(t *testing.T) {
	pairs := []PathFilePair{{"*", "*"}, {filepath.Join("*", "*"), "*"}}
	validatePathPairs(createPathFolderPairs("repo/*/*/"), pairs, "repo/*/*/", t)
	pairs = []PathFilePair{{".", "*"}, {"*", "*"}}
	validatePathPairs(createPathFolderPairs("repo/*/"), pairs, "repo/*/", t)
}

func validatePathPairs(actual, expected []PathFilePair, pattern string, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("Wrong path pairs for pattern: " + pattern)
	}
	for _, pair := range expected {
		found := false
		for _, testPair := range actual {
			if pair.path == testPair.path && pair.file == testPair.file {
				found = true
			}
		}
		if found == false {
			t.Error("Wrong path pairs for pattern: "+pattern+" , missing ", pair)
		}
	}
}

func TestArtifactoryCommonParams(t *testing.T) {
	artifactoryParams := ArtifactoryCommonParams{}
	assertIsSortLimitSpecBool(specIncludesSortOrLimit(&artifactoryParams), false, t)

	artifactoryParams.SortBy = []string{"Vava", "Bubu"}
	assertIsSortLimitSpecBool(specIncludesSortOrLimit(&artifactoryParams), true, t)

	artifactoryParams.SortBy = nil
	artifactoryParams.Limit = 0
	assertIsSortLimitSpecBool(specIncludesSortOrLimit(&artifactoryParams), false, t)

	artifactoryParams.Limit = -3
	assertIsSortLimitSpecBool(specIncludesSortOrLimit(&artifactoryParams), false, t)

	artifactoryParams.Limit = 3
	assertIsSortLimitSpecBool(specIncludesSortOrLimit(&artifactoryParams), true, t)

	artifactoryParams.SortBy = []string{"Vava", "Bubu"}
	assertIsSortLimitSpecBool(specIncludesSortOrLimit(&artifactoryParams), true, t)
}

func assertIsSortLimitSpecBool(actual, expected bool, t *testing.T) {
	if actual != expected {
		t.Error("The function specIncludesSortOrLimit() expected to return " + strconv.FormatBool(expected) + " but returned " + strconv.FormatBool(actual) + ".")
	}
}

func TestGetQueryReturnFields(t *testing.T) {
	artifactoryParams := ArtifactoryCommonParams{}
	minimalFields := []string{"name", "repo", "path", "actual_md5", "actual_sha1", "size", "type"}

	assertEqualFieldsList(getQueryReturnFields(&artifactoryParams, ALL), append(minimalFields, "property"), t)
	assertEqualFieldsList(getQueryReturnFields(&artifactoryParams, SYMLINK), append(minimalFields, "property"), t)
	assertEqualFieldsList(getQueryReturnFields(&artifactoryParams, NONE), append(minimalFields), t)

	artifactoryParams.SortBy = []string{"Vava"}
	assertEqualFieldsList(getQueryReturnFields(&artifactoryParams, NONE), append(minimalFields, "Vava"), t)
	assertEqualFieldsList(getQueryReturnFields(&artifactoryParams, ALL), append(minimalFields, "Vava"), t)
	assertEqualFieldsList(getQueryReturnFields(&artifactoryParams, SYMLINK), append(minimalFields, "Vava"), t)

	artifactoryParams.SortBy = []string{"Vava", "Bubu"}
	assertEqualFieldsList(getQueryReturnFields(&artifactoryParams, ALL), append(minimalFields, "Vava", "Bubu"), t)
}

func assertEqualFieldsList(actual, expected []string, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("The function getQueryReturnFields() expected to return the array:\n" + strings.Join(expected[:], ",") + ".\nbut returned:\n" + strings.Join(actual[:], ",") + ".")
	}
	for _, v := range actual {
		isFound := false
		for _, t := range expected {
			if v == t {
				isFound = true
				break
			}
		}
		if !isFound {
			t.Error("The function getQueryReturnFields() expected to return the array:\n'" + strings.Join(expected[:], ",") + "'.\nbut returned:\n'" + strings.Join(actual[:], ",") + "'.\n" +
				"The field " + v + "is missing!")
		}
	}
}

func TestBuildSortBody(t *testing.T) {
	assertSortBody(buildSortQueryPart([]string{"bubu"}, ""), `"$asc":["bubu"]`, t)
	assertSortBody(buildSortQueryPart([]string{"bubu", "kuku"}, ""), `"$asc":["bubu","kuku"]`, t)
}

func assertSortBody(actual, expected string, t *testing.T) {
	if actual != expected {
		t.Error("The function buildSortQueryPart expected to return the string:\n'" + expected + "'.\nbut returned:\n'" + actual + "'.")
	}
}

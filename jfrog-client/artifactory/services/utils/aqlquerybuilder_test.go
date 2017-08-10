package utils

import (
	"testing"
	"path/filepath"
)

func TestBuildAqlSearchQueryRecursiveSimple(t *testing.T) {
	params := ArtifactoryCommonParams{Pattern: "repo-local", Target: "", Props: "", Build: "", Recursive: true, Regexp: false, IncludeDirs: false}
	aqlResult, _ := createAqlBodyForItem(&params)
	expected := "{\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\": \"*\"},\"name\": {\"$match\": \"*\"}}]}]}"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryRecursiveWildcard(t *testing.T) {
	params := ArtifactoryCommonParams{Pattern: "repo-local2/a*b*c/dd/", Target: "", Props: "", Build: "", Recursive: true, Regexp: false, IncludeDirs: false}
	aqlResult, _ := createAqlBodyForItem(&params)
	expected := "{\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\": \"a*b*c/dd\"},\"path\": {\"$ne\": \".\"},\"name\": {\"$match\": \"*\"}}]},{\"$and\": [{\"path\": {\"$match\": \"a*b*c/dd/*\"},\"path\": {\"$ne\": \".\"},\"name\": {\"$match\": \"*\"}}]}]}"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryNonRecursiveSimple(t *testing.T) {
	params := ArtifactoryCommonParams{Pattern: "repo-local", Target: "", Props: "", Build: "", Recursive: false, Regexp: false, IncludeDirs: false}
	aqlResult, _ := createAqlBodyForItem(&params)
	expected := "{\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\": \".\"},\"name\": {\"$match\": \"*\"}}]}]}"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryNonRecursiveWildcard(t *testing.T) {
	specFile := ArtifactoryCommonParams{Pattern:"repo-local2/a*b*c/dd/", Target:"", Props:"", Build:"", Recursive:false, Regexp:false, IncludeDirs:false}
	aqlResult, _ := createAqlBodyForItem(&specFile)
	expected := "{\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\": \"a*b*c/dd\"},\"path\": {\"$ne\": \".\"},\"name\": {\"$match\": \"*\"}}]}]}"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestCreatePathFilePairs(t *testing.T) {
	pairs := []PathFilePair{{".","a"}}
	validatePathPairs(createPathFilePairs("a", true), pairs,"a", t)
	pairs = []PathFilePair{{"a","*"}, {"a/*","*"}}
	validatePathPairs(createPathFilePairs("a/*", true), pairs,"a/*",  t)
	pairs = []PathFilePair{{"a","a*b"}, {"a/a*","*b"}}
	validatePathPairs(createPathFilePairs("a/a*b", true), pairs,"a/a*b", t)
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
	pairs := []PathFilePair{{"*","*"}, {filepath.Join("*", "*"),"*"}}
	validatePathPairs(createPathFolderPairs("repo/*/*/"), pairs, "repo/*/*/", t)
	pairs = []PathFilePair{{".","*"}, {"*","*"}}
	validatePathPairs(createPathFolderPairs("repo/*/"), pairs, "repo/*/", t)
}

func validatePathPairs(actual, expected []PathFilePair, pattern string, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("Wrong path pairs for pattern: " + pattern)
	}
	for _, pair := range expected {
		found := false
		for _,testPair := range actual {
			if pair.path == testPair.path && pair.file== testPair.file {
				found = true
			}
		}
		if found == false {
			t.Error("Wrong path pairs for pattern: " + pattern + " , missing ", pair)
		}
	}
}
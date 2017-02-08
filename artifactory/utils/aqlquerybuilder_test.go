package utils

import (
	"testing"
)

func TestBuildAqlSearchQueryRecursiveSimple(t *testing.T) {
	specFile := CreateSpec("repo-local", "", "", "", true, true, false).Files[0]
	aqlResult, _ := createAqlBodyForItem(&specFile)
	expected := "{\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\": \"*\"},\"name\": {\"$match\": \"*\"}}]}]}"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryRecursiveWildcard(t *testing.T) {
	specFile := CreateSpec("repo-local2/a*b*c/dd/", "", "", "", true, true, false).Files[0]
	aqlResult, _ := createAqlBodyForItem(&specFile)
	expected := "{\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\": \"a*b*c/dd\"},\"name\": {\"$match\": \"*\"}}]},{\"$and\": [{\"path\": {\"$match\": \"a*b*c/dd/*\"},\"name\": {\"$match\": \"*\"}}]}]}"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryNonRecursiveSimple(t *testing.T) {
	specFile := CreateSpec("repo-local", "", "", "", false, true, false).Files[0]
	aqlResult, _ := createAqlBodyForItem(&specFile)
	expected := "{\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\": \".\"},\"name\": {\"$match\": \"*\"}}]}]}"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}

func TestBuildAqlSearchQueryNonRecursiveWildcard(t *testing.T) {
	specFile := CreateSpec("repo-local2/a*b*c/dd/", "", "", "", false, true, false).Files[0]
	aqlResult, _ := createAqlBodyForItem(&specFile)
	expected := "{\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\": \"a*b*c/dd\"},\"name\": {\"$match\": \"*\"}}]}]}"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. \nExpected: " + expected + " \nGot:      " + aqlResult)
	}
}
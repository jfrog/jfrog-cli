package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/tests"
	"testing"
)

func TestRecursiveDownload(t *testing.T) {
	flags := tests.GetFlags()
	flags.Recursive = true
	aqlResult := Download("repo-local", flags)
	expected := "items.find({\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"*\"},\"name\":{\"$match\":\"*\"}}]}]})"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}

	aqlResult = Download("repo-local2/a*b*c/dd/", flags)
	expected = "items.find({\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd\"},\"name\":{\"$match\":\"*\"}}]},{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd/*\"},\"name\":{\"$match\":\"*\"}}]}]})"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}
}

func TestNonRecursiveDownload(t *testing.T) {
	flags := tests.GetFlags()
	flags.Recursive = false
	aqlResult := Download("repo-local", flags)
	expected := "items.find({\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\".\"},\"name\":{\"$match\":\"*\"}}]}]})"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}

	aqlResult = Download("repo-local2/a*b*c/dd/", flags)
	expected = "items.find({\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd\"},\"name\":{\"$match\":\"*\"}}]}]})"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}
}

package utils

import "testing"

func TestBuildAqlSearchQueryRecursive(t *testing.T) {
	aqlResult := BuildAqlSearchQuery("repo-local", true, "")
	expected := "items.find({\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"*\"},\"name\":{\"$match\":\"*\"}}]}]}).include(\"name\",\"repo\",\"path\",\"actual_md5\",\"actual_sha1\",\"size\")"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}

	aqlResult = BuildAqlSearchQuery("repo-local2/a*b*c/dd/", true, "")
	expected = "items.find({\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd\"},\"name\":{\"$match\":\"*\"}}]},{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd/*\"},\"name\":{\"$match\":\"*\"}}]}]}).include(\"name\",\"repo\",\"path\",\"actual_md5\",\"actual_sha1\",\"size\")"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}
}

func TestBuildAqlSearchQueryNonRecursive(t *testing.T) {
	aqlResult := BuildAqlSearchQuery("repo-local", false, "")
	expected := "items.find({\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\".\"},\"name\":{\"$match\":\"*\"}}]}]}).include(\"name\",\"repo\",\"path\",\"actual_md5\",\"actual_sha1\",\"size\")"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}

	aqlResult = BuildAqlSearchQuery("repo-local2/a*b*c/dd/", false, "")
	expected = "items.find({\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd\"},\"name\":{\"$match\":\"*\"}}]}]}).include(\"name\",\"repo\",\"path\",\"actual_md5\",\"actual_sha1\",\"size\")"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}
}




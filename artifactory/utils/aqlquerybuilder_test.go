package utils

import "testing"

func TestBuildAqlSearchQueryRecursive(t *testing.T) {
	aqlResult, _ := BuildAqlSearchQuery("repo-local", true, "", []string{"\"name\"","\"repo\"","\"path\"","\"actual_md5\"","\"actual_sha1\"","\"size\""})
	expected := "items.find({\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"*\"},\"name\":{\"$match\":\"*\"}}]}]}).include(\"name\",\"repo\",\"path\",\"actual_md5\",\"actual_sha1\",\"size\")"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}

	aqlResult, _ = BuildAqlSearchQuery("repo-local2/a*b*c/dd/", true, "", []string{"\"name\"","\"repo\"","\"path\"","\"actual_md5\"","\"actual_sha1\"","\"size\""})
	expected = "items.find({\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd\"},\"name\":{\"$match\":\"*\"}}]},{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd/*\"},\"name\":{\"$match\":\"*\"}}]}]}).include(\"name\",\"repo\",\"path\",\"actual_md5\",\"actual_sha1\",\"size\")"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}
}

func TestBuildAqlSearchQueryNonRecursive(t *testing.T) {
	aqlResult, _ := BuildAqlSearchQuery("repo-local", false, "", []string{"\"name\"","\"repo\"","\"path\"","\"actual_md5\"","\"actual_sha1\"","\"size\""})
	expected := "items.find({\"repo\": \"repo-local\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\".\"},\"name\":{\"$match\":\"*\"}}]}]}).include(\"name\",\"repo\",\"path\",\"actual_md5\",\"actual_sha1\",\"size\")"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}

	aqlResult, _ = BuildAqlSearchQuery("repo-local2/a*b*c/dd/", false, "", []string{"\"name\"","\"repo\"","\"path\"","\"actual_md5\"","\"actual_sha1\"","\"size\""})
	expected = "items.find({\"repo\": \"repo-local2\",\"$or\": [{\"$and\": [{\"path\": {\"$match\":\"a*b*c/dd\"},\"name\":{\"$match\":\"*\"}}]}]}).include(\"name\",\"repo\",\"path\",\"actual_md5\",\"actual_sha1\",\"size\")"

	if aqlResult != expected {
		t.Error("Unexpected download AQL query built. Expected: " + expected + " Got " + aqlResult)
	}
}




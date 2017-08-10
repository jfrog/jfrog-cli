package utils

import (
	"testing"
)

func TestMatchingDelete(t *testing.T) {
	var actual string
	actual, _ = WildcardToDirsPath("s/*/path/", "s/a/path/b.zip")
	assertDeletePattern("s/a/path/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/path/", "s/a/b/c/path/b.zip")
	assertDeletePattern("s/a/b/c/path/", actual, t)
	actual, _ = WildcardToDirsPath("s/a/*/", "s/a/b/path/b.zip")
	assertDeletePattern("s/a/b/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/path/*/", "s/a/path/a/b.zip")
	assertDeletePattern("s/a/path/a/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/path/*/", "s/a/a/path/a/b/c/d/b.zip")
	assertDeletePattern("s/a/a/path/a/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/", "s/a/a/path/a/b/c/d/b.zip")
	assertDeletePattern("s/a/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/a/*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern("s/a/a/path/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/a/*/*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern("s/a/a/path/k/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/*l*/*/*/", "s/a/l/path/k/b/c/d/b.zip")
	assertDeletePattern("s/a/l/path/k/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/a*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern("s/a/a/", actual, t)
	actual, _ = WildcardToDirsPath("s/a*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern("s/a/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePattern("s/a/", actual, t)
	actual, _ = WildcardToDirsPath("s/*/*path*/", "s/a/h/path/k/b/c/d/b.zip")
	assertDeletePattern("s/a/h/path/", actual, t)
	actual, _ = WildcardToDirsPath("a/b/*********/*******/", "a/b/c/d/e.zip")
	assertDeletePattern("a/b/c/d/", actual, t)
	actual, err := WildcardToDirsPath("s/*/a/*/*", "s/a/a/path/k/b/c/d/b.zip")
	assertDeletePatternErr(err.Error(), "Delete pattern must end with \"/\"", t)
}

func assertDeletePattern(expected, actual string, t *testing.T) {
	if expected != actual {
		t.Error("Wrong matching expected: `" + expected + "` Got `" + actual + "`")
	}
}

func assertDeletePatternErr(expected, actual string, t *testing.T) {
	if expected != actual {
		t.Error("Wrong err message expected: `" + expected + "` Got `" + actual + "`")
	}
}

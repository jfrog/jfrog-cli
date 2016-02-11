package commands

import (
    "testing"
    "strconv"
    "github.com/jfrogdev/jfrog-cli-go/tests"
    "github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

func TestSingleFileUpload(t *testing.T) {
    flags := tests.GetFlags()
    uploaded1, _ := Upload("testdata/a.txt", "repo-local", flags)
    uploaded2, _ := Upload("testdata/aa.txt", "repo-local", flags)
    uploaded3, _ := Upload("testdata/aa1*.txt", "repo-local", flags)
    if uploaded1 != 1 {
        t.Error("Expected 1 file to be uploaded. Got " + strconv.Itoa(uploaded1) + ".")
    }
    if uploaded2 != 1 {
        t.Error("Expected 1 file to be uploaded. Got " + strconv.Itoa(uploaded2) + ".")
    }
    if uploaded3 != 0 {
        t.Error("Expected 0 file to be uploaded. Got " + strconv.Itoa(uploaded3) + ".")
    }
}

func TestPatternRecursiveUpload(t *testing.T) {
    flags := tests.GetFlags()
    flags.Recursive = true
    testPatternUpload(t, flags)
}

func TestPatternNonRecursiveUpload(t *testing.T) {
    flags := tests.GetFlags()
    flags.Recursive = false
    testPatternUpload(t, flags)
}

func testPatternUpload(t *testing.T, flags *utils.Flags) {
    sep := tests.GetFileSeperator()
    uploaded1, _ := Upload("testdata" + sep + "*", "repo-local", flags)
    uploaded2, _ := Upload("testdata" + sep + "a*", "repo-local", flags)
    uploaded3, _ := Upload("testdata" + sep + "b*", "repo-local", flags)

    if uploaded1 != 3 {
        t.Error("Expected 3 file to be uploaded. Got " + strconv.Itoa(uploaded1) + ".")
    }
    if uploaded2 != 2 {
        t.Error("Expected 2 file to be uploaded. Got " + strconv.Itoa(uploaded2) + ".")
    }
    if uploaded3 != 1 {
        t.Error("Expected 1 file to be uploaded. Got " + strconv.Itoa(uploaded3) + ".")
    }
}

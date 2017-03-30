package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"strconv"
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

func TestSingleFileUpload(t *testing.T) {
	flags := getUploadFlags()
	spec := utils.CreateSpec("testdata/a.txt", "repo-local", "", "", false, true, false)
	uploaded1, _, err := Upload(spec, flags)
	checkUploaded(t, uploaded1, err, 1)

	spec = utils.CreateSpec("testdata/aa.txt", "repo-local", "", "", false, true, false)
	uploaded2, _, err := Upload(spec, flags)
	checkUploaded(t, uploaded2, err, 1)

	spec = utils.CreateSpec("testdata/aa1*.txt", "repo-local", "", "", false, true, false)
	uploaded3, _, err := Upload(spec, flags)
	checkUploaded(t, uploaded3, err, 0)

}

func TestPatternRecursiveUpload(t *testing.T) {
	flags := getUploadFlags()
	testPatternUpload(t, true, flags)
}

func TestPatternNonRecursiveUpload(t *testing.T) {
	flags := getUploadFlags()
	testPatternUpload(t, false, flags)
}

func testPatternUpload(t *testing.T, recursive bool, flags *UploadFlags) {
	sep := cliutils.GetTestsFileSeperator()
	spec := utils.CreateSpec("testdata" + sep + "*", "repo-local", "", "", recursive, true, false)
	uploaded1, _, err := Upload(spec, flags)
	checkUploaded(t, uploaded1, err, 3)

	spec = utils.CreateSpec("testdata" + sep + "a*", "repo-local", "", "", recursive, true, false)
	uploaded2, _, err := Upload(spec, flags)
	checkUploaded(t, uploaded2, err, 2)

	spec = utils.CreateSpec("testdata" + sep + "b*", "repo-local", "", "", recursive, true, false)
	uploaded3, _, err := Upload(spec, flags)
	checkUploaded(t, uploaded3, err, 1)
}

func getUploadFlags() *UploadFlags {
	flags := new(UploadFlags)
	flags.ArtDetails = new(config.ArtifactoryDetails)
	flags.DryRun = true
	flags.Threads = 3

	return flags
}

func checkUploaded(t *testing.T, expected int, err error, actual int) {
	if err != nil {
		t.Error(err.Error())
	}
	if expected != actual {
		t.Error("Expected ", actual, " file to be uploaded. Got ", strconv.Itoa(actual), ".")
	}
}
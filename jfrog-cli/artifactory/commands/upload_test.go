package commands

//
//import (
//	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
//	"strconv"
//	"testing"
//	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
//	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
//	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/jfrog-client/services/artifactory"
//)
//
//func TestSingleFileUpload(t *testing.T) {
//	flags := getUploadFlags()
//	artDetails := new(config.ArtifactoryDetails)
//	spec := utils.CreateSpec("testdata/a.txt", "repo-local", "", "", false, true, false, false)
//	uploaded1, _, err := Upload(spec, flags, "", "" , artDetails)
//	checkUploaded(t, uploaded1, err, 1)
//
//	spec = utils.CreateSpec("testdata/aa.txt", "repo-local", "", "", false, true, false, false)
//	uploaded2, _, err := Upload(spec, flags, "", "" , artDetails)
//	checkUploaded(t, uploaded2, err, 1)
//
//	spec = utils.CreateSpec("testdata/aa1*.txt", "repo-local", "", "", false, true, false, false)
//	uploaded3, _, err := Upload(spec, flags, "", "" , artDetails)
//	checkUploaded(t, uploaded3, err, 0)
//
//}
//
//func TestPatternRecursiveUpload(t *testing.T) {
//	flags := getUploadFlags()
//	testPatternUpload(t, true, flags)
//}
//
//func TestPatternNonRecursiveUpload(t *testing.T) {
//	flags := getUploadFlags()
//	testPatternUpload(t, false, flags)
//}
//
//func testPatternUpload(t *testing.T, recursive bool, flags *artifactory.UploadConfiguration) {
//	sep := cliutils.GetTestsFileSeperator()
//	artDetails := new(config.ArtifactoryDetails)
//	spec := utils.CreateSpec("testdata" + sep + "*", "repo-local", "", "", recursive, true, false, false)
//	uploaded1, _, err := Upload(spec, flags, "", "" , artDetails)
//	checkUploaded(t, uploaded1, err, 3)
//
//	spec = utils.CreateSpec("testdata" + sep + "a*", "repo-local", "", "", recursive, true, false, false)
//	uploaded2, _, err := Upload(spec, flags, "", "" , artDetails)
//	checkUploaded(t, uploaded2, err, 2)
//
//	spec = utils.CreateSpec("testdata" + sep + "b*", "repo-local", "", "", recursive, true, false, false)
//	uploaded3, _, err := Upload(spec, flags, "", "" , artDetails)
//	checkUploaded(t, uploaded3, err, 1)
//}
//
//func getUploadFlags() *artifactory.UploadConfiguration {
//	flags := new(artifactory.UploadConfiguration)
//	flags.DryRun = true
//	flags.Threads = 3
//	return flags
//}
//
//func checkUploaded(t *testing.T, expected int, err error, actual int) {
//	if err != nil {
//		t.Error(err.Error())
//	}
//	if expected != actual {
//		t.Error("Expected ", actual, " file to be uploaded. Got ", strconv.Itoa(actual), ".")
//	}
//}

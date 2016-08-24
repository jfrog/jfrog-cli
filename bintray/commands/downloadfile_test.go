package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
    "github.com/jfrogdev/jfrog-cli-go/bintray/tests"
	"testing"
)

func TestDownloadFile(t *testing.T) {
    bintrayDetails := tests.CreateBintrayDetails()
    pathStr := "test-subject/test-repo/a/b/c/file.zip"
    pathDetails, err := utils.CreatePathDetails(pathStr)
	if err != nil {
		t.Error(err.Error())
	}
    expected := "https://dl.bintray.com/" + pathStr
    url := utils.BuildDownloadBintrayFileUrl(bintrayDetails, pathDetails)
    if expected != url {
        t.Error("Got unexpected url from BuildDownloadBintrayFileUrl. Expected: " + expected + " Got " + url)
    }

    pathStr = "test-subject/test-repo/file.zip"
    pathDetails, err = utils.CreatePathDetails(pathStr)
	if err != nil {
		t.Error(err.Error())
	}
    expected = "https://dl.bintray.com/test-subject/test-repo/file.zip"
    url = utils.BuildDownloadBintrayFileUrl(bintrayDetails, pathDetails)
    if expected != url {
        t.Error("Got unexpected url from BuildDownloadBintrayFileUrl. Expected: " + expected + " Got " + url)
    }
}
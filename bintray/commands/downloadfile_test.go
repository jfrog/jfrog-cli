package commands

import (
	"github.com/JFrogDev/jfrog-cli-go/bintray/utils"
    "github.com/JFrogDev/jfrog-cli-go/bintray/tests"
	"testing"
)

func TestDownloadFile(t *testing.T) {
    bintrayDetails := tests.CreateBintrayDetails()
    versionDetails := utils.CreateVersionDetails("test-subject/test-repo/test-package/ver-1.2")
    expected := "https://dl.bintray.com/test-subject/test-repo/a/b/c/file.zip"
    url := utils.BuildDownloadBintrayFileUrl(bintrayDetails, versionDetails, "a/b/c/file.zip")
    if expected != url {
        t.Error("Got unexpected url from BuildDownloadBintrayFileUrl. Expected: " + expected + " Got " + url)
    }

    expected = "https://dl.bintray.com/test-subject/test-repo/file.zip"
    url = utils.BuildDownloadBintrayFileUrl(bintrayDetails, versionDetails, "file.zip")
    if expected != url {
        t.Error("Got unexpected url from BuildDownloadBintrayFileUrl. Expected: " + expected + " Got " + url)
    }
}
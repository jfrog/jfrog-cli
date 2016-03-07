package commands

import (
	"github.com/JFrogDev/jfrog-cli-go/bintray/utils"
    "github.com/JFrogDev/jfrog-cli-go/bintray/tests"
	"testing"
)

func TestDownloadVersion(t *testing.T) {
    versionDetails := utils.CreateVersionDetails("test-subject/test-repo/test-package/ver-1.2")
    url := BuildDownloadVersionUrl(versionDetails, tests.CreateBintrayDetails())
    expected := "https://api.bintray.com/packages/test-subject/test-repo/test-package/versions/ver-1.2/files"
    if expected != url {
        t.Error("Got unexpected url from BuildDownloadVersionUrl. Expected: " + expected + " Got " + url)
    }
}
package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
    "github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/tests"
	"testing"
)

func TestDownloadVersion(t *testing.T) {
    versionDetails, err := utils.CreateVersionDetails("test-subject/test-repo/test-package/ver-1.2")
	if err != nil {
		t.Error(err.Error())
	}
    url := BuildDownloadVersionUrl(versionDetails, tests.CreateBintrayDetails(), false)
    expected := "https://api.bintray.com/packages/test-subject/test-repo/test-package/versions/ver-1.2/files"
    if expected != url {
        t.Error("Got unexpected url from BuildDownloadVersionUrl. Expected: " + expected + " Got " + url)
    }
}
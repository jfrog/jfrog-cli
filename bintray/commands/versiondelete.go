package commands

import (
	"fmt"
	"github.com/JFrogDev/jfrog-cli-go/bintray/utils"
	"github.com/JFrogDev/jfrog-cli-go/cliutils"
)

func DeleteVersion(versionDetails *utils.VersionDetails, bintrayDetails *cliutils.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

	fmt.Println("Deleting version: " + versionDetails.Version)
	resp, body := cliutils.SendDelete(url, bintrayDetails.User, bintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
}

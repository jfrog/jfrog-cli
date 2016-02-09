package commands

import (
    "fmt"
    "github.com/jFrogdev/jfrog-cli-go/cliutils"
    "github.com/jFrogdev/jfrog-cli-go/bintray/utils"
)

func DeleteVersion(versionDetails *utils.VersionDetails, bintrayDetails *utils.BintrayDetails) {
    if bintrayDetails.User == "" {
        bintrayDetails.User = versionDetails.Subject
    }
    url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

    fmt.Println("Deleting version: " + versionDetails.Version)
    resp, body := cliutils.SendDelete(url, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
}
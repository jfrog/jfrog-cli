package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/cliutils"
    "github.com/JFrogDev/bintray-cli-go/bintray/utils"
)

func ShowVersion(versionDetails *utils.VersionDetails, bintrayDetails *utils.BintrayDetails) {
    if bintrayDetails.User == "" {
        bintrayDetails.User = versionDetails.Subject
    }
    var message string
    if versionDetails.Version == "" {
        versionDetails.Version = "_latest"
        message = "Getting latest version"
    } else {
        message = "Getting version: " + versionDetails.Version
    }
    url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

    fmt.Println(message)
    resp, body := cliutils.SendGet(url, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode == 200 {
        fmt.Println(cliutils.IndentJson(body))
    } else {
        cliutils.Exit(cliutils.ExitCodeError, "Bintray response: " + resp.Status)
    }
}
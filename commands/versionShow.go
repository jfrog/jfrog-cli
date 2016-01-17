package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
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
    resp, body := utils.SendGet(url, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode == 200 {
        fmt.Println(utils.IndentJson(body))
    } else {
        utils.Exit("Bintray response: " + resp.Status)
    }
}
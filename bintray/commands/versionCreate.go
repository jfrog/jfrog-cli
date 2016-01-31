package commands

import (
    "fmt"
    "net/http"
    "github.com/JFrogDev/jfrog-cli-go/cliutils"
    "github.com/JFrogDev/jfrog-cli-go/bintray/utils"
)

func CreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) {
    fmt.Println("Creating version: " + versionDetails.Version)
    resp, body := DoCreateVersion(versionDetails, flags)
    if resp.StatusCode != 201 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}

func DoCreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) (*http.Response, []byte) {
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = versionDetails.Subject
    }
    data := utils.CreateVersionJson(versionDetails.Version, flags)

    url := flags.BintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions"

    return cliutils.SendPost(url, nil, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
}
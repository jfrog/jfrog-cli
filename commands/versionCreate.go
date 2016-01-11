package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func CreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) {
    data := utils.CreateVersionJson(versionDetails.Version, flags)
    url := flags.BintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions"

    fmt.Println("Creating version: " + versionDetails.Version)
    resp, body := utils.SendPost(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 201 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}
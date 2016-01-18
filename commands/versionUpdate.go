package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func UpdateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) {
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = versionDetails.Subject
    }
    data := utils.CreateVersionJson(versionDetails.Version, flags)
    url := flags.BintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

    fmt.Println("Updating version: " + versionDetails.Version)
    resp, body := utils.SendPatch(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}
package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func PublishVersion(versionDetails *utils.VersionDetails, bintrayDetails *utils.BintrayDetails) {
    url := bintrayDetails.ApiUrl + "content/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/" +
        versionDetails.Version + "/publish"

    fmt.Println("Publishing version: " + versionDetails.Version)
    resp, body := utils.SendPost(url, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}
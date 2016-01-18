package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func GpgSignVersion(versionDetails *utils.VersionDetails, passphrase string, bintrayDetails *utils.BintrayDetails) {
    if bintrayDetails.User == "" {
        bintrayDetails.User = versionDetails.Subject
    }
    url := bintrayDetails.ApiUrl + "gpg/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package +
        "/versions/" + versionDetails.Version

    var data string
    if passphrase != "" {
        data = "{ \"passphrase\": \"" + passphrase + "\" }"
    }

    fmt.Println("GPG signing version: " + versionDetails.Version)
    resp, body := utils.SendPost(url, []byte(data), bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}

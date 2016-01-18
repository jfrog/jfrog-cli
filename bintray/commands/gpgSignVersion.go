package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/cliutils"
    "github.com/JFrogDev/bintray-cli-go/bintray/utils"
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
    resp, body := cliutils.SendPost(url, []byte(data), bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}

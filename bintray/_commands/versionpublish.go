package commands

import (
    "fmt"
    "github.com/jFrogdev/jfrog-cli-go/cliutils"
    "github.com/jFrogdev/jfrog-cli-go/bintray/utils"
)

func PublishVersion(versionDetails *utils.VersionDetails, bintrayDetails *utils.BintrayDetails) {
    if bintrayDetails.User == "" {
        bintrayDetails.User = versionDetails.Subject
    }
    url := bintrayDetails.ApiUrl + "content/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/" +
        versionDetails.Version + "/publish"

    fmt.Println("Publishing version: " + versionDetails.Version)
    resp, body := cliutils.SendPost(url, nil, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}
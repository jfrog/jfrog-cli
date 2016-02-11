package commands

import (
    "fmt"
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
    "github.com/jfrogdev/jfrog-cli-go/bintray/utils"
)

func GpgSignFile(pathDetails *utils.PathDetails, passphrase string, bintrayDetails *utils.BintrayDetails) {
    if bintrayDetails.User == "" {
        bintrayDetails.User = pathDetails.Subject
    }
    url := bintrayDetails.ApiUrl + "gpg/" + pathDetails.Subject + "/" +
        pathDetails.Repo + "/" + pathDetails.Path

    var data string
    if passphrase != "" {
        data = "{ \"passphrase\": \"" + passphrase + "\" }"
    }

    fmt.Println("GPG signing file: " + pathDetails.Path)
    resp, body := cliutils.SendPost(url, nil, []byte(data), bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(cliutils.IndentJson(body))
}

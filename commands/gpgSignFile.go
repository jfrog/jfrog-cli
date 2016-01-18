package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
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
    resp, body := utils.SendPost(url, []byte(data), bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}

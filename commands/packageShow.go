package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func ShowPackage(packageDetails *utils.VersionDetails, bintrayDetails *utils.BintrayDetails) {
    url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
        packageDetails.Repo + "/" + packageDetails.Package

    fmt.Println("Getting package: " + packageDetails.Package)
    resp, body := utils.SendGet(url, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode == 200 {
        fmt.Println(utils.IndentJson(body))
    } else {
        utils.Exit("Bintray response: " + resp.Status)
    }
}
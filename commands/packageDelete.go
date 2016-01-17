package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func DeletePackage(packageDetails *utils.VersionDetails, bintrayDetails *utils.BintrayDetails) {
    if bintrayDetails.User == "" {
        bintrayDetails.User = packageDetails.Subject
    }
    url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
        packageDetails.Repo + "/" + packageDetails.Package

    fmt.Println("Deleting package: " + packageDetails.Package)
    resp, body := utils.SendDelete(url, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
}
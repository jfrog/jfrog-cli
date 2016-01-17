package commands

import (
    "fmt"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func UpdatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) {
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = packageDetails.Subject
    }
    data := utils.CreatePackageJson(packageDetails.Package, flags)
    url := flags.BintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
        packageDetails.Repo + "/" + packageDetails.Package

    fmt.Println("Updating package: " + packageDetails.Package)
    resp, body := utils.SendPatch(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}
package commands

import (
    "fmt"
    "net/http"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func CreatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) {
    fmt.Println("Creating package: " + packageDetails.Package)
    resp, body := DoCreatePackage(packageDetails, flags)
    fmt.Println("Bintray response: " + resp.Status)
    fmt.Println(utils.IndentJson(body))
}

func DoCreatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) (*http.Response, []byte) {
    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = packageDetails.Subject
    }
    data := utils.CreatePackageJson(packageDetails.Package, flags)
    url := flags.BintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
        packageDetails.Repo

    return utils.SendPost(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
}

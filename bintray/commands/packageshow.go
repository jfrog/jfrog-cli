package commands

import (
    "fmt"
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
    "github.com/jfrogdev/jfrog-cli-go/bintray/utils"
)

func ShowPackage(packageDetails *utils.VersionDetails, bintrayDetails *cliutils.BintrayDetails) {
    if bintrayDetails.User == "" {
        bintrayDetails.User = packageDetails.Subject
    }
    url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
        packageDetails.Repo + "/" + packageDetails.Package

    fmt.Println("Getting package: " + packageDetails.Package)
    resp, body := cliutils.SendGet(url, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode == 200 {
        fmt.Println(cliutils.IndentJson(body))
    } else {
        cliutils.Exit(cliutils.ExitCodeError, "Bintray response: " + resp.Status)
    }
}
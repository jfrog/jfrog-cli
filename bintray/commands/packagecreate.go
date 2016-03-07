package commands

import (
	"fmt"
	"github.com/JFrogDev/jfrog-cli-go/bintray/utils"
	"github.com/JFrogDev/jfrog-cli-go/cliutils"
	"net/http"
)

func CreatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) {
	fmt.Println("Creating package: " + packageDetails.Package)
	resp, body := DoCreatePackage(packageDetails, flags)
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func DoCreatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) (*http.Response, []byte) {
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = packageDetails.Subject
	}
	data := utils.CreatePackageJson(packageDetails.Package, flags)
	url := flags.BintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo

	return cliutils.SendPost(url, nil, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
}

package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
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
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	return ioutils.SendPost(url, []byte(data), httpClientsDetails)
}

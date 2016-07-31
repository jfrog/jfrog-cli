package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func CreatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) {
	logger.Logger.Info("Creating package: " + packageDetails.Package)
	resp, body := DoCreatePackage(packageDetails, flags)
	logger.Logger.Info("Bintray response: " + resp.Status)
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

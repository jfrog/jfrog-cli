package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

func CreatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) error {
	log.Info("Creating package...")
	resp, body, err := DoCreatePackage(packageDetails, flags)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func DoCreatePackage(packageDetails *utils.VersionDetails, flags *utils.PackageFlags) (*http.Response, []byte, error) {
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = packageDetails.Subject
	}
	data := utils.CreatePackageJson(packageDetails.Package, flags)
	url := flags.BintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
			packageDetails.Repo
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	return httputils.SendPost(url, []byte(data), httpClientsDetails)
}

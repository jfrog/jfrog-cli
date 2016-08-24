package commands

import (
    "errors"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func DeletePackage(packageDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) error {
	if bintrayDetails.User == "" {
		bintrayDetails.User = packageDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package

	logger.Logger.Info("Deleting package: " + packageDetails.Package)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, err := ioutils.SendDelete(url, nil, httpClientsDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadBintrayMessage(body)))
        if err != nil {
            return err
        }
	}
	logger.Logger.Info("Bintray response: " + resp.Status)
	return nil
}

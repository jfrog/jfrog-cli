package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func DeleteVersion(versionDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

	logger.Logger.Info("Deleting version: " + versionDetails.Version)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body := ioutils.SendDelete(url, nil, httpClientsDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	logger.Logger.Info("Bintray response: " + resp.Status)
}

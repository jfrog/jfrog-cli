package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func ShowPackage(packageDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = packageDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package

	logger.Logger.Info("Getting package: " + packageDetails.Package)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := ioutils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode == 200 {
		fmt.Println(cliutils.IndentJson(body))
	} else {
		cliutils.Exit(cliutils.ExitCodeError, "Bintray response: "+resp.Status)
	}
}

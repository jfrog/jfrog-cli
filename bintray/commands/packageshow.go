package commands

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"fmt"
)

func ShowPackage(packageDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) (err error) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = packageDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
			packageDetails.Repo + "/" + packageDetails.Package

	log.Info("Getting package details...")
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := httputils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Package", packageDetails.Package, "details:")
	fmt.Println(cliutils.IndentJson(body))
	return
}

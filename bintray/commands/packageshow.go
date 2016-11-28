package commands

import (
	"errors"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func ShowPackage(packageDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) (err error) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = packageDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package

	log.Info("Getting package:", packageDetails.Package)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := ioutils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode == 200 {
		fmt.Println(cliutils.IndentJson(body))
	} else {
		err = cliutils.CheckError(errors.New("Bintray response: "+resp.Status))
	}
	return
}

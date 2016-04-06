package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func ShowPackage(packageDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = packageDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package

	fmt.Println("Getting package: " + packageDetails.Package)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := ioutils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode == 200 {
		fmt.Println(cliutils.IndentJson(body))
	} else {
		cliutils.Exit(cliutils.ExitCodeError, "Bintray response: "+resp.Status)
	}
}

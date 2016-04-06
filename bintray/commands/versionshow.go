package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func ShowVersion(versionDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	var message string
	if versionDetails.Version == "" {
		versionDetails.Version = "_latest"
		message = "Getting latest version"
	} else {
		message = "Getting version: " + versionDetails.Version
	}
	url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

	fmt.Println(message)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := ioutils.SendGet(url, true, httpClientsDetails)
	if resp.StatusCode == 200 {
		fmt.Println(cliutils.IndentJson(body))
	} else {
		cliutils.Exit(cliutils.ExitCodeError, "Bintray response: "+resp.Status)
	}
}

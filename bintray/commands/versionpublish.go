package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func PublishVersion(versionDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "content/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/" +
		versionDetails.Version + "/publish"

	fmt.Println("Publishing version: " + versionDetails.Version)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body := ioutils.SendPost(url, nil, httpClientsDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func GpgSignVersion(versionDetails *utils.VersionDetails, passphrase string, bintrayDetails *config.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "gpg/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package +
		"/versions/" + versionDetails.Version

	var data string
	if passphrase != "" {
		data = "{ \"passphrase\": \"" + passphrase + "\" }"
	}

	logger.Logger.Info("GPG signing version: " + versionDetails.Version)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body := ioutils.SendPost(url, []byte(data), httpClientsDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	logger.Logger.Info("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func CreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) {
	logger.Logger.Info("Creating version: " + versionDetails.Version)
	resp, body := doCreateVersion(versionDetails, flags, flags.BintrayDetails)
	if resp.StatusCode != 201 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	logger.Logger.Info("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func DoCreateVersion(versionDetails *utils.VersionDetails,
    bintrayDetails *config.BintrayDetails) (*http.Response, []byte) {
    return doCreateVersion(versionDetails, nil, bintrayDetails)
}

func doCreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags,
    bintrayDetails *config.BintrayDetails) (*http.Response, []byte) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	var data string
	if flags != nil {
        data = utils.CreateVersionJson(versionDetails.Version, flags)
	} else {
		m := map[string]string{
    		"name": versionDetails.Version}
        data = cliutils.MapToJson(m)
	}

	url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/versions"
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	return ioutils.SendPost(url, []byte(data), httpClientsDetails)
}

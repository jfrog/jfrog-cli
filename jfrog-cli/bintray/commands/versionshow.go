package commands

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

func ShowVersion(versionDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) error {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	if versionDetails.Version == "" {
		versionDetails.Version = "_latest"
	}

	url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
			versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

	log.Info("Getting version details...")
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := httputils.SendGet(url, true, httpClientsDetails)

	if resp.StatusCode != 200 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

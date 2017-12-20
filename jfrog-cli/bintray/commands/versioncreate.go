package commands

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

func CreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) error {
	log.Info("Creating version...")
	resp, body, err := doCreateVersion(versionDetails, flags, flags.BintrayDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}
	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func DoCreateVersion(versionDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) (*http.Response, []byte, error) {
	return doCreateVersion(versionDetails, nil, bintrayDetails)
}

func doCreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags, bintrayDetails *config.BintrayDetails) (*http.Response, []byte, error) {
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
	return httputils.SendPost(url, []byte(data), httpClientsDetails)
}

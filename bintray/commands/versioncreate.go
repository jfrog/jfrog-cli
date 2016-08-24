package commands

import (
	"errors"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func CreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) error {
	logger.Logger.Info("Creating version: " + versionDetails.Version)
	resp, body, err := doCreateVersion(versionDetails, flags, flags.BintrayDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode != 201 {
		err = cliutils.CheckError(errors.New(resp.Status + ". "+utils.ReadBintrayMessage(body)))
        if err != nil {
            return err
        }
	}
	logger.Logger.Info("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

func DoCreateVersion(versionDetails *utils.VersionDetails,
    bintrayDetails *config.BintrayDetails) (*http.Response, []byte, error) {
    return doCreateVersion(versionDetails, nil, bintrayDetails)
}

func doCreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags,
    bintrayDetails *config.BintrayDetails) (*http.Response, []byte, error) {
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

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

func ShowVersion(versionDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) error {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	version := versionDetails.Version
	if versionDetails.Version == "" {
		versionDetails.Version = "_latest"
		version = "latest"
	}

	url := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
			versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

	log.Info("Getting version details...")
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, _, _ := httputils.SendGet(url, true, httpClientsDetails)

	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Version", version, "details:")
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

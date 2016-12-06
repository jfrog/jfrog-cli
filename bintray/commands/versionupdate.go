package commands

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func UpdateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) error {
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = versionDetails.Subject
	}
	data := utils.CreateVersionJson(versionDetails.Version, flags)
	url := flags.BintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
			versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

	log.Info("Updating version...")
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, err := ioutils.SendPatch(url, []byte(data), httpClientsDetails)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Updated version", versionDetails.Version + ".")
	return nil
}

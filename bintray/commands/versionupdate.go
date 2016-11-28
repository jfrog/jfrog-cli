package commands

import (
    "errors"
	"fmt"
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

	log.Info("Updating version:", versionDetails.Version)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, err := ioutils.SendPatch(url, []byte(data), httpClientsDetails)
    if err != nil {
        return err
    }
	if resp.StatusCode != 200 {
		err := cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadBintrayMessage(body)))
        if err != nil {
            return err
        }
	}
	log.Info("Bintray response:", resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

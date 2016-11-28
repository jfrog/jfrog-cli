package commands

import (
    "errors"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func PublishVersion(versionDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails) error {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "content/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/" +
		versionDetails.Version + "/publish"

	log.Info("Publishing version:", versionDetails.Version)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, err := ioutils.SendPost(url, nil, httpClientsDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadBintrayMessage(body)))
        if err != nil {
            return err
        }
	}
	log.Info("Bintray response:", resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

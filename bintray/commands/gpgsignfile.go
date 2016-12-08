package commands

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"fmt"
)

func GpgSignFile(pathDetails *utils.PathDetails, passphrase string, bintrayDetails *config.BintrayDetails) error {
	if bintrayDetails.User == "" {
		bintrayDetails.User = pathDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "gpg/" + pathDetails.Subject + "/" +
			pathDetails.Repo + "/" + pathDetails.Path

	var data string
	if passphrase != "" {
		data = "{ \"passphrase\": \"" + passphrase + "\" }"
	}

	log.Info("GPG signing file...")
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, err := ioutils.SendPost(url, []byte(data), httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("GPG signed file", pathDetails.Path, ", details:")
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func GpgSignFile(pathDetails *utils.PathDetails, passphrase string, bintrayDetails *config.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = pathDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "gpg/" + pathDetails.Subject + "/" +
		pathDetails.Repo + "/" + pathDetails.Path

	var data string
	if passphrase != "" {
		data = "{ \"passphrase\": \"" + passphrase + "\" }"
	}

	fmt.Println("GPG signing file: " + pathDetails.Path)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body := ioutils.SendPost(url, []byte(data), httpClientsDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

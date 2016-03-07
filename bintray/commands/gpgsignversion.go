package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
)

func GpgSignVersion(versionDetails *utils.VersionDetails, passphrase string, bintrayDetails *cliutils.BintrayDetails) {
	if bintrayDetails.User == "" {
		bintrayDetails.User = versionDetails.Subject
	}
	url := bintrayDetails.ApiUrl + "gpg/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package +
		"/versions/" + versionDetails.Version

	var data string
	if passphrase != "" {
		data = "{ \"passphrase\": \"" + passphrase + "\" }"
	}

	fmt.Println("GPG signing version: " + versionDetails.Version)
	resp, body := cliutils.SendPost(url, nil, []byte(data), bintrayDetails.User, bintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

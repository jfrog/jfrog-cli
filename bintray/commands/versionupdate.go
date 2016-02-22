package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
)

func UpdateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) {
	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = versionDetails.Subject
	}
	data := utils.CreateVersionJson(versionDetails.Version, flags)
	url := flags.BintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version

	fmt.Println("Updating version: " + versionDetails.Version)
	resp, body := cliutils.SendPatch(url, []byte(data), flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

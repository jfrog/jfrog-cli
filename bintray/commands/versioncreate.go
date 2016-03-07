package commands

import (
	"fmt"
	"github.com/JFrogDev/jfrog-cli-go/bintray/utils"
	"github.com/JFrogDev/jfrog-cli-go/cliutils"
	"net/http"
)

func CreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags) {
	fmt.Println("Creating version: " + versionDetails.Version)
	resp, body := doCreateVersion(versionDetails, flags, flags.BintrayDetails)
	if resp.StatusCode != 201 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func DoCreateVersion(versionDetails *utils.VersionDetails,
    bintrayDetails *cliutils.BintrayDetails) (*http.Response, []byte) {
    return doCreateVersion(versionDetails, nil, bintrayDetails)
}

func doCreateVersion(versionDetails *utils.VersionDetails, flags *utils.VersionFlags,
    bintrayDetails *cliutils.BintrayDetails) (*http.Response, []byte) {
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

	return cliutils.SendPost(url, nil, []byte(data), bintrayDetails.User, bintrayDetails.Key)
}

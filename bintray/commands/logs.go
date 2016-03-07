package commands

import (
    "fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
)

func LogsList(packageDetails *utils.VersionDetails, details *cliutils.BintrayDetails) {
	if details.User == "" {
		details.User = packageDetails.Subject
	}
	path := details.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package + "/logs/"
	resp, body, _, _ := cliutils.SendGet(path, nil, true, details.User, details.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
	}

	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
}

func DownloadLog(packageDetails *utils.VersionDetails, logName string,
    details *cliutils.BintrayDetails) {
	if details.User == "" {
		details.User = packageDetails.Subject
	}
	path := details.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package + "/logs/" + logName
    resp := cliutils.DownloadFile(path, "", logName, true, details.User, details.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status)
	}
	fmt.Println("Bintray response: " + resp.Status)
}
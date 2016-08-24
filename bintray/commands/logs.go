package commands

import (
    "errors"
    "fmt"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func LogsList(packageDetails *utils.VersionDetails, details *config.BintrayDetails) (err error) {
	if details.User == "" {
		details.User = packageDetails.Subject
	}
	path := details.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package + "/logs/"
	httpClientsDetails := utils.GetBintrayHttpClientDetails(details)
	resp, body, _, _ := ioutils.SendGet(path, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadBintrayMessage(body)))
        if err != nil {
            return
        }
	}

	fmt.Println("Bintray response: " + resp.Status)
	fmt.Println(cliutils.IndentJson(body))
	return
}

func DownloadLog(packageDetails *utils.VersionDetails, logName string,
    details *config.BintrayDetails) (err error) {
	if details.User == "" {
		details.User = packageDetails.Subject
	}
	path := details.ApiUrl + "packages/" + packageDetails.Subject + "/" +
		packageDetails.Repo + "/" + packageDetails.Package + "/logs/" + logName
	httpClientsDetails := utils.GetBintrayHttpClientDetails(details)
	resp := ioutils.DownloadFile(path, "", logName, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status))
        if err != nil {
            return
        }
	}
	fmt.Println("Bintray response: " + resp.Status)
	return nil
}
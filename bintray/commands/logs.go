package commands

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"fmt"
)

func LogsList(packageDetails *utils.VersionDetails, details *config.BintrayDetails) error {
	if details.User == "" {
		details.User = packageDetails.Subject
	}
	path := details.ApiUrl + "packages/" + packageDetails.Subject + "/" +
			packageDetails.Repo + "/" + packageDetails.Package + "/logs/"
	httpClientsDetails := utils.GetBintrayHttpClientDetails(details)
	log.Info("Getting logs...")
	resp, body, _, _ := httputils.SendGet(path, true, httpClientsDetails)

	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Info("Log details:")
	fmt.Println(cliutils.IndentJson(body))
	return nil
}

func DownloadLog(packageDetails *utils.VersionDetails, logName string,
details *config.BintrayDetails) error {
	if details.User == "" {
		details.User = packageDetails.Subject
	}
	path := details.ApiUrl + "packages/" + packageDetails.Subject + "/" +
			packageDetails.Repo + "/" + packageDetails.Package + "/logs/" + logName
	httpClientsDetails := utils.GetBintrayHttpClientDetails(details)
	log.Info("Downloading logs...")
	resp, err := httputils.DownloadFile(path, "", logName, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Bintray response: " + resp.Status))
	}
	log.Debug("Bintray response:", resp.Status)
	log.Info("Downloaded log.")
	return nil
}
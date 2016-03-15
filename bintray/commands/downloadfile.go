package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
)

func DownloadFile(pathDetails *utils.PathDetails, path string, flags *utils.DownloadFlags) {
	cliutils.CreateTempDirPath()
	defer cliutils.RemoveTempDir()

	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = pathDetails.Subject
	}
	utils.DownloadBintrayFile(flags.BintrayDetails, pathDetails, flags, "")
}

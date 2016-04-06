package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func DownloadFile(pathDetails *utils.PathDetails, path string, flags *utils.DownloadFlags) {
	ioutils.CreateTempDirPath()
	defer ioutils.RemoveTempDir()

	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = pathDetails.Subject
	}
	utils.DownloadBintrayFile(flags.BintrayDetails, pathDetails, flags, "")
}

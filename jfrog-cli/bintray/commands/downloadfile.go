package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

func DownloadFile(pathDetails *utils.PathDetails, targetPath string, flags *utils.DownloadFlags) (totalDownloded, totalFailed int, err error) {
	fileutils.CreateTempDirPath()
	defer fileutils.RemoveTempDir()

	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = pathDetails.Subject
	}
	err = utils.DownloadBintrayFile(flags.BintrayDetails, pathDetails, targetPath, flags, "")
	if err != nil {
		return 0, 1, err
	}
	log.Info("Downloaded 1 artifact.")
	return 1, 0, nil
}

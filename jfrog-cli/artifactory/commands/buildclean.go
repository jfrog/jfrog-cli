package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
)

func BuildClean(buildName, buildNumber string) (err error) {
	cliutils.CliLogger.Info("Cleanning build info...")
	err = utils.RemoveBuildDir(buildName, buildNumber);
	if err != nil {
		return
	}
	cliutils.CliLogger.Info("Cleaned build info", buildName, "#" + buildNumber + ".")
	return err
}

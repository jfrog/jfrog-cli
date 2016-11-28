package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func BuildClean(buildName, buildNumber string) (err error) {
	log.Info("Cleanning build info...")
	err = utils.RemoveBuildDir(buildName, buildNumber);
	if err != nil {
		return
	}
	log.Info("Cleaned build info", buildName, "#" + buildNumber + ".")
	return err
}

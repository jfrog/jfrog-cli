package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

func BuildClean(buildName, buildNumber string) error {
	log.Info("Cleaning build info...")
	err := utils.RemoveBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	log.Info("Cleaned build info", buildName+"/"+buildNumber+".")
	return nil
}

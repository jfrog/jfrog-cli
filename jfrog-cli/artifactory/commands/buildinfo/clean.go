package buildinfo

import (
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
)

func Clean(buildName, buildNumber string) error {
	log.Info("Cleaning build info...")
	err := utils.RemoveBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	log.Info("Cleaned build info", buildName+"/"+buildNumber+".")
	return nil
}

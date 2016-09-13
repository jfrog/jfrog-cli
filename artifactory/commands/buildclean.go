package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

func BuildClean(buildName, buildNumber string) error {
	err := utils.RemoveBuildDir(buildName, buildNumber);
	return err
}

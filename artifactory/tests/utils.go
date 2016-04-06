package tests

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
)

func GetFlags() *utils.Flags {
	flags := new(utils.Flags)
	flags.ArtDetails = new(config.ArtifactoryDetails)
	flags.DryRun = true
	flags.EncPassword = true
	flags.Threads = 3

	return flags
}
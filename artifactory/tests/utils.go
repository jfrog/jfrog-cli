package tests

import (
	"github.com/JFrogDev/jfrog-cli-go/artifactory/utils"
	"github.com/JFrogDev/jfrog-cli-go/cliutils"
)

func GetFlags() *utils.Flags {
	flags := new(utils.Flags)
	flags.ArtDetails = new(cliutils.ArtifactoryDetails)
	flags.DryRun = true
	flags.EncPassword = true
	flags.Threads = 3

	return flags
}
package npm

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
)

type NpmCommand struct {
	repo               string
	buildConfiguration *utils.BuildConfiguration
	npmArgs            []string
	rtDetails          *config.ArtifactoryDetails
}

func (nc *NpmCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *NpmCommand {
	nc.rtDetails = rtDetails
	return nc
}

func (nc *NpmCommand) SetNpmArgs(npmArgs []string) *NpmCommand {
	nc.npmArgs = npmArgs
	return nc
}

func (nc *NpmCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *NpmCommand {
	nc.buildConfiguration = buildConfiguration
	return nc
}

func (nc *NpmCommand) SetRepo(repo string) *NpmCommand {
	nc.repo = repo
	return nc
}

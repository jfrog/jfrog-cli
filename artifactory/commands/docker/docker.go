package docker

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
)

type DockerCommand struct {
	imageTag           string
	repo               string
	buildConfiguration *utils.BuildConfiguration
	rtDetails          *config.ArtifactoryDetails
	skipLogin          bool
}

func (dc *DockerCommand) ImageTag() string {
	return dc.imageTag
}

func (dc *DockerCommand) SetImageTag(imageTag string) *DockerCommand {
	dc.imageTag = imageTag
	return dc
}

func (dc *DockerCommand) Repo() string {
	return dc.repo
}

func (dc *DockerCommand) SetRepo(repo string) *DockerCommand {
	dc.repo = repo
	return dc
}

func (dc *DockerCommand) BuildConfiguration() *utils.BuildConfiguration {
	return dc.buildConfiguration
}

func (dc *DockerCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *DockerCommand {
	dc.buildConfiguration = buildConfiguration
	return dc
}

func (dc *DockerCommand) SetSkipLogin(skipLogin bool) *DockerCommand {
	dc.skipLogin = skipLogin
	return dc
}

func (dc *DockerCommand) RtDetails() *config.ArtifactoryDetails {
	return dc.rtDetails
}

func (dc *DockerCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DockerCommand {
	dc.rtDetails = rtDetails
	return dc
}

package docker

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type DockerPromoteCommand struct {
	rtDetails *config.ArtifactoryDetails
	params    services.DockerPromoteParams
}

func NewDockerPromoteCommand() *DockerPromoteCommand {
	return &DockerPromoteCommand{}
}

func (dp *DockerPromoteCommand) Run() error {
	// Create Service Manager
	servicesManager, err := utils.CreateServiceManager(dp.rtDetails, false)
	if err != nil {
		return err
	}

	// Promote docker
	return servicesManager.PromoteDocker(dp.params)
}

func (dp *DockerPromoteCommand) CommandName() string {
	return "rt_docker_promote"
}

func (dp *DockerPromoteCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return dp.rtDetails, nil
}

func (dp *DockerPromoteCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DockerPromoteCommand {
	dp.rtDetails = rtDetails
	return dp
}

func (dp *DockerPromoteCommand) SetParams(params services.DockerPromoteParams) *DockerPromoteCommand {
	dp.params = params
	return dp
}

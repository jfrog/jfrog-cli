package docker

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"strings"
)

type DockerPullCommand struct {
	DockerCommand
}

func NewDockerPullCommand() *DockerPullCommand {
	return &DockerPullCommand{}
}

// Pull docker image and create build info if needed
func (dpc *DockerPullCommand) Run() error {
	// Perform login
	rtDetails, err := dpc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return err
	}
	if !dpc.skipLogin {
		loginConfig := &docker.DockerLoginConfig{ArtifactoryDetails: rtDetails}
		err = docker.DockerLogin(dpc.imageTag, loginConfig)
		if err != nil {
			return err
		}
	}

	// Perform pull
	if strings.LastIndex(dpc.imageTag, ":") == -1 {
		dpc.imageTag = dpc.imageTag + ":latest"
	}
	image := docker.New(dpc.imageTag)
	err = image.Pull()
	if err != nil {
		return err
	}

	buildName := dpc.BuildConfiguration().BuildName
	buildNumber := dpc.BuildConfiguration().BuildNumber
	// Return if no build name and number was provided
	if buildName == "" || buildNumber == "" {
		return nil
	}

	if err := utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return err
	}

	serviceManager, err := docker.CreateServiceManager(rtDetails, 0)
	if err != nil {
		return err
	}

	builder := docker.BuildInfoBuilder(image, dpc.Repo(), buildName, buildNumber, serviceManager, docker.Pull)
	buildInfo, err := builder.Build(dpc.BuildConfiguration().Module)
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(buildName, buildNumber, buildInfo)
}

func (dpc *DockerPullCommand) CommandName() string {
	return "rt_docker_pull"
}

func (dpc *DockerPullCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return dpc.rtDetails, nil
}

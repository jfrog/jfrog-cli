package docker

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"strings"
)

type DockerPullCommand struct {
	DockerCommand
}

// Pull docker image and create build info if needed
func (dpc *DockerPullCommand) Run() error {
	// Perform login
	imageTag := dpc.ImageTag()
	loginConfig := &docker.DockerLoginConfig{ArtifactoryDetails: dpc.RtDetails()}
	err := docker.DockerLogin(imageTag, loginConfig)
	if err != nil {
		return err
	}

	// Perform pull
	if strings.LastIndex(imageTag, ":") == -1 {
		imageTag = imageTag + ":latest"
	}
	image := docker.New(imageTag)
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

	serviceManager, err := docker.CreateServiceManager(dpc.RtDetails(), 0)
	if err != nil {
		return err
	}

	builder := docker.BuildInfoBuilder(image, dpc.Repo(), buildName, buildNumber, serviceManager, docker.Pull)
	buildInfo, err := builder.Build()
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(buildName, buildNumber, buildInfo)
}

func (dpc *DockerPullCommand) CommandName() string {
	return "rt_docker_pull"
}

func (dpc *DockerPullCommand) RtDetails() *config.ArtifactoryDetails {
	return dpc.rtDetails
}

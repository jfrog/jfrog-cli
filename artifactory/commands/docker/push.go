package docker

import (
	"strings"

	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type DockerPushCommand struct {
	DockerCommand
	threads int
}

func NewDockerPushCommand() *DockerPushCommand {
	return &DockerPushCommand{}
}

func (dpc *DockerPushCommand) Threads() int {
	return dpc.threads
}

func (dpc *DockerPushCommand) SetThreads(threads int) *DockerPushCommand {
	dpc.threads = threads
	return dpc
}

// Push docker image and create build info if needed
func (dpc *DockerPushCommand) Run() error {
	// Perform login
	rtDetails, err := dpc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return err
	}

	err = docker.IsVersionSupported()
	if err != nil {
		return err
	}

	if !dpc.skipLogin {
		loginConfig := &docker.DockerLoginConfig{ArtifactoryDetails: rtDetails}
		err = docker.DockerLogin(dpc.imageTag, loginConfig)
		if err != nil {
			return err
		}
	}

	// Perform push
	if strings.LastIndex(dpc.imageTag, ":") == -1 {
		dpc.imageTag = dpc.imageTag + ":latest"
	}
	image := docker.New(dpc.imageTag)
	err = image.Push()
	if err != nil {
		return err
	}

	// Return if no build name and number was provided
	if dpc.buildConfiguration.BuildName == "" || dpc.buildConfiguration.BuildNumber == "" {
		return nil
	}

	if err := utils.SaveBuildGeneralDetails(dpc.buildConfiguration.BuildName, dpc.buildConfiguration.BuildNumber); err != nil {
		return err
	}

	serviceManager, err := docker.CreateServiceManager(rtDetails, dpc.threads)
	if err != nil {
		return err
	}

	builder := docker.BuildInfoBuilder(image, dpc.Repo(), dpc.BuildConfiguration().BuildName, dpc.BuildConfiguration().BuildNumber, serviceManager, docker.Push)
	buildInfo, err := builder.Build(dpc.BuildConfiguration().Module)
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(dpc.BuildConfiguration().BuildName, dpc.BuildConfiguration().BuildNumber, buildInfo)
}

func (dpc *DockerPushCommand) CommandName() string {
	return "rt_docker_push"
}

func (dpc *DockerPushCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return dpc.rtDetails, nil
}

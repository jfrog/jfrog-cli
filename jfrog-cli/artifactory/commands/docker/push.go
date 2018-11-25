package docker

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"strings"
)

// Push docker image and create build info if needed
func PushDockerImage(imageTag, targetRepo, buildName, buildNumber string, artDetails *config.ArtifactoryDetails, threads int) error {
	// Perform login
	loginConfig := &docker.DockerLoginConfig{ArtifactoryDetails: artDetails}
	err := docker.DockerLogin(imageTag, loginConfig)
	if err != nil {
		return err
	}

	// Perform push
	if strings.LastIndex(imageTag, ":") == -1 {
		imageTag = imageTag + ":latest"
	}
	image := docker.New(imageTag)
	err = image.Push()
	if err != nil {
		return err
	}

	// Return if no build name and number was provided
	if buildName == "" || buildNumber == "" {
		return nil
	}

	if err := utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return err
	}

	serviceManager, err := docker.CreateServiceManager(artDetails, threads)
	if err != nil {
		return err
	}

	builder := docker.BuildInfoBuilder(image, targetRepo, buildName, buildNumber, serviceManager, docker.Push)
	buildInfo, err := builder.Build()
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(buildName, buildNumber, buildInfo)
}

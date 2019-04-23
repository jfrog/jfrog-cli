package docker

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"strings"
)

// Pull docker image and create build info if needed
func PullDockerImage(imageTag, sourceRepo, buildName, buildNumber string, artDetails *config.ArtifactoryDetails) error {
	// Perform login
	loginConfig := &docker.DockerLoginConfig{ArtifactoryDetails: artDetails}
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

	// Return if no build name and number was provided
	if buildName == "" || buildNumber == "" {
		return nil
	}

	if err := utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return err
	}

	serviceManager, err := docker.CreateServiceManager(artDetails, 0)
	if err != nil {
		return err
	}

	builder := docker.BuildInfoBuilder(image, sourceRepo, buildName, buildNumber, serviceManager, docker.Pull)
	buildInfo, err := builder.Build()
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(buildName, buildNumber, buildInfo)
}

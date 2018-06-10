package docker

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"strings"
)

type DockerPushConfig struct {
	ArtifactoryDetails *config.ArtifactoryDetails
	Threads            int
}

// Push docker image and create build info if needed
func PushDockerImage(imageTag, targetRepo, buildName, buildNumber string, config *DockerPushConfig) error {
	if strings.LastIndex(imageTag, ":") == -1 {
		imageTag = imageTag + ":latest"
	}
	image := docker.New(imageTag)
	err := image.Push()
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

	serviceManager, err := createServiceManager(config)
	if err != nil {
		return err
	}

	builder := docker.BuildInfoBuilder(image, targetRepo, buildName, buildNumber, serviceManager)
	buildInfo, err := builder.Build()
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(buildName, buildNumber, buildInfo)
}

func createServiceManager(config *DockerPushConfig) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := utils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := config.ArtifactoryDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := artifactory.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetCertificatesPath(certPath).
		SetLogger(log.Logger).
		SetThreads(config.Threads).
		Build()

	return artifactory.New(serviceConfig)
}

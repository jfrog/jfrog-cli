package docker

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"strings"
)

// Docker login error message
const dockerLoginFailureMessage string = "Docker login failed for: %s.\nDocker image must be in the form: docker-registry-domain/path-in-repository/image-name:version."

type DockerPushConfig struct {
	ArtifactoryDetails *config.ArtifactoryDetails
	Threads            int
}

// Push docker image and create build info if needed
func PushDockerImage(imageTag, targetRepo, buildName, buildNumber string, config *DockerPushConfig) error {
	// Perform login
	loginConfig := &dockerLoginConfig{ArtifactoryDetails: config.ArtifactoryDetails}
	err := dockerLogin(imageTag, loginConfig)
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

type dockerLoginConfig struct {
	ArtifactoryDetails *config.ArtifactoryDetails
}

// First will try to login assuming a proxy-less tag (e.g. "registry-address/docker-repo/image:ver").
// If fails, we will try assuming a reverse proxy tag (e.g. "registry-address-docker-repo/image:ver").
func dockerLogin(imageTag string, config *dockerLoginConfig) error {
	imageRegistry, err := docker.ResolveRegistryFromTag(imageTag)
	if err != nil {
		return err
	}

	cmd := &docker.LoginCmd{DockerRegistry: imageRegistry, Username: config.ArtifactoryDetails.User, Password: config.ArtifactoryDetails.Password}
	err = utils.RunCmd(cmd)

	if exitCode := cliutils.GetExitCode(err, 0, 0, false); exitCode == cliutils.ExitCodeNoError {
		// Login succeeded
		return nil
	}

	indexOfSlash := strings.Index(imageRegistry, "/")
	if indexOfSlash < 0 {
		return errorutils.CheckError(errors.New(fmt.Sprintf(dockerLoginFailureMessage, imageRegistry)))
	}

	cmd = &docker.LoginCmd{DockerRegistry: imageRegistry[:indexOfSlash], Username: config.ArtifactoryDetails.User, Password: config.ArtifactoryDetails.Password}
	err = utils.RunCmd(cmd)
	if err != nil {
		// Login failed for both attempts
		return errorutils.CheckError(errors.New(fmt.Sprintf(dockerLoginFailureMessage,
			fmt.Sprintf("%s, %s", imageRegistry, imageRegistry[:indexOfSlash]))))
	}

	// Login succeeded
	return nil
}

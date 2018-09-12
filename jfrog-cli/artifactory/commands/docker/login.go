package docker

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"strings"
)

type DockerLoginConfig struct {
	ArtifactoryDetails *config.ArtifactoryDetails
}

// First will try to login assuming a proxy-less tag (e.g. "registry-address/docker-repo/image:ver").
// If fails, we will try assuming a reverse proxy tag (e.g. "registry-address-docker-repo/image:ver").
func DockerLogin(imageTag string, config *DockerLoginConfig) error {
	imageRegistry := docker.ResolveRegistryFromTag(imageTag)
	cmd := &docker.LoginCmd{DockerRegistry: imageRegistry, Username: config.ArtifactoryDetails.User, Password: config.ArtifactoryDetails.Password}
	err := utils.RunCmd(cmd)

	if exitCode := cliutils.GetExitCode(err, 0, 0, false); exitCode == cliutils.ExitCodeNoError {
		// Login succeeded
		return nil
	}

	indexOfSlash := strings.Index(imageTag, "/")
	if indexOfSlash != 0 {
		cmd = &docker.LoginCmd{DockerRegistry: imageRegistry[:indexOfSlash], Username: config.ArtifactoryDetails.User, Password: config.ArtifactoryDetails.Password}
		err = utils.RunCmd(cmd)
		if err != nil {
			// Login failed for both attempts
			return err
		}
	}

	// Login succeeded
	return nil
}

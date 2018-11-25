package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

func BuildDiscard(configuration *BuildDiscardConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, false)
	if err != nil {
		return err
	}
	return servicesManager.DiscardBuilds(configuration.DiscardBuildsParams)
}

type BuildDiscardConfiguration struct {
	ArtDetails *config.ArtifactoryDetails
	services.DiscardBuildsParams
}

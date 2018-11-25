package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

func Distribute(configuration *BuildDistributionConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, configuration.DryRun)
	if err != nil {
		return err
	}
	return servicesManager.DistributeBuild(configuration.BuildDistributionParams)
}

type BuildDistributionConfiguration struct {
	services.BuildDistributionParams
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

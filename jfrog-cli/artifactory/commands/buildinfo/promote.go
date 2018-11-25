package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

func Promote(configuration *BuildPromotionConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, configuration.DryRun)
	if err != nil {
		return err
	}
	return servicesManager.PromoteBuild(configuration.PromotionParams)
}

type BuildPromotionConfiguration struct {
	services.PromotionParams
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
)

func BuildPromote(flags *BuildPromotionConfiguration) error {
	servicesManager, err := utils.CreateDefaultServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	return servicesManager.PromoteBuild(flags.PromotionParamsImpl)
}

type BuildPromotionConfiguration struct {
	*services.PromotionParamsImpl
	ArtDetails *config.ArtifactoryDetails
	DryRun bool
}

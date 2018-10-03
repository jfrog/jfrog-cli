package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

func Promote(flags *BuildPromotionConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	return servicesManager.PromoteBuild(flags.PromotionParamsImpl)
}

type BuildPromotionConfiguration struct {
	*services.PromotionParamsImpl
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services"
)

func Distribute(flags *BuildDistributionConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	return servicesManager.DistributeBuild(flags.BuildDistributionParamsImpl)
}

type BuildDistributionConfiguration struct {
	*services.BuildDistributionParamsImpl
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

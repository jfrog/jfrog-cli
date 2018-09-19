package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services"
)

func BuildDiscard(flags *BuildDiscardConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, false)
	if err != nil {
		return err
	}
	return servicesManager.DiscardBuilds(flags.DiscardBuildsParamsImpl)
}

type BuildDiscardConfiguration struct {
	ArtDetails *config.ArtifactoryDetails
	*services.DiscardBuildsParamsImpl
}

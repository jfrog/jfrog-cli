package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
)

// Moves the artifacts using the specified move pattern.
func Move(moveSpec *utils.SpecFiles, artDetails *config.ArtifactoryDetails) error {
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return err
	}
	for i := 0; i < len(moveSpec.Files); i++ {
		params, err := moveSpec.Get(i).ToArtifatoryMoveCopyParams()
		if err != nil {
			return err
		}
		flat, err := moveSpec.Get(i).IsFlat(false)
		if err != nil {
			return err
		}
		err = servicesManager.Move(&services.MoveCopyParamsImpl{ArtifactoryCommonParams: params, Flat: flat})
		if err != nil {
			return err
		}
	}
	return nil
}

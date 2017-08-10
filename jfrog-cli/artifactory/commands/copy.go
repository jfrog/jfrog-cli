package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
)

// Copies the artifacts using the specified move pattern.
func Copy(copySpec *utils.SpecFiles, artDetails *config.ArtifactoryDetails) error {
	servicesManager, err := utils.CreateDefaultServiceManager(artDetails, false)
	if err != nil {
		return err
	}
	for i := 0; i < len(copySpec.Files); i++ {
		params, err := copySpec.Get(i).ToArtifatoryMoveCopyParams()
		if err != nil {
			return err
		}
		flat, err := copySpec.Get(i).IsFlat(false)
		if err != nil {
			return err
		}
		err = servicesManager.Copy(&services.MoveCopyParamsImpl{ArtifactoryCommonParams: params, Flat:flat})
		if err != nil {
			return err
		}
	}
	return nil
}
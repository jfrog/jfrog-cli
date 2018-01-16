package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/spec"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

// Copies the artifacts using the specified move pattern.
func Copy(copySpec *spec.SpecFiles, artDetails *config.ArtifactoryDetails) (successCount, failedCount int, err error) {
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return successCount, failedCount, err
	}
	for i := 0; i < len(copySpec.Files); i++ {
		params, err := copySpec.Get(i).ToArtifatoryMoveCopyParams()
		if err != nil {
			log.Error(err)
			continue
		}
		flat, err := copySpec.Get(i).IsFlat(false)
		if err != nil {
			log.Error(err)
			continue
		}
		partialSuccess, partialFailed, err := servicesManager.Copy(&services.MoveCopyParamsImpl{ArtifactoryCommonParams: params, Flat: flat})
		successCount += partialSuccess
		failedCount += partialFailed
		if err != nil {
			log.Error(err)
			continue
		}
	}
	return
}

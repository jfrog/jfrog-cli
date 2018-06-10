package generic

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
)

// Copies the artifacts using the specified move pattern.
func Copy(copySpec *spec.SpecFiles, flags *CopyConfiguration) (successCount, failCount int, err error) {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return
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
		failCount += partialFailed
		if err != nil {
			log.Error(err)
			continue
		}
	}
	return
}

type CopyConfiguration struct {
	DryRun                bool
	ArtDetails            *config.ArtifactoryDetails
}
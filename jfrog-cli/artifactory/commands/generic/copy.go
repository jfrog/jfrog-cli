package generic

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// Copies the artifacts using the specified move pattern.
func Copy(copySpec *spec.SpecFiles, configuration *CopyConfiguration) (successCount, failCount int, err error) {

	// Create Service Manager:
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, configuration.DryRun)
	if err != nil {
		return
	}

	// Copy Loop:
	for i := 0; i < len(copySpec.Files); i++ {

		copyParams, err := getCopyParams(copySpec.Get(i))
		if err != nil {
			log.Error(err)
			continue
		}

		partialSuccess, partialFailed, err := servicesManager.Copy(copyParams)
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
	DryRun     bool
	ArtDetails *config.ArtifactoryDetails
}

func getCopyParams(f *spec.File) (copyParams services.MoveCopyParams, err error) {
	copyParams = services.NewMoveCopyParams()
	copyParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	copyParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	copyParams.Flat, err = f.IsFlat(false)
	if err != nil {
		return
	}
	return
}

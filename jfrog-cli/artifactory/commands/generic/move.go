package generic

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// Moves the artifacts using the specified move pattern.
func Move(moveSpec *spec.SpecFiles, configuration *MoveConfiguration) (successCount, failCount int, err error) {

	// Create Service Manager:
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, configuration.DryRun)
	if err != nil {
		return
	}

	// Move Loop:
	for i := 0; i < len(moveSpec.Files); i++ {

		moveParams, err := GetMoveParams(moveSpec.Get(i))
		if err != nil {
			log.Error(err)
			continue
		}

		partialSuccess, partialFailed, err := servicesManager.Move(moveParams)
		successCount += partialSuccess
		failCount += partialFailed
		if err != nil {
			log.Error(err)
			continue
		}
	}
	return
}

type MoveConfiguration struct {
	DryRun     bool
	ArtDetails *config.ArtifactoryDetails
}

func GetMoveParams(f *spec.File) (moveParams services.MoveCopyParams, err error) {
	moveParams = services.NewMoveCopyParams()

	moveParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()

	moveParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	moveParams.Flat, err = f.IsFlat(false)
	if err != nil {
		return
	}

	return
}

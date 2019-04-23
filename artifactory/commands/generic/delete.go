package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)

func GetPathsToDelete(deleteSpec *spec.SpecFiles, configuration *DeleteConfiguration) ([]clientutils.ResultItem, error) {
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, configuration.DryRun)
	if err != nil {
		return nil, err
	}
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(deleteSpec.Files); i++ {
		deleteParams, err := getDeleteParams(deleteSpec.Get(i))
		if err != nil {
			return nil, err
		}

		currentResultItems, err := servicesManager.GetPathsToDelete(deleteParams)
		if err != nil {
			return nil, err
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	return resultItems, nil
}

func DeleteFiles(deleteItems []clientutils.ResultItem, configuration *DeleteConfiguration) (successCount, failedCount int, err error) {
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, configuration.DryRun)
	if err != nil {
		return 0, 0, err
	}
	deletedCount, err := servicesManager.DeleteFiles(deleteItems)
	return deletedCount, len(deleteItems) - deletedCount, err
}

type DeleteConfiguration struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

func getDeleteParams(f *spec.File) (deleteParams services.DeleteParams, err error) {
	deleteParams = services.NewDeleteParams()
	deleteParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	deleteParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}
	return
}

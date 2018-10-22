package generic

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)

func GetPathsToDelete(deleteSpec *spec.SpecFiles, flags *DeleteConfiguration) ([]clientutils.ResultItem, error) {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return nil, err
	}
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(deleteSpec.Files); i++ {
		deleteParams, err := GetDeleteParams(deleteSpec.Get(i))
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

func DeleteFiles(resultItems []clientutils.ResultItem, flags *DeleteConfiguration) (successCount, failedCount int, err error) {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return 0, 0, err
	}
	deleteItems := utils.ConvertResultItemArrayToDeleteItemArray(resultItems)
	deletedCount, err := servicesManager.DeleteFiles(deleteItems)
	return deletedCount, len(deleteItems) - deletedCount, err
}

type DeleteConfiguration struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

func GetDeleteParams(f *spec.File) (deleteParams services.DeleteParams, err error) {
	deleteParams = services.NewDeleteParams()

	deleteParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()

	deleteParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	return
}

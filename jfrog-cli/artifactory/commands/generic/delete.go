package generic

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)

func Delete(deleteSpec *spec.SpecFiles, flags *DeleteConfiguration) (successCount, failCount int, err error) {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return 0, 0, err
	}
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(deleteSpec.Files); i++ {
		params, err := deleteSpec.Get(i).ToArtifatoryDeleteParams()
		if err != nil {
			return 0, 0, err
		}
		currentResultItems, err := servicesManager.GetPathsToDelete(&services.DeleteParamsImpl{ArtifactoryCommonParams: params})
		if err != nil {
			return 0, 0, err
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	deleteItems := utils.ConvertResultItemArrayToDeleteItemArray(resultItems)
	deletedCount, err := servicesManager.DeleteFiles(deleteItems)
	return deletedCount, len(deleteItems) - deletedCount, err
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

func GetPathsToDelete(deleteSpec *spec.SpecFiles, flags *DeleteConfiguration) ([]clientutils.ResultItem, error) {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return nil, err
	}
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(deleteSpec.Files); i++ {
		params, err := deleteSpec.Get(i).ToArtifatoryDeleteParams()
		if err != nil {
			return nil, err
		}
		currentResultItems, err := servicesManager.GetPathsToDelete(&services.DeleteParamsImpl{ArtifactoryCommonParams: params})
		if err != nil {
			return nil, err
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	return resultItems, nil
}

type DeleteConfiguration struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

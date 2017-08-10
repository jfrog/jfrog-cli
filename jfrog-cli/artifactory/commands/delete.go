package commands

import (
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
)

func Delete(deleteSpec *utils.SpecFiles, flags *DeleteConfiguration) (err error) {
	servicesManager, err := utils.CreateDefaultServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(deleteSpec.Files); i++ {
		params, err := deleteSpec.Get(i).ToArtifatoryDeleteParams()
		if err != nil {
			return err
		}
		currentResultItems, err := servicesManager.GetPathsToDelete(&services.DeleteParamsImpl{ArtifactoryCommonParams: params})
		if err != nil {
			return err
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	deleteItems := utils.ConvertResultItemArrayToDeleteItemArray(resultItems)
	if err = servicesManager.DeleteFiles(deleteItems); err != nil {
		return
	}
	cliutils.CliLogger.Info("Deleted", len(resultItems), "items.")
	return
}

func DeleteFiles(resultItems []clientutils.ResultItem, flags *DeleteConfiguration) error {
	servicesManager, err := utils.CreateDefaultServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	deleteItems := utils.ConvertResultItemArrayToDeleteItemArray(resultItems)
	return servicesManager.DeleteFiles(deleteItems)
}

func GetPathsToDelete(deleteSpec *utils.SpecFiles, flags *DeleteConfiguration) ([]clientutils.ResultItem, error) {
	servicesManager, err := utils.CreateDefaultServiceManager(flags.ArtDetails, flags.DryRun)
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

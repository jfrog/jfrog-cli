package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func Delete(deleteSpec *utils.SpecFiles, flags *DeleteFlags) (err error) {
	err = utils.PreCommandSetup(flags)
	if err != nil {
		return
	}

	var resultItems []utils.AqlSearchResultItem
	for i := 0; i < len(deleteSpec.Files); i++ {
		isDirectoryDelete, e := isDirectoryDelete(deleteSpec.Get(i))
		if e != nil {
			err = e
			return
		}
		switch {
		case deleteSpec.Get(i).GetSpecType() == utils.AQL:
			resultItems, err = utils.AqlSearchBySpec(deleteSpec.Get(i).Aql, flags)

		case isDirectoryDelete:
			simplePathItem := utils.AqlSearchResultItem{Path:deleteSpec.Get(i).Pattern}
			resultItems = []utils.AqlSearchResultItem{simplePathItem}

		default:
			isRecursive, e := cliutils.StringToBool(deleteSpec.Get(i).Recursive, true)
			if e != nil {
				err = e
				return
			}
			resultItems, err = utils.AqlSearchDefaultReturnFields(deleteSpec.Get(i).Pattern,
				isRecursive, deleteSpec.Get(i).Props, flags)
		}

		if err != nil {
			return
		}

		if err = deleteFiles(resultItems, flags); err != nil {
			return
		}
	}
	return
}

func isDirectoryDelete(deleteFile *utils.Files) (bool, error) {
	isRecursive, err := cliutils.StringToBool(deleteFile.Recursive, true)
	if err != nil {
		return false, err
	}
	return utils.IsSimpleDirectoryPath(deleteFile.Pattern) && isRecursive == true && deleteFile.Props == "", nil
}

func deleteFiles(resultItems []utils.AqlSearchResultItem, flags *DeleteFlags) error {
	for _, v := range resultItems {
		fileUrl, err := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, v.GetFullUrl(), make(map[string]string))
		if err != nil {
			return err
		}
		if flags.DryRun {
			logger.Logger.Info("[Dry run] Deleting: " + fileUrl)
			continue
		}

		logger.Logger.Info("Deleting: " + fileUrl)
		httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
		resp, _, err := ioutils.SendDelete(fileUrl, nil, httpClientsDetails)
		if err != nil {
			return err
		}
		logger.Logger.Info("Artifactory response:", resp.Status)
	}
	return nil
}

type DeleteFlags struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

func (flags *DeleteFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *DeleteFlags) IsDryRun() bool {
	return flags.DryRun
}
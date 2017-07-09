package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"errors"
)

func Delete(deleteSpec *utils.SpecFiles, flags utils.CommonFlags) (err error) {
	err = utils.PreCommandSetup(flags)
	if err != nil {
		return
	}
	resultItems, err := getPathsToDeleteInternal(deleteSpec, flags)
	if err != nil {
		return err
	}
	if err = deleteFiles(resultItems, flags); err != nil {
		return
	}
	log.Info("Deleted", len(resultItems), "items.")
	return
}

func GetPathsToDelete(deleteSpec *utils.SpecFiles, flags utils.CommonFlags) ([]utils.AqlSearchResultItem, error) {
	if err := utils.PreCommandSetup(flags); err != nil {
		return nil, err
	}
	return getPathsToDeleteInternal(deleteSpec, flags)
}

func getPathsToDeleteInternal(deleteSpec *utils.SpecFiles, flags utils.CommonFlags) (resultItems []utils.AqlSearchResultItem, err error) {
	log.Info("Searching artifacts...")
	for i := 0; i < len(deleteSpec.Files); i++ {
		currentSpec := deleteSpec.Get(i)
		// Search paths using AQL.
		if currentSpec.GetSpecType() == utils.AQL {
			if resultItemsTemp, e := utils.AqlSearchBySpec(currentSpec, flags); e == nil {
				resultItems = append(resultItems, resultItemsTemp...)
				continue
			} else {
				err = e
				return
			}
		}

		currentSpec.IncludeDirs = "true"
		tempResultItems, e := utils.AqlSearchDefaultReturnFields(currentSpec, flags)
		if e != nil {
			err = e
			return
		}
		paths := utils.ReduceDirResult(tempResultItems, utils.FilterTopChainResults)
		resultItems = append(resultItems, paths...)
	}
	utils.LogSearchResults(len(resultItems))
	return
}

func DeleteFiles(resultItems []utils.AqlSearchResultItem, flags utils.CommonFlags) error {
	if err := utils.PreCommandSetup(flags); err != nil {
		return err
	}
	return deleteFiles(resultItems, flags)
}

func deleteFiles(resultItems []utils.AqlSearchResultItem, flags utils.CommonFlags) error {
	for _, v := range resultItems {
		fileUrl, err := utils.BuildArtifactoryUrl(flags.GetArtifactoryDetails().Url, v.GetFullUrl(), make(map[string]string))
		if err != nil {
			return err
		}
		if flags.IsDryRun() {
			log.Info("[Dry run] Deleting:", v.GetFullUrl())
			continue
		}

		log.Info("Deleting:", v.GetFullUrl())
		httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.GetArtifactoryDetails())
		resp, body, err := httputils.SendDelete(fileUrl, nil, httpClientsDetails)
		if err != nil {
			return err
		}
		if resp.StatusCode != 204 {
			return cliutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
		}

		log.Debug("Artifactory response:", resp.Status)
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
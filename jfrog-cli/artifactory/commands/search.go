package commands

import (
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
)

type SearchResult struct {
	Path string `json:"path,omitempty"`
}

func Search(searchSpec *utils.SpecFiles, artDetails *config.ArtifactoryDetails) ([]SearchResult, error) {
	servicesManager, err := utils.CreateDefaultServiceManager(artDetails, false)
	if err != nil {
		return nil, err
	}
	cliutils.CliLogger.Info("Searching artifacts...")
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(searchSpec.Files); i++ {
		params, err := searchSpec.Get(i).ToArtifatorySearchParams()
		if err != nil {
			return nil, err
		}
		currentResultItems, err := servicesManager.Search(&clientutils.SearchParamsImpl{ArtifactoryCommonParams: params})
		if err != nil {
			return nil, err
		}
		resultItems = append(resultItems, currentResultItems...)
	}

	result := aqlResultToSearchResult(resultItems)
	clientutils.LogSearchResults(len(resultItems))
	return result, err
}

func aqlResultToSearchResult(aqlResult []clientutils.ResultItem) (result []SearchResult) {
	result = make([]SearchResult, len(aqlResult))
	for i, v := range aqlResult {
		tempResult := new(SearchResult)
		if v.Path != "." {
			tempResult.Path = v.Repo + "/" + v.Path + "/" + v.Name
		} else {
			tempResult.Path = v.Repo + "/" + v.Name
		}
		result[i] = *tempResult
	}
	return
}

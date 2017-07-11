package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

type SearchResult struct {
	Path string `json:"path,omitempty"`
}

func Search(searchSpec *utils.SpecFiles, flags utils.CommonFlags) (result []SearchResult, err error) {
	err = utils.PreCommandSetup(flags)
	if err != nil {
		return
	}
	log.Info("Searching artifacts...")
	resultItems, err := utils.SearchBySpecFiles(searchSpec, flags)
	if err != nil {
		return
	}
	result = aqlResultToSearchResult(resultItems)
	utils.LogSearchResults(len(resultItems))
	return
}

func aqlResultToSearchResult(aqlResult []utils.AqlSearchResultItem) (result []SearchResult) {
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

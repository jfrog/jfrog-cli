package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
)

type SearchResult struct {
	Path string `json:"path,omitempty"`
}

func Search(searchSpec *utils.SpecFiles, flags *SearchFlags) (result []SearchResult, err error) {
	utils.PreCommandSetup(flags)
	var resultItems []utils.AqlSearchResultItem
	var itemsFound []utils.AqlSearchResultItem
	for i := 0; i < len(searchSpec.Files); i++ {
		switch searchSpec.Get(i).GetSpecType() {
		case utils.WILDCARD, utils.SIMPLE:
			itemsFound, err = utils.AqlSearchDefaultReturnFields(searchSpec.Get(i).Pattern,
				searchSpec.Get(i).Recursive, searchSpec.Get(i).Props, flags)
			if err != nil {
				return
			}
			resultItems = append(resultItems, itemsFound...)
		case utils.AQL:
			itemsFound, err = utils.AqlSearchBySpec(searchSpec.Get(i).Aql, flags)
			if err != nil {
				return
			}
			resultItems = append(resultItems, itemsFound...)
		}
	}
	result = aqlResultToSearchResult(resultItems)
	return
}

func aqlResultToSearchResult(aqlResult []utils.AqlSearchResultItem) (result []SearchResult) {
	result = make([]SearchResult, len(aqlResult))
	for i, v := range aqlResult {
		tempResult := new(SearchResult)
		if v.Path != "." {
			tempResult.Path = v.Repo + "/" + v.Path + "/" + v.Name;
		} else {
			tempResult.Path = v.Repo + "/" + v.Name;
		}
		result[i] = *tempResult
	}
	return
}

type SearchFlags struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}

func (flags *SearchFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *SearchFlags) IsDryRun() bool {
	return flags.DryRun
}
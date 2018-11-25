package generic

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type SearchResult struct {
	Path  string              `json:"path,omitempty"`
	Props map[string][]string `json:"props,omitempty"`
}

func Search(searchSpec *spec.SpecFiles, artDetails *config.ArtifactoryDetails) ([]SearchResult, error) {

	// Service Manager
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return nil, err
	}

	// Search Loop
	log.Info("Searching artifacts...")
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(searchSpec.Files); i++ {

		searchParams, err := GetSearchParams(searchSpec.Get(i))
		if err != nil {
			log.Error(err)
			return nil, err
		}

		currentResultItems, err := servicesManager.SearchFiles(searchParams)
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
		tempResult.Props = make(map[string][]string, len(v.Properties))
		for _, prop := range v.Properties {
			tempResult.Props[prop.Key] = append(tempResult.Props[prop.Key], prop.Value)
		}
		result[i] = *tempResult
	}
	return
}

func GetSearchParams(f *spec.File) (searchParams services.SearchParams, err error) {
	searchParams = services.NewSearchParams()
	searchParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	searchParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	return
}

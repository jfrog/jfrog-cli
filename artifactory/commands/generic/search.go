package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type SearchResult struct {
	Path  string              `json:"path,omitempty"`
	Type  string              `json:"type,omitempty"`
	Props map[string][]string `json:"props,omitempty"`
}

type SearchCommand struct {
	GenericCommand
	searchResult []SearchResult
	includeDirs  bool
}

func NewSearchCommand() *SearchCommand {
	return &SearchCommand{GenericCommand: *NewGenericCommand()}
}

func (sc *SearchCommand) SearchResult() []SearchResult {
	return sc.searchResult
}

func (sc *SearchCommand) CommandName() string {
	return "rt_search"
}

func (sc *SearchCommand) SetIncludeDirs(includeDirs bool) *SearchCommand {
	sc.includeDirs = includeDirs
	return sc
}

func (sc *SearchCommand) Run() error {
	return sc.Search()
}

func (sc *SearchCommand) Search() error {
	// Service Manager
	rtDetails, err := sc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return err
	}
	servicesManager, err := utils.CreateServiceManager(rtDetails, false)
	if err != nil {
		return err
	}

	// Search Loop
	log.Info("Searching artifacts...")
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(sc.Spec().Files); i++ {

		searchParams, err := GetSearchParams(sc.Spec().Get(i))
		if err != nil {
			log.Error(err)
			return err
		}
		searchParams.IncludeDirs = sc.includeDirs

		currentResultItems, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			return err
		}
		resultItems = append(resultItems, currentResultItems...)
	}

	sc.searchResult = sc.aqlResultToSearchResult(resultItems)
	clientutils.LogSearchResults(len(resultItems))
	return err
}

func (sc *SearchCommand) aqlResultToSearchResult(aqlResult []clientutils.ResultItem) (result []SearchResult) {
	result = make([]SearchResult, len(aqlResult))
	for i, v := range aqlResult {
		tempResult := new(SearchResult)
		tempResult.Path = v.Repo + "/"
		if v.Path != "." {
			tempResult.Path += v.Path + "/"
		}
		if v.Name != "." {
			tempResult.Path += v.Name
		}
		if sc.includeDirs {
			tempResult.Type = v.Type
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

package generic

import (
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type SearchResult struct {
	Path     string              `json:"path,omitempty"`
	Type     string              `json:"type,omitempty"`
	Size     int64               `json:"size,omitempty"`
	Created  string              `json:"created,omitempty"`
	Modified string              `json:"modified,omitempty"`
	Props    map[string][]string `json:"props,omitempty"`
}

type SearchCommand struct {
	GenericCommand
	searchResult []SearchResult
}

func NewSearchCommand() *SearchCommand {
	return &SearchCommand{GenericCommand: *NewGenericCommand()}
}

func (sc *SearchCommand) SearchResult() []SearchResult {
	return sc.searchResult
}

func (sc *SearchCommand) SearchResultNoDate() []SearchResult {
	for i, _ := range sc.SearchResult() {
		sc.searchResult[i].Created = ""
		sc.searchResult[i].Modified = ""
		delete(sc.searchResult[i].Props, "vcs.url")
		delete(sc.searchResult[i].Props, "vcs.revision")
	}
	return sc.searchResult
}

func (sc *SearchCommand) CommandName() string {
	return "rt_search"
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

		currentResultItems, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			return err
		}
		resultItems = append(resultItems, currentResultItems...)
	}

	sc.searchResult = aqlResultToSearchResult(resultItems)
	clientutils.LogSearchResults(len(resultItems))
	return err
}

func aqlResultToSearchResult(aqlResult []clientutils.ResultItem) (result []SearchResult) {
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
		tempResult.Type = v.Type
		tempResult.Size = v.Size
		tempResult.Created = v.Created
		tempResult.Modified = v.Modified
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

	searchParams.IncludeDirs, err = f.IsIncludeDirs(false)
	return
}

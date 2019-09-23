package tests

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

//Used in tests TestArtifactorySearchIncludeDir & 'TestArtifactorySearchProps'.
//Inorder to compare hard coded json's struction result(e.g. GetSearchPropsStep)
//witch cannot contain future 'Created' & 'Modified' dates
type SearchResultNoDate struct {
	Path  string              `json:"path,omitempty"`
	Type  string              `json:"type,omitempty"`
	Size  int64               `json:"size,omitempty"`
	Props map[string][]string `json:"props,omitempty"`
}

type SearchCommandNoDate struct {
	generic.GenericCommand
	searchResultNoDate []SearchResultNoDate
}

func NewSearchCommandNoDate() *SearchCommandNoDate {
	return &SearchCommandNoDate{GenericCommand: *generic.NewGenericCommand()}
}

func (sc *SearchCommandNoDate) SearchResultNoDate() []SearchResultNoDate {
	return sc.searchResultNoDate
}

func (sc *SearchCommandNoDate) CommandNameNoDate() string {
	return "rt_search"
}

func (sc *SearchCommandNoDate) Run() error {
	return sc.Search()
}

func (sc *SearchCommandNoDate) Search() error {
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

	sc.searchResultNoDate = aqlResultToSearchResult2(resultItems)
	clientutils.LogSearchResults(len(resultItems))
	return err
}

func aqlResultToSearchResult2(aqlResult []clientutils.ResultItem) (result []SearchResultNoDate) {
	result = make([]SearchResultNoDate, len(aqlResult))
	for i, v := range aqlResult {
		tempResult := new(SearchResultNoDate)
		tempResult.Path = v.Repo + "/"
		if v.Path != "." {
			tempResult.Path += v.Path + "/"
		}
		if v.Name != "." {
			tempResult.Path += v.Name
		}
		tempResult.Type = v.Type
		tempResult.Size = v.Size
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

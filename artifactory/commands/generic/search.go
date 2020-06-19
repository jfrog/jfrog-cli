package generic

import (
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type SearchResult struct {
	Path     string              `json:"path,omitempty"`
	Type     string              `json:"type,omitempty"`
	Size     int64               `json:"size,omitempty"`
	Created  string              `json:"created,omitempty"`
	Modified string              `json:"modified,omitempty"`
	Sha1     string              `json:"sha1,omitempty"`
	Md5      string              `json:"md5,omitempty"`
	Props    map[string][]string `json:"props,omitempty"`
}

type SearchCommand struct {
	GenericCommand
	ContentReadearchResult *content.ContentReader
}

func NewSearchCommand() *SearchCommand {
	return &SearchCommand{GenericCommand: *NewGenericCommand()}
}

func (sc *SearchCommand) SearchResult() *content.ContentReader {
	sc.ContentReadearchResult.Reset()
	return sc.ContentReadearchResult
}

func (sc *SearchCommand) SearchResultNoDate() (*content.ContentReader, error) {
	cw, err := content.NewContentWriter("results", true, false)
	if err != nil {
		return nil, err
	}
	cr := sc.SearchResult()
	for resultItem := new(SearchResult); cr.NextRecord(resultItem) == nil; resultItem = new(SearchResult) {
		if err != nil {
			return nil, err
		}
		resultItem.Created = ""
		resultItem.Modified = ""
		delete(resultItem.Props, "vcs.url")
		delete(resultItem.Props, "vcs.revision")
		cw.Write(resultItem)
	}
	if err := cr.GetError(); err != nil {
		return nil, err
	}
	cw.Close()
	cr.SetFilePath(cw.GetFilePath())
	cr.Reset()
	return cr, nil
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
	var searchResults []*content.ContentReader
	for i := 0; i < len(sc.Spec().Files); i++ {
		searchParams, err := GetSearchParams(sc.Spec().Get(i))
		if err != nil {
			log.Error(err)
			return err
		}

		currentsearchResults, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			return err
		}
		searchResults = append(searchResults, currentsearchResults)
	}

	sc.ContentReadearchResult, err = aqlResultToSearchResult(searchResults)
	length, err := sc.ContentReadearchResult.Length()
	if err != nil {
		return err
	}
	clientutils.LogSearchResults(length)
	return err
}

func aqlResultToSearchResult(crs []*content.ContentReader) (*content.ContentReader, error) {
	cw, err := content.NewContentWriter("results", true, false)
	if err != nil {
		return nil, err
	}
	for _, cr := range crs {
		for searchResult := new(clientutils.ResultItem); cr.NextRecord(searchResult) == nil; searchResult = new(clientutils.ResultItem) {
			if err != nil {
				return nil, err
			}

			tempResult := new(SearchResult)
			tempResult.Path = searchResult.Repo + "/"
			if searchResult.Path != "." {
				tempResult.Path += searchResult.Path + "/"
			}
			if searchResult.Name != "." {
				tempResult.Path += searchResult.Name
			}
			tempResult.Type = searchResult.Type
			tempResult.Size = searchResult.Size
			tempResult.Created = searchResult.Created
			tempResult.Modified = searchResult.Modified
			tempResult.Sha1 = searchResult.Actual_Sha1
			tempResult.Md5 = searchResult.Actual_Md5
			tempResult.Props = make(map[string][]string, len(searchResult.Properties))
			for _, prop := range searchResult.Properties {
				tempResult.Props[prop.Key] = append(tempResult.Props[prop.Key], prop.Value)
			}
			cw.Write(tempResult)
		}
		if err := cr.GetError(); err != nil {
			return nil, err
		}
		// TODO: Remove this
		// cr.Close()
	}
	cw.Close()
	return content.NewContentReader(cw.GetFilePath(), "results"), nil
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

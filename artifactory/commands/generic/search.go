package generic

import (
	"encoding/json"

	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientartutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
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
}

func NewSearchCommand() *SearchCommand {
	return &SearchCommand{GenericCommand: *NewGenericCommand()}
}

func SearchResultNoDate(reader *content.ContentReader) (*content.ContentReader, error) {
	writer, err := content.NewContentWriter("results", true, false)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	for resultItem := new(SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(SearchResult) {
		if err != nil {
			return nil, err
		}
		resultItem.Created = ""
		resultItem.Modified = ""
		delete(resultItem.Props, "vcs.url")
		delete(resultItem.Props, "vcs.revision")
		writer.Write(*resultItem)
	}
	if err := reader.GetError(); err != nil {
		return nil, err
	}
	reader.Reset()
	return content.NewContentReader(writer.GetFilePath(), writer.GetArrayKey()), nil
}

func (sc *SearchCommand) CommandName() string {
	return "rt_search"
}

func (sc *SearchCommand) Run() error {
	reader, err := sc.Search()
	sc.Result().SetReader(reader)
	return err
}

func (sc *SearchCommand) Search() (*content.ContentReader, error) {
	// Service Manager
	rtDetails, err := sc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	servicesManager, err := utils.CreateServiceManager(rtDetails, false)
	if err != nil {
		return nil, err
	}

	// Search Loop
	log.Info("Searching artifacts...")
	var searchResults []*content.ContentReader
	for i := 0; i < len(sc.Spec().Files); i++ {
		searchParams, err := GetSearchParams(sc.Spec().Get(i))
		if err != nil {
			log.Error(err)
			return nil, err
		}

		currentReader, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		searchResults = append(searchResults, currentReader)
	}
	defer func() {
		for _, reader := range searchResults {
			reader.Close()
		}
	}()
	reader, err := aqlResultToSearchResult(searchResults)
	if err != nil {
		return nil, err
	}
	length, err := reader.Length()
	clientartutils.LogSearchResults(length)
	return reader, err
}

func aqlResultToSearchResult(readers []*content.ContentReader) (*content.ContentReader, error) {
	writer, err := content.NewContentWriter("results", true, false)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	for _, reader := range readers {
		for searchResult := new(clientartutils.ResultItem); reader.NextRecord(searchResult) == nil; searchResult = new(clientartutils.ResultItem) {
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
			writer.Write(*tempResult)
		}
		if err := reader.GetError(); err != nil {
			return nil, err
		}
		reader.Reset()
	}
	return content.NewContentReader(writer.GetFilePath(), content.DefaultKey), nil
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

func PrintSearchResults(reader *content.ContentReader) error {
	length, err := reader.Length()
	if length == 0 {
		log.Output("[]")
		return err
	}
	log.Output("[")
	var prevSearchResult *SearchResult
	for searchResult := new(SearchResult); reader.NextRecord(searchResult) == nil; searchResult = new(SearchResult) {
		if prevSearchResult == nil {
			prevSearchResult = searchResult
			continue
		}
		performPrintSearchResults(*prevSearchResult, ",")
		prevSearchResult = searchResult
	}
	if prevSearchResult != nil {
		performPrintSearchResults(*prevSearchResult, "")
	}
	log.Output("]")
	reader.Reset()
	return reader.GetError()
}

func performPrintSearchResults(toPrint SearchResult, suffix string) error {
	data, err := json.Marshal(toPrint)
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Output("  " + clientutils.IndentJsonArray(data) + suffix)
	return nil
}

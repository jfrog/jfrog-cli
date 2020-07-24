package utils

import (
	"encoding/json"

	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/types"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func PrintSearchResults(reader *content.ContentReader) error {
	length, err := reader.Length()
	if length == 0 {
		log.Output("[]")
		return err
	}
	log.Output("[")
	suffix := ","
	for searchResult := new(types.SearchResult); reader.NextRecord(searchResult) == nil; searchResult = new(types.SearchResult) {
		if length == 1 {
			suffix = ""
		}
		printSearchResult(*searchResult, suffix)
		length--
	}
	log.Output("]")
	reader.Reset()
	return reader.GetError()
}

func printSearchResult(toPrint types.SearchResult, suffix string) error {
	data, err := json.Marshal(toPrint)
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Output("  " + clientutils.IndentJsonArray(data) + suffix)
	return nil
}

func AqlResultToSearchResult(readers []*content.ContentReader) (*content.ContentReader, error) {
	writer, err := content.NewContentWriter("results", true, false)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	for _, reader := range readers {
		for searchResult := new(utils.ResultItem); reader.NextRecord(searchResult) == nil; searchResult = new(utils.ResultItem) {
			if err != nil {
				return nil, err
			}
			tempResult := new(types.SearchResult)
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

func SearchResultNoDate(reader *content.ContentReader) (*content.ContentReader, error) {
	writer, err := content.NewContentWriter("results", true, false)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	for resultItem := new(types.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(types.SearchResult) {
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

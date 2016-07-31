package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

type AqlSearchResultItem struct {
	Repo 	   	string
	Path 	   	string
	Name 	   	string
	Actual_Md5  string
	Actual_Sha1 string
	Size 	   	int64
}

type AqlSearchResult struct {
	Results []AqlSearchResultItem
}

func AqlSearchDefaultReturnFields(pattern string, flags AqlSearchFlag) []AqlSearchResultItem {
	returnFields := []string{"\"name\"", "\"repo\"", "\"path\"", "\"actual_md5\"", "\"actual_sha1\"", "\"size\""}
	return AqlSearch(pattern, flags, returnFields)
}

func AqlSearch(pattern string, flags AqlSearchFlag, aqlReturnFields []string) []AqlSearchResultItem {
	aqlUrl := flags.GetArtifactoryDetails().Url + "api/search/aql"

	data := BuildAqlSearchQuery(pattern, flags.IsRecursive(), flags.GetProps(), aqlReturnFields)
	logger.Logger.Info("Searching Artifactory using AQL query: " + data)

	httpClientsDetails := GetArtifactoryHttpClientDetails(flags.GetArtifactoryDetails())
	resp, json := ioutils.SendPost(aqlUrl, []byte(data), httpClientsDetails)
	logger.Logger.Info("Artifactory response:", resp.Status)

	return parseAqlSearchResponse(json)
}

func parseAqlSearchResponse(resp []byte) []AqlSearchResultItem {
	var result AqlSearchResult
	err := json.Unmarshal(resp, &result)
	cliutils.CheckError(err)
	return result.Results
}

func (item AqlSearchResultItem) GetFullUrl() string {
	if item.Path == "." {
		return item.Repo + "/" + item.Name
	}

	url := item.Repo
	url = addSeparator(url, "/", item.Path)
	url = addSeparator(url, "/", item.Name)
	return url
}

func addSeparator(str1, separator, str2 string) string {
	if str2 == "" {
		return str1
	}
	if str1 == "" {
		return str2
	}

	return str1 + separator + str2
}

type AqlSearchFlag interface {
	GetArtifactoryDetails() *config.ArtifactoryDetails
	IsRecursive() bool
	GetProps() string
}
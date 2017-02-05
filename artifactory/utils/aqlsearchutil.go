package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"strings"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"errors"
)

func AqlSearchDefaultReturnFields(pattern string, recursive bool, props string, flags AqlSearchFlag) ([]AqlSearchResultItem, error) {
	query, err := BuildAqlSearchQuery(pattern, recursive, props, GetDefaultQueryReturnFields())
	if err != nil {
		return nil, err
	}
	return AqlSearch(query, flags)
}

func AqlSearchBySpec(aql Aql, flags AqlSearchFlag) ([]AqlSearchResultItem, error) {
	aqlString := aql.ItemsFind
	query := "items.find(" + aqlString + ").include(" + strings.Join(GetDefaultQueryReturnFields(), ",") + ")"
	return AqlSearch(query, flags)
}

func AqlSearch(aqlQuery string, flags AqlSearchFlag) ([]AqlSearchResultItem, error) {
	aqlUrl := flags.GetArtifactoryDetails().Url + "api/search/aql"
	log.Debug("Searching Artifactory using AQL query:", aqlQuery)

	httpClientsDetails := GetArtifactoryHttpClientDetails(flags.GetArtifactoryDetails())
	resp, json, err := ioutils.SendPost(aqlUrl, []byte(aqlQuery), httpClientsDetails)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, cliutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(json)))
	}

	log.Debug("Artifactory response: ", resp.Status)
	resultItems, err := parseAqlSearchResponse(json)
	return resultItems, err
}

func GetDefaultQueryReturnFields() []string {
	return []string{"\"name\"", "\"repo\"", "\"path\"", "\"actual_md5\"", "\"actual_sha1\"", "\"size\"", "\"property\""}
}

func LogSearchResults(numOfArtifacts int) {
	var msgSuffix = "artifacts."
	if numOfArtifacts == 1 {
		msgSuffix = "artifact."
	}
	log.Info("Found", strconv.Itoa(numOfArtifacts), msgSuffix)
}

func parseAqlSearchResponse(resp []byte) ([]AqlSearchResultItem, error) {
	var result AqlSearchResult
	err := json.Unmarshal(resp, &result)
	err = cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	return result.Results, nil
}

type AqlSearchResult struct {
	Results []AqlSearchResultItem
}

type AqlSearchResultItem struct {
	Repo        string
	Path        string
	Name        string
	Actual_Md5  string
	Actual_Sha1 string
	Size        int64
	Properties  []Property
}

type Property struct {
	Key   string
	Value string
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
}
package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"strings"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"errors"
	"sort"
)

func AqlSearchDefaultReturnFields(specFile *File, flags AqlSearchFlag) ([]AqlSearchResultItem, error) {
	query, err := createAqlBodyForItem(specFile)
	if err != nil {
		return nil, err
	}
	var dat map[string]interface{}
	json.Unmarshal([]byte(query), &dat)
	specFile.Aql = Aql{ItemsFind:dat}
	return AqlSearchBySpec(specFile, flags)
}

func AqlSearchBySpec(specFile *File, flags AqlSearchFlag) ([]AqlSearchResultItem, error) {
	aqlJson, _ := json.Marshal(specFile.Aql.ItemsFind)
	sortJson,_ := json.Marshal(specFile.Aql.Sort)
	aqlBody := string(aqlJson)
	sort := string(sortJson)
	limit := specFile.Aql.Limit
	query := "items.find(" + aqlBody + ")"
	query += ".include(" + strings.Join(GetDefaultQueryReturnFields(), ",") + ")"
	if sort != "" {
		query += ".sort(" + sort + ")"
	}
	if sort != "" {
		query += ".limit(" + strconv.Itoa(limit) + ")"
	}

	log.Info("AQL query = " +query)

	results, err := AqlSearch(query, flags)
	if err != nil {
		return nil, err
	}
	buildIdentifier := specFile.Build
	if buildIdentifier != "" && len(results) > 0 {
		results, err = filterSearchByBuild(buildIdentifier, results, flags)
		if err != nil {
			return nil, err
		}
	}
	return results, err
}

func AqlSearch(aqlQuery string, flags AqlSearchFlag) ([]AqlSearchResultItem, error) {
	json, err := execAqlSearch(aqlQuery, flags)
	if err != nil {
		return nil, err
	}

	resultItems, err := parseAqlSearchResponse(json)
	return resultItems, err
}

func execAqlSearch(aqlQuery string, flags AqlSearchFlag) ([]byte, error) {
	aqlUrl := flags.GetArtifactoryDetails().Url + "api/search/aql"
	log.Debug("Searching Artifactory using AQL query: ", aqlQuery)

	httpClientsDetails := GetArtifactoryHttpClientDetails(flags.GetArtifactoryDetails())
	resp, body, err := httputils.SendPost(aqlUrl, []byte(aqlQuery), httpClientsDetails)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, cliutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Artifactory response: ", resp.Status)
	return body, err
}

func GetDefaultQueryReturnFields() []string {
	return []string{"\"name\"", "\"repo\"", "\"path\"", "\"actual_md5\"", "\"actual_sha1\"", "\"size\"", "\"property\"", "\"type\""}
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
	if cliutils.CheckError(err) != nil {
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
	Type        string
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
	if item.Type == "folder" && !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
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

type AqlSearchResultItemFilter func(map[string]AqlSearchResultItem, []string) []AqlSearchResultItem

func FilterBottomChainResults(paths map[string]AqlSearchResultItem, pathsKeys []string) []AqlSearchResultItem {
	var result []AqlSearchResultItem
	sort.Sort(sort.Reverse(sort.StringSlice(pathsKeys)))
	for i, k := range pathsKeys {
		if i == 0 || !IsSubPath(pathsKeys, i, "/") {
			result = append(result, paths[k])
		}
	}

	return result
}

func FilterTopChainResults(paths map[string]AqlSearchResultItem, pathsKeys []string) []AqlSearchResultItem {
	sort.Strings(pathsKeys)
	for _, k := range pathsKeys {
		for _, k2 := range pathsKeys {
			prefix := k2
			if paths[k2].Type == "folder" &&  !strings.HasSuffix(k2, "/") {
				prefix += "/"
			}

			if k != k2 && strings.HasPrefix(k, prefix) {
				delete(paths, k)
				continue
			}
		}
	}

	var result []AqlSearchResultItem
	for _, v := range paths {
		result = append(result, v)
	}

	return result
}

// Reduce Dir results by using the resultsFilter
func ReduceDirResult(searchResults []AqlSearchResultItem, resultsFilter AqlSearchResultItemFilter) []AqlSearchResultItem {
	paths := make(map[string]AqlSearchResultItem)
	pathsKeys := make([]string, 0, len(searchResults))
	for _, file := range searchResults {
		if file.Name == "." {
			continue
		}

		url := file.GetFullUrl()
		paths[url] = file
		pathsKeys = append(pathsKeys, url)
	}
	return  resultsFilter(paths, pathsKeys)
}
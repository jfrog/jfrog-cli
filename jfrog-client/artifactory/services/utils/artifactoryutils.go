package utils

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const ARTIFACTORY_SYMLINK = "symlink.dest"
const SYMLINK_SHA1 = "symlink.destsha1"

func UploadFile(f *os.File, url string, artifactoryDetails auth.ArtifactoryDetails, details *fileutils.FileDetails,
	httpClientsDetails httputils.HttpClientDetails, client *httpclient.HttpClient) (*http.Response, []byte, error) {
	var err error
	if details == nil {
		details, err = fileutils.GetFileDetails(f.Name())
	}
	if err != nil {
		return nil, nil, err
	}
	headers := make(map[string]string)
	AddChecksumHeaders(headers, details)
	AddAuthHeaders(headers, artifactoryDetails)
	requestClientDetails := httpClientsDetails.Clone()
	utils.MergeMaps(headers, requestClientDetails.Headers)

	return client.UploadFile(f, url, *requestClientDetails)
}

func AddChecksumHeaders(headers map[string]string, fileDetails *fileutils.FileDetails) {
	AddHeader("X-Checksum-Sha1", fileDetails.Checksum.Sha1, &headers)
	AddHeader("X-Checksum-Md5", fileDetails.Checksum.Md5, &headers)
	if len(fileDetails.Checksum.Sha256) > 0 {
		AddHeader("X-Checksum", fileDetails.Checksum.Sha256, &headers)
	}
}

func AddAuthHeaders(headers map[string]string, artifactoryDetails auth.ArtifactoryDetails) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if artifactoryDetails.GetSshAuthHeaders() != nil {
		utils.MergeMaps(artifactoryDetails.GetSshAuthHeaders(), headers)
	}
}

func SetContentType(contentType string, headers *map[string]string) {
	AddHeader("Content-Type", contentType, headers)
}

func AddHeader(headerName, headerValue string, headers *map[string]string) {
	if *headers == nil {
		*headers = make(map[string]string)
	}
	(*headers)[headerName] = headerValue
}

func BuildArtifactoryUrl(baseUrl, path string, params map[string]string) (string, error) {
	u := url.URL{Path: path}
	escapedUrl, err := url.Parse(baseUrl + u.String())
	err = errorutils.CheckError(err)
	if err != nil {
		return "", err
	}
	q := escapedUrl.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	escapedUrl.RawQuery = q.Encode()
	return escapedUrl.String(), nil
}

func IsWildcardPattern(pattern string) bool {
	return strings.Contains(pattern, "*") || strings.HasSuffix(pattern, "/") || !strings.Contains(pattern, "/")
}

// @paths - sorted array
// @index - index of the current path which we want to check if it a prefix of any of the other previous paths
// @separator - file separator
// returns true paths[index] is a prefix of any of the paths[i] where i<index , otherwise returns false
func IsSubPath(paths []string, index int, separator string) bool {
	currentPath := paths[index]
	if !strings.HasSuffix(currentPath, separator) {
		currentPath += separator
	}
	for i := index - 1; i >= 0; i-- {
		if strings.HasPrefix(paths[i], currentPath) {
			return true
		}
	}
	return false
}

// This method parses buildIdentifier. buildIdentifier should be from the format "buildName/buildNumber".
// If no buildNumber provided LATEST wil be downloaded.
// If buildName or buildNumber contains "/" (slash) it should be escaped by "\" (backslash).
// Result examples of parsing: "aaa/123" > "aaa"-"123", "aaa" > "aaa"-"LATEST", "aaa\\/aaa" > "aaa/aaa"-"LATEST",  "aaa/12\\/3" > "aaa"-"12/3".
func getBuildNameAndNumber(buildIdentifier string, flags CommonConf) (string, string, error) {
	const Latest = "LATEST"
	const LastRelease = "LAST_RELEASE"
	buildName, buildNumber := parseBuildNameAndNumber(buildIdentifier)

	if buildNumber == Latest || buildNumber == LastRelease {
		return getBuildNumberFromArtifactory(buildName, buildNumber, flags)
	}
	return buildName, buildNumber, nil
}

func getBuildNameAndNumberFromProps(properties []Property) (buildName string, buildNumber string) {
	for _, property := range properties {
		if property.Key == "build.name" {
			buildName = property.Value
		} else if property.Key == "build.number" {
			buildNumber = property.Value
		}
		if len(buildName) > 0 && len(buildNumber) > 0 {
			return buildName, buildNumber
		}
	}
	return
}

func parseBuildNameAndNumber(buildIdentifier string) (buildName string, buildNumber string) {
	const Delimiter = "/"
	const EscapeChar = "\\"
	const Latest = "LATEST"

	if buildIdentifier == "" {
		return
	}
	if !strings.Contains(buildIdentifier, Delimiter) {
		log.Debug("No '" + Delimiter + "' is found in the build, build number is set to " + Latest)
		return buildIdentifier, Latest
	}
	buildNumberArray := []string{}
	buildAsArray := strings.Split(buildIdentifier, Delimiter)
	// The delimiter must not be prefixed with escapeChar (if it is, it should be part of the build number)
	// the code below gets substring from before the last delimiter.
	// If the new string ends with escape char it means the last delimiter was part of the build number and we need
	// to go back to the previous delimiter.
	// If no proper delimiter was found the full string will be the build name.
	for i := len(buildAsArray) - 1; i >= 1; i-- {
		buildNumberArray = append([]string{buildAsArray[i]}, buildNumberArray...)
		if !strings.HasSuffix(buildAsArray[i-1], EscapeChar) {
			buildName = strings.Join(buildAsArray[:i], Delimiter)
			buildNumber = strings.Join(buildNumberArray, Delimiter)
			break
		}
	}
	if buildName == "" {
		log.Debug("No delimiter char (" + Delimiter + ") without escaping char was found in the build, build number is set to " + Latest)
		buildName = buildIdentifier
		buildNumber = Latest
	}
	// Remove escape chars
	buildName = strings.Replace(buildName, "\\/", "/", -1)
	buildNumber = strings.Replace(buildNumber, "\\/", "/", -1)
	return buildName, buildNumber
}

type build struct {
	BuildName   string `json:"buildName"`
	BuildNumber string `json:"buildNumber"`
}

func getBuildNumberFromArtifactory(buildName, buildNumber string, flags CommonConf) (string, string, error) {
	restUrl := flags.GetArtifactoryDetails().GetUrl() + "api/build/patternArtifacts"
	body, err := createBodyForLatestBuildRequest(buildName, buildNumber)
	if err != nil {
		return "", "", err
	}
	log.Debug("Getting build name and number from Artifactory: " + buildName + ", " + buildNumber)
	httpClientsDetails := flags.GetArtifactoryDetails().CreateHttpClientDetails()
	SetContentType("application/json", &httpClientsDetails.Headers)
	log.Debug("Sending post request to: " + restUrl + ", with the following body: " + string(body))
	resp, body, err := httputils.SendPost(restUrl, body, httpClientsDetails)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + utils.IndentJson(body)))
	}
	log.Debug("Artifactory response: ", resp.Status)
	var responseBuild []build
	err = json.Unmarshal(body, &responseBuild)
	if errorutils.CheckError(err) != nil {
		return "", "", err
	}
	if responseBuild[0].BuildNumber != "" {
		log.Debug("Found build number: " + responseBuild[0].BuildNumber)
	} else {
		log.Debug("The build could not be found in Artifactory")
	}

	return buildName, responseBuild[0].BuildNumber, nil
}

func createBodyForLatestBuildRequest(buildName, buildNumber string) (body []byte, err error) {
	buildJsonArray := []build{{buildName, buildNumber}}
	body, err = json.Marshal(buildJsonArray)
	errorutils.CheckError(err)
	return
}

func filterSearchByBuild(specFile *ArtifactoryCommonParams, itemsToFilter []ResultItem, flags CommonConf) ([]ResultItem, error) {
	buildName, buildNumber, err := getBuildNameAndNumber(specFile.Build, flags)
	if err != nil {
		return nil, err
	}

	buildAqlResponse, err := fetchBuildArtifactsSha1(specFile.Aql.ItemsFind, buildName, buildNumber, itemsToFilter, flags)
	if err != nil {
		return nil, err
	}

	return filterBuildAqlSearchResults(&itemsToFilter, &buildAqlResponse, buildName, buildNumber), err
}

// This function adds to @itemsToFilter the "build.name" property and returns all the artifacts that associated with
// the provided @buildName and @buildNumber.
func fetchBuildArtifactsSha1(aqlBody, buildName, buildNumber string, itemsToFilter []ResultItem, flags CommonConf) (map[string]bool, error) {
	var wg sync.WaitGroup
	var addPropsErr error
	var aqlSearchErr error
	var buildAqlResponse []byte

	wg.Add(2)
	go func() {
		addPropsErr = searchAndAddPropsToAqlResult(itemsToFilter, aqlBody, "build.name", buildName, flags)
		wg.Done()
	}()

	go func() {
		buildQuery := createAqlQueryForBuild(buildName, buildNumber)
		buildAqlResponse, aqlSearchErr = ExecAql(buildQuery, flags)
		wg.Done()
	}()

	wg.Wait()

	if aqlSearchErr != nil {
		return nil, aqlSearchErr
	}
	if addPropsErr != nil {
		return nil, addPropsErr
	}

	buildArtifactsSha, err := extractSha1FromAqlResponse(buildAqlResponse)
	if err != nil {
		return nil, err
	}

	return buildArtifactsSha, nil
}

func searchAndAddPropsToAqlResult(itemsToFilter []ResultItem, aqlBody, filterByPropName, filterByPropValue string, flags CommonConf) error {
	propsAqlResponseJson, err := ExecAql(createPropsQuery(aqlBody, filterByPropName, filterByPropValue), flags)
	if err != nil {
		return err
	}
	propsAqlResponse, err := parseAqlSearchResponse(propsAqlResponseJson)
	if err != nil {
		return err
	}
	addPropsToAqlResult(itemsToFilter, propsAqlResponse)
	return nil
}

func addPropsToAqlResult(items []ResultItem, props []ResultItem) {
	propsMap := createPropsMap(props)
	for i := range items {
		props, propsExists := propsMap[getResultItemKey(items[i])]
		if propsExists {
			items[i].Properties = props
		}
	}
}

func createPropsMap(items []ResultItem) (propsMap map[string][]Property) {
	propsMap = make(map[string][]Property)
	for _, item := range items {
		propsMap[getResultItemKey(item)] = item.Properties
	}
	return
}

func getResultItemKey(item ResultItem) string {
	return item.Repo + item.Path + item.Name + item.Actual_Sha1
}

func extractSha1FromAqlResponse(resp []byte) (map[string]bool, error) {
	elements, err := parseAqlSearchResponse(resp)
	if err != nil {
		return nil, err
	}
	elementsMap := make(map[string]bool)
	for _, element := range elements {
		elementsMap[element.Actual_Sha1] = true
	}
	return elementsMap, nil
}

/*
 * Filter search results by the following priorities:
 * 1st priority: Match {Sha1, build name, build number}
 * 2nd priority: Match {Sha1, build name}
 * 3rd priority: Match {Sha1}
 */
func filterBuildAqlSearchResults(itemsToFilter *[]ResultItem, buildArtifactsSha *map[string]bool, buildName, buildNumber string) []ResultItem {
	filteredResults := []ResultItem{}
	firstPriority := map[string][]ResultItem{}
	secondPriority := map[string][]ResultItem{}
	thirdPriority := map[string][]ResultItem{}

	// Step 1 - Populate 3 priorities mappings.
	for _, item := range *itemsToFilter {
		if _, ok := (*buildArtifactsSha)[item.Actual_Sha1]; !ok {
			continue
		}
		resultBuildName, resultBuildNumber := getBuildNameAndNumberFromProps(item.Properties)
		isBuildNameMatched := resultBuildName == buildName
		if isBuildNameMatched && resultBuildNumber == buildNumber {
			firstPriority[item.Actual_Sha1] = append(firstPriority[item.Actual_Sha1], item)
			continue
		}
		if isBuildNameMatched {
			secondPriority[item.Actual_Sha1] = append(secondPriority[item.Actual_Sha1], item)
			continue
		}
		thirdPriority[item.Actual_Sha1] = append(thirdPriority[item.Actual_Sha1], item)
	}

	// Step 2 - Append mappings to the final results, respectively.
	for shaToMatch := range *buildArtifactsSha {
		if _, ok := firstPriority[shaToMatch]; ok {
			filteredResults = append(filteredResults, firstPriority[shaToMatch]...)
		} else if _, ok := secondPriority[shaToMatch]; ok {
			filteredResults = append(filteredResults, secondPriority[shaToMatch]...)
		} else if _, ok := thirdPriority[shaToMatch]; ok {
			filteredResults = append(filteredResults, thirdPriority[shaToMatch]...)
		}
	}

	return filteredResults
}

type CommonConf interface {
	GetArtifactoryDetails() auth.ArtifactoryDetails
	SetArtifactoryDetails(rt auth.ArtifactoryDetails)
	GetJfrogHttpClient() *httpclient.HttpClient
	IsDryRun() bool
}

type CommonConfImpl struct {
	artDetails auth.ArtifactoryDetails
	DryRun     bool
}

func (flags *CommonConfImpl) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return flags.artDetails
}

func (flags *CommonConfImpl) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	flags.artDetails = rt
}

func (flags *CommonConfImpl) IsDryRun() bool {
	return flags.DryRun
}

func (flags *CommonConfImpl) GetJfrogHttpClient() *httpclient.HttpClient {
	return httpclient.NewDefaultHttpClient()
}

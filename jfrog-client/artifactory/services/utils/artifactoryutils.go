package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"net/http"
	"net/url"
	"os"
	"errors"
	"strings"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/types/httpclient"
)

const ARTIFACTORY_SYMLINK = "symlink.dest"
const SYMLINK_SHA1 = "symlink.destsha1"

func UploadFile(f *os.File, url string, artifactoryDetails *auth.ArtifactoryDetails, details *fileutils.FileDetails,
	httpClientsDetails httputils.HttpClientDetails, client *httpclient.HttpClient) (*http.Response, []byte, error) {
	var err error
	if details == nil {
		details, err = fileutils.GetFileDetails(f.Name())
	}
	if err != nil {
		return nil, nil, err
	}
	headers := map[string]string{
		"X-Checksum":      details.Checksum.Sha256,
		"X-Checksum-Sha1": details.Checksum.Sha1,
		"X-Checksum-Md5":  details.Checksum.Md5,
	}
	AddAuthHeaders(headers, artifactoryDetails)
	requestClientDetails := httpClientsDetails.Clone()
	utils.MergeMaps(headers, requestClientDetails.Headers)

	return client.UploadFile(f, url, *requestClientDetails)
}

func AddAuthHeaders(headers map[string]string, artifactoryDetails *auth.ArtifactoryDetails) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}
	if artifactoryDetails.GetSshAuthHeaders() != nil {
		utils.MergeMaps(artifactoryDetails.GetSshAuthHeaders(), headers)
	}
	return headers
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
	escapedUrl, err := url.Parse(baseUrl + path)
	err = errorutils.CheckError(err)
	if err != nil {
		return "", nil
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

func EncodeParams(props string) (string, error) {
	propList := strings.Split(props, ";")
	result := []string{}
	for _, prop := range propList {
		if prop == "" {
			continue
		}
		key, value, err := SplitProp(prop)
		if err != nil {
			return "", err
		}
		result = append(result, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}

	return strings.Join(result, ";"), nil
}

func SplitProp(prop string) (string, string, error) {
	splitIndex := strings.Index(prop, "=")
	if splitIndex < 1 || len(prop[splitIndex+1:]) < 1 {
		err := errorutils.CheckError(errors.New("Invalid property: " + prop))
		return "", "", err
	}
	return prop[:splitIndex], prop[splitIndex+1:], nil

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
	restUrl := flags.GetArtifactoryDetails().Url + "api/build/patternArtifacts"
	body, err := createBodyForLatestBuildRequest(buildName, buildNumber)
	if err != nil {
		return "", "", err
	}
	log.Debug("Getting build name and number from Artifactory: " + buildName + ", " + buildNumber)
	httpClientsDetails := flags.GetArtifactoryDetails().CreateArtifactoryHttpClientDetails()
	SetContentType("application/json", &httpClientsDetails.Headers)
	log.Debug("Sending post request to: " + restUrl + ", with the following body: " + string(body))
	resp, body, err := httputils.SendPost(restUrl, body, httpClientsDetails)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != 200 {
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

func filterSearchByBuild(buildIdentifier string, resultsToFilter []ResultItem, flags CommonConf) ([]ResultItem, error) {
	buildName, buildNumber, err := getBuildNameAndNumber(buildIdentifier, flags)
	if err != nil {
		return nil, err
	}
	query := createAqlQueryForBuild(buildName, buildNumber)
	aqlResponse, err := execAqlSearch(query, flags)
	if err != nil {
		return nil, err
	}
	buildArtifactsSha, err := extractSearchResponseShas(aqlResponse)
	if err != nil {
		return nil, err
	}

	return filterSearchResultBySha(resultsToFilter, buildArtifactsSha), err
}

func extractSearchResponseShas(resp []byte) (map[string]bool, error) {
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

func filterSearchResultBySha(aqlSearchResultItemsToFilter []ResultItem, shasToMatch map[string]bool) (filteredResults []ResultItem) {
	for _, resultToFilter := range aqlSearchResultItemsToFilter {
		if _, matched := shasToMatch[resultToFilter.Actual_Sha1]; matched {
			filteredResults = append(filteredResults, resultToFilter)
		}
	}
	return
}

type CommonConf interface {
	GetArtifactoryDetails() *auth.ArtifactoryDetails
	SetArtifactoryDetails(rt *auth.ArtifactoryDetails)
	GetJfrogHttpClient() *httpclient.HttpClient
	IsDryRun() bool
}

type CommonConfImpl struct {
	artDetails *auth.ArtifactoryDetails
	DryRun     bool
}

func (flags *CommonConfImpl) GetArtifactoryDetails() *auth.ArtifactoryDetails {
	return flags.artDetails
}

func (flags *CommonConfImpl) SetArtifactoryDetails(rt *auth.ArtifactoryDetails) {
	flags.artDetails = rt
}

func (flags *CommonConfImpl) IsDryRun() bool {
	return flags.DryRun
}

func (flags *CommonConfImpl) GetJfrogHttpClient() *httpclient.HttpClient {
	return httpclient.NewDefaultHttpClient()
}

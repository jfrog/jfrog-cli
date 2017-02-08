package utils

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"os"
	"io/ioutil"
	"bytes"
	"time"
	"strings"
	"net/http"
	"errors"
)

const BUILD_INFO_DETAILS = "details"

func getBuildDir(buildName, buildNumber string) (string, error) {
	tempDir := os.TempDir()
	buildsDir := tempDir + "/jfrog/builds/" + buildName + "_" + buildNumber + "/"
	err := os.MkdirAll(buildsDir, 0777)
	if cliutils.CheckError(err) != nil {
		return "", err
	}
	return buildsDir, nil
}

func saveBuildData(action interface{}, buildName, buildNumber string) (err error) {
	b, err := json.Marshal(&action)
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}
	dirPath, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	tmpfile, err := ioutil.TempFile(dirPath, "temp")
	if err != nil {
		return err
	}
	defer tmpfile.Close()
	_, err = tmpfile.Write([]byte(content.String()))
	return
}

func SaveBuildGeneralDetails(buildName, buildNumber string) error {
	path, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	path += BUILD_INFO_DETAILS
	var exists bool
	exists, err = ioutils.IsFileExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	meta := BuildGeneralDetails{
		Timestamp: time.Now(),
	}
	b, err := json.Marshal(&meta)
	err = cliutils.CheckError(err)
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, []byte(content.String()), 0600)
	return err
}

type populateBuildInfoWrapper func(*ArtifactBuildInfoWrapper)

func SavePartialBuildInfo(buildName, buildNumber string, populateDataFunc populateBuildInfoWrapper) error {
	tempBuildInfo := new(ArtifactBuildInfoWrapper)
	tempBuildInfo.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	populateDataFunc(tempBuildInfo)
	return saveBuildData(tempBuildInfo, buildName, buildNumber)
}

func ReadBuildInfoFiles(buildName, buildNumber string) (BuildInfo, error) {
	var buildInfo []*ArtifactBuildInfoWrapper
	path, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	buildFiles, err := ioutils.ListFiles(path)
	if err != nil {
		return nil, err
	}
	for _, buildFile := range buildFiles {
		dir, err := ioutils.IsDir(buildFile)
		if err != nil {
			return nil, err
		}
		if dir {
			continue
		}
		if strings.HasSuffix(buildFile, BUILD_INFO_DETAILS) {
			continue
		}
		content, err := ioutils.ReadFile(buildFile)
		if err != nil {
			return nil, err
		}
		atifactBuildInfoWrapper := new(ArtifactBuildInfoWrapper)
		json.Unmarshal(content, &atifactBuildInfoWrapper)
		buildInfo = append(buildInfo, atifactBuildInfoWrapper)
	}

	return buildInfo, nil
}

func ReadBuildInfoGeneralDetails(buildName, buildNumber string) (*BuildGeneralDetails, error) {
	path, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	path += BUILD_INFO_DETAILS
	content, err := ioutils.ReadFile(path)
	if err != nil {
		return nil, err
	}
	details := new(BuildGeneralDetails)
	json.Unmarshal(content, &details)
	return details, nil
}

func PublishBuildInfo(url string, content []byte, httpClientsDetails ioutils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	return ioutils.SendPut(url + "api/build/", content, httpClientsDetails)
}

type BuildEnv map[string]string

type BuildInfoCommon struct {
	Sha1 string `json:"sha1,omitempty"`
	Md5  string `json:"md5,omitempty"`
}

type ArtifactsBuildInfo struct {
	Name string `json:"name,omitempty"`
	*BuildInfoCommon
}

type DependenciesBuildInfo struct {
	Id string `json:"id,omitempty"`
	*BuildInfoCommon
}

type BuildInfoAction string

type ArtifactBuildInfoWrapper struct {
	Artifacts    []ArtifactsBuildInfo     `json:"Artifacts,omitempty"`
	Dependencies []DependenciesBuildInfo `json:"Dependencies,omitempty"`
	Env          BuildEnv                `json:"Env,omitempty"`
	Timestamp    int64                   `json:"Timestamp,omitempty"`
}

type BuildGeneralDetails struct {
	Timestamp time.Time `json:"Timestamp,omitempty"`
}

type BuildInfo []*ArtifactBuildInfoWrapper

func (wrapper BuildInfo) Len() int {
	return len(wrapper)
}

func (wrapper BuildInfo) Less(i, j int) bool {
	return wrapper[i].Timestamp < wrapper[j].Timestamp;
}

func (wrapper BuildInfo) Swap(i, j int) {
	wrapper[i], wrapper[j] = wrapper[j], wrapper[i]
}

func RemoveBuildDir(buildName, buildNumber string) error {
	tempDirPath, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	exists, err := ioutils.IsDirExists(tempDirPath)
	if err != nil {
		return err
	}
	if exists {
		return cliutils.CheckError(os.RemoveAll(tempDirPath))
	}
	return nil
}

type BuildInfoFlags struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
	EnvInclude string
	EnvExclude string
}

type build struct {
	BuildName	string `json:"buildName"`
	BuildNumber 	string `json:"buildNumber"`
}

func (flags *BuildInfoFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *BuildInfoFlags) IsDryRun() bool {
	return flags.DryRun
}

// This method parses buildIdentifier. buildIdentifier should be from the format "buildName/buildNumber".
// If no buildNumber provided LATEST wil be downloaded.
// If buildName or buildNumber contains "/" (slash) it should be escaped by "\" (backslash).
// Result examples of parsing: "aaa/123" > "aaa"-"123", "aaa" > "aaa"-"LATEST", "aaa\\/aaa" > "aaa/aaa"-"LATEST",  "aaa/12\\/3" > "aaa"-"12/3".
func getBuildNameAndNumber(buildIdentifier string, flags AqlSearchFlag) (string, string, error) {
	const Latest = "LATEST"
	const LastRelease  = "LAST_RELEASE"
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
		buildNumberArray = append([]string{buildAsArray[i]},  buildNumberArray...)
		if !strings.HasSuffix(buildAsArray[i - 1], EscapeChar) {
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

func getBuildNumberFromArtifactory(buildName, buildNumber string, flags AqlSearchFlag) (string, string, error) {
	restUrl := flags.GetArtifactoryDetails().Url + "api/build/patternArtifacts"
	body, err := createBodyForLatestBuildRequest(buildName, buildNumber)
	if err != nil {
		return "", "", err
	}
	log.Debug("Getting build name and number from Artifactory: " + buildName + ", " + buildNumber)
	httpClientsDetails := GetArtifactoryHttpClientDetails(flags.GetArtifactoryDetails())
	SetContentType("application/json", &httpClientsDetails.Headers)
	log.Debug("Sending post request to: " + restUrl + ", with the following body: " + string(body))
	resp, body, err := ioutils.SendPost(restUrl, body, httpClientsDetails)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != 200 {
		return "", "", cliutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}
	log.Debug("Artifactory response: ", resp.Status)
	var responseBuild []build
	err = json.Unmarshal(body, &responseBuild)
	if cliutils.CheckError(err) != nil {
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
	cliutils.CheckError(err)
	return
}

func filterSearchByBuild(buildIdentifier string, resultsToFilter []AqlSearchResultItem, flags AqlSearchFlag) ([]AqlSearchResultItem, error) {
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

	return filterSearchResultBySha(resultsToFilter, buildArtifactsSha) ,err
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
	return elementsMap , nil
}

func filterSearchResultBySha(aqlSearchResultItemsToFilter []AqlSearchResultItem, shasToMatch map[string]bool) (filteredResults []AqlSearchResultItem) {
	for _, resultToFilter := range aqlSearchResultItemsToFilter {
		if _, matched := shasToMatch[resultToFilter.Actual_Sha1]; matched {
			filteredResults = append(filteredResults, resultToFilter)
		}
	}
	return
}
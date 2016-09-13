package utils

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"os"
	"io/ioutil"
	"bytes"
	"time"
	"strings"
	"net/http"
)

const BUILD_INFO_DETAILS = "details"

func GetBuildDir(buildName, buildNumber string) (string, error) {
	tempDir := os.TempDir()
	buildsDir := tempDir + "/jfrog/builds/" + buildName + "/" + buildNumber + "/"
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
	dirPath, err := GetBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	tmpfile, err := ioutil.TempFile(dirPath, "temp")
	if err != nil {
		return err
	}
	_, err = tmpfile.Write([]byte(content.String()))
	return
}

func SaveBuildGeneralDetails(buildName, buildNumber string) error {
	path, err := GetBuildDir(buildName, buildNumber)
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

func PrepareBuildInfoForSave(buildName, buildNumber string, populateDataFunc populateBuildInfoWrapper) error {
	tempBuildInfo := new(ArtifactBuildInfoWrapper)
	tempBuildInfo.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	populateDataFunc(tempBuildInfo)
	return saveBuildData(tempBuildInfo, buildName, buildNumber)
}

func ReadBuildInfoFiles(buildName, buildNumber string) (BuildInfo, error) {
	var buildInfo []*ArtifactBuildInfoWrapper
	path, err := GetBuildDir(buildName, buildNumber)
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
	path, err := GetBuildDir(buildName, buildNumber)
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

type ArtifactBuildInfo struct {
	Name string `json:"name,omitempty"`
	*BuildInfoCommon
}

type DependenciesBuildInfo struct {
	Id string `json:"id,omitempty"`
	*BuildInfoCommon
}

type BuildInfoAction string

type ArtifactBuildInfoWrapper struct {
	Artifacts    []ArtifactBuildInfo     `json:"Artifacts,omitempty"`
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
	tempDirPath, err := GetBuildDir(buildName, buildNumber)
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

func (flags *BuildInfoFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *BuildInfoFlags) IsDryRun() bool {
	return flags.DryRun
}
package utils

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"os"
	"io/ioutil"
	"bytes"
	"time"
	"strings"
	"net/http"
	"encoding/base64"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth/cert"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/types/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
)

const BUILD_INFO_DETAILS = "details"

type ArtifactsBuildInfo struct {
	*clientutils.FileHashes
	Name string `json:"name,omitempty"`
}

type DependenciesBuildInfo struct {
	*clientutils.FileHashes
	Id string `json:"id,omitempty"`
}

func getBuildDir(buildName, buildNumber string) (string, error) {
	tempDir := os.TempDir()
	encodedDirName := base64.StdEncoding.EncodeToString([]byte(buildName + "_" + buildNumber))
	buildsDir := tempDir + "/jfrog/builds/" + encodedDirName + "/"
	err := os.MkdirAll(buildsDir, 0777)
	if errorutils.CheckError(err) != nil {
		return "", err
	}
	return buildsDir, nil
}

func saveBuildData(action interface{}, buildName, buildNumber string) (err error) {
	b, err := json.Marshal(&action)
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}
	dirPath, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	cliutils.CliLogger.Debug("Creating temp build file at: " + dirPath)
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
	exists, err = fileutils.IsFileExists(path)
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
	err = errorutils.CheckError(err)
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	err = errorutils.CheckError(err)
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

func ReadBuildInfoFiles(buildName, buildNumber string) (BuildInfoData, error) {
	var BuildInfoPartialData []*ArtifactBuildInfoWrapper
	path, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	buildFiles, err := fileutils.ListFiles(path, false)
	if err != nil {
		return nil, err
	}
	for _, buildFile := range buildFiles {
		dir, err := fileutils.IsDir(buildFile)
		if err != nil {
			return nil, err
		}
		if dir {
			continue
		}
		if strings.HasSuffix(buildFile, BUILD_INFO_DETAILS) {
			continue
		}
		content, err := fileutils.ReadFile(buildFile)
		if err != nil {
			return nil, err
		}
		artifactBuildInfoWrapper := new(ArtifactBuildInfoWrapper)
		json.Unmarshal(content, &artifactBuildInfoWrapper)
		BuildInfoPartialData = append(BuildInfoPartialData, artifactBuildInfoWrapper)
	}

	return BuildInfoPartialData, nil
}

func ReadBuildInfoGeneralDetails(buildName, buildNumber string) (*BuildGeneralDetails, error) {
	path, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	path += BUILD_INFO_DETAILS
	content, err := fileutils.ReadFile(path)
	if err != nil {
		return nil, err
	}
	details := new(BuildGeneralDetails)
	json.Unmarshal(content, &details)
	return details, nil
}

func PublishBuildInfo(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, error) {
	securityDir, err := GetJfrogSecurityDir()
	if err != nil {
		return nil, nil, err
	}
	transport, err := cert.GetTransportWithLoadedCert(securityDir)
	client := httpclient.NewHttpClient(&http.Client{Transport: transport})
	return client.SendPut(url+"api/build/", content, httpClientsDetails)
}

type BuildInfoAction string
type BuildEnv map[string]string

type ArtifactBuildInfoWrapper struct {
	Artifacts    []ArtifactsBuildInfo    `json:"Artifacts,omitempty"`
	Dependencies []DependenciesBuildInfo `json:"Dependencies,omitempty"`
	Env          BuildEnv                `json:"Env,omitempty"`
	Timestamp    int64                               `json:"Timestamp,omitempty"`
	*Vcs
}

type Vcs struct {
	VcsUrl      string `json:"vcsUrl,omitempty"`
	VcsRevision string `json:"vcsRevision,omitempty"`
}

type BuildGeneralDetails struct {
	Timestamp time.Time `json:"Timestamp,omitempty"`
}

type BuildInfoData []*ArtifactBuildInfoWrapper

func (wrapper BuildInfoData) Len() int {
	return len(wrapper)
}

func (wrapper BuildInfoData) Less(i, j int) bool {
	return wrapper[i].Timestamp < wrapper[j].Timestamp;
}

func (wrapper BuildInfoData) Swap(i, j int) {
	wrapper[i], wrapper[j] = wrapper[j], wrapper[i]
}

func RemoveBuildDir(buildName, buildNumber string) error {
	tempDirPath, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	exists, err := fileutils.IsDirExists(tempDirPath)
	if err != nil {
		return err
	}
	if exists {
		return errorutils.CheckError(os.RemoveAll(tempDirPath))
	}
	return nil
}

type BuildInfoFlags struct {
	artDetails *auth.ArtifactoryDetails
	DryRun     bool
	EnvInclude string
	EnvExclude string
}

func (flags *BuildInfoFlags) GetArtifactoryDetails() *auth.ArtifactoryDetails {
	return flags.artDetails
}

func (flags *BuildInfoFlags) SetArtifactoryDetails(art *auth.ArtifactoryDetails) {
	flags.artDetails = art
}

func (flags *BuildInfoFlags) IsDryRun() bool {
	return flags.DryRun
}

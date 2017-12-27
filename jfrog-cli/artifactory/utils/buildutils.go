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
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth/cert"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"path/filepath"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

const BuildInfoDetails = "details"
const BuildTempPath = "jfrog/builds/"

func getBuildDir(buildName, buildNumber string) (string, error) {
	tempDir := os.TempDir()
	encodedDirName := base64.StdEncoding.EncodeToString([]byte(buildName + "_" + buildNumber))
	buildsDir := filepath.Join(tempDir, BuildTempPath, encodedDirName)
	err := os.MkdirAll(buildsDir, 0777)
	if errorutils.CheckError(err) != nil {
		return "", err
	}
	return buildsDir, nil
}

func getPartialsBuildDir(buildName, buildNumber string) (string, error) {
	buildDir, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return "", err
	}
	buildDir = filepath.Join(buildDir, "partials")
	err = os.MkdirAll(buildDir, 0777)
	if errorutils.CheckError(err) != nil {
		return "", err
	}
	return buildDir, nil
}

func saveBuildData(action interface{}, buildName, buildNumber string) error {
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
	dirPath, err := getPartialsBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	log.Debug("Creating temp build file at:", dirPath)
	tempFile, err := ioutil.TempFile(dirPath, "temp")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	_, err = tempFile.Write([]byte(content.String()))
	return err
}

func SaveBuildGeneralDetails(buildName, buildNumber string) error {
	partialsBuildDir, err := getPartialsBuildDir(buildName, buildNumber)
	log.Debug("Saving build general details at: " + partialsBuildDir)
	if err != nil {
		return err
	}
	detailsFilePath := filepath.Join(partialsBuildDir, BuildInfoDetails)
	var exists bool
	exists, err = fileutils.IsFileExists(detailsFilePath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	meta := buildinfo.General{
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
	err = ioutil.WriteFile(detailsFilePath, []byte(content.String()), 0600)
	return err
}

type populatePartialBuildInfo func(partial *buildinfo.Partial)

func SavePartialBuildInfo(buildName, buildNumber string, populatePartialBuildInfoFunc populatePartialBuildInfo) error {
	partialBuildInfo := new(buildinfo.Partial)
	partialBuildInfo.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	populatePartialBuildInfoFunc(partialBuildInfo)
	return saveBuildData(partialBuildInfo, buildName, buildNumber)
}

func GetGeneratedBuildsInfo(buildName, buildNumber string) ([]*buildinfo.BuildInfo, error) {
	buildDir, err := getBuildDir(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	buildFiles, err := fileutils.ListFiles(buildDir, false)
	if err != nil {
		return nil, err
	}

	var generatedBuildsInfo []*buildinfo.BuildInfo
	for _, buildFile := range buildFiles {
		dir, err := fileutils.IsDir(buildFile)
		if err != nil {
			return nil, err
		}
		if dir {
			continue
		}
		content, err := fileutils.ReadFile(buildFile)
		if err != nil {
			return nil, err
		}
		buildInfo := new(buildinfo.BuildInfo)
		json.Unmarshal(content, &buildInfo)
		generatedBuildsInfo = append(generatedBuildsInfo, buildInfo)
	}
	return generatedBuildsInfo, nil
}

func ReadPartialBuildInfoFiles(buildName, buildNumber string) (buildinfo.Partials, error) {
	var partials buildinfo.Partials
	partialsBuildDir, err := getPartialsBuildDir(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	buildFiles, err := fileutils.ListFiles(partialsBuildDir, false)
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
		if strings.HasSuffix(buildFile, BuildInfoDetails) {
			continue
		}
		content, err := fileutils.ReadFile(buildFile)
		if err != nil {
			return nil, err
		}
		partial := new(buildinfo.Partial)
		json.Unmarshal(content, &partial)
		partials = append(partials, partial)
	}

	return partials, nil
}

func ReadBuildInfoGeneralDetails(buildName, buildNumber string) (*buildinfo.General, error) {
	partialsBuildDir, err := getPartialsBuildDir(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	generalDetailsFilePath := filepath.Join(partialsBuildDir, BuildInfoDetails)
	content, err := fileutils.ReadFile(generalDetailsFilePath)
	if err != nil {
		return nil, err
	}
	details := new(buildinfo.General)
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

type BuildConfigFlags struct {
	BuildName   string
	BuildNumber string
}

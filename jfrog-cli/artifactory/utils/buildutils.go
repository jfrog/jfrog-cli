package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const BuildInfoDetails = "details"
const BuildTempPath = "jfrog/builds/"

func GetBuildDir(buildName, buildNumber string) (string, error) {
	tempDir := os.TempDir()
	encodedDirName := base64.StdEncoding.EncodeToString([]byte(buildName + "_" + buildNumber))
	buildsDir := filepath.Join(tempDir, BuildTempPath, encodedDirName)
	err := os.MkdirAll(buildsDir, 0777)
	if errorutils.CheckError(err) != nil {
		return "", err
	}
	return buildsDir, nil
}

func CreateBuildProperties(buildName, buildNumber string) (string, error) {
	if buildName == "" || buildNumber == "" {
		return "", nil
	}
	buildGeneralDetails, err := ReadBuildInfoGeneralDetails(buildName, buildNumber)
	if err != nil {
		return fmt.Sprintf("build.name=%s;build.number=%s", buildName, buildNumber), err
	}
	timestamp := strconv.FormatInt(buildGeneralDetails.Timestamp.UnixNano()/int64(time.Millisecond), 10)
	return fmt.Sprintf("build.name=%s;build.number=%s;build.timestamp=%s", buildName, buildNumber, timestamp), nil
}

func getPartialsBuildDir(buildName, buildNumber string) (string, error) {
	buildDir, err := GetBuildDir(buildName, buildNumber)
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
	if errorutils.CheckError(err) != nil {
		return err
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	if errorutils.CheckError(err) != nil {
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

func SaveBuildInfo(buildName, buildNumber string, buildInfo *buildinfo.BuildInfo) error {
	b, err := json.Marshal(buildInfo)
	if errorutils.CheckError(err) != nil {
		return err
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	if errorutils.CheckError(err) != nil {
		return err
	}
	dirPath, err := GetBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	log.Debug("Creating temp build file at: " + dirPath)
	tempFile, err := ioutil.TempFile(dirPath, "temp")
	if errorutils.CheckError(err) != nil {
		return err
	}
	defer tempFile.Close()
	_, err = tempFile.Write([]byte(content.String()))
	return errorutils.CheckError(err)
}

func SaveBuildGeneralDetails(buildName, buildNumber string) error {
	partialsBuildDir, err := getPartialsBuildDir(buildName, buildNumber)
	log.Debug("Saving build general details at: " + partialsBuildDir)
	if err != nil {
		return err
	}
	detailsFilePath := filepath.Join(partialsBuildDir, BuildInfoDetails)
	var exists bool
	exists, err = fileutils.IsFileExists(false, detailsFilePath)
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
	if err != nil {
		return errorutils.CheckError(err)
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	if err != nil {
		return errorutils.CheckError(err)
	}
	err = ioutil.WriteFile(detailsFilePath, []byte(content.String()), 0600)
	return errorutils.CheckError(err)
}

type populatePartialBuildInfo func(partial *buildinfo.Partial)

func SavePartialBuildInfo(buildName, buildNumber string, populatePartialBuildInfoFunc populatePartialBuildInfo) error {
	partialBuildInfo := new(buildinfo.Partial)
	partialBuildInfo.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	populatePartialBuildInfoFunc(partialBuildInfo)
	return saveBuildData(partialBuildInfo, buildName, buildNumber)
}

func GetGeneratedBuildsInfo(buildName, buildNumber string) ([]*buildinfo.BuildInfo, error) {
	buildDir, err := GetBuildDir(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	buildFiles, err := fileutils.ListFiles(buildDir, false)
	if err != nil {
		return nil, err
	}

	var generatedBuildsInfo []*buildinfo.BuildInfo
	for _, buildFile := range buildFiles {
		dir, err := fileutils.IsDir(false, buildFile)
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
		dir, err := fileutils.IsDir(false, buildFile)
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

func RemoveBuildDir(buildName, buildNumber string) error {
	tempDirPath, err := GetBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	exists, err := fileutils.IsDirExists(false, tempDirPath)
	if err != nil {
		return err
	}
	if exists {
		return errorutils.CheckError(os.RemoveAll(tempDirPath))
	}
	return nil
}

type BuildInfoConfiguration struct {
	artDetails auth.ArtifactoryDetails
	DryRun     bool
	EnvInclude string
	EnvExclude string
}

func (config *BuildInfoConfiguration) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return config.artDetails
}

func (config *BuildInfoConfiguration) SetArtifactoryDetails(art auth.ArtifactoryDetails) {
	config.artDetails = art
}

func (config *BuildInfoConfiguration) IsDryRun() bool {
	return config.DryRun
}

type BuildConfiguration struct {
	BuildName   string
	BuildNumber string
}

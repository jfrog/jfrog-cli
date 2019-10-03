package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jfrog/gocmd/cmd"
	"github.com/jfrog/gocmd/executers"
	executersutils "github.com/jfrog/gocmd/executers/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	_go "github.com/jfrog/jfrog-client-go/artifactory/services/go"
	cliutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils/checksum"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/version"
)

// Represent go project
type Go interface {
	Dependencies() []executers.Package
	CreateBuildInfoDependencies(includeInfoFiles bool) error
	PublishPackage(targetRepo, buildName, buildNumber string, servicesManager *artifactory.ArtifactoryServicesManager) error
	PublishDependencies(targetRepo string, servicesManager *artifactory.ArtifactoryServicesManager, includeDepSlice []string) (succeeded, failed int, err error)
	BuildInfo(includeArtifacts bool, module string) *buildinfo.BuildInfo
	LoadDependencies() error
}

type goProject struct {
	dependencies []executers.Package
	artifacts    []buildinfo.Artifact
	modContent   []byte
	moduleName   string
	version      string
	projectPath  string
}

// Load go project.
func Load(version string) (Go, error) {
	goProject := &goProject{version: version}
	err := goProject.readModFile()
	return goProject, err
}

// Get the go project dependencies.
func (project *goProject) Dependencies() []executers.Package {
	return project.dependencies
}

// Get the go project dependencies.
func (project *goProject) CreateBuildInfoDependencies(includeInfoFiles bool) error {
	for i, dep := range project.dependencies {
		err := dep.CreateBuildInfoDependencies(includeInfoFiles)
		if err != nil {
			return err
		}
		project.dependencies[i] = dep
	}
	return nil
}

// Get the project dependencies.
func (project *goProject) LoadDependencies() error {
	var err error
	project.dependencies, err = project.loadDependencies()
	return err
}

func (project *goProject) loadDependencies() ([]executers.Package, error) {
	cachePath, err := executersutils.GetCachePath()
	if err != nil {
		return nil, err
	}
	modulesMap, err := cmd.GetDependenciesGraph()
	if err != nil {
		return nil, err
	}
	if modulesMap == nil {
		return nil, nil
	}
	return executers.GetDependencies(cachePath, modulesMap)
}

// Publish go project to Artifactory.
func (project *goProject) PublishPackage(targetRepo, buildName, buildNumber string, servicesManager *artifactory.ArtifactoryServicesManager) error {
	log.Info("Publishing", project.getId(), "to", targetRepo)

	props, err := utils.CreateBuildProperties(buildName, buildNumber)
	if err != nil {
		return err
	}

	// Temp directory for the project archive.
	// The directory will be deleted at the end.
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		return err
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	params := _go.NewGoParams()
	params.Version = project.version
	params.Props = props
	params.TargetRepo = targetRepo
	params.ModuleId = project.getId()
	params.ModContent = project.modContent
	params.ModPath = filepath.Join(project.projectPath, "go.mod")
	params.ZipPath, err = project.archiveProject(project.version, tempDirPath)
	if err != nil {
		return err
	}
	// Create the info file if Artifactory version is 6.10.0 and above.
	artifactoryVersion, err := servicesManager.GetConfig().GetArtDetails().GetVersion()
	if err != nil {
		return err
	}
	version := version.NewVersion(artifactoryVersion)
	if version.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile) {
		pathToInfo, err := project.createInfoFile()
		if err != nil {
			return err
		}
		defer os.Remove(pathToInfo)
		if len(buildName) > 0 && len(buildNumber) > 0 {
			err = project.addInfoFileToBuildInfo(pathToInfo)
			if err != nil {
				return err
			}
		}
		params.InfoPath = pathToInfo
	}

	return servicesManager.PublishGoProject(params)
}

// Creates the info file.
// Returns the path to that file.
func (project *goProject) createInfoFile() (string, error) {
	log.Debug("Creating info file", project.projectPath)
	currentTime := time.Now().Format("2006-01-02T15:04:05Z")
	goInfoContent := goInfo{Version: project.version, Time: currentTime}
	content, err := json.Marshal(&goInfoContent)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	file, err := os.Create(project.version + ".info")
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	path, err := filepath.Abs(file.Name())
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	log.Debug("Info file was successfully created:", path)
	return path, nil
}

func (project *goProject) PublishDependencies(targetRepo string, servicesManager *artifactory.ArtifactoryServicesManager, includeDepSlice []string) (succeeded, failed int, err error) {
	log.Info("Publishing package dependencies...")
	includeDep := cliutils.ConvertSliceToMap(includeDepSlice)

	skip := 0
	_, includeAll := includeDep["ALL"]
	dependencies := project.Dependencies()
	for _, dependency := range dependencies {
		includeDependency := includeAll
		if !includeDependency {
			if _, included := includeDep[dependency.GetId()]; included {
				includeDependency = true
			}
		}
		if includeDependency {
			err = dependency.Publish("", targetRepo, servicesManager)
			if err != nil {
				err = errors.New("Failed to publish " + dependency.GetId() + " due to: " + err.Error())
				log.Error("Failed to publish", dependency.GetId(), ":", err)
			} else {
				succeeded++
			}
			continue
		}
		skip++
	}

	failed = len(dependencies) - succeeded - skip
	if failed > 0 {
		err = errors.New("Publishing project dependencies finished with errors. Please review the logs.")
	}
	return succeeded, failed, err
}

// Get the build info of the go project
func (project *goProject) BuildInfo(includeArtifacts bool, module string) *buildinfo.BuildInfo {
	buildInfoDependencies := []buildinfo.Dependency{}
	for _, dep := range project.dependencies {
		buildInfoDependencies = append(buildInfoDependencies, dep.Dependencies()...)
	}
	var artifacts []buildinfo.Artifact
	if includeArtifacts {
		artifacts = project.artifacts
	}
	buildInfoModule := buildinfo.Module{Id: module, Artifacts: artifacts, Dependencies: buildInfoDependencies}
	if module == "" {
		buildInfoModule.Id = project.getId()
	}
	return &buildinfo.BuildInfo{Modules: []buildinfo.Module{buildInfoModule}}
}

// Get go project ID in the form of projectName:version
func (project *goProject) getId() string {
	return project.moduleName
}

// Read go.mod file and add it as an artifact to the build info.
func (project *goProject) readModFile() error {
	var err error
	project.projectPath, err = cmd.GetProjectRoot()
	if err != nil {
		return errorutils.CheckError(err)
	}

	modFilePath := filepath.Join(project.projectPath, "go.mod")
	modFile, err := os.Open(modFilePath)
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer modFile.Close()
	content, err := ioutil.ReadAll(modFile)
	if err != nil {
		return errorutils.CheckError(err)
	}

	// Read module name
	project.moduleName, err = parseModuleName(string(content))
	if err != nil {
		return err
	}

	checksums, err := checksum.Calc(bytes.NewBuffer(content))
	if err != nil {
		return err
	}
	project.modContent = content

	// Add mod file as artifact
	artifact := buildinfo.Artifact{Name: project.version + ".mod"}
	artifact.Checksum = &buildinfo.Checksum{Sha1: checksums[checksum.SHA1], Md5: checksums[checksum.MD5]}
	project.artifacts = append(project.artifacts, artifact)
	return nil
}

// Archive the go project.
// Returns the path of the temp archived project file.
func (project *goProject) archiveProject(version, tempDir string) (string, error) {
	tempFile, err := ioutil.TempFile(tempDir, "project.zip")

	if err != nil {
		return "", errorutils.CheckError(err)
	}
	defer tempFile.Close()
	regex, err := getPathExclusionRegExp()
	if err != nil {
		return "", err
	}
	err = archiveProject(tempFile, project.projectPath, project.moduleName, version, regex)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	fileDetails, err := fileutils.GetFileDetails(tempFile.Name())
	if err != nil {
		return "", err
	}

	artifact := buildinfo.Artifact{Name: version + ".zip"}
	artifact.Checksum = &buildinfo.Checksum{Sha1: fileDetails.Checksum.Sha1, Md5: fileDetails.Checksum.Md5}
	project.artifacts = append(project.artifacts, artifact)
	return tempFile.Name(), nil
}

// Add the info file also as an artifact to be part of the build info.
func (project *goProject) addInfoFileToBuildInfo(infoFilePath string) error {
	fileDetails, err := fileutils.GetFileDetails(infoFilePath)
	if err != nil {
		return err
	}

	artifact := buildinfo.Artifact{Name: project.version + ".info"}
	artifact.Checksum = &buildinfo.Checksum{Sha1: fileDetails.Checksum.Sha1, Md5: fileDetails.Checksum.Md5}
	project.artifacts = append(project.artifacts, artifact)
	return nil
}

// Parse module name from go.mod content.
func parseModuleName(modContent string) (string, error) {
	r, err := regexp.Compile(`module "?([\w\.@:%_\+-.~#?&]+/?.+\w)`)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	lines := strings.Split(modContent, "\n")
	for _, v := range lines {
		matches := r.FindStringSubmatch(v)
		if len(matches) == 2 {
			return matches[1], nil
		}
	}

	return "", errorutils.CheckError(errors.New("Module name missing in go.mod file"))
}

// Returns a regex that matches the following:
// 1. .DS_Store.
// 2. .git.
func getPathExclusionRegExp() (*regexp.Regexp, error) {
	excludePathsRegExp, err := regexp.Compile("(" + filepath.Join("^*", ".git", ".*$") + ")|(" + filepath.Join("^*", ".DS_Store") + ")")
	if err != nil {
		return nil, err
	}

	return excludePathsRegExp, nil
}

type goInfo struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

package project

import (
	"bytes"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang/project/dependencies"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services/go"
	cliutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils/checksum"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Represent go project
type Go interface {
	Dependencies() []dependencies.Package
	PublishPackage(targetRepo, buildName, buildNumber string, servicesManager *artifactory.ArtifactoryServicesManager) error
	PublishDependencies(targetRepo string, servicesManager *artifactory.ArtifactoryServicesManager, includeDepSlice []string) (succeeded, failed int, err error)
	BuildInfo(includeArtifacts bool) *buildinfo.BuildInfo
	LoadDependencies() error
	DownloadFromVcsAndPublish(targetRepo, goArg string, recursiveTidy, recursiveTidyOverwrite bool, details *config.ArtifactoryDetails) error
}

type goProject struct {
	dependencies []dependencies.Package
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
	if err != nil {
		return nil, err
	}

	err = os.Chdir(goProject.projectPath)
	if err != nil {
		return nil, err
	}
	return goProject, err
}

// Get the go project dependencies.
func (project *goProject) Dependencies() []dependencies.Package {
	return project.dependencies
}

// Get the project dependencies.
func (project *goProject) LoadDependencies() error {
	var err error
	project.dependencies, err = project.loadDependencies()
	return err
}

// Downloads all dependencies from VCS and publish them to Artifactory.
func (project *goProject) DownloadFromVcsAndPublish(targetRepo, goArg string, recursiveTidy, recursiveTidyOverwrite bool, details *config.ArtifactoryDetails) error {
	wd, err := os.Getwd()
	if err != nil {
		return errorutils.CheckError(err)
	}
	rootProjectDir, err := golang.GetProjectRoot()
	if err != nil {
		return err
	}

	// Need to run Go without Artifactory to resolve all dependencies.
	cache := golang.DependenciesCache{}
	err = collectDependenciesPopulateAndPublish(targetRepo, recursiveTidy, recursiveTidyOverwrite, &cache, details)
	if err != nil {
		if !recursiveTidy {
			return err
		}
		log.Error("Received an error:", err)
	}
	// Lets run the same command again now that all the dependencies were downloaded.
	// Need to run only if the command is not go mod download and go mod tidy since this was run by the CLI to download and publish to Artifactory
	log.Info(fmt.Sprintf("Done building and publishing %d go dependencies to Artifactory out of a total of %d dependencies.", cache.GetSuccesses(), cache.GetTotal()))
	if !strings.Contains(goArg, "mod download") && !strings.Contains(goArg, "mod tidy") {
		if recursiveTidy {
			// Remove the go.sum file, since it includes information which is not up to date (it was created by the "go mod tidy" command executed without Artifactory
			err = removeGoSumFile(wd, rootProjectDir)
			if err != nil {
				log.Error("Received an error:", err)
			}
		}
		err = golang.RunGo(goArg)
	}
	return err
}

// Download the dependencies from VCS and publish them to Artifactory.
func collectDependenciesPopulateAndPublish(targetRepo string, recursiveTidy, recursiveTidyOverwrite bool, cache *golang.DependenciesCache, details *config.ArtifactoryDetails) error {
	err := os.Unsetenv(golang.GOPROXY)
	if err != nil {
		return err
	}
	dependenciesToPublish, err := dependencies.CollectProjectDependencies(targetRepo, cache, details)
	if err != nil || len(dependenciesToPublish) == 0 {
		return err
	}

	var dependency dependencies.GoPackage
	if recursiveTidy {
		err = fileutils.CreateTempDirPath()
		if err != nil {
			return err
		}
		defer fileutils.RemoveTempDir()

		dependency = &dependencies.PackageWithDeps{}
		err = dependency.Init()
		if err != nil {
			return err
		}
	} else {
		dependency = &dependencies.Package{}
	}

	return runPopulateAndPublishDependencies(targetRepo, recursiveTidy, recursiveTidyOverwrite, dependency, dependenciesToPublish, cache, details)
}

func runPopulateAndPublishDependencies(targetRepo string, recursiveTidy, recursiveTidyOverwrite bool, dependenciesInterface dependencies.GoPackage, dependenciesToPublish map[string]bool, cache *golang.DependenciesCache, details *config.ArtifactoryDetails) error {
	cachePath, err := dependencies.GetCachePath()
	if err != nil {
		return err
	}

	dependencies, err := dependencies.GetDependencies(cachePath, dependenciesToPublish)
	if err != nil {
		return err
	}

	cache.IncrementTotal(len(dependencies))
	for _, dep := range dependencies {
		dependenciesInterface = dependenciesInterface.New(cachePath, dep, recursiveTidyOverwrite)
		err := dependenciesInterface.PopulateModAndPublish(targetRepo, cache, details)
		if err != nil {
			if recursiveTidy {
				log.Warn(err)
				continue
			}
			return err
		}
	}
	return nil
}

func removeGoSumFile(wd, rootDir string) error {
	log.Debug("Changing back to the working directory")
	err := os.Chdir(wd)
	if err != nil {
		return errorutils.CheckError(err)
	}

	goSumFile := filepath.Join(rootDir, "go.sum")
	exists, err := fileutils.IsFileExists(goSumFile, false)
	if err != nil {
		return err
	}
	if exists {
		return errorutils.CheckError(os.Remove(goSumFile))
	}
	return nil
}

func (project *goProject) loadDependencies() ([]dependencies.Package, error) {
	cachePath, err := dependencies.GetCachePath()
	if err != nil {
		return nil, err
	}
	modulesMap, err := golang.GetDependenciesGraph()
	if err != nil {
		return nil, err
	}
	if modulesMap == nil {
		return nil, nil
	}
	return dependencies.GetDependencies(cachePath, modulesMap)
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
	err = fileutils.CreateTempDirPath()
	if err != nil {
		return err
	}
	defer fileutils.RemoveTempDir()

	params := _go.NewGoParams()
	params.Version = project.version
	params.Props = props
	params.TargetRepo = targetRepo
	params.ModuleId = project.getId()
	params.ModContent = project.modContent
	params.ModPath = filepath.Join(project.projectPath, "go.mod")
	params.ZipPath, err = project.archiveProject(project.version)
	if err != nil {
		return err
	}

	return servicesManager.PublishGoProject(params)
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
func (project *goProject) BuildInfo(includeArtifacts bool) *buildinfo.BuildInfo {
	buildInfoDependencies := []buildinfo.Dependency{}
	for _, dep := range project.dependencies {
		buildInfoDependencies = append(buildInfoDependencies, dep.Dependencies()...)
	}
	var artifacts []buildinfo.Artifact
	if includeArtifacts {
		artifacts = project.artifacts
	}
	return &buildinfo.BuildInfo{Modules: []buildinfo.Module{{Id: project.getId(), Artifacts: artifacts, Dependencies: buildInfoDependencies}}}
}

// Get go project ID in the form of projectName:version
func (project *goProject) getId() string {
	return project.moduleName
}

// Read go.mod file and add it as an artifact to the xbuild info
func (project *goProject) readModFile() error {
	var err error
	project.projectPath, err = golang.GetProjectRoot()
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
func (project *goProject) archiveProject(version string) (string, error) {
	tempDir, err := fileutils.GetTempDirPath()
	if err != nil {
		return "", err
	}
	tempFile, err := ioutil.TempFile(tempDir, "project.zip")
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	regex, err := getPathExclusionRegExp()
	if err != nil {
		tempFile.Close()
		return "", err
	}
	err = archiveProject(tempFile, project.projectPath, project.moduleName, version, regex)
	if err != nil {
		tempFile.Close()
		return "", errorutils.CheckError(err)
	}
	tempFile.Close()

	fileDetails, err := fileutils.GetFileDetails(tempFile.Name())
	if err != nil {
		return "", err
	}

	artifact := buildinfo.Artifact{Name: version + ".zip"}
	artifact.Checksum = &buildinfo.Checksum{Sha1: fileDetails.Checksum.Sha1, Md5: fileDetails.Checksum.Md5}
	project.artifacts = append(project.artifacts, artifact)
	return tempFile.Name(), nil
}

// Parse module name from go.mod content.
func parseModuleName(modContent string) (string, error) {
	r, err := regexp.Compile(`module ([\w\.@:%_\+-.~#?&]+/?.+)`)
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

// Returns a regex that match the following:
// 1. .git folder.
// 2. .gitignore file
// 3. .DS_Store
func getPathExclusionRegExp() (*regexp.Regexp, error) {
	excludePathsRegExp, err := regexp.Compile("(" + filepath.Join("^*", ".git", ".*$") + ")|(" + filepath.Join("^*", ".gitignore") + ")|(" + filepath.Join("^*", ".DS_Store") + ")")
	if err != nil {
		return nil, err
	}

	return excludePathsRegExp, nil
}

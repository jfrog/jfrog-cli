package project

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/vgo/project/dependencies"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/vgo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils/checksum"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Represent go project
type Go interface {
	Dependencies() []dependencies.Dependency
	Publish(targetRepo, buildName, buildNumber string, details *config.ArtifactoryDetails) error
	BuildInfo() *buildinfo.BuildInfo
}

type goProject struct {
	dependencies []dependencies.Dependency
	artifacts    []buildinfo.Artifact
	modContent   []byte
	moduleName   string
	version      string
	projectPath  string
}

// Load vgo project.
func Load(version string) (Go, error) {
	goProject := &goProject{version: version}
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = goProject.readModFile(pwd)
	if err != nil {
		return nil, err
	}
	goProject.projectPath = pwd
	goProject.dependencies, err = dependencies.Load()
	return goProject, err
}

// Get the vgo project dependencies.
func (project *goProject) Dependencies() []dependencies.Dependency {
	return project.dependencies
}

// Publish vgo project to Artifactory.
func (project *goProject) Publish(targetRepo, buildName, buildNumber string, details *config.ArtifactoryDetails) error {
	log.Info("Publishing", project.getId(), "to", targetRepo)
	servicesManager, err := utils.CreateServiceManager(details, false)
	if err != nil {
		return err
	}

	props, err := utils.CreateBuildProperties(buildName, buildNumber)
	if err != nil {
		return err
	}

	// Temp folder for the project archive.
	// The folder will be deleted at the end.
	err = fileutils.CreateTempDirPath()
	if err != nil {
		return err
	}
	defer fileutils.RemoveTempDir()

	params := &vgo.VgoParamsImpl{}
	params.Version = project.version
	params.Props = props
	params.TargetRepo = targetRepo

	params.ModContent = project.modContent
	params.ZipPath, err = project.archiveProject(project.version)
	if err != nil {
		return err
	}

	return servicesManager.PublishVgoProject(params)
}

// Get the build info of the vgo project
func (project *goProject) BuildInfo() *buildinfo.BuildInfo {
	buildInfoDependencies := []buildinfo.Dependency{}
	for _, dep := range project.dependencies {
		buildInfoDependencies = append(buildInfoDependencies, dep.Dependencies()...)
	}
	return &buildinfo.BuildInfo{Modules: []buildinfo.Module{{Id: project.getId(), Artifacts: project.artifacts, Dependencies: buildInfoDependencies}}}
}

// Get vgo project ID in the form of projectName:version
func (project *goProject) getId() string {
	return fmt.Sprintf("%s:%s", project.moduleName, project.version)
}

// Read go.mod file and add it as a dependency to the build info
func (project *goProject) readModFile(projectPath string) error {
	modFile, err := os.Open(filepath.Join(projectPath, "go.mod"))
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

// Archive the vgo project.
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

	err = archiveProject(tempFile, project.projectPath, project.moduleName, version)
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
	r, err := regexp.Compile(`module "([\w\.@:%_\+-.~#?&]+/.+)"`)
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

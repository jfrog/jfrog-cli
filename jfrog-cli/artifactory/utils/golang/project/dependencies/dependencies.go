package dependencies

import (
	"bytes"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	golangutil "github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services/go"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils/checksum"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

func Load() ([]Dependency, error) {
	goPath, err := getGOPATH()
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	cachePath := filepath.Join(goPath, "pkg", "mod", "cache", "download")
	return getDependencies(cachePath)
}

// Represent go dependency project.
// Includes publishing capabilities and build info dependencies.
type Dependency struct {
	buildInfoDependencies []buildinfo.Dependency
	id                    string
	modContent            []byte
	zipPath               string
	version               string
}

func (dependency *Dependency) GetId() string {
	return dependency.id
}

func (dependency *Dependency) Publish(targetRepo string, servicesManager *artifactory.ArtifactoryServicesManager) error {
	log.Info("Publishing:", dependency.id, "to", targetRepo)
	params := &_go.GoParamsImpl{}
	params.ZipPath = dependency.zipPath
	params.ModContent = dependency.modContent
	params.Version = dependency.version
	params.TargetRepo = targetRepo
	params.ModuleId = dependency.id

	return servicesManager.PublishGoProject(params)
}

func (dependency *Dependency) Dependencies() []buildinfo.Dependency {
	return dependency.buildInfoDependencies
}

func getDependencies(cachePath string) ([]Dependency, error) {
	goCmd, err := golangutil.NewCmd()
	if err != nil {
		return nil, err
	}
	goCmd.Command = []string{"list"}
	goCmd.CommandFlags = []string{"-m", "all"}
	output, err := utils.RunCmdOutput(goCmd)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	nameVersionMap, err := parseListOutput(output)
	if err != nil {
		return nil, nil
	}

	deps := []Dependency{}
	for name, ver := range nameVersionMap {
		name := getDependencyName(name)
		dep, err := createDependency(cachePath, name, ver)
		if err != nil {
			return nil, err
		}
		if dep != nil {
			deps = append(deps, *dep)
		}
	}
	return deps, nil
}

// Returns the actual path to the dependency.
// If in the path there are capital letters, the Go convention is to use "!" before the letter.
// The letter itself in lowercase.
func getDependencyName(name string) string {
	path := ""
	for _, letter := range name {
		if unicode.IsUpper(letter) {
			path += "!" + strings.ToLower(string(letter))
		} else {
			path += string(letter)
		}
	}
	return path
}

// Creates a go dependency.
// Returns a nil value in case the dependency does not include a zip in the cache.
func createDependency(cachePath, dependencyName, version string) (*Dependency, error) {
	// We first check if the this dependency has a zip binary in the local go cache.
	// If it does not, nil is returned. This seems to be a bug in go.
	zipPath, err := getPackageZipLocation(cachePath, dependencyName, version)

	if err != nil {
		return nil, err
	}

	if zipPath == "" {
		return nil, nil
	}

	dep := Dependency{}

	dep.id = strings.Join([]string{dependencyName, version}, ":")
	dep.version = version
	dep.zipPath = zipPath
	dep.modContent, err = ioutil.ReadFile(filepath.Join(cachePath, dependencyName, "@v", version+".mod"))
	if err != nil {
		return &dep, errorutils.CheckError(err)
	}

	// Mod file dependency
	modDependency := buildinfo.Dependency{Id: dep.id}
	checksums, err := checksum.Calc(bytes.NewBuffer(dep.modContent))
	if err != nil {
		return &dep, err
	}
	modDependency.Checksum = &buildinfo.Checksum{Sha1: checksums[checksum.SHA1], Md5: checksums[checksum.MD5]}

	// Zip file dependency
	zipDependency := buildinfo.Dependency{Id: dep.id}
	fileDetails, err := fileutils.GetFileDetails(dep.zipPath)
	if err != nil {
		return &dep, err
	}
	zipDependency.Checksum = &buildinfo.Checksum{Sha1: fileDetails.Checksum.Sha1, Md5: fileDetails.Checksum.Md5}

	dep.buildInfoDependencies = append(dep.buildInfoDependencies, modDependency, zipDependency)
	return &dep, nil
}

// Returns the path to the package zip file if exists.
func getPackageZipLocation(cachePath, dependencyName, version string) (string, error) {
	zipPath, err := getPackagePathIfExists(cachePath, dependencyName, version)
	if err != nil {
		return "", err
	}

	if zipPath != "" {
		return zipPath, nil
	}

	zipPath, err = getPackagePathIfExists(filepath.Dir(cachePath), dependencyName, version)

	if err != nil {
		return "", err
	}

	return zipPath, nil
}

// Validates if the package zip file exists.
func getPackagePathIfExists(cachePath, dependencyName, version string) (zipPath string, err error) {
	zipPath = filepath.Join(cachePath, dependencyName, "@v", version+".zip")
	fileExists, err := fileutils.IsFileExists(zipPath, false)
	if err != nil {
		log.Warn(fmt.Sprintf("Could not find zip binary for dependency '%s' at %s.", dependencyName, zipPath))
		return "", err
	}
	// Zip binary does not exist, so we skip it by returning a nil dependency.
	if !fileExists {
		log.Debug("The following file is missing:", zipPath)
		return "", nil
	}
	return zipPath, nil
}

func parseListOutput(content []byte) (map[string]string, error) {
	depRegexp, err := regexp.Compile("(\\S+)\\s+(\\S+)")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	depMap := map[string]string{}
	lines := bytes.Split(content, []byte("\n"))
	for i := 0; i < len(lines); i++ {
		dependency := depRegexp.FindStringSubmatch(string(lines[i]))
		if len(dependency) == 3 {
			depMap[dependency[1]] = dependency[2]
		}
	}
	return depMap, nil
}

func getGOPATH() (string, error) {
	goCmd, err := golangutil.NewCmd()
	if err != nil {
		return "", err
	}
	goCmd.Command = []string{"env", "GOPATH"}
	output, err := utils.RunCmdOutput(goCmd)
	if err != nil {
		return "", fmt.Errorf("Could not find GOPATH env: %s", err.Error())
	}
	return strings.TrimSpace(string(output)), nil
}

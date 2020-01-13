package dependencies

import (
	"encoding/json"
	"errors"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	serviceutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// The extractor responsible to calculate the project dependencies.
type Extractor interface {
	// Get all the dependencies for the project.
	AllDependencies() map[string]*buildinfo.Dependency
	// Get all the root dependencies of the project.
	DirectDependencies() []string
	// Dependencies relations map.
	ChildrenMap() map[string][]string
	// Decide package name.
	PackageName() (string, error)

	Extract() error
}

func CreateCompatibleExtractor(pythonExecutablePath string, pipArgs []string) (Extractor, error) {
	// Check if using requirements file.
	filePath, err := getRequirementsFilePath(pipArgs)
	if err != nil {
		return nil, err
	}
	if filePath != "" {
		return NewRequirementsExtractor(filePath, pythonExecutablePath), nil
	}

	// Setup.py should be in current dir.
	filePath, err = getSetuppyFilePath()
	if err != nil {
		return nil, err
	}
	if filePath != "" {
		return NewSetupExtractor(filePath, pythonExecutablePath), nil
	}

	// Couldn't resolve requirements file or setup.py.
	return nil, errorutils.CheckError(errors.New("Could not find installation file for pip command, the command must include '--requirement' or be executed from within the directory containing the 'setup.py' file."))
}

// Look for 'requirements' flag in command args.
// If found, validate the file exists and return its path.
func getRequirementsFilePath(args []string) (string, error) {
	// Get requirements file path from args.
	_, _, requirementsFilePath, err := utils.FindFlagFirstMatch([]string{"-r", "--requirement"}, args)
	if err != nil || requirementsFilePath == "" {
		// Args don't include a path to requirements file.
		return "", err
	}

	// Validate path exists.
	validPath, err := fileutils.IsFileExists(requirementsFilePath, false)
	if err != nil {
		return "", err
	}
	if !validPath {
		return "", errorutils.CheckError(errors.New(fmt.Sprintf("Could not find requirements file at provided location: %s", requirementsFilePath)))
	}

	// Return absolute path.
	return filepath.Abs(requirementsFilePath)
}

// Look for 'setup.py' file in current work dir.
// If found, return its absolute path.
func getSetuppyFilePath() (string, error) {
	wd, err := os.Getwd()
	if errorutils.CheckError(err) != nil {
		return "", err
	}

	filePath := filepath.Join(wd, "setup.py")
	// Check if setup.py exists.
	validPath, err := fileutils.IsFileExists(filePath, false)
	if err != nil {
		return "", err
	}
	if !validPath {
		return "", errorutils.CheckError(errors.New(fmt.Sprintf("Could not find setup.py file in current directory: %s", wd)))
	}

	return filePath, nil
}

// Execute pip-dependency-map script, return dependency map of all installed pip packages in current environment.
// pythonExecPath - Execution path python.
func BuildPipDependencyMap(pythonExecPath string) (map[string]pipDependencyPackage, error) {
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	pipDependencyMapScriptPath, err := GetDepTreeScriptPath()
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	// Execute the python pip-dependency-map script.
	pipDependencyMapCmd := &pip.PipCmd{
		Executable:  pythonExecPath,
		Command:     pipDependencyMapScriptPath,
		CommandArgs: []string{"--json"},
		StrWriter:   pipeWriter,
	}
	var pythonErr error
	go func() {
		pythonErr = gofrogcmd.RunCmd(pipDependencyMapCmd)
	}()
	data, err := ioutil.ReadAll(pipeReader)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	if pythonErr != nil {
		return nil, errorutils.CheckError(pythonErr)
	}

	// Parse the result.
	return parsePipDependencyMapOutput(data)
}

// Parse pip-dependency-map raw output to dependencies map.
func parsePipDependencyMapOutput(data []byte) (map[string]pipDependencyPackage, error) {
	// Parse into array.
	packages := make([]pipDependencyPackage, 0)
	if err := json.Unmarshal(data, &packages); err != nil {
		return nil, errorutils.CheckError(err)
	}

	// Create packages map.
	packagesMap := make(map[string]pipDependencyPackage)
	for _, pkg := range packages {
		packagesMap[pkg.Package.Key] = pkg
	}

	return packagesMap, nil
}

// Extract all dependencies, based on provided root-dependencies 'dependencies'.
// Return allDependencies and childrenMap.
func extractDependencies(dependencies []string, environmentPackages map[string]pipDependencyPackage) (allDependencies map[string]*buildinfo.Dependency, childrenMap map[string][]string, err error) {
	allDependencies = make(map[string]*buildinfo.Dependency)
	childrenMap = make(map[string][]string)
	// Iterate over dependencies, resolve and discover more dependencies.
	index := -1
	var currentDep string
	for {
		index++

		// Check if should stop.
		if len(dependencies) < index+1 {
			break
		}

		currentDep = dependencies[index]
		// Check if current dependency already resolved.
		if _, ok := allDependencies[currentDep]; ok {
			// Already resolved.
			continue
		}

		// Resolve dependency.
		depTreePkg, ok := environmentPackages[currentDep]
		if !ok {
			// Current dependency was not found in pip-deps-tree script output.
			// This may happen for packages which are being installed from the requirements.txt or setup.py files, but also
			// integrated within the Python standard library. Thus won't be included in pip-freeze output.
			log.Debug(fmt.Sprintf("Package name: %s appears in installation file, but not shown in environment's installed dependencies.", currentDep))
			allDependencies[currentDep] = &buildinfo.Dependency{}
			continue
		}

		// Extract pip-dependency from pip-package.
		var pipDep *pipDependency
		pipDep, err = depTreePkg.extractPipDependency()
		if err != nil {
			return
		}

		// Update extractor's map.
		if pipDep != nil {
			allDependencies[pipDep.id] = pipDep.dependency
			childrenMap[pipDep.id] = pipDep.dependencies
		}

		// Add pipDep dependency for resolution.
		dependencies = append(dependencies, pipDep.dependencies...)
	}
	return
}

type pipDependency struct {
	id           string
	version      string
	dependency   *buildinfo.Dependency
	dependencies []string
}

func (pipDepTreePkg *pipDependencyPackage) extractPipDependency() (*pipDependency, error) {
	// Create pip-dependency.
	pipDependency := &pipDependency{id: pipDepTreePkg.Package.Key, version: pipDepTreePkg.Package.InstalledVersion, dependencies: pipDepTreePkg.getDependencies()}

	// Build build-info dependency.
	pipDependency.dependency = &buildinfo.Dependency{Id: pipDepTreePkg.Package.PackageName + ":" + pipDepTreePkg.Package.InstalledVersion}

	return pipDependency, nil
}

func (pipDepTreePkg *pipDependencyPackage) getDependencies() []string {
	var dependencies []string
	for _, dep := range pipDepTreePkg.Dependencies {
		dependencies = append(dependencies, strings.ToLower(dep.Key))
	}
	return dependencies
}

// Return path to the dependency-tree script, if not exists it creates the file.
func GetDepTreeScriptPath() (string, error) {
	pipDependenciesPath, err := config.GetJfrogDependenciesPath()
	if err != nil {
		return "", err
	}
	pipDependenciesPath = filepath.Join(pipDependenciesPath, "pip", pipDepTreeVersion)
	depTreeScriptName := "pipdeptree.py"
	depTreeScriptPath := path.Join(pipDependenciesPath, depTreeScriptName)
	err = writeScriptIfNeeded(pipDependenciesPath, depTreeScriptName)

	return depTreeScriptPath, err
}

func writeScriptIfNeeded(targetDirPath, scriptName string) error {
	scriptPath := path.Join(targetDirPath, scriptName)
	exists, err := fileutils.IsFileExists(scriptPath, false)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if !exists {
		err = os.MkdirAll(targetDirPath, os.ModeDir|os.ModePerm)
		if errorutils.CheckError(err) != nil {
			return err
		}
		err = ioutil.WriteFile(scriptPath, pipDepTreeContent, os.ModePerm)
		if errorutils.CheckError(err) != nil {
			return err
		}
	}
	return nil
}

// Populate project's dependencies with checksums and file names.
// If the dependency was downloaded in this pip-install execution, checksum will be fetched from Artifactory.
// Otherwise, check if exists in cache.
// Return dependency-names of all dependencies which its information could not be obtained.
func AddDepsInfoAndReturnMissingDeps(dependenciesMap map[string]*buildinfo.Dependency, dependenciesCache *DependenciesCache, dependencyToFileMap map[string]string, servicesManager *artifactory.ArtifactoryServicesManager, repository string) ([]string, error) {
	var missingDeps []string
	// Iterate dependencies map to update info.
	for depName := range dependenciesMap {
		// Get dependency info.
		depFileName, depChecksum, err := getDependencyInfo(depName, repository, dependenciesCache, dependencyToFileMap, servicesManager)
		if err != nil {
			return nil, err
		}

		// Check if info not found.
		if depFileName == "" || depChecksum == nil {
			// Dependency either wasn't downloaded in this run nor stored in cache.
			missingDeps = append(missingDeps, depName)

			// dependenciesMapT should contain only dependencies with checksums.
			delete(dependenciesMap, depName)

			continue
		}
		fileType := ""
		// Update dependency info.
		dependenciesMap[depName].Id = depFileName
		if i := strings.LastIndex(depFileName, "."); i != -1 {
			fileType = depFileName[i+1:]
		}
		dependenciesMap[depName].Type = fileType
		dependenciesMap[depName].Checksum = depChecksum
	}

	return missingDeps, nil
}

// Get dependency information.
// If dependency was downloaded in this pip-install execution, fetch info from Artifactory.
// Otherwise, fetch info from cache.
func getDependencyInfo(depName, repository string, dependenciesCache *DependenciesCache, dependencyToFileMap map[string]string, servicesManager *artifactory.ArtifactoryServicesManager) (string, *buildinfo.Checksum, error) {
	// Check if this dependency was updated during this pip-install execution, and we have its file-name.
	// If updated - fetch checksum from Artifactory, regardless of what was previously stored in cache.
	depFileName, ok := dependencyToFileMap[depName]
	if ok && depFileName != "" {
		checksum, err := getDependencyChecksumFromArtifactory(servicesManager, repository, depFileName)
		return depFileName, checksum, err
	}

	// Check cache for dependency checksum.
	if dependenciesCache != nil {
		dep := dependenciesCache.GetDependency(depName)
		if dep != nil {
			// Checksum found in cache, return info
			return dep.Id, dep.Checksum, nil
		}
	}

	return "", nil, nil
}

// Fetch checksum for file from Artifactory.
// If the file isn't found, or md5 or sha1 are missing, return nil.
func getDependencyChecksumFromArtifactory(servicesManager *artifactory.ArtifactoryServicesManager, repository, dependencyFile string) (*buildinfo.Checksum, error) {
	log.Debug(fmt.Sprintf("Fetching checksums for: %s", dependencyFile))
	result, err := servicesManager.Aql(serviceutils.CreateAqlQueryForPypi(repository, dependencyFile))
	if err != nil {
		return nil, err
	}

	parsedResult := new(aqlResult)
	err = json.Unmarshal(result, parsedResult)
	if err = errorutils.CheckError(err); err != nil {
		return nil, err
	}
	if len(parsedResult.Results) == 0 {
		log.Debug(fmt.Sprintf("File: %s could not be found in repository: %s", dependencyFile, repository))
		return nil, nil
	}

	// Verify checksum exist.
	sha1 := parsedResult.Results[0].Actual_sha1
	md5 := parsedResult.Results[0].Actual_md5
	if sha1 == "" || md5 == "" {
		// Missing checksum.
		log.Debug(fmt.Sprintf("Missing checksums for file: %s, sha1: '%s', md5: '%s'", dependencyFile, sha1, md5))
		return nil, nil
	}

	// Update checksum.
	checksum := &buildinfo.Checksum{Sha1: sha1, Md5: md5}
	log.Debug(fmt.Sprintf("Found checksums for file: %s, sha1: '%s', md5: '%s'", dependencyFile, sha1, md5))

	return checksum, nil
}

type aqlResult struct {
	Results []*results `json:"results,omitempty"`
}

type results struct {
	Actual_md5  string `json:"actual_md5,omitempty"`
	Actual_sha1 string `json:"actual_sha1,omitempty"`
}

// Structs for parsing the pip-dependency-map result.

type pipDependencyPackage struct {
	Package      packageType  `json:"package,omitempty"`
	Dependencies []dependency `json:"dependencies,omitempty"`
}

type packageType struct {
	Key              string `json:"key,omitempty"`
	PackageName      string `json:"package_name,omitempty"`
	InstalledVersion string `json:"installed_version,omitempty"`
}

type dependency struct {
	Key              string `json:"key,omitempty"`
	PackageName      string `json:"package_name,omitempty"`
	InstalledVersion string `json:"installed_version,omitempty"`
}

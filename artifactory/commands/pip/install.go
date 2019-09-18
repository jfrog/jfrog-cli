package pip

import (
	"encoding/json"
	errors2 "errors"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	piputils "github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip/dependencies"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PipInstallCommand struct {
	rtDetails              *config.ArtifactoryDetails
	buildConfiguration     *utils.BuildConfiguration
	args                   []string
	repository             string
	shouldCollectBuildInfo bool
	dependencyToFileMap    map[string]string//Parsed from pip-install logs, maps dependency name to its actual downloaded file from Artifactory.
}

func NewPipInstallCommand() *PipInstallCommand {
	return &PipInstallCommand{}
}

func (pic *PipInstallCommand) Run() error {
	log.Info("Running pip Install.")

	// Prepare for running.
	pipExecutablePath, pythonExecutablePath, pipIndexUrl, err := pic.prepare()
	if err != nil {
		return nil
	}

	// Run pip install.
	err = pic.runPipInstall(pipExecutablePath, pipIndexUrl)
	if err != nil {
		pic.cleanBuildInfoDir()
		return err
	}

	// Check if need to collect build-info.
	if !pic.shouldCollectBuildInfo {
		log.Info("pip install finished successfully.")
		return nil
	}

	// Collect build-info.
	if err := pic.collectBuildInfo(pythonExecutablePath); err != nil {
		pic.cleanBuildInfoDir()
		return err
	}

	log.Info("pip install finished successfully.")
	return nil
}

func (pic *PipInstallCommand) runPipInstall(pipExecutablePath, pipIndexUrl string) error {
	// Create pip install command.
	pipInstallCmd := &piputils.PipCmd{
		Executable:  pipExecutablePath,
		Command:     "install",
		CommandArgs: append(pic.args, "-i", pipIndexUrl),
	}

	// If need to collect build-info, run pip-install with log parsing.
	if pic.shouldCollectBuildInfo {
		return pic.runPipInstallWithLogParsing(pipInstallCmd)
	}

	// Run without log parsing.
	return gofrogcmd.RunCmd(pipInstallCmd)
}

// Run pip-install command while parsing the logs for downloaded packages.
// Supports running pip either in non-verbose and verbose mode.
// Populates 'dependencyToFileMap' with downloaded package-name and its actual downloaded file (wheel/egg/zip...).
func (pic *PipInstallCommand) runPipInstallWithLogParsing(pipInstallCmd *piputils.PipCmd) error {
	// Create regular expressions for log parsing.
	collectingPackageRegexp, err := clientutils.GetRegExp(`^Collecting\s(\w[\w-\.]+).*`)
	if err != nil {
		return err
	}
	downloadFileRegexp, err := clientutils.GetRegExp(`^\s\sDownloading\s[^\s]*\/packages\/[^\s]*\/([^\s]*)`)
	if err != nil {
		return err
	}

	downloadedDependencies := make(map[string]string)
	var packageName string
	expectingPackageFilePath := false

	// Extract downloaded package name.
	dependencyNameParser := gofrogcmd.CmdOutputPattern{
		RegExp: collectingPackageRegexp,
		ExecFunc: func(pattern *gofrogcmd.CmdOutputPattern) (string, error) {
			// If this pattern matched a second time before downloaded-file-name was found, prompt a message.
			if expectingPackageFilePath {
				// This may occur when a package-installation file is saved in pip-cache-dir, thus not being downloaded during the installation.
				// Re-running pip-install with 'no-cache-dir' fixes this issue.
				log.Debug(fmt.Sprintf("Could not resolve download path for package: %s, continuing...", packageName))
			}

			// Check for out of bound results.
			if len(pattern.MatchedResults)-1 < 0 {
				log.Debug(fmt.Sprintf("Failed extracting package name from line: %s", pattern.Line))
				return pattern.Line, nil
			}

			// Save dependency information.
			expectingPackageFilePath = true
			packageName = pattern.MatchedResults[1]

			return pattern.Line, nil
		},
	}

	// Extract downloaded file, stored in Artifactory.
	dependencyFileParser := gofrogcmd.CmdOutputPattern{
		RegExp: downloadFileRegexp,
		ExecFunc: func(pattern *gofrogcmd.CmdOutputPattern) (string, error) {
			// Check for out of bound results.
			if len(pattern.MatchedResults)-1 < 0 {
				log.Debug(fmt.Sprintf("Failed extracting download path from line: %s", pattern.Line))
				return pattern.Line, nil
			}

			// If this pattern matched before package-name was found, do not collect this path.
			if !expectingPackageFilePath {
				log.Debug(fmt.Sprintf("Could not resolve package name for download path: %s , continuing...", packageName))
				return pattern.Line, nil
			}

			// Save dependency information.
			filePath := pattern.MatchedResults[1]
			downloadedDependencies[strings.ToLower(packageName)] = filePath
			expectingPackageFilePath = false

			log.Debug(fmt.Sprintf("Found package: %s installed with: %s", packageName, filePath))
			return pattern.Line, nil
		},
	}

	// Execute command.
	_, _, _, err = gofrogcmd.RunCmdWithOutputParser(pipInstallCmd, true, &dependencyNameParser, &dependencyFileParser)
	if errorutils.CheckError(err) != nil {
		return err
	}

	// Update dependencyToFileMap.
	pic.dependencyToFileMap = downloadedDependencies

	return nil
}

func (pic *PipInstallCommand) collectBuildInfo(pythonExecutablePath string) error {
	// Create compatible extractor for dependencies resolution, and extract dependencies.
	extractor, err := pic.createCompatibleExtractor(pythonExecutablePath)
	if err != nil {
		return err
	}

	err = extractor.Extract()
	if err != nil {
		return err
	}

	// Determine module name for build-info.
	if err := pic.determineModuleName(extractor); err != nil {
		return err
	}

	// Get project dependencies.
	allDependencies := extractor.AllDependencies()

	// Populate dependencies information - checksums and file-name.
	pic.addDepsInfo(allDependencies)

	// Update project cache with correct dependencies.
	dependencies.UpdateDependenciesCache(allDependencies)

	// Save build-info built from allDependencies.
	pic.saveBuildInfo(allDependencies)

	return nil
}

func (pic *PipInstallCommand) createCompatibleExtractor(pythonExecutablePath string) (dependencies.Extractor, error) {
	// Check if using requirements file.
	filePath, err := getRequirementsFilePath(pic.args)
	if err != nil {
		return nil, err
	}
	if filePath != "" {
		// Create requirements extractor.
		return dependencies.NewRequirementsExtractor(filePath, pythonExecutablePath), nil
	}

	// Setup.py should be in current dir.
	filePath, err = getSetuppyFilePath()
	if err != nil {
		return nil, err
	}
	if filePath != "" {
		// Create setuppy extractor.
		return dependencies.NewSetupExtractor(filePath, pythonExecutablePath), nil
	}

	// Couldn't resolve requirements file or setup.py.
	return nil, errorutils.CheckError(errors.New("Could not find installation file for pip command, the command must contain either '--requirement <file>' or run from the directory containing 'setup.py' file."))
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
		return "", errorutils.CheckError(errors.New(fmt.Sprintf("Could not find setup.py file at current directory: %s", wd)))
	}

	return filePath, nil
}

func (pic *PipInstallCommand) saveBuildInfo(allDependencies map[string]*buildinfo.Dependency) {
	buildInfo := &buildinfo.BuildInfo{}
	var modules []buildinfo.Module
	var projectDependencies []buildinfo.Dependency

	for _, dep := range allDependencies {
		projectDependencies = append(projectDependencies, *dep)
	}

	// Save build-info.
	module := buildinfo.Module{Id: pic.buildConfiguration.Module, Dependencies: projectDependencies}
	modules = append(modules, module)

	buildInfo.Modules = modules
	utils.SaveBuildInfo(pic.buildConfiguration.BuildName, pic.buildConfiguration.BuildNumber, buildInfo)
}

func (pic *PipInstallCommand) determineModuleName(extractor dependencies.Extractor) error {
	// If module-name was set in command, don't change it.
	if pic.buildConfiguration.Module != "" {
		return nil
	}

	// Get package-name from extractor.
	moduleName, err := extractor.PackageName()
	if err != nil {
		return err
	}

	// If package-name unknown, set module as build-name.
	if moduleName == "" {
		moduleName = pic.buildConfiguration.BuildName
	}

	pic.buildConfiguration.Module = moduleName
	return nil
}

func (pic *PipInstallCommand) prepare() (pipExecutablePath, pythonExecutablePath, pipIndexUrl string, err error) {
	log.Debug("Preparing prerequisites.")

	// Get pip executable path.
	pipExecutablePath, err = getExecutablePath("pip")
	if err != nil {
		return
	}

	// Get python executable path.
	pythonExecutablePath, err = getExecutablePath("python")
	if err != nil {
		return
	}

	// Set URL for dependencies resolution.
	pipIndexUrl, err = pic.getArtifactoryUrlWithCredentials()
	if err != nil {
		return
	}

	// Extract build-info information from args.
	pic.args, pic.buildConfiguration, err = utils.ExtractBuildDetailsFromArgs(pic.args)
	if err != nil {
		return
	}

	// Prepare build-info.
	if pic.buildConfiguration.BuildName != "" && pic.buildConfiguration.BuildNumber != "" {
		pic.shouldCollectBuildInfo = true
		if err = utils.SaveBuildGeneralDetails(pic.buildConfiguration.BuildName, pic.buildConfiguration.BuildNumber); err != nil {
			return
		}
	}

	return
}

func (pic *PipInstallCommand) cleanBuildInfoDir() {
	if err := utils.RemoveBuildDir(pic.buildConfiguration.BuildName, pic.buildConfiguration.BuildNumber); err != nil {
		log.Error(fmt.Sprintf("Failed cleaning build-info directory: %s", err.Error()))
	}
}

func (pic *PipInstallCommand) getArtifactoryUrlWithCredentials() (string, error) {
	rtUrl, err := url.Parse(pic.rtDetails.GetUrl())
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	username := pic.rtDetails.GetUser()
	password := pic.rtDetails.GetPassword()

	// Get credentials from access-token if exists.
	if pic.rtDetails.GetAccessToken() != "" {
		username, err = auth.ExtractUsernameFromAccessToken(pic.rtDetails.GetAccessToken())
		if err != nil {
			return "", err
		}
		password = pic.rtDetails.GetAccessToken()
	}

	if username != "" && password != "" {
		rtUrl.User = url.UserPassword(username, password)
	}
	rtUrl.Path += "api/pypi/" + pic.repository + "/simple"

	return rtUrl.String(), nil
}

// Populate project's dependencies with checksums and file names.
// If the dependency was downloaded in this pip-install execution, checksum will be fetched from Artifactory.
// Otherwise, check if exists in cache.
func (pic *PipInstallCommand) addDepsInfo(dependenciesMap map[string]*buildinfo.Dependency) error {
	servicesManager, err := utils.CreateServiceManager(pic.rtDetails, false)
	if err != nil {
		return err
	}

	var missingDepsText []string
	dependenciesCache, err := dependencies.GetProjectDependenciesCache()
	if err != nil {
		return err
	}
	// Iterate dependencies map to update info.
	for depName := range dependenciesMap {
		// Check if this dependency was updated during this pip-install execution.
		// If updated - fetch checksum from Artifactory, regardless of what was previously stored in cache.
		depFileName, ok := pic.dependencyToFileMap[depName]
		if ok {
			// Fetch from Artifactory.
			checksum, err := getDependencyChecksumFromArtifactory(servicesManager, pic.repository, depFileName)
			if err != nil {
				return err
			}
			if checksum != nil {
				// Update dependency.
				dependenciesMap[depName].Checksum = checksum
				dependenciesMap[depName].Id = depFileName
				continue
			}
			// Failed receiving data from Artifactory, or missing checksum for file.
		}

		// If dependency wasn't downloaded in this run, check cache for dependency checksum.
		if !ok && dependenciesCache != nil {
			dep := dependenciesCache.GetDependency(depName)
			if dep != nil {
				// Checksum found in cache - update dependency.
				dependenciesMap[depName].Checksum = dep.Checksum
				dependenciesMap[depName].Id = dep.Id
				continue
			}
		}

		// Dependency not found in cache.
		missingDepsText = append(missingDepsText, depName)

		// The build-info should only contain dependencies with checksums.
		delete(dependenciesMap, depName)
	}

	// Prompt missing dependencies.
	if len(missingDepsText) > 0 {
		log.Warn(strings.Join(missingDepsText, "\n"))
		log.Warn("The pypi packages above could not be found in Artifactory or were not downloaded in this execution, therefore they are not included in the build-info.\n" +
			"Make sure the packages are available in Artifactory for this build.\n" +
			"Reinstalling in clean environment or using '--no-cache-dir' and '--force-reinstall' flags, will force downloading and populating Artifactory with these packages, the next time this command is executed.")
	}

	return nil
}

func (pic *PipInstallCommand) CommandName() string {
	return "rt_pip_install"
}

func (pic *PipInstallCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return pic.rtDetails, nil
}

// Setters.

func (pic *PipInstallCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *PipInstallCommand {
	pic.rtDetails = rtDetails
	return pic
}

func (pic *PipInstallCommand) SetRepo(repo string) *PipInstallCommand {
	pic.repository = repo
	return pic
}

func (pic *PipInstallCommand) SetArgs(arguments []string) *PipInstallCommand {
	pic.args = arguments
	return pic
}

// Get executable path.
// If run inside a virtual-env, this should return the path for the correct executable.
func getExecutablePath(executableName string) (string, error) {
	executablePath, err := exec.LookPath(executableName)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	if executablePath == "" {
		return "", errorutils.CheckError(errors2.New(fmt.Sprintf("Could not find '%s' executable", executableName)))
	}

	log.Debug(fmt.Sprintf("Found %s executable at: %s", executableName, executablePath))
	return executablePath, nil
}

func getDependencyChecksumFromArtifactory(servicesManager *artifactory.ArtifactoryServicesManager, repository, dependencyFile string) (*buildinfo.Checksum, error) {
	log.Debug(fmt.Sprintf("Fetching checksums for: %s", dependencyFile))
	result, err := servicesManager.Aql(createAqlQueryForPypi(repository, dependencyFile))
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

// TODO: Move this function to jfrog-client-go/artifactory/services/utils/aqlquerybuilder.go
func createAqlQueryForPypi(repo, file string) string {
	itemsPart :=
		`items.find({` +
			`"repo": "%s",` +
			`"$or": [{` +
			`"$and":[{` +
			`"path": {"$match": "*"},` +
			`"name": {"$match": "%s"}` +
			`}]` +
			`}]` +
			`}).include("actual_md5","actual_sha1")`
	return fmt.Sprintf(itemsPart, repo, file)
}

type aqlResult struct {
	Results []*results `json:"results,omitempty"`
}

type results struct {
	Actual_md5  string `json:"actual_md5,omitempty"`
	Actual_sha1 string `json:"actual_sha1,omitempty"`
}

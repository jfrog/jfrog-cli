package pip

import (
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	piputils "github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip/dependencies"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type PipInstallCommand struct {
	rtDetails            *config.ArtifactoryDetails
	pipExecutablePath    string
	pythonExecutablePath string
	pipIndexUrl          string
	pypiRepo             string
	buildName            string
	buildNumber          string
	moduleName           string
	collectBuildInfo     bool
	args                 []string
	projectPath          string
	dependencyToFileMap  map[string]string
	buildFile            string //Parsed from pip-install logs, maps dependency name to its actual downloaded file from Artifactory.
}

func NewPipInstallCommand() *PipInstallCommand {
	return &PipInstallCommand{}
}

func (pic *PipInstallCommand) Run() error {
	log.Info("Running pip Install.")

	// Prepare for running.
	if err := pic.preparePrerequisites(); err != nil {
		return nil
	}

	// Run pip install.
	err := pic.executePipInstall()
	if err != nil {
		pic.cleanBuildInfoDir()
		return err
	}

	// Check if need to collect build-info.
	if !pic.collectBuildInfo {
		return nil
	}

	// Collect build-info.
	if err := pic.doCollectBuildInfo(); err != nil {
		pic.cleanBuildInfoDir()
		return err
	}

	return nil
}

func (pic *PipInstallCommand) executePipInstall() error {
	// Create pip install command.
	pipInstallCmd := &piputils.PipCmd{
		Executable:  pic.pipExecutablePath,
		Command:     "install",
		CommandArgs: append(pic.args, "-i", pic.pipIndexUrl),
		EnvVars:     nil,
		StrWriter:   nil,
		ErrWriter:   nil,
	}

	// If need to collect build-info, run pip-install with log parsing.
	if pic.collectBuildInfo {
		return pic.executePipInstallWithLogParsing(pipInstallCmd)
	}

	// Run without log parsing.
	return gofrogcmd.RunCmd(pipInstallCmd)
}

// Run pip-install command while parsing the logs for downloaded packages.
// Supports running pip either in non-verbose and verbose mode.
// Populates 'dependencyToFileMap' with downloaded package-name and its actual downloaded file (wheel/egg/zip...).
func (pic *PipInstallCommand) executePipInstallWithLogParsing(pipInstallCmd *piputils.PipCmd) error {
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
	waitingForPackageFilePath := false

	// Extract downloaded package name.
	dependencyNameParser := gofrogcmd.CmdOutputPattern{
		RegExp: collectingPackageRegexp,
		ExecFunc: func(pattern *gofrogcmd.CmdOutputPattern) (string, error) {
			// If this pattern matched a second time before downloaded-file-name was found, prompt a message.
			if waitingForPackageFilePath {
				log.Debug(fmt.Sprintf("Could not resolve download path for package: %s, continuing...", packageName))
			}

			// Check for out of bound results.
			if len(pattern.MatchedResults)-1 < 0 {
				log.Debug(fmt.Sprintf("Failed extracting package name from line: %s", pattern.Line))
				return pattern.Line, nil
			}

			// Save dependency information.
			waitingForPackageFilePath = true
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
			if !waitingForPackageFilePath {
				log.Debug(fmt.Sprintf("Could not resolve package name for download path: %s , continuing...", packageName))
				return pattern.Line, nil
			}

			// Save dependency information.
			filePath := pattern.MatchedResults[1]
			downloadedDependencies[strings.ToLower(packageName)] = filePath
			waitingForPackageFilePath = false

			log.Debug(fmt.Sprintf("Found package: %s installed with: %s", packageName, filePath))
			return pattern.Line, nil
		},
	}

	// Execute command.
	_, _, err = gofrogcmd.RunCmdWithOutputParser(pipInstallCmd, true, &dependencyNameParser, &dependencyFileParser)
	if errorutils.CheckError(err) != nil {
		return err
	}

	// Update dependencyToFileMap.
	pic.dependencyToFileMap = downloadedDependencies

	return nil
}

func (pic *PipInstallCommand) doCollectBuildInfo() error {
	// Create compatible extractor for dependencies resolution, and extract dependencies.
	extractor, err := pic.createCompatibleExtractor()
	if err != nil {
		return err
	}

	err = extractor.Extract()
	if err != nil {
		return err
	}

	// If module-name wasn't set by the user, determine it.
	if pic.moduleName == "" {
		err := pic.determineModuleName(extractor)
		if err != nil {
			return err
		}
	}

	// Get project dependencies.
	allDependencies := extractor.AllDependencies()

	// Populate dependencies information - checksums and file-name.
	pic.populateDependenciesInfoAndPromptMissingDependencies(allDependencies)

	// Update project cache with correct dependencies.
	dependencies.UpdateDependenciesCache(allDependencies)

	// Save build-info built from allDependencies.
	pic.saveBuildInfo(allDependencies)

	return nil
}

func (pic *PipInstallCommand) createCompatibleExtractor() (dependencies.Extractor, error) {
	// Check if using requirements file or setup.py.
	success, err := pic.calculateRequirementsFilePathFromArgs()
	if err != nil {
		return nil, err
	}
	if success {
		// Create requirements extractor.
		return dependencies.NewRequirementsExtractor(pic.buildFile, pic.projectPath, pic.pythonExecutablePath), nil
	}

	// Setup.py should be in current dir.
	success, err = pic.calculateSetuppyFilePath()
	if err != nil {
		return nil, err
	}
	if success {
		// Create setup.py extractor.
		return dependencies.NewSetupExtractor(pic.buildFile, pic.projectPath, pic.pythonExecutablePath), nil
	}

	// Couldn't resolve requirements file or setup.py.
	return nil, errorutils.CheckError(errors.New("Could not determine installation file for pip command, the command must contain either '--requirement <file>' or run from the directory containing 'setup.py' file."))
}

func (pic *PipInstallCommand) saveBuildInfo(allDependencies map[string]*buildinfo.Dependency) {
	buildInfo := &buildinfo.BuildInfo{}
	var modules []buildinfo.Module
	var projectDependencies []buildinfo.Dependency

	for _, dep := range allDependencies {
		projectDependencies = append(projectDependencies, *dep)
	}

	// Save build-info.
	module := buildinfo.Module{Id: pic.moduleName, Dependencies: projectDependencies}
	modules = append(modules, module)

	buildInfo.Modules = modules
	utils.SaveBuildInfo(pic.buildName, pic.buildNumber, buildInfo)
}

// Cannot resolve the project path from args when using setup.py, look for setup.py in current dir.
func (pic *PipInstallCommand) calculateSetuppyFilePath() (bool, error) {
	wd, err := os.Getwd()
	if errorutils.CheckError(err) != nil {
		return false, err
	}

	// Check if setup.py exists.
	validPath, err := fileutils.IsFileExists(filepath.Join(wd, "setup.py"), false)
	if err != nil {
		return false, err
	}
	if !validPath {
		return false, errorutils.CheckError(errors.New(fmt.Sprintf("Could not find setup.py file at current directory: %s", wd)))
	}

	// Valid path.
	pic.buildFile = "setup.py"
	pic.projectPath = wd

	return true, nil
}

func (pic *PipInstallCommand) determineModuleName(extractor dependencies.Extractor) error {
	// If module-name was set in command, don't change it.
	if pic.moduleName != "" {
		return nil
	}

	// Get package-name from extractor.
	moduleName, err := extractor.PackageName()
	if err != nil {
		return err
	}

	// If package-name unknown, set module as build-name.
	if moduleName == "" {
		moduleName = pic.buildName
	}

	pic.moduleName = moduleName
	return nil
}

func (pic *PipInstallCommand) calculateRequirementsFilePathFromArgs() (bool, error) {
	requirementsFilePath, err := pic.getFlagValueFromArgs([]string{"-r", "--requirement"})
	if err != nil || requirementsFilePath == "" {
		// Args don't include a path to requirements file.
		return false, err
	}

	// Requirements file path found.
	validPath, err := fileutils.IsFileExists(requirementsFilePath, false)
	if err != nil {
		return false, err
	}
	if !validPath {
		return false, errorutils.CheckError(errors.New(fmt.Sprintf("Could not find requirements file at provided location: %s", requirementsFilePath)))
	}

	// Valid path.
	absolutePath, err := filepath.Abs(requirementsFilePath)
	if err != nil {
		return false, err
	}
	pic.projectPath, pic.buildFile = filepath.Split(absolutePath)

	return true, nil
}

// Used to search for flag value in args.
// The method returns the value of the first flag it finds from the provided slice.
func (pic *PipInstallCommand) getFlagValueFromArgs(flags []string) (string, error) {
	// Look for provided flags.
	for _, flag := range flags {
		_, _, value, err := utils.FindFlag(flag, pic.args)
		if err != nil {
			return "", err
		}
		if value != "" {
			// Found value for flag.
			return value, nil
		}
	}
	return "", nil
}

func (pic *PipInstallCommand) preparePrerequisites() error {
	log.Debug("Preparing prerequisites.")

	// Set pip executable path.
	pipExecutable, err := piputils.GetExecutablePath("pip")
	if err != nil {
		return err
	}
	pic.pipExecutablePath = pipExecutable

	// Set python executable path.
	pythonExecutable, err := piputils.GetExecutablePath("python")
	if err != nil {
		return err
	}
	pic.pythonExecutablePath = pythonExecutable

	// Set url for dependency resolution.
	artifactoryUrl, err := pic.getArtifactoryUrlWithCredentials()
	if err != nil {
		return err
	}
	pic.pipIndexUrl = artifactoryUrl

	// Extract build-info information from args.
	if err := pic.extractBuildInfoParametersFromArgs(); err != nil {
		return err
	}

	// Prepare build-info.
	if pic.buildName != "" && pic.buildNumber != "" {
		pic.collectBuildInfo = true
		if err := utils.SaveBuildGeneralDetails(pic.buildName, pic.buildNumber); err != nil {
			return err
		}
	}

	return nil
}

func (pic *PipInstallCommand) extractBuildInfoParametersFromArgs() (err error) {
	// Extract build-info information from the args.
	var flagIndex, valueIndex int
	flagIndex, valueIndex, pic.buildName, err = utils.FindFlag("--build-name", pic.args)
	if err != nil {
		return
	}
	utils.RemoveFlagFromCommand(&pic.args, flagIndex, valueIndex)

	flagIndex, valueIndex, pic.buildNumber, err = utils.FindFlag("--build-number", pic.args)
	if err != nil {
		return
	}
	utils.RemoveFlagFromCommand(&pic.args, flagIndex, valueIndex)

	flagIndex, valueIndex, pic.moduleName, err = utils.FindFlag("--module", pic.args)
	if err != nil {
		return
	}
	utils.RemoveFlagFromCommand(&pic.args, flagIndex, valueIndex)

	return
}

func (pic *PipInstallCommand) cleanBuildInfoDir() {
	if err := utils.RemoveBuildDir(pic.buildName, pic.buildNumber); err != nil {
		log.Info(fmt.Sprintf("Attempted cleaning build-info directory: %s", err.Error()))
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
		log.Debug("Using proxy with access-token.")
		username, err = auth.ExtractUsernameFromAccessToken(pic.rtDetails.GetAccessToken())
		if err != nil {
			return "", err
		}
		password = pic.rtDetails.GetAccessToken()
	}

	if username != "" && password != "" {
		rtUrl.User = url.UserPassword(username, password)
	}
	rtUrl.Path += "api/pypi/" + pic.pypiRepo + "/simple"

	return rtUrl.String(), nil
}

// Populate project's dependencies with its checksums and file-names.
// If the dependency was downloaded in this pip-install execution, checksum will be fetched from Artifactory.
// Otherwise, check if exist in cache.
func (pic *PipInstallCommand) populateDependenciesInfoAndPromptMissingDependencies(dependenciesMap map[string]*buildinfo.Dependency) error {
	servicesManager, err := utils.CreateServiceManager(pic.rtDetails, false)
	if err != nil {
		return err
	}

	var missingDependenciesText []string
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
			checksum, err := piputils.GetDependencyChecksumFromArtifactory(servicesManager, pic.pypiRepo, depFileName)
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
		missingDependenciesText = append(missingDependenciesText, depName)
		delete(dependenciesMap, depName)
	}

	// Prompt missing dependencies.
	if len(missingDependenciesText) > 0 {
		log.Warn(strings.Join(missingDependenciesText, "\n"))
		log.Warn("The pypi packages above could not be found in Artifactory and therefore are not included in the build-info.\n" +
			"Make sure the packages are available in Artifactory for this build.\n" +
			"Uninstalling packages from the environment will force populating Artifactory with these packages.")
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
	pic.pypiRepo = repo
	return pic
}

func (pic *PipInstallCommand) SetArgs(arguments []string) *PipInstallCommand {
	pic.args = arguments
	return pic
}

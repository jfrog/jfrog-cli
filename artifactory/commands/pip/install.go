package pip

import (
	"encoding/json"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
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
	"path/filepath"
	"regexp"
	"strings"
)

type PipInstallCommand struct {
	buildConfiguration   *utils.BuildConfiguration
	rtDetails            *config.ArtifactoryDetails
	pipExecutablePath    string
	pythonExecutablePath string
	pipIndexUrl          string
	pypiRepo             string

	buildName           string
	buildNumber         string
	moduleName          string
	collectBuildInfo    bool
	args                []string
	projectPath         string
	buildFile           string
	dependencyToFileMap map[string]string //Parsed from pip-install logs, maps dependency name to its actual downloaded file from Artifactory.
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
		return nil
	}

	return nil
}

func (pic *PipInstallCommand) executePipInstall() error {
	// Create pip install command.
	pipInstallCmd := &pip.PipCmd{
		Executable:  pic.pipExecutablePath,
		Command:     fmt.Sprintf("install"),
		CommandArgs: append(pic.args, fmt.Sprintf("-i %s", pic.pipIndexUrl)),
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

func (pic *PipInstallCommand) executePipInstallWithLogParsing(pipInstallCmd *pip.PipCmd) error {
	regexp, nameMatchGroup, fileMatchGroup, err := pic.getParsingRegexpForPipInstallCommand()
	if err != nil {
		return err
	}

	dependencies := make(map[string]string)
	// Read the log line, identify dependency name and its corresponding file from Artifactory.
	protocolRegExp := gofrogcmd.CmdOutputPattern{
		RegExp: regexp,
		ExecFunc: func(pattern *gofrogcmd.CmdOutputPattern) (string, error) {
			// Match - extract dependency information.

			// Check for out of bound results.
			if len(pattern.MatchedResults)-1 < nameMatchGroup || len(pattern.MatchedResults)-1 < fileMatchGroup {
				log.Debug(fmt.Sprintf("Failed extracting dependency information from line: %s", pattern.Line))
				return "", nil
			}

			// Save dependency information.
			depName := pattern.MatchedResults[nameMatchGroup]
			file := pattern.MatchedResults[fileMatchGroup]
			dependencies[strings.ToLower(depName)] = file

			log.Debug(fmt.Sprintf("Found dependency: %s installed with: %s", pattern.MatchedResults[nameMatchGroup], file))
			return "", nil
		},
	}

	// Execute command.
	_, _, err = gofrogcmd.RunCmdWithOutputParser(pipInstallCmd, true, &protocolRegExp)
	if errorutils.CheckError(err) != nil {
		return err
	}

	// Update dependencyToFileMap.
	pic.dependencyToFileMap = dependencies

	return nil
}

// Return the regexp to use for parsing pip-install log for dependencies name and file.
// regexp - The regular expression to run.
// depNameMatchGroup - Match group for the resolved dependency name.
// fileMatchGroup - Match group for the downloaded file in Artifactory.
func (pic *PipInstallCommand) getParsingRegexpForPipInstallCommand() (regexp *regexp.Regexp, depNameMatchGroup, fileMatchGroup int, err error) {
	// Check if args contain -v or --verbose -> use appropriate regex for these cases.
	if utils.FindBooleanFlag("-v", pic.args) != -1 || utils.FindBooleanFlag("--verbose", pic.args) != -1 {
		// Command is in verbose mode.
		regexp, err = clientutils.GetRegExp(`^\s\sDownloading\sfrom\sURL\s.*\/packages\/.*\/(.*)\#sha256=[A-Fa-f0-9]{64}\s\(from\s.*\/(\w[\w-\.]+)\/\)$`)
		depNameMatchGroup = 2
		fileMatchGroup = 1
		return
	}

	// Command is non-verbose.
	// TODO: This regexp isn't good for the non-verbose output, as we have to parse 2 log lines each time (notice the \n in the regexp).
	regexp, err = clientutils.GetRegExp(`\nCollecting\s(\w[\w-\.]+).*\n\s\sDownloading\s.*\/packages\/.*\/(.*)`)
	depNameMatchGroup = 1
	fileMatchGroup = 2
	return
}

func (pic *PipInstallCommand) doCollectBuildInfo() error {
	// Create compatible extractor for dependencies resolution, and extract dependencies.
	extractor, err := pic.createCompatibleExtractor()
	if err != nil {
		return err
	}
	extractor.Extract()

	// TODO: decide module name:
	// If setup.py: Run python egg_info and parse PKG-INFO for module name.
	// If requirements, put build-name as module name.
	pic.moduleName = pic.buildName

	// Get project dependencies.
	allDependencies := extractor.AllDependencies()

	// Populate dependencies checksums.
	pic.populateDependenciesChecksum(allDependencies)

	// Clear cache from unnecessary dependencies.
	dependencies.UpdateDependenciesCache(allDependencies)

	// Save build-info built from allDependencies.
	pic.saveBuildInfoAndPromptMissingDependencyChecksums(allDependencies)

	return nil
}

func (pic *PipInstallCommand) createCompatibleExtractor() (dependencies.Extractor, error) {
	// Check if using requirements file.
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
		// Create setuppy extractor.
		return dependencies.NewSetupExtractor(pic.buildFile, pic.projectPath, pic.pythonExecutablePath), nil
	}

	// Couldn't resolve requirements file or setup.py.
	return nil, errorutils.CheckError(errors.New("Could not determine installation file for pip command, the command must contain either '--requirement <file>' or run from the directory containing 'setup.py' file."))
}

func (pic *PipInstallCommand) saveBuildInfoAndPromptMissingDependencyChecksums(allDependencies map[string]*buildinfo.Dependency) {
	buildInfo := &buildinfo.BuildInfo{}
	var modules []buildinfo.Module
	var projectDependencies []buildinfo.Dependency
	var missingDependenciesText []string

	for _, dep := range allDependencies {
		projectDependencies = append(projectDependencies, *dep)
		if dep.Checksum == nil || dep.Checksum.Sha1 == "" || dep.Checksum.Md5 == "" {
			// Dependency missing checksum.
			missingDependenciesText = append(missingDependenciesText, dep.Id)
		}
	}

	// Save build-info.
	module := buildinfo.Module{Id: pic.moduleName, Dependencies: projectDependencies}
	modules = append(modules, module)
	utils.SaveBuildInfo(pic.buildConfiguration.BuildName, pic.buildConfiguration.BuildNumber, buildInfo)

	// Prompt missing dependencies.
	if len(missingDependenciesText) > 0 {
		log.Warn(strings.Join(missingDependenciesText, "\n"))
		log.Warn("The npm dependencies above could not be found in Artifactory and therefore are not included in the build-info.\n" +
			"Make sure the dependencies are available in Artifactory for this build.\n" +
			"Uninstalling packages from the environment will force populating Artifactory with these dependencies.")
	}
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
	pic.buildFile, pic.projectPath = filepath.Split(absolutePath)

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
	pipExecutable, err := pip.GetExecutablePath("pip")
	if err != nil {
		return err
	}
	pic.pipExecutablePath = pipExecutable

	// Set python executable path.
	pythonExecutable, err := pip.GetExecutablePath("python")
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
	rtUrl.Path += "api/pypi/" + pic.pypiRepo + "simple"

	return rtUrl.String(), nil
}

// This methods populates project's dependencies with its checksums.
// If the dependency was downloaded in this pip-install execution, checksum will be fetched from Artifactory.
// if not, check if exist in cache.
func (pic *PipInstallCommand) populateDependenciesChecksum(dependenciesMap map[string]*buildinfo.Dependency) error {
	servicesManager, err := utils.CreateServiceManager(pic.rtDetails, false)
	if err != nil {
		return err
	}

	// Iterate dependencies map to update checksums.
	for depName := range dependenciesMap {
		// Check if this dependency was updated during this pip-install execution.
		// If updated - fetch checksum from Artifactory, regardless of what was previously stored in cache.
		depFileName, ok := pic.dependencyToFileMap[depName]
		if ok {
			// Fetch from Artifactory.
			checksum, err := getDependencyChecksumFromArtifactory(servicesManager, pic.pypiRepo, depFileName)
			if err != nil {
				return err
			}
			// Update dependency.
			dependenciesMap[depName].Checksum = checksum

			continue
		}

		// Check cache for dependency checksum.
		checksum, err := dependencies.GetDependencyChecksum(depName, depFileName)
		if err != nil {
			return err
		}
		if checksum == nil {
			// Checksum not found in cache.
			continue
		}
		// Checksum found in cache - update dependency.
		dependenciesMap[depName].Checksum = checksum
	}

	return nil
}

func getDependencyChecksumFromArtifactory(servicesManager *artifactory.ArtifactoryServicesManager, repository, dependencyFile string) (checksum *buildinfo.Checksum, err error) {
	log.Debug(fmt.Sprintf("Fetching checksums for: %s", dependencyFile))
	result, err := servicesManager.Aql(createAqlQueryForPypi(repository, dependencyFile))
	if err != nil {
		return
	}

	parsedResult := new(aqlResult)
	err = json.Unmarshal(result, parsedResult)
	if err = errorutils.CheckError(err); err != nil {
		return
	}

	if len(parsedResult.Results) == 0 {
		log.Debug(fmt.Sprintf("File: %s could not be found in repository: %s", dependencyFile, repository))
		return
	}

	checksum = &buildinfo.Checksum{Sha1: parsedResult.Results[0].Actual_sha1, Md5: parsedResult.Results[0].Actual_md5}
	log.Debug(fmt.Sprintf("Found checksums for file: %s, sha1: %s, md5:%s", dependencyFile, parsedResult.Results[0].Actual_sha1, parsedResult.Results[0].Actual_md5))

	return
}

// TODO: Move this function to jfrog-client-go/artifactory/services/utils/aqlquerybuilder.go
func createAqlQueryForPypi(repo, file string) string {
	itemsPART :=
		`items.find({` +
			`"repo": "%s",` +
			`"$or": [{` +
			`"$and":[{` +
			`"path": {"$match": "*"},` +
			`"name": {"$match": "%s"}` +
			`}]` +
			`}]` +
			`}).include("actual_md5","actual_sha1")`
	return fmt.Sprintf(itemsPART, repo, file)
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

func (pic *PipInstallCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *PipInstallCommand {
	pic.buildConfiguration = buildConfiguration
	return pic
}

type aqlResult struct {
	Results []*results `json:"results,omitempty"`
}

type results struct {
	Actual_md5  string `json:"actual_md5,omitempty"`
	Actual_sha1 string `json:"actual_sha1,omitempty"`
}

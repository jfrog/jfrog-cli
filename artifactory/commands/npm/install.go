package npm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/gofrog/parallel"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/npm"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/ioutils"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	serviceutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/httpclient"
	cliutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"github.com/mattn/go-shellwords"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const npmrcFileName = ".npmrc"
const npmrcBackupFileName = "jfrog.npmrc.backup"
const minSupportedArtifactoryVersion = "5.5.2"
const minSupportedNpmVersion = "5.4.0"

type NpmInstallCommand struct {
	threads          int
	executablePath   string
	npmrcFileMode    os.FileMode
	workingDirectory string
	registry         string
	npmAuth          string
	collectBuildInfo bool
	dependencies     map[string]*dependency
	typeRestriction  string
	artDetails       auth.ArtifactoryDetails
	packageInfo      *npm.PackageInfo
	NpmCommand
}

func (nic *NpmInstallCommand) SetThreads(threads int) *NpmInstallCommand {
	nic.threads = threads
	return nic
}

func (nic *NpmInstallCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return nic.rtDetails, nil
}

func (nic *NpmInstallCommand) Run() error {
	log.Info("Running npm Install.")
	if err := nic.preparePrerequisites(nic.repo); err != nil {
		return err
	}

	if err := nic.createTempNpmrc(); err != nil {
		return nic.restoreNpmrcAndError(err)
	}

	if err := nic.runInstall(); err != nil {
		return nic.restoreNpmrcAndError(err)
	}

	if err := nic.restoreNpmrc(); err != nil {
		return err
	}

	if !nic.collectBuildInfo {
		log.Info("npm install finished successfully.")
		return nil
	}

	if err := nic.setDependenciesList(); err != nil {
		return err
	}

	if err := nic.collectDependenciesChecksums(); err != nil {
		return err
	}

	if err := nic.saveDependenciesData(); err != nil {
		return err
	}

	log.Info("npm install finished successfully.")
	return nil
}

func (nic *NpmInstallCommand) CommandName() string {
	return "rt_npm_install"
}

func (nic *NpmInstallCommand) preparePrerequisites(repo string) error {
	log.Debug("Preparing prerequisites.")
	if err := nic.setNpmExecutable(); err != nil {
		return err
	}

	if err := nic.validateNpmVersion(); err != nil {
		return err
	}

	if err := nic.setWorkingDirectory(); err != nil {
		return err
	}

	if err := nic.prepareArtifactoryPrerequisites(repo); err != nil {
		return err
	}

	if err := nic.prepareBuildInfo(); err != nil {
		return err
	}

	return nic.backupProjectNpmrc()
}

func (nic *NpmInstallCommand) prepareArtifactoryPrerequisites(repo string) (err error) {
	npmAuth, artifactoryVersion, err := getArtifactoryDetails(nic.artDetails)
	if err != nil {
		return err
	}

	nic.npmAuth = npmAuth
	if version.Compare(artifactoryVersion, minSupportedArtifactoryVersion) < 0 && artifactoryVersion != "development" {
		return errorutils.CheckError(errors.New("This operation requires Artifactory version " + minSupportedArtifactoryVersion + " or higher."))
	}

	if err = utils.CheckIfRepoExists(repo, nic.artDetails); err != nil {
		return err
	}

	nic.registry = getNpmRepositoryUrl(repo, nic.artDetails.GetUrl())
	return nil
}

func (nic *NpmInstallCommand) prepareBuildInfo() error {
	var err error
	if len(nic.buildConfiguration.BuildName) > 0 && len(nic.buildConfiguration.BuildNumber) > 0 {
		nic.collectBuildInfo = true
		if err = utils.SaveBuildGeneralDetails(nic.buildConfiguration.BuildName, nic.buildConfiguration.BuildNumber); err != nil {
			return err
		}

		if nic.packageInfo, err = npm.ReadPackageInfoFromPackageJson(nic.workingDirectory); err != nil {
			return err
		}
	}
	return err
}

func (nic *NpmInstallCommand) setWorkingDirectory() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return errorutils.CheckError(err)
	}

	if currentDir, err = filepath.Abs(currentDir); err != nil {
		return errorutils.CheckError(err)
	}

	nic.workingDirectory = currentDir
	log.Debug("Working directory set to:", nic.workingDirectory)
	if err = nic.setArtifactoryAuth(); err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

// In order to make sure the install downloads the dependencies from Artifactory, we are creating a.npmrc file in the project's root directory.
// If such a file already exists, we are copying it aside.
// This method restores the backed up file and deletes the one created by the command.
func (nic *NpmInstallCommand) restoreNpmrc() (err error) {
	log.Debug("Restoring project .npmrc file")
	if err = os.Remove(filepath.Join(nic.workingDirectory, npmrcFileName)); err != nil {
		return errorutils.CheckError(errors.New(createRestoreErrorPrefix(nic.workingDirectory) + err.Error()))
	}
	log.Debug("Deleted the temporary .npmrc file successfully")

	if _, err = os.Stat(filepath.Join(nic.workingDirectory, npmrcBackupFileName)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errorutils.CheckError(errors.New(createRestoreErrorPrefix(nic.workingDirectory) + err.Error()))
	}

	if err = ioutils.CopyFile(
		filepath.Join(nic.workingDirectory, npmrcBackupFileName),
		filepath.Join(nic.workingDirectory, npmrcFileName), nic.npmrcFileMode); err != nil {
		return errorutils.CheckError(err)
	}
	log.Debug("Restored project .npmrc file successfully")

	if err = os.Remove(filepath.Join(nic.workingDirectory, npmrcBackupFileName)); err != nil {
		return errorutils.CheckError(errors.New(createRestoreErrorPrefix(nic.workingDirectory) + err.Error()))
	}
	log.Debug("Deleted project", npmrcBackupFileName, "file successfully")
	return nil
}

func createRestoreErrorPrefix(workingDirectory string) string {
	return fmt.Sprintf("Error occurred while restoring project .npmrc file. "+
		"Delete '%s' and move '%s' (if exists) to '%s' in order to restore the project. Failure cause: \n",
		filepath.Join(workingDirectory, npmrcFileName),
		filepath.Join(workingDirectory, npmrcBackupFileName),
		filepath.Join(workingDirectory, npmrcFileName))
}

// In order to make sure the install downloads the artifacts from Artifactory we creating in the project .npmrc file.
// If such a file exists we storing a copy of it in npmrcBackupFileName.
func (nic *NpmInstallCommand) createTempNpmrc() error {
	log.Debug("Creating project .npmrc file.")
	data, err := npm.GetConfigList(nic.npmArgs, nic.executablePath)
	configData, err := nic.prepareConfigData(data)
	if err != nil {
		return errorutils.CheckError(err)
	}

	if err = removeNpmrcIfExists(nic.workingDirectory); err != nil {
		return err
	}

	return errorutils.CheckError(ioutil.WriteFile(filepath.Join(nic.workingDirectory, npmrcFileName), configData, nic.npmrcFileMode))
}

func (nic *NpmInstallCommand) runInstall() error {
	log.Debug("Running npmi install command.")
	splitArgs, err := shellwords.Parse(nic.npmArgs)
	if err != nil {
		return errorutils.CheckError(err)
	}
	filteredArgs := filterFlags(splitArgs)
	installCmdConfig := &npm.NpmConfig{
		Npm:          nic.executablePath,
		Command:      append([]string{"install"}, filteredArgs...),
		CommandFlags: nil,
		StrWriter:    nil,
		ErrWriter:    nil,
	}

	if nic.collectBuildInfo && len(filteredArgs) > 0 {
		log.Warn("Build info dependencies collection with npm arguments is not supported. Build info creation will be skipped.")
		nic.collectBuildInfo = false
	}

	return errorutils.CheckError(gofrogcmd.RunCmd(installCmdConfig))
}

func (nic *NpmInstallCommand) setDependenciesList() (err error) {
	nic.dependencies = make(map[string]*dependency)
	// nic.scope can be empty, "production" or "development" in case of empty both of the functions should run
	if nic.typeRestriction != "production" {
		if err = nic.prepareDependencies("development"); err != nil {
			return
		}
	}
	if nic.typeRestriction != "development" {
		err = nic.prepareDependencies("production")
	}
	return
}

func (nic *NpmInstallCommand) collectDependenciesChecksums() error {
	log.Info("Collecting dependencies information... This may take a few minuets...")
	servicesManager, err := utils.CreateServiceManager(nic.rtDetails, false)
	if err != nil {
		return err
	}

	producerConsumer := parallel.NewBounedRunner(nic.threads, false)
	errorsQueue := serviceutils.NewErrorsQueue(1)
	handlerFunc := nic.createGetDependencyInfoFunc(servicesManager)
	go func() {
		defer producerConsumer.Done()
		for i := range nic.dependencies {
			producerConsumer.AddTaskWithError(handlerFunc(i), errorsQueue.AddError)
		}
	}()
	producerConsumer.Run()
	return errorsQueue.GetError()
}

func (nic *NpmInstallCommand) saveDependenciesData() error {
	log.Debug("Saving install data.")
	dependencies, missingDependencies := nic.transformDependencies()
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Dependencies = dependencies
		partial.ModuleId = nic.packageInfo.BuildInfoModuleId()
	}

	if err := utils.SavePartialBuildInfo(nic.buildConfiguration.BuildName, nic.buildConfiguration.BuildNumber, populateFunc); err != nil {
		return err
	}

	if len(missingDependencies) > 0 {
		var missingDependenciesText []string
		for _, dependency := range missingDependencies {
			missingDependenciesText = append(missingDependenciesText, dependency.name+"-"+dependency.version)
		}
		log.Warn(strings.Join(missingDependenciesText, "\n"))
		log.Warn("The npm dependencies above could not be found in Artifactory and therefore are not included in the build-info.\n" +
			"Make sure the dependencies are available in Artifactory for this build.\n" +
			"Deleting the local cache will force populating Artifactory with these dependencies.")
	}
	return nil
}

func (nic *NpmInstallCommand) validateNpmVersion() error {
	npmVersion, err := npm.Version(nic.executablePath)
	if err != nil {
		return err
	}
	if version.Compare(string(npmVersion), minSupportedNpmVersion) < 0 {
		return errorutils.CheckError(errors.New("JFrog cli npm-install command requires npm client version " + minSupportedNpmVersion + " or higher."))
	}
	return nil
}

// To make npm do the resolution from Artifactory we are creating .npmrc file in the project dir.
// If a .npmrc file already exists we will backup it and override while running the command
func (nic *NpmInstallCommand) backupProjectNpmrc() error {
	fileInfo, err := os.Stat(filepath.Join(nic.workingDirectory, npmrcFileName))
	if err != nil {
		if os.IsNotExist(err) {
			nic.npmrcFileMode = 0644
			return nil
		}
		return errorutils.CheckError(err)
	}

	nic.npmrcFileMode = fileInfo.Mode()
	src := filepath.Join(nic.workingDirectory, npmrcFileName)
	dst := filepath.Join(nic.workingDirectory, npmrcBackupFileName)
	if err = ioutils.CopyFile(src, dst, nic.npmrcFileMode); err != nil {
		return err
	}
	log.Debug("Project .npmrc file backed up successfully to", filepath.Join(nic.workingDirectory, npmrcBackupFileName))
	return nil
}

// This func transforms "npm config list --json" result to key=val list of values that can be set to .npmrc file.
// it filters any nil values key, changes registry and scope registries to Artifactory url and adds Artifactory authentication to the list
func (nic *NpmInstallCommand) prepareConfigData(data []byte) ([]byte, error) {
	var collectedConfig map[string]interface{}
	var filteredConf []string
	if err := json.Unmarshal(data, &collectedConfig); err != nil {
		return nil, errorutils.CheckError(err)
	}

	for i := range collectedConfig {
		if isValidKeyVal(i, collectedConfig[i]) {
			filteredConf = append(filteredConf, i, " = ", fmt.Sprint(collectedConfig[i]), "\n")
		} else if strings.HasPrefix(i, "@") {
			// Override scoped registries (@scope = xyz)
			filteredConf = append(filteredConf, i, " = ", nic.registry, "\n")
		}
		nic.setTypeRestriction(i, collectedConfig[i])
	}
	filteredConf = append(filteredConf, "registry = ", nic.registry, "\n")
	filteredConf = append(filteredConf, nic.npmAuth)
	return []byte(strings.Join(filteredConf, "")), nil
}

// npm install type restriction can be set by "--production" or "-only={prod[uction]|dev[elopment]}" flags
func (nic *NpmInstallCommand) setTypeRestriction(key string, val interface{}) {
	if key == "production" && val != nil && (val == true || val == "true") {
		nic.typeRestriction = "production"
	} else if key == "only" && val != nil {
		if strings.Contains(val.(string), "prod") {
			nic.typeRestriction = "production"
		} else if strings.Contains(val.(string), "dev") {
			nic.typeRestriction = "development"
		}
	}
}

// Run npm list and parse the returned json
func (nic *NpmInstallCommand) prepareDependencies(typeRestriction string) error {
	// Run npm list
	data, errData, err := npm.RunList(nic.npmArgs+" -only="+typeRestriction, nic.executablePath)
	if err != nil {
		log.Warn("npm list command failed with error:", err.Error())
	}
	if len(errData) > 0 {
		log.Warn("Some errors occurred while collecting dependencies info:\n" + string(errData))
	}

	// Parse the dependencies json object
	return jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if string(key) == "dependencies" {
			err := nic.parseDependencies(value, typeRestriction)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Parses npm dependencies recursively and adds the collected dependencies to nic.dependencies
func (nic *NpmInstallCommand) parseDependencies(data []byte, scope string) error {
	var transitiveDependencies [][]byte
	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		ver, _, _, err := jsonparser.Get(data, string(key), "version")
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			return errorutils.CheckError(err)
		} else if err == jsonparser.KeyPathNotFoundError {
			log.Warn(fmt.Sprintf("npm dependencies list contains the package '%s' without version information. The dependency will not be added to build-info.", string(key)))
		} else {
			nic.appendDependency(key, ver, scope)
		}
		transitive, _, _, err := jsonparser.Get(data, string(key), "dependencies")
		if err != nil && err.Error() != "Key path not found" {
			return errorutils.CheckError(err)
		}

		if len(transitive) > 0 {
			transitiveDependencies = append(transitiveDependencies, transitive)
		}
		return nil
	})

	if err != nil {
		return err
	}

	for _, element := range transitiveDependencies {
		err := nic.parseDependencies(element, scope)
		if err != nil {
			return err
		}
	}
	return nil
}

func (nic *NpmInstallCommand) appendDependency(key []byte, ver []byte, scope string) {
	dependencyKey := string(key) + "-" + string(ver)
	if nic.dependencies[dependencyKey] == nil {
		nic.dependencies[dependencyKey] = &dependency{name: string(key), version: string(ver), scopes: []string{scope}}
	} else if !scopeAlreadyExists(scope, nic.dependencies[dependencyKey].scopes) {
		nic.dependencies[dependencyKey].scopes = append(nic.dependencies[dependencyKey].scopes, scope)
	}
}

// Creates a function that fetches dependency data from Artifactory. Can be applied from a producer-consumer mechanism
func (nic *NpmInstallCommand) createGetDependencyInfoFunc(servicesManager *artifactory.ArtifactoryServicesManager) getDependencyInfoFunc {
	return func(dependencyIndex string) parallel.TaskFunc {
		return func(threadId int) error {
			name := nic.dependencies[dependencyIndex].name
			ver := nic.dependencies[dependencyIndex].version
			log.Debug(cliutils.GetLogMsgPrefix(threadId, false), "Fetching checksums for", name, "-", ver)
			result, err := servicesManager.Aql(serviceutils.CreateAqlQueryForNpm(name, ver))
			if err != nil {
				return err
			}

			parsedResult := new(aqlResult)
			if err = json.Unmarshal(result, parsedResult); err != nil {
				return errorutils.CheckError(err)
			}
			if len(parsedResult.Results) == 0 {
				log.Debug(cliutils.GetLogMsgPrefix(threadId, false), name, "-", ver, "could not be found in Artifactory.")
				return nil
			}
			nic.dependencies[dependencyIndex].artifactName = parsedResult.Results[0].Name
			nic.dependencies[dependencyIndex].checksum =
				&buildinfo.Checksum{Sha1: parsedResult.Results[0].Actual_sha1, Md5: parsedResult.Results[0].Actual_md5}
			log.Debug(cliutils.GetLogMsgPrefix(threadId, false), "Found", parsedResult.Results[0].Name,
				"sha1:", parsedResult.Results[0].Actual_sha1,
				"md5", parsedResult.Results[0].Actual_md5)
			return nil
		}
	}
}

// Transforms the list of dependencies to buildinfo.Dependencies list and creates a list of dependencies that are missing in Artifactory.
func (nic *NpmInstallCommand) transformDependencies() (dependencies []buildinfo.Dependency, missingDependencies []dependency) {
	for _, dependency := range nic.dependencies {
		if dependency.artifactName != "" {
			dependencies = append(dependencies,
				buildinfo.Dependency{Id: dependency.artifactName, Scopes: dependency.scopes, Checksum: dependency.checksum})
		} else {
			missingDependencies = append(missingDependencies, *dependency)
		}
	}
	return
}

func (nic *NpmInstallCommand) restoreNpmrcAndError(err error) error {
	if restoreErr := nic.restoreNpmrc(); restoreErr != nil {
		return errors.New(fmt.Sprintf("Two errors occurred:\n %s\n %s", restoreErr.Error(), err.Error()))
	}
	return err
}

func (nic *NpmInstallCommand) setArtifactoryAuth() error {
	authArtDetails, err := nic.rtDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	if authArtDetails.GetSshAuthHeaders() != nil {
		return errorutils.CheckError(errors.New("SSH authentication is not supported in this command."))
	}
	nic.artDetails = authArtDetails
	return nil
}

func removeNpmrcIfExists(workingDirectory string) error {
	if _, err := os.Stat(filepath.Join(workingDirectory, npmrcFileName)); err != nil {
		if os.IsNotExist(err) { // The file dose not exist, nothing to do.
			return nil
		}
		return errorutils.CheckError(err)
	}

	log.Debug("Removing Existing .npmrc file")
	return errorutils.CheckError(os.Remove(filepath.Join(workingDirectory, npmrcFileName)))
}

func (nic *NpmInstallCommand) setNpmExecutable() error {
	npmExecPath, err := exec.LookPath("npm")
	if err != nil {
		return errorutils.CheckError(err)
	}

	if npmExecPath == "" {
		return errorutils.CheckError(errors.New("Could not find 'npm' executable"))
	}
	nic.executablePath = npmExecPath
	log.Debug("Found npm executable at:", nic.executablePath)
	return nil
}

func getArtifactoryDetails(artDetails auth.ArtifactoryDetails) (npmAuth string, artifactoryVersion string, err error) {
	if artDetails.GetAccessToken() == "" {
		return getDetailsUsingBasicAuth(artDetails)
	}

	return getDetailsUsingAccessToken(artDetails)
}

func getDetailsUsingAccessToken(artDetails auth.ArtifactoryDetails) (npmAuth string, artifactoryVersion string, err error) {
	npmAuthString := "_auth = %s\nalways-auth = true"
	// Build npm token, consists of <username:password> encoded.
	// Use Artifactory's access-token as username and password to create npm token.
	username, err := auth.ExtractUsernameFromAccessToken(artDetails.GetAccessToken())
	if err != nil {
		return "", "", err
	}
	encodedNpmToken := base64.StdEncoding.EncodeToString([]byte(username + ":" + artDetails.GetAccessToken()))
	npmAuth = fmt.Sprintf(npmAuthString, encodedNpmToken)

	// Get Artifactory version.
	rtVersion, err := artDetails.GetVersion()
	if err != nil {
		return "", "", err
	}

	return npmAuth, rtVersion, err
}

func getDetailsUsingBasicAuth(artDetails auth.ArtifactoryDetails) (npmAuth string, artifactoryVersion string, err error) {
	authApiUrl := artDetails.GetUrl() + "api/npm/auth"
	log.Debug("Sending npm auth request")

	// Get npm token from Artifactory.
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return "", "", err
	}
	resp, body, _, err := client.SendGet(authApiUrl, true, artDetails.CreateHttpClientDetails())
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	// Extract Artifactory version from response header.
	serverValues := strings.Split(resp.Header.Get("Server"), "/")
	if len(serverValues) != 2 {
		errorutils.CheckError(errors.New("Cannot parse Artifactory version from the server header."))
	}

	return string(body), strings.TrimSpace(serverValues[1]), err
}

func getNpmRepositoryUrl(repo, url string) string {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "api/npm/" + repo
	return url
}

func scopeAlreadyExists(scope string, existingScopes []string) bool {
	for _, existingScope := range existingScopes {
		if existingScope == scope {
			return true
		}
	}
	return false
}

// Valid configs keys are not related to registry (registry = xyz) or scoped registry (@scope = xyz)) and have data in their value
func isValidKeyVal(key string, val interface{}) bool {
	return !strings.HasPrefix(key, "//") &&
		!strings.HasPrefix(key, "@") &&
		key != "registry" &&
		key != "metrics-registry" &&
		val != nil &&
		val != ""
}

func filterFlags(splitArgs []string) []string {
	var filteredArgs []string
	for _, arg := range splitArgs {
		if !strings.HasPrefix(arg, "-") {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	return filteredArgs
}

type getDependencyInfoFunc func(string) parallel.TaskFunc

type npmInstall struct {
	npmCommandConfig *NpmCommand
}

type dependency struct {
	name         string
	version      string
	scopes       []string
	artifactName string
	checksum     *buildinfo.Checksum
}

type aqlResult struct {
	Results []*results `json:"results,omitempty"`
}

type results struct {
	Name        string `json:"name,omitempty"`
	Actual_md5  string `json:"actual_md5,omitempty"`
	Actual_sha1 string `json:"actual_sha1,omitempty"`
}

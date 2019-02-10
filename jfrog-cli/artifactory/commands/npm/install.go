package npm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/gofrog/parallel"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/npm"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/ioutils"
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

func Install(repo string, cliConfiguration *npm.CliConfiguration) (err error) {
	log.Info("Running npm Install.")
	npmi := npmInstall{cliConfig: cliConfiguration}
	if err = npmi.preparePrerequisites(repo); err != nil {
		return err
	}

	if err = npmi.createTempNpmrc(); err != nil {
		return npmi.restoreNpmrcAndError(err)
	}

	if err = npmi.runInstall(); err != nil {
		return npmi.restoreNpmrcAndError(err)
	}

	if err = npmi.restoreNpmrc(); err != nil {
		return err
	}

	if !npmi.collectBuildInfo {
		log.Info("npm install finished successfully.")
		return nil
	}

	if err = npmi.setDependenciesList(); err != nil {
		return err
	}

	if err = npmi.collectDependenciesChecksums(); err != nil {
		return err
	}

	if err = npmi.saveDependenciesData(); err != nil {
		return err
	}

	log.Info("npm install finished successfully.")
	return
}

func (npmi *npmInstall) preparePrerequisites(repo string) error {
	log.Debug("Preparing prerequisites.")
	if err := npmi.setNpmExecutable(); err != nil {
		return err
	}

	if err := npmi.validateNpmVersion(); err != nil {
		return err
	}

	if err := npmi.setWorkingDirectory(); err != nil {
		return err
	}

	if err := npmi.prepareArtifactoryPrerequisites(repo); err != nil {
		return err
	}

	if err := npmi.prepareBuildInfo(); err != nil {
		return err
	}

	return npmi.backupProjectNpmrc()
}

func (npmi *npmInstall) prepareArtifactoryPrerequisites(repo string) (err error) {
	npmAuth, artifactoryVersion, err := getArtifactoryDetails(npmi.artDetails)
	if err != nil {
		return err
	}

	npmi.npmAuth = string(npmAuth)
	if version.Compare(artifactoryVersion, minSupportedArtifactoryVersion) < 0 && artifactoryVersion != "development" {
		return errorutils.CheckError(errors.New("This operation requires Artifactory version " + minSupportedArtifactoryVersion + " or higher."))
	}

	if err = utils.CheckIfRepoExists(repo, npmi.artDetails); err != nil {
		return err
	}

	npmi.registry = getNpmRepositoryUrl(repo, npmi.artDetails.GetUrl())
	return nil
}

func (npmi *npmInstall) prepareBuildInfo() error {
	var err error
	if len(npmi.cliConfig.BuildName) > 0 && len(npmi.cliConfig.BuildNumber) > 0 {
		npmi.collectBuildInfo = true
		if err = utils.SaveBuildGeneralDetails(npmi.cliConfig.BuildName, npmi.cliConfig.BuildNumber); err != nil {
			return err
		}

		if npmi.packageInfo, err = npm.ReadPackageInfoFromPackageJson(npmi.workingDirectory); err != nil {
			return err
		}
	}
	return err
}

func (npmi *npmInstall) setWorkingDirectory() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return errorutils.CheckError(err)
	}

	if currentDir, err = filepath.Abs(currentDir); err != nil {
		return errorutils.CheckError(err)
	}

	npmi.workingDirectory = currentDir
	log.Debug("Working directory set to:", npmi.workingDirectory)
	if err = npmi.setArtifactoryAuth(); err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

// In order to make sure the install downloads the dependencies from Artifactory, we are creating a.npmrc file in the project's root directory.
// If such a file already exists, we are copying it aside.
// This method restores the backed up file and deletes the one created by the command.
func (npmi *npmInstall) restoreNpmrc() (err error) {
	log.Debug("Restoring project .npmrc file")
	if err = os.Remove(filepath.Join(npmi.workingDirectory, npmrcFileName)); err != nil {
		return errorutils.CheckError(errors.New(createRestoreErrorPrefix(npmi.workingDirectory) + err.Error()))
	}
	log.Debug("Deleted the temporary .npmrc file successfully")

	if _, err = os.Stat(filepath.Join(npmi.workingDirectory, npmrcBackupFileName)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errorutils.CheckError(errors.New(createRestoreErrorPrefix(npmi.workingDirectory) + err.Error()))
	}

	if err = ioutils.CopyFile(
		filepath.Join(npmi.workingDirectory, npmrcBackupFileName),
		filepath.Join(npmi.workingDirectory, npmrcFileName), npmi.npmrcFileMode); err != nil {
		return errorutils.CheckError(err)
	}
	log.Debug("Restored project .npmrc file successfully")

	if err = os.Remove(filepath.Join(npmi.workingDirectory, npmrcBackupFileName)); err != nil {
		return errorutils.CheckError(errors.New(createRestoreErrorPrefix(npmi.workingDirectory) + err.Error()))
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
func (npmi *npmInstall) createTempNpmrc() error {
	log.Debug("Creating project .npmrc file.")
	data, err := npm.GetConfigList(npmi.cliConfig.NpmArgs, npmi.executablePath)
	configData, err := npmi.prepareConfigData(data)
	if err != nil {
		return errorutils.CheckError(err)
	}

	if err = removeNpmrcIfExists(npmi.workingDirectory); err != nil {
		return err
	}

	return errorutils.CheckError(ioutil.WriteFile(filepath.Join(npmi.workingDirectory, npmrcFileName), configData, npmi.npmrcFileMode))
}

func (npmi *npmInstall) runInstall() error {
	log.Debug("Running npmi install command.")
	splitArgs, err := shellwords.Parse(npmi.cliConfig.NpmArgs)
	if err != nil {
		return errorutils.CheckError(err)
	}
	filteredArgs := filterFlags(splitArgs)
	installCmdConfig := &npm.NpmConfig{
		Npm:          npmi.executablePath,
		Command:      append([]string{"install"}, filteredArgs...),
		CommandFlags: nil,
		StrWriter:    nil,
		ErrWriter:    nil,
	}

	if npmi.collectBuildInfo && len(filteredArgs) > 0 {
		log.Warn("Build info dependencies collection with npm arguments is not supported. Build info creation will be skipped.")
		npmi.collectBuildInfo = false
	}

	return errorutils.CheckError(gofrogcmd.RunCmd(installCmdConfig))
}

func (npmi *npmInstall) setDependenciesList() (err error) {
	npmi.dependencies = make(map[string]*dependency)
	// npmi.scope can be empty, "production" or "development" in case of empty both of the functions should run
	if npmi.typeRestriction != "production" {
		if err = npmi.prepareDependencies("development"); err != nil {
			return
		}
	}
	if npmi.typeRestriction != "development" {
		err = npmi.prepareDependencies("production")
	}
	return
}

func (npmi *npmInstall) collectDependenciesChecksums() error {
	log.Info("Collecting dependencies information... This may take a few minuets...")
	servicesManager, err := utils.CreateServiceManager(npmi.cliConfig.ArtDetails, false)
	if err != nil {
		return err
	}

	producerConsumer := parallel.NewBounedRunner(10, false)
	errorsQueue := serviceutils.NewErrorsQueue(1)
	handlerFunc := npmi.createGetDependencyInfoFunc(servicesManager)
	go func() {
		defer producerConsumer.Done()
		for i := range npmi.dependencies {
			producerConsumer.AddTaskWithError(handlerFunc(i), errorsQueue.AddError)
		}
	}()
	producerConsumer.Run()
	return errorsQueue.GetError()
}

func (npmi *npmInstall) saveDependenciesData() error {
	log.Debug("Saving install data.")
	dependencies, missingDependencies := npmi.transformDependencies()
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Dependencies = dependencies
		partial.ModuleId = npmi.packageInfo.BuildInfoModuleId()
	}

	if err := utils.SavePartialBuildInfo(npmi.cliConfig.BuildName, npmi.cliConfig.BuildNumber, populateFunc); err != nil {
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

func (npmi *npmInstall) validateNpmVersion() error {
	npmVersion, err := npm.Version(npmi.executablePath)
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
func (npmi *npmInstall) backupProjectNpmrc() error {
	fileInfo, err := os.Stat(filepath.Join(npmi.workingDirectory, npmrcFileName))
	if err != nil {
		if os.IsNotExist(err) {
			npmi.npmrcFileMode = 0644
			return nil
		}
		return errorutils.CheckError(err)
	}

	npmi.npmrcFileMode = fileInfo.Mode()
	src := filepath.Join(npmi.workingDirectory, npmrcFileName)
	dst := filepath.Join(npmi.workingDirectory, npmrcBackupFileName)
	if err = ioutils.CopyFile(src, dst, npmi.npmrcFileMode); err != nil {
		return err
	}
	log.Debug("Project .npmrc file backed up successfully to", filepath.Join(npmi.workingDirectory, npmrcBackupFileName))
	return nil
}

// This func transforms "npm config list --json" result to key=val list of values that can be set to .npmrc file.
// it filters any nil values key, changes registry and scope registries to Artifactory url and adds Artifactory authentication to the list
func (npmi *npmInstall) prepareConfigData(data []byte) ([]byte, error) {
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
			filteredConf = append(filteredConf, i, " = ", npmi.registry, "\n")
		}
		npmi.setTypeRestriction(i, collectedConfig[i])
	}
	filteredConf = append(filteredConf, "registry = ", npmi.registry, "\n")
	filteredConf = append(filteredConf, npmi.npmAuth)
	return []byte(strings.Join(filteredConf, "")), nil
}

// npm install type restriction can be set by "--production" or "-only={prod[uction]|dev[elopment]}" flags
func (npmi *npmInstall) setTypeRestriction(key string, val interface{}) {
	if key == "production" && val != nil && (val == true || val == "true") {
		npmi.typeRestriction = "production"
	} else if key == "only" && val != nil {
		if strings.Contains(val.(string), "prod") {
			npmi.typeRestriction = "production"
		} else if strings.Contains(val.(string), "dev") {
			npmi.typeRestriction = "development"
		}
	}
}

// Run npm list and parse the returned json
func (npmi *npmInstall) prepareDependencies(typeRestriction string) error {
	// Run npm list
	data, errData, err := npm.RunList(npmi.cliConfig.NpmArgs+" -only="+typeRestriction, npmi.executablePath)
	if err != nil {
		log.Warn("npm list command failed with error:", err.Error())
	}
	if len(errData) > 0 {
		log.Warn("Some errors occurred while collecting dependencies info:\n" + string(errData))
	}

	// Parse the dependencies json object
	return jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if string(key) == "dependencies" {
			err := npmi.parseDependencies(value, typeRestriction)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Parses npm dependencies recursively and adds the collected dependencies to npmi.dependencies
func (npmi *npmInstall) parseDependencies(data []byte, scope string) error {
	var transitiveDependencies [][]byte
	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		ver, _, _, err := jsonparser.Get(data, string(key), "version")
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			return errorutils.CheckError(err)
		} else if err == jsonparser.KeyPathNotFoundError {
			log.Warn(fmt.Sprintf("npm dependencies list contains the package '%s' without version information. The dependency will not be added to build-info.", string(key)))
		} else {
			npmi.appendDependency(key, ver, scope)
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
		err := npmi.parseDependencies(element, scope)
		if err != nil {
			return err
		}
	}
	return nil
}

func (npmi *npmInstall) appendDependency(key []byte, ver []byte, scope string) {
	dependencyKey := string(key) + "-" + string(ver)
	if npmi.dependencies[dependencyKey] == nil {
		npmi.dependencies[dependencyKey] = &dependency{name: string(key), version: string(ver), scopes: []string{scope}}
	} else if !scopeAlreadyExists(scope, npmi.dependencies[dependencyKey].scopes) {
		npmi.dependencies[dependencyKey].scopes = append(npmi.dependencies[dependencyKey].scopes, scope)
	}
}

// Creates a function that fetches dependency data from Artifactory. Can be applied from a producer-consumer mechanism
func (npmi *npmInstall) createGetDependencyInfoFunc(servicesManager *artifactory.ArtifactoryServicesManager) getDependencyInfoFunc {
	return func(dependencyIndex string) parallel.TaskFunc {
		return func(threadId int) error {
			name := npmi.dependencies[dependencyIndex].name
			ver := npmi.dependencies[dependencyIndex].version
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
			npmi.dependencies[dependencyIndex].artifactName = parsedResult.Results[0].Name
			npmi.dependencies[dependencyIndex].checksum =
				&buildinfo.Checksum{Sha1: parsedResult.Results[0].Actual_sha1, Md5: parsedResult.Results[0].Actual_md5}
			log.Debug(cliutils.GetLogMsgPrefix(threadId, false), "Found", parsedResult.Results[0].Name,
				"sha1:", parsedResult.Results[0].Actual_sha1,
				"md5", parsedResult.Results[0].Actual_md5)
			return nil
		}
	}
}

// Transforms the list of dependencies to buildinfo.Dependencies list and creates a list of dependencies that are missing in Artifactory.
func (npmi *npmInstall) transformDependencies() (dependencies []buildinfo.Dependency, missingDependencies []dependency) {
	for _, dependency := range npmi.dependencies {
		if dependency.artifactName != "" {
			dependencies = append(dependencies,
				buildinfo.Dependency{Id: dependency.artifactName, Scopes: dependency.scopes, Checksum: dependency.checksum})
		} else {
			missingDependencies = append(missingDependencies, *dependency)
		}
	}
	return
}

func (npmi *npmInstall) restoreNpmrcAndError(err error) error {
	if restoreErr := npmi.restoreNpmrc(); restoreErr != nil {
		return errors.New(fmt.Sprintf("Two errors occurred:\n %s\n %s", restoreErr.Error(), err.Error()))
	}
	return err
}

func (npmi *npmInstall) setArtifactoryAuth() error {
	authArtDetails, err := npmi.cliConfig.ArtDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	if authArtDetails.GetSshAuthHeaders() != nil {
		return errorutils.CheckError(errors.New("SSH authentication is not supported in this command."))
	}
	npmi.artDetails = authArtDetails
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

func (npmi *npmInstall) setNpmExecutable() error {
	npmExecPath, err := exec.LookPath("npm")
	if err != nil {
		return errorutils.CheckError(err)
	}

	if npmExecPath == "" {
		return errorutils.CheckError(errors.New("Could not find 'npm' executable"))
	}
	npmi.executablePath = npmExecPath
	log.Debug("Found npm executable at:", npmi.executablePath)
	return nil
}

func getArtifactoryDetails(artDetails auth.ArtifactoryDetails) (body []byte, artifactoryVersion string, err error) {
	authApiUrl := artDetails.GetUrl() + "api/npm/auth"
	log.Debug("Sending npm auth request")
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return nil, "", err
	}

	resp, body, _, err := client.SendGet(authApiUrl, true, artDetails.CreateHttpClientDetails())
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	serverValues := strings.Split(resp.Header.Get("Server"), "/")
	if len(serverValues) != 2 {
		errorutils.CheckError(errors.New("Cannot parse Artifactory version from the server header."))
	}
	return body, strings.TrimSpace(serverValues[1]), err
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
	executablePath   string
	cliConfig        *npm.CliConfiguration
	npmrcFileMode    os.FileMode
	workingDirectory string
	registry         string
	npmAuth          string
	collectBuildInfo bool
	dependencies     map[string]*dependency
	typeRestriction  string
	artDetails       auth.ArtifactoryDetails
	packageInfo      *npm.PackageInfo
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

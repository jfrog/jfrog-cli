package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	utils2 "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/yarn"

	buildutils "github.com/jfrog/build-info-go/build/utils"
	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/gofrog/version"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	clientutils "github.com/jfrog/jfrog-client-go/utils"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-core/v2/utils/ioutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

const (
	minimumWorkspacesNpmVersion = "7.24.2"
)

type npmTestParams struct {
	testName      string
	nativeCommand string
	// Deprecated
	legacyCommand  string
	repo           string
	npmArgs        string
	wd             string
	buildNumber    string
	moduleName     string
	validationFunc func(*testing.T, npmTestParams, bool)
}

func cleanNpmTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	deleteSpec := spec.NewBuilder().Pattern(tests.NpmRepo).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.NoError(t, err)
	tests.CleanFileSystem()
}

func TestNpmNativeSyntax(t *testing.T) {
	testNpm(t, false)
}

// Deprecated
func TestNpmLegacy(t *testing.T) {
	testNpm(t, true)
}

func testNpm(t *testing.T, isLegacy bool) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	npmVersion, _, err := buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	log.Info("npm version:", npmVersion.GetVersion())
	isNpm7 := isNpm7(npmVersion)

	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath, npmProjectCi, npmPostInstallProjectPath := initNpmFilesTest(t)
	var npmTests = []npmTestParams{
		{testName: "npm ci", nativeCommand: "npm ci", legacyCommand: "rt npmci", repo: tests.NpmRemoteRepo, wd: npmProjectCi, validationFunc: validateNpmInstall},
		{testName: "npm ci with module", nativeCommand: "npm ci", legacyCommand: "rt npmci", repo: tests.NpmRemoteRepo, wd: npmProjectCi, moduleName: ModuleNameJFrogTest, validationFunc: validateNpmInstall},
		{testName: "npm i with module", nativeCommand: "npm install", legacyCommand: "rt npm-install", repo: tests.NpmRemoteRepo, wd: npmProjectPath, moduleName: ModuleNameJFrogTest, validationFunc: validateNpmInstall},
		{testName: "npm i with scoped project", nativeCommand: "npm install", legacyCommand: "rt npm-install", repo: tests.NpmRemoteRepo, wd: npmScopedProjectPath, validationFunc: validateNpmInstall},
		{testName: "npm i with npmrc project", nativeCommand: "npm install", legacyCommand: "rt npm-install", repo: tests.NpmRemoteRepo, wd: npmNpmrcProjectPath, validationFunc: validateNpmInstall},
		{testName: "npm i with production", nativeCommand: "npm install", legacyCommand: "rt npm-install", repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateNpmInstall, npmArgs: "--production"},
		{testName: "npm p with module", nativeCommand: "npm p", legacyCommand: "rt npmp", repo: tests.NpmRepo, wd: npmScopedProjectPath, moduleName: ModuleNameJFrogTest, validationFunc: validateNpmScopedPublish},
		{testName: "npm p", nativeCommand: "npm publish", legacyCommand: "rt npm-publish", repo: tests.NpmRepo, wd: npmProjectPath, validationFunc: validateNpmPublish},
		{testName: "npm postinstall", nativeCommand: "npm i", legacyCommand: "rt npmi", repo: tests.NpmRemoteRepo, wd: npmPostInstallProjectPath, validationFunc: validateNpmInstall},
	}

	for i, npmTest := range npmTests {
		t.Run(npmTest.testName, func(t *testing.T) {
			npmCmd := npmTest.nativeCommand
			if isLegacy {
				npmCmd = npmTest.legacyCommand
			}
			clientTestUtils.ChangeDirAndAssert(t, filepath.Dir(npmTest.wd))
			npmrcFileInfo, err := os.Stat(".npmrc")
			if err != nil && !os.IsNotExist(err) {
				assert.Fail(t, err.Error())
			}
			var buildNumber string
			commandArgs := strings.Split(npmCmd, " ")
			buildNumber = strconv.Itoa(i + 100)
			commandArgs = append(commandArgs, npmTest.npmArgs)

			// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
			commandArgs = append(commandArgs, "--cache="+tempCacheDirPath)

			commandArgs = append(commandArgs, "--build-name="+tests.NpmBuildName, "--build-number="+buildNumber)

			if npmTest.moduleName != "" {
				runJfrogCli(t, append(commandArgs, "--module="+npmTest.moduleName)...)
			} else {
				npmTest.moduleName = readModuleId(t, npmTest.wd, npmVersion)
				runJfrogCli(t, commandArgs...)
			}
			validateNpmLocalBuildInfo(t, tests.NpmBuildName, buildNumber, npmTest.moduleName)
			assert.NoError(t, artifactoryCli.Exec("bp", tests.NpmBuildName, buildNumber))
			npmTest.buildNumber = buildNumber
			npmTest.validationFunc(t, npmTest, isNpm7)

			// make sure npmrc file was not changed (if existed)
			postTestFileInfo, postTestFileInfoErr := os.Stat(".npmrc")
			validateNpmrcFileInfo(t, npmTest, npmrcFileInfo, postTestFileInfo, err, postTestFileInfoErr)
		})
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.NpmBuildName, artHttpDetails)
}

func TestNpmPublishWithNpmrc(t *testing.T) {
	testNpmPublishWithNpmrc(t, validateNpmPublish, "npmpublishrcproject", tests.NpmRepo, false)
}

func TestNpmPublishWithNpmrcScoped(t *testing.T) {
	testNpmPublishWithNpmrc(t, validateNpmScopedPublish, "npmpublishrcscopedproject", tests.NpmScopedRepo, true)
}

func testNpmPublishWithNpmrc(t *testing.T, validationFunc func(t *testing.T, npmTest npmTestParams, isNpm7 bool), projectName string, repoName string, isScoped bool) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	buildNumber := "1"
	npmVersion, _, err := buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Create project without a JFrog config file — native mode must not require it.
	// npm publish does not require node_modules, so no prepareArtifactoryForNpmBuild needed.
	npmProjectPath := filepath.Dir(createNpmProject(t, projectName))
	clientTestUtils.ChangeDirAndAssert(t, npmProjectPath)

	// fetch module id
	packageJsonPath := filepath.Join(npmProjectPath, "package.json")
	moduleName := readModuleId(t, packageJsonPath, npmVersion)

	// Write .npmrc directly — no npm.yaml config file.
	assert.NoError(t, writeNpmrcForNativeMode(t, tests.NpmRepo))

	if isScoped {
		addNpmScopeRegistryToNpmRc(t, npmProjectPath, packageJsonPath, npmVersion)
	}

	// Use deprecated --run-native flag to exercise backward-compat path through the CLI layer.
	runJfrogCli(t, "npm", "publish", "--run-native=true", "--build-name="+tests.NpmBuildName, "--build-number="+buildNumber)

	validateNpmLocalBuildInfo(t, tests.NpmBuildName, buildNumber, moduleName)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.NpmBuildName, buildNumber))

	// validation
	testParams := npmTestParams{testName: "npm p",
		nativeCommand:  "npm publish",
		legacyCommand:  "rt npm-publish",
		repo:           repoName,
		wd:             npmProjectPath,
		validationFunc: validateNpmPublish,
		buildNumber:    buildNumber,
		moduleName:     moduleName,
	}
	validationFunc(t, testParams, false)
}

// TestNpmInstallClientNative verifies that in native mode the user's .npmrc is
// never modified — no backup/restore, no temp file, mod-time unchanged.
// The test intentionally has no JFrog config file so it exercises the no-config
// native path introduced alongside JFROG_RUN_NATIVE support.
func TestNpmInstallClientNative(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	defer enableNativeMode(t)()

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	npmVersion, _, err := buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	buildNumber := "1"

	// Create project without a JFrog config file — native mode must not require it.
	npmProjectDirectory := filepath.Dir(createNpmProject(t, "npmproject"))
	prepareArtifactoryForNpmBuild(t, npmProjectDirectory)
	clientTestUtils.ChangeDirAndAssert(t, npmProjectDirectory)

	// Write .npmrc directly; this is the file native mode must leave untouched.
	assert.NoError(t, writeNpmrcForNativeMode(t, tests.NpmRemoteRepo))
	npmrcFileInfo, err := os.Stat(".npmrc")
	assert.NoError(t, err)

	packageJsonPath := filepath.Join(npmProjectDirectory, "package.json")
	moduleName := readModuleId(t, packageJsonPath, npmVersion)

	runJfrogCli(t, "npm", "i", "--build-name="+tests.NpmBuildName, "--build-number="+buildNumber)
	validateNpmLocalBuildInfo(t, tests.NpmBuildName, buildNumber, moduleName)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.NpmBuildName, buildNumber))

	npmTest := npmTestParams{
		testName:    "npm with run-native (JFROG_RUN_NATIVE=true)",
		buildNumber: buildNumber,
	}

	validateNpmInstall(t, npmTest, isNpm7(npmVersion))
	postTestFileInfo, postTestFileInfoErr := os.Stat(".npmrc")
	validateNpmrcFileInfo(t, npmTest, npmrcFileInfo, postTestFileInfo, err, postTestFileInfoErr)
	validateIfFileWasEverModified(t, npmrcFileInfo, postTestFileInfo)
}

// writeNpmrcForNativeMode creates a .npmrc in the current directory pointing to
// the given Artifactory npm repo, using serverDetails for auth.
// This simulates a project configured for Artifactory without a JFrog CLI config file.
func writeNpmrcForNativeMode(t *testing.T, repo string) error {
	repoURL := utils2.GetNpmRepositoryUrl(repo, serverDetails.ArtifactoryUrl)
	// Ensure trailing slash so GetNpmAuthKeyValue produces "//host/path/:_auth",
	// matching what npm's nerfDart() generates when looking up credentials.
	if !strings.HasSuffix(repoURL, "/") {
		repoURL += "/"
	}
	key, value := utils2.GetNpmAuthKeyValue(serverDetails, repoURL)
	require.NotEmpty(t, key, "no auth credentials available in serverDetails")
	content := fmt.Sprintf("registry = %s\n%s=%s\n", repoURL, key, value)
	return os.WriteFile(".npmrc", []byte(content), 0644)
}

// appendRegistryAuthToNpmrc appends auth credentials for registryURL to the current .npmrc.
// If registryURL is empty, reads it from the "registry = " line in the current .npmrc.
// Handles both access token and user/password auth. Ensures trailing slash in the registry
// URL so the nerf-dart key matches what npm's nerfDart() generates during credential lookup.
func appendRegistryAuthToNpmrc(t *testing.T, registryURL string) error {
	if registryURL == "" {
		data, err := os.ReadFile(".npmrc")
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "registry = ") {
				registryURL = strings.TrimSpace(strings.TrimPrefix(line, "registry = "))
				break
			}
		}
	}
	if registryURL == "" {
		return nil
	}
	// Ensure trailing slash so the nerf-dart key matches npm's nerfDart() output.
	if !strings.HasSuffix(registryURL, "/") {
		registryURL += "/"
	}
	key, value := utils2.GetNpmAuthKeyValue(serverDetails, registryURL)
	if key == "" {
		return nil
	}
	f, err := os.OpenFile(".npmrc", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		assert.NoError(t, f.Close())
	}()
	_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	return err
}

func readModuleId(t *testing.T, wd string, npmVersion *version.Version) string {
	packageInfo, err := buildutils.ReadPackageInfoFromPackageJsonIfExists(filepath.Dir(wd), npmVersion)
	assert.NoError(t, err)
	return packageInfo.BuildInfoModuleId()
}

func addNpmScopeRegistryToNpmRc(t *testing.T, projectPath string, packageJsonPath string, npmVersion *version.Version) {
	scope := getScopeFromPackageJson(t, packageJsonPath, npmVersion)
	authConfig, err := serverDetails.CreateArtAuthConfig()
	assert.NoError(t, err)
	_, registry, err := utils2.GetArtifactoryNpmRepoDetails(tests.NpmScopedRepo, authConfig, false)
	assert.NoError(t, err)
	// Ensure trailing slash so npm's nerfDart() finds the matching auth key.
	if !strings.HasSuffix(registry, "/") {
		registry += "/"
	}
	scopedRegistry := scope + ":registry=" + registry
	npmrcFilePath := filepath.Join(projectPath, ".npmrc")
	func() {
		npmrcFile, err := os.OpenFile(npmrcFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		assert.NoError(t, err)
		defer func() {
			_ = npmrcFile.Close()
		}()
		_, err = npmrcFile.WriteString(scopedRegistry + "\n")
		assert.NoError(t, err)
	}()

	assert.NoError(t, appendRegistryAuthToNpmrc(t, registry))
}

func getScopeFromPackageJson(t *testing.T, wd string, npmVersion *version.Version) string {
	packageInfo, err := buildutils.ReadPackageInfoFromPackageJsonIfExists(filepath.Dir(wd), npmVersion)
	assert.NoError(t, err)
	return packageInfo.Scope
}

func TestNpmWithGlobalConfig(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	npmProjectPath := initGlobalNpmFilesTest(t)
	clientTestUtils.ChangeDirAndAssert(t, filepath.Dir(npmProjectPath))
	runJfrogCli(t, "npm", "install", "--build-name="+tests.NpmBuildName, "--build-number=1", "--module="+ModuleNameJFrogTest)
	validateNpmLocalBuildInfo(t, tests.NpmBuildName, "1", ModuleNameJFrogTest)
}

func validateNpmLocalBuildInfo(t *testing.T, buildName, buildNumber, moduleName string) {
	buildInfoService := build.CreateBuildInfoService()
	npmBuild, err := buildInfoService.GetOrCreateBuildWithProject(buildName, buildNumber, "")
	assert.NoError(t, err)
	bi, err := npmBuild.ToBuildInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, bi.Started)
	if assert.Len(t, bi.Modules, 1) {
		assert.Equal(t, moduleName, bi.Modules[0].Id)
		assert.Equal(t, buildinfo.Npm, bi.Modules[0].Type)
	}
}

func TestNpmWithoutPackageJson(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	// Create temp dir that does not contain an npm project
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tempDirPath)
	defer chdirCallback()

	// Run config to allow resolution from Artifactory
	err = createConfigFileForTest([]string{tempDirPath}, tests.NpmRemoteRepo, "", t, project.Npm, false)
	assert.NoError(t, err)

	// Run npm install and make sure that package.json and package-lock.json were created
	runJfrogCli(t, "npm", "i", "json@9.0.6", "--save-exact")
	assert.FileExists(t, filepath.Join(tempDirPath, "package.json"))
	assert.FileExists(t, filepath.Join(tempDirPath, "package-lock.json"))
}

func TestNpmConditionalUpload(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	searchSpec, err := tests.CreateSpec(tests.SearchAllNpm)
	assert.NoError(t, err)
	npmVersion, _, err := buildutils.GetNpmVersionAndExecPath(log.Logger)
	assert.NoError(t, err)
	npmProjectPath := initNpmProjectTest(t)
	clientTestUtils.ChangeDirAndAssert(t, npmProjectPath)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	buildName := tests.NpmBuildName + "-scan"
	buildNumber := "505"
	runJfrogCli(t, []string{"npm", "install", "--build-name=" + buildName, "--build-number=" + buildNumber}...)
	execFunc := func() error {
		return runNpmConditionalUploadTest(buildName, buildNumber)
	}
	testConditionalUpload(t, execFunc, searchSpec, tests.GetNpmDeployedArtifacts(isNpm7(npmVersion))...)
}

func runNpmConditionalUploadTest(buildName, buildNumber string) (err error) {
	configFilePath, exists, err := project.GetProjectConfFilePath(project.Npm)
	if err != nil {
		return
	} else if !exists {
		return errorutils.CheckErrorf("no config file was found!")
	}
	npmCmd := npm.NewNpmPublishCommand()
	npmCmd.SetConfigFilePath(configFilePath).SetArgs([]string{"--scan", "--build-name=" + buildName, "--build-number=" + buildNumber})
	if err = npmCmd.Init(); err != nil {
		return err
	}
	printDeploymentView, detailedSummary := log.IsStdErrTerminal(), npmCmd.IsDetailedSummary()
	if !detailedSummary {
		npmCmd.SetDetailedSummary(printDeploymentView)
	}
	err = commands.Exec(npmCmd)
	result := npmCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(npmCmd.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func validateNpmrcFileInfo(t *testing.T, npmTest npmTestParams, npmrcFileInfo, postTestNpmrcFileInfo os.FileInfo, err, postTestFileInfoErr error) {
	if postTestFileInfoErr != nil && !os.IsNotExist(postTestFileInfoErr) {
		assert.Fail(t, postTestFileInfoErr.Error())
	}
	assert.False(t, err == nil && postTestFileInfoErr != nil, ".npmrc file existed and was not restored at the end of the install command.")
	assert.False(t, err != nil && postTestFileInfoErr == nil, ".npmrc file was not deleted at the end of the install command.")
	assert.False(t, err == nil && postTestFileInfoErr == nil && (npmrcFileInfo.Mode() != postTestNpmrcFileInfo.Mode() || npmrcFileInfo.Size() != postTestNpmrcFileInfo.Size()),
		".npmrc file was changed after running npm command! it was:\n%v\nnow it is:\n%v\nTest arguments are:\n%v", npmrcFileInfo, postTestNpmrcFileInfo, npmTest)
	// make sue the temp .npmrc was deleted.
	bcpNpmrc, err := os.Stat("jfrog.npmrc.backup")
	if err != nil && !os.IsNotExist(err) {
		assert.Fail(t, err.Error())
	}
	assert.Nil(t, bcpNpmrc, "The file 'jfrog.npmrc.backup' was supposed to be deleted but it was not when running the configuration:\n%v", npmTest)
}

// if file was backed up then it's mod time should be changed
func validateIfFileWasEverModified(t *testing.T, fileInfo, postTestFileInfo os.FileInfo) {
	assert.Equal(t, fileInfo.ModTime(), postTestFileInfo.ModTime())
}

func initNpmFilesTest(t *testing.T) (npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath, npmProjectCi, npmPostInstallProjectPath string) {
	npmProjectPath = createNpmProject(t, "npmproject")
	npmScopedProjectPath = createNpmProject(t, "npmscopedproject")
	npmNpmrcProjectPath = createNpmProject(t, "npmnpmrcproject")
	npmProjectCi = createNpmProject(t, "npmprojectci")
	npmPostInstallProjectPath = createNpmProject(t, "npmpostinstall")
	_ = createNpmProject(t, filepath.Join("npmpostinstall", "subdir"))
	err := createConfigFileForTest([]string{filepath.Dir(npmProjectPath), filepath.Dir(npmScopedProjectPath),
		filepath.Dir(npmNpmrcProjectPath), filepath.Dir(npmProjectCi), filepath.Dir(npmPostInstallProjectPath)}, tests.NpmRemoteRepo, tests.NpmRepo, t, project.Npm, false)
	assert.NoError(t, err)
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectPath))
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectCi))
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmPostInstallProjectPath))
	return
}

func initNpmProjectTest(t *testing.T) (npmProjectPath string) {
	npmProjectPath = filepath.Dir(createNpmProject(t, "npmproject"))
	err := createConfigFileForTest([]string{npmProjectPath}, tests.NpmRemoteRepo, tests.NpmRepo, t, project.Npm, false)
	assert.NoError(t, err)
	prepareArtifactoryForNpmBuild(t, npmProjectPath)
	return
}

func initNpmWorkspacesProjectTest(t *testing.T) (npmProjectPath string) {
	npmProjectPath = filepath.Dir(createNpmProject(t, "npmworkspaces"))
	err := createConfigFileForTest([]string{npmProjectPath}, tests.NpmRemoteRepo, tests.NpmRepo, t, project.Npm, false)
	assert.NoError(t, err)
	testFolder := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "npm", "npmworkspaces")
	err = biutils.CopyDir(testFolder, npmProjectPath, true, []string{})
	assert.NoError(t, err)
	prepareArtifactoryForNpmBuild(t, npmProjectPath)
	return
}

func initGlobalNpmFilesTest(t *testing.T) (npmProjectPath string) {
	npmProjectPath = createNpmProject(t, "npmproject")
	jfrogHomeDir, err := coreutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	err = createConfigFileForTest([]string{jfrogHomeDir}, tests.NpmRemoteRepo, tests.NpmRepo, t, project.Npm, true)
	assert.NoError(t, err)
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectPath))
	return
}

func createNpmProject(t *testing.T, dir string) string {
	srcPackageJson := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "npm", dir, "package.json")
	targetPackageJson := filepath.Join(tests.Out, dir)
	packageJson, err := tests.ReplaceTemplateVariables(srcPackageJson, targetPackageJson)
	assert.NoError(t, err)

	// failure can be ignored
	npmrcExists, err := fileutils.IsFileExists(filepath.Join(filepath.Dir(srcPackageJson), ".npmrc"), false)
	assert.NoError(t, err)

	if npmrcExists {
		_, err = tests.ReplaceTemplateVariables(filepath.Join(filepath.Dir(srcPackageJson), ".npmrc"), targetPackageJson)
		assert.NoError(t, err)
	}
	packageJson, err = filepath.Abs(packageJson)
	assert.NoError(t, err)
	return packageJson
}

func validateNpmInstall(t *testing.T, npmTestParams npmTestParams, isNpm7 bool) {
	expectedDependencies := []expectedDependency{{id: "xml:1.0.1", scopes: []string{"prod"}}}
	if !strings.Contains(npmTestParams.npmArgs, "-only=prod") && !strings.Contains(npmTestParams.npmArgs, "-production") {
		expectedDependencies = append(expectedDependencies, expectedDependency{id: "json:9.0.6", scopes: []string{"dev"}})
	}
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.NpmBuildName, npmTestParams.buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	if buildInfo.Modules == nil {
		assert.NotNil(t, buildInfo.Modules)
		return
	}
	assert.NotEmpty(t, buildInfo.Modules)
	equalDependenciesSlices(t, expectedDependencies, buildInfo.Modules[0].Dependencies)
}

type expectedDependency struct {
	id     string
	scopes []string
}

func validateNpmPublish(t *testing.T, npmTestParams npmTestParams, isNpm7 bool) {
	verifyExistInArtifactoryByProps(tests.GetNpmDeployedArtifacts(isNpm7),
		tests.NpmRepo+"/*",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", tests.NpmBuildName, npmTestParams.buildNumber), t)
	validateNpmCommonPublish(t, npmTestParams, isNpm7, false)
}

func validateNpmScopedPublish(t *testing.T, npmTestParams npmTestParams, isNpm7 bool) {
	verifyExistInArtifactoryByProps(tests.GetNpmDeployedScopedArtifacts(npmTestParams.repo, isNpm7),
		npmTestParams.repo+"/*",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", tests.NpmBuildName, npmTestParams.buildNumber), t)
	validateNpmCommonPublish(t, npmTestParams, isNpm7, true)
}

func validateNpmCommonPublish(t *testing.T, npmTestParams npmTestParams, isNpm7, isScoped bool) {
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.NpmBuildName, npmTestParams.buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	expectedArtifactName := tests.GetNpmArtifactName(isNpm7, isScoped)
	if len(buildInfo.Modules) == 0 {
		// Case no module was created
		assert.Fail(t, "npm publish test failed", "params: \n%v \nexpected to have module with the following artifact: \n%v \nbut has no modules: \n%v",
			npmTestParams, expectedArtifactName, buildInfo)
		return
	}
	// The checksums are ignored when comparing the actual and the expected
	assert.Len(t, buildInfo.Modules[0].Artifacts, 1, "npm publish test with the arguments: \n%v \nexpected to have the following artifact: \n%v \nbut has: \n%v",
		npmTestParams, expectedArtifactName, buildInfo.Modules[0].Artifacts)
	assert.Equal(t, npmTestParams.moduleName, buildInfo.Modules[0].Id, "npm publish test with the arguments: \n%v \nexpected to have the following module name: \n%v \nbut has: \n%v",
		npmTestParams, npmTestParams.moduleName, buildInfo.Modules[0].Id)
	assert.Equal(t, expectedArtifactName, buildInfo.Modules[0].Artifacts[0].Name, "npm publish test with the arguments: \n%v \nexpected to have the following artifact: \n%v \nbut has: \n%v",
		npmTestParams, expectedArtifactName, buildInfo.Modules[0].Artifacts[0].Name)
}

func prepareArtifactoryForNpmBuild(t *testing.T, workingDirectory string) {
	clientTestUtils.ChangeDirAndAssert(t, workingDirectory)

	caches := ioutils.DoubleWinPathSeparator(filepath.Join(workingDirectory, "caches"))
	// Run install with -cache argument to download the artifacts from Artifactory
	// This done to be sure the artifacts exists in Artifactory
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("npm", "install", "-cache="+caches))

	clientTestUtils.RemoveAllAndAssert(t, filepath.Join(workingDirectory, "node_modules"))
	clientTestUtils.RemoveAllAndAssert(t, caches)
}

func initNpmTest(t *testing.T) {
	if !*tests.TestNpm {
		t.Skip("Skipping Npm test. To run Npm test add the '-test.npm=true' option.")
	}
	// Ensure JFROG_RUN_NATIVE is not set (clean state for non-native tests)
	_ = os.Unsetenv("JFROG_RUN_NATIVE")
	createJfrogHomeConfig(t, true)
}

// enableNativeMode sets JFROG_RUN_NATIVE=true and returns a cleanup function that unsets it.
func enableNativeMode(t *testing.T) func() {
	clientTestUtils.SetEnvAndAssert(t, "JFROG_RUN_NATIVE", "true")
	return func() {
		clientTestUtils.UnSetEnvAndAssert(t, "JFROG_RUN_NATIVE")
	}
}

func TestNpmPublishDetailedSummary(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	npmVersion, _, err := buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Init npm project & npmp command for testing
	npmProjectPath := initNpmProjectTest(t)
	configFilePath := filepath.Join(npmProjectPath, ".jfrog", "projects", "npm.yaml")
	args := []string{"--detailed-summary=true"}
	npmpCmd := npm.NewNpmPublishCommand()
	npmpCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	assert.NoError(t, npmpCmd.Init())
	err = commands.Exec(npmpCmd)
	assert.NoError(t, err)

	result := npmpCmd.Result()
	assert.NotNil(t, result)
	reader := result.Reader()
	readerGetErrorAndAssert(t, reader)
	defer readerCloseAndAssert(t, reader)
	// Read result
	var files []clientutils.FileTransferDetails
	for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
		files = append(files, *transferDetails)
	}
	if files == nil {
		assert.NotNil(t, files)
		return
	}

	// Verify deploy details
	tarballName := "jfrog-cli-tests-v1.0.0.tgz"
	// In npm under v7 prefix is removed.
	if npmVersion.Compare("7.0.0") > 0 {
		tarballName = "jfrog-cli-tests-1.0.0.tgz"
	}
	expectedSourcePath := filepath.Join(npmProjectPath, tarballName)
	expectedTargetPath := serverDetails.ArtifactoryUrl + tests.NpmRepo + "/jfrog-cli-tests/-/" + tarballName
	assert.Equal(t, expectedSourcePath, files[0].SourcePath, "Summary validation failed - unmatched SourcePath.")
	assert.Equal(t, expectedTargetPath, files[0].RtUrl+files[0].TargetPath, "Summary validation failed - unmatched TargetPath.")
	assert.Equal(t, 1, len(files), "Summary validation failed - only one archive should be deployed.")
	// Verify sha256 is valid (a string size 256 characters) and not an empty string.
	assert.Equal(t, 64, len(files[0].Sha256), "Summary validation failed - sha256 should be in size 64 digits.")
}

func TestNpmDistTag(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	npmPath := initNpmProjectTest(t)
	chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, npmPath)
	defer chdirCallBack()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Publish package with tag.
	tagP := "tag-from-publish"
	assert.NoError(t, jfrogCli.Exec("npm", "p", "--tag="+tagP))

	// Add tag using dist-tag add command.
	tagDt := "tag-from-dist-tag"
	assert.NoError(t, jfrogCli.Exec("npm", "dist-tag", "add", "jfrog-cli-tests@v1.0.0", tagDt))

	assertDistTagsExist(t, []string{tagP, tagDt, "latest"})
}

func assertDistTagsExist(t *testing.T, expectedTags []string) {
	searchSpecBuilder := spec.NewBuilder().Pattern(tests.NpmRepo + "/*jfrog-cli-tests*1.0.0.tgz").Recursive(true)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails)
	searchCmd.SetSpec(searchSpecBuilder.BuildSpec())

	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	readerGetErrorAndAssert(t, reader)
	defer readerCloseAndAssert(t, reader)
	length, err := reader.Length()
	assert.NoError(t, err)
	if !assert.Equal(t, length, 1) {
		return
	}
	for resultItem := new(utils.SearchResult); reader.NextRecord(resultItem) == nil; resultItem = new(utils.SearchResult) {
		assert.ElementsMatch(t, resultItem.Props[npm.DistTagPropKey], expectedTags)
	}
}

func TestNpmPublishWithDeploymentView(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	npmPath := initNpmProjectTest(t)
	chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, npmPath)
	defer chdirCallBack()
	assertPrintedDeploymentViewFunc, cleanupFunc := initDeploymentViewTest(t)
	defer cleanupFunc()
	runGenericNpm(t, "npm", "publish")
	// Check deployment view
	assertPrintedDeploymentViewFunc()
}

func TestNpmPackInstall(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	command := "npm i"
	testWorkingDir, err := filepath.Abs(createNpmProject(t, "npmnpmrcproject"))
	assert.NoError(t, err)
	err = createConfigFileForTest([]string{filepath.Dir(testWorkingDir)}, tests.NpmRemoteRepo, tests.NpmRepo, t, project.Npm, false)
	assert.NoError(t, err)
	clientTestUtils.ChangeDirAndAssert(t, filepath.Dir(testWorkingDir))
	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	buildNumber := "999"
	commandArgs := strings.Split(command, " ")
	commandArgs = append(commandArgs, "yaml")

	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	commandArgs = append(commandArgs, "--cache="+tempCacheDirPath)

	commandArgs = append(commandArgs, "--build-name="+tests.NpmBuildName, "--build-number="+buildNumber)
	runJfrogCli(t, commandArgs...)

	// Validate that no dependencies were collected
	buildInfoService := build.CreateBuildInfoService()
	npmBuild, err := buildInfoService.GetOrCreateBuild(tests.NpmBuildName, buildNumber)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, npmBuild.Clean())
	}()
	npmBuildInfo, err := npmBuild.ToBuildInfo()
	assert.NoError(t, err)
	assert.NotNil(t, npmBuildInfo)
	assert.Len(t, npmBuildInfo.Modules, 0)
}

// Test npm publish --workspaces command
// When using the -w flag npm itself knows to handle multiple modules,
// And the CLI needs to know to publish multiple packages.
// Workspaces has been introduced in npm v7.0.0+
// Read more about npm workspaces here: https://docs.npmjs.com/cli/v7/using-npm/workspaces
func TestNpmPublishWithWorkspaces(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("JGC-417 - Test is flaky on Windows, skipping...")
	}
	// Check npm version
	npmVersion, _, err := buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	// In npm under v7 skip test
	if npmVersion.Compare(minimumWorkspacesNpmVersion) > 0 {
		log.Info("Test skipped as this function in not supported in npm version " + npmVersion.GetVersion())
		return
	}

	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	// Init npm project & npmp command for testing
	npmProjectPath := initNpmWorkspacesProjectTest(t)
	configFilePath := filepath.Join(npmProjectPath, ".jfrog", "projects", "npm.yaml")

	// Add build info parameters
	buildName := tests.NpmBuildName + "-workspaces"
	buildNumber := "789"
	args := []string{"--detailed-summary=true", "--workspaces", "--verbose",
		"--build-name=" + buildName, "--build-number=" + buildNumber}

	npmpCmd := npm.NewNpmPublishCommand()
	npmpCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	npmpCmd.SetNpmArgs(args)
	assert.NoError(t, npmpCmd.Init())
	err = commands.Exec(npmpCmd)
	assert.NoError(t, err)

	files := assertNpmPublishResultFiles(t, npmpCmd)

	expectedTars := []string{"nested1", "nested2"}
	for index, tar := range expectedTars {
		// Verify deploy details
		tarballName := tar + "-1.0.0.tgz"
		expectedSourcePath := filepath.Join(npmProjectPath, tarballName)
		expectedTargetPath := serverDetails.ArtifactoryUrl + tests.NpmRepo + "/" + tar + "/-/" + tarballName
		assert.Equal(t, expectedSourcePath, files[index].SourcePath, "Summary validation failed - unmatched SourcePath.")
		assert.Equal(t, expectedTargetPath, files[index].RtUrl+files[index].TargetPath, "Summary validation failed - unmatched TargetPath.")
		assert.Equal(t, len(expectedTars), len(files), "Summary validation failed - two archive should be deployed.")
		assert.Len(t, files[index].Sha256, 64)
	}

	// Validate build info was created
	buildInfoService := build.CreateBuildInfoService()
	npmBuild, err := buildInfoService.GetOrCreateBuild(buildName, buildNumber)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, npmBuild.Clean())
	}()

	npmBuildInfo, err := npmBuild.ToBuildInfo()
	assert.NoError(t, err)
	assert.NotNil(t, npmBuildInfo)
	assert.NotEmpty(t, npmBuildInfo.Started)

	// Should have multiple modules for workspaces (one per workspace package)
	assert.GreaterOrEqual(t, len(npmBuildInfo.Modules), 1, "There should be a single module created as part of workspaces publish")

	module := npmBuildInfo.Modules[0]
	assert.NotEmpty(t, module.Id, "Module %d should have an ID")
	assert.Equal(t, buildinfo.Npm, module.Type, "Module %d should be npm type")
	assert.Equal(t, len(module.Artifacts), 2, "Module %d should have artifacts")

	// Validate artifact properties
	for j, artifact := range module.Artifacts {
		assert.NotEmpty(t, artifact.Name, "Artifact %d in module %d should have a name", j)
		assert.NotEmpty(t, artifact.Path, "Artifact %d in module %d should have a path", j)
		assert.NotEmpty(t, artifact.Sha1, "Artifact %d in module %d should have SHA1", j)
		assert.NotEmpty(t, artifact.Sha256, "Artifact %d in module %d should have SHA256", j)
		assert.NotEmpty(t, artifact.Md5, "Artifact %d in module %d should have MD5", j)
		assert.True(t, containsTarName(artifact.Name, expectedTars))
	}

	// Publish build info to Artifactory
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func TestNpmPublishWithWorkspacesRunNative(t *testing.T) {
	if coreutils.IsWindows() {
		t.Skip("JGC-417 - Test is flaky on Windows, skipping...")
	}
	// Check npm version
	npmVersion, _, err := buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	// In npm under v7 skip test
	if npmVersion.Compare(minimumWorkspacesNpmVersion) > 0 {
		log.Info("Test skipped as this function in not supported in npm version " + npmVersion.GetVersion())
		return
	}

	initNpmTest(t)
	defer cleanNpmTest(t)
	defer enableNativeMode(t)()

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	// Init npm workspaces project without a JFrog config file — native mode must not require it.
	npmProjectPath := filepath.Dir(createNpmProject(t, "npmworkspaces"))
	testFolder := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "npm", "npmworkspaces")
	assert.NoError(t, biutils.CopyDir(testFolder, npmProjectPath, true, []string{}))
	prepareArtifactoryForNpmBuild(t, npmProjectPath)
	clientTestUtils.ChangeDirAndAssert(t, npmProjectPath)

	// Write .npmrc directly — no npm.yaml config file.
	assert.NoError(t, writeNpmrcForNativeMode(t, tests.NpmRepo))

	// Add build info parameters
	buildName := tests.NpmBuildName + "-workspaces-native"
	buildNumber := "890"
	args := []string{"--workspaces", "--build-name=" + buildName, "--build-number=" + buildNumber}

	npmpCmd := npm.NewNpmPublishCommand()
	npmpCmd.SetArgs(args).SetUseNative(true)
	assert.NoError(t, npmpCmd.Init())
	err = commands.Exec(npmpCmd)
	assert.NoError(t, err)

	expectedTars := []string{"nested1", "nested2"}
	// Validate build info was created
	buildInfoService := build.CreateBuildInfoService()
	npmBuild, err := buildInfoService.GetOrCreateBuild(buildName, buildNumber)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, npmBuild.Clean())
	}()

	npmBuildInfo, err := npmBuild.ToBuildInfo()
	assert.NoError(t, err)
	assert.NotNil(t, npmBuildInfo)
	assert.NotEmpty(t, npmBuildInfo.Started)

	// Should have single module with multiple artifacts for workspaces with run-native
	if !assert.GreaterOrEqual(t, len(npmBuildInfo.Modules), 1, "There should be a single module created as part of workspaces publish with run-native") {
		return
	}

	module := npmBuildInfo.Modules[0]
	assert.NotEmpty(t, module.Id, "Module should have an ID")
	assert.Equal(t, buildinfo.Npm, module.Type, "Module should be npm type")
	assert.Equal(t, len(module.Artifacts), 2, "Module should have exactly 2 artifacts for workspaces")

	// Validate artifact properties
	for j, artifact := range module.Artifacts {
		assert.NotEmpty(t, artifact.Name, "Artifact %d should have a name", j)
		assert.NotEmpty(t, artifact.Path, "Artifact %d should have a path", j)
		assert.NotEmpty(t, artifact.Sha1, "Artifact %d should have SHA1", j)
		assert.NotEmpty(t, artifact.Sha256, "Artifact %d should have SHA256", j)
		assert.NotEmpty(t, artifact.Md5, "Artifact %d should have MD5", j)
		assert.True(t, containsTarName(artifact.Name, expectedTars))
	}

	// Publish build info to Artifactory
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// Test npm publish command with provided tarball
func TestNpmPackProvidedTarball(t *testing.T) {
	// Check npm version
	npmVersion, _, err := buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	// In npm under v7 skip test
	if npmVersion.Compare(minimumWorkspacesNpmVersion) > 0 {
		log.Info("Test skipped as this function in not supported in npm version " + npmVersion.GetVersion())
		return
	}

	// Prepare test
	initNpmTest(t)
	defer cleanNpmTest(t)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	testFolder := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "npm", "npmprovidedtarball")
	err = biutils.CopyDir(testFolder, tempDirPath, false, []string{})
	assert.NoError(t, err)

	// CD inside the copied project and create npm config
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tempDirPath)
	defer chdirCallback()
	err = createConfigFileForTest([]string{tempDirPath}, tests.NpmRemoteRepo, tests.NpmRepo, t, project.Npm, false)
	assert.NoError(t, err)

	// Init npm project & npmp command for testing
	configFilePath := filepath.Join(tempDirPath, ".jfrog", "projects", "npm.yaml")
	args := []string{"jfrog-cli-tests-v1.0.0.tgz", "--detailed-summary=true", "--workspaces", "--verbose"}
	npmpCmd := npm.NewNpmPublishCommand()
	npmpCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	npmpCmd.SetNpmArgs(args)
	assert.NoError(t, npmpCmd.Init())
	err = commands.Exec(npmpCmd)
	assert.NoError(t, err)

	// Check result
	assertNpmPublishResultFiles(t, npmpCmd)
}

func TestYarn(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	yarnProjectPath := filepath.Join(testDataTarget, "yarnprojectV2")
	assert.NoError(t, createConfigFileForTest([]string{yarnProjectPath}, tests.NpmRemoteRepo, "", t, project.Yarn, false))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, yarnProjectPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	// Add "localhost" to http whitelist
	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)
	// Get original http white list config
	origWhitelist, err := yarn.ConfigGet("unsafeHttpWhitelist", yarnExecPath, true)
	assert.NoError(t, err)
	assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[\"localhost\"]", yarnExecPath, true))
	defer func() {
		// Restore original whitelist config
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", origWhitelist, yarnExecPath, true))
	}()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("yarn", "--build-name="+tests.YarnBuildName, "--build-number=1", "--module="+ModuleNameJFrogTest))

	validateNpmLocalBuildInfo(t, tests.YarnBuildName, "1", ModuleNameJFrogTest)

	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("bp", tests.YarnBuildName, "1"))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.YarnBuildName, "1")
	assert.NoError(t, err)
	assert.True(t, found)
	if assert.NotNil(t, publishedBuildInfo) && assert.NotNil(t, publishedBuildInfo.BuildInfo) {
		assert.Equal(t, 1, len(publishedBuildInfo.BuildInfo.Modules))
		if len(publishedBuildInfo.BuildInfo.Modules) > 0 {
			assert.Equal(t, buildinfo.Npm, publishedBuildInfo.BuildInfo.Modules[0].Type)
			assert.Equal(t, "jfrog-test", publishedBuildInfo.BuildInfo.Modules[0].Id)
			assert.Equal(t, 0, len(publishedBuildInfo.BuildInfo.Modules[0].Artifacts))

			expectedDependencies := []expectedDependency{{id: "xml:1.0.1"}, {id: "json:9.0.6"}}
			equalDependenciesSlices(t, expectedDependencies, publishedBuildInfo.BuildInfo.Modules[0].Dependencies)
		}
	}
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.YarnBuildName, artHttpDetails)
}

func TestYarnSetVersion(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	yarnProjectPath := filepath.Join(testDataTarget, "yarnprojectV2")
	assert.NoError(t, createConfigFileForTest([]string{yarnProjectPath}, tests.NpmRemoteRepo, "", t, project.Yarn, false))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, yarnProjectPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	// Add "localhost" to http whitelist
	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)
	// Get original http white list config
	origWhitelist, err := yarn.ConfigGet("unsafeHttpWhitelist", yarnExecPath, true)
	assert.NoError(t, err)
	assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[\"localhost\"]", yarnExecPath, true))
	defer func() {
		// Restore original whitelist config
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", origWhitelist, yarnExecPath, true))
	}()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("yarn", "set", "version", "3.2.1")
	assert.NoError(t, err)
	modifyExistingYarnRc(t, "3.2.1")
}

// TestYarnUpgradeToV5 verifies that upgrading to an unsupported yarn major version (v5+) is blocked.
func TestYarnUpgradeToV5(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	yarnProjectPath := filepath.Join(testDataTarget, "yarnprojectV2")
	assert.NoError(t, createConfigFileForTest([]string{yarnProjectPath}, tests.NpmRemoteRepo, "", t, project.Yarn, false))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, yarnProjectPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)
	// Get original http white list config
	origWhitelist, err := yarn.ConfigGet("unsafeHttpWhitelist", yarnExecPath, true)
	assert.NoError(t, err)
	assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[\"localhost\"]", yarnExecPath, true))
	defer func() {
		// Restore original whitelist config
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", origWhitelist, yarnExecPath, true))
	}()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// Yarn v5 is not yet supported — should fail
	err = jfrogCli.Exec("yarn", "set", "version", "5.0.0")
	assert.Error(t, err)
}

func TestYarnInV4(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")

	yarnProjectPath := filepath.Join(testDataTarget, "yarnprojectV4")
	assert.NoError(t, createConfigFileForTest([]string{yarnProjectPath}, tests.NpmRemoteRepo, "", t, project.Yarn, false))
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, yarnProjectPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	// Add "localhost" to http whitelist (CI Artifactory uses HTTP)
	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)
	origWhitelist, err := yarn.ConfigGet("unsafeHttpWhitelist", yarnExecPath, true)
	assert.NoError(t, err)
	assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[\"localhost\"]", yarnExecPath, true))
	defer func() {
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", origWhitelist, yarnExecPath, true))
	}()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// Yarn v4 is now supported — install should succeed and collect build-info
	assert.NoError(t, jfrogCli.Exec("yarn", "--build-name="+tests.YarnBuildName, "--build-number=2", "--module=yarnV4Module"))

	validateNpmLocalBuildInfo(t, tests.YarnBuildName, "2", "yarnV4Module")

	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("bp", tests.YarnBuildName, "2"))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.YarnBuildName, "2")
	assert.NoError(t, err)
	assert.True(t, found)
	if assert.NotNil(t, publishedBuildInfo) && assert.NotNil(t, publishedBuildInfo.BuildInfo) {
		assert.Equal(t, 1, len(publishedBuildInfo.BuildInfo.Modules))
		if len(publishedBuildInfo.BuildInfo.Modules) > 0 {
			assert.Equal(t, buildinfo.Npm, publishedBuildInfo.BuildInfo.Modules[0].Type)
			expectedDependencies := []expectedDependency{{id: "xml:1.0.1"}, {id: "json:9.0.6"}}
			equalDependenciesSlices(t, expectedDependencies, publishedBuildInfo.BuildInfo.Modules[0].Dependencies)
		}
	}
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.YarnBuildName, artHttpDetails)
}

// TestYarnV4WorkspaceRoot verifies that running `jf yarn install` from a monorepo workspace
// root correctly captures all dependencies declared by workspace member packages.
// Previously the build-info was empty because the root package has no direct dependency
// edges in `yarn info` output — the fix seeds additional walks from each workspace member.
func TestYarnV4WorkspaceRoot(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")

	yarnProjectPath := filepath.Join(testDataTarget, "yarnprojectV4workspace")
	assert.NoError(t, createConfigFileForTest([]string{yarnProjectPath}, tests.NpmRemoteRepo, "", t, project.Yarn, false))
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, yarnProjectPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)
	origWhitelist, err := yarn.ConfigGet("unsafeHttpWhitelist", yarnExecPath, true)
	assert.NoError(t, err)
	assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[\"localhost\"]", yarnExecPath, true))
	defer func() {
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", origWhitelist, yarnExecPath, true))
	}()

	buildName := tests.YarnBuildName + "-workspace-root"
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("yarn", "--build-name="+buildName, "--build-number=1", "--module=yarnV4WorkspaceModule"))

	validateNpmLocalBuildInfo(t, buildName, "1", "yarnV4WorkspaceModule")

	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("bp", buildName, "1"))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, "1")
	assert.NoError(t, err)
	assert.True(t, found)
	if assert.NotNil(t, publishedBuildInfo) && assert.NotNil(t, publishedBuildInfo.BuildInfo) {
		assert.Equal(t, 1, len(publishedBuildInfo.BuildInfo.Modules))
		if len(publishedBuildInfo.BuildInfo.Modules) > 0 {
			assert.Equal(t, buildinfo.Npm, publishedBuildInfo.BuildInfo.Modules[0].Type)
			// xml comes from pkg-a, json comes from pkg-b — both must be captured
			expectedDependencies := []expectedDependency{{id: "xml:1.0.1"}, {id: "json:9.0.6"}}
			equalDependenciesSlices(t, expectedDependencies, publishedBuildInfo.BuildInfo.Modules[0].Dependencies)
		}
	}
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestYarnV4WorkspaceMemberWithSibling runs `jf yarn install` from inside a workspace member
// (packages/pkg-a) that depends on a registry package (xml) and a workspace sibling (pkg-b
// via workspace:*). pkg-b itself depends on json.
//
// Verifies:
//   - xml:1.0.1 and json:9.0.6 appear in build-info
//   - pkg-b (workspace:* sibling) does NOT appear as a build-info dependency
//   - No misleading "could not be found in Artifactory" warning for pkg-b
func TestYarnV4WorkspaceMemberWithSibling(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")

	// The jfrog config file must be created at the workspace root so yarn can resolve all packages.
	workspaceRoot := filepath.Join(testDataTarget, "yarnprojectV4workspacemember")
	assert.NoError(t, createConfigFileForTest([]string{workspaceRoot}, tests.NpmRemoteRepo, "", t, project.Yarn, false))

	// Run from the workspace member directory (pkg-a), not the root.
	pkgAPath := filepath.Join(workspaceRoot, "packages", "pkg-a")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, pkgAPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)
	origWhitelist, err := yarn.ConfigGet("unsafeHttpWhitelist", yarnExecPath, true)
	assert.NoError(t, err)
	assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[\"localhost\"]", yarnExecPath, true))
	defer func() {
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", origWhitelist, yarnExecPath, true))
	}()

	buildName := tests.YarnBuildName + "-workspace-member"
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("yarn", "--build-name="+buildName, "--build-number=1", "--module=yarnV4WorkspaceMemberModule"))

	validateNpmLocalBuildInfo(t, buildName, "1", "yarnV4WorkspaceMemberModule")

	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("bp", buildName, "1"))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, "1")
	assert.NoError(t, err)
	assert.True(t, found)
	if assert.NotNil(t, publishedBuildInfo) && assert.NotNil(t, publishedBuildInfo.BuildInfo) {
		assert.Equal(t, 1, len(publishedBuildInfo.BuildInfo.Modules))
		if len(publishedBuildInfo.BuildInfo.Modules) > 0 {
			assert.Equal(t, buildinfo.Npm, publishedBuildInfo.BuildInfo.Modules[0].Type)
			deps := publishedBuildInfo.BuildInfo.Modules[0].Dependencies

			// Registry deps from pkg-a (direct) and pkg-b (via workspace sibling) must be captured.
			expectedDependencies := []expectedDependency{{id: "xml:1.0.1"}, {id: "json:9.0.6"}}
			equalDependenciesSlices(t, expectedDependencies, deps)

			// pkg-b is a workspace sibling — must not appear as a build-info dependency.
			for _, dep := range deps {
				assert.NotEqual(t, "pkg-b:1.0.0", dep.Id, "workspace sibling pkg-b must not appear as a build-info dependency")
			}
		}
	}
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestYarnChangeVersionInV4ToV5 verifies that upgrading from v4 to unsupported v5 is blocked,
// while downgrading to v3 is allowed.
func TestYarnChangeVersionInV4ToV5(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	yarnProjectPath := filepath.Join(testDataTarget, "yarnprojectV4")
	assert.NoError(t, createConfigFileForTest([]string{yarnProjectPath}, tests.NpmRemoteRepo, "", t, project.Yarn, false))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, yarnProjectPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)

	yarnrcPath := ".yarnrc.yml"
	data, err := os.ReadFile(yarnrcPath)
	assert.NoError(t, err)
	var config = make(map[string]any)
	err = yaml.Unmarshal(data, &config)
	assert.NoError(t, err)
	config["unsafeHttpWhitelist"] = []string{"localhost"}
	updatedYamlData, err := yaml.Marshal(&config)
	assert.NoError(t, err)
	err = os.WriteFile(yarnrcPath, updatedYamlData, 0644)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[]", yarnExecPath, true))
	}()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Upgrading to v5 should fail
	err = jfrogCli.Exec("yarn", "set", "version", "5.0.0")
	assert.Error(t, err)

	// Downgrading to v3 should succeed
	err = jfrogCli.Exec("yarn", "set", "version", "3.2.1")
	assert.NoError(t, err)
	modifyExistingYarnRc(t, "3.2.1")

	err = jfrogCli.Exec("yarn", "--version")
	assert.NoError(t, err)
}

// TestYarnV4NativeMode verifies that yarn v4 works in native mode (JFROG_RUN_NATIVE=true)
// without a jfrog config file, using the default server for build-info collection.
func TestYarnV4NativeMode(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	defer enableNativeMode(t)()

	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")

	yarnProjectPath := filepath.Join(testDataTarget, "yarnprojectV4")
	// No config file created — native mode should work without it
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, yarnProjectPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	// Add "localhost" to http whitelist (CI Artifactory uses HTTP)
	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)
	origWhitelist, err := yarn.ConfigGet("unsafeHttpWhitelist", yarnExecPath, true)
	assert.NoError(t, err)
	assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[\"localhost\"]", yarnExecPath, true))
	defer func() {
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", origWhitelist, yarnExecPath, true))
	}()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("yarn", "--build-name="+tests.YarnBuildName, "--build-number=3", "--module=yarnV4NativeModule"))

	validateNpmLocalBuildInfo(t, tests.YarnBuildName, "3", "yarnV4NativeModule")

	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("bp", tests.YarnBuildName, "3"))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.YarnBuildName, "3")
	assert.NoError(t, err)
	assert.True(t, found)
	if assert.NotNil(t, publishedBuildInfo) && assert.NotNil(t, publishedBuildInfo.BuildInfo) {
		assert.Equal(t, 1, len(publishedBuildInfo.BuildInfo.Modules))
		if len(publishedBuildInfo.BuildInfo.Modules) > 0 {
			assert.Equal(t, buildinfo.Npm, publishedBuildInfo.BuildInfo.Modules[0].Type)
			expectedDependencies := []expectedDependency{{id: "xml:1.0.1"}, {id: "json:9.0.6"}}
			equalDependenciesSlices(t, expectedDependencies, publishedBuildInfo.BuildInfo.Modules[0].Dependencies)
		}
	}
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.YarnBuildName, artHttpDetails)
}

// TestYarnV4NativeModeWithServerId verifies that yarn v4 works in native mode
// with an explicit --server-id flag for build-info collection.
func TestYarnV4NativeModeWithServerId(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	defer enableNativeMode(t)()

	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, biutils.CopyDir(testDataSource, testDataTarget, true, nil))

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")

	yarnProjectPath := filepath.Join(testDataTarget, "yarnprojectV4")
	// No config file created — native mode should work without it
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, yarnProjectPath)
	defer chdirCallback()
	cleanUpYarnGlobalFolder := clientTestUtils.SetEnvWithCallbackAndAssert(t, "YARN_GLOBAL_FOLDER", tempDirPath)
	defer cleanUpYarnGlobalFolder()

	// Add "localhost" to http whitelist
	yarnExecPath, err := exec.LookPath("yarn")
	assert.NoError(t, err)
	origWhitelist, err := yarn.ConfigGet("unsafeHttpWhitelist", yarnExecPath, true)
	assert.NoError(t, err)
	assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", "[\"localhost\"]", yarnExecPath, true))
	defer func() {
		assert.NoError(t, yarn.ConfigSet("unsafeHttpWhitelist", origWhitelist, yarnExecPath, true))
	}()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("yarn", "--build-name="+tests.YarnBuildName, "--build-number=4", "--module=yarnV4NativeModule", "--server-id=default"))

	validateNpmLocalBuildInfo(t, tests.YarnBuildName, "4", "yarnV4NativeModule")

	assert.NoError(t, artifactoryCli.WithoutCredentials().Exec("bp", tests.YarnBuildName, "4"))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.YarnBuildName, "4")
	assert.NoError(t, err)
	assert.True(t, found)
	if assert.NotNil(t, publishedBuildInfo) && assert.NotNil(t, publishedBuildInfo.BuildInfo) {
		assert.Equal(t, 1, len(publishedBuildInfo.BuildInfo.Modules))
		if len(publishedBuildInfo.BuildInfo.Modules) > 0 {
			assert.Equal(t, buildinfo.Npm, publishedBuildInfo.BuildInfo.Modules[0].Type)
			expectedDependencies := []expectedDependency{{id: "xml:1.0.1"}, {id: "json:9.0.6"}}
			equalDependenciesSlices(t, expectedDependencies, publishedBuildInfo.BuildInfo.Modules[0].Dependencies)
		}
	}
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.YarnBuildName, artHttpDetails)
}

// Checks if the expected dependencies match the actual dependencies. Only the dependencies' IDs and scopes (not more than one scope) are compared.
func equalDependenciesSlices(t *testing.T, expectedDependencies []expectedDependency, actualDependencies []buildinfo.Dependency) {
	assert.Equal(t, len(expectedDependencies), len(actualDependencies))
	for _, dependency := range expectedDependencies {
		found := false
		for _, actualDependency := range actualDependencies {
			if actualDependency.Id == dependency.id &&
				len(actualDependency.Scopes) == len(dependency.scopes) &&
				(len(actualDependency.Scopes) == 0 || actualDependency.Scopes[0] == dependency.scopes[0]) {
				found = true
				break
			}
		}
		// The checksums are ignored when comparing the actual and the expected
		assert.True(t, found, "The dependencies from the build-info did not match the expected. expected: %v, actual: %v",
			expectedDependencies, dependenciesToPrintableArray(actualDependencies))
	}
}

func modifyExistingYarnRc(t *testing.T, version string) {
	yarnConfig := make(map[string]any)
	yarnRcPath := ".yarnrc.yml"
	yarnConfig["yarnPath"] = ".yarn/releases/yarn-" + version + ".cjs"
	updatedYamlData, err := yaml.Marshal(&yarnConfig)
	assert.NoError(t, err)
	err = os.WriteFile(yarnRcPath, updatedYamlData, 0644)
	assert.NoError(t, err)
}

func isNpm7(npmVersion *version.Version) bool {
	return npmVersion.Compare("7.0.0") <= 0
}

func TestGenericNpm(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	npmPath := initNpmProjectTest(t)
	chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, npmPath)
	defer chdirCallBack()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{"npm", "version"}
	output := jfrogCli.WithoutCredentials().RunCliCmdWithOutput(t, args...)
	assert.Contains(t, output, "'jfrog-cli-tests': 'v1.0.0'")
	// Check we don't fail with JFrog flags.
	output = jfrogCli.WithoutCredentials().RunCliCmdWithOutput(t, append(args, "--build-name=d", "--build-number=1", "--module=1")...)
	assert.Contains(t, output, "'jfrog-cli-tests': 'v1.0.0'")
}

func runGenericNpm(t *testing.T, args ...string) {
	jfCli := coretests.NewJfrogCli(execMain, "jf", "")
	assert.NoError(t, jfCli.WithoutCredentials().Exec(args...))
}

func assertNpmPublishResultFiles(t *testing.T, npmpCmd *npm.NpmPublishCommand) (files []clientutils.FileTransferDetails) {
	result := npmpCmd.Result()
	assert.NotNil(t, result)
	reader := result.Reader()
	readerGetErrorAndAssert(t, reader)
	defer readerCloseAndAssert(t, reader)
	for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
		files = append(files, *transferDetails)
	}
	assert.NotNil(t, files)
	return files
}

func TestSetupNpmCommand(t *testing.T) {
	initNpmTest(t)
	// Validate that the module does not exist in the cache before running the test.
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	moduleCacheUrl := serverDetails.ArtifactoryUrl + tests.NpmRemoteRepo + "-cache/chalk-animation/-/chalk-animation-2.0.3.tgz"
	_, _, err = client.GetRemoteFileDetails(moduleCacheUrl, artHttpDetails)
	assert.ErrorContains(t, err, "404")

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	require.NoError(t, execGo(jfrogCli, "setup", "npm", "--repo="+tests.NpmRemoteRepo))

	// Run 'npm install' to resolve the module from Artifactory and force it to be downloaded from Artifactory.
	output, err := exec.Command("npm", "install", "chalk-animation@2.0.3", "--cache", t.TempDir(), "--prefix", t.TempDir()).Output()
	assert.NoError(t, err, fmt.Sprintf("%s\n%q", string(output), err))
	// Validate that the module exists in the cache after running the test.
	// That means that the setup command worked and the 'go get' resolved the module from Artifactory.
	_, res, err := client.GetRemoteFileDetails(moduleCacheUrl, artHttpDetails)
	if assert.NoError(t, err, "Failed to find the artifact in the cache: "+moduleCacheUrl) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
}

func containsTarName(tarName string, expectedTars []string) bool {
	isTarPresent := false
	for _, tar := range expectedTars {
		strings.Contains(tarName, tar)
		isTarPresent = true
		break
	}
	return isTarPresent
}

// TestNpmBuildPublishWithCIVcsProps tests that CI VCS properties are set on npm artifacts
// when running build-publish in a CI environment (GitHub Actions).
func TestNpmBuildPublishWithCIVcsProps(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	buildName := "npm-civcs-test"
	buildNumber := "1"

	// Setup GitHub Actions environment (uses real env vars on CI, mock values locally)
	cleanupEnv, actualOrg, actualRepo := tests.SetupGitHubActionsEnv(t)
	defer cleanupEnv()

	// Clean old build
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	// Setup npm project and change to that directory
	npmPath := initNpmProjectTest(t)
	chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, npmPath)
	defer chdirCallBack()

	// Run npm publish with build info collection
	runJfrogCli(t, "npm", "publish", "--build-name="+buildName, "--build-number="+buildNumber)

	// Publish build info - should set CI VCS props on artifacts
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Restore working directory before getting build info
	clientTestUtils.ChangeDirAndAssert(t, wd)

	// Get the published build info to find artifact paths
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "Build info was not found")

	// Create service manager for getting artifact properties
	serviceManager, err := utils.CreateServiceManager(serverDetails, 3, 1000, false)
	assert.NoError(t, err)

	// Verify VCS properties on each artifact from build info
	artifactCount := 0
	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			fullPath := artifact.OriginalDeploymentRepo + "/" + artifact.Path

			props, err := serviceManager.GetItemProps(fullPath)
			assert.NoError(t, err, "Failed to get properties for artifact: %s", fullPath)
			assert.NotNil(t, props, "Properties are nil for artifact: %s", fullPath)

			// Validate VCS properties
			assert.Contains(t, props.Properties, "vcs.provider", "Missing vcs.provider on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.provider"], "github", "Wrong vcs.provider on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.org", "Missing vcs.org on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.org"], actualOrg, "Wrong vcs.org on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.repo", "Missing vcs.repo on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.repo"], actualRepo, "Wrong vcs.repo on %s", artifact.Name)

			artifactCount++
		}
	}
	assert.Greater(t, artifactCount, 0, "No artifacts in build info")
}

// TestNpmPublishWithLocalGitVcsProps verifies local git VCS props on npm artifacts
// when running publish followed by build-publish with VCS collection enabled and no CI env.
func TestNpmPublishWithLocalGitVcsProps(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	buildName := "npm-local-git-test"
	buildNumber := "1"

	cleanupEnv := tests.SetupLocalGitVcsEnv(t)
	defer cleanupEnv()

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	wd, err := os.Getwd()
	require.NoError(t, err)

	npmPath := initNpmProjectTest(t)
	tests.CopyGitFixtureIntoProject(t, npmPath)
	chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, npmPath)
	defer chdirCallBack()

	runJfrogCli(t, "npm", "publish", "--build-name="+buildName, "--build-number="+buildNumber)
	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	clientTestUtils.ChangeDirAndAssert(t, wd)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)

	serviceManager, err := utils.CreateServiceManager(serverDetails, 3, 1000, false)
	require.NoError(t, err)

	count := tests.ValidateLocalGitVcsPropsOnBuildInfoArtifacts(t, serviceManager, publishedBuildInfo, tests.NpmRepo,
		tests.VcsFixtureMainURL, tests.VcsFixtureMainRevision, tests.VcsFixtureMainBranch)
	assert.Greater(t, count, 0)
}

// TestNpmFailOnMissingDeps - COMPREHENSIVE SUITE
// Tests all permutations and combinations of --fail-on-missing-deps flag
//
// SUCCESS PATHS (what we test end-to-end with real apmtest server):
// - Backward compatibility (no flag)
// - All individual flag values: all, peer, optional, regular, bundle
// - 8 permutations/combinations of 2+ flags
// - 3 semantic edge cases verifying exclusion logic
// Total: 15 subtests covering all realistic success scenarios
//
// FAILURE PATHS (tested in build-info-go unit tests):
// - TestHandleFailOnMissingDeps verifies all flag/depType combinations trigger correct failures
// - All 4 dependency types: peer, optional, regular, bundle
// - All flag combinations tested with proper mocking
// - See: build-info-go/build/utils/npm_test.go line 816+
func TestNpmFailOnMissingDeps(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	_, _, err = buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err, "npm must be available for this test")
		return
	}

	testCases := []struct {
		name            string
		flagValue       string
		buildName       string
		buildNumber     string
		expectedSuccess bool
		description     string
		category        string // "backward_compat", "individual", "combo", "semantic"
	}{
		// ===== 1. BACKWARD COMPATIBILITY =====
		{
			name:            "backward_compat_no_flag",
			flagValue:       "",
			buildName:       "npm-no-flag",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Without flag: warns but doesn't fail (backward compat preserved)",
			category:        "backward_compat",
		},

		// ===== 2. INDIVIDUAL FLAG VALUES (5 tests) =====
		{
			name:            "flag_all",
			flagValue:       "all",
			buildName:       "npm-flag-all",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Flag: all (monitors all 4 dep types)",
			category:        "individual",
		},
		{
			name:            "flag_peer",
			flagValue:       "peer",
			buildName:       "npm-flag-peer",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Flag: peer (peerDependencies only)",
			category:        "individual",
		},
		{
			name:            "flag_optional",
			flagValue:       "optional",
			buildName:       "npm-flag-optional",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Flag: optional (optionalDependencies only)",
			category:        "individual",
		},
		{
			name:            "flag_regular",
			flagValue:       "regular",
			buildName:       "npm-flag-regular",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Flag: regular (regular/dev/bundle, NOT optional)",
			category:        "individual",
		},
		{
			name:            "flag_bundle",
			flagValue:       "bundle",
			buildName:       "npm-flag-bundle",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Flag: bundle (bundleDependencies only)",
			category:        "individual",
		},

		// ===== 3. PERMUTATIONS & COMBINATIONS (8 tests) =====
		// 2-flag combinations
		{
			name:            "combo_peer_optional",
			flagValue:       "peer,optional",
			buildName:       "npm-combo-peer-opt",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Combo: peer + optional (2-way combination)",
			category:        "combo",
		},
		{
			name:            "combo_peer_bundle",
			flagValue:       "peer,bundle",
			buildName:       "npm-combo-peer-bundle",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Combo: peer + bundle (2-way combination)",
			category:        "combo",
		},
		{
			name:            "combo_optional_bundle",
			flagValue:       "optional,bundle",
			buildName:       "npm-combo-opt-bundle",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Combo: optional + bundle (2-way combination)",
			category:        "combo",
		},
		{
			name:            "combo_regular_optional",
			flagValue:       "regular,optional",
			buildName:       "npm-combo-reg-opt",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Combo: regular + optional (2-way combination)",
			category:        "combo",
		},
		{
			name:            "combo_all_peer",
			flagValue:       "all,peer",
			buildName:       "npm-combo-all-peer",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Combo: all + peer (redundant but valid - all subsumes peer)",
			category:        "combo",
		},
		// 3-flag combinations
		{
			name:            "combo_peer_optional_bundle",
			flagValue:       "peer,optional,bundle",
			buildName:       "npm-combo-trio",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Combo: peer + optional + bundle (3-way combination)",
			category:        "combo",
		},
		{
			name:            "combo_all_optional_bundle",
			flagValue:       "all,optional,bundle",
			buildName:       "npm-combo-all-opt-bundle",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Combo: all + optional + bundle (all subsumes others)",
			category:        "combo",
		},
		{
			name:            "combo_regular_peer_bundle",
			flagValue:       "regular,peer,bundle",
			buildName:       "npm-combo-reg-peer-bundle",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Combo: regular + peer + bundle (3-way, no overlap)",
			category:        "combo",
		},

		// ===== 4. SEMANTIC CORRECTNESS - EDGE CASES (3 tests) =====
		// These verify that flags correctly EXCLUDE certain dependency types
		{
			name:            "semantic_regular_excludes_optional",
			flagValue:       "regular",
			buildName:       "npm-sem-reg-excl-opt",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Semantic: 'regular' flag correctly EXCLUDES optional deps from monitoring",
			category:        "semantic",
		},
		{
			name:            "semantic_optional_excludes_regular",
			flagValue:       "optional",
			buildName:       "npm-sem-opt-excl-reg",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Semantic: 'optional' flag ONLY monitors optional deps (excludes regular)",
			category:        "semantic",
		},
		{
			name:            "semantic_peer_excludes_optional",
			flagValue:       "peer",
			buildName:       "npm-sem-peer-excl-opt",
			buildNumber:     "1",
			expectedSuccess: true,
			description:     "Semantic: 'peer' flag correctly EXCLUDES optional deps from monitoring",
			category:        "semantic",
		},

		// ===== NEGATIVE SCENARIOS (Invalid Inputs) =====
		{
			name:            "invalid_flag_unknown_value",
			flagValue:       "invalid",
			buildName:       "npm-invalid-flag",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Should reject: unknown flag value 'invalid'",
			category:        "negative",
		},
		{
			name:            "invalid_flag_case_sensitive_ALL",
			flagValue:       "ALL",
			buildName:       "npm-case-ALL",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Should reject: flag is case-sensitive ('ALL' not valid, must be 'all')",
			category:        "negative",
		},
		{
			name:            "invalid_flag_malformed_trailing_comma",
			flagValue:       "peer,",
			buildName:       "npm-malformed-trailing",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Should reject: malformed flag with trailing comma 'peer,'",
			category:        "negative",
		},
		{
			name:            "invalid_flag_malformed_leading_comma",
			flagValue:       ",peer",
			buildName:       "npm-malformed-leading",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Should reject: malformed flag with leading comma ',peer'",
			category:        "negative",
		},
		{
			name:            "invalid_flag_double_comma",
			flagValue:       "peer,,bundle",
			buildName:       "npm-double-comma",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Should reject: malformed flag with double comma 'peer,,bundle'",
			category:        "negative",
		},
		{
			name:            "invalid_flag_with_spaces",
			flagValue:       "peer, optional",
			buildName:       "npm-spaces-flag",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Should reject: flag with spaces 'peer, optional' (spaces not trimmed)",
			category:        "negative",
		},
		{
			name:            "invalid_flag_special_chars",
			flagValue:       "peer@bundle",
			buildName:       "npm-special-chars",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Should reject: flag with special characters 'peer@bundle'",
			category:        "negative",
		},

		// ===== ERROR MESSAGE FORMAT SCENARIOS (Verify error message structure with actual missing deps) =====
		// These test cases verify the error message format when various combinations of dependencies are missing
		{
			name:            "error_format_regular_only",
			flagValue:       "regular",
			buildName:       "npm-err-regular",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Error format: regular deps missing → shows npm cache hint only",
			category:        "error_format",
		},
		{
			name:            "error_format_peer_only",
			flagValue:       "peer",
			buildName:       "npm-err-peer",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Error format: peer deps missing → shows npm ls hint",
			category:        "error_format",
		},
		{
			name:            "error_format_bundle_only",
			flagValue:       "bundle",
			buildName:       "npm-err-bundle",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Error format: bundle deps missing → shows npm ls hint",
			category:        "error_format",
		},
		{
			name:            "error_format_optional_only",
			flagValue:       "optional",
			buildName:       "npm-err-optional",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Error format: optional deps missing → shows npm ls hint",
			category:        "error_format",
		},
		{
			name:            "error_format_peer_and_optional",
			flagValue:       "peer,optional",
			buildName:       "npm-err-peer-opt",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Error format: peer+optional missing → combines both in npm ls hint",
			category:        "error_format",
		},
		{
			name:            "error_format_peer_and_bundle",
			flagValue:       "peer,bundle",
			buildName:       "npm-err-peer-bundle",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Error format: peer+bundle missing → combines both in npm ls hint",
			category:        "error_format",
		},
		{
			name:            "error_format_all_types_missing",
			flagValue:       "all",
			buildName:       "npm-err-all",
			buildNumber:     "1",
			expectedSuccess: false,
			description:     "Error format: all 4 types missing → shows both npm cache + npm ls hints with all types",
			category:        "error_format",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tt.buildName, artHttpDetails)
			defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tt.buildName, artHttpDetails)

			projectPath := initNpmProjectTest(t)
			chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
			defer chdirCallBack()

			if tt.category == "error_format" {
				// ===== ERROR FORMAT TESTS: Actually recreate missing dependency scenarios =====
				// Pattern: useIsolatedCache → install (populate) → wipeCache (corrupt) → install (detect missing)

				cacheDir, restoreCache := useIsolatedNpmCache(t)
				defer restoreCache()

				// STEP 1: Initial install to populate the isolated cache
				installArgs := []string{"npm", "install", "--cache=" + cacheDir}
				initialErr := runJfrogCliWithoutAssertion(installArgs...)
				assert.NoError(t, initialErr, "Initial cache population should succeed for: %s", tt.description)

				// STEP 2: Corrupt the cache by removing tarballs to simulate missing dependencies
				wipeNpmCacacheTarballs(t, cacheDir)

				// STEP 3: Second install with flag should fail because tarballs are missing
				args := []string{"npm", "install", "--cache=" + cacheDir,
					"--build-name=" + tt.buildName, "--build-number=" + tt.buildNumber,
					"--fail-on-missing-deps=" + tt.flagValue}

				err := runJfrogCliWithoutAssertion(args...)

				// STEP 4: Verify error occurs and message is properly formatted
				assert.Error(t, err, tt.description)
				if err != nil {
					errMsg := err.Error()
					// Verify error message contains appropriate hints based on missing deps type
					if strings.Contains(tt.flagValue, "regular") || tt.flagValue == "all" {
						// Should contain npm cache hint for regular deps
						assert.Contains(t, errMsg, "npm cache",
							"Error should mention npm cache for regular deps: %s", tt.description)
					}
					if tt.flagValue != "regular" && tt.flagValue != "" {
						// Should contain npm ls hint for peer/bundle/optional
						assert.Contains(t, errMsg, "npm ls",
							"Error should mention npm ls for peer/bundle/optional: %s", tt.description)
					}
				}
				t.Logf("[PASS-%s] %s (error recreated with isolated cache corruption)", strings.ToUpper(tt.category), tt.description)

			} else if tt.expectedSuccess {
				// ===== SUCCESS PATH TESTS: Normal flow with valid flags =====
				args := []string{"npm", "install", "--build-name=" + tt.buildName, "--build-number=" + tt.buildNumber}
				if tt.flagValue != "" {
					args = append(args, "--fail-on-missing-deps="+tt.flagValue)
				}

				err := runJfrogCliWithoutAssertion(args...)
				assert.NoError(t, err, tt.description)

				// Publish build info
				assert.NoError(t, artifactoryCli.Exec("bp", tt.buildName, tt.buildNumber),
					"Failed to publish build for: %s", tt.buildName)

				// Verify build info exists and contains modules
				publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tt.buildName, tt.buildNumber)
				assert.NoError(t, err)
				assert.True(t, found, "Build info should exist: %s", tt.description)
				if assert.NotNil(t, publishedBuildInfo) && assert.NotNil(t, publishedBuildInfo.BuildInfo) {
					assert.Greater(t, len(publishedBuildInfo.BuildInfo.Modules), 0,
						"Modules should be present: %s", tt.description)
				}
				t.Logf("[PASS-%s] %s", strings.ToUpper(tt.category), tt.description)

			} else if tt.category == "negative" {
				// ===== NEGATIVE TESTS: Invalid flag values should be rejected =====
				args := []string{"npm", "install", "--build-name=" + tt.buildName, "--build-number=" + tt.buildNumber,
					"--fail-on-missing-deps=" + tt.flagValue}

				err := runJfrogCliWithoutAssertion(args...)
				// Negative test case: should fail with validation error
				assert.Error(t, err, tt.description)
				// Verify the error is about validation (invalid flag value)
				if err != nil {
					assert.Contains(t, err.Error(), "invalid", "Error should mention invalid flag: %s", tt.description)
				}
				t.Logf("[PASS-%s] %s (correctly rejected with validation error)", strings.ToUpper(tt.category), tt.description)
			}

			clientTestUtils.ChangeDirAndAssert(t, wd)
		})
	}
}

// useIsolatedNpmCache points npm at a dedicated cache directory via npm_config_cache.
// Callers must also pass --cache=<dir> to every npm invocation: the env var alone loses to an
// NPM_CONFIG_CACHE already exported by the environment, which makes 'npm config get cache'
// (how the build-info collector locates the cache) report a directory the test never wiped.
func useIsolatedNpmCache(t *testing.T) (cacheDir string, restore func()) {
	cacheDir = t.TempDir()
	return cacheDir, clientTestUtils.SetEnvWithCallbackAndAssert(t, "npm_config_cache", cacheDir)
}

// npmCachedTarballs lists the content-v2 tarballs in the cache, relative to cacheDir.
func npmCachedTarballs(t *testing.T, cacheDir string) []string {
	contentPath := filepath.Join(cacheDir, "_cacache", "content-v2")
	entries, err := os.ReadDir(contentPath)
	if err != nil {
		return nil
	}
	tarballs := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			subentries, _ := os.ReadDir(filepath.Join(contentPath, entry.Name()))
			for _, subentry := range subentries {
				tarballs = append(tarballs, filepath.Join(contentPath, entry.Name(), subentry.Name()))
			}
		}
	}
	return tarballs
}

// wipeNpmCacacheTarballs removes cached tarballs and index-v5 so xml/json cannot be checksummed.
// GetNpmConfigCache requires _cacache to exist; node_modules is left in place so the next
// npm install stays up to date and does not refill the cache from the registry.
func wipeNpmCacacheTarballs(t *testing.T, cacheDir string) {
	cacachePath := filepath.Join(cacheDir, "_cacache")
	tarballs := npmCachedTarballs(t, cacheDir)
	require.NotEmpty(t, tarballs, "cache should hold tarballs before wiping, otherwise the test proves nothing")
	require.NoError(t, os.RemoveAll(filepath.Join(cacachePath, "content-v2")))
	require.NoError(t, os.RemoveAll(filepath.Join(cacachePath, "index-v5")))
	require.NoError(t, os.MkdirAll(cacachePath, 0755))
}

// TestNpmFailOnMissingDepsNegative tests invalid flag values and error handling.
// These tests verify that the flag validation rejects malformed input with clear error messages.
func TestNpmFailOnMissingDepsNegative(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	_, _, err = buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err, "npm must be available for this test")
		return
	}

	testCases := []struct {
		name        string
		flagValue   string
		buildName   string
		buildNumber string
		description string
	}{
		{
			name:        "invalid_unknown_value",
			flagValue:   "invalid",
			buildName:   "npm-invalid-value",
			buildNumber: "1",
			description: "Should reject unknown flag value 'invalid'",
		},
		{
			name:        "invalid_case_sensitive",
			flagValue:   "ALL",
			buildName:   "npm-case-all",
			buildNumber: "1",
			description: "Should reject case-insensitive 'ALL' (must be 'all')",
		},
		{
			name:        "invalid_trailing_comma",
			flagValue:   "peer,",
			buildName:   "npm-trailing-comma",
			buildNumber: "1",
			description: "Should reject trailing comma 'peer,'",
		},
		{
			name:        "invalid_leading_comma",
			flagValue:   ",peer",
			buildName:   "npm-leading-comma",
			buildNumber: "1",
			description: "Should reject leading comma ',peer'",
		},
		{
			name:        "invalid_double_comma",
			flagValue:   "peer,,bundle",
			buildName:   "npm-double-comma",
			buildNumber: "1",
			description: "Should reject double comma 'peer,,bundle'",
		},
		{
			name:        "invalid_with_spaces",
			flagValue:   "peer, optional",
			buildName:   "npm-with-spaces",
			buildNumber: "1",
			description: "Should reject spaces in flag 'peer, optional'",
		},
		{
			name:        "invalid_special_chars",
			flagValue:   "peer@bundle",
			buildName:   "npm-special-chars",
			buildNumber: "1",
			description: "Should reject special characters 'peer@bundle'",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			projectPath := initNpmProjectTest(t)
			chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
			defer chdirCallBack()

			args := []string{"npm", "install",
				"--build-name=" + tt.buildName,
				"--build-number=" + tt.buildNumber,
				"--fail-on-missing-deps=" + tt.flagValue}

			err := runJfrogCliWithoutAssertion(args...)
			// Should fail with validation error
			assert.Error(t, err, tt.description)
			if err != nil {
				assert.Contains(t, err.Error(), "invalid", "Error should mention 'invalid' for: %s", tt.description)
			}
			t.Logf("[PASS-NEGATIVE] %s", tt.description)

			clientTestUtils.ChangeDirAndAssert(t, wd)
		})
	}
}

// TestNpmFailOnMissingDepsErrorFormat tests error message formatting when dependencies are missing.
// Uses isolated cache corruption to actually recreate missing dependency scenarios.
func TestNpmFailOnMissingDepsErrorFormat(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	_, _, err = buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err, "npm must be available for this test")
		return
	}

	testCases := []struct {
		name        string
		flagValue   string
		buildName   string
		buildNumber string
		expectHints []string // Expected hints in error message
		description string
	}{
		{
			name:        "error_regular_deps",
			flagValue:   "regular",
			buildName:   "npm-err-regular",
			buildNumber: "1",
			expectHints: []string{"npm cache"},
			description: "Error should mention npm cache for regular deps",
		},
		{
			name:        "error_peer_deps",
			flagValue:   "peer",
			buildName:   "npm-err-peer",
			buildNumber: "1",
			expectHints: []string{"npm ls"},
			description: "Error should mention npm ls for peer deps",
		},
		{
			name:        "error_bundle_deps",
			flagValue:   "bundle",
			buildName:   "npm-err-bundle",
			buildNumber: "1",
			expectHints: []string{"npm ls"},
			description: "Error should mention npm ls for bundle deps",
		},
		{
			name:        "error_optional_deps",
			flagValue:   "optional",
			buildName:   "npm-err-optional",
			buildNumber: "1",
			expectHints: []string{"npm ls"},
			description: "Error should mention npm ls for optional deps",
		},
		{
			name:        "error_peer_and_optional",
			flagValue:   "peer,optional",
			buildName:   "npm-err-peer-opt",
			buildNumber: "1",
			expectHints: []string{"npm ls"},
			description: "Error should mention npm ls for peer+optional deps",
		},
		{
			name:        "error_peer_and_bundle",
			flagValue:   "peer,bundle",
			buildName:   "npm-err-peer-bundle",
			buildNumber: "1",
			expectHints: []string{"npm ls"},
			description: "Error should mention npm ls for peer+bundle deps",
		},
		{
			name:        "error_all_types",
			flagValue:   "all",
			buildName:   "npm-err-all",
			buildNumber: "1",
			expectHints: []string{"npm cache", "npm ls"},
			description: "Error should mention both npm cache + npm ls for all dep types",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			projectPath := initNpmProjectTest(t)
			chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
			defer chdirCallBack()

			// ===== RECREATE ERROR SCENARIO =====
			// STEP 1: Create isolated cache
			cacheDir, restoreCache := useIsolatedNpmCache(t)
			defer restoreCache()

			// STEP 2: Initial install to populate cache
			installArgs := []string{"npm", "install", "--cache=" + cacheDir}
			initialErr := runJfrogCliWithoutAssertion(installArgs...)
			assert.NoError(t, initialErr, "Cache population should succeed")

			// STEP 3: Corrupt cache to simulate missing dependencies
			wipeNpmCacacheTarballs(t, cacheDir)

			// STEP 4: Run with flag - should fail with missing deps error
			args := []string{"npm", "install", "--cache=" + cacheDir,
				"--build-name=" + tt.buildName,
				"--build-number=" + tt.buildNumber,
				"--fail-on-missing-deps=" + tt.flagValue}

			err := runJfrogCliWithoutAssertion(args...)

			// Verify error occurs and has proper hints
			assert.Error(t, err, tt.description)
			if err != nil {
				errMsg := err.Error()
				for _, hint := range tt.expectHints {
					assert.Contains(t, errMsg, hint, "Error should mention '%s' for: %s", hint, tt.description)
				}
			}
			t.Logf("[PASS-ERROR-FORMAT] %s", tt.description)

			clientTestUtils.ChangeDirAndAssert(t, wd)
		})
	}
}

// TestNpmMissingDepsLegacyBehavior tests backward compatibility: without the flag, missing deps generate debug/warn logs but don't fail.
// This ensures existing workflows that don't use the flag continue to work as before.
func TestNpmMissingDepsLegacyBehavior(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	_, _, err = buildutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err, "npm must be available for this test")
		return
	}

	t.Run("no_flag_with_missing_deps", func(t *testing.T) {
		buildName := "npm-legacy-warn"
		buildNumber := "1"

		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
		defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

		projectPath := initNpmProjectTest(t)
		chdirCallBack := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
		defer chdirCallBack()

		// Setup isolated cache and corrupt it
		cacheDir, restoreCache := useIsolatedNpmCache(t)
		defer restoreCache()

		// Initial install to populate cache
		installArgs := []string{"npm", "install", "--cache=" + cacheDir}
		initialErr := runJfrogCliWithoutAssertion(installArgs...)
		require.NoError(t, initialErr, "Cache population should succeed")

		// Corrupt cache
		wipeNpmCacacheTarballs(t, cacheDir)

		// WITHOUT --fail-on-missing-deps flag: should succeed (legacy behavior - warns/logs but doesn't fail)
		args := []string{"npm", "install", "--cache=" + cacheDir,
			"--build-name=" + buildName,
			"--build-number=" + buildNumber}

		err := runJfrogCliWithoutAssertion(args...)
		// Legacy behavior: should NOT fail even with missing deps
		assert.NoError(t, err, "WITHOUT flag: missing deps should warn but NOT fail (legacy behavior)")

		// Verify build-info was still published (partial build info is OK without the flag)
		clientTestUtils.ChangeDirAndAssert(t, wd)
		publishErr := artifactoryCli.Exec("bp", buildName, buildNumber)
		// May or may not succeed depending on whether build-info was collected, but the install itself should have succeeded
		if publishErr == nil {
			publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
			assert.NoError(t, err)
			if found && publishedBuildInfo != nil {
				// Build info exists (may be partial without strict mode)
				assert.NotNil(t, publishedBuildInfo.BuildInfo, "Build info should be populated")
			}
		}

		t.Logf("[PASS-LEGACY] Without flag: missing deps warn/log but don't fail (backward compat preserved)")
	})
}

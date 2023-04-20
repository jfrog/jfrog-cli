package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/yarn"

	biutils "github.com/jfrog/build-info-go/build/utils"
	"github.com/jfrog/gofrog/version"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	clientutils "github.com/jfrog/jfrog-client-go/utils"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/ioutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
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
	npmVersion, _, err := biutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}
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

func readModuleId(t *testing.T, wd string, npmVersion *version.Version) string {
	packageInfo, err := biutils.ReadPackageInfoFromPackageJson(filepath.Dir(wd), npmVersion)
	assert.NoError(t, err)
	return packageInfo.BuildInfoModuleId()
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
	buildInfoService := utils.CreateBuildInfoService()
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

func TestNpmConditionalUpload(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	npmProjectPath := initNpmProjectTest(t)
	clientTestUtils.ChangeDirAndAssert(t, filepath.Dir(npmProjectPath))
	buildName := tests.NpmBuildName + "-scan"
	buildNumber := "505"
	runJfrogCli(t, []string{"npm", "install", "--build-name=" + buildName, "--build-number=" + buildNumber}...)

	execFunc := func() error {
		defer clientTestUtils.ChangeDirAndAssert(t, wd)
		return runJfrogCliWithoutAssertion([]string{"npm", "publish", "--scan", "--build-name=" + buildName, "--build-number=" + buildNumber}...)
	}
	testConditionalUpload(t, execFunc, tests.SearchAllNpm)
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

func initNpmFilesTest(t *testing.T) (npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath, npmProjectCi, npmPostInstallProjectPath string) {
	npmProjectPath = createNpmProject(t, "npmproject")
	npmScopedProjectPath = createNpmProject(t, "npmscopedproject")
	npmNpmrcProjectPath = createNpmProject(t, "npmnpmrcproject")
	npmProjectCi = createNpmProject(t, "npmprojectci")
	npmPostInstallProjectPath = createNpmProject(t, "npmpostinstall")
	_ = createNpmProject(t, filepath.Join("npmpostinstall", "subdir"))
	err := createConfigFileForTest([]string{filepath.Dir(npmProjectPath), filepath.Dir(npmScopedProjectPath),
		filepath.Dir(npmNpmrcProjectPath), filepath.Dir(npmProjectCi), filepath.Dir(npmPostInstallProjectPath)}, tests.NpmRemoteRepo, tests.NpmRepo, t, utils.Npm, false)
	assert.NoError(t, err)
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectPath))
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectCi))
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmPostInstallProjectPath))
	return
}

func initNpmProjectTest(t *testing.T) (npmProjectPath string) {
	npmProjectPath = createNpmProject(t, "npmproject")
	err := createConfigFileForTest([]string{filepath.Dir(npmProjectPath)}, tests.NpmRemoteRepo, tests.NpmRepo, t, utils.Npm, false)
	assert.NoError(t, err)
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectPath))
	return
}

func initGlobalNpmFilesTest(t *testing.T) (npmProjectPath string) {
	npmProjectPath = createNpmProject(t, "npmproject")
	jfrogHomeDir, err := coreutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	err = createConfigFileForTest([]string{jfrogHomeDir}, tests.NpmRemoteRepo, tests.NpmRepo, t, utils.Npm, true)
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
	verifyExistInArtifactoryByProps(tests.GetNpmDeployedScopedArtifacts(isNpm7),
		tests.NpmRepo+"/*",
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
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
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
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("npm", "install", "-cache="+caches))

	clientTestUtils.RemoveAllAndAssert(t, filepath.Join(workingDirectory, "node_modules"))
	clientTestUtils.RemoveAllAndAssert(t, caches)
}

func initNpmTest(t *testing.T) {
	if !*tests.TestNpm {
		t.Skip("Skipping Npm test. To run Npm test add the '-test.npm=true' option.")
	}
	createJfrogHomeConfig(t, true)
}

func TestNpmPublishDetailedSummary(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)

	npmVersion, _, err := biutils.GetNpmVersionAndExecPath(log.Logger)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Init npm project & npmp command for testing
	npmProjectPath := strings.TrimSuffix(initNpmProjectTest(t), "package.json")
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
	expectedSourcePath := npmProjectPath + tarballName
	expectedTargetPath := serverDetails.ArtifactoryUrl + tests.NpmRepo + "/jfrog-cli-tests/-/" + tarballName
	assert.Equal(t, expectedSourcePath, files[0].SourcePath, "Summary validation failed - unmatched SourcePath.")
	assert.Equal(t, expectedTargetPath, files[0].RtUrl+files[0].TargetPath, "Summary validation failed - unmatched TargetPath.")
	assert.Equal(t, 1, len(files), "Summary validation failed - only one archive should be deployed.")
	// Verify sha256 is valid (a string size 256 characters) and not an empty string.
	assert.Equal(t, 64, len(files[0].Sha256), "Summary validation failed - sha256 should be in size 64 digits.")
}

func TestNpmPublishWithDeploymentView(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	defer clientTestUtils.ChangeDirAndAssert(t, wd)
	initNpmProjectTest(t)
	assertPrintedDeploymentViewFunc, cleanupFunc := initDeploymentViewTest(t)
	defer cleanupFunc()
	runGenericNpm(t, "npm", "publish")
	// Check deployment view
	assertPrintedDeploymentViewFunc()
	// Restore workspace
	clientTestUtils.ChangeDirAndAssert(t, wd)
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
	err = createConfigFileForTest([]string{filepath.Dir(testWorkingDir)}, tests.NpmRemoteRepo, tests.NpmRepo, t, utils.Npm, false)
	assert.NoError(t, err)
	clientTestUtils.ChangeDirAndAssert(t, filepath.Dir(testWorkingDir))
	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	tempCacheDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	buildNumber := "1"
	commandArgs := strings.Split(command, " ")
	commandArgs = append(commandArgs, "yaml")

	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	commandArgs = append(commandArgs, "--cache="+tempCacheDirPath)

	commandArgs = append(commandArgs, "--build-name="+tests.NpmBuildName, "--build-number="+buildNumber)
	runJfrogCli(t, commandArgs...)

	// Validate that no dependencies were collected
	buildInfoService := utils.CreateBuildInfoService()
	goBuild, err := buildInfoService.GetOrCreateBuild(tests.NpmBuildName, buildNumber)
	assert.NoError(t, err)
	defer assert.NoError(t, goBuild.Clean())
	_, err = goBuild.ToBuildInfo()
	assert.Error(t, err)
}

func TestYarn(t *testing.T) {
	initNpmTest(t)
	defer cleanNpmTest(t)

	// Temporarily change the cache folder to a temporary folder - to make sure the cache is clean and dependencies will be downloaded from Artifactory
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	testDataSource := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "yarn")
	testDataTarget := filepath.Join(tempDirPath, tests.Out, "yarn")
	assert.NoError(t, fileutils.CopyDir(testDataSource, testDataTarget, true, nil))

	yarnProjectPath := filepath.Join(testDataTarget, "yarnproject")
	assert.NoError(t, createConfigFileForTest([]string{yarnProjectPath}, tests.NpmRemoteRepo, "", t, utils.Yarn, false))

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

	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
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

func isNpm7(npmVersion *version.Version) bool {
	return npmVersion.Compare("7.0.0") <= 0
}

func TestGenericNpm(t *testing.T) {
	initNpmTest(t)
	runGenericNpm(t, "npm", "--version")
	// Check we don't fail with JFrog flags.
	runGenericNpm(t, "npm", "--version", "--build-name=d", "--build-number=1", "--module=1")
}

func runGenericNpm(t *testing.T, args ...string) {
	jfCli := tests.NewJfrogCli(execMain, "jf", "")
	assert.NoError(t, jfCli.WithoutCredentials().Exec(args...))
}

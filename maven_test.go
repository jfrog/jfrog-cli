package main

import (
	"errors"
	"fmt"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	outputFormat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/build-info-go/build"
	buildinfo "github.com/jfrog/build-info-go/entities"
	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/mvn"
	buildUtils "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	cliproxy "github.com/jfrog/jfrog-cli/utils/tests/proxy/server"
	"github.com/jfrog/jfrog-cli/utils/tests/proxy/server/certificate"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
)

const mavenTestsProxyPort = "1028"
const localRepoSystemProperty = "-Dmaven.repo.local="

var localRepoDir string

func cleanMavenTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	deleteFilesFromRepo(t, tests.MvnRepo1)
	deleteFilesFromRepo(t, tests.MvnRepo2)
	tests.CleanFileSystem()
}

func TestMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t, false)
	assert.NoError(t, runMaven(t, createSimpleMavenProject, tests.MavenConfig, "install"))
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, serverDetails, t)
	cleanMavenTest(t)
}

func TestMavenBuildWithNoProxy(t *testing.T) {
	initMavenTest(t, false)
	// jfrog-ignore - not a real password
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, buildUtils.HttpProxyEnvKey, "http://login:pass@proxy.mydomain:8888")
	defer setEnvCallBack()
	// Set noProxy to match all to skip http proxy configuration
	setNoProxyEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, buildUtils.NoProxyEnvKey, "*")
	defer setNoProxyEnvCallBack()
	assert.NoError(t, runMaven(t, createSimpleMavenProject, tests.MavenConfig, "install"))
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, serverDetails, t)
	cleanMavenTest(t)
}

func TestMavenBuildWithNoProxyHttps(t *testing.T) {
	initMavenTest(t, false)
	// jfrog-ignore - not a real password
	setHttpsEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, buildUtils.HttpsProxyEnvKey, "https://logins:passw@proxys.mydomains:8889")
	defer setHttpsEnvCallBack()
	// Set noProxy to match all to skip https proxy configuration
	setNoProxyEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, buildUtils.NoProxyEnvKey, "*")
	defer setNoProxyEnvCallBack()
	assert.NoError(t, runMaven(t, createSimpleMavenProject, tests.MavenConfig, "install"))
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, serverDetails, t)
	cleanMavenTest(t)
}

func TestMavenBuildWithConditionalUpload(t *testing.T) {
	initMavenTest(t, false)
	buildName := tests.MvnBuildName + "-scan"
	buildNumber := "505"

	execFunc := func() error {
		oldHomeDir := changeWD(t, beforeRunMaven(t, createSimpleMavenProject, tests.MavenConfig))
		defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)
		return runMvnConditionalUploadTest(buildName, buildNumber)
	}
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	testConditionalUpload(t, execFunc, searchSpec, tests.GetMavenDeployedArtifacts()...)
	cleanMavenTest(t)
}

func runMvnConditionalUploadTest(buildName, buildNumber string) (err error) {
	configFilePath, exists, err := project.GetProjectConfFilePath(project.Maven)
	if err != nil {
		return
	}
	if !exists {
		return errors.New("no config file was found!")
	}
	buildConfig := buildUtils.NewBuildConfiguration(buildName, buildNumber, "", "")
	if err = buildConfig.ValidateBuildAndModuleParams(); err != nil {
		return
	}
	printDeploymentView := log.IsStdErrTerminal()
	mvnCmd := mvn.NewMvnCommand().
		SetGoals([]string{"clean", "install", "-B", localRepoSystemProperty + localRepoDir}).
		SetConfiguration(buildConfig).
		SetXrayScan(true).SetScanOutputFormat(outputFormat.Table).
		SetConfigPath(configFilePath).SetDetailedSummary(printDeploymentView).SetThreads(commonCliUtils.Threads)
	err = commands.Exec(mvnCmd)
	result := mvnCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	return cliutils.PrintCommandSummary(mvnCmd.Result(), false, printDeploymentView, false, err)
}

func TestMavenBuildWithServerIDAndDetailedSummary(t *testing.T) {
	initMavenTest(t, false)
	pomDir := createSimpleMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenConfig)
	destPath := filepath.Join(pomDir, ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)

	oldHomeDir := changeWD(t, pomDir)
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	filteredMavenArgs := []string{"clean", "install", "-B", repoLocalSystemProp}
	mvnCmd := mvn.NewMvnCommand().SetConfiguration(new(buildUtils.BuildConfiguration)).SetConfigPath(filepath.Join(destPath, tests.MavenConfig)).SetGoals(filteredMavenArgs).SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(mvnCmd))
	// Validate
	assert.NotNil(t, mvnCmd.Result())
	if mvnCmd.Result() != nil {
		tests.VerifySha256DetailedSummaryFromResult(t, mvnCmd.Result())
	}
	inttestutils.VerifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, serverDetails, t)
	cleanMavenTest(t)
}

func TestMavenBuildWithoutDeployer(t *testing.T) {
	initMavenTest(t, false)
	assert.NoError(t, runMaven(t, createSimpleMavenProject, tests.MavenWithoutDeployerConfig, "install"))
	cleanMavenTest(t)
}

func TestInsecureTlsMavenBuild(t *testing.T) {
	initMavenTest(t, true)
	// Establish a reverse proxy without any certificates
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, tests.HttpsProxyEnvVar, mavenTestsProxyPort)
	defer setEnvCallBack()
	go cliproxy.StartLocalReverseHttpProxy(serverDetails.ArtifactoryUrl, false)
	// Wait for the reverse proxy to start up.
	assert.NoError(t, checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", false))
	// The two certificate files are created by the reverse proxy on startup in the current directory.
	clientTestUtils.RemoveAndAssert(t, certificate.KeyFile)
	clientTestUtils.RemoveAndAssert(t, certificate.CertFile)
	// Save the original Artifactory url, and change the url to proxy url
	oldUrl := tests.JfrogUrl
	proxyUrl := "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort()
	tests.JfrogUrl = &proxyUrl

	assert.NoError(t, createHomeConfigAndLocalRepo(t, false))
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	pomDir := createSimpleMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenConfig)
	destPath := filepath.Join(pomDir, ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)

	oldHomeDir := changeWD(t, pomDir)
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")

	// First, try to run without the insecure-tls flag, failure is expected.
	err = jfrogCli.Exec("mvn", "clean", "install", "-B", repoLocalSystemProp)
	assert.Error(t, err)

	// Run with the insecure-tls flag
	err = jfrogCli.Exec("mvn", "clean", "install", "-B", repoLocalSystemProp, "--insecure-tls")
	if assert.NoError(t, err) {
		// Validate Successful deployment
		inttestutils.VerifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, serverDetails, t)
	}

	tests.JfrogUrl = oldUrl
	cleanMavenTest(t)
}

func createSimpleMavenProject(t *testing.T) string {
	srcPomFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "maven", "mavenproject", "pom.xml")
	pomPath, err := tests.ReplaceTemplateVariables(srcPomFile, "")
	assert.NoError(t, err)
	return filepath.Dir(pomPath)
}

func createMultiMavenProject(t *testing.T) string {
	projectDir := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "maven", "multiproject")
	destPath, err := os.Getwd()
	if err != nil {
		assert.NoError(t, err)
		return ""
	}
	destPath = filepath.Join(destPath, tests.Temp)
	assert.NoError(t, biutils.CopyDir(projectDir, destPath, true, nil))
	return destPath
}

func initMavenTest(t *testing.T, disableConfig bool) {
	if !*tests.TestMaven {
		t.Skip("Skipping Maven test. To run Maven test add the '-test.maven=true' option.")
	}
	if !disableConfig {
		err := createHomeConfigAndLocalRepo(t, true)
		assert.NoError(t, err)
	}
}

func createHomeConfigAndLocalRepo(t *testing.T, encryptPassword bool) (err error) {
	createJfrogHomeConfig(t, encryptPassword)
	// To make sure we download the dependencies from  Artifactory, we will run with customize .m2 directory.
	// The directory wil be deleted on the test cleanup as part as the out dir.
	localRepoDir, err = os.MkdirTemp(os.Getenv(coreutils.HomeDir), "tmp.m2")
	return err
}

// Get the build timestamp from the build info.
func getBuildTimestamp(buildName, buildNumber string, t *testing.T) string {
	service := build.NewBuildInfoService()
	bld, err := service.GetOrCreateBuild(buildName, buildNumber)
	if assert.NoError(t, err) {
		return fmt.Sprintf("%d", bld.GetBuildTimestamp().UnixMilli())
	}
	return ""
}

func TestMavenBuildIncludePatterns(t *testing.T) {
	initMavenTest(t, false)
	buildNumber := "123"
	assert.NoError(t, runMaven(t, createMultiMavenProject, tests.MavenIncludeExcludePatternsConfig, "install", "--build-name="+tests.MvnBuildName, "--build-number="+buildNumber))

	// Validate deployed artifacts.
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetMavenMultiIncludedDeployedArtifacts(), searchSpec, serverDetails, t)
	verifyExistInArtifactoryByProps(tests.GetMavenMultiIncludedDeployedArtifacts(), tests.MvnRepo1+"/*", "build.name="+tests.MvnBuildName+";build.number="+buildNumber+";build.timestamp="+getBuildTimestamp(tests.MvnBuildName, buildNumber, t), t)

	// Validate build info.
	assert.NoError(t, artifactoryCli.Exec("build-publish", tests.MvnBuildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.MvnBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	if len(buildInfo.Modules) != 4 {
		assert.Len(t, buildInfo.Modules, 4)
		return
	}
	validateSpecificModule(buildInfo, t, 13, 2, 1, "org.jfrog.test:multi1:3.7-SNAPSHOT", buildinfo.Maven)
	validateSpecificModule(buildInfo, t, 1, 0, 2, "org.jfrog.test:multi2:3.7-SNAPSHOT", buildinfo.Maven)
	validateSpecificModule(buildInfo, t, 15, 1, 1, "org.jfrog.test:multi3:3.7-SNAPSHOT", buildinfo.Maven)
	validateSpecificModule(buildInfo, t, 0, 1, 0, "org.jfrog.test:multi:3.7-SNAPSHOT", buildinfo.Maven)
	cleanMavenTest(t)
}

func TestMavenDeploy(t *testing.T) {
	initMavenTest(t, false)
	runMavenAndValidateDeployedArtifacts(t, true, "install")
	deleteDeployedArtifacts(t)
	runMavenAndValidateDeployedArtifacts(t, true, "deploy")
	deleteDeployedArtifacts(t)
	runMavenAndValidateDeployedArtifacts(t, false, "package")
}

func runMavenAndValidateDeployedArtifacts(t *testing.T, shouldDeployArtifact bool, args ...string) {
	assert.NoError(t, runMaven(t, createMultiMavenProject, tests.MavenIncludeExcludePatternsConfig, args...))
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	if shouldDeployArtifact {
		inttestutils.VerifyExistInArtifactory(tests.GetMavenMultiIncludedDeployedArtifacts(), searchSpec, serverDetails, t)
	} else {
		results, _ := inttestutils.SearchInArtifactory(searchSpec, serverDetails, t)
		assert.Zero(t, results)
	}
}
func TestMavenWithSummary(t *testing.T) {
	testcases := []struct {
		isDetailedSummary bool
		isDeploymentView  bool
		expectedString    string
		expectedError     error
	}{
		{true, false, `"status": "success",`, nil},
		{false, true, "These files were uploaded:", nil},
	}
	initMavenTest(t, false)
	outputBuffer, stderrBuffer, previousLog := coreTests.RedirectLogOutputToBuffer()
	revertFlags := log.SetIsTerminalFlagsWithCallback(true)
	// Restore previous logger and terminal mode when the function returns
	defer func() {
		log.SetLogger(previousLog)
		revertFlags()
	}()
	for _, test := range testcases {
		args := []string{"install"}
		if test.isDetailedSummary {
			args = append(args, "--detailed-summary")
		}

		assert.NoError(t, runMaven(t, createMultiMavenProject, tests.MavenIncludeExcludePatternsConfig, args...))
		var output []byte
		if test.isDetailedSummary {
			output = outputBuffer.Bytes()
			outputBuffer.Truncate(0)
		} else {
			output = stderrBuffer.Bytes()
			stderrBuffer.Truncate(0)
		}
		assert.True(t, strings.Contains(string(output), test.expectedString), fmt.Sprintf("cant find '%s' in '%s'", test.expectedString, string(output)))
	}
	deleteDeployedArtifacts(t)
}

func deleteDeployedArtifacts(t *testing.T) {
	deleteSpec := spec.NewBuilder().Pattern(tests.MvnRepo1).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.NoError(t, err)
}

func runMaven(t *testing.T, createProjectFunction func(*testing.T) string, configFileName string, args ...string) error {
	oldHomeDir := changeWD(t, beforeRunMaven(t, createProjectFunction, configFileName))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir

	args = append([]string{"mvn", "clean"}, args...)
	args = append(args, "-B", repoLocalSystemProp)
	return runJfrogCliWithoutAssertion(args...)
}

func beforeRunMaven(t *testing.T, createProjectFunction func(*testing.T) string, configFileName string) string {
	projDir := createProjectFunction(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", configFileName)
	destPath := filepath.Join(projDir, ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	assert.NoError(t, os.Rename(filepath.Join(destPath, configFileName), filepath.Join(destPath, "maven.yaml")))
	return projDir
}

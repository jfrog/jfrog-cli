package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	cliproxy "github.com/jfrog/jfrog-cli/utils/tests/proxy/server"
	"github.com/jfrog/jfrog-cli/utils/tests/proxy/server/certificate"
	"github.com/stretchr/testify/assert"
)

const mavenFlagName = "maven"
const mavenTestsProxyPort = "1028"
const localRepoSystemProperty = "-Dmaven.repo.local="

var localRepoDir string

func cleanMavenTest(t *testing.T) {
	assert.NoError(t, os.Unsetenv(coreutils.HomeDir))
	deleteSpec := spec.NewBuilder().Pattern(tests.MvnRepo1).BuildSpec()
	tests.DeleteFiles(deleteSpec, serverDetails)
	deleteSpec = spec.NewBuilder().Pattern(tests.MvnRepo2).BuildSpec()
	tests.DeleteFiles(deleteSpec, serverDetails)
	tests.CleanFileSystem()
}

func TestMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t, false)
	runMavenCleanInstall(t, createSimpleMavenProject, tests.MavenConfig, []string{})
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
	cleanMavenTest(t)
}

func TestMavenBuildWithConditionalUpload(t *testing.T) {
	initMavenTest(t, false)
	runMavenCleanInstall(t, createSimpleMavenProject, tests.MavenConfig, []string{"--scan"})
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
	cleanMavenTest(t)
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
	defer tests.ChangeDirAndAssert(t, oldHomeDir)
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	filteredMavenArgs := []string{"clean", "install", repoLocalSystemProp}
	mvnCmd := mvn.NewMvnCommand().SetConfiguration(new(utils.BuildConfiguration)).SetConfigPath(filepath.Join(destPath, tests.MavenConfig)).SetGoals(filteredMavenArgs).SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(mvnCmd))
	// Validate
	assert.NotNil(t, mvnCmd.Result())
	if mvnCmd.Result() != nil {
		tests.VerifySha256DetailedSummaryFromResult(t, mvnCmd.Result())
	}
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
	cleanMavenTest(t)
}

func TestMavenBuildWithoutDeployer(t *testing.T) {
	initMavenTest(t, false)
	runMavenCleanInstall(t, createSimpleMavenProject, tests.MavenWithoutDeployerConfig, []string{})
	cleanMavenTest(t)
}

func TestInsecureTlsMavenBuild(t *testing.T) {
	initMavenTest(t, true)
	// Establish a reverse proxy without any certificates
	assert.NoError(t, os.Setenv(tests.HttpsProxyEnvVar, mavenTestsProxyPort))
	defer func() { assert.NoError(t, os.Unsetenv(tests.HttpsProxyEnvVar)) }()
	go cliproxy.StartLocalReverseHttpProxy(serverDetails.ArtifactoryUrl, false)
	// The two certificate files are created by the reverse proxy on startup in the current directory.
	os.Remove(certificate.KEY_FILE)
	os.Remove(certificate.CERT_FILE)
	// Wait for the reverse proxy to start up.
	assert.NoError(t, checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", false))
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
	defer tests.ChangeDirAndAssert(t, oldHomeDir)
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")

	// First, try to run without the insecure-tls flag, failure is expected.
	err = jfrogCli.Exec("mvn", "clean", "install", repoLocalSystemProp)
	assert.Error(t, err)
	// Run with the insecure-tls flag
	err = jfrogCli.Exec("mvn", "clean", "install", repoLocalSystemProp, "--insecure-tls")
	assert.NoError(t, err)

	// Validate Successful deployment
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)

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
	assert.NoError(t, fileutils.CopyDir(projectDir, destPath, true, nil))
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
	localRepoDir, err = ioutil.TempDir(os.Getenv(coreutils.HomeDir), "tmp.m2")
	return err
}

func TestMavenBuildIncludePatterns(t *testing.T) {
	initMavenTest(t, false)
	buildNumber := "123"
	commandArgs := []string{"--build-name=" + tests.MvnBuildName, "--build-number=" + buildNumber}
	runMavenCleanInstall(t, createMultiMavenProject, tests.MavenIncludeExcludePatternsConfig, commandArgs)
	assert.NoError(t, artifactoryCli.Exec("build-publish", tests.MvnBuildName, buildNumber))

	// Validate deployed artifacts.
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMavenMultiIncludedDeployedArtifacts(), searchSpec, t)

	// Validate build info.
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

func runMavenCleanInstall(t *testing.T, createProjectFunction func(*testing.T) string, configFileName string, additionalArgs []string) {
	projDir := createProjectFunction(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", configFileName)
	destPath := filepath.Join(projDir, ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	assert.NoError(t, os.Rename(filepath.Join(destPath, configFileName), filepath.Join(destPath, "maven.yaml")))
	oldHomeDir := changeWD(t, projDir)
	defer tests.ChangeDirAndAssert(t, oldHomeDir)
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir

	args := []string{"mvn", "clean", "install", repoLocalSystemProp}
	args = append(args, additionalArgs...)
	runJfrogCli(t, args...)
}

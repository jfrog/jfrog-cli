package main

import (
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/common/commands"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	cliproxy "github.com/jfrog/jfrog-cli/utils/tests/proxy/server"
	"github.com/jfrog/jfrog-cli/utils/tests/proxy/server/certificate"
	"github.com/stretchr/testify/assert"
)

const mavenFlagName = "maven"
const mavenTestsProxyPort = "1028"
const localRepoSystemProperty = "-Dmaven.repo.local="

var localRepoDir string

func cleanMavenTest() {
	os.Unsetenv(coreutils.HomeDir)
	deleteSpec := spec.NewBuilder().Pattern(tests.MvnRepo1).BuildSpec()
	tests.DeleteFiles(deleteSpec, serverDetails)
	deleteSpec = spec.NewBuilder().Pattern(tests.MvnRepo2).BuildSpec()
	tests.DeleteFiles(deleteSpec, serverDetails)
	tests.CleanFileSystem()
}

func TestMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t, false)

	pomPath := filepath.Join(createSimpleMavenProject(t), "pom.xml")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	assert.NoError(t, err)
	runAndValidateMaven(pomPath, configFilePath, t)
	cleanMavenTest()
}

func TestNativeMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t, false)
	runNativeMavenCleanInstall(t, createSimpleMavenProject, tests.MavenConfig, []string{})
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
	cleanMavenTest()
}

func TestNativeMavenBuildWithServerIDAndDetailedSummary(t *testing.T) {
	initMavenTest(t, false)
	pomDir := createSimpleMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenConfig)
	destPath := filepath.Join(pomDir, ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)

	oldHomeDir := changeWD(t, pomDir)
	defer func() {
		err := os.Chdir(oldHomeDir)
		assert.NoError(t, err)
	}()
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	filteredMavenArgs := []string{"clean", "install", repoLocalSystemProp}

	mvnCmd := mvn.NewMvnCommand().SetConfiguration(new(utils.BuildConfiguration)).SetConfigPath(filepath.Join(destPath, tests.MavenConfig)).SetGoals(filteredMavenArgs).SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(mvnCmd))

	// Validate
	tests.VerifySha256DetailedSummaryFromResult(t, mvnCmd.Result())
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
	cleanMavenTest()
}

func TestMavenBuildWithoutDeployer(t *testing.T) {
	initMavenTest(t, false)
	runNativeMavenCleanInstall(t, createSimpleMavenProject, tests.MavenWithoutDeployerConfig, []string{})
	cleanMavenTest()
}

// This test check legacy behavior whereby the Maven config yml contains the username, url and password.
func TestMavenBuildWithCredentials(t *testing.T) {
	initMavenTest(t, false)

	if *tests.RtAccessToken != "" {
		origUsername, origPassword := tests.SetBasicAuthFromAccessToken(t)
		defer func() {
			*tests.RtUser = origUsername
			*tests.RtPassword = origPassword
		}()
	}

	pomPath := filepath.Join(createSimpleMavenProject(t), "pom.xml")
	srcConfigTemplate := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenUsernamePasswordTemplate)
	configFilePath, err := tests.ReplaceTemplateVariables(srcConfigTemplate, "")
	assert.NoError(t, err)

	runAndValidateMaven(pomPath, configFilePath, t)
	cleanMavenTest()
}

func TestInsecureTlsMavenBuild(t *testing.T) {
	initMavenTest(t, true)
	// Establish a reverse proxy without any certificates
	os.Setenv(tests.HttpsProxyEnvVar, mavenTestsProxyPort)
	go cliproxy.StartLocalReverseHttpProxy(serverDetails.ArtifactoryUrl, false)
	// The two certificate files are created by the reverse proxy on startup in the current directory.
	os.Remove(certificate.KEY_FILE)
	os.Remove(certificate.CERT_FILE)
	// Wait for the reverse proxy to start up.
	assert.NoError(t, checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", false))
	// Save the original Artifactory url, and change the url to proxy url
	oldRtUrl := tests.RtUrl
	parsedUrl, err := url.Parse(serverDetails.ArtifactoryUrl)
	proxyUrl := "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()
	tests.RtUrl = &proxyUrl

	assert.NoError(t, createHomeConfigAndLocalRepo(t, false))
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	pomDir := createSimpleMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenConfig)
	destPath := filepath.Join(pomDir, ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)

	oldHomeDir := changeWD(t, pomDir)
	defer func() {
		err := os.Chdir(oldHomeDir)
		assert.NoError(t, err)
	}()
	rtCli := tests.NewJfrogCli(execMain, "jfrog rt", "")

	// First, try to run without the insecure-tls flag, failure is expected.
	err = rtCli.Exec("mvn", "clean", "install", repoLocalSystemProp)
	assert.Error(t, err)
	// Run with the insecure-tls flag
	err = rtCli.Exec("mvn", "clean", "install", repoLocalSystemProp, "--insecure-tls")
	assert.NoError(t, err)

	// Validate Successful deployment
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)

	tests.RtUrl = oldRtUrl
	cleanMavenTest()
}

func runAndValidateMaven(pomPath, configFilePath string, t *testing.T) {
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	runCliWithLegacyBuildtoolsCmd(t, "mvn", "clean install -f "+pomPath+" "+repoLocalSystemProp, configFilePath)
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
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
	localRepoDir, err = os.MkdirTemp(os.Getenv(coreutils.HomeDir), "tmp.m2")
	return err
}

func TestMavenBuildIncludePatterns(t *testing.T) {
	initMavenTest(t, false)
	buildNumber := "123"
	commandArgs := []string{"--build-name=" + tests.MvnBuildName, "--build-number=" + buildNumber}
	runNativeMavenCleanInstall(t, createMultiMavenProject, tests.MavenIncludeExcludePatternsConfig, commandArgs)
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
	cleanMavenTest()
}

func runNativeMavenCleanInstall(t *testing.T, createProjectFunction func(*testing.T) string, configFileName string, additionalArgs []string) {
	projDir := createProjectFunction(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", configFileName)
	destPath := filepath.Join(projDir, ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	assert.NoError(t, os.Rename(filepath.Join(destPath, configFileName), filepath.Join(destPath, "maven.yaml")))
	oldHomeDir := changeWD(t, projDir)
	defer func() {
		err := os.Chdir(oldHomeDir)
		assert.NoError(t, err)
	}()
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir

	args := []string{"mvn", "clean", "install", repoLocalSystemProp}
	args = append(args, additionalArgs...)
	runCli(t, args...)
}

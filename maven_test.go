package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands/buildinfo"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	cliproxy "github.com/jfrog/jfrog-cli/utils/tests/proxy/server"
	"github.com/jfrog/jfrog-cli/utils/tests/proxy/server/certificate"
	buildinfocmd "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

const mavenFlagName = "maven"
const mavenTestsProxyPort = "1028"
const localRepoSystemProperty = "-Dmaven.repo.local="

var localRepoDir string

func cleanMavenTest() {
	os.Unsetenv(coreutils.HomeDir)
	deleteSpec := spec.NewBuilder().Pattern(tests.MvnRepo1).BuildSpec()
	tests.DeleteFiles(deleteSpec, artifactoryDetails)
	deleteSpec = spec.NewBuilder().Pattern(tests.MvnRepo2).BuildSpec()
	tests.DeleteFiles(deleteSpec, artifactoryDetails)
	tests.CleanFileSystem()
}

func TestMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t, false)

	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	assert.NoError(t, err)
	runAndValidateMaven(pomPath, configFilePath, t)
	cleanMavenTest()
}

func TestNativeMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t, false)
	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenConfig)
	destPath := filepath.Join(filepath.Dir(pomPath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(pomPath))
	pomPath = strings.Replace(pomPath, `\`, "/", -1) // Windows compatibility.
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	runCli(t, "mvn", "clean", "install", "-f", pomPath, repoLocalSystemProp)
	err := os.Chdir(oldHomeDir)
	assert.NoError(t, err)
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
	cleanMavenTest()
}

func TestMavenBuildWithoutDeployer(t *testing.T) {
	initMavenTest(t, false)
	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", "maven_without_deployer", tests.MavenConfig)
	destPath := filepath.Join(filepath.Dir(pomPath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(pomPath))
	pomPath = strings.Replace(pomPath, `\`, "/", -1) // Windows compatibility.
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	runCli(t, "mvn", "clean", "install", "-f", pomPath, repoLocalSystemProp)
	err := os.Chdir(oldHomeDir)
	assert.NoError(t, err)
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

	pomPath := createMavenProject(t)
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
	go cliproxy.StartLocalReverseHttpProxy(artifactoryDetails.Url, false)
	// The two certificate files are created by the reverse proxy on startup in the current directory.
	os.Remove(certificate.KEY_FILE)
	os.Remove(certificate.CERT_FILE)
	// Wait for the reverse proxy to start up.
	assert.NoError(t, checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", false))
	// Save the original Artifactory url, and change the url to proxy url
	oldRtUrl := tests.RtUrl
	parsedUrl, err := url.Parse(artifactoryDetails.Url)
	proxyUrl := "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()
	tests.RtUrl = &proxyUrl

	assert.NoError(t, createHomeConfigAndLocalRepo(t, false))
	repoLocalSystemProp := localRepoSystemProperty + localRepoDir
	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenConfig)
	destPath := filepath.Join(filepath.Dir(pomPath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(pomPath))
	pomPath = strings.Replace(pomPath, `\`, "/", -1) // Windows compatibility.
	rtCli := tests.NewJfrogCli(execMain, "jfrog rt", "")

	// First, try to run without the insecure-tls flag, failure is expected.
	err = rtCli.Exec("mvn", "clean", "install", "-f", pomPath, repoLocalSystemProp)
	assert.Error(t, err)
	// Run with the insecure-tls flag
	err = rtCli.Exec("mvn", "clean", "install", "-f", pomPath, repoLocalSystemProp, "--insecure-tls")
	assert.NoError(t, err)
	err = os.Chdir(oldHomeDir)
	assert.NoError(t, err)

	// Validate Successful deployment
	searchSpec, err := tests.CreateSpec(tests.SearchAllMaven)
	assert.NoError(t, err)
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

func createMavenProject(t *testing.T) string {
	srcPomFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "mavenproject", "pom.xml")
	pomPath, err := tests.ReplaceTemplateVariables(srcPomFile, "")
	assert.NoError(t, err)
	return pomPath
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

// build two maven projects named DependencyMvnProject & MavenProject. MavenProject uses DependencyMvnProject as its dependency.
// By building & publishing MavenProject to Artifactory, one of the steps to do so is to collect its dependencies' build property.
// This tests validate that each dependency inside the build info, update its build property by searching it in Artifactory.
func TestBuildOfBuildsMaven(t *testing.T) {
	// Initialize
	initMavenTest(t, false)
	dependencyMvnProject, mvnProject := copyMavenProjectsToOutDir(t)
	buildNumber := "1"

	//install dependency project.
	changeWD(t, dependencyMvnProject)
	runCli(t, "mvn", "clean", "install", "--build-name="+tests.RtBuildOfBuildMavenDependencyProject, "--build-number="+buildNumber)

	// Publish maven dependency project's buildInfo.
	runCli(t, "bp", tests.RtBuildOfBuildMavenDependencyProject, buildNumber)

	// Install maven project which depends on DependencyMvnProject.
	changeWD(t, mvnProject)
	runCli(t, "mvn", "clean", "install", "--build-name="+tests.RtBuildOfBuildMavenProject, "--build-number="+buildNumber)

	// Genarete build info for maven project.
	buildConfig := &utils.BuildConfiguration{BuildName: tests.RtBuildOfBuildMavenProject, BuildNumber: buildNumber}
	publishConfig := new(buildinfocmd.Configuration)
	publishCmd := buildinfo.NewBuildPublishCommand().SetBuildConfiguration(buildConfig).SetConfig(publishConfig).SetRtDetails(artifactoryDetails)
	servicesManager, err := utils.CreateServiceManager(artifactoryDetails, false)
	assert.NoError(t, err)
	buildInfo, err := publishCmd.GenerateBuildInfo(servicesManager)
	assert.NoError(t, err)

	// Validate that the dependency got the value of its build.
	dep := buildInfo.Modules[0].Dependencies[0]
	assert.True(t, strings.HasPrefix(dep.Build, tests.RtBuildOfBuildMavenDependencyProject+"/"+buildNumber+"/"))
	idx := strings.LastIndex(dep.Build, "/")
	assert.True(t, len(dep.Build[idx:]) > 3)

	// Cleanup
	assert.NoError(t, utils.RemoveBuildDir(tests.RtBuildOfBuildMavenDependencyProject, buildNumber))
	assert.NoError(t, utils.RemoveBuildDir(tests.RtBuildOfBuildMavenProject, buildNumber))
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.RtBuildOfBuildGenericDownload, artHttpDetails)
	cleanMavenTest()
}

func copyMavenProjectsToOutDir(t *testing.T) (string, string) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	mavenTestdataPath := filepath.Join(wd, "testdata", "maven")
	assert.NoError(t, fileutils.CopyDir(mavenTestdataPath, tests.Out, true, nil), err)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", "buildofbuilds", tests.MavenConfig)
	createConfigFile(filepath.Join(tests.Out, "dependencyproject", ".jfrog", "projects"), configFilePath, t)
	createConfigFile(filepath.Join(tests.Out, "project", ".jfrog", "projects"), configFilePath, t)
	return filepath.Join(wd, tests.Out, "dependencyproject"), filepath.Join(wd, tests.Out, "project")
}

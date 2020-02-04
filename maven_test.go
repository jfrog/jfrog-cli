package main

import (
	"github.com/jfrog/jfrog-cli-go/utils/tests/proxy/server/certificate"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-go/utils/tests"
	cliproxy "github.com/jfrog/jfrog-cli-go/utils/tests/proxy/server"
)

const mavenFlagName = "maven"
const localRepoSystemProperty = "-Dmaven.repo.local="

func TestMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t, false)

	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	assert.NoError(t, err)
	runAndValidateMaven(pomPath, configFilePath, t)
	cleanBuildToolsTest()
}

func TestNativeMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t, false)
	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenConfig)
	destPath := filepath.Join(filepath.Dir(pomPath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(pomPath))
	pomPath = strings.Replace(pomPath, `\`, "/", -1) // Windows compatibility.
	runCli(t, "mvn", "clean", "install", "-f", pomPath)
	err := os.Chdir(oldHomeDir)
	assert.NoError(t, err)
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
	cleanBuildToolsTest()
}

func TestMavenBuildWithCredentials(t *testing.T) {
	if *tests.RtUser == "" || *tests.RtPassword == "" {
		t.SkipNow()
	}
	initMavenTest(t, false)
	pomPath := createMavenProject(t)
	srcConfigTemplate := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenUsernamePasswordTemplate)
	configFilePath, err := tests.ReplaceTemplateVariables(srcConfigTemplate, "")
	assert.NoError(t, err)

	runAndValidateMaven(pomPath, configFilePath, t)
	cleanBuildToolsTest()
}

func TestInsecureTlsMavenBuild(t *testing.T) {
	initMavenTest(t, true)
	// Establish a reverse proxy without any certificates
	os.Setenv(tests.HttpsProxyEnvVar, "1024")
	go cliproxy.StartLocalReverseHttpProxy(artifactoryDetails.Url, false)
	// The two certificate files are created by the reverse proxy on startup in the current directory.
	os.Remove(certificate.KEY_FILE)
	os.Remove(certificate.CERT_FILE)
	// Wait for the reverse proxy to start up.
	err := checkIfServerIsUp(cliproxy.GetProxyHttpsPort(), "https", false)
	if err != nil {
		t.Error(err)
	}
	// Save the original Artifactory url, and change the url to proxy url
	oldRtUrl := tests.RtUrl
	parsedUrl, err := url.Parse(artifactoryDetails.Url)
	proxyUrl := "https://127.0.0.1:" + cliproxy.GetProxyHttpsPort() + parsedUrl.RequestURI()
	tests.RtUrl = &proxyUrl

	createJfrogHomeConfig(t)
	// To make sure we download the dependencies from  Artifactory, we will run with customize .m2 directory.
	tmpM2, err := ioutil.TempDir("", "tmp.m2")
	assert.NoError(t, err)
	defer fileutils.RemoveTempDir(tmpM2)
	repoLocalSystemProp := localRepoSystemProperty + tmpM2
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
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)

	tests.RtUrl = oldRtUrl
	cleanBuildToolsTest()
}

func runAndValidateMaven(pomPath, configFilePath string, t *testing.T) {
	runCliWithLegacyBuildtoolsCmd(t, "mvn", "clean install -f "+pomPath, configFilePath)
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
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
	createJfrogHomeConfig(t)
}

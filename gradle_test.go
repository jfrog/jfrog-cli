package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

const gradleFlagName = "gradle"

func TestGradleBuildWithServerID(t *testing.T) {
	initGradleTest(t)

	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	assert.NoError(t, err)
	buildName := "gradle-cli"
	buildNumber := "1"
	runAndValidateGradle(buildGradlePath, configFilePath, buildName, buildNumber, t)
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo, _ := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")

	cleanBuildToolsTest()
}

func TestNativeGradleBuildWithServerID(t *testing.T) {
	initGradleTest(t)
	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	buildName := "gradle-cli"
	buildNumber := "1"
	buildGradlePath = strings.Replace(buildGradlePath, `\`, "/", -1) // Windows compatibility.
	runCli(t, "gradle", "clean artifactoryPublish", "-b"+buildGradlePath, "--build-name="+buildName, "--build-number="+buildNumber)
	err := os.Chdir(oldHomeDir)
	assert.NoError(t, err)
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
	verifyExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.Repo1+"/*", "build.name="+buildName+";build.number="+buildNumber, t)
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo, _ := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")
	cleanBuildToolsTest()
}

func TestGradleBuildWithServerIDWithUsesPlugin(t *testing.T) {
	initGradleTest(t)

	buildGradlePath := createGradleProject(t, "projectwithplugin")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleServerIDUsesPluginConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	assert.NoError(t, err)
	buildName := "gradle-cli"
	buildNumber := "1"
	runAndValidateGradle(buildGradlePath, configFilePath, buildName, buildNumber, t)

	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo, _ := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")
	cleanBuildToolsTest()
}

func TestGradleBuildWithCredentials(t *testing.T) {
	if *tests.RtUser == "" || *tests.RtPassword == "" {
		t.SkipNow()
	}

	initGradleTest(t)

	buildName := "gradle-cli"
	buildNumber := "1"
	buildGradlePath := createGradleProject(t, "gradleproject")
	srcConfigTemplate := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleUsernamePasswordTemplate)
	configFilePath, err := tests.ReplaceTemplateVariables(srcConfigTemplate, "")
	assert.NoError(t, err)

	runAndValidateGradle(buildGradlePath, configFilePath, buildName, buildNumber, t)
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo, _ := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")
	cleanBuildToolsTest()
}

func runAndValidateGradle(buildGradlePath, configFilePath, buildName, buildNumber string, t *testing.T) {
	runCliWithLegacyBuildtoolsCmd(t, "gradle", "clean artifactoryPublish -b "+buildGradlePath, configFilePath, "--build-name="+buildName, "--build-number="+buildNumber)
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
	verifyExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.Repo1+"/*", "build.name="+buildName+";build.number="+buildNumber, t)
}

func createGradleProject(t *testing.T, projectName string) string {
	srcBuildFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradle", projectName, "build.gradle")
	buildGradlePath, err := tests.ReplaceTemplateVariables(srcBuildFile, "")
	assert.NoError(t, err)

	srcSettingsFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradle", projectName, "settings.gradle")
	_, err = tests.ReplaceTemplateVariables(srcSettingsFile, "")
	assert.NoError(t, err)

	return buildGradlePath
}
func initGradleTest(t *testing.T) {
	if !*tests.TestGradle {
		t.Skip("Skipping Gradle test. To run Gradle test add the '-test.gradle=true' option.")
	}
	createJfrogHomeConfig(t, true)
}

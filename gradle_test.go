package main

import (
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/common/commands"
	"github.com/jfrog/jfrog-cli-core/common/spec"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"

	"github.com/stretchr/testify/assert"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

const (
	gradleFlagName = "gradle"
	gradleModuleId = ":minimal-example:1.0"
)

func cleanGradleTest() {
	os.Unsetenv(coreutils.HomeDir)
	deleteSpec := spec.NewBuilder().Pattern(tests.GradleRepo).BuildSpec()
	tests.DeleteFiles(deleteSpec, serverDetails)
	tests.CleanFileSystem()
}

func TestGradleBuildWithServerID(t *testing.T) {
	initGradleTest(t)

	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	assert.NoError(t, err)
	buildName := "gradle-cli"
	buildNumber := "1"
	runAndValidateGradle(buildGradlePath, configFilePath, buildName, buildNumber, t)
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, 0, 1, gradleModuleId, buildinfo.Gradle)

	cleanGradleTest()
}

func TestNativeGradleBuildWithServerID(t *testing.T) {
	initGradleTest(t)
	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	buildNumber := "1"
	buildGradlePath = strings.Replace(buildGradlePath, `\`, "/", -1) // Windows compatibility.
	runCli(t, "gradle", "clean artifactoryPublish", "-b"+buildGradlePath, "--build-name="+tests.GradleBuildName, "--build-number="+buildNumber)
	err := os.Chdir(oldHomeDir)
	assert.NoError(t, err)
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
	verifyExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.GradleRepo+"/*", "build.name="+tests.GradleBuildName+";build.number="+buildNumber, t)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.GradleBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.GradleBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, 0, 1, gradleModuleId, buildinfo.Gradle)
	cleanGradleTest()
}

func TestNativeGradleBuildWithServerIDAndDetailedSummary(t *testing.T) {
	initGradleTest(t)
	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	buildNumber := "1"
	buildGradlePath = strings.Replace(buildGradlePath, `\`, "/", -1) // Windows compatibility.

	// Test gradle with detailed summary without buildinfo props.
	filteredGradleArgs := []string{"clean artifactoryPublish", "-b" + buildGradlePath}
	gradleCmd := gradle.NewGradleCommand().SetConfiguration(new(utils.BuildConfiguration)).SetTasks(strings.Join(filteredGradleArgs, " ")).SetConfigPath(filepath.Join(destPath, "gradle.yaml")).SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(gradleCmd))
	// Validate sha256
	tests.VerifySha256DetailedSummaryFromResult(t, gradleCmd.Result())

	// Test gradle with detailed summary + buildinfo.
	buildConfiguration := &utils.BuildConfiguration{BuildName: tests.GradleBuildName, BuildNumber: buildNumber}
	gradleCmd = gradle.NewGradleCommand().SetConfiguration(buildConfiguration).SetTasks(strings.Join(filteredGradleArgs, " ")).SetConfigPath(filepath.Join(destPath, "gradle.yaml")).SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(gradleCmd))
	// Validate sha256
	tests.VerifySha256DetailedSummaryFromResult(t, gradleCmd.Result())

	err := os.Chdir(oldHomeDir)
	assert.NoError(t, err)
	// Validate build info
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
	verifyExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.GradleRepo+"/*", "build.name="+tests.GradleBuildName+";build.number="+buildNumber, t)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.GradleBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.GradleBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, 0, 1, gradleModuleId, buildinfo.Gradle)
	cleanGradleTest()
}

func TestGradleBuildWithServerIDWithUsesPlugin(t *testing.T) {
	initGradleTest(t)

	buildGradlePath := createGradleProject(t, "projectwithplugin")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleServerIDUsesPluginConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	assert.NoError(t, err)
	buildNumber := "1"
	runAndValidateGradle(buildGradlePath, configFilePath, tests.GradleBuildName, buildNumber, t)

	assert.NoError(t, artifactoryCli.Exec("bp", tests.GradleBuildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.GradleBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, 0, 1, gradleModuleId, buildinfo.Gradle)
	cleanGradleTest()
}

// This test check legacy behavior whereby the Gradle config yml contains the username, url and password.
func TestGradleBuildWithCredentials(t *testing.T) {
	initGradleTest(t)

	if *tests.RtAccessToken != "" {
		origUsername, origPassword := tests.SetBasicAuthFromAccessToken(t)
		defer func() {
			*tests.RtUser = origUsername
			*tests.RtPassword = origPassword
		}()
	}

	buildNumber := "1"
	buildGradlePath := createGradleProject(t, "gradleproject")
	srcConfigTemplate := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleUsernamePasswordTemplate)
	configFilePath, err := tests.ReplaceTemplateVariables(srcConfigTemplate, "")
	assert.NoError(t, err)

	runAndValidateGradle(buildGradlePath, configFilePath, tests.GradleBuildName, buildNumber, t)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.GradleBuildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.GradleBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, 0, 1, gradleModuleId, buildinfo.Gradle)
	cleanGradleTest()
}

func runAndValidateGradle(buildGradlePath, configFilePath, buildName, buildNumber string, t *testing.T) {
	runCliWithLegacyBuildtoolsCmd(t, "gradle", "clean artifactoryPublish -b "+buildGradlePath, configFilePath, "--build-name="+buildName, "--build-number="+buildNumber)
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
	verifyExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.GradleRepo+"/*", "build.name="+buildName+";build.number="+buildNumber, t)
	inttestutils.ValidateGeneratedBuildInfoModule(t, buildName, buildNumber, "", []string{gradleModuleId}, buildinfo.Gradle)
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

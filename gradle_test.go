package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-go/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/inttestutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
)

const gradleFlagName = "gradle"

func TestGradleBuildWithServerID(t *testing.T) {
	initGradleTest(t)

	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	if err != nil {
		t.Error(err)
	}
	buildName := "gradle-cli"
	buildNumber := "1"
	runAndValidateGradle(buildGradlePath, configFilePath, t, &utils.BuildConfiguration{BuildName: buildName, BuildNumber: buildNumber})
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
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
	runCli(t, "gradle", "clean artifactoryPublish","-b"+buildGradlePath, "--build-name="+buildName, "--build-number="+buildNumber)
	err := os.Chdir(oldHomeDir)
	if err != nil {
		t.Error(err)
	}
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	if err != nil {
		t.Error(err)
	}
	isExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
	isExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.Repo1+"/*", "build.name="+buildName+";build.number="+buildNumber, t)
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")
	cleanBuildToolsTest()
}

func TestGradleBuildWithServerIDWithUsesPlugin(t *testing.T) {
	initGradleTest(t)

	buildGradlePath := createGradleProject(t, "projectwithplugin")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleServerIDUsesPluginConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	if err != nil {
		t.Error(err)
	}
	buildName := "gradle-cli"
	buildNumber := "1"
	runAndValidateGradle(buildGradlePath, configFilePath, t, &utils.BuildConfiguration{BuildName: buildName, BuildNumber: buildNumber})

	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
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
	srcConfigTemplate := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleUseramePasswordTemplate)
	configFilePath, err := tests.ReplaceTemplateVariables(srcConfigTemplate, "")
	if err != nil {
		t.Error(err)
	}

	runAndValidateGradle(buildGradlePath, configFilePath, t, &utils.BuildConfiguration{BuildName: buildName, BuildNumber: buildNumber})
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")
	cleanBuildToolsTest()
}

func runAndValidateGradle(buildGradlePath, configFilePath string, t *testing.T, buildConfig *utils.BuildConfiguration) {
	gradleCmd := gradle.NewGradleCommand().SetTasks("clean artifactoryPublish -b " + buildGradlePath).SetConfigPath(configFilePath).SetConfiguration(buildConfig)
	err := gradleCmd.Run()
	if err != nil {
		t.Error(err)
	}
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	if err != nil {
		t.Error(err)
	}

	isExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
	isExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.Repo1+"/*", "build.name="+buildConfig.BuildName+";build.number="+buildConfig.BuildNumber, t)
}

func createGradleProject(t *testing.T, projectName string) string {
	srcBuildFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradle", projectName, "build.gradle")
	buildGradlePath, err := tests.ReplaceTemplateVariables(srcBuildFile, "")
	if err != nil {
		t.Error(err)
	}

	srcSettingsFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradle", projectName, "settings.gradle")
	_, err = tests.ReplaceTemplateVariables(srcSettingsFile, "")
	if err != nil {
		t.Error(err)
	}

	return buildGradlePath
}
func initGradleTest(t *testing.T) {
	if !*tests.TestGradle {
		t.Skip("Skipping Gradle test. To run Gradle test add the '-test.gradle=true' option.")
	}
	createJfrogHomeConfig(t)
}

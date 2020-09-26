package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	buildinfocmd "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"

	"github.com/stretchr/testify/assert"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands/buildinfo"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

const gradleFlagName = "gradle"

func cleanGradleTest() {
	os.Unsetenv(coreutils.HomeDir)
	deleteSpec := spec.NewBuilder().Pattern(tests.GradleRepo).BuildSpec()
	tests.DeleteFiles(deleteSpec, artifactoryDetails)
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
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo, _ := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")

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
	artifactoryCli.Exec("bp", tests.GradleBuildName, buildNumber)
	buildInfo, _ := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.GradleBuildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")
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

	artifactoryCli.Exec("bp", tests.GradleBuildName, buildNumber)
	buildInfo, _ := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.GradleBuildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")
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
	artifactoryCli.Exec("bp", tests.GradleBuildName, buildNumber)
	buildInfo, _ := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.GradleBuildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 0, 1, ":minimal-example:1.0")
	cleanGradleTest()
}

func runAndValidateGradle(buildGradlePath, configFilePath, buildName, buildNumber string, t *testing.T) {
	runCliWithLegacyBuildtoolsCmd(t, "gradle", "clean artifactoryPublish -b "+buildGradlePath, configFilePath, "--build-name="+buildName, "--build-number="+buildNumber)
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)

	verifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
	verifyExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.GradleRepo+"/*", "build.name="+buildName+";build.number="+buildNumber, t)
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

func TestBuildOfBuildsGradle(t *testing.T) {
	// Initialize
	initGradleTest(t)
	dependencyGradleProject, gradleProject := copyGradleProjectsToOutDir(t)
	buildNumber := "1"

	// Clean failed runs.
	tests.CleanLocalPartialBuildInfo(t, tests.RtBuildOfBuildGradleProject, tests.RtBuildOfBuildGradleDependencyProject, buildNumber)

	//install dependency project.
	oldHomeDir := changeWD(t, dependencyGradleProject)
	runCli(t, "gradle", "clean", "build", "artifactoryPublish", "--build-name="+tests.RtBuildOfBuildGradleDependencyProject, "--build-number="+buildNumber)

	// Install maven project which depends on DependencyGradleProject.
	changeWD(t, gradleProject)
	runCli(t, "gradle", "clean", "build", "artifactoryPublish", "--build-name="+tests.RtBuildOfBuildGradleProject, "--build-number="+buildNumber)

	// Genarete build info for maven project.
	buildConfig := &utils.BuildConfiguration{BuildName: tests.RtBuildOfBuildGradleProject, BuildNumber: buildNumber}
	publishConfig := new(buildinfocmd.Configuration)
	publishCmd := buildinfo.NewBuildPublishCommand().SetBuildConfiguration(buildConfig).SetConfig(publishConfig).SetRtDetails(artifactoryDetails)
	servicesManager, err := utils.CreateServiceManager(artifactoryDetails, false)
	assert.NoError(t, err)
	buildInfo, err := publishCmd.GenerateBuildInfo(servicesManager)
	assert.NoError(t, err)

	// Validate that the dependency got the value of its build.
	dep := buildInfo.Modules[0].Dependencies[0]
	assert.True(t, strings.HasPrefix(dep.Build, tests.RtBuildOfBuildGradleDependencyProject+"/"+buildNumber+"/"))
	idx := strings.LastIndex(dep.Build, "/")
	assert.True(t, len(dep.Build[idx:]) > 3)

	// Cleanup
	assert.NoError(t, os.Chdir(oldHomeDir))
	tests.CleanLocalPartialBuildInfo(t, tests.RtBuildOfBuildGradleProject, tests.RtBuildOfBuildGradleDependencyProject, buildNumber)
	cleanGradleTest()
}

func copyGradleProjectsToOutDir(t *testing.T) (string, string) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	gradleTestdataPath := filepath.Join(wd, "testdata", "gradle", "buildofbuildsprojects")
	assert.NoError(t, fileutils.CopyDir(gradleTestdataPath, tests.Out, true, nil), err)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", "buildofbuilds", tests.GradleConfig)
	createConfigFile(filepath.Join(tests.Out, "greeter", ".jfrog", "projects"), configFilePath, t)
	createConfigFile(filepath.Join(tests.Out, "hello", ".jfrog", "projects"), configFilePath, t)
	createConfigFile(filepath.Join(tests.Out, "hello"), filepath.Join(tests.Out, "hello", "build.gradle"), t)
	return filepath.Join(wd, tests.Out, "greeter"), filepath.Join(wd, tests.Out, "hello")
}

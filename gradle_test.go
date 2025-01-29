package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"

	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	outputFormat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

	"github.com/stretchr/testify/assert"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

const (
	gradleModuleId = ":minimal-example:1.0"
)

func cleanGradleTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	deleteSpec := spec.NewBuilder().Pattern(tests.GradleRepo).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.NoError(t, err)
	tests.CleanFileSystem()
}

func TestGradleBuildConditionalUpload(t *testing.T) {
	initGradleTest(t)
	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)
	execFunc := func() error {
		return runGradleConditionalUploadTest(buildGradlePath)
	}
	testConditionalUpload(t, execFunc, searchSpec, tests.GetGradleDeployedArtifacts()...)
	cleanGradleTest(t)
}

func runGradleConditionalUploadTest(buildGradlePath string) (err error) {
	configFilePath, exists, err := project.GetProjectConfFilePath(project.Gradle)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("no config file was found!")
	}
	buildConfig := build.NewBuildConfiguration("", "", "", "")
	if err = buildConfig.ValidateBuildAndModuleParams(); err != nil {
		return
	}
	printDeploymentView := log.IsStdErrTerminal()
	gradleCmd := gradle.NewGradleCommand().
		SetTasks([]string{"clean", "artifactoryPublish", "-b" + buildGradlePath}).
		SetConfiguration(buildConfig).
		SetXrayScan(true).SetScanOutputFormat(outputFormat.Table).
		SetDetailedSummary(printDeploymentView).SetConfigPath(configFilePath).SetThreads(commonCliUtils.Threads)
	err = commands.Exec(gradleCmd)
	result := gradleCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(gradleCmd.Result(), false, printDeploymentView, false, err)
	return
}

func TestGradleWithDeploymentView(t *testing.T) {
	initGradleTest(t)
	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)
	assertPrintedDeploymentViewFunc, cleanupFunc := initDeploymentViewTest(t)
	defer cleanupFunc()
	assert.NoError(t, runJfrogCliWithoutAssertion("gradle", "clean", "artifactoryPublish", "-b"+buildGradlePath))
	assertPrintedDeploymentViewFunc()
	cleanGradleTest(t)
}

func TestGradleBuildWithServerID(t *testing.T) {
	initGradleTest(t)
	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	buildNumber := "1"
	// Windows compatibility
	buildGradlePath = strings.ReplaceAll(buildGradlePath, `\`, "/")
	runJfrogCli(t, "gradle", "clean", "artifactoryPublish", "-b"+buildGradlePath, "--build-name="+tests.GradleBuildName, "--build-number="+buildNumber)
	clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, serverDetails, t)
	verifyExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.GradleRepo+"/*", "build.name="+tests.GradleBuildName+";build.number="+buildNumber+";build.timestamp="+getBuildTimestamp(tests.GradleBuildName, buildNumber, t), t)
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
	cleanGradleTest(t)
}

func TestGradleBuildWithServerIDAndDetailedSummary(t *testing.T) {
	initGradleTest(t)
	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	buildNumber := "1"
	// Windows compatibility.
	buildGradlePath = strings.ReplaceAll(buildGradlePath, `\`, "/")

	// Test gradle with detailed summary without buildinfo props.
	filteredGradleArgs := []string{"clean", "artifactoryPublish", "-b" + buildGradlePath}
	gradleCmd := gradle.NewGradleCommand().SetConfiguration(new(build.BuildConfiguration)).SetTasks(filteredGradleArgs).SetConfigPath(filepath.Join(destPath, "gradle.yaml")).SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(gradleCmd))
	// Validate sha256
	assert.NotNil(t, gradleCmd.Result())
	if gradleCmd.Result() != nil {
		tests.VerifySha256DetailedSummaryFromResult(t, gradleCmd.Result())
	}

	// Test gradle with detailed summary + buildinfo.
	gradleCmd = gradle.NewGradleCommand().SetConfiguration(build.NewBuildConfiguration(tests.GradleBuildName, buildNumber, "", "")).SetTasks(filteredGradleArgs).SetConfigPath(filepath.Join(destPath, "gradle.yaml")).SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(gradleCmd))
	// Validate sha256
	tests.VerifySha256DetailedSummaryFromResult(t, gradleCmd.Result())

	clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)
	// Validate build info
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, serverDetails, t)
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
	cleanGradleTest(t)
}

func TestGradleBuildWithServerIDWithUsesPlugin(t *testing.T) {
	initGradleTest(t)
	// Create gradle project in a tmp dir
	buildGradlePath := createGradleProject(t, "projectwithplugin")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleServerIDUsesPluginConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	err := os.Rename(filepath.Join(destPath, tests.GradleServerIDUsesPluginConfig), filepath.Join(destPath, "gradle.yaml"))
	assert.NoError(t, err)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	buildName := tests.GradleBuildName
	buildNumber := "1"
	runJfrogCli(t, "gradle", "clean", "artifactoryPublish", "-b"+buildGradlePath, "--build-name="+buildName, "--build-number="+buildNumber)
	changeWD(t, oldHomeDir)
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, serverDetails, t)
	verifyExistInArtifactoryByProps(tests.GetGradleDeployedArtifacts(), tests.GradleRepo+"/*", "build.name="+buildName+";build.number="+buildNumber, t)
	inttestutils.ValidateGeneratedBuildInfoModule(t, buildName, buildNumber, "", []string{gradleModuleId}, buildinfo.Gradle)

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
	cleanGradleTest(t)
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

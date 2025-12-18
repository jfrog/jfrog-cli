package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/gradle"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/stretchr/testify/require"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"

	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	buildinfo "github.com/jfrog/build-info-go/entities"
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
	// Expect 1 dependency (junit:4.7) since the project has source code using JUnit
	validateBuildInfo(buildInfo, t, 1, 1, gradleModuleId, buildinfo.Gradle)
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
	// Expect 1 dependency (junit:4.7) since the project has source code using JUnit
	validateBuildInfo(buildInfo, t, 1, 1, gradleModuleId, buildinfo.Gradle)
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

func TestSetupGradleCommand(t *testing.T) {
	restoreFunc := prepareGradleSetupTest(t)
	defer restoreFunc()
	// Validate that the module does not exist in the cache before running the test.
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	// This module is part of the dependencies in the build.gradle file of the current test project.
	// We want to ensure that it is not exist in the cache before running the build command.
	moduleCacheUrl := serverDetails.ArtifactoryUrl + tests.GradleRemoteRepo + "-cache/com/fasterxml/jackson/core/jackson-core/2.13.2/jackson-core-2.13.2.jar"
	_, _, err = client.GetRemoteFileDetails(moduleCacheUrl, artHttpDetails)
	assert.ErrorContains(t, err, "404")

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, execGo(jfrogCli, "setup", "gradle", "--repo="+tests.GradleRemoteRepo))

	// Run `gradle clean` to resolve the artifact from Artifactory and force it to be downloaded.
	output, err := exec.Command("gradle",
		"clean",
		"build",
		"--info",
		"--refresh-dependencies").CombinedOutput()
	assert.NoError(t, err, fmt.Sprintf("%s\n%q", string(output), err))

	// Validate that the module exists in the cache after running the build command.
	_, res, err := client.GetRemoteFileDetails(moduleCacheUrl, artHttpDetails)
	if assert.NoError(t, err, "Failed to find the artifact in the cache: "+moduleCacheUrl) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
}

// TestGradleBuildWithFlexPack tests Gradle build with JFROG_RUN_NATIVE=true (FlexPack mode)
func TestGradleBuildWithFlexPack(t *testing.T) {
	initGradleTest(t)

	// Check if Gradle is available in the environment
	if _, err := exec.LookPath("gradle"); err != nil {
		t.Skip("Gradle not found in PATH, skipping Gradle FlexPack test")
	}

	buildGradlePath := createGradleProject(t, "gradleproject")
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)

	// Set environment for native FlexPack implementation
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallBack()

	// Windows compatibility
	buildGradlePath = strings.ReplaceAll(buildGradlePath, `\`, "/")

	// Run gradle without config file to trigger FlexPack mode
	err := runJfrogCliWithoutAssertion("gradle", "clean", "build", "-b"+buildGradlePath)
	assert.NoError(t, err)

	cleanGradleTest(t)
}

// TestGradleBuildWithFlexPackBuildInfo tests Gradle build info collection with JFROG_RUN_NATIVE=true
func TestGradleBuildWithFlexPackBuildInfo(t *testing.T) {
	initGradleTest(t)

	// Check if Gradle is available in the environment
	if _, err := exec.LookPath("gradle"); err != nil {
		t.Skip("Gradle not found in PATH, skipping Gradle FlexPack build info test")
	}

	buildGradlePath := createGradleProject(t, "gradleproject")
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)

	buildName := tests.GradleBuildName + "-flexpack"
	buildNumber := "1"

	// Set environment for native FlexPack implementation
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallBack()

	// Windows compatibility
	buildGradlePath = strings.ReplaceAll(buildGradlePath, `\`, "/")

	// Run gradle with build info (FlexPack mode - no config file)
	err := runJfrogCliWithoutAssertion("gradle", "clean", "build", "-b"+buildGradlePath, "--build-name="+buildName, "--build-number="+buildNumber)
	assert.NoError(t, err)

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Validate build info was created with FlexPack dependencies
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}

	// Validate build info structure
	buildInfo := publishedBuildInfo.BuildInfo
	assert.NotEmpty(t, buildInfo.Modules, "Build info should have modules")
	if len(buildInfo.Modules) > 0 {
		module := buildInfo.Modules[0]
		assert.Equal(t, buildinfo.Gradle, module.Type, "Module type should be Gradle")
		assert.NotEmpty(t, module.Id, "Module should have ID")

		// FlexPack should collect dependencies with scopes and checksums.
		assert.NotEmpty(t, module.Dependencies, "FlexPack build info should include dependencies")
		for _, dep := range module.Dependencies {
			assert.NotEmpty(t, dep.Id, "Dependency should have ID")
			// FlexPack should provide checksums
			hasChecksum := dep.Sha1 != "" || dep.Sha256 != "" || dep.Md5 != ""
			assert.True(t, hasChecksum, "Dependency %s should have at least one checksum", dep.Id)
			// FlexPack should provide scopes (Gradle configurations)
			assert.NotEmpty(t, dep.Scopes, "Dependency %s should have scopes/configurations", dep.Id)
		}
	}

	cleanGradleTest(t)
}

// TestGradleBuildWithFlexPackEnvVars tests build info collection using JFROG_CLI_BUILD_NAME and JFROG_CLI_BUILD_NUMBER
func TestGradleBuildWithFlexPackEnvVars(t *testing.T) {
	initGradleTest(t)

	// Check if Gradle is available in the environment
	if _, err := exec.LookPath("gradle"); err != nil {
		t.Skip("Gradle not found in PATH, skipping Gradle FlexPack env vars test")
	}

	buildGradlePath := createGradleProject(t, "gradleproject")
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)

	buildName := tests.GradleBuildName + "-flexpack-env"
	buildNumber := "123"

	// Set environment for native FlexPack implementation
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallBack()

	// Set build name and number via environment variables
	setBuildNameCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_CLI_BUILD_NAME", buildName)
	defer setBuildNameCallback()
	setBuildNumberCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_CLI_BUILD_NUMBER", buildNumber)
	defer setBuildNumberCallback()

	// Windows compatibility
	buildGradlePath = strings.ReplaceAll(buildGradlePath, `\`, "/")

	// Run gradle without explicit build name/number flags - should use env vars
	err := runJfrogCliWithoutAssertion("gradle", "clean", "build", "-b"+buildGradlePath)
	assert.NoError(t, err)

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Validate build info was created with the env var values
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found with env var build name/number")
		return
	}

	// Validate build info has the correct build name and number
	assert.Equal(t, buildName, publishedBuildInfo.BuildInfo.Name, "Build name should match env var")
	assert.Equal(t, buildNumber, publishedBuildInfo.BuildInfo.Number, "Build number should match env var")

	cleanGradleTest(t)
}

// TestGradleBuildWithFlexPackInvalidArgs tests 'jf gradle build' with invalid Gradle arguments
func TestGradleBuildWithFlexPackInvalidArgs(t *testing.T) {
	initGradleTest(t)

	// Check if Gradle is available in the environment
	if _, err := exec.LookPath("gradle"); err != nil {
		t.Skip("Gradle not found in PATH, skipping Gradle FlexPack invalid args test")
	}

	buildGradlePath := createGradleProject(t, "gradleproject")
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)

	// Set environment for native FlexPack implementation
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallBack()

	// Windows compatibility
	buildGradlePath = strings.ReplaceAll(buildGradlePath, `\`, "/")

	// Run gradle with invalid task name - should fail
	err := runJfrogCliWithoutAssertion("gradle", "nonExistentTask", "-b"+buildGradlePath)
	assert.Error(t, err, "Gradle should fail with invalid task name")

	// Run gradle with invalid option
	err = runJfrogCliWithoutAssertion("gradle", "build", "--invalid-option-xyz", "-b"+buildGradlePath)
	assert.Error(t, err, "Gradle should fail with invalid option")

	// This test intentionally fails before any deployment/build-info publish.
	clientTestUtils.UnSetEnvAndAssert(t, coreutils.HomeDir)
	tests.CleanFileSystem()
}

// TestGradleBuildWithFlexPackFallback verifies that gradle falls back to traditional approach
// when JFROG_RUN_NATIVE is not set (covered by existing traditional tests, this is explicit verification)
func TestGradleBuildWithFlexPackFallback(t *testing.T) {
	initGradleTest(t)

	// Explicitly ensure JFROG_RUN_NATIVE is NOT set
	_ = os.Unsetenv("JFROG_RUN_NATIVE")

	buildGradlePath := createGradleProject(t, "gradleproject")
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleConfig)
	destPath := filepath.Join(filepath.Dir(buildGradlePath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)

	buildName := tests.GradleBuildName + "-fallback"
	buildNumber := "1"

	// Windows compatibility
	buildGradlePath = strings.ReplaceAll(buildGradlePath, `\`, "/")

	// Run gradle with config file (traditional approach)
	runJfrogCli(t, "gradle", "clean", "artifactoryPublish", "-b"+buildGradlePath, "--build-name="+buildName, "--build-number="+buildNumber)

	// Validate artifacts were deployed (traditional approach deploys to Artifactory)
	searchSpec, err := tests.CreateSpec(tests.SearchAllGradle)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, serverDetails, t)

	cleanGradleTest(t)
}

// TestGradleHelpCommand verifies 'jf gradle --help' displays correct usage instructions
func TestGradleHelpCommand(t *testing.T) {
	initGradleTest(t)

	// Run gradle help command
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("gradle", "--help")
	// Help command should succeed
	assert.NoError(t, err, "gradle --help should succeed")

	cleanGradleTest(t)
}

// TestGradleBuildWithFlexPackKotlinDSL tests build info collection for Kotlin DSL (build.gradle.kts)
// Note: This test requires a Kotlin DSL project to be available
func TestGradleBuildWithFlexPackKotlinDSL(t *testing.T) {
	initGradleTest(t)

	// Check if Gradle is available in the environment
	if _, err := exec.LookPath("gradle"); err != nil {
		t.Skip("Gradle not found in PATH, skipping Gradle FlexPack Kotlin DSL test")
	}

	// Check if Kotlin DSL project exists
	kotlinProjectPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradle", "kotlinproject")
	if _, err := os.Stat(kotlinProjectPath); os.IsNotExist(err) {
		t.Skip("Kotlin DSL project not found, skipping Kotlin DSL test")
	}

	// Create gradle project with Kotlin DSL
	buildGradlePath := createGradleProjectKotlin(t, "kotlinproject")
	oldHomeDir := changeWD(t, filepath.Dir(buildGradlePath))
	defer clientTestUtils.ChangeDirAndAssert(t, oldHomeDir)

	buildName := tests.GradleBuildName + "-flexpack-kotlin"
	buildNumber := "1"

	// Set environment for native FlexPack implementation
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallBack()

	// Windows compatibility
	buildGradlePath = strings.ReplaceAll(buildGradlePath, `\`, "/")

	// Run gradle build
	err := runJfrogCliWithoutAssertion("gradle", "clean", "build", "-b"+buildGradlePath, "--build-name="+buildName, "--build-number="+buildNumber)
	assert.NoError(t, err)

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Validate build info was created
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}

	// Validate build info structure
	buildInfo := publishedBuildInfo.BuildInfo
	assert.NotEmpty(t, buildInfo.Modules, "Build info should have modules")
	if len(buildInfo.Modules) > 0 {
		module := buildInfo.Modules[0]
		assert.Equal(t, buildinfo.Gradle, module.Type, "Module type should be Gradle")
	}

	cleanGradleTest(t)
}

func createGradleProject(t *testing.T, projectName string) string {
	// Copy the entire project directory including source files
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradle", projectName)
	projectTarget := filepath.Join(tests.Temp, projectName)
	// Use absolute paths to avoid issues when tests change working directory and pass relative -b flags.
	// (e.g. running from tmp/gradleproject and passing -btmp/gradleproject/build.gradle results in duplicated paths)
	var err error
	projectTarget, err = filepath.Abs(projectTarget)
	assert.NoError(t, err)
	err = io.CopyDir(projectSrc, projectTarget, true, nil)
	assert.NoError(t, err)

	// Replace template variables in build.gradle
	srcBuildFile := filepath.Join(projectTarget, "build.gradle")
	buildGradlePath, err := tests.ReplaceTemplateVariables(srcBuildFile, projectTarget)
	assert.NoError(t, err)
	buildGradlePath, err = filepath.Abs(buildGradlePath)
	assert.NoError(t, err)
	assert.FileExists(t, buildGradlePath)
	// Gradle accepts forward slashes on all platforms, and this avoids Windows path edge-cases.
	buildGradlePath = filepath.ToSlash(buildGradlePath)

	// Replace template variables in settings.gradle
	srcSettingsFile := filepath.Join(projectTarget, "settings.gradle")
	settingsPath, err := tests.ReplaceTemplateVariables(srcSettingsFile, projectTarget)
	assert.NoError(t, err)
	settingsPath, err = filepath.Abs(settingsPath)
	assert.NoError(t, err)
	assert.FileExists(t, settingsPath)

	return buildGradlePath
}

// createGradleProjectKotlin creates a Kotlin DSL gradle project for testing
func createGradleProjectKotlin(t *testing.T, projectName string) string {
	// Copy the entire project directory including source files
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradle", projectName)
	projectTarget := filepath.Join(tests.Temp, projectName)
	// Use absolute paths to avoid issues when tests change working directory and pass relative -b flags.
	var err error
	projectTarget, err = filepath.Abs(projectTarget)
	assert.NoError(t, err)
	err = io.CopyDir(projectSrc, projectTarget, true, nil)
	assert.NoError(t, err)

	// Replace template variables in build.gradle.kts
	srcBuildFile := filepath.Join(projectTarget, "build.gradle.kts")
	buildGradlePath, err := tests.ReplaceTemplateVariables(srcBuildFile, projectTarget)
	assert.NoError(t, err)
	buildGradlePath, err = filepath.Abs(buildGradlePath)
	assert.NoError(t, err)
	assert.FileExists(t, buildGradlePath)
	// Gradle accepts forward slashes on all platforms, and this avoids Windows path edge-cases.
	buildGradlePath = filepath.ToSlash(buildGradlePath)

	// Replace template variables in settings.gradle.kts
	srcSettingsFile := filepath.Join(projectTarget, "settings.gradle.kts")
	settingsPath, err := tests.ReplaceTemplateVariables(srcSettingsFile, projectTarget)
	assert.NoError(t, err)
	settingsPath, err = filepath.Abs(settingsPath)
	assert.NoError(t, err)
	assert.FileExists(t, settingsPath)

	return buildGradlePath
}

func initGradleTest(t *testing.T) {
	if !*tests.TestGradle {
		t.Skip("Skipping Gradle test. To run Gradle test add the '-test.gradle=true' option.")
	}
	// Ensure clean state - unset native flag so traditional tests run correctly
	_ = os.Unsetenv("JFROG_RUN_NATIVE")
	createJfrogHomeConfig(t, true)
}

func prepareGradleSetupTest(t *testing.T) func() {
	initGradleTest(t)
	gradleHome := t.TempDir()
	t.Setenv(gradle.UserHomeEnv, gradleHome)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	gradleProjectDir := t.TempDir()
	err = io.CopyDir(filepath.Join(tests.GetTestResourcesPath(), "gradle", "setupcmd"), gradleProjectDir, true, nil)
	require.NoError(t, err)
	assert.NoError(t, os.Chdir(gradleProjectDir))
	restoreDir := clientTestUtils.ChangeDirWithCallback(t, wd, gradleProjectDir)
	return func() {
		restoreDir()
	}
}

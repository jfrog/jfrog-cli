package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/golang"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli/inttestutils"

	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

func TestGoConfigWithModuleNameChange(t *testing.T) {
	_, cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()
	buildNumber := "1"

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")

	prepareGoProject("project1", "", t, true)
	runGo(t, ModuleNameJFrogTest, tests.GoBuildName, buildNumber, 4, 0, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	clientTestUtils.ChangeDirAndAssert(t, wd)
}

func TestGoGetSpecificVersion(t *testing.T) {
	// Init test and prepare Global config
	_, cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()
	buildNumber := "1"
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	prepareGoProject("project1", "", t, true)
	// Build and publish a go project.
	// We do so in order to make sure the rsc.io/quote:v1.5.2 will be available for the get command
	runGo(t, "", tests.GoBuildName, buildNumber, 4, 0, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)

	// Go get one of the known dependencies
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	err = execGo(jfrogCli, "go", "get", "rsc.io/quote@v1.5.2", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = execGo(artifactoryCli, "bp", tests.GoBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.GoBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo

	validateBuildInfo(buildInfo, t, 2, 0, "rsc.io/quote", buildinfo.Go)

	// Cleanup
	clientTestUtils.ChangeDirAndAssert(t, wd)
}

// Test 'go get' with a nested package (a specific directory inside a package) and validate it was cached successfully.
// The whole outer package should be downloaded.
func TestGoGetNestedPackage(t *testing.T) {
	goPath, cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	prepareGoProject("project1", "", t, true)
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")

	// Download 'mockgen', which is a nested package inside 'github.com/golang/mock@v1.4.1'. Then validate it was downloaded correctly.
	err = execGo(jfrogCli, "go", "get", "github.com/golang/mock/mockgen@v1.4.1")
	if err != nil {
		assert.NoError(t, err)
	}
	packageCachePath := filepath.Join(goPath, "pkg", "mod")
	exists, err := fileutils.IsDirExists(filepath.Join(packageCachePath, "github.com/golang/mock@v1.4.1"), false)
	assert.NoError(t, err)
	assert.True(t, exists)
	clientTestUtils.ChangeDirAndAssert(t, wd)
	cleanGoTest(t)
}

// Testing publishing and resolution capabilities for go projects.
// Build first project ->
// Publish first project to Artifactory ->
// Build second project using go resolving from Artifactory - should download project1 as dependency.
func TestGoPublishResolve(t *testing.T) {
	_, cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	project1Path := prepareGoProject("project1", "", t, true)
	clientTestUtils.ChangeDirAndAssert(t, wd)
	project2Path := prepareGoProject("project2", "", t, true)
	clientTestUtils.ChangeDirAndAssert(t, project1Path)

	// Build the first project and download its dependencies from Artifactory
	buildNumber := "1"
	runGo(t, "", tests.GoBuildName, buildNumber, 4, 0, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)

	// Publish the first project to Artifactory
	buildNumber = "2"
	runGo(t, "", tests.GoBuildName, buildNumber, 0, 3, "gp", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "v1.0.0")

	clientTestUtils.ChangeDirAndAssert(t, project2Path)

	// Build the second project and download its dependencies from Artifactory
	err = execGo(artifactoryCli, "go", "build", "--mod=mod")
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Restore workspace
	clientTestUtils.ChangeDirAndAssert(t, wd)
}

func TestGoPublishWithDetailedSummary(t *testing.T) {
	_, cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()

	// Init environment
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	projectPath := prepareGoProject("project1", "", t, true)

	// Publish with detailed summary and buildinfo.
	// Build project
	buildNumber := "1"
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, execGo(jfrogCli, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest))

	// GoPublish with detailed summary without buildinfo.
	goPublishCmd := golang.NewGoPublishCommand()
	goPublishCmd.SetConfigFilePath(filepath.Join(projectPath, ".jfrog", "projects", "go.yaml")).SetBuildConfiguration(new(utils.BuildConfiguration)).SetVersion("v1.0.0").SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(goPublishCmd))
	tests.VerifySha256DetailedSummaryFromResult(t, goPublishCmd.Result())

	// GoPublish with buildinfo configuration
	goPublishCmd.SetBuildConfiguration(utils.NewBuildConfiguration(tests.GoBuildName, buildNumber, ModuleNameJFrogTest, ""))
	assert.NoError(t, commands.Exec(goPublishCmd))
	tests.VerifySha256DetailedSummaryFromResult(t, goPublishCmd.Result())

	// Build publish
	assert.NoError(t, artifactoryCli.Exec("bp", tests.GoBuildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.GoBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, 4, 3, ModuleNameJFrogTest, buildinfo.Go)

	// Restore workspace
	clientTestUtils.ChangeDirAndAssert(t, wd)
}

func TestGoPublishWithDeploymentView(t *testing.T) {
	_, goCleanupFunc := initGoTest(t)
	defer goCleanupFunc()
	assertPrintedDeploymentViewFunc, cleanupFunc := initDeploymentViewTest(t)
	defer cleanupFunc()

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	prepareGoProject("project1", "", t, true)
	jfrogCli := tests.NewJfrogCli(execMain, "jf", "")
	err = execGo(jfrogCli, "gp", "v1.1.1")
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assertPrintedDeploymentViewFunc()

	// Restore workspace
	clientTestUtils.ChangeDirAndAssert(t, wd)
}

func TestGoVcsFallback(t *testing.T) {
	_, cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()

	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	_ = prepareGoProject("vcsfallback", "", t, false)

	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	// Run "go get github.com/octocat/Hello-World" with --no-fallback.
	// This package is not a Go package and therefore we'd expect the command to fail.
	err = execGo(jfrogCli, "go", "get", "github.com/octocat/Hello-World", "--no-fallback")
	assert.Error(t, err)

	// Run "go get github.com/octocat/Hello-World" with the default --no-fallback=false.
	// Eventually, this package should be downloaded from GitHub.
	err = execGo(jfrogCli, "go", "get", "github.com/octocat/Hello-World")
	assert.NoError(t, err)

	clientTestUtils.ChangeDirAndAssert(t, wd)
}

func prepareGoProject(projectName, configDestDir string, t *testing.T, copyDirs bool) string {
	projectPath := createGoProject(t, projectName, copyDirs)
	testdataTarget := filepath.Join(tests.Out, "testdata")
	testdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testdata")
	err := fileutils.CopyDir(testdataSrc, testdataTarget, copyDirs, nil)
	assert.NoError(t, err)
	if configDestDir == "" {
		configDestDir = filepath.Join(projectPath, ".jfrog")
	}
	configFileDir := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", projectName, ".jfrog", "projects")
	_, err = tests.ReplaceTemplateVariables(filepath.Join(configFileDir, "go.yaml"), filepath.Join(configDestDir, "projects"))
	assert.NoError(t, err)
	clientTestUtils.ChangeDirAndAssert(t, projectPath)
	log.Info("Using Go project located at ", projectPath)
	return projectPath
}

func initGoTest(t *testing.T) (tempGoPath string, cleanUp func()) {
	if !*tests.TestGo {
		t.Skip("Skipping go test. To run go test add the '-test.go=true' option.")
	}
	clientTestUtils.SetEnvAndAssert(t, "GONOSUMDB", "github.com/jfrog")
	clientTestUtils.UnSetEnvAndAssert(t, "GOMODCACHE")
	createJfrogHomeConfig(t, true)
	tempGoPath, cleanUpGoPath := createTempGoPath(t)
	return tempGoPath, func() {
		cleanUpGoPath()
		cleanGoTest(t)
	}
}

func cleanGoTest(t *testing.T) {
	clientTestUtils.UnSetEnvAndAssert(t, "GONOSUMDB")
	deleteSpec := spec.NewBuilder().Pattern(tests.GoRepo).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.NoError(t, err)
	cleanTestsHomeEnv()
}

func createTempGoPath(t *testing.T) (tempGoPath string, cleanUp func()) {
	tempDirPath, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	log.Info(fmt.Sprintf("Changing GOPATH to: %s", tempDirPath))
	cleanUpGoPath := clientTestUtils.SetEnvWithCallbackAndAssert(t, "GOPATH", tempDirPath)
	return tempDirPath, func() {
		// Sometimes we don't have permissions to delete Go cache folders, so we tell Go to delete their content and then we just delete the empty folders.
		cleanGoCache(t)
		cleanUpGoPath()
	}
}

func createGoProject(t *testing.T, projectName string, includeDirs bool) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", projectName)
	projectTarget := filepath.Join(tests.Out, projectName)
	err := fileutils.CopyDir(projectSrc, projectTarget, includeDirs, nil)
	assert.NoError(t, err)
	projectTarget, err = filepath.Abs(projectTarget)
	assert.NoError(t, err)
	goModeOriginalPath := filepath.Join(projectTarget, "createGoProject_go.mod_suffix")
	goModeTargetPath := filepath.Join(projectTarget, "go.mod")
	assert.NoError(t, os.Rename(goModeOriginalPath, goModeTargetPath))
	return projectTarget
}

// runGo runs 'jfrog' command with the given args, publishes a build info, validates it and finally deletes it.
func runGo(t *testing.T, module, buildName, buildNumber string, expectedDependencies, expectedArtifacts int, args ...string) {
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	err := execGo(jfrogCli, args...)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = execGo(artifactoryCli, "bp", buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

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

	if module == "" {
		module = "github.com/jfrog/dependency"
	}

	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, module, buildinfo.Go)
}

func execGo(cli *tests.JfrogCli, args ...string) error {
	return cli.WithoutCredentials().Exec(args...)
}

func cleanGoCache(t *testing.T) {
	log.Info("Cleaning go cache by running: 'go clean -modcache'")

	cmd := exec.Command("go", "clean", "-modcache")
	cmd.Env = append(cmd.Env, "GOPATH="+os.Getenv("GOPATH"))
	assert.NoError(t, cmd.Run())
}

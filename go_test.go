package main

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/golang"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

func TestGoConfigWithModuleNameChange(t *testing.T) {
	cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()
	buildNumber := "1"

	wd, err := os.Getwd()
	assert.NoError(t, err)

	prepareGoProject("project1", "", t, true)
	runGo(t, ModuleNameJFrogTest, tests.GoBuildName, buildNumber, 4, 0, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	assert.NoError(t, os.Chdir(wd))
}

func TestGoGetSpecificVersion(t *testing.T) {
	// Init test and prepare Global config
	cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()
	buildNumber := "1"
	wd, err := os.Getwd()
	assert.NoError(t, err)
	prepareGoProject("project1", "", t, true)
	// Build and publish a go project.
	// We do so in order to make sure the rsc.io/quote:v1.5.2 will be available for the get command
	runGo(t, "", tests.GoBuildName, buildNumber, 4, 0, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)

	// Go get one of the known dependencies
	artifactoryGoCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err = execGo(artifactoryGoCli, "go", "get", "rsc.io/quote@v1.5.2", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)
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
	assert.NoError(t, os.Chdir(wd))
}

// Testing publishing and resolution capabilities for go projects.
// Build first project ->
// Publish first project to Artifactory ->
// Build second project using go resolving from Artifactory - should download project1 as dependency.
func TestGoPublishResolve(t *testing.T) {
	cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()
	wd, err := os.Getwd()
	assert.NoError(t, err)
	project1Path := prepareGoProject("project1", "", t, true)
	assert.NoError(t, os.Chdir(wd))
	project2Path := prepareGoProject("project2", "", t, true)
	assert.NoError(t, os.Chdir(project1Path))

	// Build the first project and download its dependencies from Artifactory
	buildNumber := "1"
	runGo(t, "", tests.GoBuildName, buildNumber, 4, 0, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)

	// Publish the first project to Artifactory
	buildNumber = "2"
	runGo(t, "", tests.GoBuildName, buildNumber, 4, 3, "gp", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "v1.0.0")

	assert.NoError(t, os.Chdir(project2Path))

	// Build the second project and download its dependencies from Artifactory
	err = execGo(artifactoryCli, "go", "build", "--mod=mod")
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Restore workspace
	assert.NoError(t, os.Chdir(wd))
}

func TestGoPublishWithDetailedSummary(t *testing.T) {
	cleanUpFunc := initGoTest(t)
	defer cleanUpFunc()

	// Init environment
	wd, err := os.Getwd()
	assert.NoError(t, err)
	projectPath := prepareGoProject("project1", "", t, true)

	// Publish with detailed summary and buildinfo.
	// Build project
	buildNumber := "1"
	artifactoryGoCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	assert.NoError(t, execGo(artifactoryGoCli, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest))

	// GoPublish with detailed summary without buildinfo.
	goPublishCmd := golang.NewGoPublishCommand()
	goPublishCmd.SetConfigFilePath(filepath.Join(projectPath, ".jfrog", "projects", "go.yaml")).SetBuildConfiguration(new(utils.BuildConfiguration)).SetVersion("v1.0.0").SetDetailedSummary(true)
	assert.NoError(t, commands.Exec(goPublishCmd))
	tests.VerifySha256DetailedSummaryFromResult(t, goPublishCmd.Result())

	// GoPublish with buildinfo configuration
	buildConf := utils.BuildConfiguration{BuildName: tests.GoBuildName, BuildNumber: buildNumber, Module: ModuleNameJFrogTest}
	goPublishCmd.SetBuildConfiguration(&buildConf)
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
	assert.NoError(t, os.Chdir(wd))
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
	configFileDir, err = tests.ReplaceTemplateVariables(filepath.Join(configFileDir, "go.yaml"), filepath.Join(configDestDir, "projects"))
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir(projectPath))
	log.Info("Using Go project located at ", projectPath)
	return projectPath
}

func initGoTest(t *testing.T) (cleanUp func()) {
	if !*tests.TestGo {
		t.Skip("Skipping go test. To run go test add the '-test.go=true' option.")
	}
	assert.NoError(t, os.Setenv("GONOSUMDB", "github.com/jfrog"))
	createJfrogHomeConfig(t, true)
	cleanUpGoPath := createTempGoPath(t)
	return func() {
		cleanUpGoPath()
		cleanGoTest(t)
	}
}

func cleanGoTest(t *testing.T) {
	assert.NoError(t, os.Unsetenv("GONOSUMDB"))
	deleteSpec := spec.NewBuilder().Pattern(tests.GoRepo).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.NoError(t, err)
	cleanBuildToolsTest()
}

func createTempGoPath(t *testing.T) (cleanUp func()) {
	tempDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	log.Info(fmt.Sprintf("Changing GOPATH to: %s", tempDirPath))
	cleanUpGoPath := setEnvVar(t, "GOPATH", tempDirPath)
	return func() {
		cleanUpGoPath()
		assert.NoError(t, fileutils.RemoveTempDir(tempDirPath))
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

// runGo runs 'jfrog rt' command with the given args, publishes a build info, validates it and finally deletes it.
func runGo(t *testing.T, module, buildName, buildNumber string, expectedDependencies, expectedArtifacts int, args ...string) {
	artifactoryGoCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err := execGo(artifactoryGoCli, args...)
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

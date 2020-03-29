package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

func initNugetTest(t *testing.T) {
	if !*tests.TestNuget {
		t.Skip("Skipping NuGet test. To run Nuget test add the '-test.nuget=true' option.")
	}

	if !cliutils.IsWindows() {
		t.Skip("Skipping nuget tests, since this is not a Windows machine.")
	}

	// This is due to Artifactory bug, we cant create remote repository with REST API.
	require.True(t, isRepoExist(tests.NugetRemoteRepo), "Create nuget remote repository:", tests.NugetRemoteRepo, "in order to run nuget tests")
	createJfrogHomeConfig(t)
}

func TestNugetResolve(t *testing.T) {
	initNugetTest(t)
	projects := []struct {
		name                 string
		project              string
		moduleId             string
		args                 []string
		expectedDependencies int
	}{
		{"packagesconfigwithoutmodulechnage", "packagesconfig", "packagesconfig", []string{"nuget", "restore", tests.NugetRemoteRepo, "--build-name=" + tests.NugetBuildName}, 6},
		{"packagesconfigwithmodulechnage", "packagesconfig", ModuleNameJFrogTest, []string{"nuget", "restore", tests.NugetRemoteRepo, "--build-name=" + tests.NugetBuildName, "--module=" + ModuleNameJFrogTest}, 6},
		{"referencewithoutmodulechnage", "reference", "reference", []string{"nuget", "restore", tests.NugetRemoteRepo, "--build-name=" + tests.NugetBuildName}, 6},
		{"referencewithmodulechnage", "reference", ModuleNameJFrogTest, []string{"nuget", "restore", tests.NugetRemoteRepo, "--build-name=" + tests.NugetBuildName, "--module=" + ModuleNameJFrogTest}, 6},
	}
	for buildNumber, test := range projects {
		t.Run(test.project, func(t *testing.T) {
			testNugetCmd(t, createNugetProject(t, test.project), strconv.Itoa(buildNumber), test.moduleId, test.expectedDependencies, test.args, false)
		})
	}
	cleanBuildToolsTest()
}

func TestNativeNugetResolve(t *testing.T) {
	initNugetTest(t)
	projects := []struct {
		name                 string
		project              string
		moduleId             string
		args                 []string
		expectedDependencies int
	}{
		{"packagesconfigwithoutmodulechnage", "packagesconfig", "packagesconfig", []string{"nuget", "restore", "--build-name=" + tests.NugetBuildName}, 6},
		{"packagesconfigwithmodulechnage", "packagesconfig", ModuleNameJFrogTest, []string{"nuget", "restore", "--build-name=" + tests.NugetBuildName, "--module=" + ModuleNameJFrogTest}, 6},
		{"referencewithoutmodulechnage", "reference", "reference", []string{"nuget", "restore", "--build-name=" + tests.NugetBuildName}, 6},
		{"referencewithmodulechnage", "reference", ModuleNameJFrogTest, []string{"nuget", "restore", "--build-name=" + tests.NugetBuildName, "--module=" + ModuleNameJFrogTest}, 6},
	}
	for buildNumber, test := range projects {
		projectPath := createNugetProject(t, test.project)
		err := createConfigFileForTest([]string{projectPath}, tests.NugetRemoteRepo, "", t, utils.Nuget, false)
		assert.NoError(t, err)
		t.Run(test.project, func(t *testing.T) {
			testNugetCmd(t, projectPath, strconv.Itoa(buildNumber), test.moduleId, test.expectedDependencies, test.args, true)
		})
	}
	cleanBuildToolsTest()
}

func createNugetProject(t *testing.T, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "nuget", projectName)
	projectTarget := filepath.Join(tests.Out, projectName)
	err := fileutils.CreateDirIfNotExist(projectTarget)
	assert.NoError(t, err)

	files, err := fileutils.ListFiles(projectSrc, false)
	assert.NoError(t, err)

	for _, v := range files {
		err = fileutils.CopyFile(projectTarget, v)
		assert.NoError(t, err)
	}
	return projectTarget
}

func TestNuGetWithGlobalConfig(t *testing.T) {
	initNugetTest(t)
	projectPath := createNugetProject(t, "packagesconfig")
	jfrogHomeDir, err := cliutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	err = createConfigFileForTest([]string{jfrogHomeDir}, tests.NugetRemoteRepo, "", t, utils.Nuget, true)
	assert.NoError(t, err)
	testNugetCmd(t, projectPath, "1", "packagesconfig", 6, []string{"nuget", "restore", "--build-name=" + tests.NugetBuildName}, true)

	cleanBuildToolsTest()
}

func testNugetCmd(t *testing.T, projectPath, buildNumber, module string, expectedDependencies int, args []string, native bool) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(projectPath)
	assert.NoError(t, err)
	args = append(args, "--build-number="+buildNumber)
	if native {
		runNuGet(t, args...)
	} else {
		artifactoryCli.Exec(args...)
	}
	artifactoryCli.Exec("bp", tests.NugetBuildName, buildNumber)

	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.NugetBuildName, buildNumber, t, artHttpDetails)
	require.NotEmpty(t, buildInfo.Modules, "Nuget build info was not generated correctly, no modules were created.")
	assert.Len(t, buildInfo.Modules[0].Dependencies, expectedDependencies, "Incorrect number of artifacts found in the build-info")
	assert.Equal(t, module, buildInfo.Modules[0].Id, "Unexpected module name")
	assert.NoError(t, os.Chdir(wd))

	// cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.NugetBuildName, artHttpDetails)
}

func runNuGet(t *testing.T, args ...string) {
	artifactoryNuGetCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err := artifactoryNuGetCli.Exec(args...)
	assert.NoError(t, err)
}

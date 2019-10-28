package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jfrog/jfrog-cli-go/inttestutils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
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
	if !isRepoExist(tests.NugetRemoteRepo) {
		t.Error("Create nuget remote repository:", tests.NugetRemoteRepo, "in order to run nuget tests")
		t.FailNow()
	}
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
			testNugetCmd(t, createNugetProject(t, test.project), strconv.Itoa(buildNumber), test.moduleId, test.expectedDependencies, test.args)
		})
	}
	cleanBuildToolsTest()
}

func createNugetProject(t *testing.T, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "nuget", projectName)
	projectTarget := filepath.Join(tests.Out, projectName)
	err := fileutils.CreateDirIfNotExist(projectTarget)
	if err != nil {
		t.Error(err)
	}

	files, err := fileutils.ListFiles(projectSrc, false)
	if err != nil {
		t.Error(err)
	}

	for _, v := range files {
		err = fileutils.CopyFile(projectTarget, v)
		if err != nil {
			t.Error(err)
		}
	}
	return projectTarget
}

func testNugetCmd(t *testing.T, projectPath, buildNumber, module string, expectedDependencies int, args []string) {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	err = os.Chdir(projectPath)
	if err != nil {
		t.Error(err)
	}
	args = append(args, "--build-number="+buildNumber)
	artifactoryCli.Exec(args...)
	artifactoryCli.Exec("bp", tests.NugetBuildName, buildNumber)

	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.NugetBuildName, buildNumber, t, artHttpDetails)
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		t.Error("Nuget build info was not generated correctly, no modules were created.")
	}

	if expectedDependencies != len(buildInfo.Modules[0].Dependencies) {
		t.Error("Incorrect number of artifacts found in the build-info, expected:", expectedDependencies, " Found:", len(buildInfo.Modules[0].Dependencies))
	}

	if module != buildInfo.Modules[0].Id {
		t.Error(fmt.Errorf("Expected module name %s, got %s", module, buildInfo.Modules[0].Id))
	}

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	// cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.NugetBuildName, artHttpDetails)
	cleanBuildToolsTest()
}

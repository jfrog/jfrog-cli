package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoetryInstall(t *testing.T) {
	// Init poetry test.
	initPoetryTest(t)
	tests.SkipKnownFailingTest(t)
	// Populate cli config with 'default' server.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		assert.NoError(t, os.Setenv(coreutils.HomeDir, oldHomeDir))
		assert.NoError(t, fileutils.RemoveTempDir(newHomeDir))
	}()

	// Create test cases.
	allTests := []struct {
		name                string
		project             string
		outputFolder        string
		moduleId            string
		args                []string
		cleanAfterExecution bool
	}{
		{"poetry", "poetryproject", "cli-poetry-build", "cli-poetry-build:0.1.0", []string{"poetry", "install", "--build-name=" + tests.PoetryBuildName}, true},
	}

	// Run test cases.
	for buildNumber, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			testPoetryCmd(t, createPoetryProject(t, test.outputFolder, test.project), strconv.Itoa(buildNumber), test.moduleId, test.args)
			if test.cleanAfterExecution {
				// cleanup
				inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PoetryBuildName, artHttpDetails)
			}
		})
	}
	tests.CleanFileSystem()
}

func testPoetryCmd(t *testing.T, projectPath, buildNumber, module string, args []string) {
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	args = append(args, "--build-number="+buildNumber)

	jfrogCli := tests.NewJfrogCli(execMain, "jf", "")
	err = jfrogCli.WithoutCredentials().Exec(args...)
	if err != nil {
		assert.Fail(t, "Failed executing poetry install command", err.Error())
		return
	}

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.PoetryBuildName, buildNumber, "", []string{module}, buildinfo.Python)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.PoetryBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PoetryBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}

	buildInfo := publishedBuildInfo.BuildInfo
	require.NotEmpty(t, buildInfo.Modules, "Poetry build info was not generated correctly, no modules were created.")
	assert.Len(t, buildInfo.Modules[0].Dependencies, 3, "Incorrect number of artifacts found in the build-info")
	assert.Equal(t, module, buildInfo.Modules[0].Id, "Unexpected module name")
	assertPoetryDependenciesRequestedBy(t, buildInfo.Modules[0], module)
}

func assertPoetryDependenciesRequestedBy(t *testing.T, module buildinfo.Module, moduleName string) {
	for _, dependency := range module.Dependencies {
		switch dependency.Id {
		case "toml:0.10.2", "pexpect:4.8.0":
			assert.EqualValues(t, [][]string{{moduleName}}, dependency.RequestedBy)
		case "ptyprocess:0.7.0":
			assert.EqualValues(t, [][]string{{"pexpect:4.8.0", moduleName}}, dependency.RequestedBy)
		default:
			assert.Fail(t, "Unexpected dependency "+dependency.Id)
		}
	}
}

func createPoetryProject(t *testing.T, outFolder, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "poetry", projectName)
	projectTarget := filepath.Join(tests.Out, outFolder+"-poetry-"+projectName)
	err := fileutils.CreateDirIfNotExist(projectTarget)
	assert.NoError(t, err)

	// Copy poetry project
	err = fileutils.CopyDir(projectSrc, projectTarget, true, nil)
	assert.NoError(t, err)

	// Copy poetry-config file.
	configSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "poetry", "poetry.yaml")
	configTarget := filepath.Join(projectTarget, ".jfrog", "projects")
	_, err = tests.ReplaceTemplateVariables(configSrc, configTarget)
	assert.NoError(t, err)
	return projectTarget
}

func initPoetryTest(t *testing.T) {
	if !*tests.TestPoetry {
		t.Skip("Skipping Poetry test. To run Poetry test add the '-test.poetry=true' option.")
	}
	require.True(t, isRepoExist(tests.PoetryRemoteRepo), tests.PoetryRemoteRepo+" test repository doesn't exist.")
	require.True(t, isRepoExist(tests.PoetryVirtualRepo), tests.PoetryVirtualRepo+" test repository doesn't exist.")
}

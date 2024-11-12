package main

import (
	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli-security/commands/audit/sca/python"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipInstallNativeSyntax(t *testing.T) {
	testPipInstall(t, false)
}

// Deprecated
func TestPipInstallLegacy(t *testing.T) {
	testPipInstall(t, true)
}

func testPipInstall(t *testing.T, isLegacy bool) {
	// Init pip.
	initPipTest(t)

	// Populate cli config with 'default' server.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	// Create test cases.
	allTests := []struct {
		name                 string
		project              string
		outputFolder         string
		moduleId             string
		args                 []string
		expectedDependencies int
	}{
		{"setuppy", "setuppyproject", "setuppy", "jfrog-python-example:1.0", []string{".", "--no-cache-dir", "--force-reinstall", "--build-name=" + tests.PipBuildName}, 3},
		{"setuppy-verbose", "setuppyproject", "setuppy-verbose", "jfrog-python-example:1.0", []string{".", "--no-cache-dir", "--force-reinstall", "-v", "--build-name=" + tests.PipBuildName}, 3},
		{"setuppy-with-module", "setuppyproject", "setuppy-with-module", "setuppy-with-module", []string{".", "--no-cache-dir", "--force-reinstall", "--build-name=" + tests.PipBuildName, "--module=setuppy-with-module"}, 3},
		{"requirements", "requirementsproject", "requirements", tests.PipBuildName, []string{"-r", "requirements.txt", "--no-cache-dir", "--force-reinstall", "--build-name=" + tests.PipBuildName}, 5},
		{"requirements-verbose", "requirementsproject", "requirements-verbose", tests.PipBuildName, []string{"-r", "requirements.txt", "--no-cache-dir", "--force-reinstall", "-v", "--build-name=" + tests.PipBuildName}, 5},
		{"requirements-use-cache", "requirementsproject", "requirements-verbose", "requirements-verbose-use-cache", []string{"-r", "requirements.txt", "--module=requirements-verbose-use-cache", "--build-name=" + tests.PipBuildName}, 5},
	}

	// Run test cases.
	for buildNumber, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			cleanVirtualEnv, err := prepareVirtualEnv(t)
			assert.NoError(t, err)

			if isLegacy {
				test.args = append([]string{"rt", "pip-install"}, test.args...)
			} else {
				test.args = append([]string{"pip", "install"}, test.args...)
			}
			testPipCmd(t, createPipProject(t, test.outputFolder, test.project), strconv.Itoa(buildNumber), test.moduleId, test.expectedDependencies, test.args)

			// cleanup
			cleanVirtualEnv()
			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PipBuildName, artHttpDetails)
		})
	}
	tests.CleanFileSystem()
}

func prepareVirtualEnv(t *testing.T) (func(), error) {
	// Create temp directory
	tmpDir, removeTempDir := coretests.CreateTempDirWithCallbackAndAssert(t)

	// Change current working directory to the temp directory
	currentDir, err := os.Getwd()
	if err != nil {
		return removeTempDir, err
	}
	restoreCwd := clientTestUtils.ChangeDirWithCallback(t, currentDir, tmpDir)
	defer restoreCwd()

	// Create virtual environment
	restorePathEnv, err := python.SetPipVirtualEnvPath()
	if err != nil {
		return removeTempDir, err
	}
	// Set cache dir
	unSetEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "PIP_CACHE_DIR", filepath.Join(tmpDir, "cache"))
	return func() {
		removeTempDir()
		assert.NoError(t, restorePathEnv())
		unSetEnvCallback()
	}, err
}

func testPipCmd(t *testing.T, projectPath, buildNumber, module string, expectedDependencies int, args []string) {
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	args = append(args, "--build-number="+buildNumber)

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec(args...)
	if err != nil {
		assert.Fail(t, "Failed executing pip install command", err.Error())
		return
	}

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.PipBuildName, buildNumber, "", []string{module}, buildinfo.Python)
	assert.NoError(t, artifactoryCli.Exec("bp", tests.PipBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PipBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	require.NotEmpty(t, buildInfo.Modules, "Pip build info was not generated correctly, no modules were created.")
	assert.Len(t, buildInfo.Modules[0].Dependencies, expectedDependencies, "Incorrect number of dependencies found in the build-info")
	assert.Equal(t, module, buildInfo.Modules[0].Id, "Unexpected module name")
	assertDependenciesRequestedByAndChecksums(t, buildInfo.Modules[0], module)
}

func assertDependenciesRequestedByAndChecksums(t *testing.T, module buildinfo.Module, moduleName string) {
	for _, dependency := range module.Dependencies {
		assertDependencyChecksums(t, dependency.Checksum)
		switch dependency.Id {
		case "pyyaml:5.1.2", "nltk:3.4.5", "macholib:1.11":
			assert.EqualValues(t, [][]string{{moduleName}}, dependency.RequestedBy)
		case "six:1.16.0":
			assert.EqualValues(t, [][]string{{"nltk:3.4.5", moduleName}}, dependency.RequestedBy)
		default:
			// Altgraph version can change
			if assert.Contains(t, dependency.Id, "altgraph") {
				assert.EqualValues(t, [][]string{{"macholib:1.11", moduleName}}, dependency.RequestedBy)
			} else {
				assert.Fail(t, "Unexpected dependency "+dependency.Id)
			}
		}
	}
}

func assertDependencyChecksums(t *testing.T, checksum buildinfo.Checksum) {
	if assert.NotEmpty(t, checksum) {
		assert.NotEmpty(t, checksum.Md5)
		assert.NotEmpty(t, checksum.Sha1)
		assert.NotEmpty(t, checksum.Sha256)
	}
}

func createPipProject(t *testing.T, outFolder, projectName string) string {
	return createPypiProject(t, outFolder, projectName, "pip")
}

func createPypiProject(t *testing.T, outFolder, projectName, projectSrcDir string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), projectSrcDir, projectName)
	projectTarget := filepath.Join(tests.Out, outFolder+"-"+projectName)
	err := fileutils.CreateDirIfNotExist(projectTarget)
	assert.NoError(t, err)

	// Copy pip-installation file.
	err = biutils.CopyDir(projectSrc, projectTarget, true, nil)
	assert.NoError(t, err)

	// Copy pip-config file.
	configSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), projectSrcDir, "pip.yaml")
	configTarget := filepath.Join(projectTarget, ".jfrog", "projects")
	_, err = tests.ReplaceTemplateVariables(configSrc, configTarget)
	assert.NoError(t, err)
	return projectTarget
}

func initPipTest(t *testing.T) {
	if !*tests.TestPip {
		t.Skip("Skipping Pip test. To run Pip test add the '-test.pip=true' option.")
	}
	require.True(t, isRepoExist(tests.PypiLocalRepo), "Pypi test local repository doesn't exist.")
	require.True(t, isRepoExist(tests.PypiRemoteRepo), "Pypi test remote repository doesn't exist.")
	require.True(t, isRepoExist(tests.PypiVirtualRepo), "Pypi test virtual repository doesn't exist.")
}

func TestTwine(t *testing.T) {
	// Init pip.
	initPipTest(t)

	// Populate cli config with 'default' server.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	// Create test cases.
	allTests := []struct {
		name              string
		project           string
		outputFolder      string
		expectedModuleId  string
		args              []string
		expectedArtifacts int
	}{
		{"twine", "pyproject", "twine", "jfrog-python-example:1.0", []string{}, 2},
		{"twine-with-module", "pyproject", "twine-with-module", "twine-with-module", []string{"--module=twine-with-module"}, 2},
	}

	// Run test cases.
	for testNumber, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			cleanVirtualEnv, err := prepareVirtualEnv(t)
			assert.NoError(t, err)

			buildNumber := strconv.Itoa(100 + testNumber)
			test.args = append([]string{"twine", "upload", "dist/*", "--build-name=" + tests.PipBuildName, "--build-number=" + buildNumber}, test.args...)
			testTwineCmd(t, createPypiProject(t, test.outputFolder, test.project, "twine"), buildNumber, test.expectedModuleId, test.expectedArtifacts, test.args)

			// cleanup
			cleanVirtualEnv()
			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PipBuildName, artHttpDetails)
		})
	}
}

func testTwineCmd(t *testing.T, projectPath, buildNumber, expectedModuleId string, expectedArtifacts int, args []string) {
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec(args...)
	if err != nil {
		assert.Fail(t, "Failed executing twine upload command", err.Error())
		return
	}

	assert.NoError(t, artifactoryCli.Exec("bp", tests.PipBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PipBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	require.Len(t, buildInfo.Modules, 1)
	twineModule := buildInfo.Modules[0]
	assert.Equal(t, buildinfo.Python, twineModule.Type)
	assert.Len(t, twineModule.Artifacts, expectedArtifacts)
	assert.Equal(t, expectedModuleId, twineModule.Id)
}

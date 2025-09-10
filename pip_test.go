package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli-security/sca/bom/buildinfo/technologies/python"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

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

func TestPipDepsCacheOutput(t *testing.T) {
	// Init pip.
	initPipTest(t)

	// Populate cli config with 'default' server.
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	cleanVirtualEnv, err := prepareVirtualEnv(t)
	assert.NoError(t, err)
	defer cleanVirtualEnv()

	// Use the new test project with requirements.txt and expected deps.cache.json
	projectPath := createPipProject(t, "pip-deps-cache-test", "depscachetest")
	defer func() { assert.NoError(t, fileutils.RemoveTempDir(projectPath)) }()

	// Change to project directory
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Run pip install with JFrog CLI
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("pip", "install", "-r", "requirements.txt", "--no-cache-dir", "--force-reinstall", "--build-name="+tests.PipBuildName, "--build-number=999")
	assert.NoError(t, err)

	// Read the generated deps.cache.json
	generatedCacheFile := filepath.Join(".jfrog", "projects", "deps.cache.json")
	assert.FileExists(t, generatedCacheFile)
	generatedContent, err := os.ReadFile(generatedCacheFile)
	assert.NoError(t, err)

	// Read the expected deps.cache.json
	expectedContent, err := os.ReadFile("expected_deps_cache.json")
	assert.NoError(t, err)

	// Parse both JSON files for comparison
	var generatedCache, expectedCache struct {
		Version      int                             `json:"version"`
		Dependencies map[string]buildinfo.Dependency `json:"dependencies"`
	}
	err = json.Unmarshal(generatedContent, &generatedCache)
	assert.NoError(t, err)
	err = json.Unmarshal(expectedContent, &expectedCache)
	assert.NoError(t, err)

	// Compare the dependencies
	assert.Equal(t, len(expectedCache.Dependencies), len(generatedCache.Dependencies),
		"Number of dependencies should match")

	// Verify each package mapping
	for pkgName, expectedDep := range expectedCache.Dependencies {
		generatedDep, exists := generatedCache.Dependencies[pkgName]
		assert.True(t, exists, "Package %s not found in generated cache", pkgName)

		// Compare wheel file names
		assert.Equal(t, expectedDep.Id, generatedDep.Id,
			"Package %s has incorrect wheel file. Expected: %s, Got: %s",
			pkgName, expectedDep.Id, generatedDep.Id)

		// Compare checksums
		assert.Equal(t, expectedDep.Checksum.Sha1, generatedDep.Checksum.Sha1,
			"Package %s SHA1 mismatch", pkgName)
		assert.Equal(t, expectedDep.Checksum.Sha256, generatedDep.Checksum.Sha256,
			"Package %s SHA256 mismatch", pkgName)
		assert.Equal(t, expectedDep.Checksum.Md5, generatedDep.Checksum.Md5,
			"Package %s MD5 mismatch", pkgName)
	}

	// Specifically verify the fix - each package should map to its own wheel
	alembicDep := generatedCache.Dependencies["alembic"]
	assert.Contains(t, alembicDep.Id, "alembic",
		"alembic should map to its own wheel file")

	beautifulsoup4Dep := generatedCache.Dependencies["beautifulsoup4"]
	assert.Contains(t, beautifulsoup4Dep.Id, "beautifulsoup4",
		"beautifulsoup4 should map to its own wheel file")
	assert.NotContains(t, beautifulsoup4Dep.Id, "alembic",
		"beautifulsoup4 should NOT have alembic's wheel file")
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
		// Note: Sub-dependency versions may vary because root dependencies can specify version ranges (e.g., >=1.0.0) for their sub-dependencies.
		switch dependency.Id {
		case "pyyaml:5.1.2", "nltk:3.4.5", "macholib:1.11":
			assert.EqualValues(t, [][]string{{moduleName}}, dependency.RequestedBy)
		case "six:1.17.0":
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
		{"twine-columns-env-long-filename", "pyprojectlongfilename", "twine-with-long-filename",
			"twine-with-long-filename", []string{"--module=twine-with-long-filename"}, 2},
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

func TestSetupPipCommand(t *testing.T) {
	if !*tests.TestPip {
		t.Skip("Skipping Pip test. To run Pip test add the '-test.pip=true' option.")
	}
	createJfrogHomeConfig(t, true)
	// Set custom pip.conf file.
	t.Setenv("PIP_CONFIG_FILE", filepath.Join(t.TempDir(), "pip.conf"))

	// Validate that the package does not exist in the cache before running the test.
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	packageCacheUrl := serverDetails.ArtifactoryUrl + tests.PypiRemoteRepo + "-cache/54/16/12b82f791c7f50ddec566873d5bdd245baa1491bac11d15ffb98aecc8f8b/pefile-2024.8.26-py3-none-any.whl"

	_, _, err = client.GetRemoteFileDetails(packageCacheUrl, artHttpDetails)
	assert.ErrorContains(t, err, "404")

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	require.NoError(t, execGo(jfrogCli, "setup", "pip", "--repo="+tests.PypiRemoteRepo))

	// Run 'pip install' to resolve the package from Artifactory and force it to be cached.
	output, err := exec.Command("pip", "install", "--target", t.TempDir(), "--no-cache-dir", "pefile==2024.8.26").CombinedOutput()
	assert.NoError(t, err, fmt.Sprintf("%s\n%q", string(output), err))

	// Validate that the package exists in the cache after running the test.
	_, res, err := client.GetRemoteFileDetails(packageCacheUrl, artHttpDetails)
	if assert.NoError(t, err, "Failed to find the package in the cache: "+packageCacheUrl) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
}

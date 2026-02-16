package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	buildinfo "github.com/jfrog/build-info-go/entities"
	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoetryInstallNativeSyntax(t *testing.T) {
	testPoetryInstall(t, false, false)
}

func TestPoetryInstallNativeFlexPack(t *testing.T) {
	testPoetryInstall(t, false, true)
}

// Deprecated - Test legacy syntax for backward compatibility
func TestPoetryInstallLegacy(t *testing.T) {
	testPoetryInstall(t, true, false)
}

func testPoetryInstall(t *testing.T, isLegacy bool, useFlexPack bool) {
	// Init Poetry test
	initPoetryTest(t)

	// Set environment for FlexPack implementation if requested
	var setEnvCallback func()
	if useFlexPack {
		setEnvCallback = clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	}
	defer func() {
		if setEnvCallback != nil {
			setEnvCallback()
		}
	}()

	// Populate cli config with 'default' server
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	// Create test cases for Poetry FlexPack implementation
	allTests := []struct {
		name                 string
		project              string
		outputFolder         string
		moduleId             string
		args                 []string
		expectedDependencies int
	}{
		// Basic Poetry install tests
		{"poetry-basic", "poetryproject", "poetry-basic", "poetry-basic", []string{".", "--build-name=" + tests.PoetryBuildName}, 3},
		{"poetry-verbose", "poetryproject", "poetry-verbose", "poetry-verbose", []string{".", "-v", "--build-name=" + tests.PoetryBuildName}, 3},
		{"poetry-with-module", "poetryproject", "poetry-with-module", "poetry-with-module", []string{".", "--build-name=" + tests.PoetryBuildName, "--module=poetry-with-module"}, 3},

		// Poetry with dev dependencies
		{"poetry-with-dev", "poetryproject", "poetry-with-dev", "poetry-with-dev", []string{".", "--with=dev", "--build-name=" + tests.PoetryBuildName}, 5},

		// Poetry with specific groups
		{"poetry-without-dev", "poetryproject", "poetry-without-dev", "poetry-without-dev", []string{".", "--without=dev", "--build-name=" + tests.PoetryBuildName}, 2},
	}

	// Run tests
	for buildNumber, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			buildNumberStr := strconv.Itoa(buildNumber + 1)

			if isLegacy {
				// Legacy syntax (if still supported)
				test.args = append([]string{"rt", "poetry-install"}, test.args...)
			} else {
				// Native FlexPack syntax
				test.args = append([]string{"poetry", "install"}, test.args...)
			}

			testPoetryCmd(t, createPoetryProject(t, test.outputFolder, test.project), buildNumberStr, test.moduleId, test.expectedDependencies, test.args)
		})

		// Clean up build info
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PoetryBuildName, artHttpDetails)
	}
}

func testPoetryCmd(t *testing.T, projectPath, buildNumber, module string, expectedDependencies int, args []string) {
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current directory")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Poetry cache directory to avoid conflicts
	tmpDir, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	unSetEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "POETRY_CACHE_DIR", filepath.Join(tmpDir, "cache"))
	defer unSetEnvCallback()

	// Run JFrog CLI command
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec(args...))

	// Validate build info was created
	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.PoetryBuildName, buildNumber, "", []string{module}, buildinfo.Python)

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.PoetryBuildName, buildNumber))

	// Get published build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PoetryBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}

	// Validate build info content
	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	assert.Len(t, buildInfoModules, 1)
	assert.Equal(t, module, buildInfoModules[0].Id)
	assert.Len(t, buildInfoModules[0].Dependencies, expectedDependencies)

	// Validate that dependencies have checksums (FlexPack feature)
	for _, dep := range buildInfoModules[0].Dependencies {
		if dep.Type == "pypi" {
			// Main dependencies should have checksums
			assert.NotEmpty(t, dep.Sha1, "SHA1 checksum should be present for %s", dep.Id)
		}
	}
}

func TestPoetryPublish(t *testing.T) {
	initPoetryTest(t)

	// Populate cli config with 'default' server
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	// Test cases for Poetry publish
	allTests := []struct {
		name              string
		project           string
		outputFolder      string
		expectedModuleId  string
		expectedArtifacts int
		args              []string
	}{
		{"poetry-publish", "poetryproject", "poetry-publish", "my-poetry-project:0.1.0", 2, []string{"--repository=" + tests.PoetryLocalRepo}},
		{"poetry-publish-with-module", "poetryproject", "poetry-publish-with-module", "poetry-publish-module", 2, []string{"--repository=" + tests.PoetryLocalRepo, "--module=poetry-publish-module"}},
	}

	for buildNumber, test := range allTests {
		t.Run(test.name, func(t *testing.T) {
			buildNumberStr := strconv.Itoa(buildNumber + 1)
			test.args = append([]string{"poetry", "publish", "--build-name=" + tests.PoetryBuildName, "--build-number=" + buildNumberStr}, test.args...)
			testPoetryPublishCmd(t, createPoetryProject(t, test.outputFolder, test.project), buildNumberStr, test.expectedModuleId, test.expectedArtifacts, test.args)
		})

		// Clean up
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.PoetryBuildName, artHttpDetails)
	}
}

func testPoetryPublishCmd(t *testing.T, projectPath, buildNumber, expectedModuleId string, expectedArtifacts int, args []string) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Build the package first
	buildCmd := exec.Command("poetry", "build")
	buildCmd.Dir = projectPath
	assert.NoError(t, buildCmd.Run(), "Failed to build Poetry package")

	// Run publish command
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec(args...))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.PoetryBuildName, buildNumber))

	// Validate artifacts have build properties (like npm/maven/gradle do)
	validatePoetryPublishProperties(t, tests.PoetryLocalRepo, tests.PoetryBuildName, buildNumber)

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.PoetryBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}

	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	assert.Len(t, buildInfoModules, 1)
	assert.Equal(t, expectedModuleId, buildInfoModules[0].Id)
	assert.Len(t, buildInfoModules[0].Artifacts, expectedArtifacts)

	// Validate artifacts have checksums
	for _, artifact := range buildInfoModules[0].Artifacts {
		assert.NotEmpty(t, artifact.Sha1, "SHA1 checksum should be present for artifact %s", artifact.Name)
		assert.NotEmpty(t, artifact.Sha256, "SHA256 checksum should be present for artifact %s", artifact.Name)
	}
}

// validatePoetryPublishProperties validates that Poetry artifacts have build properties set
func validatePoetryPublishProperties(t *testing.T, repo, buildName, buildNumber string) {
	expectedProps := fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", buildName, buildNumber)

	// Search for Poetry artifacts with build properties
	verifyExistInArtifactoryByProps([]string{}, repo+"/*", expectedProps, t)
}

// TestPoetryPublishTraditionalFlowWithBuildInfo tests the traditional flow with build-info flags
// Ensures traditional flow publishes artifacts when --build-name and --build-number are provided
// This validates the fix for bug introduced in v2.79.0 where FlexPack code caused early return
func TestPoetryPublishTraditionalFlowWithBuildInfo(t *testing.T) {
	initPoetryTest(t)

	// Populate cli config with 'default' server
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	// Ensure JFROG_RUN_NATIVE is NOT set (test traditional flow)
	assert.NoError(t, os.Unsetenv("JFROG_RUN_NATIVE"))

	buildName := "poetry-traditional-buildinfo-test"
	buildNumber := "1"
	projectPath := createPoetryProject(t, "traditional-buildinfo-test", "poetryproject")

	// Change to project directory
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Build the package
	buildCmd := exec.Command("poetry", "build")
	buildCmd.Dir = projectPath
	assert.NoError(t, buildCmd.Run(), "Failed to build Poetry package")

	// Publish with build-info flags (this would fail in v2.79.0-v2.82.0)
	args := []string{
		"poetry", "publish",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
		"--repository=" + tests.PoetryLocalRepo,
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec(args...))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// CRITICAL: Validate artifacts were uploaded
	// In buggy versions (v2.79.0-v2.82.0), this would fail with 0 artifacts
	validatePoetryPublishProperties(t, tests.PoetryLocalRepo, buildName, buildNumber)

	// Validate build info has artifacts
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info was expected to be found")

	// Validate 2 artifacts exist (.whl and .tar.gz)
	assert.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	assert.Len(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts, 2,
		"Expected 2 artifacts (.whl and .tar.gz), validates fix for v2.79.0-v2.82.0 bug")

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestPoetryPublishFlexPackFlow tests the FlexPack flow for poetry publish
// Ensures FlexPack flow continues to work correctly with JFROG_RUN_NATIVE=true
func TestPoetryPublishFlexPackFlow(t *testing.T) {
	initPoetryTest(t)

	// Set JFROG_RUN_NATIVE=true for FlexPack flow
	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallback()

	// Populate cli config with 'default' server
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	buildName := "poetry-flexpack-test"
	buildNumber := "1"
	projectPath := createPoetryProject(t, "flexpack-test", "poetryproject")

	// Change to project directory
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Build the package
	buildCmd := exec.Command("poetry", "build")
	buildCmd.Dir = projectPath
	assert.NoError(t, buildCmd.Run(), "Failed to build Poetry package")

	// Publish with FlexPack flow
	args := []string{
		"poetry", "publish",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
		"--repository=" + tests.PoetryLocalRepo,
	}

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec(args...))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Validate artifacts were uploaded
	validatePoetryPublishProperties(t, tests.PoetryLocalRepo, buildName, buildNumber)

	// Validate build info has artifacts
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info was expected to be found")

	// Validate 2 artifacts exist
	assert.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	assert.Len(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts, 2,
		"Expected 2 artifacts (.whl and .tar.gz) in FlexPack flow")

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestPoetryPublishBothFlowsComparison tests both traditional and FlexPack flows
// Ensures both flows produce the same results (feature parity)
func TestPoetryPublishBothFlowsComparison(t *testing.T) {
	initPoetryTest(t)

	// Populate cli config with 'default' server
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	flows := []struct {
		name      string
		useNative bool
		buildName string
	}{
		{"Traditional", false, "poetry-flow-traditional"},
		{"FlexPack", true, "poetry-flow-flexpack"},
	}

	for _, flow := range flows {
		t.Run(flow.name, func(t *testing.T) {
			// Set up environment for FlexPack if needed
			if flow.useNative {
				setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
				defer setEnvCallback()
			} else {
				assert.NoError(t, os.Unsetenv("JFROG_RUN_NATIVE"))
			}

			buildNumber := "1"
			projectPath := createPoetryProject(t, "flow-comparison-"+flow.name, "poetryproject")

			// Change to project directory
			wd, err := os.Getwd()
			assert.NoError(t, err)
			chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
			defer chdirCallback()

			// Build the package
			buildCmd := exec.Command("poetry", "build")
			buildCmd.Dir = projectPath
			assert.NoError(t, buildCmd.Run(), "Failed to build Poetry package")

			// Publish
			args := []string{
				"poetry", "publish",
				"--build-name=" + flow.buildName,
				"--build-number=" + buildNumber,
				"--repository=" + tests.PoetryLocalRepo,
			}

			jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
			assert.NoError(t, jfrogCli.Exec(args...))

			// Publish build info
			assert.NoError(t, artifactoryCli.Exec("bp", flow.buildName, buildNumber))

			// Validate artifacts were uploaded (both flows should upload same artifacts)
			validatePoetryPublishProperties(t, tests.PoetryLocalRepo, flow.buildName, buildNumber)

			// Validate build info
			publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, flow.buildName, buildNumber)
			assert.NoError(t, err)
			assert.True(t, found, "build info was expected to be found for %s flow", flow.name)

			// Both flows should produce 2 artifacts
			assert.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
			assert.Len(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts, 2,
				"%s flow should upload 2 artifacts (.whl and .tar.gz)", flow.name)

			// Validate artifacts have checksums
			for _, artifact := range publishedBuildInfo.BuildInfo.Modules[0].Artifacts {
				assert.NotEmpty(t, artifact.Sha1, "SHA1 checksum should be present in %s flow", flow.name)
				assert.NotEmpty(t, artifact.Sha256, "SHA256 checksum should be present in %s flow", flow.name)
			}

			// Clean up
			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, flow.buildName, artHttpDetails)
		})
	}
}

func TestPoetryBuildInfoCollection(t *testing.T) {
	// Test the FlexPack build info collection functionality
	initPoetryTest(t)

	// Set environment for FlexPack implementation
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallBack()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	projectPath := createPoetryProject(t, "poetry-buildinfo", "poetryproject")

	// Test build info collection with FlexPack
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Run Poetry install with build info
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{"poetry", "install", "--build-name=poetry-buildinfo-test", "--build-number=1"}
	assert.NoError(t, jfrogCli.Exec(args...))

	// Check that FlexPack cache was created
	cacheDir := filepath.Join(projectPath, ".jfrog", "projects")
	assert.DirExists(t, cacheDir, "FlexPack cache directory should exist")

	// Check for Poetry-specific cache files
	poetryCacheFile := filepath.Join(cacheDir, "poetry-deps.cache.json")
	if fileutils.IsPathExists(poetryCacheFile, false) {
		t.Logf("Poetry dependencies cache found at: %s", poetryCacheFile)
	}

	// Publish and validate build info
	assert.NoError(t, artifactoryCli.Exec("bp", "poetry-buildinfo-test", "1"))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, "poetry-buildinfo-test", "1")
	assert.NoError(t, err)
	assert.True(t, found, "build info should be found")

	if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		t.Logf("Build info module: %s with %d dependencies", module.Id, len(module.Dependencies))

		// Validate FlexPack-specific features
		for _, dep := range module.Dependencies {
			if dep.Type == "pypi" {
				t.Logf("Dependency: %s (SHA1: %s)", dep.Id, dep.Sha1)
			}
		}
	}

	// Clean up
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, "poetry-buildinfo-test", artHttpDetails)
}

func createPoetryProject(t *testing.T, outputFolder, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "poetry", projectName)
	tmpDir, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()

	projectPath := filepath.Join(tmpDir, outputFolder)
	assert.NoError(t, biutils.CopyDir(projectSrc, projectPath, true, nil))

	// No need for poetry.yaml - FlexPack uses native Poetry configuration

	return projectPath
}

func initPoetryTest(t *testing.T) {
	if !*tests.TestPoetry {
		t.Skip("Skipping Poetry test. To run Poetry test add the '-test.poetry=true' option.")
	}
	require.True(t, isRepoExist(tests.PoetryRemoteRepo), "Poetry test remote repository doesn't exist.")
	require.True(t, isRepoExist(tests.PoetryVirtualRepo), "Poetry test virtual repository doesn't exist.")
}

func TestSetupPoetryCommand(t *testing.T) {
	if !*tests.TestPoetry {
		t.Skip("Skipping Poetry test. To run Poetry test add the '-test.poetry=true' option.")
	}

	// Test the setup command for Poetry (if implemented)
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	// Verify that packages can be resolved from the remote repository
	packageUrl := serverDetails.ArtifactoryUrl + tests.PoetryRemoteRepo + "-cache/packages/f9/9b/335f9764261e915ed497fcdeb11df5dfd6f7bf257d4a6a2a686d80da4d54/requests-2.32.3-py3-none-any.whl"

	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)
	resp, _, _, err := client.SendGet(packageUrl, true, httputils.HttpClientDetails{}, "")
	if err == nil && resp.StatusCode == http.StatusOK {
		t.Log("Poetry remote repository is accessible and contains packages")
	}

	// Test setup command (when implemented)
	// This would configure Poetry to use JFrog repositories
	// require.NoError(t, execGo(jfrogCli, "setup", "poetry", "--repo="+tests.PoetryRemoteRepo))
}

func TestPoetryFlexPackFeatures(t *testing.T) {
	// Test specific FlexPack features for Poetry
	initPoetryTest(t)

	// Set environment for FlexPack implementation
	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, "JFROG_RUN_NATIVE", "true")
	defer setEnvCallBack()

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	projectPath := createPoetryProject(t, "poetry-flexpack", "poetryproject")

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Test 1: Dependency caching
	t.Run("dependency-caching", func(t *testing.T) {
		jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

		// First run - should build cache
		start := time.Now().UnixMilli()
		args := []string{"poetry", "install", "--build-name=poetry-cache-test", "--build-number=1"}
		assert.NoError(t, jfrogCli.Exec(args...))
		firstRunTime := time.Now().UnixMilli() - start

		// Second run - should use cache (faster)
		start = time.Now().UnixMilli()
		args = []string{"poetry", "install", "--build-name=poetry-cache-test", "--build-number=2"}
		assert.NoError(t, jfrogCli.Exec(args...))
		secondRunTime := time.Now().UnixMilli() - start

		t.Logf("First run: %dms, Second run: %dms", firstRunTime, secondRunTime)
		// Second run should be faster due to caching (allow some variance)
		if secondRunTime < firstRunTime {
			t.Log("Caching improved performance")
		}

		// Clean up
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, "poetry-cache-test", artHttpDetails)
	})

	// Test 2: Checksum calculation
	t.Run("checksum-calculation", func(t *testing.T) {
		jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
		args := []string{"poetry", "install", "--build-name=poetry-checksum-test", "--build-number=1"}
		assert.NoError(t, jfrogCli.Exec(args...))

		// Publish and check checksums
		assert.NoError(t, artifactoryCli.Exec("bp", "poetry-checksum-test", "1"))

		publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, "poetry-checksum-test", "1")
		assert.NoError(t, err)
		assert.True(t, found)

		if found && len(publishedBuildInfo.BuildInfo.Modules) > 0 {
			checksumCount := 0
			for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
				if dep.Sha1 != "" {
					checksumCount++
					t.Logf("Dependency %s has checksum: %s", dep.Id, dep.Sha1)
				}
			}
			assert.Greater(t, checksumCount, 0, "At least some dependencies should have checksums")
		}

		// Clean up
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, "poetry-checksum-test", artHttpDetails)
	})
}

// TestPoetryBuildPublishWithCIVcsProps tests that CI VCS properties are set on Poetry artifacts
// when running build-publish in a CI environment (GitHub Actions).
// Poetry relies on build-publish to set CI VCS properties via batch AQL query.
func TestPoetryBuildPublishWithCIVcsProps(t *testing.T) {
	initPoetryTest(t)

	buildName := tests.PoetryBuildName + "-civcs"
	buildNumber := "1"

	// Setup GitHub Actions environment (uses real env vars on CI, mock values locally)
	cleanupEnv, actualOrg, actualRepo := tests.SetupGitHubActionsEnv(t)
	defer cleanupEnv()

	// Populate cli config with 'default' server
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}()

	// Clean old build
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	projectPath := createPoetryProject(t, "poetry-civcs", "poetryproject")

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Build the package first
	buildCmd := exec.Command("poetry", "build")
	buildCmd.Dir = projectPath
	assert.NoError(t, buildCmd.Run(), "Failed to build Poetry package")

	// Run poetry publish with build info collection
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"poetry", "publish",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
		"--repository=" + tests.PoetryLocalRepo,
	}
	err = jfrogCli.Exec(args...)
	assert.NoError(t, err, "Failed executing poetry publish command")

	// Publish build info - should set CI VCS props on artifacts
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Get the published build info to find artifact paths
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "Build info was not found")

	// Create service manager for getting artifact properties
	serviceManager, err := utils.CreateServiceManager(serverDetails, 3, 1000, false)
	assert.NoError(t, err)

	// Verify VCS properties on each artifact from build info
	// Use same fallback logic as CI VCS: OriginalDeploymentRepo + Path, or Path directly
	artifactCount := 0
	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			var fullPath string
			switch {
			case artifact.OriginalDeploymentRepo != "":
				fullPath = artifact.OriginalDeploymentRepo + "/" + artifact.Path
			case artifact.Path != "":
				fullPath = artifact.Path
			default:
				continue // Skip artifacts without any path info
			}

			props, err := serviceManager.GetItemProps(fullPath)
			assert.NoError(t, err, "Failed to get properties for artifact: %s", fullPath)
			assert.NotNil(t, props, "Properties are nil for artifact: %s", fullPath)

			// Validate VCS properties
			assert.Contains(t, props.Properties, "vcs.provider", "Missing vcs.provider on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.provider"], "github", "Wrong vcs.provider on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.org", "Missing vcs.org on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.org"], actualOrg, "Wrong vcs.org on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.repo", "Missing vcs.repo on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.repo"], actualRepo, "Wrong vcs.repo on %s", artifact.Name)

			artifactCount++
		}
	}

	assert.Greater(t, artifactCount, 0, "No artifacts were validated for CI VCS properties")
}

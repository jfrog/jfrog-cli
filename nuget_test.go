package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/utils/ioutils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/io"

	dotnetUtils "github.com/jfrog/build-info-go/build/utils/dotnet"
	buildInfo "github.com/jfrog/build-info-go/entities"
	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func initNugetTest(t *testing.T) {
	if !*tests.TestNuget {
		t.Skip("Skipping NuGet test. To run Nuget test add the '-test.nuget=true' option.")
	}
	createJfrogHomeConfig(t, true)
}

type testDescriptor struct {
	name                 string
	project              string
	args                 []string
	expectedModules      []string
	expectedDependencies []int
}

func TestNugetResolve(t *testing.T) {
	uniqueNugetTests := []testDescriptor{
		{"nugetargswithspaces", "packagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "-PackagesDirectory", "./packages dir with spaces"}, []string{"packagesconfig"}, []int{6}},
		{"packagesconfigwithoutmodulechange", "packagesconfig", []string{dotnetUtils.Nuget.String(), "restore"}, []string{"packagesconfig"}, []int{6}},
		{"packagesconfigwithmodulechange", "packagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
		{"packagesconfigwithconfigpath", "packagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "./packages.config", "-SolutionDirectory", "."}, []string{"packagesconfig"}, []int{6}},
		{"multipackagesconfigwithoutmodulechange", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore"}, []string{"proj1", "proj2", "proj3"}, []int{4, 3, 2}},
		{"multipackagesconfigwithmodulechange", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{8}},
		{"multipackagesconfigwithslnPath", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "./multipackagesconfig.sln"}, []string{"proj1", "proj2", "proj3"}, []int{4, 3, 2}},
		{"multipackagesconfigsingleprojectdir", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "./proj2/", "-SolutionDirectory", "."}, []string{"proj2"}, []int{3}},
		{"multipackagesconfigsingleprojectconfig", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "./proj1/packages.config", "-SolutionDirectory", "."}, []string{"proj1"}, []int{4}},
	}
	testNativeNugetDotnetResolve(t, uniqueNugetTests, tests.NuGetBuildName, project.Nuget)
}

func TestDotnetResolve(t *testing.T) {
	uniqueDotnetTests := []testDescriptor{
		{"dotnetargswithspaces", "multireference", []string{dotnetUtils.DotnetCore.String(), "restore", "src/multireference.proj1/", "--packages", "./packages dir with spaces"}, []string{"proj1"}, []int{5}},
		{"multireferencesingleprojectdir", "multireference", []string{dotnetUtils.DotnetCore.String(), "restore", "src/multireference.proj1/"}, []string{"proj1"}, []int{5}},
	}
	testNativeNugetDotnetResolve(t, uniqueDotnetTests, tests.DotnetBuildName, project.Dotnet)
}

func testNativeNugetDotnetResolve(t *testing.T, uniqueTests []testDescriptor, buildName string, projectType project.ProjectType) {
	initNugetTest(t)
	testDescriptors := append(slices.Clone(uniqueTests), []testDescriptor{
		{"referencewithoutmodulechange", "reference", []string{projectType.String(), "restore"}, []string{"reference"}, []int{6}},
		{"referencewithmodulechange", "reference", []string{projectType.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
		{"multireferencewithoutmodulechange", "multireference", []string{projectType.String(), "restore"}, []string{"proj1", "proj2"}, []int{5, 3}},
		{"multireferencewithmodulechange", "multireference", []string{projectType.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
		{"multireferencewithslnpath", "multireference", []string{projectType.String(), "restore", "src/multireference.sln"}, []string{"proj1", "proj2"}, []int{5, 3}},
		{"multireferencewithslndir", "multireference", []string{projectType.String(), "restore", "src/"}, []string{"proj1", "proj2"}, []int{5, 3}},
		{"multireferencesingleprojectcsproj", "multireference", []string{projectType.String(), "restore", "src/multireference.proj2/proj2.csproj"}, []string{"proj2"}, []int{3}},
		{"sln_and_proj_different_locations", "differentlocations", []string{projectType.String(), "restore", "solutions/differentlocations.sln"}, []string{"proj1", "proj2"}, []int{5, 3}},
	}...)
	for buildNumber, test := range testDescriptors {
		projectPath := createNugetProject(t, test.project)
		err := createConfigFileForTest([]string{projectPath}, tests.NugetRemoteRepo, "", t, projectType, false)
		if err != nil {
			assert.NoError(t, err)
			return
		}
		t.Run(test.name, func(t *testing.T) {
			testNugetCmd(t, projectPath, buildName, strconv.Itoa(buildNumber), test.expectedModules, test.args, test.expectedDependencies)
		})
	}
	cleanTestsHomeEnv()
}

func createNugetProject(t *testing.T, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "nuget", projectName)
	projectTarget := filepath.Join(tests.Out, projectName)
	err := fileutils.CreateDirIfNotExist(projectTarget)
	assert.NoError(t, err)

	err = biutils.CopyDir(projectSrc, projectTarget, true, nil)
	assert.NoError(t, err)
	return projectTarget
}

func TestNuGetWithGlobalConfig(t *testing.T) {
	initNugetTest(t)
	projectPath := createNugetProject(t, "packagesconfig")
	jfrogHomeDir, err := coreutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	err = createConfigFileForTest([]string{jfrogHomeDir}, tests.NugetRemoteRepo, "", t, project.Nuget, true)
	assert.NoError(t, err)
	// allow insecure connection for testings to work with localhost server
	testNugetCmd(t, projectPath, tests.NuGetBuildName, "1", []string{"packagesconfig"}, []string{"nuget", "restore"}, []int{6})

	cleanTestsHomeEnv()
}

func testNugetCmd(t *testing.T, projectPath, buildName, buildNumber string, expectedModule, args []string, expectedDependencies []int) {
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	allowInsecureConnectionForTests(&args)
	args = append(args, "--build-name="+buildName, "--build-number="+buildNumber)

	err = runNuGet(t, args...)
	if err != nil {
		return
	}
	inttestutils.ValidateGeneratedBuildInfoModule(t, buildName, buildNumber, "", expectedModule, buildInfo.Nuget)
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	bi := publishedBuildInfo.BuildInfo
	require.NotEmpty(t, bi.Modules, buildName+" build info was not generated correctly, no modules were created.")

	for i, module := range bi.Modules {
		assert.Equal(t, expectedModule[i], bi.Modules[i].Id, "Unexpected module name")
		assert.Len(t, module.Dependencies, expectedDependencies[i], "Incorrect number of artifacts found in the build-info")
		if strings.HasSuffix(projectPath, "multipackagesconfig") {
			assertNugetMultiPackagesConfigDependencies(t, module, expectedModule[i])
		} else {
			assertNugetDependencies(t, module, expectedModule[i])
		}
	}
	chdirCallback()

	// cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// Add allow insecure connection for testings to work with localhost server
func allowInsecureConnectionForTests(args *[]string) {
	*args = append(*args, "--allow-insecure-connections")
}

func assertNugetDependencies(t *testing.T, module buildInfo.Module, moduleName string) {
	for _, dependency := range module.Dependencies {
		switch dependency.Id {
		case "Microsoft.Web.Xdt:2.1.0", "Microsoft.Web.Xdt:2.1.1":
			assert.EqualValues(t, [][]string{{"NuGet.Core:2.14.0", moduleName}}, dependency.RequestedBy)
		case "popper.js:1.12.9", "jQuery:3.0.0":
			assert.EqualValues(t, [][]string{{"bootstrap:4.0.0", moduleName}}, dependency.RequestedBy)
		case "bootstrap:4.0.0", "Newtonsoft.Json:11.0.2", "NuGet.Core:2.14.0":
			assert.EqualValues(t, [][]string{{moduleName}}, dependency.RequestedBy)
		default:
			assert.Fail(t, "Unexpected dependency "+dependency.Id)
		}
	}
}

func assertNugetMultiPackagesConfigDependencies(t *testing.T, module buildInfo.Module, moduleName string) {
	for _, dependency := range module.Dependencies {
		switch dependency.Id {
		case "Microsoft.Web.Xdt:2.1.0", "Microsoft.Web.Xdt:2.1.1":
			assert.EqualValues(t, [][]string{{"NuGet.Core:2.14.0", moduleName}}, dependency.RequestedBy)
		case "jQuery:3.0.0":
			assert.EqualValues(t, [][]string{{"bootstrap:4.0.0", moduleName}}, dependency.RequestedBy)
		case "bootstrap:4.0.0", "Newtonsoft.Json:11.0.2", "NuGet.Core:2.14.0", "StyleCop.Analyzers:1.0.2",
			"Microsoft.VisualStudio.Setup.Configuration.Interop:1.11.2290", "popper.js:1.12.9":
			assert.EqualValues(t, [][]string{{moduleName}}, dependency.RequestedBy)
		default:
			assert.Fail(t, "Unexpected dependency "+dependency.Id)
		}
	}
}

func runNuGet(t *testing.T, args ...string) error {
	artifactoryNuGetCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	err := artifactoryNuGetCli.Exec(args...)
	assert.NoError(t, err)
	return err
}

type testInitNewConfigDescriptor struct {
	testName          string
	useNugetV2        bool
	expectedSourceUrl string
}

func TestInitNewConfig(t *testing.T) {
	// jfrog-ignore test url
	baseRtUrl := "http://some/url"
	expectedV2Url := baseRtUrl + "/api/nuget"
	expectedV3Url := baseRtUrl + "/api/nuget/v3/index.json"
	testsSuites := []testInitNewConfigDescriptor{
		{"useNugetAddSourceV2", true, expectedV2Url},
		{"useNugetAddSourceV3", false, expectedV3Url},
	}

	for _, test := range testsSuites {
		t.Run(test.testName, func(t *testing.T) {
			runInitNewConfig(t, test, baseRtUrl)
		})
	}
}

func runInitNewConfig(t *testing.T, testSuite testInitNewConfigDescriptor, baseRtUrl string) {
	initNugetTest(t)

	tempDirPath, createTempDirCallback := coreTests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	if tempDirPath == "" {
		return
	}

	params := &dotnet.DotnetCommand{}
	server := &config.ServerDetails{ArtifactoryUrl: baseRtUrl, User: "user", Password: "password"}
	params.SetServerDetails(server).
		SetUseNugetV2(testSuite.useNugetV2).
		SetAllowInsecureConnections(true)

	// Prepare the config file with NuGet authentication
	configFile, err := dotnet.InitNewConfig(tempDirPath, "", server, testSuite.useNugetV2, true)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// #nosec G703 -- configFile path is created by test setup, not user input
	content, err := os.ReadFile(configFile.Name())
	if err != nil {
		assert.NoError(t, err)
		return
	}

	nugetConfig := NugetConfig{}
	err = xml.Unmarshal(content, &nugetConfig)
	if err != nil {
		assert.NoError(t, err, "unmarshalling failed with an error")
		return
	}

	assert.Len(t, nugetConfig.PackageSources, 1)

	for _, packageSource := range nugetConfig.PackageSources {
		assert.Equal(t, dotnet.SourceName, packageSource.Key)
		assert.Equal(t, testSuite.expectedSourceUrl, packageSource.Value)
	}
	assert.Len(t, nugetConfig.PackageSourceCredentials, 1)
	assert.Len(t, nugetConfig.PackageSourceCredentials[0].JFrogCli, 2)
}

type NugetConfig struct {
	XMLName                  xml.Name                   `xml:"configuration"`
	PackageSources           []PackageSources           `xml:"packageSources>add"`
	PackageSourceCredentials []PackageSourceCredentials `xml:"packageSourceCredentials"`
}

type PackageSources struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type PackageSourceCredentials struct {
	JFrogCli []PackageSources `xml:">add"`
}

func TestSetupNugetCommand(t *testing.T) {
	testSetupCommand(t, project.Nuget)
}

func TestSetupDotnetCommand(t *testing.T) {
	testSetupCommand(t, project.Dotnet)
}

func testSetupCommand(t *testing.T, packageManager project.ProjectType) {
	initNugetTest(t)
	restoreFunc := prepareSetupTest(t, packageManager)
	defer func() {
		restoreFunc()
	}()
	// Validate that the package does not exist in the cache before running the test.
	client, err := httpclient.ClientBuilder().Build()
	assert.NoError(t, err)

	// We use different versions of the Nunit package for Nuget and Dotnet to differentiate between the two tests.
	version := "4.0.0"
	if packageManager == project.Dotnet {
		version = "4.1.0"
	}
	moduleCacheUrl := serverDetails.ArtifactoryUrl + tests.NugetRemoteRepo + "-cache/nunit." + version + ".nupkg"
	_, _, err = client.GetRemoteFileDetails(moduleCacheUrl, artHttpDetails)
	assert.ErrorContains(t, err, "404")

	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	require.NoError(t, execGo(jfrogCli, "setup", packageManager.String(), "--repo="+tests.NugetRemoteRepo))

	// Run install some random (Nunit) package to test the setup command.
	var output []byte
	if packageManager == project.Dotnet {
		output, err = exec.Command(packageManager.String(), "add", "package", "NUnit", "--version", version).Output()
	} else {
		output, err = exec.Command(packageManager.String(), "install", "NUnit", "-Version", version, "-OutputDirectory", t.TempDir(), "-NoHttpCache").Output()
	}
	assert.NoError(t, err, fmt.Sprintf("%s\n%q", string(output), err))

	// Validate that the package exists in the cache after running the test.
	// That means that the setup command worked and the package resolved from Artifactory.
	_, res, err := client.GetRemoteFileDetails(moduleCacheUrl, artHttpDetails)
	if assert.NoError(t, err, "Failed to find the artifact in the cache: "+moduleCacheUrl) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
}

func prepareSetupTest(t *testing.T, packageManager project.ProjectType) func() {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)
	var nugetConfigDir string
	switch {
	case io.IsWindows():
		nugetConfigDir = filepath.Join("AppData", "Roaming")
	case packageManager == project.Nuget:
		nugetConfigDir = ".config"
	default:
		nugetConfigDir = ".nuget"
	}

	wd, err := os.Getwd()
	assert.NoError(t, err)
	restoreDir := clientTestUtils.ChangeDirWithCallback(t, wd, t.TempDir())

	// Back up the existing NuGet.config and ensure restoration after the test.
	restoreConfigFunc, err := ioutils.BackupFile(filepath.Join(homeDir, nugetConfigDir, "NuGet", "NuGet.Config"), packageManager.String()+".config.backup")
	require.NoError(t, err)

	if packageManager == project.Dotnet {
		// Dotnet requires creating a new project to install packages.
		assert.NoError(t, exec.Command(packageManager.String(), "new", "console").Run())
		// Clear the NuGet cache to ensure the package is resolved from Artifactory.
		assert.NoError(t, exec.Command(packageManager.String(), "nuget", "locals", "all", "--clear").Run())
		// Remove the default nuget.org source to force resolving the package from Artifactory.
		// We ignore the error since the source might not exist.
		_ = exec.Command(packageManager.String(), "nuget", "remove", "source", "nuget.org").Run()
	} else {
		// Remove the default nuget.org source to force resolving the package from Artifactory.
		// We ignore the error since the source might not exist.
		_ = exec.Command(packageManager.String(), "sources", "remove", "-name", "nuget.org").Run()
	}
	return func() {
		assert.NoError(t, restoreConfigFunc())
		restoreDir()
	}
}

// TestDotnetRequestedByDeterminism verifies that the RequestedBy paths in build-info
// are deterministic across multiple runs. This addresses the bug where dependencies
// could flip between "direct" and "transitive" attribution due to non-deterministic
// map iteration order in Go.
//
// The test uses the "determinism" project which has dependencies that are BOTH
// direct AND transitive (e.g., Newtonsoft.Json is a direct dep and also a transitive
// dep of NuGet.Core). This creates multiple RequestedBy paths that must be
// consistently sorted across runs.
func TestDotnetRequestedByDeterminism(t *testing.T) {
	t.Skip("Skipping test temporarily")
	initNugetTest(t)
	const numRuns = 5
	buildName := tests.DotnetBuildName + "-determinism"

	// Use "determinism" project which has deps that are both direct AND transitive
	projectPath := createNugetProject(t, "determinism")
	err := createConfigFileForTest([]string{projectPath}, tests.NugetRemoteRepo, "", t, project.Dotnet, false)
	require.NoError(t, err)

	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Store RequestedBy maps from each run for comparison
	var allRunsRequestedBy []map[string][][]string

	for i := 1; i <= numRuns; i++ {
		buildNumber := strconv.Itoa(i)
		args := []string{dotnetUtils.DotnetCore.String(), "restore"}
		allowInsecureConnectionForTests(&args)
		args = append(args, "--build-name="+buildName, "--build-number="+buildNumber)

		err := runNuGet(t, args...)
		require.NoError(t, err, "Run %d failed", i)

		// Publish build info
		require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

		// Get published build info
		publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
		require.NoError(t, err)
		require.True(t, found, "Build info not found for run %d", i)

		bi := publishedBuildInfo.BuildInfo
		require.NotEmpty(t, bi.Modules, "No modules in build info for run %d", i)

		// Extract RequestedBy for each dependency
		requestedByMap := make(map[string][][]string)
		for _, module := range bi.Modules {
			for _, dep := range module.Dependencies {
				requestedByMap[dep.Id] = dep.RequestedBy
			}
		}
		allRunsRequestedBy = append(allRunsRequestedBy, requestedByMap)

		// Clean up this build number
		inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	}

	// Compare all runs - they should be identical
	firstRun := allRunsRequestedBy[0]
	for i := 1; i < numRuns; i++ {
		currentRun := allRunsRequestedBy[i]

		// Check that all dependencies have the same RequestedBy paths
		for depId, expectedRequestedBy := range firstRun {
			actualRequestedBy, exists := currentRun[depId]
			assert.True(t, exists, "Dependency %s missing in run %d", depId, i+1)
			assert.Equal(t, expectedRequestedBy, actualRequestedBy,
				"RequestedBy mismatch for %s between run 1 and run %d.\nExpected: %v\nActual: %v",
				depId, i+1, expectedRequestedBy, actualRequestedBy)
		}

		// Also check for any extra dependencies in current run
		for depId := range currentRun {
			_, exists := firstRun[depId]
			assert.True(t, exists, "Extra dependency %s found in run %d but not in run 1", depId, i+1)
		}
	}

	// Verify that dependencies with multiple paths have direct path first
	// Newtonsoft.Json should have direct path ["determinism"] before transitive paths
	for depId, requestedBy := range firstRun {
		if len(requestedBy) > 1 {
			// Direct paths (length 1) should come before transitive paths (length > 1)
			for j := 1; j < len(requestedBy); j++ {
				assert.LessOrEqual(t, len(requestedBy[j-1]), len(requestedBy[j]),
					"RequestedBy paths not sorted by length for %s: path %d has length %d, path %d has length %d",
					depId, j-1, len(requestedBy[j-1]), j, len(requestedBy[j]))
			}
			t.Logf("Dependency %s has %d RequestedBy paths (verified sorted)", depId, len(requestedBy))
		}
	}

	t.Logf("Successfully verified RequestedBy determinism across %d runs", numRuns)
	cleanTestsHomeEnv()
}

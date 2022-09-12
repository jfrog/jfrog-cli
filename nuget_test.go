package main

import (
	"encoding/xml"
	dotnetUtils "github.com/jfrog/build-info-go/build/utils/dotnet"
	buildInfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
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
		{"packagesconfigwithoutmodulechnage", "packagesconfig", []string{dotnetUtils.Nuget.String(), "restore"}, []string{"packagesconfig"}, []int{6}},
		{"packagesconfigwithmodulechnage", "packagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
		{"packagesconfigwithconfigpath", "packagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "./packages.config", "-SolutionDirectory", "."}, []string{"packagesconfig"}, []int{6}},
		{"multipackagesconfigwithoutmodulechnage", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore"}, []string{"proj1", "proj2", "proj3"}, []int{4, 3, 2}},
		{"multipackagesconfigwithmodulechnage", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{8}},
		{"multipackagesconfigwithslnPath", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "./multipackagesconfig.sln"}, []string{"proj1", "proj2", "proj3"}, []int{4, 3, 2}},
		{"multipackagesconfigsingleprojectdir", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "./proj2/", "-SolutionDirectory", "."}, []string{"proj2"}, []int{3}},
		{"multipackagesconfigsingleprojectconfig", "multipackagesconfig", []string{dotnetUtils.Nuget.String(), "restore", "./proj1/packages.config", "-SolutionDirectory", "."}, []string{"proj1"}, []int{4}},
	}
	testNativeNugetDotnetResolve(t, uniqueNugetTests, tests.NuGetBuildName, utils.Nuget)
}

func TestDotnetResolve(t *testing.T) {
	uniqueDotnetTests := []testDescriptor{
		{"dotnetargswithspaces", "multireference", []string{dotnetUtils.DotnetCore.String(), "restore", "src/multireference.proj1/", "--packages", "./packages dir with spaces"}, []string{"proj1"}, []int{5}},
		{"multireferencesingleprojectdir", "multireference", []string{dotnetUtils.DotnetCore.String(), "restore", "src/multireference.proj1/"}, []string{"proj1"}, []int{5}},
	}
	testNativeNugetDotnetResolve(t, uniqueDotnetTests, tests.DotnetBuildName, utils.Dotnet)
}

func testNativeNugetDotnetResolve(t *testing.T, uniqueTests []testDescriptor, buildName string, projectType utils.ProjectType) {
	initNugetTest(t)
	testDescriptors := append(uniqueTests, []testDescriptor{
		{"referencewithoutmodulechnage", "reference", []string{projectType.String(), "restore"}, []string{"reference"}, []int{6}},
		{"referencewithmodulechnage", "reference", []string{projectType.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
		{"multireferencewithoutmodulechnage", "multireference", []string{projectType.String(), "restore"}, []string{"proj1", "proj2"}, []int{5, 3}},
		{"multireferencewithmodulechnage", "multireference", []string{projectType.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
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

	err = fileutils.CopyDir(projectSrc, projectTarget, true, nil)
	assert.NoError(t, err)
	return projectTarget
}

func TestNuGetWithGlobalConfig(t *testing.T) {
	initNugetTest(t)
	projectPath := createNugetProject(t, "packagesconfig")
	jfrogHomeDir, err := coreutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	err = createConfigFileForTest([]string{jfrogHomeDir}, tests.NugetRemoteRepo, "", t, utils.Nuget, true)
	assert.NoError(t, err)
	testNugetCmd(t, projectPath, tests.NuGetBuildName, "1", []string{"packagesconfig"}, []string{"nuget", "restore"}, []int{6})

	cleanTestsHomeEnv()
}

func testNugetCmd(t *testing.T, projectPath, buildName, buildNumber string, expectedModule, args []string, expectedDependencies []int) {
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()
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
	artifactoryNuGetCli := tests.NewJfrogCli(execMain, "jfrog", "")
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
	baseRtUrl := "http://some/url"
	expectedV2Url := baseRtUrl + "/api/nuget"
	expectedV3Url := baseRtUrl + "/api/nuget/v3"
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
	params.SetServerDetails(&config.ServerDetails{ArtifactoryUrl: baseRtUrl, User: "user", Password: "password"}).
		SetUseNugetV2(testSuite.useNugetV2)
	// Prepare the config file with NuGet authentication
	configFile, err := params.InitNewConfig(tempDirPath)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	content, err := ioutil.ReadFile(configFile.Name())
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

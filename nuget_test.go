package main

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jfrog/jfrog-cli-core/utils/coreutils"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	dotnetutils "github.com/jfrog/jfrog-cli-core/artifactory/utils/dotnet"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initNugetTest(t *testing.T) {
	if !*tests.TestNuget {
		t.Skip("Skipping NuGet test. To run Nuget test add the '-test.nuget=true' option.")
	}

	// This is due to Artifactory bug, we cant create remote repository with REST API.
	require.True(t, isRepoExist(tests.NugetRemoteRepo), "Create nuget remote repository:", tests.NugetRemoteRepo, "in order to run nuget tests")
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
	initNugetTest(t)
	projects := []testDescriptor{
		{"packagesconfigwithoutmodulechnage", "packagesconfig", []string{"nuget", "restore", tests.NugetRemoteRepo}, []string{"packagesconfig"}, []int{6}},
		{"packagesconfigwithmodulechnage", "packagesconfig", []string{"nuget", "restore", tests.NugetRemoteRepo, "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
		{"referencewithoutmodulechnage", "reference", []string{"nuget", "restore", tests.NugetRemoteRepo}, []string{"reference"}, []int{6}},
		{"referencewithmodulechnage", "reference", []string{"nuget", "restore", tests.NugetRemoteRepo, "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
	}
	for buildNumber, test := range projects {
		t.Run(test.project, func(t *testing.T) {
			testNugetCmd(t, createNugetProject(t, test.project), tests.DotnetBuildName, strconv.Itoa(buildNumber), test.expectedModules, test.args, test.expectedDependencies, false)
		})
	}
	cleanBuildToolsTest()
}

func TestNativeNugetResolve(t *testing.T) {
	uniqueNugetTests := []testDescriptor{
		{"nugetargswithspaces", "packagesconfig", []string{dotnetutils.Nuget.String(), "restore", "-PackagesDirectory", "./packages dir with spaces"}, []string{"packagesconfig"}, []int{6}},
		{"packagesconfigwithoutmodulechnage", "packagesconfig", []string{dotnetutils.Nuget.String(), "restore"}, []string{"packagesconfig"}, []int{6}},
		{"packagesconfigwithmodulechnage", "packagesconfig", []string{dotnetutils.Nuget.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{6}},
		{"packagesconfigwithconfigpath", "packagesconfig", []string{dotnetutils.Nuget.String(), "restore", "./packages.config", "-SolutionDirectory", "."}, []string{"packagesconfig"}, []int{6}},
		{"multipackagesconfigwithoutmodulechnage", "multipackagesconfig", []string{dotnetutils.Nuget.String(), "restore"}, []string{"proj1", "proj2", "proj3"}, []int{4, 3, 2}},
		{"multipackagesconfigwithmodulechnage", "multipackagesconfig", []string{dotnetutils.Nuget.String(), "restore", "--module=" + ModuleNameJFrogTest}, []string{ModuleNameJFrogTest}, []int{8}},
		{"multipackagesconfigwithslnPath", "multipackagesconfig", []string{dotnetutils.Nuget.String(), "restore", "./multipackagesconfig.sln"}, []string{"proj1", "proj2", "proj3"}, []int{4, 3, 2}},
		{"multipackagesconfigsingleprojectdir", "multipackagesconfig", []string{dotnetutils.Nuget.String(), "restore", "./proj2/", "-SolutionDirectory", "."}, []string{"proj2"}, []int{3}},
		{"multipackagesconfigsingleprojectconfig", "multipackagesconfig", []string{dotnetutils.Nuget.String(), "restore", "./proj1/packages.config", "-SolutionDirectory", "."}, []string{"proj1"}, []int{4}},
	}
	testNativeNugetDotnetResolve(t, uniqueNugetTests, tests.NuGetBuildName, utils.Nuget)
}

func TestDotnetResolve(t *testing.T) {
	uniqueDotnetTests := []testDescriptor{
		{"dotnetargswithspaces", "multireference", []string{dotnetutils.DotnetCore.String(), "restore", "src/multireference.proj1/", "--packages", "./packages dir with spaces"}, []string{"proj1"}, []int{5}},
		{"multireferencesingleprojectdir", "multireference", []string{dotnetutils.DotnetCore.String(), "restore", "src/multireference.proj1/"}, []string{"proj1"}, []int{5}},
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
	}...)
	for buildNumber, test := range testDescriptors {
		projectPath := createNugetProject(t, test.project)
		err := createConfigFileForTest([]string{projectPath}, tests.NugetRemoteRepo, "", t, projectType, false)
		if err != nil {
			assert.NoError(t, err)
			return
		}
		t.Run(test.name, func(t *testing.T) {
			testNugetCmd(t, projectPath, buildName, strconv.Itoa(buildNumber), test.expectedModules, test.args, test.expectedDependencies, true)
		})
	}
	cleanBuildToolsTest()
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
	testNugetCmd(t, projectPath, tests.NuGetBuildName, "1", []string{"packagesconfig"}, []string{"nuget", "restore"}, []int{6}, true)

	cleanBuildToolsTest()
}

func testNugetCmd(t *testing.T, projectPath, buildName, buildNumber string, expectedModule, args []string, expectedDependencies []int, native bool) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(projectPath)
	assert.NoError(t, err)
	args = append(args, "--build-name="+buildName, "--build-number="+buildNumber)
	if native {
		err = runNuGet(t, args...)
		if err != nil {
			return
		}
	} else {
		assert.NoError(t, artifactoryCli.Exec(args...))
	}
	inttestutils.ValidateGeneratedBuildInfoModule(t, buildName, buildNumber, "", expectedModule, buildinfo.Nuget)
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
	buildInfo := publishedBuildInfo.BuildInfo
	require.NotEmpty(t, buildInfo.Modules, buildName+" build info was not generated correctly, no modules were created.")
	for i, module := range buildInfo.Modules {
		assert.Equal(t, expectedModule[i], buildInfo.Modules[i].Id, "Unexpected module name")
		assert.Len(t, module.Dependencies, expectedDependencies[i], "Incorrect number of artifacts found in the build-info")
	}
	assert.NoError(t, os.Chdir(wd))

	// cleanup
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func runNuGet(t *testing.T, args ...string) error {
	artifactoryNuGetCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err := artifactoryNuGetCli.Exec(args...)
	assert.NoError(t, err)
	return err
}

type testInitNewConfigDescriptor struct {
	testName          string
	useNugetAddSource bool
	useNugetV2        bool
	expectedSourceUrl string
}

func TestInitNewConfig(t *testing.T) {
	baseRtUrl := "http://some/url"
	expectedV2Url := baseRtUrl + "/api/nuget"
	expectedV3Url := baseRtUrl + "/api/nuget/v3"
	testsSuites := []testInitNewConfigDescriptor{
		{"useNugetAddSourceV2", true, true, expectedV2Url},
		{"useNugetAddSourceV3", true, false, expectedV3Url},
		{"doNotUseNugetAddSourceV2", false, true, expectedV2Url},
		{"doNotUseNugetAddSourceV3", false, false, expectedV3Url},
	}

	for _, test := range testsSuites {
		t.Run(test.testName, func(t *testing.T) {
			runInitNewConfig(t, test, baseRtUrl)
		})
	}
}

func runInitNewConfig(t *testing.T, testSuite testInitNewConfigDescriptor, baseRtUrl string) {
	initNugetTest(t)

	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	params := &dotnet.DotnetCommand{}
	params.SetServerDetails(&config.ServerDetails{ArtifactoryUrl: baseRtUrl, User: "user", Password: "password"}).
		SetUseNugetAddSource(testSuite.useNugetAddSource).SetUseNugetV2(testSuite.useNugetV2)
	// Prepare the config file with NuGet authentication
	configFile, err := params.InitNewConfig(tempDirPath)
	if err != nil {
		assert.NoError(t, err)
		return
	}

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
	Apikeys                  []PackageSources           `xml:"apikeys>add"`
}

type PackageSources struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type PackageSourceCredentials struct {
	JFrogCli []PackageSources `xml:">add"`
}

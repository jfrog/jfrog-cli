package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jfrog/gocmd/executers"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	_go "github.com/jfrog/jfrog-client-go/artifactory/services/go"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"github.com/stretchr/testify/assert"
)

// Testing build info capabilities for go project.
// Build project using go without Artifactory ->
// Publish dependencies to Artifactory ->
// Publish project to Artifactory->
// Publish and validate build-info
func TestGoBuildInfo(t *testing.T) {
	initGoTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	project1Path := createGoProject(t, "project1", false)
	testdataTarget := filepath.Join(tests.Out, "testdata")
	testdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testdata")
	assert.NoError(t, fileutils.CopyDir(testdataSrc, testdataTarget, true, nil))
	assert.NoError(t, os.Chdir(project1Path))
	defer os.Chdir(wd)

	log.Info("Using Go project located at ", project1Path)

	// 1. Download dependencies.
	// 2. Publish build-info.
	// 3. Validate the total count of dependencies added to the build-info.
	buildNumber := "1"

	err = execGo(t, artifactoryCli, "go", "build --mod=mod", tests.GoRepo, "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.GoBuildName, buildNumber, "", []string{"github.com/jfrog/dependency"}, buildinfo.Go)
	err = execGo(t, artifactoryCli, "bp", tests.GoBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	module := "github.com/jfrog/dependency"
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
	artifactoryVersion, err := artAuth.GetVersion()
	assert.NoError(t, err)

	// Since Artifactory doesn't support info file before version 6.10.0, the artifacts count in the build info is different between versions
	ver := version.NewVersion(artifactoryVersion)
	expectedDependencies := 8
	expectedArtifacts := 2
	if ver.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile) {
		expectedDependencies = 4
		expectedArtifacts = 3
	}
	validateBuildInfo(buildInfo, t, expectedDependencies, 0, module)

	// Now, using a new build number, do the following:
	// 1. Build the project again.
	// 2. Publish the go package.
	// 3. Validate the total count of dependencies and artifacts added to the build-info.
	// 4. Validate that the artifacts are tagged with the build.name and build.number properties.
	buildNumber = "2"

	err = execGo(t, artifactoryCli, "go", "build --mod=mod", tests.GoRepo, "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = execGo(t, artifactoryCli, "gp", tests.GoRepo, "v1.0.0", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--deps=rsc.io/quote:v1.5.2", "--module="+ModuleNameJFrogTest)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = execGo(t, artifactoryCli, "bp", tests.GoBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	publishedBuildInfo, found, err = tests.GetBuildInfo(serverDetails, tests.GoBuildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo = publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, ModuleNameJFrogTest)

	assert.NoError(t, os.Chdir(filepath.Join(wd, "testdata", "go")))

	resultItems := getResultItemsFromArtifactory(tests.SearchGo, t)
	assert.Equal(t, len(buildInfo.Modules[0].Artifacts), len(resultItems), "Incorrect number of artifacts were uploaded")
	propsMap := map[string][]string{
		"build.name":   {tests.GoBuildName},
		"build.number": {buildNumber},
		"go.version":   {"v1.0.0"},
	}
	validateArtifactsProperties(resultItems, t, propsMap)

	assert.NoError(t, os.Chdir(wd))
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.GoBuildName, artHttpDetails)
	cleanGoTest(t)
}

func TestGoConfigWithModuleNameChange(t *testing.T) {
	initGoTest(t)
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(coreutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	prepareGoProject("", t, true)
	runGo(ModuleNameJFrogTest, tests.GoBuildName, buildNumber, t, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	assert.NoError(t, os.Chdir(wd))

	cleanGoTest(t)
}

func TestGoConfigWithoutModuleChange(t *testing.T) {
	initGoTest(t)
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(coreutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	prepareGoProject("", t, true)
	runGo("", tests.GoBuildName, buildNumber, t, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)

	assert.NoError(t, os.Chdir(wd))

	cleanGoTest(t)
}

func TestGoPublishWithConfig(t *testing.T) {
	initGoTest(t)
	buildNumber := "11"
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(coreutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)
	wd, err := os.Getwd()
	defer os.Chdir(wd)
	assert.NoError(t, err)
	prepareGoProject("", t, true)
	artifactoryGoCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err = artifactoryGoCli.Exec("gp", "v1.0.0", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--deps=rsc.io/quote:v1.5.2", "--module="+ModuleNameJFrogTest)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = artifactoryCli.Exec("build-publish", tests.GoBuildName, buildNumber)
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
	validateBuildInfo(buildInfo, t, 0, 3, ModuleNameJFrogTest)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.GoBuildName, artHttpDetails)
	assert.NoError(t, os.Chdir(wd))
	cleanGoTest(t)
}

func TestGoWithGlobalConfig(t *testing.T) {
	initGoTest(t)
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)

	defer os.Setenv(coreutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	prepareGoProject(newHomeDir, t, false)
	runGo(ModuleNameJFrogTest, tests.GoBuildName, buildNumber, t, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
	assert.NoError(t, os.Chdir(wd))

	cleanGoTest(t)
}

func TestGoGetSpecificVersion(t *testing.T) {
	// Init test and prepare Global config
	initGoTest(t)
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(coreutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	prepareGoProject("", t, true)
	// Build and publish a go project.
	// We do so in order to make sure the rsc.io/quote:v1.5.2 will be available for the get command
	runGo("", tests.GoBuildName, buildNumber, t, "go", "build", "--mod=mod", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)

	// Go get one of the known dependencies
	artifactoryGoCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err = execGo(t, artifactoryGoCli, "go", "get", "rsc.io/quote@v1.5.2", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = execGo(t, artifactoryCli, "bp", tests.GoBuildName, buildNumber)
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

	validateBuildInfo(buildInfo, t, 2, 0, "rsc.io/quote")

	// Cleanup
	assert.NoError(t, os.Chdir(wd))
	cleanGoTest(t)
}

func runGo(module, buildName, buildNumber string, t *testing.T, args ...string) {
	artifactoryGoCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err := execGo(t, artifactoryGoCli, args...)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = execGo(t, artifactoryCli, "bp", buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
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
	artifactoryVersion, err := artAuth.GetVersion()
	assert.NoError(t, err)

	// Since Artifactory doesn't support info file before version 6.10.0, the artifacts count in the build info is different between versions
	ver := version.NewVersion(artifactoryVersion)
	if ver.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile) {
		validateBuildInfo(buildInfo, t, 4, 0, module)
	} else {
		validateBuildInfo(buildInfo, t, 8, 0, module)
	}

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

func prepareGoProject(configDestDir string, t *testing.T, copyDirs bool) {
	project1Path := createGoProject(t, "project1", copyDirs)
	testdataTarget := filepath.Join(tests.Out, "testdata")
	testdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testdata")
	err := fileutils.CopyDir(testdataSrc, testdataTarget, copyDirs, nil)
	assert.NoError(t, err)
	if configDestDir == "" {
		configDestDir = filepath.Join(project1Path, ".jfrog")
	}
	configFileDir := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "project1", ".jfrog", "projects")
	configFileDir, err = tests.ReplaceTemplateVariables(filepath.Join(configFileDir, "go.yaml"), filepath.Join(configDestDir, "projects"))
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir(project1Path))
	log.Info("Using Go project located at ", project1Path)
}

// Testing publishing and resolution capabilities for go projects.
// Build first project using go without Artifactory ->
// Publish dependencies to Artifactory ->
// Publish first project to Artifactory ->
// Set go to resolve from Artifactory (set GOPROXY) ->
// Build second project using go resolving from Artifactory, should download project1 as dependency.
func TestGoPublishResolve(t *testing.T) {
	initGoTest(t)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	project1Path := createGoProject(t, "project1", false)
	project2Path := createGoProject(t, "project2", false)
	assert.NoError(t, os.Chdir(project1Path))

	// Download dependencies without Artifactory
	err = execGo(t, artifactoryCli, "go", "build --mod=mod", tests.GoRepo, "--publish-deps=true")
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Publish dependency project to Artifactory
	err = execGo(t, artifactoryCli, "gp", tests.GoRepo, "v1.0.0")
	if err != nil {
		assert.NoError(t, err)
		return
	}

	assert.NoError(t, os.Chdir(project2Path))

	// Build the second project, download dependencies from Artifactory
	err = execGo(t, artifactoryCli, "go", "build --mod=mod", tests.GoRepo)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Restore workspace
	assert.NoError(t, os.Chdir(wd))
	cleanGoTest(t)
}

// Testing the fallback mechanism
// 1. Building a project with a dependency that doesn't exists not in Artifactory and not in VCS.
// 2. The fallback mechanism will try to download from both VCS and Artifactory and then fail with an error
// 3. Testing that the error that is returned is the right error of the fallback.
func TestGoFallback(t *testing.T) {
	initGoTest(t)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	projectBuild := createGoProject(t, "projectbuild", false)

	assert.NoError(t, os.Chdir(projectBuild))

	err = execGo(t, artifactoryCli, "go", "build --mod=mod", tests.GoRepo)
	if err != nil {
		log.Warn(err)
		assert.Contains(t, err.Error(), executers.FailedToRetrieve)
		assert.Contains(t, err.Error(), executers.FromBothArtifactoryAndVcs)
	} else {
		assert.Error(t, err, "Expected error but got success")
		return
	}

	assert.NoError(t, os.Chdir(wd))
	cleanGoTest(t)
}

// Publish also the missing dependencies to Artifactory with the publishDeps flag.
// Checks that the dependency exists in Artifactory.
func TestGoWithPublishDeps(t *testing.T) {
	initGoTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	project1Path := createGoProject(t, "project1", false)
	testdataTarget := filepath.Join(tests.Out, "testdata")
	testdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testdata")
	assert.NoError(t, fileutils.CopyDir(testdataSrc, testdataTarget, true, nil))
	assert.NoError(t, os.Chdir(project1Path))
	defer os.Chdir(wd)

	log.Info("Using Go project located at ", project1Path)
	err = execGo(t, artifactoryCli, "go", "build --mod=mod", tests.GoRepo, "--publish-deps=true")
	if err != nil {
		assert.NoError(t, err)
		return
	}

	assert.NoError(t, os.Chdir(filepath.Join(wd, "testdata", "go")))

	content := downloadModFile(tests.DownloadModOfDependencyGo, wd, "errors", t)
	assert.NotContains(t, string(content), " module github.com/pkg/errors", "Wrong mod content was downloaded")

	assert.NoError(t, os.Chdir(wd))

	cleanGoTest(t)
}

func initGoTest(t *testing.T) {
	if !*tests.TestGo {
		t.Skip("Skipping go test. To run go test add the '-test.go=true' option.")
	}
	assert.NoError(t, os.Setenv("GONOSUMDB", "github.com/jfrog"))
}

func cleanGoTest(t *testing.T) {
	cleanGoCache(t)
	assert.NoError(t, os.Unsetenv("GONOSUMDB"))
	deleteSpec := spec.NewBuilder().Pattern(tests.GoRepo).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.NoError(t, err)
	cleanBuildToolsTest()
}

func uploadGoProject(projectPath string, t *testing.T) {
	assert.NoError(t, os.Chdir(projectPath))
	// Publish project to Artifactory
	err := execGo(t, artifactoryCli, "gp", tests.GoRepo, "v1.0.0")
	if err != nil {
		assert.NoError(t, err)
		return
	}
}

func cleanGoCache(t *testing.T) {
	log.Info("Cleaning go cache by running: 'go clean -modcache'")

	cmd := exec.Command("go", "clean", "-modcache")
	assert.NoError(t, cmd.Run())
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

func downloadModFile(specName, wd, subDir string, t *testing.T) []byte {
	specFile, err := tests.CreateSpec(specName)
	assert.NoError(t, err)

	modDir := filepath.Join(wd, tests.Out, subDir)
	assert.NoError(t, os.MkdirAll(modDir, 0777))

	assert.NoError(t, os.Chdir(modDir))
	assert.NoError(t, artifactoryCli.Exec("download", "--spec="+specFile, "--flat=true"))
	files, err := fileutils.ListFiles(modDir, false)
	assert.NoError(t, err)
	assert.Len(t, files, 1, "Expected to get one mod file")

	content, err := ioutil.ReadFile(files[0])
	assert.NoError(t, err)
	return content
}

func execGo(t *testing.T, cli *tests.JfrogCli, args ...string) error {
	cleanGoCache(t)
	return cli.Exec(args...)
}

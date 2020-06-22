package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jfrog/gocmd/executers"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	_go "github.com/jfrog/jfrog-client-go/artifactory/services/go"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/version"
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
	testsdataTarget := filepath.Join(tests.Out, "testsdata")
	testsdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testsdata")
	assert.NoError(t, fileutils.CopyDir(testsdataSrc, testsdataTarget, true))
	assert.NoError(t, os.Chdir(project1Path))
	defer os.Chdir(wd)

	log.Info("Using Go project located at ", project1Path)

	// 1. Download dependencies.
	// 2. Publish build-info.
	// 3. Validate the total count of dependencies added to the build-info.
	buildNumber := "1"

	artifactoryCli.Exec("go", "build", tests.GoRepo, "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)
	cleanGoCache(t)

	artifactoryCli.Exec("bp", tests.GoBuildName, buildNumber)
	module := "github.com/jfrog/dependency"
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.GoBuildName, buildNumber, t, artHttpDetails)
	artifactoryVersion, err := artAuth.GetVersion()
	assert.NoError(t, err)

	// Since Artifactory doesn't support info file before version 6.10.0, the artifacts count in the build info is different between versions
	version := version.NewVersion(artifactoryVersion)
	expectedDependencies := 8
	expectedArtifacts := 2
	if version.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile) {
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

	artifactoryCli.Exec("go", "build", tests.GoRepo, "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
	cleanGoCache(t)

	artifactoryCli.Exec("gp", tests.GoRepo, "v1.0.0", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--deps=rsc.io/quote:v1.5.2", "--module="+ModuleNameJFrogTest)
	cleanGoCache(t)

	artifactoryCli.Exec("bp", tests.GoBuildName, buildNumber)
	buildInfo = inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.GoBuildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, ModuleNameJFrogTest)

	assert.NoError(t, os.Chdir(filepath.Join(wd, "testsdata", "go")))

	resultItems := getResultItemsFromArtifactory(tests.SearchGo, t)
	assert.Equal(t, len(buildInfo.Modules[0].Artifacts), len(resultItems), "Incorrect number of artifacts were uploaded")
	propsMap := map[string]string{
		"build.name":   tests.GoBuildName,
		"build.number": buildNumber,
		"go.version":   "v1.0.0",
	}
	validateArtifactsProperties(resultItems, t, propsMap)

	assert.NoError(t, os.Chdir(wd))
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.GoBuildName, artHttpDetails)
	cleanGoTest()
}

func TestGoConfigWithModuleNameChange(t *testing.T) {
	initGoTest(t)
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(cliutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	prepareGoProject("", t, true)
	runGo(ModuleNameJFrogTest, tests.GoBuildName, buildNumber, t, "go", "build", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	assert.NoError(t, os.Chdir(wd))

	cleanGoTest()
}

func TestGoConfigWithoutModuleChange(t *testing.T) {
	initGoTest(t)
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(cliutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	prepareGoProject("", t, true)
	runGo("", tests.GoBuildName, buildNumber, t, "go", "build", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber)

	assert.NoError(t, os.Chdir(wd))

	cleanGoTest()
}

func TestGoWithGlobalConfig(t *testing.T) {
	initGoTest(t)
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)

	defer os.Setenv(cliutils.HomeDir, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	assert.NoError(t, err)

	prepareGoProject(newHomeDir, t, false)
	runGo(ModuleNameJFrogTest, tests.GoBuildName, buildNumber, t, "go", "build", "--build-name="+tests.GoBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	assert.NoError(t, os.Chdir(wd))

	cleanGoTest()
}

func runGo(module, buildName, buildNumber string, t *testing.T, args ...string) {
	artifactoryGoCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	assert.NoError(t, artifactoryGoCli.Exec(args...))
	cleanGoCache(t)
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	if module == "" {
		module = "github.com/jfrog/dependency"
	}
	artifactoryVersion, err := artAuth.GetVersion()
	assert.NoError(t, err)

	// Since Artifactory doesn't support info file before version 6.10.0, the artifacts count in the build info is different between versions
	version := version.NewVersion(artifactoryVersion)
	if version.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile) {
		validateBuildInfo(buildInfo, t, 4, 0, module)
	} else {
		validateBuildInfo(buildInfo, t, 8, 0, module)
	}

	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
}

func prepareGoProject(configDestDir string, t *testing.T, copyDirs bool) {
	project1Path := createGoProject(t, "project1", copyDirs)
	testsdataTarget := filepath.Join(tests.Out, "testsdata")
	testsdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testsdata")
	err := fileutils.CopyDir(testsdataSrc, testsdataTarget, copyDirs)
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
	artifactoryCli.Exec("go", "build", tests.GoRepo)
	cleanGoCache(t)

	// Publish dependency project to Artifactory
	artifactoryCli.Exec("gp", tests.GoRepo, "v1.0.0")
	cleanGoCache(t)

	assert.NoError(t, os.Chdir(project2Path))

	// Build the second project, download dependencies from Artifactory
	artifactoryCli.Exec("go", "build", tests.GoRepo)
	cleanGoCache(t)

	// Restore workspace
	assert.NoError(t, os.Chdir(wd))
	cleanGoTest()
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

	err = artifactoryCli.Exec("go", "build", tests.GoRepo)
	if err != nil {
		log.Warn(err)
		assert.Contains(t, err.Error(), executers.FailedToRetrieve)
		assert.Contains(t, err.Error(), executers.FromBothArtifactoryAndVcs)
	} else {
		assert.Fail(t, "Expected error but got success")
	}

	assert.NoError(t, os.Chdir(wd))
	cleanGoTest()
}

// Builds a project with a dependency of gofrog that is missing a mod file.
// Test the recursive overwrite capability.
// 1. Upload dependency.
// 2. Upload a project that is using that dependency
// 3. Build with recursive-tidy-overwrite set to true so the gofrog dependency will be downloaded from VCS
// 4. Check mod file (in Artifactory) of the dependency gofrog that populated.
// 5. Check mod file (in Artifactory) of the gofrog dependency (pkg/errors) that exists with the right content
func TestGoRecursivePublish(t *testing.T) {
	initGoTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)

	testsdataTarget := filepath.Join(tests.Out, "testsdata")
	testsdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testsdata")
	assert.NoError(t, fileutils.CopyDir(testsdataSrc, testsdataTarget, true))
	project1Path := createGoProject(t, "dependency", false)
	projectMissingDependency := createGoProject(t, "projectmissingdependency", false)
	projectBuild := createGoProject(t, "projectbuild", false)

	uploadGoProject(project1Path, t)
	uploadGoProject(projectMissingDependency, t)

	assert.NoError(t, os.Chdir(projectBuild))
	defer os.Chdir(wd)

	assert.NoError(t, artifactoryCli.Exec("grp", tests.GoRepo))
	sumFileExists, err := fileutils.IsFileExists("go.sum", false)
	assert.NoError(t, err)
	if sumFileExists {
		assert.NoError(t, os.Remove("go.sum"))
	}
	cleanGoCache(t)

	assert.NoError(t, os.Chdir(filepath.Join(wd, "testsdata", "go")))

	// Need to check the mod file within Artifactory of the gofrog dependency.
	content := downloadModFile(tests.DownloadModFileGo, wd, "gofrog", t)

	// Check that the file was signed:
	assert.NotContains(t, string(content), "// Generated by JFrog", "Expected file to be not signed")
	// Check that the mod file was populated with the dependency
	assert.NotContains(t, string(content), "require github.com/pkg/errors", "Expected to get empty mod file")

	assert.NoError(t, os.Chdir(filepath.Join(wd, "testsdata", "go")))

	// Need to check the mod file within Artifactory of the dependency of gofrog => pkg/errors.
	content = downloadModFile(tests.DownloadModOfDependencyGo, wd, "errors", t)

	// Check that the file was signed:
	assert.NotContains(t, string(content), "// Generated by JFrog", "Expected file to be not signed")
	// Check that the mod file contains dependency module.
	assert.Contains(t, string(content), "module github.com/pkg/errors", "Expected to get module github.com/pkg/errors")

	assert.NoError(t, os.Chdir(wd))

	cleanGoTest()
}

// Publish also the missing dependencies to Artifactory with the publishDeps flag.
// Checks that the dependency exists in Artifactory.
func TestGoWithPublishDeps(t *testing.T) {
	initGoTest(t)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	project1Path := createGoProject(t, "project1", false)
	testsdataTarget := filepath.Join(tests.Out, "testsdata")
	testsdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testsdata")
	assert.NoError(t, fileutils.CopyDir(testsdataSrc, testsdataTarget, true))
	assert.NoError(t, os.Chdir(project1Path))
	defer os.Chdir(wd)

	log.Info("Using Go project located at ", project1Path)
	artifactoryCli.Exec("go", "build", tests.GoRepo, "--publish-deps=true")
	cleanGoCache(t)

	assert.NoError(t, os.Chdir(filepath.Join(wd, "testsdata", "go")))

	content := downloadModFile(tests.DownloadModOfDependencyGo, wd, "errors", t)
	assert.NotContains(t, string(content), " module github.com/pkg/errors", "Wrong mod content was downloaded")

	assert.NoError(t, os.Chdir(wd))

	cleanGoTest()
}

func initGoTest(t *testing.T) {
	if !*tests.TestGo {
		t.Skip("Skipping go test. To run go test add the '-test.go=true' option.")
	}

	os.Setenv("GONOSUMDB", "github.com/jfrog")

	// Move when go will be supported and check Artifactory version.
	if !isRepoExist(tests.GoRepo) {
		repoConfig := filepath.FromSlash(tests.GetTestResourcesPath()) + tests.GoLocalRepositoryConfig
		repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
		require.NoError(t, err)
		execCreateRepoRest(repoConfig, tests.GoRepo)
	}
	authenticate()
}

func cleanGoTest() {
	os.Unsetenv("GONOSUMDB")
	if isRepoExist(tests.GoRepo) {
		execDeleteRepoRest(tests.GoRepo)
	}
	cleanBuildToolsTest()
}

func uploadGoProject(projectPath string, t *testing.T) {
	assert.NoError(t, os.Chdir(projectPath))
	// Publish project to Artifactory
	assert.NoError(t, artifactoryCli.Exec("gp", tests.GoRepo, "v1.0.0"))
	cleanGoCache(t)
}

func cleanGoCache(t *testing.T) {
	log.Info("Cleaning go cache by running: 'go clean -modcache'")

	cmd := exec.Command("go", "clean", "-modcache")
	assert.NoError(t, cmd.Run())
}

func createGoProject(t *testing.T, projectName string, includeDirs bool) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", projectName)
	projectTarget := filepath.Join(tests.Out, projectName)
	err := fileutils.CopyDir(projectSrc, projectTarget, includeDirs)
	assert.NoError(t, err)
	projectTarget, err = filepath.Abs(projectTarget)
	assert.NoError(t, err)
	return projectTarget
}

func downloadModFile(specName, wd, subDir string, t *testing.T) []byte {
	specFile, err := tests.CreateSpec(specName)
	assert.NoError(t, err)

	modDir := filepath.Join(wd, tests.Out, subDir)
	assert.NoError(t, os.MkdirAll(modDir, 0777))

	assert.NoError(t, os.Chdir(modDir))
	artifactoryCli.Exec("download", "--spec="+specFile, "--flat=true")
	files, err := fileutils.ListFiles(modDir, false)
	assert.NoError(t, err)
	assert.Len(t, files, 1, "Expected to get one mod file")

	content, err := ioutil.ReadFile(files[0])
	assert.NoError(t, err)
	return content
}

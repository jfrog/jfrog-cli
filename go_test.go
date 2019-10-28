package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/gocmd/executers"
	"github.com/jfrog/jfrog-cli-go/inttestutils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
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
	if err != nil {
		t.Error(err)
	}
	project1Path := createGoProject(t, "project1", false)
	testsdataTarget := filepath.Join(tests.Out, "testsdata")
	testsdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testsdata")
	err = fileutils.CopyDir(testsdataSrc, testsdataTarget, true)
	if err != nil {
		t.Error(err)
	}
	err = os.Chdir(project1Path)
	if err != nil {
		t.Error(err)
	}
	defer os.Chdir(wd)

	log.Info("Using Go project located at ", project1Path)

	buildName := "go-build"

	// 1. Download dependencies.
	// 2. Publish build-info.
	// 3. Validate the total count of dependencies added to the build-info.
	buildNumber := "1"

	artifactoryCli.Exec("go", "build", tests.GoLocalRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	cleanGoCache(t)

	artifactoryCli.Exec("bp", buildName, buildNumber)
	module := "github.com/jfrog/dependency"
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	artifactoryVersion, err := artAuth.GetVersion()
	if err != nil {
		t.Error(err)
	}

	// Since Artifactory doesn't support info file before version 6.10.0, the artifacts count in the build info is different between versions
	version := version.NewVersion(artifactoryVersion)
	expectedDependencies := 8
	expectedArtifacts := 2
	if version.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile) {
		expectedDependencies = 12
		expectedArtifacts = 3
	}
	validateBuildInfo(buildInfo, t, expectedDependencies, 0, module)

	// Now, using a new build number, do the following:
	// 1. Build the project again.
	// 2. Publish the go package.
	// 3. Validate the total count of dependencies and artifacts added to the build-info.
	// 4. Validate that the artifacts are tagged with the build.name and build.number properties.
	buildNumber = "2"

	artifactoryCli.Exec("go", "build", tests.GoLocalRepo, "--build-name="+buildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
	cleanGoCache(t)

	artifactoryCli.Exec("gp", tests.GoLocalRepo, "v1.0.0", "--build-name="+buildName, "--build-number="+buildNumber, "--deps=rsc.io/quote:v1.5.2", "--module="+ModuleNameJFrogTest)
	cleanGoCache(t)

	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo = inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, ModuleNameJFrogTest)

	err = os.Chdir(filepath.Join(wd, "testsdata", "go"))
	if err != nil {
		t.Error(err)
	}

	resultItems := getResultItemsFromArtifactory(tests.SearchGo, t)
	if len(buildInfo.Modules[0].Artifacts) != len(resultItems) {
		t.Error("Incorrect number of artifacts were uploaded, expected:", len(buildInfo.Modules[0].Artifacts), " Found:", len(resultItems))
	}
	propsMap := map[string]string{
		"build.name":   buildName,
		"build.number": buildNumber,
		"go.version":   "v1.0.0",
	}
	validateArtifactsProperties(resultItems, t, propsMap)

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanGoTest()
}

func TestGoConfigWithModuleNameChange(t *testing.T) {
	initGoTest(t)
	buildName := "go-build"
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(cliutils.JfrogHomeDirEnv, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	prepareGoProject("", t, true)
	runGo(ModuleNameJFrogTest, buildName, buildNumber, t, "go", "build", "--build-name="+buildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	cleanGoTest()
}

func TestGoConfigWithoutModuleChange(t *testing.T) {
	initGoTest(t)
	buildName := "go-build"
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)
	defer os.Setenv(cliutils.JfrogHomeDirEnv, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	prepareGoProject("", t, true)
	runGo("", buildName, buildNumber, t, "go", "build", "--build-name="+buildName, "--build-number="+buildNumber)

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	cleanGoTest()
}

func TestGoWithGlobalConfig(t *testing.T) {
	initGoTest(t)
	buildName := "go-build"
	buildNumber := "1"
	oldHomeDir, newHomeDir := prepareHomeDir(t)

	defer os.Setenv(cliutils.JfrogHomeDirEnv, oldHomeDir)
	defer os.RemoveAll(newHomeDir)

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	prepareGoProject(newHomeDir, t, false)
	runGo(ModuleNameJFrogTest, buildName, buildNumber, t, "go", "build", "--build-name="+buildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	cleanGoTest()
}

func runGo(module, buildName, buildNumber string, t *testing.T, args ...string) {
	artifactoryGoCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err := artifactoryGoCli.Exec(args...)
	if err != nil {
		t.Error(err)
	}
	cleanGoCache(t)
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	if module == "" {
		module = "github.com/jfrog/dependency"
	}
	artifactoryVersion, err := artAuth.GetVersion()
	if err != nil {
		t.Error(err)
	}

	// Since Artifactory doesn't support info file before version 6.10.0, the artifacts count in the build info is different between versions
	version := version.NewVersion(artifactoryVersion)
	if version.AtLeast(_go.ArtifactoryMinSupportedVersionForInfoFile) {
		validateBuildInfo(buildInfo, t, 12, 0, module)
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
	if err != nil {
		t.Error(err)
	}
	if configDestDir == "" {
		configDestDir = filepath.Join(project1Path, ".jfrog")
	}
	configFileDir := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "project1", ".jfrog", "projects")
	configFileDir, err = tests.ReplaceTemplateVariables(filepath.Join(configFileDir, "go.yaml"), filepath.Join(configDestDir, "projects"))
	if err != nil {
		t.Error(err)
	}
	err = os.Chdir(project1Path)
	if err != nil {
		t.Error(err)
	}
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
	if err != nil {
		t.Error(err)
	}
	project1Path := createGoProject(t, "project1", false)
	project2Path := createGoProject(t, "project2", false)
	err = os.Chdir(project1Path)
	if err != nil {
		t.Error(err)
	}

	// Download dependencies without Artifactory
	artifactoryCli.Exec("go", "build", tests.GoLocalRepo)
	cleanGoCache(t)

	// Publish dependency project to Artifactory
	artifactoryCli.Exec("gp", tests.GoLocalRepo, "v1.0.0")
	cleanGoCache(t)

	err = os.Chdir(project2Path)
	if err != nil {
		t.Error(err)
	}

	// Build the second project, download dependencies from Artifactory
	artifactoryCli.Exec("go", "build", tests.GoLocalRepo)
	cleanGoCache(t)

	// Restore workspace
	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}
	cleanGoTest()
}

// Testing the fallback mechanism
// 1. Building a project with a dependency that doesn't exists not in Artifactory and not in VCS.
// 2. The fallback mechanism will try to download from both VCS and Artifactory and then fail with an error
// 3. Testing that the error that is returned is the right error of the fallback.
func TestGoFallback(t *testing.T) {
	initGoTest(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	projectBuild := createGoProject(t, "projectbuild", false)

	err = os.Chdir(projectBuild)
	if err != nil {
		t.Error(err)
	}

	err = artifactoryCli.Exec("go", "build", tests.GoLocalRepo)
	if err != nil {
		log.Warn(err)
		if !strings.Contains(err.Error(), executers.FailedToRetrieve) || !strings.Contains(err.Error(), executers.FromBothArtifactoryAndVcs) {
			t.Error(err)
		}
	} else {
		t.Error("Expected error but got success")
	}

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}
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
	if err != nil {
		t.Error(err)
	}

	testsdataTarget := filepath.Join(tests.Out, "testsdata")
	testsdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testsdata")
	err = fileutils.CopyDir(testsdataSrc, testsdataTarget, true)
	if err != nil {
		t.Error(err)
	}
	project1Path := createGoProject(t, "dependency", false)
	projectMissingDependency := createGoProject(t, "projectmissingdependency", false)
	projectBuild := createGoProject(t, "projectbuild", false)

	uploadGoProject(project1Path, t)
	uploadGoProject(projectMissingDependency, t)

	err = os.Chdir(projectBuild)
	if err != nil {
		t.Error(err)
	}
	defer os.Chdir(wd)

	err = artifactoryCli.Exec("grp", tests.GoLocalRepo)
	if err != nil {
		t.Error(err)
	}
	sumFileExists, err := fileutils.IsFileExists("go.sum", false)
	if err != nil {
		t.Error(err)
	}
	if sumFileExists {
		err = os.Remove("go.sum")
		if err != nil {
			t.Error(err)
		}
	}
	cleanGoCache(t)

	err = os.Chdir(filepath.Join(wd, "testsdata", "go"))
	if err != nil {
		t.Error(err)
	}

	// Need to check the mod file within Artifactory of the gofrog dependency.
	content := downloadModFile(tests.DownloadModFileGo, wd, "gofrog", t)

	// Check that the file was signed:
	if strings.Contains(string(content), "// Generated by JFrog") {
		t.Error(fmt.Sprintf("Expected file to be not signed, however, the file is signed: %s", string(content)))
	}
	// Check that the mod file was populated with the dependency
	if strings.Contains(string(content), "require github.com/pkg/errors") {
		t.Error(fmt.Sprintf("Expected to get empty mod file, however, got: %s", string(content)))
	}

	err = os.Chdir(filepath.Join(wd, "testsdata", "go"))
	if err != nil {
		t.Error(err)
	}

	// Need to check the mod file within Artifactory of the dependency of gofrog => pkg/errors.
	content = downloadModFile(tests.DownloadModOfDependencyGo, wd, "errors", t)

	// Check that the file was signed:
	if strings.Contains(string(content), "// Generated by JFrog") {
		t.Error(fmt.Sprintf("Expected file to be not signed, however, the file is signed: %s", string(content)))
	}
	// Check that the mod file contains dependency module.
	if !strings.Contains(string(content), "module github.com/pkg/errors") {
		t.Error(fmt.Sprintf("Expected to get module github.com/pkg/errors, however, got: %s", string(content)))
	}

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	cleanGoTest()
}

// Publish also the missing dependencies to Artifactory with the publishDeps flag.
// Checks that the dependency exists in Artifactory.
func TestGoWithPublishDeps(t *testing.T) {
	initGoTest(t)
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	project1Path := createGoProject(t, "project1", false)
	testsdataTarget := filepath.Join(tests.Out, "testsdata")
	testsdataSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", "testsdata")
	err = fileutils.CopyDir(testsdataSrc, testsdataTarget, true)
	if err != nil {
		t.Error(err)
	}
	err = os.Chdir(project1Path)
	if err != nil {
		t.Error(err)
	}
	defer os.Chdir(wd)

	log.Info("Using Go project located at ", project1Path)
	artifactoryCli.Exec("go", "build", tests.GoLocalRepo, "--publish-deps=true")
	cleanGoCache(t)

	err = os.Chdir(filepath.Join(wd, "testsdata", "go"))
	if err != nil {
		t.Error(err)
	}

	content := downloadModFile(tests.DownloadModOfDependencyGo, wd, "errors", t)
	if strings.Contains(string(content), " module github.com/pkg/errors") {
		t.Error(fmt.Sprintf("Wrong mod content was downloaded: %s", string(content)))
	}

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	cleanGoTest()
}

func initGoTest(t *testing.T) {
	if !*tests.TestGo {
		t.Skip("Skipping go test. To run go test add the '-test.go=true' option.")
	}

	os.Setenv("GONOSUMDB", "github.com/jfrog")

	// Move when go will be supported and check Artifactory version.
	if !isRepoExist(tests.GoLocalRepo) {
		repoConfig := filepath.FromSlash(tests.GetTestResourcesPath()) + tests.GoLocalRepositoryConfig
		repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		execCreateRepoRest(repoConfig, tests.GoLocalRepo)
	}
	authenticate()
}

func cleanGoTest() {
	os.Unsetenv("GONOSUMDB")
	if isRepoExist(tests.GoLocalRepo) {
		execDeleteRepoRest(tests.GoLocalRepo)
	}
	cleanBuildToolsTest()
}

func uploadGoProject(projectPath string, t *testing.T) {
	err := os.Chdir(projectPath)
	if err != nil {
		t.Error(err)
	}
	// Publish project to Artifactory
	err = artifactoryCli.Exec("gp", tests.GoLocalRepo, "v1.0.0")
	if err != nil {
		t.Error(err)
	}
	cleanGoCache(t)
}

func cleanGoCache(t *testing.T) {
	log.Info("Cleaning go cache by running: 'go clean -modcache'")

	cmd := exec.Command("go", "clean", "-modcache")
	err := cmd.Run()
	if err != nil {
		t.Error(err)
	}
}

func createGoProject(t *testing.T, projectName string, includeDirs bool) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "go", projectName)
	projectTarget := filepath.Join(tests.Out, projectName)
	err := fileutils.CopyDir(projectSrc, projectTarget, includeDirs)
	if err != nil {
		t.Error(err)
	}
	projectTarget, err = filepath.Abs(projectTarget)
	if err != nil {
		t.Error(err)
	}
	return projectTarget
}
func downloadModFile(specName, wd, subDir string, t *testing.T) []byte {
	specFile, err := tests.CreateSpec(specName)
	if err != nil {
		t.Error(err)
	}

	modDir := filepath.Join(wd, tests.Out, subDir)
	err = os.MkdirAll(modDir, 0777)
	if err != nil {
		t.Error(err)
	}

	err = os.Chdir(modDir)
	if err != nil {
		t.Error(err)
	}
	artifactoryCli.Exec("download", "--spec="+specFile, "--flat=true")
	files, err := fileutils.ListFiles(modDir, false)
	if err != nil {
		t.Error(err)
	}

	if len(files) != 1 {
		t.Error(fmt.Sprintf("Expected to get one mod file but got %d", len(files)))
	}
	content, err := ioutil.ReadFile(files[0])
	if err != nil {
		t.Error(err)
	}
	return content
}

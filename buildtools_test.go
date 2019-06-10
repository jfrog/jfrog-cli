package main

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/gocmd/executers"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/inttestutils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/ioutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services/go"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const DockerTestImage string = "jfrog_cli_test_image"

func InitBuildToolsTests() {
	os.Setenv(cliutils.OfferConfig, "false")
	os.Setenv(cliutils.ReportUsage, "false")
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", cred)
	createReposIfNeeded()
	cleanBuildToolsTest()
}

func InitDockerTests() {
	if !*tests.TestDocker {
		return
	}
	os.Setenv(cliutils.ReportUsage, "false")
	os.Setenv(cliutils.OfferConfig, "false")
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", cred)
}

func CleanBuildToolsTests() {
	cleanBuildToolsTest()
	deleteRepos()
}

func createJfrogHomeConfig(t *testing.T) {
	templateConfigPath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "configtemplate", config.JfrogConfigFile)
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	err = os.Setenv(cliutils.JfrogHomeDirEnv, filepath.Join(wd, tests.Out, "jfroghome"))
	if err != nil {
		t.Error(err)
	}
	jfrogHomePath, err := config.GetJfrogHomeDir()
	if err != nil {
		t.Error(err)
	}
	_, err = tests.ReplaceTemplateVariables(templateConfigPath, jfrogHomePath)
	if err != nil {
		t.Error(err)
	}
}

func TestMavenBuildWithServerID(t *testing.T) {
	initBuildToolsTest(t)

	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	if err != nil {
		t.Error(err)
	}
	runAndValidateMaven(pomPath, configFilePath, t)
	cleanBuildToolsTest()
}

func TestMavenBuildWithCredentials(t *testing.T) {
	if *tests.RtUser == "" || *tests.RtPassword == "" {
		t.SkipNow()
	}

	initBuildToolsTest(t)

	pomPath := createMavenProject(t)
	srcConfigTemplate := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenUsernamePasswordTemplate)
	configFilePath, err := tests.ReplaceTemplateVariables(srcConfigTemplate, "")
	if err != nil {
		t.Error(err)
	}

	runAndValidateMaven(pomPath, configFilePath, t)
	cleanBuildToolsTest()
}

func runAndValidateMaven(pomPath, configFilePath string, t *testing.T) {
	buildConfig := &utils.BuildConfiguration{}
	mvnCmd := mvn.NewMvnCommand().SetConfiguration(buildConfig).SetGoals("clean install -f" + pomPath).SetConfigPath(configFilePath)
	err := mvnCmd.Run()
	if err != nil {
		t.Error(err)
	}
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	if err != nil {
		t.Error(err)
	}

	isExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
}

func TestGradleBuildWithServerID(t *testing.T) {
	initBuildToolsTest(t)

	buildGradlePath := createGradleProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	if err != nil {
		t.Error(err)
	}

	runAndValidateGradle(buildGradlePath, configFilePath, t)
	cleanBuildToolsTest()
}

func TestGradleBuildWithCredentials(t *testing.T) {
	if *tests.RtUser == "" || *tests.RtPassword == "" {
		t.SkipNow()
	}

	initBuildToolsTest(t)

	buildGradlePath := createGradleProject(t)
	srcConfigTemplate := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.GradleUseramePasswordTemplate)
	configFilePath, err := tests.ReplaceTemplateVariables(srcConfigTemplate, "")
	if err != nil {
		t.Error(err)
	}

	runAndValidateGradle(buildGradlePath, configFilePath, t)
	cleanBuildToolsTest()
}

// Image get parent image id command
type buildDockerImage struct {
	dockerFilePath string
	dockerTag      string
}

func (image *buildDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "build")
	cmd = append(cmd, image.dockerFilePath)
	cmd = append(cmd, "--tag", image.dockerTag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (image *buildDockerImage) GetEnv() map[string]string {
	return map[string]string{}
}

func (image *buildDockerImage) GetStdWriter() io.WriteCloser {
	return nil
}
func (image *buildDockerImage) GetErrWriter() io.WriteCloser {
	return nil
}

func initGoTest(t *testing.T) {
	if !*tests.TestGo {
		t.Skip("Skipping go test. To run go test add the '-test.go=true' option.")
	}

	err := os.Setenv("GO111MODULE", "on")
	if err != nil {
		t.Error(err)
	}
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

func initNugetTest(t *testing.T) {
	if !*tests.TestNuget {
		t.Skip("Skipping NuGet test. To run Nuget test add the '-test.nuget=true' option.")
	}

	if !cliutils.IsWindows() {
		t.Skip("Skipping nuget tests, since this is not a Windows machine.")
	}

	// This is due to Artifactory bug, we cant create remote repository with REST API.
	if !isRepoExist(tests.NugetRemoteRepo) {
		t.Error("Create nuget remote repository:", tests.NugetRemoteRepo, "in order to run nuget tests")
		t.FailNow()
	}
}

func cleanGoTest(gopath string) {
	if isRepoExist(tests.GoLocalRepo) {
		execDeleteRepoRest(tests.GoLocalRepo)
	}
	os.Setenv("GOPATH", gopath)
	cleanBuildToolsTest()
}

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
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", filepath.Join(wd, tests.Out))
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
	cleanGoTest(gopath)
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
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", filepath.Join(wd, tests.Out))

	prepareGoProject("", t, true)
	runGo(ModuleNameJFrogTest, buildName, buildNumber, t, "go", "build", "--build-name="+buildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	cleanGoTest(gopath)
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
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", filepath.Join(wd, tests.Out))

	prepareGoProject("", t, true)
	runGo("", buildName, buildNumber, t, "go", "build", "--build-name="+buildName, "--build-number="+buildNumber)

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	cleanGoTest(gopath)
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
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", filepath.Join(wd, tests.Out))

	prepareGoProject(newHomeDir, t, false)
	runGo(ModuleNameJFrogTest, buildName, buildNumber, t, "go", "build", "--build-name="+buildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	cleanGoTest(gopath)
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

func prepareHomeDir(t *testing.T) (string, string) {
	oldHomeDir := os.Getenv(cliutils.JfrogHomeDirEnv)
	// Populate cli config with 'default' server
	createJfrogHomeConfig(t)
	newHomeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		t.Error(err)
	}
	return oldHomeDir, newHomeDir
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
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", filepath.Join(wd, tests.Out))
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
	cleanGoTest(gopath)
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
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", filepath.Join(wd, tests.Out))
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
	cleanGoTest(gopath)
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

	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", filepath.Join(wd, tests.Out))

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

	cleanGoTest(gopath)
}

// Publish also the missing dependencies to Artifactory with the publishDeps flag.
// Checks that the dependency exists in Artifactory.
func TestGoWithPublishDeps(t *testing.T) {
	initGoTest(t)
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", filepath.Join(wd, tests.Out))
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

	cleanGoTest(gopath)
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

func TestDockerPush(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}
	runDockerPushTest(DockerTestImage, DockerTestImage+":1", false, t)
}

func TestDockerPushWithModuleName(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}
	runDockerPushTest(DockerTestImage, ModuleNameJFrogTest, true, t)
}

func TestDockerPushWithMultipleSlash(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}
	runDockerPushTest(DockerTestImage+"/multiple", "multiple:1", false, t)
}

// Run docker push to Artifactory
func runDockerPushTest(imageName, module string, withModule bool, t *testing.T) {
	imageTag := buildTestDockerImage(imageName)
	buildName := "docker-build"
	buildNumber := "1"

	// Push docker image using docker client
	if withModule {
		artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber, "--module="+module)
	} else {
		artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	}
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	imagePath := path.Join(*tests.DockerTargetRepo, imageName, "1") + "/"
	validateDockerBuild(buildName, buildNumber, imagePath, module, 7, 5, 7, t)

	dockerTestCleanup(imageName, buildName)
}

func TestDockerPull(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}

	imageName := DockerTestImage
	imageTag := buildTestDockerImage(imageName)

	// Push docker image using docker client
	artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo)

	buildName := "docker-pull"
	buildNumber := "1"

	// Pull docker image using docker client
	artifactoryCli.Exec("docker-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	imagePath := path.Join(*tests.DockerTargetRepo, imageName, "1") + "/"
	validateDockerBuild(buildName, buildNumber, imagePath, imageName+":1", 0, 7, 7, t)

	buildNumber = "2"
	artifactoryCli.Exec("docker-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
	artifactoryCli.Exec("build-publish", buildName, buildNumber)
	validateDockerBuild(buildName, buildNumber, imagePath, ModuleNameJFrogTest, 0, 7, 7, t)

	dockerTestCleanup(imageName, buildName)
}

func buildTestDockerImage(imageName string) string {
	imageTag := path.Join(*tests.DockerRepoDomain, imageName+":1")
	dockerFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "docker")
	imageBuilder := &buildDockerImage{dockerTag: imageTag, dockerFilePath: dockerFilePath}
	gofrogcmd.RunCmd(imageBuilder)
	return imageTag
}

func validateDockerBuild(buildName, buildNumber, imagePath, module string, expectedArtifacts, expectedDependencies, expectedItemsInArtifactory int, t *testing.T) {
	specFile := spec.NewBuilder().Pattern(imagePath + "*").BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(specFile)
	err := searchCmd.Search()
	if err != nil {
		log.Error(err)
		t.Error(err)
	}
	if expectedItemsInArtifactory != len(searchCmd.SearchResult()) {
		t.Error("Docker build info was not pushed correctly, expected:", expectedArtifacts, " Found:", len(searchCmd.SearchResult()))
	}

	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, module)
}

func dockerTestCleanup(imageName, buildName string) {
	// Remove build from Artifactory
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)

	// Remove image from Artifactory
	deleteSpec := spec.NewBuilder().Pattern(path.Join(*tests.DockerTargetRepo, imageName)).BuildSpec()
	tests.DeleteFiles(deleteSpec, artifactoryDetails)
}

func validateBuildInfo(buildInfo buildinfo.BuildInfo, t *testing.T, expectedDependencies int, expectedArtifacts int, moduleName string) {
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		t.Error("build info was not generated correctly, no modules were created.")
	}
	if buildInfo.Modules[0].Id != moduleName {
		t.Error(fmt.Errorf("Expected module name %s, got %s", moduleName, buildInfo.Modules[0].Id))
	}
	if expectedDependencies != len(buildInfo.Modules[0].Dependencies) {
		t.Error("Incorrect number of dependencies found in the build-info, expected:", expectedDependencies, " Found:", len(buildInfo.Modules[0].Dependencies))
	}
	if expectedArtifacts != len(buildInfo.Modules[0].Artifacts) {
		t.Error("Incorrect number of artifacts found in the build-info, expected:", expectedArtifacts, " Found:", len(buildInfo.Modules[0].Artifacts))
	}
}

func TestNugetResolve(t *testing.T) {
	initNugetTest(t)
	projects := []struct {
		name                 string
		project              string
		moduleId             string
		args                 []string
		expectedDependencies int
	}{
		{"packagesconfigwithoutmodulechnage", "packagesconfig", "packagesconfig", []string{"nuget", "restore", tests.NugetRemoteRepo, "--build-name=" + tests.NugetBuildName}, 6},
		{"packagesconfigwithmodulechnage", "packagesconfig", ModuleNameJFrogTest, []string{"nuget", "restore", tests.NugetRemoteRepo, "--build-name=" + tests.NugetBuildName, "--module=" + ModuleNameJFrogTest}, 6},
		{"referencewithoutmodulechnage", "reference", "reference", []string{"nuget", "restore", tests.NugetRemoteRepo, "--build-name=" + tests.NugetBuildName}, 6},
		{"referencewithmodulechnage", "reference", ModuleNameJFrogTest, []string{"nuget", "restore", tests.NugetRemoteRepo, "--build-name=" + tests.NugetBuildName, "--module=" + ModuleNameJFrogTest}, 6},
	}
	for buildNumber, test := range projects {
		t.Run(test.project, func(t *testing.T) {
			testNugetCmd(t, createNugetProject(t, test.project), strconv.Itoa(buildNumber), test.moduleId, test.expectedDependencies, test.args)
		})
	}
	cleanBuildToolsTest()
}

func createNugetProject(t *testing.T, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "nuget", projectName)
	projectTarget := filepath.Join(tests.Out, projectName)
	err := fileutils.CreateDirIfNotExist(projectTarget)
	if err != nil {
		t.Error(err)
	}

	files, err := fileutils.ListFiles(projectSrc, false)
	if err != nil {
		t.Error(err)
	}

	for _, v := range files {
		err = fileutils.CopyFile(projectTarget, v)
		if err != nil {
			t.Error(err)
		}
	}
	return projectTarget
}

func testNugetCmd(t *testing.T, projectPath, buildNumber, module string, expectedDependencies int, args []string) {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	err = os.Chdir(projectPath)
	if err != nil {
		t.Error(err)
	}
	args = append(args, "--build-number="+buildNumber)
	artifactoryCli.Exec(args...)
	artifactoryCli.Exec("bp", tests.NugetBuildName, buildNumber)

	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.NugetBuildName, buildNumber, t, artHttpDetails)
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		t.Error("Nuget build info was not generated correctly, no modules were created.")
	}

	if expectedDependencies != len(buildInfo.Modules[0].Dependencies) {
		t.Error("Incorrect number of artifacts found in the build-info, expected:", expectedDependencies, " Found:", len(buildInfo.Modules[0].Dependencies))
	}

	if module != buildInfo.Modules[0].Id {
		t.Error(fmt.Errorf("Expected module name %s, got %s", module, buildInfo.Modules[0].Id))
	}

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}

	// cleanup
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.NugetBuildName, artHttpDetails)
	cleanBuildToolsTest()
}

func TestNpm(t *testing.T) {
	initBuildToolsTest(t)
	npmi := "npm-install"
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath, npmProjectCi := initNpmTest(t)
	var npmTests = []npmTestParams{
		{command: "npmci", repo: tests.NpmRemoteRepo, wd: npmProjectCi, validationFunc: validateNpmInstall},
		{command: "npmci", repo: tests.NpmRemoteRepo, wd: npmProjectCi, moduleName: ModuleNameJFrogTest, validationFunc: validateNpmInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, moduleName: ModuleNameJFrogTest, validationFunc: validateNpmInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmScopedProjectPath, validationFunc: validateNpmInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmNpmrcProjectPath, validationFunc: validateNpmInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateNpmInstall, npmArgs: "--production"},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateNpmInstall, npmArgs: "-only=dev"},
		{command: "npmi", repo: tests.NpmRemoteRepo, wd: npmNpmrcProjectPath, validationFunc: validateNpmPackInstall, npmArgs: "yaml"},
		{command: "npmp", repo: tests.NpmLocalRepo, wd: npmScopedProjectPath, moduleName: ModuleNameJFrogTest, validationFunc: validateNpmScopedPublish},
		{command: "npm-publish", repo: tests.NpmLocalRepo, wd: npmProjectPath, validationFunc: validateNpmPublish},
	}

	for i, npmTest := range npmTests {
		err = os.Chdir(filepath.Dir(npmTest.wd))
		if err != nil {
			t.Error(err)
		}
		npmrcFileInfo, err := os.Stat(".npmrc")
		if err != nil && !os.IsNotExist(err) {
			t.Error(err)
		}
		if npmTest.moduleName != "" {
			artifactoryCli.Exec(npmTest.command, npmTest.repo, "--npm-args="+npmTest.npmArgs, "--build-name="+tests.NpmBuildName, "--build-number="+strconv.Itoa(i+1), "--module="+npmTest.moduleName)
		} else {
			npmTest.moduleName = "jfrog-cli-tests:1.0.0"
			artifactoryCli.Exec(npmTest.command, npmTest.repo, "--npm-args="+npmTest.npmArgs, "--build-name="+tests.NpmBuildName, "--build-number="+strconv.Itoa(i+1))
		}
		artifactoryCli.Exec("bp", tests.NpmBuildName, strconv.Itoa(i+1))
		npmTest.buildNumber = strconv.Itoa(i + 1)
		npmTest.validationFunc(t, npmTest)

		// make sure npmrc file was not changed (if existed)
		postTestFileInfo, postTestFileInfoErr := os.Stat(".npmrc")
		validateNpmrcFileInfo(t, npmTest, npmrcFileInfo, postTestFileInfo, err, postTestFileInfoErr)
	}

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}
	cleanBuildToolsTest()
	inttestutils.DeleteBuild(artifactoryDetails.Url, tests.NpmBuildName, artHttpDetails)
}

func TestGetJcenterRemoteDetails(t *testing.T) {
	initBuildToolsTest(t)
	createServerConfigAndReturnPassphrase()

	unsetEnvVars := func() {
		err := os.Unsetenv(utils.JCenterRemoteServerEnv)
		if err != nil {
			t.Error(err)
		}
		err = os.Unsetenv(utils.JCenterRemoteRepoEnv)
		if err != nil {
			t.Error(err)
		}
	}
	unsetEnvVars()
	defer unsetEnvVars()

	// The utils.JCenterRemoteServerEnv env var is not set, so extractor1.jar should be downloaded from jcenter.
	downloadPath := "org/jfrog/buildinfo/build-info-extractor/extractor1.jar"
	expectedRemotePath := path.Join("bintray/jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Still, the utils.JCenterRemoteServerEnv env var is not set, so the download should be from jcenter.
	// Expecting a different download path this time.
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor2.jar"
	expectedRemotePath = path.Join("bintray/jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Setting the utils.JCenterRemoteServerEnv env var now,
	// Expecting therefore the download to be from the the server ID configured by this env var.
	err := os.Setenv(utils.JCenterRemoteServerEnv, tests.RtServerId)
	if err != nil {
		t.Error(err)
	}
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor3.jar"
	expectedRemotePath = path.Join("jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Still expecting the download to be from the same server ID, but this time, not through a remote repo named
	// jcenter, but through test-remote-repo.
	testRemoteRepo := "test-remote-repo"
	err = os.Setenv(utils.JCenterRemoteRepoEnv, testRemoteRepo)
	if err != nil {
		t.Error(err)
	}
	downloadPath = "1org/jfrog/buildinfo/build-info-extractor/extractor4.jar"
	expectedRemotePath = path.Join(testRemoteRepo, downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	cleanBuildToolsTest()
}

func validateJcenterRemoteDetails(t *testing.T, downloadPath, expectedRemotePath string) {
	artDetails, remotePath, err := utils.GetJcenterRemoteDetails(downloadPath)
	if err != nil {
		t.Error(err)
	}
	if remotePath != expectedRemotePath {
		t.Error("Expected remote path to be", expectedRemotePath, "but got", remotePath)
	}
	if os.Getenv(utils.JCenterRemoteServerEnv) != "" && artDetails == nil {
		t.Error("Expected a server to be returned")
	}
}

func validateNpmrcFileInfo(t *testing.T, npmTest npmTestParams, npmrcFileInfo, postTestNpmrcFileInfo os.FileInfo, err, postTestFileInfoErr error) {
	if postTestFileInfoErr != nil && !os.IsNotExist(postTestFileInfoErr) {
		t.Error(postTestFileInfoErr)
	}
	if err == nil && postTestFileInfoErr != nil {
		t.Error(".npmrc file existed and was not resored at the end of the install command.")
	}
	if err != nil && postTestFileInfoErr == nil {
		t.Error(".npmrc file was not deleted at the end of the install command.")
	}
	if err == nil && postTestFileInfoErr == nil && (npmrcFileInfo.Mode() != postTestNpmrcFileInfo.Mode() || npmrcFileInfo.Size() != postTestNpmrcFileInfo.Size()) {
		t.Errorf(".npmrc file was changed after running npm command! it was:\n%v\nnow it is:\n%v\nTest arguments are:\n%v", npmrcFileInfo, postTestNpmrcFileInfo, npmTest)
	}
	// make sue the temp .npmrc was deleted.
	bcpNpmrc, err := os.Stat("jfrog.npmrc.backup")
	if err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
	if bcpNpmrc != nil {
		t.Errorf("The file 'jfrog.npmrc.backup' was supposed to be deleted but it was not when running the configuration:\n%v", npmTest)
	}
}

func initNpmTest(t *testing.T) (npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath, npmProjectCi string) {
	npmProjectPath, err := filepath.Abs(createNpmProject(t, "npmproject"))
	if err != nil {
		t.Error(err)
	}
	npmScopedProjectPath, err = filepath.Abs(createNpmProject(t, "npmscopedproject"))
	if err != nil {
		t.Error(err)
	}
	npmNpmrcProjectPath, err = filepath.Abs(createNpmProject(t, "npmnpmrcproject"))
	if err != nil {
		t.Error(err)
	}
	npmProjectCi, err = filepath.Abs(createNpmProject(t, "npmprojectci"))
	if err != nil {
		t.Error(err)
	}
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectPath))
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectCi))
	return
}

func runAndValidateGradle(buildGradlePath, configFilePath string, t *testing.T) {
	buildConfig := &utils.BuildConfiguration{}
	gradleCmd := gradle.NewGradleCommand().SetTasks("clean artifactoryPublish -b " + buildGradlePath).SetConfigPath(configFilePath).SetConfiguration(buildConfig)
	err := gradleCmd.Run()
	if err != nil {
		t.Error(err)
	}
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	if err != nil {
		t.Error(err)
	}

	isExistInArtifactory(tests.GetGradleDeployedArtifacts(), searchSpec, t)
}

func createGradleProject(t *testing.T) string {
	srcBuildFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradleproject", "build.gradle")
	buildGradlePath, err := tests.ReplaceTemplateVariables(srcBuildFile, "")
	if err != nil {
		t.Error(err)
	}

	srcSettingsFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "gradleproject", "settings.gradle")
	_, err = tests.ReplaceTemplateVariables(srcSettingsFile, "")
	if err != nil {
		t.Error(err)
	}

	return buildGradlePath
}

func createMavenProject(t *testing.T) string {
	srcPomFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "mavenproject", "pom.xml")
	pomPath, err := tests.ReplaceTemplateVariables(srcPomFile, "")
	if err != nil {
		t.Error(err)
	}
	return pomPath
}

func createNpmProject(t *testing.T, dir string) string {
	srcPackageJson := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "npm", dir, "package.json")
	targetPackageJson := filepath.Join(tests.Out, dir)
	packageJson, err := tests.ReplaceTemplateVariables(srcPackageJson, targetPackageJson)
	if err != nil {
		t.Error(err)
	}

	// failure can be ignored
	npmrcExists, err := fileutils.IsFileExists(filepath.Join(filepath.Dir(srcPackageJson), ".npmrc"), false)
	if err != nil {
		t.Error(err)
	}

	if npmrcExists {
		if _, err = tests.ReplaceTemplateVariables(filepath.Join(filepath.Dir(srcPackageJson), ".npmrc"), targetPackageJson); err != nil {
			t.Error(err)
		}
	}
	return packageJson
}

func initBuildToolsTest(t *testing.T) {
	if !*tests.TestBuildTools {
		t.Skip("Skipping build tools test. To run build tools tests add the '-test.buildTools=true' option.")
	}
	createJfrogHomeConfig(t)
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", cred)
	configArtifactoryCli = createConfigJfrogCLI(cred)
}

func prepareArtifactoryForNpmBuild(t *testing.T, workingDirectory string) {
	if err := os.Chdir(workingDirectory); err != nil {
		t.Error(err)
	}

	caches := ioutils.DoubleWinPathSeparator(filepath.Join(workingDirectory, "caches"))
	// Run install with -cache argument to download the artifacts from Artifactory
	// This done to be sure the artifacts exists in Artifactory
	artifactoryCli.Exec("npm-install", tests.NpmRemoteRepo, "--npm-args=-cache="+caches)

	if err := os.RemoveAll(filepath.Join(workingDirectory, "node_modules")); err != nil {
		t.Error(err)
	}

	if err := os.RemoveAll(caches); err != nil {
		t.Error(err)
	}
}

func cleanBuildToolsTest() {
	if *tests.TestBuildTools || *tests.TestGo || *tests.TestNuget {
		os.Unsetenv(cliutils.JfrogHomeDirEnv)
		cleanArtifactory()
		tests.CleanFileSystem()
	}
}

func validateNpmInstall(t *testing.T, npmTestParams npmTestParams) {
	type expectedDependency struct {
		id     string
		scopes []string
	}
	var expectedDependencies []expectedDependency
	if !strings.Contains(npmTestParams.npmArgs, "-only=dev") {
		expectedDependencies = append(expectedDependencies, expectedDependency{id: "xml-1.0.1.tgz", scopes: []string{"production"}})
	}
	if !strings.Contains(npmTestParams.npmArgs, "-only=prod") && !strings.Contains(npmTestParams.npmArgs, "-production") {
		expectedDependencies = append(expectedDependencies, expectedDependency{id: "json-9.0.6.tgz", scopes: []string{"development"}})
	}
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.NpmBuildName, npmTestParams.buildNumber, t, artHttpDetails)
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		// Case no module was created
		t.Errorf("npm install test with the arguments: \n%v \nexpected to have module with the following dependencies: \n%v \nbut has no modules: \n%v",
			npmTestParams, expectedDependencies, buildInfo)
	}
	if len(expectedDependencies) != len(buildInfo.Modules[0].Dependencies) {
		// The checksums are ignored when comparing the actual and the expected
		t.Errorf("npm install test with the arguments: \n%v \nexpected to have the following dependencies: \n%v \nbut has: \n%v",
			npmTestParams, expectedDependencies, dependenciesToPrintableArray(buildInfo.Modules[0].Dependencies))
	}
	for _, expectedDependency := range expectedDependencies {
		found := false
		for _, actualDependency := range buildInfo.Modules[0].Dependencies {
			if actualDependency.Id == expectedDependency.id &&
				len(actualDependency.Scopes) == len(expectedDependency.scopes) &&
				actualDependency.Scopes[0] == expectedDependency.scopes[0] {
				found = true
				break
			}
		}
		if !found {
			// The checksums are ignored when comparing the actual and the expected
			t.Errorf("npm install test with the arguments: \n%v \nexpected to have the following dependencies: \n%v \nbut has: \n%v",
				npmTestParams, expectedDependencies, dependenciesToPrintableArray(buildInfo.Modules[0].Dependencies))
		}
	}
}

func validateNpmPackInstall(t *testing.T, npmTestParams npmTestParams) {
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.NpmBuildName, npmTestParams.buildNumber, t, artHttpDetails)
	if len(buildInfo.Modules) > 0 {
		t.Errorf("npm install test with the arguments: \n%v \nexpected to have no modules but has: \n%v",
			npmTestParams, buildInfo.Modules[0])
	}

	packageJsonFile, err := ioutil.ReadFile(npmTestParams.wd)
	if err != nil {
		t.Error(err)
	}

	var packageJson struct {
		Dependencies map[string]string `json:"dependencies,omitempty"`
	}
	if err := json.Unmarshal(packageJsonFile, &packageJson); err != nil {
		t.Error(err)
	}

	if len(packageJson.Dependencies) != 2 || packageJson.Dependencies[npmTestParams.npmArgs] == "" {
		t.Errorf("npm install test with the arguments: \n%v \nexpected have the dependency %v in the following package.json file: \n%v",
			npmTestParams, npmTestParams.npmArgs, packageJsonFile)
	}
}

func validateNpmPublish(t *testing.T, npmTestParams npmTestParams) {
	isExistInArtifactoryByProps(tests.GetNpmDeployedArtifacts(),
		tests.NpmLocalRepo+"/*",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", tests.NpmBuildName, npmTestParams.buildNumber), t)
	validateNpmCommonPublish(t, npmTestParams)
}

func validateNpmScopedPublish(t *testing.T, npmTestParams npmTestParams) {
	isExistInArtifactoryByProps(tests.GetNpmDeployedScopedArtifacts(),
		tests.NpmLocalRepo+"/*",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", tests.NpmBuildName, npmTestParams.buildNumber), t)
	validateNpmCommonPublish(t, npmTestParams)
}

func validateNpmCommonPublish(t *testing.T, npmTestParams npmTestParams) {
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.NpmBuildName, npmTestParams.buildNumber, t, artHttpDetails)
	expectedArtifactName := "jfrog-cli-tests-1.0.0.tgz"
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		// Case no module was created
		t.Errorf("npm publish test with the arguments: \n%v \nexpected to have module with the following artifact: \n%v \nbut has no modules: \n%v",
			npmTestParams, expectedArtifactName, buildInfo)
	}
	if len(buildInfo.Modules[0].Artifacts) != 1 {
		// The checksums are ignored when comparing the actual and the expected
		t.Errorf("npm publish test with the arguments: \n%v \nexpected to have the following artifact: \n%v \nbut has: \n%v",
			npmTestParams, expectedArtifactName, buildInfo.Modules[0].Artifacts)
	}
	if buildInfo.Modules[0].Id != npmTestParams.moduleName {
		t.Errorf("npm publish test with the arguments: \n%v \nexpected to have the following module name: \n%v \nbut has: \n%v",
			npmTestParams, npmTestParams.moduleName, buildInfo.Modules[0].Id)
	}
	if buildInfo.Modules[0].Artifacts[0].Name != expectedArtifactName {
		t.Errorf("npm publish test with the arguments: \n%v \nexpected to have the following artifact: \n%v \nbut has: \n%v",
			npmTestParams, expectedArtifactName, buildInfo.Modules[0].Artifacts[0].Name)
	}
}

type npmTestParams struct {
	command        string
	repo           string
	npmArgs        string
	wd             string
	buildNumber    string
	moduleName     string
	validationFunc func(*testing.T, npmTestParams)
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/jfrog/inttestutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	rtutils "github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func InitBuildToolsTests() {
	os.Setenv("JFROG_CLI_OFFER_CONFIG", "false")
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(main, "jfrog rt", cred)
	createReposIfNeeded()
	cleanBuildToolsTest()
}

func InitDockerTests() {
	if !*tests.TestDocker {
		return
	}
	os.Setenv("JFROG_CLI_OFFER_CONFIG", "false")
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(main, "jfrog rt", cred)
}

func CleanBuildToolsTests() {
	cleanBuildToolsTest()
	deleteRepos()
}

func createJfrogHomeConfig(t *testing.T) {
	templateConfigPath := filepath.Join(tests.GetTestResourcesPath(), "configtemplate", config.JfrogConfigFile)

	err := os.Setenv(config.JfrogHomeEnv, filepath.Join(tests.Out, "jfroghome"))
	jfrogHomePath, err := config.GetJfrogHomeDir()
	if err != nil {
		t.Error(err)
	}
	_, err = copyTemplateFile(templateConfigPath, jfrogHomePath, config.JfrogConfigFile, true)
	if err != nil {
		t.Error(err)
	}
}

func TestMavenBuildWithServerID(t *testing.T) {
	initBuildToolsTest(t)

	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(tests.GetTestResourcesPath(), "buildspecs", tests.MavenServerIDConfig)

	runAndValidateMaven(pomPath, configFilePath, t)
	cleanBuildToolsTest()
}

func TestMavenBuildWithCredentials(t *testing.T) {
	initBuildToolsTest(t)

	pomPath := createMavenProject(t)
	srcConfigTemplate := filepath.Join(tests.GetTestResourcesPath(), "buildspecs", tests.MavenUsernamePasswordTemplate)
	targetBuildSpecPath := filepath.Join(tests.Out, "buildspecs")
	configFilePath, err := copyTemplateFile(srcConfigTemplate, targetBuildSpecPath, tests.MavenUsernamePasswordTemplate, true)
	if err != nil {
		t.Error(err)
	}

	runAndValidateMaven(pomPath, configFilePath, t)
	cleanBuildToolsTest()
}

func runAndValidateMaven(pomPath, configFilePath string, t *testing.T) {
	buildConfig := &utils.BuildConfiguration{}
	err := mvn.Mvn("clean install -f"+pomPath, configFilePath, buildConfig)
	if err != nil {
		t.Error(err)
	}
	isExistInArtifactory(tests.MavenDeployedArtifacts, tests.GetFilePath(tests.SearchAllRepo1), t)
}

func TestGradleBuildWithServerID(t *testing.T) {
	initBuildToolsTest(t)

	buildGradlePath := createGradleProject(t)
	configFilePath := filepath.Join(tests.GetTestResourcesPath(), "buildspecs", tests.GradleServerIDConfig)

	runAndValidateGradle(buildGradlePath, configFilePath, t)
	cleanBuildToolsTest()
}

func TestGradleBuildWithCredentials(t *testing.T) {
	initBuildToolsTest(t)

	buildGradlePath := createGradleProject(t)
	srcConfigTemplate := filepath.Join(tests.GetTestResourcesPath(), "buildspecs", tests.GradleUseramePasswordTemplate)
	targetBuildSpecPath := filepath.Join(tests.Out, "buildspecs")
	configFilePath, err := copyTemplateFile(srcConfigTemplate, targetBuildSpecPath, tests.GradleUseramePasswordTemplate, true)
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
		repoConfig := tests.GetTestResourcesPath() + tests.GoLocalRepositoryConfig
		execCreateRepoRest(repoConfig, tests.GoLocalRepo)
	}
}

func initNugetTest(t *testing.T) {
	if !*tests.TestNuget {
		t.Skip("Skipping NuGet test. To run Nuget test add the '-test.nuget=true' option.")
	}

	if runtime.GOOS != "windows" {
		t.Skip("Skipping nuget tests, since this is not a Windows machine.")
	}

	// This is due to Artifactory bug, we cant create remote repository with REST API.
	if !isRepoExist(tests.NugetRemoteRepo) {
		t.Error("Create nuget remote repository:", tests.NugetRemoteRepo, "in order to run nuget tests")
		t.FailNow()
	}
}

func cleanGoTest() {
	if isRepoExist(tests.GoLocalRepo) {
		execDeleteRepoRest(tests.GoLocalRepo)
	}
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
	project1Path := createGoProject(t, "project1")
	testsdataTarget := filepath.Join(tests.Out, "testsdata")
	testsdataSrc := filepath.Join(tests.GetTestResourcesPath(), "go", "testsdata")
	err = fileutils.CopyDir(testsdataSrc, testsdataTarget, true)
	if err != nil {
		t.Error(err)
	}
	err = os.Chdir(project1Path)
	if err != nil {
		t.Error(err)
	}

	log.Info("Using Go project located at ", project1Path)

	buildName := "go-build"

	// 1. Download dependencies without Artifactory.
	// 2. Publish build-info.
	// 3. Validate the total count of dependencies added to the build-info.
	buildNumber := "1"
	artifactoryCli.Exec("go", "build", tests.GoLocalRepo, "--no-registry=true", "--build-name="+buildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 6, 0)

	// Do the same, with one exception - download the dependencies from Artifactory.
	// Use a new build number. Expect the same results as the previous build.
	buildNumber = "2"
	artifactoryCli.Exec("go", "build", tests.GoLocalRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo = inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 6, 0)

	// Now, using a new build number, do the following:
	// 1. Build the project again.
	// 2. Publish the go package.
	// 3. Validate the total count of dependencies and artifacts added to the build-info.
	// 4. Validate that the artifacts are tagged with the build.name and build.number properties.
	buildNumber = "3"
	artifactoryCli.Exec("go", "build", tests.GoLocalRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("gp", tests.GoLocalRepo, "v1.0.0", "--build-name="+buildName, "--build-number="+buildNumber, "--deps=rsc.io/quote:v1.5.2")
	artifactoryCli.Exec("bp", buildName, buildNumber)
	buildInfo = inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 6, 2)
	validateBuildInfoProperties(buildInfo, t)

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	cleanGoTest()
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
	project1Path := createGoProject(t, "project1")
	project2Path := createGoProject(t, "project2")
	err = os.Chdir(project1Path)
	if err != nil {
		t.Error(err)
	}

	// Download dependencies without Artifactory
	artifactoryCli.Exec("go", "build", tests.GoLocalRepo, "--no-registry=true")
	// Publish dependency project to Artifactory
	artifactoryCli.Exec("gp", tests.GoLocalRepo, "v1.0.0", "--deps=ALL")

	err = os.Chdir(project2Path)
	if err != nil {
		t.Error(err)
	}

	// Build the second project, download dependencies from Artifactory
	artifactoryCli.Exec("go", "build", tests.GoLocalRepo)

	// Restore workspace
	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}
	cleanGoTest()
}

func createGoProject(t *testing.T, projectName string) string {
	projectSrc := filepath.Join(tests.GetTestResourcesPath(), "go", projectName)
	projectTarget := filepath.Join(tests.Out, projectName)
	err := fileutils.CopyDir(projectSrc, projectTarget, false)
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
	runDockerTest("jfrog_cli_test_image", t)
}

func TestDockerPushWithMultipleSlash(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}
	runDockerTest("jfrog_cli_test_image/multiple", t)
}

// Run docker push to Artifactory
func runDockerTest(imageName string, t *testing.T) {
	imageTag := path.Join(*tests.DockerRepoDomain, imageName+":1")
	dockerFilePath := filepath.Join(tests.GetTestResourcesPath(), "docker")
	imageBuilder := &buildDockerImage{dockerTag: imageTag, dockerFilePath: dockerFilePath}
	utils.RunCmd(imageBuilder)

	buildName := "docker-build"
	buildNumber := "1"

	// Push docker image using docker client
	artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	imagePath := path.Join(*tests.DockerTargetRepo, imageName, "1") + "/"
	validateDockerBuild(buildName, buildNumber, imagePath, 7, 5, t)
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)

	deleteFlags := new(generic.DeleteConfiguration)
	deleteSpec := spec.NewBuilder().Pattern(path.Join(*tests.DockerTargetRepo, imageName)).BuildSpec()
	deleteFlags.ArtDetails = artifactoryDetails
	generic.Delete(deleteSpec, deleteFlags)
}

func validateDockerBuild(buildName, buildNumber, imagePath string, expectedArtifacts, expectedDependencies int, t *testing.T) {
	specFile := spec.NewBuilder().Pattern(imagePath + "*").BuildSpec()
	result, err := generic.Search(specFile, artifactoryDetails)
	if err != nil {
		log.Error(err)
		t.Error(err)
	}
	if expectedArtifacts != len(result) {
		t.Error("Docker build info was not pushed correctly, expected:", expectedArtifacts, " Found:", len(result))
	}

	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts)
}

func validateBuildInfo(buildInfo buildinfo.BuildInfo, t *testing.T, expectedDependencies int, expectedArtifacts int) {
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		t.Error("build info was not generated correctly, no modules were created.")
	}
	if expectedDependencies != len(buildInfo.Modules[0].Dependencies) {
		t.Error("Incorrect number of dependencies found in the build-info, expected:", expectedDependencies, " Found:", len(buildInfo.Modules[0].Dependencies))
	}
	if expectedArtifacts != len(buildInfo.Modules[0].Artifacts) {
		t.Error("Incorrect number of artifacts found in the build-info, expected:", expectedArtifacts, " Found:", len(buildInfo.Modules[0].Artifacts))
	}
}

// This function counts the following:
// #1 The number of artifacts in the build-info JSON.
// #2 The number of artifact with the build.name and build.number properties.
// Validates that #1 == #2
func validateBuildInfoProperties(buildInfo buildinfo.BuildInfo, t *testing.T) {
	spec, flags := getSpecAndCommonFlags(tests.GetFilePath(tests.SearchGo))
	flags.SetArtifactoryDetails(artAuth)
	var resultItems []rtutils.ResultItem
	for i := 0; i < len(spec.Files); i++ {
		params, err := spec.Get(i).ToArtifatorySetPropsParams()
		if err != nil {
			t.Error(err)
		}
		currentResultItems, err := rtutils.SearchBySpecFiles(&rtutils.SearchParamsImpl{ArtifactoryCommonParams: params}, flags, rtutils.ALL)
		if err != nil {
			t.Error("Failed Searching files:", err)
		}
		resultItems = append(resultItems, currentResultItems...)
	}

	if len(buildInfo.Modules[0].Artifacts) != len(resultItems) {
		t.Error("Incorrect number of artifacts were uploaded, expected:", len(buildInfo.Modules[0].Artifacts), " Found:", len(resultItems))
	}

	for _, item := range resultItems {
		properties := item.Properties
		if len(properties) < 1 {
			t.Error("Failed setting properties on item:", item.GetItemRelativePath())
		}
		propertiesMap := convertSliceToMap(properties)
		value, contains := propertiesMap["build.name"]

		if !contains {
			t.Error("Failed setting up build.name property on", item.Name)
		}
		if value != buildInfo.Name {
			t.Error("Wrong value for build.name property on", item.Name, "expected", buildInfo.Name, "got", value)
		}

		value, contains = propertiesMap["build.number"]
		if !contains {
			t.Error("Failed setting up build.number property on", item.Name)
		}
		if value != buildInfo.Number {
			t.Error("Wrong value for build.number property on", item.Name, "expected", buildInfo.Number, "got", value)
		}

	}
}

func TestNugetResolve(t *testing.T) {
	initNugetTest(t)
	projects := []struct {
		project              string
		expectedDependencies int
	}{
		{"packagesconfig", 6},
		{"reference", 6},
	}
	for buildNumber, test := range projects {
		t.Run(test.project, func(t *testing.T) {
			testNugetCmd(t, createNugetProject(t, test.project), strconv.Itoa(buildNumber), 6)
		})
	}
	cleanBuildToolsTest()
}

func createNugetProject(t *testing.T, projectName string) string {
	projectSrc := filepath.Join(tests.GetTestResourcesPath(), "nuget", projectName)
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

func convertSliceToMap(props []rtutils.Property) map[string]string {
	propsMap := make(map[string]string)
	for _, item := range props {
		propsMap[item.Key] = item.Value
	}
	return propsMap
}

func testNugetCmd(t *testing.T, projectPath string, buildNumber string, expectedDependencies int) {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	err = os.Chdir(projectPath)
	if err != nil {
		t.Error(err)
	}

	artifactoryCli.Exec("nuget", "restore", tests.NugetRemoteRepo, "--build-name="+tests.NugetBuildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("bp", tests.NugetBuildName, buildNumber)

	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, tests.NugetBuildName, buildNumber, t, artHttpDetails)
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		t.Error("Nuget build info was not generated correctly, no modules were created.")
	}

	if expectedDependencies != len(buildInfo.Modules[0].Dependencies) {
		t.Error("Incorrect number of artifacts found in the build-info, expected:", expectedDependencies, " Found:", len(buildInfo.Modules[0].Dependencies))
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

	npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath := initNpmTest(t)
	var npmTests = []npmTestParams{
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateNpmInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmScopedProjectPath, validationFunc: validateNpmInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmNpmrcProjectPath, validationFunc: validateNpmInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateNpmInstall, npmArgs: "--production"},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateNpmInstall, npmArgs: "-only=dev"},
		{command: "npmi", repo: tests.NpmRemoteRepo, wd: npmNpmrcProjectPath, validationFunc: validateNpmPackInstall, npmArgs: "yaml"},
		{command: "npmp", repo: tests.NpmLocalRepo, wd: npmScopedProjectPath, validationFunc: validateNpmScopedPublish},
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
		artifactoryCli.Exec(npmTest.command, npmTest.repo, "--npm-args="+npmTest.npmArgs, "--build-name="+tests.NpmBuildName, "--build-number="+strconv.Itoa(i+1))
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
		t.Errorf("The file 'jfrog.npmrc.backup' was supposed to be deleted but it was not when runnung the configuration:\n%v", npmTest)
	}
}

func initNpmTest(t *testing.T) (npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath string) {
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
	prepareArtifactoryForNpmBuild(t, filepath.Dir(npmProjectPath))
	return npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath
}

func runAndValidateGradle(buildGradlePath, configFilePath string, t *testing.T) {
	buildConfig := &utils.BuildConfiguration{}
	err := gradle.Gradle("clean artifactoryPublish -b "+buildGradlePath, configFilePath, buildConfig)
	if err != nil {
		t.Error(err)
	}
	isExistInArtifactory(tests.GradleDeployedArtifacts, tests.GetFilePath(tests.SearchAllRepo1), t)
}

func createGradleProject(t *testing.T) string {
	srcBuildFile := filepath.Join(tests.GetTestResourcesPath(), "gradleproject", "build.gradle")
	targetPomPath := filepath.Join(tests.Out, "gradleproject")
	buildGradlePath, err := copyTemplateFile(srcBuildFile, targetPomPath, "build.gradle", false)
	if err != nil {
		t.Error(err)
	}

	srcSettingsFile := filepath.Join(tests.GetTestResourcesPath(), "gradleproject", "settings.gradle")
	_, err = copyTemplateFile(srcSettingsFile, targetPomPath, "settings.gradle", false)
	if err != nil {
		t.Error(err)
	}

	return buildGradlePath
}

func createMavenProject(t *testing.T) string {
	srcPomFile := filepath.Join(tests.GetTestResourcesPath(), "mavenproject", "pom.xml")
	targetPomPath := filepath.Join(tests.Out, "mavenproject")
	pomPath, err := copyTemplateFile(srcPomFile, targetPomPath, "pom.xml", false)
	if err != nil {
		t.Error(err)
	}
	return pomPath
}

func createNpmProject(t *testing.T, dir string) string {
	srcPackageJson := filepath.Join(tests.GetTestResourcesPath(), "npm", dir, "package.json")
	targetPackageJson := filepath.Join(tests.Out, dir)
	packageJson, err := copyTemplateFile(srcPackageJson, targetPackageJson, "package.json", false)
	if err != nil {
		t.Error(err)
	}

	// failure can be ignored
	npmrcExists, err := fileutils.IsFileExists(filepath.Join(filepath.Dir(srcPackageJson), ".npmrc"))
	if err != nil {
		t.Error(err)
	}

	if npmrcExists {
		if _, err = copyTemplateFile(filepath.Join(filepath.Dir(srcPackageJson), ".npmrc"), targetPackageJson, ".npmrc", false); err != nil {
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
}

func prepareArtifactoryForNpmBuild(t *testing.T, workingDirectory string) {
	if err := os.Chdir(workingDirectory); err != nil {
		t.Error(err)
	}

	caches := filepath.Join(workingDirectory, "caches")
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
		os.Unsetenv(config.JfrogHomeEnv)
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
	isExistInArtifactoryByProps(tests.NpmDeployedArtifacts,
		tests.NpmLocalRepo+"/*",
		fmt.Sprintf("build.name=%v;build.number=%v;build.timestamp=*", tests.NpmBuildName, npmTestParams.buildNumber), t)
	validateNpmCommonPublish(t, npmTestParams)
}

func validateNpmScopedPublish(t *testing.T, npmTestParams npmTestParams) {
	isExistInArtifactoryByProps(tests.NpmDeployedScopedArtifacts,
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
	validationFunc func(*testing.T, npmTestParams)
}

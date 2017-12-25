package main

import (
	"os"
	"testing"
	"path/filepath"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/commands"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
)

func InitBuildToolsTests() {
	if !*tests.TestBuildTools {
		return
	}
	os.Setenv("JFROG_CLI_OFFER_CONFIG", "false")
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(main, "jfrog rt", cred)
	createReposIfNeeded()
	cleanBuildToolsTest()
}

func CleanBuildToolsTests() {
	cleanBuildToolsTest()
	deleteRepos()
}

func TestMavenBuildWithServerID(t *testing.T) {
	initBuildToolsTest(t)

	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(tests.GetTestResourcesPath(), "buildspecs", tests.MavenServerIDConfig)
	createJfrogHomeConfig(t)

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
	mavenFlags := &utils.BuildConfigFlags{}
	err := commands.Mvn("clean install -f"+pomPath, configFilePath, mavenFlags)
	if err != nil {
		t.Error(err)
	}
	isExistInArtifactory(tests.MavenDeployedArtifacts, tests.GetFilePath(tests.SearchAllRepo1), t)
}

func TestGradleBuildWithServerID(t *testing.T) {
	initBuildToolsTest(t)

	buildGradlePath := createGradleProject(t)
	configFilePath := filepath.Join(tests.GetTestResourcesPath(), "buildspecs", tests.GradleServerIDConfig)
	createJfrogHomeConfig(t)

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

func TestNpm(t *testing.T) {
	initBuildToolsTest(t)
	npmi := "npm-install"
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	npmProjectPath, npmScopedProjectPath, npmNpmrcProjectPath := initNpmTest(t)
	var npmTests = []npmTestParams{
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmScopedProjectPath, validationFunc: validateInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmNpmrcProjectPath, validationFunc: validateInstall},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateInstall, npmArgs: "--production"},
		{command: npmi, repo: tests.NpmRemoteRepo, wd: npmProjectPath, validationFunc: validateInstall, npmArgs: "-only=dev",},
		{command: "npmi", repo: tests.NpmRemoteRepo, wd: npmNpmrcProjectPath, validationFunc: validatePackInstall, npmArgs: "yaml"},
		{command: "npmp", repo: tests.NpmLocalRepo, wd: npmScopedProjectPath, validationFunc: validateScopedPublish},
		{command: "npm-publish", repo: tests.NpmLocalRepo, wd: npmProjectPath, validationFunc: validatePublish},
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
		validateNpmrcFileInfo(t, npmTest, npmrcFileInfo , postTestFileInfo, err, postTestFileInfoErr)
	}

	err = os.Chdir(wd)
	if err != nil {
		t.Error(err)
	}
	cleanBuildToolsTest()
	deleteBuild(tests.NpmBuildName)
}

func validateNpmrcFileInfo( t *testing.T, npmTest npmTestParams, npmrcFileInfo, postTestNpmrcFileInfo os.FileInfo, err, postTestFileInfoErr error) {
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
		t.Errorf(".npmrc file was changed after running npm command! it was:\n%s\nnow it is:\n%s\nTest arguments are:\n%s", npmrcFileInfo, postTestNpmrcFileInfo, npmTest)
	}
	// make sue the temp .npmrc was deleted.
	bcpNpmrc, err := os.Stat("jfrog.npmrc.backup")
	if err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
	if bcpNpmrc != nil {
		t.Errorf("The file 'jfrog.npmrc.backup' was supposed to be deleted but it was not when runnung the configuration:\n%s", npmTest)
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
	buildConfigFlags := &utils.BuildConfigFlags{}
	err := commands.Gradle("clean artifactoryPublish -b "+buildGradlePath, configFilePath, buildConfigFlags)
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
		t.Skip("Skipping build tools test. To run build tools tests add the '--test.buildTools=true' option.")
	}
}

func prepareArtifactoryForNpmBuild(t *testing.T, workingDirectory string)  {
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
	if !*tests.TestBuildTools {
		return
	}
	os.Unsetenv(config.JfrogHomeEnv)
	cleanArtifactory()
	tests.CleanFileSystem()
}

func validateInstall(t *testing.T, npmTestParams npmTestParams) {
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

	buildInfoJson := getBuildInfo(t, npmTestParams)
	if len(expectedDependencies) != len(buildInfoJson.BuildInfo.Modules[0].Dependencies) {
		// The checksums are ignored when comparing the actual and the expected
		t.Errorf("npm install test with the arguments: \n%s \nexpected to have the following dependencies: \n%s \nbut has: \n%s",
			npmTestParams, expectedDependencies, buildInfoJson.BuildInfo.Modules[0].Dependencies)
	}
	for _, expectedDependency := range expectedDependencies {
		found := false
		for _, actualDependency := range buildInfoJson.BuildInfo.Modules[0].Dependencies {
			if actualDependency.Id == expectedDependency.id &&
				len(actualDependency.Scopes) == len(expectedDependency.scopes) &&
				actualDependency.Scopes[0] == expectedDependency.scopes[0] {
				found = true
				break
			}
		}
		if !found {
			// The checksums are ignored when comparing the actual and the expected
			t.Errorf("npm install test with the arguments: \n%s \nexpected to have the following dependencies: \n%s \nbut has: \n%s",
				npmTestParams, expectedDependencies, buildInfoJson.BuildInfo.Modules[0].Dependencies)
		}
	}
}

func validatePackInstall(t *testing.T, npmTestParams npmTestParams) {
	buildInfoJson := getBuildInfo(t, npmTestParams)
	if len(buildInfoJson.BuildInfo.Modules) > 0 {
		t.Errorf("npm install test with the arguments: \n%s \nexpected to have no modules but has: \n%s",
			npmTestParams, buildInfoJson.BuildInfo.Modules[0])
	}

	packageJsonFile, err := ioutil.ReadFile(npmTestParams.wd)
	if err != nil {
		t.Error(err)
	}

	var packageJson struct{ Dependencies map[string]string `json:"dependencies,omitempty"` }
	if err := json.Unmarshal(packageJsonFile, &packageJson); err != nil {
		t.Error(err)
	}

	if len(packageJson.Dependencies) != 2 || packageJson.Dependencies[npmTestParams.npmArgs] == "" {
		t.Errorf("npm install test with the arguments: \n%s \nexpected have the dependency %s in the following package.json file: \n%s",
			npmTestParams, npmTestParams.npmArgs, packageJsonFile)
	}
}

func validatePublish(t *testing.T, npmTestParams npmTestParams) {
	isExistInArtifactoryByProps(tests.NpmDeployedArtifacts,
		tests.NpmLocalRepo+"/*",
		fmt.Sprintf("build.name=%s;build.number=%s;build.timestamp=*", tests.NpmBuildName, npmTestParams.buildNumber), t)
	validateCommonPublish(t, npmTestParams)
}

func validateScopedPublish(t *testing.T, npmTestParams npmTestParams) {
	isExistInArtifactoryByProps(tests.NpmDeployedScopedArtifacts,
		tests.NpmLocalRepo+"/*",
		fmt.Sprintf("build.name=%s;build.number=%s;build.timestamp=*", tests.NpmBuildName, npmTestParams.buildNumber), t)
	validateCommonPublish(t, npmTestParams)
}

func validateCommonPublish(t *testing.T, npmTestParams npmTestParams) {
	buildInfoJson := getBuildInfo(t, npmTestParams)
	expectedArtifactName := "jfrog-cli-tests-1.0.0.tgz"
	if len(buildInfoJson.BuildInfo.Modules[0].Artifacts) != 1 {
		// The checksums are ignored when comparing the actual and the expected
		t.Errorf("npm publish test with the arguments: \n%s \nexpected to have the following artifact: \n%s \nbut has: \n%s",
			npmTestParams, expectedArtifactName, buildInfoJson.BuildInfo.Modules[0].Artifacts)
	}
	if buildInfoJson.BuildInfo.Modules[0].Artifacts[0].Name != expectedArtifactName {
		t.Errorf("npm publish test with the arguments: \n%s \nexpected to have the following artifact: \n%s \nbut has: \n%s",
			npmTestParams, expectedArtifactName, buildInfoJson.BuildInfo.Modules[0].Artifacts[0].Name)
	}
}

func getBuildInfo(t *testing.T, npmTestParams npmTestParams) struct{ BuildInfo buildinfo.BuildInfo `json:"buildInfo,omitempty"` } {
	_, body, _, err := httputils.SendGet(artifactoryDetails.Url+"api/build/"+tests.NpmBuildName+"/"+npmTestParams.buildNumber, true, artHttpDetails)
	if err != nil {
		t.Error(err)
	}

	var buildInfoJson struct{ BuildInfo buildinfo.BuildInfo `json:"buildInfo,omitempty"` }
	if err := json.Unmarshal(body, &buildInfoJson); err != nil {
		t.Error(err)
	}
	return buildInfoJson
}

type npmTestParams struct {
	command        string
	repo           string
	npmArgs        string
	wd             string
	buildNumber    string
	validationFunc func(*testing.T, npmTestParams)
}

package main

import (
	"os"
	"testing"
	"path/filepath"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/commands"
)

func InitBuildToolsTests() {
	if !*tests.TestBuildTools {
		return
	}
	os.Setenv("JFROG_CLI_OFFER_CONFIG", "false")
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(main, "jfrog rt", cred)
	if err := createReposIfNeeded(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	cleanBuildToolsTest()
}

func CleanBuildToolsTests() {
	cleanBuildToolsTest()
	if err := deleteRepos(); err != nil {
		log.Error(err)
	}
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
	srcConfigTemplate := filepath.Join(tests.GetTestResourcesPath(), "buildspecs", tests.MavenUseramePasswordTemplate)
	targetBuildSpecPath := filepath.Join(tests.Out, "buildspecs")
	configFilePath, err := copyTemplateFile(srcConfigTemplate, targetBuildSpecPath, tests.MavenUseramePasswordTemplate, true)
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

func initBuildToolsTest(t *testing.T) {
	if !*tests.TestBuildTools {
		t.Skip("Build tools are not being tested, skipping...")
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
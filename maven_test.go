package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-go/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
)

const mavenFlagName = "maven"

func TestMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t)

	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenServerIDConfig)
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, "")
	if err != nil {
		t.Error(err)
	}
	runAndValidateMaven(pomPath, configFilePath, t)
	cleanBuildToolsTest()
}

func TestNativeMavenBuildWithServerID(t *testing.T) {
	initMavenTest(t)
	pomPath := createMavenProject(t)
	configFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "buildspecs", tests.MavenConfig)
	destPath := filepath.Join(filepath.Dir(pomPath), ".jfrog", "projects")
	createConfigFile(destPath, configFilePath, t)
	oldHomeDir := changeWD(t, filepath.Dir(pomPath))
	pomPath = strings.Replace(pomPath, `\`, "/", -1) // Windows compatibility.
	runNewCli(t, "mvn", "clean install -f"+pomPath)
	err := os.Chdir(oldHomeDir)
	if err != nil {
		t.Error(err)
	}
	// Validate
	searchSpec, err := tests.CreateSpec(tests.SearchAllRepo1)
	if err != nil {
		t.Error(err)
	}
	isExistInArtifactory(tests.GetMavenDeployedArtifacts(), searchSpec, t)
	cleanBuildToolsTest()
}

func TestMavenBuildWithCredentials(t *testing.T) {
	if *tests.RtUser == "" || *tests.RtPassword == "" {
		t.SkipNow()
	}
	initMavenTest(t)
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

func createMavenProject(t *testing.T) string {
	srcPomFile := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "mavenproject", "pom.xml")
	pomPath, err := tests.ReplaceTemplateVariables(srcPomFile, "")
	if err != nil {
		t.Error(err)
	}
	return pomPath
}

func initMavenTest(t *testing.T) {
	if !*tests.TestMaven {
		t.Skip("Skipping Maven test. To run Maven test add the '-test.maven=true' option.")
	}
	createJfrogHomeConfig(t)
}

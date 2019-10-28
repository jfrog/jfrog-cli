package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
)

func InitBuildToolsTests() {
	os.Setenv(cliutils.OfferConfig, "false")
	os.Setenv(cliutils.ReportUsage, "false")
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", cred)
	createReposIfNeeded()
	cleanBuildToolsTest()
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

func initBuildToolsTest(t *testing.T) {
	if !*tests.TestBuildTools {
		t.Skip("Skipping build tools test. To run build tools tests add the '-test.buildTools=true' option.")
	}
	createJfrogHomeConfig(t)
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", cred)
	configArtifactoryCli = createConfigJfrogCLI(cred)
}

func cleanBuildToolsTest() {
	if *tests.TestBuildTools || *tests.TestGo || *tests.TestNuget || *tests.TestPip {
		os.Unsetenv(cliutils.JfrogHomeDirEnv)
		cleanArtifactory()
		tests.CleanFileSystem()
	}
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

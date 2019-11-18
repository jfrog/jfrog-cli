package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	artifactoryUtils "github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/prompt"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"gopkg.in/yaml.v2"
)

func TestMain(m *testing.M) {
	setupIntegrationTests()
	result := m.Run()
	tearDownIntegrationTests()
	os.Exit(result)
}

func setupIntegrationTests() {
	os.Setenv(cliutils.ReportUsage, "false")
	os.Setenv(cliutils.OfferConfig, "false")
	flag.Parse()
	log.SetDefaultLogger()

	if *tests.TestBintray {
		InitBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		initArtifactoryCli()
		InitArtifactoryTests()
	}
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip {
		if artifactoryCli == nil {
			initArtifactoryCli()
		}
		InitBuildToolsTests()
	}
	if *tests.TestDocker {
		if artifactoryCli == nil {
			initArtifactoryCli()
		}
	}
}

func tearDownIntegrationTests() {
	if *tests.TestBintray {
		CleanBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		CleanArtifactoryTests()
	}
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip {
		CleanBuildToolsTests()
	}
	os.Setenv(cliutils.OfferConfig, "true")
	os.Setenv(cliutils.ReportUsage, "true")
}

func InitBuildToolsTests() {
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

func cleanBuildToolsTest() {
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip {
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

func initArtifactoryCli() {
	*tests.RtUrl = utils.AddTrailingSlashIfNeeded(*tests.RtUrl)
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", cred)
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		configArtifactoryCli = createConfigJfrogCLI(cred)
	}
}

func testCreateConfFile(dirs []string, resolver, deployer string, t *testing.T, confType artifactoryUtils.ProjectType) error {
	var atDirectory string
	for _, atDir := range dirs {
		atDirectory = filepath.Dir(atDir)
		d, err := yaml.Marshal(&prompt.ConfigFile{
			CommonConfig: prompt.CommonConfig{
				Version:    1,
				ConfigType: confType.String(),
			},
			Resolver: artifactoryUtils.Repository{
				Repo:     resolver,
				ServerId: "default",
			},
			Deployer: artifactoryUtils.Repository{
				Repo:     deployer,
				ServerId: "default",
			},
		})
		if err != nil {
			return err
		}
		filePath := filepath.Join(atDirectory, ".jfrog", "projects")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			os.MkdirAll(filePath, 0777)
		}
		filePath = filepath.Join(filePath, confType.String()+".yaml")
		// Create config file to make sure the path is valid
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			t.Error("Couldn't create file:", err)
		}
		defer f.Close()
		_, err = f.Write(d)
		if err != nil {
			t.Error(err)
		}
	}
	return nil
}

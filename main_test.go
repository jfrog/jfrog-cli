package main

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	commandUtils "github.com/jfrog/jfrog-cli/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-cli/utils/tests"
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
	// Disable progress bar and confirmation messages.
	os.Setenv(cliutils.CI, "true")
	flag.Parse()
	log.SetDefaultLogger()

	if *tests.TestBintray {
		InitBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		InitArtifactoryTests()
	}
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip {
		InitBuildToolsTests()
	}
	if *tests.TestDocker {
		initArtifactoryCli()
	}
	if *tests.TestDistribution {
		InitDistributionTests()
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
	if *tests.TestDistribution {
		CleanDistributionTests()
	}
}

func InitBuildToolsTests() {
	initArtifactoryCli()
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
	assert.NoError(t, err)
	err = os.Setenv(cliutils.HomeDir, filepath.Join(wd, tests.Out, "jfroghome"))
	assert.NoError(t, err)
	jfrogHomePath, err := cliutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	_, err = tests.ReplaceTemplateVariables(templateConfigPath, jfrogHomePath)
	assert.NoError(t, err)
}

func prepareHomeDir(t *testing.T) (string, string) {
	oldHomeDir := os.Getenv(cliutils.HomeDir)
	// Populate cli config with 'default' server
	createJfrogHomeConfig(t)
	newHomeDir, err := cliutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	return oldHomeDir, newHomeDir
}

func cleanBuildToolsTest() {
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip {
		os.Unsetenv(cliutils.HomeDir)
		cleanArtifactory()
		tests.CleanFileSystem()
	}
}

func validateBuildInfo(buildInfo buildinfo.BuildInfo, t *testing.T, expectedDependencies int, expectedArtifacts int, moduleName string) {
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		assert.Fail(t, "build info was not generated correctly, no modules were created.")
		return
	}
	assert.Equal(t, moduleName, buildInfo.Modules[0].Id, "Unexpected module name")
	assert.Len(t, buildInfo.Modules[0].Dependencies, expectedDependencies, "Incorrect number of dependencies found in the build-info")
	assert.Len(t, buildInfo.Modules[0].Artifacts, expectedArtifacts, "Incorrect number of artifacts found in the build-info")
}

func initArtifactoryCli() {
	if artifactoryCli != nil {
		return
	}
	*tests.RtUrl = utils.AddTrailingSlashIfNeeded(*tests.RtUrl)
	cred := authenticate()
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", cred)
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		configArtifactoryCli = createConfigJfrogCLI(cred)
	}
}

func createConfigFileForTest(dirs []string, resolver, deployer string, t *testing.T, confType artifactoryUtils.ProjectType, global bool) error {
	var filePath string
	for _, atDir := range dirs {
		d, err := yaml.Marshal(&commandUtils.ConfigFile{
			Version:    1,
			ConfigType: confType.String(),
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
		if global {
			filePath = filepath.Join(atDir, "projects")

		} else {
			filePath = filepath.Join(atDir, ".jfrog", "projects")

		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			os.MkdirAll(filePath, 0777)
		}
		filePath = filepath.Join(filePath, confType.String()+".yaml")
		// Create config file to make sure the path is valid
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err, "Couldn't create file")
		defer f.Close()
		_, err = f.Write(d)
		assert.NoError(t, err)
	}
	return nil
}

func runCli(t *testing.T, args ...string) {
	rtCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err := rtCli.Exec(args...)
	assert.NoError(t, err)
}
func runCliWithLegacyBuildtoolsCmd(t *testing.T, args ...string) {
	rtCli := tests.NewJfrogCli(execMain, "jfrog rt", "")
	err := rtCli.LegacyBuildToolExec(args...)
	assert.NoError(t, err)
}

func changeWD(t *testing.T, newPath string) string {
	prevDir, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(newPath)
	assert.NoError(t, err)
	return prevDir
}

// Copy config file from `configFilePath` to `inDir`
func createConfigFile(inDir, configFilePath string, t *testing.T) {
	if _, err := os.Stat(inDir); os.IsNotExist(err) {
		os.MkdirAll(inDir, 0777)
	}
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, inDir)
	assert.NoError(t, err)
}

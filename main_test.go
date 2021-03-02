package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jfrog/jfrog-cli-core/common/commands"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/utils/log"

	commandUtils "github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestMain(m *testing.M) {
	setupIntegrationTests()
	result := m.Run()
	tearDownIntegrationTests()
	os.Exit(result)
}

func setupIntegrationTests() {
	os.Setenv(coreutils.ReportUsage, "false")
	// Disable progress bar and confirmation messages.
	os.Setenv(coreutils.CI, "true")

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
		InitDockerTests()
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
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip || *tests.TestDocker {
		CleanBuildToolsTests()
	}
	if *tests.TestDistribution {
		CleanDistributionTests()
	}
}

func InitBuildToolsTests() {
	initArtifactoryCli()
	cleanUpOldBuilds()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
	cleanBuildToolsTest()
}

func CleanBuildToolsTests() {
	cleanBuildToolsTest()
	deleteCreatedRepos()
}

func createJfrogHomeConfig(t *testing.T, encryptPassword bool) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Setenv(coreutils.HomeDir, filepath.Join(wd, tests.Out, "jfroghome"))
	assert.NoError(t, err)
	var credentials string
	if *tests.RtAccessToken != "" {
		credentials = "--access-token=" + *tests.RtAccessToken
	} else {
		credentials = "--user=" + *tests.RtUser + " --password=" + *tests.RtPassword
	}
	// Delete the default server if exist
	config, err := commands.GetConfig("default", false)
	if err == nil && config.ServerId != "" {
		err = tests.NewJfrogCli(execMain, "jfrog config", "").Exec("rm", "default", "--quiet")
	}
	err = tests.NewJfrogCli(execMain, "jfrog config", credentials).Exec("add", "default", "--interactive=false", "--artifactory-url="+*tests.RtUrl, "--enc-password="+strconv.FormatBool(encryptPassword))
	assert.NoError(t, err)
}

func prepareHomeDir(t *testing.T) (string, string) {
	oldHomeDir := os.Getenv(coreutils.HomeDir)
	// Populate cli config with 'default' server
	createJfrogHomeConfig(t, true)
	newHomeDir, err := coreutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	return oldHomeDir, newHomeDir
}

func cleanBuildToolsTest() {
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip {
		os.Unsetenv(coreutils.HomeDir)
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
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", authenticate(false))
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		configCli = createConfigJfrogCLI(authenticate(true))
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

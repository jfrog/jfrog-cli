package main

import (
	"flag"
	"fmt"
	buildinfo "github.com/jfrog/build-info-go/entities"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"

	commandUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/tests"
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
	err := os.Setenv(coreutils.ReportUsage, "false")
	if err != nil {
		clientlog.Error(fmt.Sprintf("Couldn't set env: %s. Error: %s", coreutils.ReportUsage, err.Error()))
		os.Exit(1)
	}
	// Disable progress bar and confirmation messages.
	err = os.Setenv(coreutils.CI, "true")
	if err != nil {
		clientlog.Error(fmt.Sprintf("Couldn't set env: %s. Error: %s", coreutils.CI, err.Error()))
		os.Exit(1)
	}
	flag.Parse()
	log.SetDefaultLogger()
	validateCmdAliasesUniqueness()
	if (*tests.TestArtifactory && !*tests.TestArtifactoryProxy) || *tests.TestArtifactoryProject {
		InitArtifactoryTests()
	}
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip || *tests.TestPipenv {
		InitBuildToolsTests()
	}
	if *tests.TestDocker {
		InitDockerTests()
	}
	if *tests.TestDistribution {
		InitDistributionTests()
	}
	if *tests.TestPlugins {
		InitPluginsTests()
	}
	if *tests.TestXray {
		InitXrayTests()
	}
}

func tearDownIntegrationTests() {
	if (*tests.TestArtifactory && !*tests.TestArtifactoryProxy) || *tests.TestArtifactoryProject {
		CleanArtifactoryTests()
	}
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip || *tests.TestPipenv || *tests.TestDocker {
		CleanBuildToolsTests()
	}
	if *tests.TestDistribution {
		CleanDistributionTests()
	}
	if *tests.TestPlugins {
		CleanPluginsTests()
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
	assert.NoError(t, err, "Failed to get current dir")
	clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, filepath.Join(wd, tests.Out, "jfroghome"))
	var credentials string
	if *tests.JfrogAccessToken != "" {
		credentials = "--access-token=" + *tests.JfrogAccessToken
	} else {
		credentials = "--user=" + *tests.JfrogUser + " --password=" + *tests.JfrogPassword
	}
	// Delete the default server if exist
	config, err := commands.GetConfig("default", false)
	if err == nil && config.ServerId != "" {
		err = tests.NewJfrogCli(execMain, "jfrog config", "").Exec("rm", "default", "--quiet")
	}
	*tests.JfrogUrl = utils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	err = tests.NewJfrogCli(execMain, "jfrog config", credentials).Exec("add", "default", "--interactive=false", "--artifactory-url="+*tests.JfrogUrl+tests.ArtifactoryEndpoint, "--xray-url="+*tests.JfrogUrl+tests.XrayEndpoint, "--enc-password="+strconv.FormatBool(encryptPassword))
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
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip || *tests.TestPipenv || *tests.TestDocker {
		err := os.Unsetenv(coreutils.HomeDir)
		if err != nil {
			clientlog.Error(fmt.Sprintf("Couldn't unset env: %s. Error: %s", coreutils.HomeDir, err.Error()))
		}
		tests.CleanFileSystem()
	}
}

func validateBuildInfo(buildInfo buildinfo.BuildInfo, t *testing.T, expectedDependencies int, expectedArtifacts int, moduleName string, moduleType buildinfo.ModuleType) {
	if buildInfo.Modules == nil || len(buildInfo.Modules) == 0 {
		assert.Fail(t, "build info was not generated correctly, no modules were created.")
		return
	}
	validateModule(buildInfo.Modules[0], t, expectedDependencies, expectedArtifacts, 0, moduleName, moduleType)
}

func validateModule(module buildinfo.Module, t *testing.T, expectedDependencies, expectedArtifacts, expectedExcludedArtifacts int, moduleName string, moduleType buildinfo.ModuleType) {
	assert.Equal(t, moduleName, module.Id, "Unexpected module name")
	assert.Len(t, module.Dependencies, expectedDependencies, "Incorrect number of dependencies found in the build-info")
	assert.Len(t, module.Artifacts, expectedArtifacts, "Incorrect number of artifacts found in the build-info")
	assert.Len(t, module.ExcludedArtifacts, expectedExcludedArtifacts, "Incorrect number of excluded artifacts found in the build-info")
	assert.Equal(t, module.Type, moduleType)
}

func validateSpecificModule(buildInfo buildinfo.BuildInfo, t *testing.T, expectedDependencies, expectedArtifacts, expectedExcludedArtifacts int, moduleName string, moduleType buildinfo.ModuleType) {
	for _, module := range buildInfo.Modules {
		if module.Id == moduleName {
			validateModule(module, t, expectedDependencies, expectedArtifacts, expectedExcludedArtifacts, moduleName, moduleType)
			return
		}
	}
}

func initArtifactoryCli() {
	if artifactoryCli != nil {
		return
	}
	*tests.JfrogUrl = utils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	artifactoryCli = tests.NewJfrogCli(execMain, "jfrog rt", authenticate(false))
	if (*tests.TestArtifactory && !*tests.TestArtifactoryProxy) || *tests.TestPlugins || *tests.TestArtifactoryProject {
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
			assert.NoError(t, os.MkdirAll(filePath, 0777))
		}
		filePath = filepath.Join(filePath, confType.String()+".yaml")
		// Create config file to make sure the path is valid
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err, "Couldn't create file")
		defer func() {
			assert.NoError(t, f.Close())
		}()
		_, err = f.Write(d)
		assert.NoError(t, err)
	}
	return nil
}

func runJfrogCli(t *testing.T, args ...string) {
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec(args...)
	assert.NoError(t, err)
}

func changeWD(t *testing.T, newPath string) string {
	prevDir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current dir")
	clientTestUtils.ChangeDirAndAssert(t, newPath)
	return prevDir
}

// Copy config file from `configFilePath` to `inDir`
func createConfigFile(inDir, configFilePath string, t *testing.T) {
	if _, err := os.Stat(inDir); os.IsNotExist(err) {
		assert.NoError(t, os.MkdirAll(inDir, 0777))
	}
	configFilePath, err := tests.ReplaceTemplateVariables(configFilePath, inDir)
	assert.NoError(t, err)
}

// setEnvVar sets an environment variable and returns a clean up function that reverts it.
func setEnvVar(t *testing.T, key, value string) (cleanUp func()) {
	oldValue, exist := os.LookupEnv(key)
	clientTestUtils.SetEnvAndAssert(t, key, value)

	if exist {
		return func() {
			clientTestUtils.SetEnvAndAssert(t, key, oldValue)
		}
	}

	return func() {
		clientTestUtils.UnSetEnvAndAssert(t, key)
	}
}

// Validate that all CLI commands' aliases are unique, and that two commands don't use the same alias.
func validateCmdAliasesUniqueness() {
	for _, command := range getCommands() {
		subcommands := command.Subcommands
		aliasesMap := map[string]bool{}
		for _, subcommand := range subcommands {
			for _, alias := range subcommand.Aliases {
				if aliasesMap[alias] {
					clientlog.Error(fmt.Sprintf("Duplicate alias '%s' found on %s %s command.", alias, command.Name, subcommand.Name))
					os.Exit(1)
				}
				aliasesMap[alias] = true
			}
		}
	}
}

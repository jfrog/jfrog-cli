package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	xrayutils "github.com/jfrog/jfrog-cli-core/v2/xray/utils"
	"github.com/urfave/cli"

	buildinfo "github.com/jfrog/build-info-go/entities"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"

	commandUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli/artifactory"
	"github.com/jfrog/jfrog-cli/inttestutils"
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
	if *tests.TestAccess {
		InitAccessTests()
		InitArtifactoryTests()
	}
	if *tests.TestTransfer {
		InitTransferTests()
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
	if *tests.TestTransfer {
		CleanTransferTests()
	}
}

func InitBuildToolsTests() {
	initArtifactoryCli()
	cleanUpOldBuilds()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
	cleanTestsHomeEnv()
}

func CleanBuildToolsTests() {
	cleanTestsHomeEnv()
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
		assert.NoError(t, err)
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

func cleanTestsHomeEnv() {
	os.Unsetenv(coreutils.HomeDir)
	tests.CleanFileSystem()
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
	if (*tests.TestArtifactory && !*tests.TestArtifactoryProxy) || *tests.TestPlugins || *tests.TestArtifactoryProject || *tests.TestAccess || *tests.TestTransfer {
		configCli = createConfigJfrogCLI(authenticate(true))
		platformCli = tests.NewJfrogCli(execMain, "jfrog", authenticate(false))
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
	assert.NoError(t, runJfrogCliWithoutAssertion(args...))
}

func runJfrogCliWithoutAssertion(args ...string) error {
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(args...)
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
	_, err := tests.ReplaceTemplateVariables(configFilePath, inDir)
	assert.NoError(t, err)
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

func testConditionalUpload(t *testing.T, execFunc func() error, validationSpecFileName string) {
	// Mock the scan function
	expectedErrMsg := "This error was expected"
	commandUtils.ConditionalUploadScanFunc = func(serverDetails *config.ServerDetails, fileSpec *spec.SpecFiles, threads int, scanOutputFormat xrayutils.OutputFormat) error {
		return errors.New(expectedErrMsg)
	}

	// Run conditional publish and verify the expected error returned.
	err := execFunc()
	assert.EqualError(t, err, expectedErrMsg)

	searchSpec, err := tests.CreateSpec(validationSpecFileName)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(nil, searchSpec, serverDetails, t)
}

func TestSearchSimilarCmds(t *testing.T) {
	testData := []struct {
		badCmdSyntax string
		searchIn     []cli.Command
		expectedRes  []string
	}{
		{"rtt", getCommands(), []string{"rt"}},
		{"bp", getCommands(), []string{"rt bp"}},
		{"asdfewrwqfaxf", getCommands(), []string{}},
		{"bpp", artifactory.GetCommands(), []string{"bpr", "bp", "pp"}},
		{"uplid", artifactory.GetCommands(), []string{"upload"}},
		{"downlo", artifactory.GetCommands(), []string{"download"}},
		{"ownload", artifactory.GetCommands(), []string{"download"}},
		{"ownload", artifactory.GetCommands(), []string{"download"}},
	}
	for _, testCase := range testData {
		actualRes := searchSimilarCmds(testCase.searchIn, testCase.badCmdSyntax)
		assert.ElementsMatch(t, actualRes, testCase.expectedRes)
	}
}

// Prepare and return the tool to check if the deployment view was printed after any command, by redirecting all the logs output into a buffer
// Returns:
// 1. assertDeploymentViewFunc - A function to check if the deployment view was printed to the screen after running jfrog cli command
// 2. cleanup func to be run at the end of the test
func initDeploymentViewTest(t *testing.T) (assertDeploymentViewFunc func(), cleanupFunc func()) {
	_, buffer, previousLog := tests.RedirectLogOutputToBuffer()
	revertFlags := clientlog.SetIsTerminalFlagsWithCallback(true)
	// Restore previous logger and terminal mode when the function returns
	assertDeploymentViewFunc = func() {
		output := buffer.Bytes()
		// Clean buffer for future runs.
		buffer.Truncate(0)
		expectedStringInOutput := "These files were uploaded:"
		assert.True(t, strings.Contains(string(output), expectedStringInOutput), fmt.Sprintf("cant find '%s' in '%s'", expectedStringInOutput, string(output)))
	}
	// Restore previous logger and terminal mode when the function returns
	cleanupFunc = func() {
		clientlog.SetLogger(previousLog)
		revertFlags()
	}
	return
}

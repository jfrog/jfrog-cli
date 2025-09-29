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

	artifactoryCLI "github.com/jfrog/jfrog-cli-artifactory/cli"
	"github.com/jfrog/jfrog-cli/artifactory"
	"github.com/stretchr/testify/require"

	buildinfo "github.com/jfrog/build-info-go/entities"
	commandUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
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
	if *tests.TestDocker || *tests.TestPodman || *tests.TestDockerScan {
		InitContainerTests()
	}
	if *tests.TestDistribution {
		InitDistributionTests()
	}
	if *tests.TestPlugins {
		InitPluginsTests()
	}
	if *tests.TestAccess {
		InitAccessTests()
	}
	if *tests.TestTransfer {
		InitTransferTests()
	}
	if *tests.TestLifecycle {
		InitLifecycleTests()
	}
}

func tearDownIntegrationTests() {
	if (*tests.TestArtifactory && !*tests.TestArtifactoryProxy) || *tests.TestArtifactoryProject {
		CleanArtifactoryTests()
	}
	if *tests.TestNpm || *tests.TestGradle || *tests.TestMaven || *tests.TestGo || *tests.TestNuget || *tests.TestPip || *tests.TestPipenv || *tests.TestDocker || *tests.TestPodman || *tests.TestDockerScan {
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
	if *tests.TestLifecycle {
		CleanLifecycleTests()
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
		err = coreTests.NewJfrogCli(execMain, "jfrog config", "").Exec("rm", "default", "--quiet")
		assert.NoError(t, err)
	}
	*tests.JfrogUrl = utils.AddTrailingSlashIfNeeded(*tests.JfrogUrl)
	err = coreTests.NewJfrogCli(execMain, "jfrog config", credentials).Exec("add", "default", "--interactive=false", "--url="+*tests.JfrogUrl, "--enc-password="+strconv.FormatBool(encryptPassword))
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
	if len(buildInfo.Modules) == 0 {
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
	artifactoryCli = coreTests.NewJfrogCli(execMain, "jfrog rt", authenticate(false))
	if (*tests.TestArtifactory && !*tests.TestArtifactoryProxy) || *tests.TestPlugins || *tests.TestArtifactoryProject ||
		*tests.TestAccess || *tests.TestTransfer || *tests.TestLifecycle {
		configCli = createConfigJfrogCLI(authenticate(true))
		platformCli = coreTests.NewJfrogCli(execMain, "jfrog", authenticate(false))
	}
}

func createConfigFileForTest(dirs []string, resolver, deployer string, t *testing.T, confType project.ProjectType, global bool) error {
	var filePath string
	for _, atDir := range dirs {
		d, err := yaml.Marshal(&commands.ConfigFile{
			Version:    1,
			ConfigType: confType.String(),
			Resolver: project.Repository{
				Repo:     resolver,
				ServerId: "default",
			},
			Deployer: project.Repository{
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
			assert.NoError(t, os.MkdirAll(filePath, 0o777))
		}
		filePath = filepath.Join(filePath, confType.String()+".yaml")
		// Create config file to make sure the path is valid
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		assert.NoError(t, err, "Couldn't create file")
		defer func(file *os.File) {
			assert.NoError(t, file.Close())
		}(f)
		_, err = f.Write(d)
		assert.NoError(t, err)
	}
	return nil
}

func runJfrogCli(t *testing.T, args ...string) {
	assert.NoError(t, runJfrogCliWithoutAssertion(args...))
}

func runJfrogCliWithoutAssertion(args ...string) error {
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
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
		assert.NoError(t, os.MkdirAll(inDir, 0o777))
	}
	_, err := tests.ReplaceTemplateVariables(configFilePath, inDir)
	assert.NoError(t, err)
}

// Validate that all CLI commands' aliases are unique, and that two commands don't use the same alias.
func validateCmdAliasesUniqueness() {
	cmds, err := getCommands()
	if err != nil {
		clientlog.Error(err)
		os.Exit(1)
	}
	for _, command := range cmds {
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

func testConditionalUpload(t *testing.T, execFunc func() error, searchSpec string, expectedDeployed ...string) {
	// Mock the scan function (failure) and verify the expected error returned.
	expectedErrMsg := "This error was expected"
	commandUtils.ConditionalUploadScanFunc = func(serverDetails *config.ServerDetails, fileSpec *spec.SpecFiles, threads int, scanOutputFormat format.OutputFormat) error {
		return errors.New(expectedErrMsg)
	}
	err := execFunc()
	assert.EqualError(t, err, expectedErrMsg)
	inttestutils.VerifyExistInArtifactory(nil, searchSpec, serverDetails, t)
	// Mock the scan function (success) and verify the expected artifacts deployed.
	commandUtils.ConditionalUploadScanFunc = func(serverDetails *config.ServerDetails, fileSpec *spec.SpecFiles, threads int, scanOutputFormat format.OutputFormat) error {
		return nil
	}
	err = execFunc()
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(expectedDeployed, searchSpec, serverDetails, t)
}

func TestSearchSimilarCmds(t *testing.T) {
	cmds, err := getCommands()
	assert.NoError(t, err)
	// fetch all legacy commands
	rtCmdsLegacy := artifactory.GetCommands()
	// fetch all new commands present as part of jfrog-cli-artifactory
	rtCmdsCombined, err := ConvertEmbeddedPlugin(artifactoryCLI.GetJfrogCliArtifactoryApp())
	assert.NoError(t, err)
	rtCmdsNew, err := fetchOnlyRTCmdsFromNewCmds(rtCmdsCombined)
	assert.NoError(t, err)
	// slice containing all artifactory commands
	rtCmdsNew = append(rtCmdsNew, rtCmdsLegacy...)
	assert.NoError(t, err)
	testData := []struct {
		badCmdSyntax string
		searchIn     []cli.Command
		expectedRes  []string
	}{
		{"rtt", cmds, []string{"rt"}},
		{"bp", cmds, []string{"rt bp"}},
		{"asdfewrwqfaxf", cmds, []string{}},
		{"bpp", rtCmdsNew, []string{"bpr", "bp", "pp"}},
		{"uplid", rtCmdsNew, []string{"upload"}},
		{"downlo", rtCmdsNew, []string{"download"}},
		{"ownload", rtCmdsNew, []string{"download"}},
		{"ownload", rtCmdsNew, []string{"download"}},
	}
	for _, testCase := range testData {
		actualRes := searchSimilarCmds(testCase.searchIn, testCase.badCmdSyntax)
		assert.ElementsMatch(t, actualRes, testCase.expectedRes)
	}
}

func fetchOnlyRTCmdsFromNewCmds(commands []cli.Command) ([]cli.Command, error) {
	var rtCmds cli.Commands
	for _, cmd := range commands {
		if cmd.Name == "rt" {
			rtCmds = cmd.Subcommands
			break
		}
	}
	if len(rtCmds) == 0 {
		return nil, errors.New("No rt commands found")
	}
	return rtCmds, nil
}

// Prepare and return the tool to check if the deployment view was printed after any command, by redirecting all the logs output into a buffer
// Returns:
// 1. assertDeploymentViewFunc - A function to check if the deployment view was printed to the screen after running jfrog cli command
// 2. cleanup func to be run at the end of the test
func initDeploymentViewTest(t *testing.T) (assertDeploymentViewFunc func(), cleanupFunc func()) {
	_, buffer, previousLog := coreTests.RedirectLogOutputToBuffer()
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

func deleteFilesFromRepo(t *testing.T, repoName string) {
	deleteSpec := spec.NewBuilder().Pattern(repoName).BuildSpec()
	_, _, err := tests.DeleteFiles(deleteSpec, serverDetails)
	// Mostly used during cleanup, no need to fail the test
	if err != nil {
		t.Logf("Error deleting files from repo %s: %+v", repoName, err)
	}
}

func TestIntro(t *testing.T) {
	buffer, _, previousLog := coreTests.RedirectLogOutputToBuffer()
	defer clientlog.SetLogger(previousLog)

	setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, "CI", "false")
	defer setEnvCallBack()

	runJfrogCli(t, "intro")
	assert.Contains(t, buffer.String(), "Thank you for installing version")
}

func TestSurvey(t *testing.T) {
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	_, contentErr, err := tests.GetCmdOutput(t, jfrogCli, "intro")
	require.NoError(t, err)
	assert.Contains(t, string(contentErr), "https://") // not doing more check as url can change
}

func TestGenerateAndLogTraceIdToken(t *testing.T) {
	traceIdToken, err := generateTraceIdToken()
	assert.NoError(t, err)
	assert.Len(t, traceIdToken, 16)
	_, err = strconv.ParseUint(traceIdToken, 16, 64)
	assert.NoError(t, err, "unexpected: trace ID token contains non-hex characters")
}

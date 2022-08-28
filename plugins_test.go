package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/buger/jsonparser"
	"github.com/jfrog/jfrog-cli-core/v2/plugins"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	pluginsutils "github.com/jfrog/jfrog-cli-core/v2/utils/plugins"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/plugins/commands/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

const officialPluginForTest = "rt-fs"
const officialPluginVersion = "v1.0.0"
const customPluginName = "custom-plugin"

func TestPluginInstallUninstallOfficialRegistry(t *testing.T) {
	initPluginsTest(t)
	// Create temp jfrog home
	cleanUpJfrogHome, err := coreTests.SetJfrogHome()
	if err != nil {
		return
	}
	defer cleanUpJfrogHome()

	// Set empty plugins server to run against official registry.
	oldServer := os.Getenv(utils.PluginsServerEnv)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, utils.PluginsServerEnv, oldServer)
	}()
	clientTestUtils.SetEnvAndAssert(t, utils.PluginsServerEnv, "")
	oldRepo := os.Getenv(utils.PluginsRepoEnv)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, utils.PluginsRepoEnv, oldRepo)
	}()
	clientTestUtils.SetEnvAndAssert(t, utils.PluginsRepoEnv, "")
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")

	// Try installing a plugin with specific version.
	err = installAndAssertPlugin(t, jfrogCli, officialPluginForTest, officialPluginVersion)
	if err != nil {
		return
	}

	// Try installing the latest version of the plugin. Also verifies replacement was successful.
	err = installAndAssertPlugin(t, jfrogCli, officialPluginForTest, "")
	if err != nil {
		return
	}

	// Uninstall plugin from home dir.
	err = jfrogCli.Exec("plugin", "uninstall", officialPluginForTest)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = verifyPluginInPluginsDir(t, officialPluginForTest, false, false)
	if err != nil {
		return
	}
}

func TestPluginInstallWithProgressBar(t *testing.T) {
	initPluginsTest(t)

	callback := tests.MockProgressInitialization()
	defer callback()

	// Create temp jfrog home
	cleanUpJfrogHome, err := coreTests.SetJfrogHome()
	if err != nil {
		return
	}
	defer cleanUpJfrogHome()

	// Set empty plugins server to run against official registry.
	oldServer := os.Getenv(utils.PluginsServerEnv)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, utils.PluginsServerEnv, oldServer)
	}()
	clientTestUtils.SetEnvAndAssert(t, utils.PluginsServerEnv, "")
	oldRepo := os.Getenv(utils.PluginsRepoEnv)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, utils.PluginsRepoEnv, oldRepo)
	}()
	clientTestUtils.SetEnvAndAssert(t, utils.PluginsRepoEnv, "")
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")

	// Try installing a plugin with specific version.
	err = installAndAssertPlugin(t, jfrogCli, officialPluginForTest, officialPluginVersion)
	if err != nil {
		return
	}

	// Try installing the latest version of the plugin. Also verifies replacement was successful.
	err = installAndAssertPlugin(t, jfrogCli, officialPluginForTest, "")
	if err != nil {
		return
	}
}

func installAndAssertPlugin(t *testing.T, jfrogCli *tests.JfrogCli, pluginName, pluginVersion string) error {
	// If version required, concat to plugin name
	identifier := pluginName
	if pluginVersion != "" {
		identifier += "@" + pluginVersion
	}

	// Install plugin from registry.
	err := jfrogCli.Exec("plugin", "install", identifier)
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	err = verifyPluginInPluginsDir(t, pluginName, true, false)
	if err != nil {
		return err
	}

	err = verifyPluginSignature(t, jfrogCli)
	if err != nil {
		return err
	}

	return verifyPluginVersion(t, jfrogCli, pluginVersion)
}

func verifyPluginSignature(t *testing.T, jfrogCli *tests.JfrogCli) error {
	// Get signature from plugin.
	content, err := tests.GetCmdOutput(t, jfrogCli, officialPluginForTest, plugins.SignatureCommandName)
	if err != nil {
		return err
	}

	// Extract the name from the output.
	name, err := jsonparser.GetString(content, "name")
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	assert.Equal(t, officialPluginForTest, name)

	// Extract the usage from the output.
	usage, err := jsonparser.GetString(content, "usage")
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	assert.NotEmpty(t, usage)
	return nil
}

func verifyPluginVersion(t *testing.T, jfrogCli *tests.JfrogCli, expectedVersion string) error {
	// Run plugin's -v command.
	content, err := tests.GetCmdOutput(t, jfrogCli, officialPluginForTest, "-v")
	if err != nil {
		return err
	}
	if expectedVersion != "" {
		assert.NoError(t, utils.AssertPluginVersion(string(content), expectedVersion))
	}
	return err
}

func verifyPluginInPluginsDir(t *testing.T, pluginName string, execShouldExist, resourcesShouldExist bool) error {
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	// Check plugins executable exists
	actualExists, err := fileutils.IsFileExists(filepath.Join(pluginsDir, pluginName, coreutils.PluginsExecDirName, pluginsutils.GetLocalPluginExecutableName(pluginName)), false)
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	if execShouldExist {
		assert.True(t, actualExists, "expected plugin executable to be preset in plugins dir after installing")
	} else {
		assert.False(t, actualExists, "expected plugin executable not to be preset in plugins dir after uninstalling")
	}

	// Check plugins resources directory exists
	actualExists, err = fileutils.IsFileExists(filepath.Join(pluginsDir, pluginName, coreutils.PluginsResourcesDirName, "dir", "resource"), false)
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	if resourcesShouldExist {
		assert.True(t, actualExists, "expected resources to be preset in plugins dir after installing")
	} else {
		assert.False(t, actualExists, "expected resources not to be preset in plugins dir after uninstalling")
	}
	return nil
}

func initPluginsTest(t *testing.T) {
	if !*tests.TestPlugins {
		t.Skip("Skipping Plugins test. To run Plugins test add the '-test.plugins=true' option.")
	}
}

func TestPublishInstallCustomServer(t *testing.T) {
	initPluginsTest(t)
	// Create temp jfrog home
	cleanUpJfrogHome, err := coreTests.SetJfrogHome()
	if err != nil {
		return
	}
	defer cleanUpJfrogHome()

	// Create server to use with the command.
	_, err = createServerConfigAndReturnPassphrase(t)
	defer deleteServerConfig(t)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Set plugins server to run against the configured server.
	oldServer := os.Getenv(utils.PluginsServerEnv)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, utils.PluginsServerEnv, oldServer)
	}()
	clientTestUtils.SetEnvAndAssert(t, utils.PluginsServerEnv, tests.ServerId)
	oldRepo := os.Getenv(utils.PluginsRepoEnv)
	defer func() {
		clientTestUtils.SetEnvAndAssert(t, utils.PluginsRepoEnv, oldRepo)
	}()
	clientTestUtils.SetEnvAndAssert(t, utils.PluginsRepoEnv, tests.RtRepo1)

	err = setOnlyLocalArc(t)
	if err != nil {
		return
	}

	// Test without resources directory
	testPublishAndInstall(t, false)
	// Create 'resources' directory for testing
	wd, err := os.Getwd()
	assert.NoError(t, err)
	exists, err := fileutils.IsDirExists(filepath.Join(wd, coreutils.PluginsResourcesDirName), false)
	assert.NoError(t, err)
	assert.False(t, exists)
	err = fileutils.CopyDir(filepath.Join(wd, "testdata", "plugins", "plugin-mock", coreutils.PluginsResourcesDirName), filepath.Join(wd, coreutils.PluginsResourcesDirName), true, nil)
	assert.NoError(t, err)
	// Test with resources directory
	testPublishAndInstall(t, true)
	assert.NoError(t, os.RemoveAll(filepath.Join(wd, "resources")))
}

func testPublishAndInstall(t *testing.T, resources bool) {
	// Publish the CLI as a plugin to the registry.
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	err := jfrogCli.Exec("plugin", "p", customPluginName, cliutils.GetVersion())
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = verifyPluginExistsInRegistry(t, resources)
	if err != nil {
		return
	}

	// Install plugin from registry.
	err = jfrogCli.Exec("plugin", "install", customPluginName)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = verifyPluginInPluginsDir(t, customPluginName, true, resources)
	if err != nil {
		return
	}
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	clientTestUtils.RemoveAllAndAssert(t, filepath.Join(pluginsDir, customPluginName))
	// Deleting plugin from Artifactory for other plugin's tests
	err = jfrogCli.Exec("rt", "del", tests.RtRepo1+"/"+customPluginName)
	assert.NoError(t, err)
}

func verifyPluginExistsInRegistry(t *testing.T, checkResources bool) error {
	searchFilePath, err := tests.CreateSpec(tests.SearchAllRepo1)
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	localArc, err := utils.GetLocalArchitecture()
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	expectedPath := path.Join(utils.GetPluginDirPath(customPluginName, cliutils.GetVersion(), localArc), pluginsutils.GetLocalPluginExecutableName(customPluginName))
	// Expected to find the plugin in the version and latest dir.
	expected := []string{
		expectedPath,
		strings.Replace(expectedPath, cliutils.GetVersion(), utils.LatestVersionName, 1),
	}
	// Add resources to expected paths if needed
	if checkResources {
		expectedPath = path.Join(utils.GetPluginDirPath(customPluginName, cliutils.GetVersion(), localArc), coreutils.PluginsResourcesDirName+".zip")
		expected = append(expected, expectedPath, strings.Replace(expectedPath, cliutils.GetVersion(), utils.LatestVersionName, 1))
	}
	inttestutils.VerifyExistInArtifactory(expected, searchFilePath, serverDetails, t)
	return nil
}

// Set the local architecture to be the only one in map to avoid building for all architectures.
func setOnlyLocalArc(t *testing.T) error {
	localArcName, err := utils.GetLocalArchitecture()
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	localArc := utils.ArchitecturesMap[localArcName]
	utils.ArchitecturesMap = map[string]utils.Architecture{
		localArcName: localArc,
	}
	return nil
}

func InitPluginsTests() {
	initArtifactoryCli()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

func CleanPluginsTests() {
	deleteCreatedRepos()
}

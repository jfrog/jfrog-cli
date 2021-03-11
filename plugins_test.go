package main

import (
	"github.com/buger/jsonparser"
	"github.com/jfrog/jfrog-cli-core/plugins"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/utils/tests"
	"github.com/jfrog/jfrog-cli/plugins/commands/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pluginTemplateName = "hello-frog"

func TestPluginInstallUninstallOfficialRegistry(t *testing.T) {
	initPluginsTest(t)
	// Create temp jfrog home
	coreTests.CleanUnitTestsJfrogHome()
	oldHome, err := coreTests.SetJfrogHome()
	if err != nil {
		return
	}
	defer os.Setenv(coreutils.HomeDir, oldHome)
	defer coreTests.CleanUnitTestsJfrogHome()

	// Set empty plugins server to run against official registry.
	oldServer := os.Getenv(utils.PluginsServerEnv)
	defer os.Setenv(utils.PluginsServerEnv, oldServer)
	assert.NoError(t, os.Setenv(utils.PluginsServerEnv, ""))
	oldRepo := os.Getenv(utils.PluginsRepoEnv)
	defer os.Setenv(utils.PluginsRepoEnv, oldRepo)
	assert.NoError(t, os.Setenv(utils.PluginsRepoEnv, ""))

	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")

	// Install plugin from registry.
	err = jfrogCli.Exec("plugin", "install", pluginTemplateName)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = verifyPluginInPluginsDir(t, true)
	if err != nil {
		return
	}

	err = verifyPluginSignature(t, jfrogCli)
	if err != nil {
		return
	}

	err = verifyPluginCommand(t, jfrogCli)
	if err != nil {
		return
	}

	// Uninstall plugin from home dir.
	err = jfrogCli.Exec("plugin", "uninstall", pluginTemplateName)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = verifyPluginInPluginsDir(t, false)
	if err != nil {
		return
	}
}

func verifyPluginSignature(t *testing.T, jfrogCli *tests.JfrogCli) error {
	// Get signature from plugin.
	content, err := getCmdOutput(t, jfrogCli, pluginTemplateName, plugins.SignatureCommandName)
	if err != nil {
		return err
	}

	// Extract the the name from the output.
	name, err := jsonparser.GetString(content, "name")
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	assert.Equal(t, pluginTemplateName, name)

	// Extract the the usage from the output.
	usage, err := jsonparser.GetString(content, "usage")
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	assert.NotEmpty(t, usage)
	return nil
}

func verifyPluginCommand(t *testing.T, jfrogCli *tests.JfrogCli) error {
	// Run plugin's command.
	content, err := getCmdOutput(t, jfrogCli, pluginTemplateName, "hello", "\"hello world\"", "--shout")
	if err != nil {
		return err
	}

	assert.Contains(t, string(content), "HELLO WORLD")
	return nil
}

func getCmdOutput(t *testing.T, jfrogCli *tests.JfrogCli, cmd ...string) ([]byte, error) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		assert.NoError(t, err)
		return nil, err
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		r.Close()
	}()
	err = jfrogCli.Exec(cmd...)
	if err != nil {
		assert.NoError(t, err)
		w.Close()
		return nil, err
	}
	err = w.Close()
	if err != nil {
		assert.NoError(t, err)
		return nil, err
	}
	content, err := ioutil.ReadAll(r)
	assert.NoError(t, err)
	return content, err
}

func verifyPluginInPluginsDir(t *testing.T, shouldExist bool) error {
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		assert.NoError(t, err)
		return err
	}

	actualExists, err := fileutils.IsFileExists(filepath.Join(pluginsDir, utils.GetLocalPluginExecutableName(pluginTemplateName)), false)
	if err != nil {
		assert.NoError(t, err)
		return err
	}
	if shouldExist {
		assert.True(t, actualExists, "expected plugin executable to be preset in plugins dir after installing")
	} else {
		assert.False(t, actualExists, "expected plugin executable not to be preset in plugins dir after uninstalling")
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
	coreTests.CleanUnitTestsJfrogHome()
	oldHome, err := coreTests.SetJfrogHome()
	if err != nil {
		return
	}
	defer os.Setenv(coreutils.HomeDir, oldHome)
	defer coreTests.CleanUnitTestsJfrogHome()

	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")

	// Create server to use with the command.
	_, err = createServerConfigAndReturnPassphrase()
	defer deleteServerConfig()
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Set plugins server to run against the configured server.
	oldServer := os.Getenv(utils.PluginsServerEnv)
	defer os.Setenv(utils.PluginsServerEnv, oldServer)
	assert.NoError(t, os.Setenv(utils.PluginsServerEnv, tests.RtServerId))
	oldRepo := os.Getenv(utils.PluginsRepoEnv)
	defer os.Setenv(utils.PluginsRepoEnv, oldRepo)
	assert.NoError(t, os.Setenv(utils.PluginsRepoEnv, tests.RtRepo1))

	err = setOnlyLocalArc(t)
	if err != nil {
		return
	}

	// Publish the CLI as a plugin to the registry.
	err = jfrogCli.Exec("plugin", "p", pluginTemplateName, cliutils.GetVersion())
	if err != nil {
		assert.NoError(t, err)
		return
	}

	err = verifyPluginExistsInRegistry(t)
	if err != nil {
		return
	}

	// Install plugin from registry.
	err = jfrogCli.Exec("plugin", "install", pluginTemplateName)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = verifyPluginInPluginsDir(t, true)
	if err != nil {
		return
	}
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.NoError(t, os.Remove(filepath.Join(pluginsDir, utils.GetLocalPluginExecutableName(pluginTemplateName))))
}

func verifyPluginExistsInRegistry(t *testing.T) error {
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
	expectedPath := utils.GetPluginPathInArtifactory(pluginTemplateName, cliutils.GetVersion(), localArc)
	// Expected to find the plugin in the version and latest dir.
	expected := []string{
		expectedPath,
		strings.Replace(expectedPath, cliutils.GetVersion(), utils.LatestVersionName, 1),
	}
	verifyExistInArtifactory(expected, searchFilePath, t)
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

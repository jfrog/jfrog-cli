package main

import (
	"github.com/buger/jsonparser"
	"github.com/jfrog/jfrog-cli-core/plugins"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/utils/tests"
	"github.com/jfrog/jfrog-cli/plugins/utils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const jfrogPluginTestsHome = ".jfrogCliPluginsTest"
const pluginTemplateName = "hello-frog"

func TestPluginFullCycle(t *testing.T) {
	initPluginsTest(t)
	// Create temp jfrog home
	oldHome, err := coreTests.SetJfrogHome(jfrogPluginTestsHome)
	if err != nil {
		return
	}
	defer os.Setenv(coreutils.HomeDir, oldHome)
	// Clean from previous tests.
	os.RemoveAll(jfrogPluginTestsHome)
	defer os.RemoveAll(jfrogPluginTestsHome)

	// Set CI to true to prevent interactive.
	oldCi := os.Getenv(coreutils.CI)
	os.Setenv(coreutils.CI, "true")
	defer os.Setenv(coreutils.CI, oldCi)

	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")

	// Install plugin from registry.
	err = jfrogCli.Exec("plugin install " + pluginTemplateName)
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
	err = jfrogCli.Exec("plugin uninstall " + pluginTemplateName)
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
	cmd := exec.Command("jfrog", pluginTemplateName, plugins.SignatureCommandName)
	content, err := cmd.Output()
	if err != nil {
		assert.NoError(t, err)
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
	cmd := exec.Command("jfrog", pluginTemplateName, "hello", "hello world", "--shout")
	content, err := cmd.Output()
	if err != nil {
		assert.NoError(t, err)
		return err
	}

	assert.Contains(t, string(content), "HELLO WORLD")
	return nil
}

func verifyPluginInPluginsDir(t *testing.T, shouldExist bool) error {
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		assert.NoError(t, err)
		return err
	}

	actualExists, err := fileutils.IsFileExists(filepath.Join(pluginsDir, utils.GetPluginExecutableName(pluginTemplateName)), false)
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

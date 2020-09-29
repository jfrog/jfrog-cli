package commands

import (
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/utils/log"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	log.SetDefaultLogger()
}

const jfrogPluginTestsHome = ".jfrogCliPluginsTest"
const pluginMockPath = "../../testdata/plugins/plugin-mock"

func TestRunUninstallCmd(t *testing.T) {
	// Create temp jfrog home
	oldHome, err := tests.SetJfrogHome(jfrogPluginTestsHome)
	if err != nil {
		return
	}
	defer os.Setenv(coreutils.HomeDir, oldHome)
	defer os.RemoveAll(jfrogPluginTestsHome)

	// Set CI to true to prevent interactive.
	oldCi := os.Getenv(coreutils.CI)
	os.Setenv(coreutils.CI, "true")
	defer os.Setenv(coreutils.CI, oldCi)

	// Create a file in plugins dir to mock a plugin.
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	err = fileutils.CopyFile(pluginsDir, pluginMockPath)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	pluginName := filepath.Base(pluginMockPath)
	pluginPath := filepath.Join(pluginsDir, pluginName)

	// Try uninstalling a plugin that doesn't exist.
	err = runUninstallCmd("non-existing-plugin")
	expectedError := generateNoPluginFoundError("non-existing-plugin")
	assert.Error(t, err)
	assert.Equal(t, expectedError.Error(), err.Error())
	exists, err := fileutils.IsFileExists(pluginPath, false)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.True(t, exists)

	// Try uninstalling a plugin that exists.
	assert.NoError(t, runUninstallCmd(pluginName))
	exists, err = fileutils.IsFileExists(pluginPath, false)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.False(t, exists)
}

package commands

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	"github.com/jfrog/jfrog-cli-core/v2/utils/plugins"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	log.SetDefaultLogger()
}

const pluginMockPath = "../../testdata/plugins/plugin-mock"

func TestRunUninstallCmd(t *testing.T) {
	// Create temp jfrog home
	cleanUpJfrogHome, err := coreTests.SetJfrogHome()
	if err != nil {
		return
	}
	defer cleanUpJfrogHome()

	// Set CI to true to prevent interactive.
	oldCi := os.Getenv(coreutils.CI)
	clientTestUtils.SetEnvAndAssert(t, coreutils.CI, "true")
	defer clientTestUtils.SetEnvAndAssert(t, coreutils.CI, oldCi)

	// Create a file in plugins dir to mock a plugin.
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		assert.NoError(t, err)
		return
	}
	pluginName := filepath.Base(pluginMockPath)
	err = fileutils.CopyDir(pluginMockPath, filepath.Join(pluginsDir, pluginName), true, nil)
	if err != nil {
		assert.NoError(t, err)
		return
	}

	pluginExePath := filepath.Join(pluginsDir, pluginName, coreutils.PluginsExecDirName, plugins.GetLocalPluginExecutableName(pluginName))
	pluginResourcePath := filepath.Join(pluginsDir, pluginName, coreutils.PluginsResourcesDirName, "dir", "resource")
	// Fix path for windows.
	assert.NoError(t, os.Rename(filepath.Join(pluginsDir, pluginName, coreutils.PluginsExecDirName, pluginName), pluginExePath))

	// Try uninstalling a plugin that doesn't exist.
	err = runUninstallCmd("non-existing-plugin")
	expectedError := generateNoPluginFoundError("non-existing-plugin")
	// Assert error was returned.
	assert.Error(t, err)
	assert.Error(t, expectedError)
	assert.Equal(t, expectedError.Error(), err.Error())
	exists, err := fileutils.IsFileExists(pluginExePath, false)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.True(t, exists)
	exists, err = fileutils.IsFileExists(pluginResourcePath, false)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.True(t, exists)

	// Try uninstalling a plugin that exists.
	assert.NoError(t, runUninstallCmd(pluginName))
	exists, err = fileutils.IsFileExists(pluginExePath, false)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.False(t, exists)
	exists, err = fileutils.IsFileExists(pluginResourcePath, false)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.False(t, exists)
}

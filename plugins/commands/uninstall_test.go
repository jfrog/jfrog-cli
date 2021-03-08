package commands

import (
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/utils/log"
	coreTests "github.com/jfrog/jfrog-cli-core/utils/tests"
	"github.com/jfrog/jfrog-cli/plugins/commands/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
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
	oldHome, err := coreTests.SetJfrogHome()
	if err != nil {
		return
	}
	defer os.Setenv(coreutils.HomeDir, oldHome)
	// Clean from previous tests.
	coreTests.CleanUnitTestsJfrogHome()
	defer coreTests.CleanUnitTestsJfrogHome()

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
	pluginExePath := filepath.Join(pluginsDir, utils.GetLocalPluginExecutableName(pluginName))
	// Fix path for windows.
	assert.NoError(t, os.Rename(filepath.Join(pluginsDir, pluginName), pluginExePath))

	// Try uninstalling a plugin that doesn't exist.
	err = runUninstallCmd("non-existing-plugin")
	expectedError := generateNoPluginFoundError("non-existing-plugin")
	assert.Error(t, err)
	if expectedError == nil {
		return
	}
	assert.Equal(t, expectedError.Error(), err.Error())
	exists, err := fileutils.IsFileExists(pluginExePath, false)
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
}

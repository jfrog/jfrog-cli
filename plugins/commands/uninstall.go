package commands

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/plugins"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

func UninstallCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	err := plugins.CheckPluginsVersionAndConvertIfNeeded()
	if err != nil {
		return err
	}
	return runUninstallCmd(c.Args().Get(0))
}

func runUninstallCmd(requestedPlugin string) error {
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		return err
	}
	requestedPluginDirPath := filepath.Join(pluginsDir, requestedPlugin)
	exists, err := fileutils.IsDirExists(requestedPluginDirPath, false)
	if err != nil {
		return err
	}
	if !exists {
		return generateNoPluginFoundError(requestedPlugin)
	}

	ci, err := clientutils.GetBoolEnvValue(coreutils.CI, false)
	if err != nil {
		return err
	}
	if !ci {
		if !coreutils.AskYesNo("Are you sure you want to uninstall plugin: \""+requestedPlugin+"\"?", false) {
			return nil
		}
	}
	return errorutils.CheckError(os.RemoveAll(requestedPluginDirPath))
}

func generateNoPluginFoundError(pluginName string) error {
	return errorutils.CheckErrorf("plugin '" + pluginName + "' could not be found")
}

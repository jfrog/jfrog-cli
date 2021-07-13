package commands

import (
	"errors"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/plugins/commands/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"os"
	"path/filepath"
)

func UninstallCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return runUninstallCmd(c.Args().Get(0))
}

func runUninstallCmd(requestedPlugin string) error {
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		return err
	}
	pluginExePath := filepath.Join(pluginsDir, utils.GetLocalPluginExecutableName(requestedPlugin))
	exists, err := fileutils.IsFileExists(pluginExePath, false)
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
	return os.Remove(pluginExePath)
}

func generateNoPluginFoundError(pluginName string) error {
	return errorutils.CheckError(errors.New("plugin '" + pluginName + "' could not be found"))
}

package commands

import (
	"errors"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/plugins/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"os"
	"path/filepath"
	"strings"
)

func UninstallCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return runUninstallCmd(c.Args().Get(0))
}

func runUninstallCmd(requestedPlugin string) error {
	plugins, err := utils.GetAllPluginsNames()
	if err != nil {
		return err
	}

	if len(plugins) == 0 {
		return errors.New("no plugins found")
	}

	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		if strings.EqualFold(plugin, requestedPlugin) {
			ci, err := clientutils.GetBoolEnvValue(coreutils.CI, false)
			if err != nil {
				return err
			}
			if !ci {
				if !coreutils.AskYesNo("Are you sure you want to delete plugin: \""+plugin+"\"?", false) {
					return nil
				}
			}
			return os.Remove(filepath.Join(pluginsDir, plugin))
		}
	}
	return generateNoPluginFoundError(requestedPlugin)
}

func generateNoPluginFoundError(pluginName string) error {
	return errors.New("plugin '" + pluginName + "' could not be found")
}

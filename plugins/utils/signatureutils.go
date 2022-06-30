package utils

import (
	"encoding/json"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	coreplugins "github.com/jfrog/jfrog-cli-core/v2/plugins"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/plugins"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const pluginsErrorPrefix = "jfrog cli plugins: "
const pluginsCategory = "Plugins"

// Gets all the installed plugins' signatures by looping over the plugins' dir.
func getPluginsSignatures() ([]*components.PluginSignature, error) {
	var signatures []*components.PluginSignature
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		return signatures, errorutils.CheckError(err)
	}
	plugins, err := coreutils.GetPluginsDirContent()
	if err != nil {
		return signatures, errorutils.CheckError(err)
	}
	var finalErr error
	for _, p := range plugins {
		// Skip 'plugins.yml'
		if p.Name() == coreutils.JfrogPluginsFileName {
			continue
		}
		if !p.IsDir() {
			logSkippablePluginsError("unexpected file in plugins directory", p.Name(), nil)
			continue
		}
		pluginName := strings.TrimSuffix(p.Name(), filepath.Ext(p.Name()))
		execPath := filepath.Join(pluginsDir, pluginName, coreutils.PluginsExecDirName, p.Name())
		output, err := gofrogcmd.RunCmdOutput(
			&PluginExecCmd{
				execPath,
				[]string{coreplugins.SignatureCommandName},
			})
		if err != nil {
			finalErr = err
			logSkippablePluginsError("failed getting signature from plugin", pluginName, err)
			continue
		}
		curSignature := new(components.PluginSignature)
		err = json.Unmarshal([]byte(output), &curSignature)
		if err != nil {
			finalErr = err
			logSkippablePluginsError("failed unmarshalling signature from plugin", pluginName, err)
			continue
		}
		curSignature.ExecutablePath = execPath
		signatures = append(signatures, curSignature)
	}
	return signatures, finalErr
}

func logSkippablePluginsError(msg, pluginName string, err error) {
	log.Error(fmt.Sprintf("%s%s: '%s'. Skiping...", pluginsErrorPrefix, msg, pluginName))
	if err != nil {
		log.Error("Error was: " + err.Error())
	}
}

// Converts signatures to commands to be appended to the CLI commands.
func signaturesToCommands(signatures []*components.PluginSignature) []cli.Command {
	var commands []cli.Command
	for _, sig := range signatures {
		commands = append(commands, cli.Command{
			Name:            sig.Name,
			Usage:           sig.Usage,
			SkipFlagParsing: true,
			Action:          getAction(*sig),
			Category:        pluginsCategory,
		})
	}
	return commands
}

func getAction(sig components.PluginSignature) func(*cli.Context) error {
	return func(c *cli.Context) error {
		cmd := exec.Command(sig.ExecutablePath, cliutils.ExtractCommand(c)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
}

func GetPlugins() []cli.Command {
	err := plugins.CheckPluginsVersionAndConvertIfNeeded()
	if err != nil {
		log.Error("failed adding certain plugins as commands. Last error: " + err.Error())
		return []cli.Command{}
	}
	signatures, err := getPluginsSignatures()
	if err != nil {
		// Intentionally ignoring error to avoid failing if running other commands.
		log.Error("failed adding certain plugins as commands. Last error: " + err.Error())
		return []cli.Command{}
	}
	return signaturesToCommands(signatures)
}

package utils

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/plugins"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const pluginsErrorPrefix = "jfrog cli plugins: "

// Gets all the installed plugins' signatures by looping over the plugins dir.
func getPluginsSignatures() ([]*components.PluginSignature, error) {
	var signatures []*components.PluginSignature
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		return signatures, err
	}
	exists, err := fileutils.IsDirExists(pluginsDir, false)
	if err != nil || !exists {
		return signatures, err
	}

	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		return signatures, errorutils.CheckError(err)
	}

	var finalErr error
	for _, f := range files {
		if f.IsDir() {
			logSkippablePluginsError("unexpected directory in plugins directory", f.Name(), nil)
			continue
		}
		pluginName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		execPath := filepath.Join(pluginsDir, f.Name())
		output, err := gofrogcmd.RunCmdOutput(
			&PluginExecCmd{
				execPath,
				[]string{plugins.SignatureCommandName},
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
	signatures, err := getPluginsSignatures()
	if err != nil {
		// Intentionally ignoring error to avoid failing if running other commands.
		log.Error("failed adding certain plugins as commands. Last error: " + err.Error())
		return []cli.Command{}
	}
	return signaturesToCommands(signatures)
}

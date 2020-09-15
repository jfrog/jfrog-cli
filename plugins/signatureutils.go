package plugins

import (
	"encoding/json"
	"github.com/codegangsta/cli"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/plugins"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

// Command used to extract a signature plugin, and to execute plugin commands.
type SignatureCmd struct {
	execPath string
	Command  []string
}

func (signatureCmd *SignatureCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, signatureCmd.execPath)
	cmd = append(cmd, signatureCmd.Command...)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (signatureCmd *SignatureCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (signatureCmd *SignatureCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (signatureCmd *SignatureCmd) GetErrWriter() io.WriteCloser {
	return nil
}

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

	files, err := ioutil.ReadDir(pluginsDir)
	if err != nil {
		return signatures, errorutils.CheckError(err)
	}

	var finalErr error
	for _, f := range files {
		execPath := filepath.Join(pluginsDir, f.Name())
		output, err := gofrogcmd.RunCmdOutput(
			&SignatureCmd{
				execPath,
				[]string{plugins.SignatureCommandName},
			})
		if err != nil {
			finalErr = err
			log.Error("failed getting signature from plugin: " + f.Name())
			continue
		}
		curSignature := new(components.PluginSignature)
		err = json.Unmarshal([]byte(output), &curSignature)
		if err != nil {
			finalErr = err
			log.Error("failed unmarshalling signature from plugin: " + f.Name())
		}
		curSignature.ExecutablePath = execPath
		signatures = append(signatures, curSignature)
	}
	return signatures, finalErr
}

// Converts signatures to commands to be appended to the CLI commands.
func signaturesToCommands(signatures []*components.PluginSignature) []cli.Command {
	var commands []cli.Command
	for _, sig := range signatures {
		commands = append(commands, cli.Command{
			Name:            sig.Name,
			Usage:           sig.Usage,
			SkipFlagParsing: true,
			Action: func(c *cli.Context) error {
				return gofrogcmd.RunCmd(
					&SignatureCmd{
						sig.ExecutablePath,
						c.Args()})
			},
		})
	}
	return commands
}

func GetPlugins() []cli.Command {
	signatures, err := getPluginsSignatures()
	if err != nil {
		// Intentionally ignoring error to avoid failing if running other commands.
		log.Error("failed adding plugins as commands: " + err.Error())
		return []cli.Command{}
	}
	return signaturesToCommands(signatures)
}

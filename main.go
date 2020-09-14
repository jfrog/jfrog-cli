package main

import (
	"github.com/jfrog/jfrog-cli-core/docs/common"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/utils/log"
	"github.com/jfrog/jfrog-cli/plugins"
	"os"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli/artifactory"
	"github.com/jfrog/jfrog-cli/bintray"
	"github.com/jfrog/jfrog-cli/completion"
	"github.com/jfrog/jfrog-cli/missioncontrol"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/xray"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientLog "github.com/jfrog/jfrog-client-go/utils/log"
)

func main() {
	log.SetDefaultLogger()
	err := execMain()
	if cleanupErr := fileutils.CleanOldDirs(); cleanupErr != nil {
		clientLog.Warn(cleanupErr)
	}
	coreutils.ExitOnErr(err)
}

func execMain() error {
	// Set JFrog CLI's user-agent on the jfrog-client-go.
	clientutils.SetUserAgent(coreutils.GetUserAgent())

	app := cli.NewApp()
	app.Name = "jfrog"
	app.Usage = "See https://github.com/jfrog/jfrog-cli-go for usage instructions."
	app.Version = cliutils.GetVersion()
	args := os.Args
	app.EnableBashCompletion = true
	app.Commands = getCommands()
	cli.CommandHelpTemplate = common.CommandHelpTemplate
	cli.AppHelpTemplate = common.AppHelpTemplate
	cli.SubcommandHelpTemplate = common.SubcommandHelpTemplate
	err := app.Run(args)
	return err
}

func getCommands() []cli.Command {
	cliNameSpaces := []cli.Command{
		{
			Name:        cliutils.CmdArtifactory,
			Usage:       "Artifactory commands",
			Subcommands: artifactory.GetCommands(),
		},
		{
			Name:        cliutils.CmdBintray,
			Usage:       "Bintray commands",
			Subcommands: bintray.GetCommands(),
		},
		{
			Name:        cliutils.CmdMissionControl,
			Usage:       "Mission Control commands",
			Subcommands: missioncontrol.GetCommands(),
		},
		{
			Name:        cliutils.CmdXray,
			Usage:       "Xray commands",
			Subcommands: xray.GetCommands(),
		},
		{
			Name:        cliutils.CmdCompletion,
			Usage:       "Generate autocomplete scripts",
			Subcommands: completion.GetCommands(),
		},
	}
	return append(cliNameSpaces, plugins.GetPlugins()...)
}

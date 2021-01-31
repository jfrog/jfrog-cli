package main

import (
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/utils/log"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/plugins"
	"github.com/jfrog/jfrog-cli/plugins/utils"
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

const commandHelpTemplate string = `{{.HelpName}}{{if .UsageText}}
Arguments:
{{.UsageText}}
{{end}}{{if .VisibleFlags}}
Options:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}{{if .ArgsUsage}}
Environment Variables:
{{.ArgsUsage}}{{end}}

`

const appHelpTemplate string = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} [arguments...]{{end}}
   {{if .Version}}
VERSION:
   {{.Version}}
   {{end}}{{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
Environment Variables:
` + common.GlobalEnvVars + `{{end}}

`

const subcommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}}[arguments...]

COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
Environment Variables:
` + common.GlobalEnvVars + `{{end}}

`

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
	app.Usage = "See https://github.com/jfrog/jfrog-cli for usage instructions."
	app.Version = cliutils.GetVersion()
	args := os.Args
	app.EnableBashCompletion = true
	app.Commands = getCommands()
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.AppHelpTemplate = appHelpTemplate
	cli.SubcommandHelpTemplate = subcommandHelpTemplate
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
		{
			Name:        cliutils.CmdPlugin,
			Usage:       "Plugins commands",
			Subcommands: plugins.GetCommands(),
		},
	}
	return append(cliNameSpaces, utils.GetPlugins()...)
}

package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/bintray"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/missioncontrol"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/xray"
	"github.com/jfrog/jfrog-client-go/utils"
	"os"
)

const commandHelpTemplate string = `{{.HelpName}}{{if .UsageText}}
Arguments:
{{.UsageText}}
{{end}}{{if .Flags}}
Options:
	{{range .Flags}}{{.}}
	{{end}}{{end}}{{if .ArgsUsage}}
Environment Variables:
{{.ArgsUsage}}{{end}}

`

const appHelpTemplate string = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .Flags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} [arguments...]{{end}}
   {{if .Version}}
VERSION:
   {{.Version}}
   {{end}}{{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
Environment Variables:
` + common.GlobalEnvVars + `{{end}}

`

const subcommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{.HelpName}} command{{if .Flags}} [command options]{{end}}[arguments...]

COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
Environment Variables:
` + common.GlobalEnvVars + `{{end}}

`

func main() {
	// Set JFrog CLI's user-agent on the jfrog-client-go.
	utils.SetUserAgent(fmt.Sprintf("%s/%s", cliutils.ClientAgent, cliutils.GetVersion()))

	app := cli.NewApp()
	app.Name = "jfrog"
	app.Usage = "See https://github.com/jfrog/jfrog-cli-go for usage instructions."
	app.Version = cliutils.GetVersion()
	args := os.Args
	app.Commands = getCommands()
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.AppHelpTemplate = appHelpTemplate
	cli.SubcommandHelpTemplate = subcommandHelpTemplate
	err := app.Run(args)
	cliutils.ExitOnErr(err)
}

func getCommands() []cli.Command {
	return []cli.Command{
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
	}
}

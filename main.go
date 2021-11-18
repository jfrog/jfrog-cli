package main

import (
	"fmt"
	"github.com/jfrog/jfrog-cli/distribution"
	"github.com/jfrog/jfrog-cli/scan"
	"os"

	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	"github.com/jfrog/jfrog-cli/config"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/docs/general/cisetup"
	commands "github.com/jfrog/jfrog-cli/general/cisetup"
	"github.com/jfrog/jfrog-cli/plugins"
	"github.com/jfrog/jfrog-cli/plugins/utils"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli/artifactory"
	"github.com/jfrog/jfrog-cli/buildtools"
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

const subcommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Description}}

USAGE:
	{{if .Usage}}{{.Usage}}{{ "\n\t" }}{{end}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}}[arguments...]

COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Description}}
   {{end}}{{if .VisibleFlags}}{{if .ArgsUsage}}
Arguments:
{{.ArgsUsage}}{{ "\n" }}{{end}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
{{end}}
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
	clientutils.SetUserAgent(coreutils.GetCliUserAgent())

	app := cli.NewApp()
	app.Name = "jfrog"
	app.Usage = "See https://github.com/jfrog/jfrog-cli for usage instructions."
	app.Version = cliutils.GetVersion()
	args := os.Args
	cliutils.SetCliExecutableName(args[0])
	app.EnableBashCompletion = true
	app.Commands = getCommands()
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.AppHelpTemplate = getAppHelpTemplate()
	cli.SubcommandHelpTemplate = subcommandHelpTemplate
	err := app.Run(args)
	return err
}

const otherCategory = "Other"

func getCommands() []cli.Command {
	cliNameSpaces := []cli.Command{
		{
			Name:        cliutils.CmdArtifactory,
			Description: "Artifactory commands",
			Subcommands: artifactory.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdMissionControl,
			Description: "Mission Control commands",
			Subcommands: missioncontrol.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdXray,
			Description: "Xray commands",
			Subcommands: xray.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdDistribution,
			Description: "Distribution commands",
			Subcommands: distribution.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdCompletion,
			Description: "Generate autocomplete scripts",
			Subcommands: completion.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdPlugin,
			Description: "Plugins handling commands",
			Subcommands: plugins.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdConfig,
			Aliases:     []string{"c"},
			Description: "Config commands",
			Subcommands: config.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:         "ci-setup",
			Usage:        cisetup.GetDescription(),
			HelpName:     corecommon.CreateUsage("ci-setup", cisetup.GetDescription(), cisetup.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     otherCategory,
			Action: func(c *cli.Context) error {
				return commands.RunCiSetupCmd()
			},
		},
		{
			Name:        cliutils.CmdOptions,
			Description: "Show all supported environment variables",
			Category:    otherCategory,
			Action: func(*cli.Context) {
				fmt.Printf(common.GetGlobalEnvVars())
			},
		},
	}
	allCommands := append(cliNameSpaces, utils.GetPlugins()...)
	allCommands = append(allCommands, scan.GetCommands()...)
	return append(allCommands, buildtools.GetCommands()...)
}

func getAppHelpTemplate() string {
	return `NAME:
   ` + coreutils.GetCliExecutableName() + ` - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} [arguments...]{{end}}
   {{if .Version}}
VERSION:
   {{.Version}}
   {{end}}{{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:{{range .VisibleCategories}}{{if .Name}}

   {{.Name}}:{{end}}{{range .Commands}}
     {{join .Names ", "}}{{ "\t" }}{{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
{{end}}
`
}

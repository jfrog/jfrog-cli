package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/jfrog/jfrog-cli/distribution"
	"github.com/jfrog/jfrog-cli/scan"

	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	"github.com/jfrog/jfrog-cli/config"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/docs/general/cisetup"
	cisetupcommand "github.com/jfrog/jfrog-cli/general/cisetup"
	"github.com/jfrog/jfrog-cli/general/envsetup"
	"github.com/jfrog/jfrog-cli/general/project"
	"github.com/jfrog/jfrog-cli/plugins"
	"github.com/jfrog/jfrog-cli/plugins/utils"

	"github.com/jfrog/jfrog-cli/artifactory"
	"github.com/jfrog/jfrog-cli/buildtools"
	"github.com/jfrog/jfrog-cli/completion"
	"github.com/jfrog/jfrog-cli/missioncontrol"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/xray"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientLog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
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
	app.Name = "jf"
	app.Usage = "See https://github.com/jfrog/jfrog-cli for usage instructions."
	app.Version = cliutils.GetVersion()
	args := os.Args
	cliutils.SetCliExecutableName(args[0])
	app.EnableBashCompletion = true
	app.Commands = getCommands()
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.AppHelpTemplate = getAppHelpTemplate()
	cli.SubcommandHelpTemplate = subcommandHelpTemplate
	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "'"+command+"' is not a jf command. See --help\n")
		if bestSimilarity := getSimilarCmds(c, command); len(bestSimilarity) > 0 {
			text := "The most similar "
			if len(bestSimilarity) == 1 {
				text += "command is\n\t" + bestSimilarity[0]
			} else {
				text += "commands are\n\t" + strings.Join(bestSimilarity, ",")
			}
			fmt.Fprintln(c.App.Writer, text)
		}
		os.Exit(1)
	}
	err := app.Run(args)
	return err
}

// Detects typos and can identify exactly one or more valid commands similar to the error command.
func getSimilarCmds(c *cli.Context, toCompare string) (bestSimilarity []string) {
	// Set the max diff
	minDistance := 2
	for _, c := range c.App.Commands {
		for _, n := range c.Names() {
			distance := levenshtein.ComputeDistance(n, toCompare)
			if distance == minDistance {
				bestSimilarity = append(bestSimilarity, n)
			}
			if distance < minDistance {
				minDistance = distance
				bestSimilarity = []string{n}
			}
		}
	}
	return
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
			Name:        cliutils.CmdProject,
			Description: "Project commands",
			Subcommands: project.GetCommands(),
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
				return cisetupcommand.RunCiSetupCmd()
			},
		},
		{
			Name:     "setup",
			HideHelp: true,
			Hidden:   true,
			Action: func(c *cli.Context) error {
				return envsetup.RunEnvSetupCmd()
			},
		},
		{
			Name:        cliutils.CmdOptions,
			Description: "Show all supported environment variables",
			Category:    otherCategory,
			Action: func(*cli.Context) {
				fmt.Println(common.GetGlobalEnvVars())
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
   {{end}}{{if .VisibleCommands}}
COMMANDS:{{range .VisibleCategories}}{{if .Name}}

   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{ "\t" }}{{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
{{end}}
`
}

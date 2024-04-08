package main

import (
	"fmt"
	"github.com/jfrog/jfrog-cli/general/ai"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	setupcore "github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	securityCLI "github.com/jfrog/jfrog-cli-security/cli"
	"github.com/jfrog/jfrog-cli/artifactory"
	"github.com/jfrog/jfrog-cli/buildtools"
	"github.com/jfrog/jfrog-cli/completion"
	"github.com/jfrog/jfrog-cli/config"
	"github.com/jfrog/jfrog-cli/distribution"
	"github.com/jfrog/jfrog-cli/docs/common"
	aiDocs "github.com/jfrog/jfrog-cli/docs/general/ai"
	"github.com/jfrog/jfrog-cli/docs/general/cisetup"
	loginDocs "github.com/jfrog/jfrog-cli/docs/general/login"
	tokenDocs "github.com/jfrog/jfrog-cli/docs/general/token"
	cisetupcommand "github.com/jfrog/jfrog-cli/general/cisetup"
	"github.com/jfrog/jfrog-cli/general/envsetup"
	"github.com/jfrog/jfrog-cli/general/login"
	"github.com/jfrog/jfrog-cli/general/project"
	"github.com/jfrog/jfrog-cli/general/token"
	"github.com/jfrog/jfrog-cli/lifecycle"
	"github.com/jfrog/jfrog-cli/missioncontrol"
	"github.com/jfrog/jfrog-cli/pipelines"
	"github.com/jfrog/jfrog-cli/plugins"
	"github.com/jfrog/jfrog-cli/plugins/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"golang.org/x/exp/slices"
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
   {{.HelpName}} - {{.Usage}}

USAGE:
	{{if .Usage}}{{.Usage}}{{ "\n\t" }}{{end}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} [arguments...]

COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .VisibleFlags}}{{if .ArgsUsage}}
Arguments:
{{.ArgsUsage}}{{ "\n" }}{{end}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
{{end}}
`

const jfrogAppName = "jf"

func main() {
	log.SetDefaultLogger()
	err := execMain()
	if cleanupErr := fileutils.CleanOldDirs(); cleanupErr != nil {
		clientlog.Warn(cleanupErr)
	}
	coreutils.ExitOnErr(err)
}

func execMain() error {
	// Set JFrog CLI's user-agent on the jfrog-client-go.
	clientutils.SetUserAgent(coreutils.GetCliUserAgent())

	app := cli.NewApp()
	app.Name = jfrogAppName
	app.Usage = "See https://github.com/jfrog/jfrog-cli for usage instructions."
	app.Version = cliutils.GetVersion()
	args := os.Args
	cliutils.SetCliExecutableName(args[0])
	app.EnableBashCompletion = true
	commands, err := getCommands()
	if err != nil {
		clientlog.Error(err)
		os.Exit(1)
	}
	sort.Slice(commands, func(i, j int) bool { return commands[i].Name < commands[j].Name })
	app.Commands = commands
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.AppHelpTemplate = getAppHelpTemplate()
	cli.SubcommandHelpTemplate = subcommandHelpTemplate
	app.CommandNotFound = func(c *cli.Context, command string) {
		_, err := fmt.Fprintf(c.App.Writer, "'"+c.App.Name+" "+command+"' is not a jf command. See --help\n")
		if err != nil {
			clientlog.Debug(err)
			os.Exit(1)
		}
		if bestSimilarity := searchSimilarCmds(c.App.Commands, command); len(bestSimilarity) > 0 {
			text := "The most similar "
			if len(bestSimilarity) == 1 {
				text += "command is:\n\tjf " + bestSimilarity[0]
			} else {
				sort.Strings(bestSimilarity)
				text += "commands are:\n\tjf " + strings.Join(bestSimilarity, "\n\tjf ")
			}
			_, err = fmt.Fprintln(c.App.Writer, text)
			if err != nil {
				clientlog.Debug(err)
			}
		}
		os.Exit(1)
	}
	app.Before = func(ctx *cli.Context) error {
		clientlog.Debug("JFrog CLI version:", app.Version)
		clientlog.Debug("OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH)
		warningMessage, err := cliutils.CheckNewCliVersionAvailable(app.Version)
		if err != nil {
			clientlog.Debug("failed while trying to check latest JFrog CLI version:", err.Error())
		}
		if warningMessage != "" {
			clientlog.Warn(warningMessage)
		}
		if err = getGithubActionJobSummaries(ctx); err != nil {
			return err
		}
		return nil
	}
	err = app.Run(args)
	return err
}

func getGithubActionJobSummaries(ctx *cli.Context) error {
	markdownLocation := os.Getenv("GITHUB_STEP_SUMMARY")
	if markdownLocation == "" {
		return nil
	}
	return ctx.GlobalSet("githubJobSummary", "true")

}

// Detects typos and can identify one or more valid commands similar to the error command.
// In Addition, if a subcommand is found with exact match, preferred it over similar commands, for example:
// "jf bp" -> return "jf rt bp"
func searchSimilarCmds(cmds []cli.Command, toCompare string) (bestSimilarity []string) {
	// Set min diff between two commands.
	minDistance := 2
	for _, cmd := range cmds {
		// Check if we have an exact match with the next level.
		for _, subCmd := range cmd.Subcommands {
			for _, subCmdName := range subCmd.Names() {
				// Found exact match, return it.
				distance := levenshtein.ComputeDistance(subCmdName, toCompare)
				if distance == 0 {
					return []string{cmd.Name + " " + subCmdName}
				}
			}
		}
		// Search similar commands with max diff of 'minDistance'.
		for _, cmdName := range cmd.Names() {
			distance := levenshtein.ComputeDistance(cmdName, toCompare)
			if distance == minDistance {
				// In the case of an alias, we don't want to show the full command name, but the alias.
				// Therefore, we trim the end of the full name and concat the actual matched (alias/full command name)
				bestSimilarity = append(bestSimilarity, strings.Replace(cmd.FullName(), cmd.Name, cmdName, 1))
			}
			if distance < minDistance {
				// Found a cmd with a smaller distance.
				minDistance = distance
				bestSimilarity = []string{strings.Replace(cmd.FullName(), cmd.Name, cmdName, 1)}
			}
		}
	}
	return
}

const otherCategory = "Other"

func getCommands() ([]cli.Command, error) {
	cliNameSpaces := []cli.Command{
		{
			Name:        cliutils.CmdArtifactory,
			Usage:       "Artifactory commands.",
			Subcommands: artifactory.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdMissionControl,
			Usage:       "Mission Control commands.",
			Subcommands: missioncontrol.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdDistribution,
			Usage:       "Distribution commands.",
			Subcommands: distribution.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdPipelines,
			Usage:       "JFrog Pipelines commands.",
			Subcommands: pipelines.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdCompletion,
			Usage:       "Generate autocomplete scripts.",
			Subcommands: completion.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdPlugin,
			Usage:       "Plugins handling commands.",
			Subcommands: plugins.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdConfig,
			Aliases:     []string{"c"},
			Usage:       "Config commands.",
			Subcommands: config.GetCommands(),
			Category:    otherCategory,
		},
		{
			Name:        cliutils.CmdProject,
			Usage:       "Project commands.",
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
		// {
		//	Name:         "invite",
		//	Usage:        invite.GetDescription(),
		//	HelpName:     corecommon.CreateUsage("invite", invite.GetDescription(), invite.Usage),
		//	ArgsUsage:    common.CreateEnvVars(),
		//	BashComplete: corecommon.CreateBashCompletionFunc(),
		//	Category:     otherCategory,
		//	Action: func(c *cli.Context) error {
		//		return invitecommand.RunInviteCmd(c)
		//	},
		// },
		{
			Name:     "setup",
			HideHelp: true,
			Hidden:   true,
			Flags:    cliutils.GetCommandFlags(cliutils.Setup),
			Action:   SetupCmd,
		},
		{
			Name:     "intro",
			HideHelp: true,
			Hidden:   true,
			Flags:    cliutils.GetCommandFlags(cliutils.Intro),
			Action:   IntroCmd,
		},
		{
			Name:     cliutils.CmdOptions,
			Usage:    "Show all supported environment variables.",
			Category: otherCategory,
			Action: func(*cli.Context) {
				fmt.Println(common.GetGlobalEnvVars())
			},
		},
		{
			Name:         "login",
			Usage:        loginDocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("login", loginDocs.GetDescription(), loginDocs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     otherCategory,
			Action:       login.LoginCmd,
		},
		{
			Hidden:       true,
			Name:         "how",
			Usage:        aiDocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("how", aiDocs.GetDescription(), aiDocs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     otherCategory,
			Action:       ai.HowCmd,
		},
		{
			Name:         "access-token-create",
			Aliases:      []string{"atc"},
			Flags:        cliutils.GetCommandFlags(cliutils.AccessTokenCreate),
			Usage:        tokenDocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("atc", tokenDocs.GetDescription(), tokenDocs.Usage),
			UsageText:    tokenDocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       token.AccessTokenCreateCmd,
		},
	}

	securityCmds, err := ConvertEmbeddedPlugin(securityCLI.GetJfrogCliSecurityApp())
	if err != nil {
		return nil, err
	}
	allCommands := append(slices.Clone(cliNameSpaces), securityCmds...)
	allCommands = append(allCommands, utils.GetPlugins()...)
	allCommands = append(allCommands, buildtools.GetCommands()...)
	allCommands = append(allCommands, lifecycle.GetCommands()...)
	return append(allCommands, buildtools.GetBuildToolsHelpCommands()...), nil
}

// Embedded plugins are CLI plugins that are embedded in the JFrog CLI and not require any installation.
// This function converts an embedded plugin to a cli.Command slice to be registered as commands of the cli.
func ConvertEmbeddedPlugin(jfrogPlugin components.App) (converted []cli.Command, err error) {
	for i := range jfrogPlugin.Subcommands {
		// commands name-space without category are considered as 'other' category
		if jfrogPlugin.Subcommands[i].Category == "" {
			jfrogPlugin.Subcommands[i].Category = otherCategory
		}
	}
	if converted, err = components.ConvertAppCommands(jfrogPlugin); err != nil {
		err = fmt.Errorf("failed adding '%s' embedded plugin commands. Last error: %s", jfrogPlugin.Name, err.Error())
	}
	return
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

func SetupCmd(c *cli.Context) error {
	format := setupcore.Human
	formatFlag := c.String("format")
	if formatFlag == string(setupcore.Machine) {
		format = setupcore.Machine
	}
	return envsetup.RunEnvSetupCmd(c, format)
}

func IntroCmd(_ *cli.Context) error {
	ci, err := clientutils.GetBoolEnvValue(coreutils.CI, false)
	if ci || err != nil {
		return err
	}
	clientlog.Output()
	clientlog.Output(coreutils.PrintTitle(fmt.Sprintf("Thank you for installing version %s of JFrog CLI! 🐸", cliutils.CliVersion)))
	var serverExists bool
	serverExists, err = coreconfig.IsServerConfExists()
	if serverExists || err != nil {
		return err
	}
	clientlog.Output(coreutils.PrintTitle("So what's next?"))
	clientlog.Output()
	clientlog.Output(coreutils.PrintTitle("Authenticate with your JFrog Platform by running one of the following two commands:"))
	clientlog.Output()
	clientlog.Output("jf login")
	clientlog.Output(coreutils.PrintTitle("or"))
	clientlog.Output("jf c add")
	return nil
}

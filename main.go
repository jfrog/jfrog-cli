package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/gofrog/version"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"net/http"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	setupcore "github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	"github.com/jfrog/jfrog-cli/artifactory"
	"github.com/jfrog/jfrog-cli/buildtools"
	"github.com/jfrog/jfrog-cli/completion"
	"github.com/jfrog/jfrog-cli/config"
	"github.com/jfrog/jfrog-cli/distribution"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/docs/general/cisetup"
	cisetupcommand "github.com/jfrog/jfrog-cli/general/cisetup"
	"github.com/jfrog/jfrog-cli/general/envsetup"
	"github.com/jfrog/jfrog-cli/general/project"
	"github.com/jfrog/jfrog-cli/missioncontrol"
	"github.com/jfrog/jfrog-cli/pipelines"
	"github.com/jfrog/jfrog-cli/plugins"
	"github.com/jfrog/jfrog-cli/plugins/utils"
	"github.com/jfrog/jfrog-cli/scan"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/xray"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
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

type githubResponse struct {
	TagName string `json:"tag_name,omitempty"`
	URL     string `json:"html_url"`
}

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
		warningMessage, err := checkNewCliVersionAvailable(app.Version)
		if err != nil {
			clientlog.Debug("failed while trying to check latest JFrog CLI version:", err.Error())
		}
		if warningMessage != "" {
			clientlog.Warn(warningMessage)
		}
		return nil
	}
	err := app.Run(args)
	return err
}

// Checks if the requested plugin exists in registry and does not exist locally.
func checkNewCliVersionAvailable(currentVersion string) (warningMessage string, err error) {
	shouldCheck, err := shouldCheckLatestCliVersion()
	if err != nil || !shouldCheck {
		return
	}
	githubVersionInfo, err := getLatestCliVersionFromGithubAPI()
	if err != nil {
		return
	}
	latestVersion := strings.TrimPrefix(githubVersionInfo.TagName, "v")
	if version.NewVersion(latestVersion).Compare(currentVersion) < 0 {
		warningMessage = strings.Join([]string{
			fmt.Sprintf("You are using JFrog CLI version %s, however version %s is available.", coreutils.PrintComment(currentVersion), coreutils.PrintTitle(latestVersion)),
			fmt.Sprintf("To install the latest version, visit: %sgetcli", coreutils.JFrogComUrl),
			"To see the release notes, visit: " + githubVersionInfo.URL,
			fmt.Sprintf("To ignore this message you can use %s=TRUE", cliutils.JfrogCliAvoidNewVersionWarning),
		},
			"\n")
	}
	return
}

func shouldCheckLatestCliVersion() (shouldCheck bool, err error) {
	if strings.ToLower(os.Getenv(cliutils.JfrogCliAvoidNewVersionWarning)) == "true" {
		return
	}
	homeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return
	}
	indicatorFile := path.Join(homeDir, "Latest_Cli_Version_Check_Indicator")
	fileInfo, err := os.Stat(indicatorFile)
	if err != nil && !os.IsNotExist(err) {
		err = fmt.Errorf("couldn't get indicator file %s info: %s", indicatorFile, err.Error())
		return
	}
	if err == nil && (time.Now().UnixMilli()-fileInfo.ModTime().UnixMilli()) < cliutils.LatestCliVersionCheckInterval.Milliseconds() {
		// Timestamp file exists and updated less than 3 hours ago, therefor no need to check version again
		return
	}
	return true, os.WriteFile(indicatorFile, []byte{}, 0666)
}

func getLatestCliVersionFromGithubAPI() (githubVersionInfo githubResponse, err error) {
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return
	}
	resp, body, _, err := client.SendGet("https://api.github.com/repos/jfrog/jfrog-cli/releases/latest", true, httputils.HttpClientDetails{HttpTimeout: time.Second * 2}, "")
	if err != nil {
		err = errors.New("couldn't get latest JFrog CLI latest version info from GitHub API: " + err.Error())
		return
	}
	err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &githubVersionInfo)
	return
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

func getCommands() []cli.Command {
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
			Name:        cliutils.CmdXray,
			Usage:       "Xray commands.",
			Subcommands: xray.GetCommands(),
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
		//{
		//	Name:         "invite",
		//	Usage:        invite.GetDescription(),
		//	HelpName:     corecommon.CreateUsage("invite", invite.GetDescription(), invite.Usage),
		//	ArgsUsage:    common.CreateEnvVars(),
		//	BashComplete: corecommon.CreateBashCompletionFunc(),
		//	Category:     otherCategory,
		//	Action: func(c *cli.Context) error {
		//		return invitecommand.RunInviteCmd(c)
		//	},
		//},
		{
			Name:     "setup",
			HideHelp: true,
			Hidden:   true,
			Flags:    cliutils.GetCommandFlags(cliutils.Setup),
			Action: func(c *cli.Context) error {
				return SetupCmd(c)
			},
		},
		{
			Name:     "intro",
			HideHelp: true,
			Hidden:   true,
			Flags:    cliutils.GetCommandFlags(cliutils.Intro),
			Action: func(*cli.Context) error {
				return IntroCmd()
			},
		},
		{
			Name:     cliutils.CmdOptions,
			Usage:    "Show all supported environment variables.",
			Category: otherCategory,
			Action: func(*cli.Context) {
				fmt.Println(common.GetGlobalEnvVars())
			},
		},
	}
	allCommands := append(cliNameSpaces, utils.GetPlugins()...)
	allCommands = append(allCommands, scan.GetCommands()...)
	allCommands = append(allCommands, buildtools.GetCommands()...)
	return append(allCommands, buildtools.GetBuildToolsHelpCommands()...)
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

func IntroCmd() error {
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
	clientlog.Output(coreutils.PrintTitle("Here's how you get started."))
	clientlog.Output("🐸 If you already have a JFrog environment, run the 'jf c add' command to set its connection details.")
	clientlog.Output("🐸 Don't have a JFrog environment? No problem!\n   Simply run the 'jf setup' command.\n   This command will set you up with a free JFrog environment in the cloud, and also configure JFrog CLI to use it, all in less then two minutes.\n")
	return nil
}

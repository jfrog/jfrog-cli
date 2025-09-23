package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"
	gofrogcmd "github.com/jfrog/gofrog/io"
	artifactoryCLI "github.com/jfrog/jfrog-cli-artifactory/cli"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/log"
	platformServicesCLI "github.com/jfrog/jfrog-cli-platform-services/cli"
	securityCLI "github.com/jfrog/jfrog-cli-security/cli"
	"github.com/jfrog/jfrog-cli/artifactory"
	"github.com/jfrog/jfrog-cli/buildtools"
	"github.com/jfrog/jfrog-cli/completion"
	"github.com/jfrog/jfrog-cli/config"
	"github.com/jfrog/jfrog-cli/docs/common"
	aiDocs "github.com/jfrog/jfrog-cli/docs/general/ai"
	loginDocs "github.com/jfrog/jfrog-cli/docs/general/login"
	oidcDocs "github.com/jfrog/jfrog-cli/docs/general/oidc"
	summaryDocs "github.com/jfrog/jfrog-cli/docs/general/summary"
	tokenDocs "github.com/jfrog/jfrog-cli/docs/general/token"
	"github.com/jfrog/jfrog-cli/general/ai"
	"github.com/jfrog/jfrog-cli/general/login"
	"github.com/jfrog/jfrog-cli/general/summary"
	"github.com/jfrog/jfrog-cli/general/token"
	"github.com/jfrog/jfrog-cli/missioncontrol"
	"github.com/jfrog/jfrog-cli/pipelines"
	"github.com/jfrog/jfrog-cli/plugins"
	"github.com/jfrog/jfrog-cli/plugins/utils"
	"github.com/jfrog/jfrog-cli/utils/buildinfo"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
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

const (
	jfrogAppName  = "jf"
	traceIdLogMsg = "Trace ID for JFrog Platform logs:"
)

// Trace ID that is generated for the Uber Trace ID header.
var traceID string

func main() {
	log.SetDefaultLogger()
	err := execMain()
	if cleanupErr := fileutils.CleanOldDirs(); cleanupErr != nil {
		clientlog.Warn("failed while attempting to cleanup old CLI temp directories:", cleanupErr)
	}
	coreutils.ExitOnErr(err)
}

func execMain() error {
	// Set JFrog CLI's user-agent on the jfrog-client-go.
	clientutils.SetUserAgent(coreutils.GetCliUserAgent())

	app := cli.NewApp()
	app.Name = jfrogAppName
	app.Usage = "See https://docs.jfrog-applications.jfrog.io/jfrog-applications/jfrog-cli for full documentation."
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
	app.CommandNotFound = func(c *cli.Context, command string) {
		_, err = fmt.Fprintf(c.App.Writer, "'%s %s' is not a jf command. See --help\n", c.App.Name, command)
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
		if err = setUberTraceIdToken(); err != nil {
			clientlog.Warn("failed generating a trace ID token:", err.Error())
		}
		if os.Getenv("JFROG_RUN_NATIVE") == "true" {
			// If the JFROG_RUN_NATIVE environment variable is set to true, we run the new implementation
			// but only for package manager commands, not for JFrog CLI commands
			args := ctx.Args()
			if args.Present() && len(args) > 0 {
				firstArg := args.Get(0)
				if isPackageManagerCommand(firstArg) {
					if err = runNativeImplementation(ctx); err != nil {
						clientlog.Error("Failed to run native implementation:", err)
						os.Exit(1)
					}
					os.Exit(0)
				}
			}
			// For non-package-manager commands, continue with normal CLI processing
		}
		return nil
	}

	app.CommandNotFound = func(c *cli.Context, command string) {
		// Try to handle as native package manager command only when JFROG_RUN_NATIVE is true
		if os.Getenv("JFROG_RUN_NATIVE") == "true" && isPackageManagerCommand(command) {
			clientlog.Debug("Attempting to handle as native package manager command:", command)
			err := runNativeImplementation(c)
			if err != nil {
				clientlog.Error("Failed to run native implementation:", err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		// Original behavior for unknown commands
		_, err = fmt.Fprintf(c.App.Writer, "'%s %s' is not a jf command. See --help\n", c.App.Name, command)
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

	err = app.Run(args)
	logTraceIdOnFailure(err)

	if err == nil {
		displaySurveyLinkIfNeeded()
	}

	return err
}

// displaySurveyLinkIfNeeded checks if the survey should be hidden based on the JFROG_CLI_HIDE_SURVEY environment variable
func displaySurveyLinkIfNeeded() {
	if cliutils.ShouldHideSurveyLink() {
		return
	}
	fmt.Fprintln(os.Stderr, "\n💬 Help us improve JFrog CLI! \033]8;;https://www.surveymonkey.com/r/JFCLICLI\033\\https://www.surveymonkey.com/r/JFCLICLI\033]8;;\033\\")
}

func runNativeImplementation(ctx *cli.Context) error {
	clientlog.Debug("Starting native implementation...")

	// Extract the build name and number from the command arguments
	args, buildArgs, err := build.ExtractBuildDetailsFromArgs(ctx.Args())
	if err != nil {
		clientlog.Error("Failed to extract build details from args: ", err)
		return fmt.Errorf("ExtractBuildDetailsFromArgs failed: %w", err)
	}

	if len(args) < 2 {
		return fmt.Errorf("insufficient arguments: expected at least package-manager and command, got %v", args)
	}

	packageManager := args[0]
	command := args[1]
	clientlog.Debug("Executing native command: " + packageManager + " " + command)

	buildName, err := buildArgs.GetBuildName()
	if err != nil {
		clientlog.Error("Failed to get build name: ", err)
		return fmt.Errorf("GetBuildName failed: %w", err)
	}

	buildNumber, err := buildArgs.GetBuildNumber()
	if err != nil {
		clientlog.Error("Failed to get build number: ", err)
		return fmt.Errorf("GetBuildNumber failed: %w", err)
	}

	// Execute the native command
	err = RunActions(args)
	if err != nil {
		clientlog.Error("Failed to run actions: ", err)
		return fmt.Errorf("RunActions failed: %w", err)
	}

	// Collect build info if build name and number are provided
	if buildName != "" && buildNumber != "" {
		clientlog.Info("Collecting build info for executed command...")
		workingDir := ctx.GlobalString("working-dir")
		if workingDir == "" {
			workingDir = "."
		}

		// Use the enhanced build info collection that supports Poetry
		err = buildinfo.GetBuildInfoForPackageManager(packageManager, workingDir, buildArgs)
		if err != nil {
			clientlog.Error("Failed to collect build info: ", err)
			return fmt.Errorf("GetBuildInfoForPackageManager failed: %w", err)
		}
	}

	clientlog.Info("Native implementation completed successfully.")
	return nil
}

// isPackageManagerCommand checks if the command is a supported package manager
func isPackageManagerCommand(command string) bool {
	supportedPackageManagers := []string{"poetry", "pip", "pipenv", "gem", "bundle", "npm", "yarn", "gradle", "mvn", "maven", "nuget", "go"}
	for _, pm := range supportedPackageManagers {
		if command == pm {
			return true
		}
	}
	return false
}

// cleanProgressOutput handles carriage returns and progress updates properly
func cleanProgressOutput(output string) string {
	if output == "" {
		return ""
	}

	// First, split by both \n and \r to handle all line breaks
	lines := strings.FieldsFunc(output, func(c rune) bool {
		return c == '\n' || c == '\r'
	})

	var cleanedLines []string
	var progressLines = make(map[string]string) // Track progress lines by filename

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a progress line (contains % and "Uploading")
		if strings.Contains(line, "Uploading") && strings.Contains(line, "%") {
			// Extract filename for tracking progress
			if strings.Contains(line, " - Uploading ") {
				parts := strings.Split(line, " - Uploading ")
				if len(parts) == 2 {
					filename := strings.Split(parts[1], " ")[0]
					progressLines[filename] = line
					continue
				}
			}
		}

		// Add non-progress lines immediately
		cleanedLines = append(cleanedLines, line)
	}

	// Add final progress states
	for _, progressLine := range progressLines {
		cleanedLines = append(cleanedLines, progressLine)
	}

	if len(cleanedLines) > 0 {
		return strings.Join(cleanedLines, "\n") + "\n"
	}
	return ""
}

func RunActions(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("insufficient arguments for RunActions: expected at least 2, got %d", len(args))
	}

	packageManager := args[0]
	subCommand := args[1]
	executableCommand := append([]string{}, args[2:]...)

	clientlog.Debug("Executing command: " + packageManager + " " + subCommand)
	command := gofrogcmd.NewCommand(packageManager, subCommand, executableCommand)

	// Use RunCmdWithOutputParser but handle the output better
	stdout, stderr, exitOk, err := gofrogcmd.RunCmdWithOutputParser(command, false)
	if err != nil {
		clientlog.Error("Command execution failed: ", err)
		if stderr != "" {
			clientlog.Error("Command stderr: ", stderr)
		}
		return fmt.Errorf("command '%s %s' failed (exitOk=%t): %w", packageManager, subCommand, exitOk, err)
	}

	// Print stdout directly without parsing to preserve Poetry's output format
	if stdout != "" {
		fmt.Print(stdout)
	}

	// Also print stderr, but clean it up for progress information
	if stderr != "" {
		cleanStderr := cleanProgressOutput(stderr)
		if cleanStderr != "" {
			fmt.Print(cleanStderr)
		}
	}

	clientlog.Debug("Command executed successfully")
	return nil
}

// This command generates and sets an Uber Trace ID token which will be attached as a header to every request.
// This allows users to easily identify which logs on the server side are related to the command executed by the CLI.
func setUberTraceIdToken() error {
	var err error
	traceID, err = generateTraceIdToken()
	if err != nil {
		return err
	}
	httpclient.SetUberTraceIdToken(traceID)
	clientlog.Debug(traceIdLogMsg, traceID)
	return nil
}

// Generates a 16 chars hexadecimal string to be used as a Trace ID token.
func generateTraceIdToken() (string, error) {
	// Generate 8 random bytes.
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	// Convert the random bytes to a 16 chars hexadecimal string.
	return hex.EncodeToString(buf), nil
}

func logTraceIdOnFailure(err error) {
	if err == nil || traceID == "" {
		return
	}
	clientlog.Info(traceIdLogMsg, traceID)
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
const commandNamespacesCategory = "Command Namespaces"

func getCommands() ([]cli.Command, error) {
	cliNameSpaces := []cli.Command{
		{
			Name:        cliutils.CmdArtifactory,
			Usage:       "Artifactory commands.",
			Subcommands: artifactory.GetCommands(),
			Category:    commandNamespacesCategory,
		},
		{
			Name:        cliutils.CmdMissionControl,
			Usage:       "Mission Control commands.",
			Subcommands: missioncontrol.GetCommands(),
			Category:    commandNamespacesCategory,
		},
		{
			Name:        cliutils.CmdPipelines,
			Usage:       "Pipelines commands.",
			Subcommands: pipelines.GetCommands(),
			Category:    commandNamespacesCategory,
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
			Category:    commandNamespacesCategory,
		},
		{
			Name:        cliutils.CmdConfig,
			Aliases:     []string{"c"},
			Usage:       "Server configurations commands.",
			Subcommands: config.GetCommands(),
			Category:    commandNamespacesCategory,
		},
		{
			Name:   "intro",
			Hidden: true,
			Flags:  cliutils.GetCommandFlags(cliutils.Intro),
			Action: IntroCmd,
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
			Name:         "how",
			Usage:        aiDocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("how", aiDocs.GetDescription(), aiDocs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
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
			Category:     otherCategory,
			Action:       token.AccessTokenCreateCmd,
		},
		{
			Name:         "exchange-oidc-token",
			Aliases:      []string{"eot"},
			Flags:        cliutils.GetCommandFlags(cliutils.ExchangeOidcToken),
			Usage:        oidcDocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("eot", oidcDocs.GetDescription(), oidcDocs.Usage),
			UsageText:    oidcDocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     otherCategory,
			Action:       token.ExchangeOidcTokenCmd,
		},
		{
			Name:     "generate-summary-markdown",
			Aliases:  []string{"gsm"},
			Usage:    summaryDocs.GetDescription(),
			HelpName: corecommon.CreateUsage("gsm", summaryDocs.GetDescription(), summaryDocs.Usage),
			Category: otherCategory,
			Action:   summary.FinalizeCommandSummaries,
		},
	}

	securityCmds, err := ConvertEmbeddedPlugin(securityCLI.GetJfrogCliSecurityApp())
	if err != nil {
		return nil, err
	}
	artifactoryCmds, err := ConvertEmbeddedPlugin(artifactoryCLI.GetJfrogCliArtifactoryApp())
	if err != nil {
		return nil, err
	}
	platformServicesCmds, err := ConvertEmbeddedPlugin(platformServicesCLI.GetPlatformServicesApp())
	if err != nil {
		return nil, err
	}

	allCommands := append(slices.Clone(cliNameSpaces), securityCmds...)
	allCommands = mergeCommands(allCommands, artifactoryCmds)
	allCommands = append(allCommands, platformServicesCmds...)
	allCommands = append(allCommands, utils.GetPlugins()...)
	allCommands = append(allCommands, buildtools.GetCommands()...)
	return append(allCommands, buildtools.GetBuildToolsHelpCommands()...), nil
}

// mergeCommands merges two slices of cli.Command into one, combining subcommands of commands with the same name.
func mergeCommands(a, b []cli.Command) []cli.Command {
	cmdMap := make(map[string]*cli.Command)

	for _, cmd := range append(a, b...) {
		if existing, found := cmdMap[cmd.Name]; found {
			existing.Subcommands = append(existing.Subcommands, cmd.Subcommands...)
		} else {
			cmdMap[cmd.Name] = &cmd
		}
	}

	merged := make([]cli.Command, 0, len(cmdMap))
	for _, cmd := range cmdMap {
		merged = append(merged, *cmd)
	}
	return merged
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

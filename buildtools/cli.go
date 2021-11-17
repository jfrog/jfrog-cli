package buildtools

import (
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/golang"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/pip"
	commandsutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/yarn"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npminstall"
	dotnetdocs "github.com/jfrog/jfrog-cli/docs/buildtools/dotnet"
	"github.com/jfrog/jfrog-cli/docs/buildtools/dotnetconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/gocommand"
	"github.com/jfrog/jfrog-cli/docs/buildtools/goconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/gopublish"
	gradledoc "github.com/jfrog/jfrog-cli/docs/buildtools/gradle"
	"github.com/jfrog/jfrog-cli/docs/buildtools/gradleconfig"
	mvndoc "github.com/jfrog/jfrog-cli/docs/buildtools/mvn"
	"github.com/jfrog/jfrog-cli/docs/buildtools/mvnconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/npmcommand"
	"github.com/jfrog/jfrog-cli/docs/buildtools/npmconfig"
	nugetdocs "github.com/jfrog/jfrog-cli/docs/buildtools/nuget"
	"github.com/jfrog/jfrog-cli/docs/buildtools/nugetconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipinstall"
	yarndocs "github.com/jfrog/jfrog-cli/docs/buildtools/yarn"
	"github.com/jfrog/jfrog-cli/docs/buildtools/yarnconfig"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"strconv"
	"strings"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "mvn-config",
			Aliases:      []string{"mvnc"},
			Flags:        cliutils.GetCommandFlags(cliutils.MvnConfig),
			Description:  mvnconfig.Description,
			HelpName:     corecommon.CreateUsage("mvn-config", mvnconfig.Description, mvnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, utils.Maven)
			},
		},
		{
			Name:            "mvn",
			Flags:           cliutils.GetCommandFlags(cliutils.Mvn),
			Description:     mvndoc.Description,
			HelpName:        corecommon.CreateUsage("mvn", mvndoc.Description, mvndoc.Usage),
			UsageText:       mvndoc.Arguments,
			ArgsUsage:       common.CreateEnvVars(mvndoc.EnvVar),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return MvnCmd(c)
			},
		},
		{
			Name:         "gradle-config",
			Aliases:      []string{"gradlec"},
			Flags:        cliutils.GetCommandFlags(cliutils.GradleConfig),
			Description:  gradleconfig.Description,
			HelpName:     corecommon.CreateUsage("gradle-config", gradleconfig.Description, gradleconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, utils.Gradle)
			},
		},
		{
			Name:            "gradle",
			Flags:           cliutils.GetCommandFlags(cliutils.Gradle),
			Description:     gradledoc.Description,
			HelpName:        corecommon.CreateUsage("gradle", gradledoc.Description, gradledoc.Usage),
			UsageText:       gradledoc.Arguments,
			ArgsUsage:       common.CreateEnvVars(gradledoc.EnvVar),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return GradleCmd(c)
			},
		},
		{
			Name:         "yarn-config",
			Aliases:      []string{"yarnc"},
			Flags:        cliutils.GetCommandFlags(cliutils.YarnConfig),
			Description:  yarnconfig.Description,
			HelpName:     corecommon.CreateUsage("yarn-config", yarnconfig.Description, yarnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, utils.Yarn)
			},
		},
		{
			Name:            "yarn",
			Flags:           cliutils.GetCommandFlags(cliutils.Yarn),
			Description:     yarndocs.Description,
			HelpName:        corecommon.CreateUsage("yarn", yarndocs.Description, yarndocs.Usage),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return YarnCmd(c)
			},
		},
		{
			Name:         "nuget-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NugetConfig),
			Aliases:      []string{"nugetc"},
			Description:  nugetconfig.Description,
			HelpName:     corecommon.CreateUsage("nuget-config", nugetconfig.Description, nugetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, utils.Nuget)
			},
		},
		{
			Name:            "nuget",
			Flags:           cliutils.GetCommandFlags(cliutils.Nuget),
			Description:     nugetdocs.Description,
			HelpName:        corecommon.CreateUsage("nuget", nugetdocs.Description, nugetdocs.Usage),
			UsageText:       nugetdocs.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return NugetCmd(c)
			},
		},
		{
			Name:         "dotnet-config",
			Flags:        cliutils.GetCommandFlags(cliutils.DotnetConfig),
			Aliases:      []string{"dotnetc"},
			Description:  dotnetconfig.Description,
			HelpName:     corecommon.CreateUsage("dotnet-config", dotnetconfig.Description, dotnetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, utils.Dotnet)
			},
		},
		{
			Name:            "dotnet",
			Flags:           cliutils.GetCommandFlags(cliutils.Dotnet),
			Description:     dotnetdocs.Description,
			HelpName:        corecommon.CreateUsage("dotnet", dotnetdocs.Description, dotnetdocs.Usage),
			UsageText:       dotnetdocs.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return DotnetCmd(c)
			},
		},
		{
			Name:         "go-config",
			Aliases:      []string{"goc"},
			Flags:        cliutils.GetCommandFlags(cliutils.GoConfig),
			Description:  goconfig.Description,
			HelpName:     corecommon.CreateUsage("go-config", goconfig.Description, goconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, utils.Go)
			},
		},
		{
			Name:            "go",
			Flags:           cliutils.GetCommandFlags(cliutils.Go),
			Aliases:         []string{"go"},
			Description:     gocommand.Description,
			HelpName:        corecommon.CreateUsage("go", gocommand.Description, gocommand.Usage),
			UsageText:       gocommand.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return GoCmd(c)
			},
		},
		{
			Name:         "go-publish",
			Flags:        cliutils.GetCommandFlags(cliutils.GoPublish),
			Aliases:      []string{"gp"},
			Description:  gopublish.Description,
			HelpName:     corecommon.CreateUsage("go-publish", gopublish.Description, gopublish.Usage),
			UsageText:    gopublish.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return GoPublishCmd(c)
			},
		},
		{
			Name:         "pip-config",
			Flags:        cliutils.GetCommandFlags(cliutils.PipConfig),
			Aliases:      []string{"pipc"},
			Description:  pipconfig.Description,
			HelpName:     corecommon.CreateUsage("pip-config", pipconfig.Description, pipconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, utils.Pip)
			},
		},
		{
			Name:            "pip",
			Flags:           cliutils.GetCommandFlags(cliutils.PipInstall),
			Description:     pipinstall.Description,
			HelpName:        corecommon.CreateUsage("pip", pipinstall.Description, pipinstall.Usage),
			UsageText:       pipinstall.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return pipCmd(c)
			},
		},
		{
			Name:         "npm-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NpmConfig),
			Aliases:      []string{"npmc"},
			Description:  npmconfig.Description,
			HelpName:     corecommon.CreateUsage("npm-config", npmconfig.Description, npmconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, utils.Npm)
			},
		},
		{
			Name:            "npm",
			Flags:           cliutils.GetCommandFlags(cliutils.Npm),
			Description:     npmcommand.Description,
			HelpName:        corecommon.CreateUsage("npm", npminstall.Description, npminstall.Usage),
			UsageText:       npminstall.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        "BuildTools Commands",
			Action: func(c *cli.Context) error {
				return npmCmd(c)
			},
		},
	})
}

func MvnCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Maven)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("No config file was found! Before running the mvn command on a project for the first time, the project should be configured with the mvn-config command. ")
	}
	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	args := cliutils.ExtractCommand(c)
	filteredMavenArgs, insecureTls, err := coreutils.ExtractInsecureTlsFromArgs(args)
	if err != nil {
		return err
	}
	filteredMavenArgs, buildConfiguration, err := utils.ExtractBuildDetailsFromArgs(filteredMavenArgs)
	if err != nil {
		return err
	}
	filteredMavenArgs, threads, err := extractThreadsFlag(filteredMavenArgs)
	if err != nil {
		return err
	}
	filteredMavenArgs, detailedSummary, err := coreutils.ExtractDetailedSummaryFromArgs(filteredMavenArgs)
	if err != nil {
		return err
	}
	filteredMavenArgs, xrayScan, err := coreutils.ExtractXrayScanFromArgs(filteredMavenArgs)
	if err != nil {
		return err
	}
	filteredMavenArgs, format, err := coreutils.ExtractXrayOutputFormatFromArgs(filteredMavenArgs)
	if err != nil {
		return err
	}
	if !xrayScan && format != "" {
		return cliutils.PrintHelpAndReturnError("The --format option can be sent only with the --scan option", c)
	}
	scanOutputFormat, err := commandsutils.GetXrayOutputFormat(format)
	if err != nil {
		return err
	}
	mvnCmd := mvn.NewMvnCommand().SetConfiguration(buildConfiguration).SetConfigPath(configFilePath).SetGoals(filteredMavenArgs).SetThreads(threads).SetInsecureTls(insecureTls).SetDetailedSummary(detailedSummary).SetXrayScan(xrayScan).SetScanOutputFormat(scanOutputFormat)
	err = commands.Exec(mvnCmd)
	if err != nil {
		return err
	}
	if mvnCmd.IsDetailedSummary() {
		return printDetailedSummaryReportMvnGradle(err, mvnCmd.Result())
	}
	return nil
}

func GradleCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Gradle)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("No config file was found! Before running the gradle command on a project for the first time, the project should be configured with the gradle-config command. ")
	}
	// Found a config file. Continue as native command.
	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	args := cliutils.ExtractCommand(c)
	filteredGradleArgs, buildConfiguration, err := utils.ExtractBuildDetailsFromArgs(args)
	if err != nil {
		return err
	}
	filteredGradleArgs, threads, err := extractThreadsFlag(filteredGradleArgs)
	if err != nil {
		return err
	}
	filteredGradleArgs, detailedSummary, err := coreutils.ExtractDetailedSummaryFromArgs(filteredGradleArgs)
	if err != nil {
		return err
	}
	filteredGradleArgs, xrayScan, err := coreutils.ExtractXrayScanFromArgs(filteredGradleArgs)
	if err != nil {
		return err
	}
	filteredGradleArgs, format, err := coreutils.ExtractXrayOutputFormatFromArgs(filteredGradleArgs)
	if err != nil {
		return err
	}
	if !xrayScan && format != "" {
		return cliutils.PrintHelpAndReturnError("The --format option can be sent only with the --scan option", c)
	}
	scanOutputFormat, err := commandsutils.GetXrayOutputFormat(format)
	if err != nil {
		return err
	}
	gradleCmd := gradle.NewGradleCommand().SetConfiguration(buildConfiguration).SetTasks(strings.Join(filteredGradleArgs, " ")).SetConfigPath(configFilePath).SetThreads(threads).SetDetailedSummary(detailedSummary).SetXrayScan(xrayScan).SetScanOutputFormat(scanOutputFormat)
	err = commands.Exec(gradleCmd)
	if err != nil {
		return err
	}
	if gradleCmd.IsDetailedSummary() {
		return printDetailedSummaryReportMvnGradle(err, gradleCmd.Result())
	}
	return nil
}

func YarnCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Yarn)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(fmt.Sprintf("No config file was found! Before running the yarn command on a project for the first time, the project should be configured using the yarn-config command."))
	}

	yarnCmd := yarn.NewYarnCommand().SetConfigFilePath(configFilePath).SetArgs(c.Args())
	return commands.Exec(yarnCmd)
}

func NugetCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Nuget)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New(fmt.Sprintf("No config file was found! Before running the nuget command on a project for the first time, the project should be configured using the nuget-config command."))
	}

	rtDetails, targetRepo, useNugetV2, err := getNugetAndDotnetConfigFields(configFilePath)
	if err != nil {
		return err
	}
	args := cliutils.ExtractCommand(c)
	filteredNugetArgs, buildConfiguration, err := utils.ExtractBuildDetailsFromArgs(args)
	if err != nil {
		return err
	}

	nugetCmd := dotnet.NewNugetCommand()
	nugetCmd.SetServerDetails(rtDetails).SetRepoName(targetRepo).SetBuildConfiguration(buildConfiguration).
		SetBasicCommand(filteredNugetArgs[0]).SetUseNugetV2(useNugetV2)
	// Since we are using the values of the command's arguments and flags along the buildInfo collection process,
	// we want to separate the actual NuGet basic command (restore/build...) from the arguments and flags
	if len(filteredNugetArgs) > 1 {
		nugetCmd.SetArgAndFlags(filteredNugetArgs[1:])
	}
	return commands.Exec(nugetCmd)
}

func DotnetCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	// Get configuration file path.
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Dotnet)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(fmt.Sprintf("No config file was found! Before running the dotnet command on a project for the first time, the project should be configured using the dotnet-config command."))
	}

	rtDetails, targetRepo, useNugetV2, err := getNugetAndDotnetConfigFields(configFilePath)
	if err != nil {
		return err
	}

	args := cliutils.ExtractCommand(c)

	filteredDotnetArgs, buildConfiguration, err := utils.ExtractBuildDetailsFromArgs(args)
	if err != nil {
		return err
	}

	// Run command.
	dotnetCmd := dotnet.NewDotnetCoreCliCommand()
	dotnetCmd.SetServerDetails(rtDetails).SetRepoName(targetRepo).SetBuildConfiguration(buildConfiguration).
		SetBasicCommand(filteredDotnetArgs[0]).SetUseNugetV2(useNugetV2)
	// Since we are using the values of the command's arguments and flags along the buildInfo collection process,
	// we want to separate the actual .NET basic command (restore/build...) from the arguments and flags
	if len(filteredDotnetArgs) > 1 {
		dotnetCmd.SetArgAndFlags(filteredDotnetArgs[1:])
	}
	return commands.Exec(dotnetCmd)
}

func getNugetAndDotnetConfigFields(configFilePath string) (rtDetails *coreConfig.ServerDetails, targetRepo string, useNugetV2 bool, err error) {
	vConfig, err := utils.ReadConfigFile(configFilePath, utils.YAML)
	if err != nil {
		return nil, "", false, errors.New(fmt.Sprintf("Error occurred while attempting to read nuget-configuration file: %s", err.Error()))
	}
	projectConfig, err := utils.GetRepoConfigByPrefix(configFilePath, utils.ProjectConfigResolverPrefix, vConfig)
	if err != nil {
		return nil, "", false, err
	}
	rtDetails, err = projectConfig.ServerDetails()
	if err != nil {
		return nil, "", false, err
	}
	targetRepo = projectConfig.TargetRepo()
	useNugetV2 = vConfig.GetBool(utils.ProjectConfigResolverPrefix + "." + "nugetV2")
	return
}

func extractThreadsFlag(args []string) (cleanArgs []string, threadsCount int, err error) {
	// Extract threads flag.
	cleanArgs = append([]string(nil), args...)
	threadsFlagIndex, threadsValueIndex, threads, err := coreutils.FindFlag("--threads", cleanArgs)
	if err != nil || threadsFlagIndex < 0 {
		threadsCount = cliutils.Threads
		return
	}
	coreutils.RemoveFlagFromCommand(&cleanArgs, threadsFlagIndex, threadsValueIndex)

	// Convert flag value to int.
	threadsCount, err = strconv.Atoi(threads)
	if err != nil {
		err = errors.New("The '--threads' option should have a numeric value. " + cliutils.GetDocumentationMessage())
	}
	return
}

func GoCmd(c *cli.Context) error {
	configFilePath, err := goCmdVerification(c)
	if err != nil {
		return err
	}
	args := cliutils.ExtractCommand(c)
	goCommand := golang.NewGoCommand()
	goCommand.SetConfigFilePath(configFilePath).SetGoArg(args)
	return commands.Exec(goCommand)
}

func GoPublishCmd(c *cli.Context) error {
	configFilePath, err := goCmdVerification(c)
	if err != nil {
		return err
	}
	buildConfiguration, err := CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	version := c.Args().Get(0)
	goPublishCmd := golang.NewGoPublishCommand()
	goPublishCmd.SetConfigFilePath(configFilePath).SetBuildConfiguration(buildConfiguration).SetVersion(version).SetDetailedSummary(c.Bool("detailed-summary"))
	err = commands.Exec(goPublishCmd)
	result := goPublishCmd.Result()
	return cliutils.PrintDetailedSummaryReport(result.SuccessCount(), result.FailCount(), result.Reader(), true, false, err)
}

func goCmdVerification(c *cli.Context) (string, error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return "", err
	}
	if c.NArg() < 1 {
		return "", cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Go)
	if err != nil {
		return "", err
	}
	// Verify config file is found.
	if !exists {
		return "", errors.New(fmt.Sprintf("No config file was found! Before running the go command on a project for the first time, the project should be configured using the go-config command."))
	}
	log.Debug("Go config file was found in:", configFilePath)
	return configFilePath, nil
}

func printDetailedSummaryReportMvnGradle(originalErr error, result *commandsutils.Result) (err error) {
	if len(result.Reader().GetFilesPaths()) == 0 {
		return errorutils.CheckError(errors.New("empty reader - no files paths"))
	}
	defer func() {
		e := os.Remove(result.Reader().GetFilesPaths()[0])
		if err == nil {
			err = e
		}
	}()
	err = cliutils.PrintDetailedSummaryReport(result.SuccessCount(), result.FailCount(), result.Reader(), true, false, originalErr)
	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), false)
}

func CreateBuildConfigurationWithModule(c *cli.Context) (buildConfigConfiguration *utils.BuildConfiguration, err error) {
	buildConfigConfiguration = new(utils.BuildConfiguration)
	buildConfigConfiguration.BuildName, buildConfigConfiguration.BuildNumber = utils.GetBuildNameAndNumber(c.String("build-name"), c.String("build-number"))
	buildConfigConfiguration.Project = utils.GetBuildProject(c.String("project"))
	buildConfigConfiguration.Module = c.String("module")
	err = utils.ValidateBuildAndModuleParams(buildConfigConfiguration)
	return
}

func npmCmd(c *cli.Context) error {
	configFilePath, orgArgs, err := GetNpmConfigAndArgs(c)
	if err != nil {
		return err
	}
	cmdName, filteredArgs := getCommandName(orgArgs)
	switch cmdName {
	// Aliases accepted by npm.
	case "install", "i", "isntall", "add":
		return npmInstallCmd(configFilePath, filteredArgs)
	case "ci":
		return npmCiCmd(configFilePath, filteredArgs)
	case "publish", "p":
		return npmPublishCmd(configFilePath, filteredArgs)
	default:
		return npmNativeCmd(cmdName, configFilePath, orgArgs)
	}
}

// Assuming command name is the first argument that isn't a flag.
// Returns the command name, and the filtered arguments slice without it.
func getCommandName(orgArgs []string) (string, []string) {
	cmdArgs := make([]string, len(orgArgs))
	copy(cmdArgs, orgArgs)
	for i, arg := range cmdArgs {
		if !strings.HasPrefix(arg, "-") {
			return arg, append(cmdArgs[:i], cmdArgs[i+1:]...)
		}
	}
	return "", cmdArgs
}

func npmInstallCmd(configFilePath string, args []string) error {
	npmCmd := npm.NewNpmInstallCommand()
	npmCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	err := npmCmd.Init()
	if err != nil {
		return err
	}
	return commands.Exec(npmCmd)
}

func npmCiCmd(configFilePath string, args []string) error {
	npmCmd := npm.NewNpmCiCommand()
	npmCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	err := npmCmd.Init()
	if err != nil {
		return err
	}
	return commands.Exec(npmCmd)
}

func npmPublishCmd(configFilePath string, args []string) error {
	npmCmd := npm.NewNpmPublishCommand()
	npmCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	err := npmCmd.Init()
	if err != nil {
		return err
	}
	return commands.Exec(npmCmd)
}

func npmNativeCmd(cmdName, configFilePath string, fullCmd []string) error {
	npmCmd := npm.NewNpmNativeCommand(cmdName)
	npmCmd.SetConfigFilePath(configFilePath).SetNpmArgs(fullCmd)
	err := npmCmd.Init()
	if err != nil {
		return err
	}
	return commands.Exec(npmCmd)
}

func GetNpmConfigAndArgs(c *cli.Context) (configFilePath string, args []string, err error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return "", nil, err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Npm)
	if err != nil {
		return "", nil, err
	}

	if !exists {
		return "", nil, errors.New("No config file was found! Before running the npm command on a project for the first time, the project should be configured using the npm-config command. ")
	}
	args = cliutils.ExtractCommand(c)
	return
}

func pipCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	// Get pip configuration.
	pipConfig, err := utils.GetResolutionOnlyConfiguration(utils.Pip)
	if err != nil {
		return errors.New(fmt.Sprintf("Error occurred while attempting to read pip-configuration file: %s\n"+
			"Please run 'jfrog rt pip-config' command prior to running 'jfrog %s'.", err.Error(), "pip"))
	}

	// Set arg values.
	rtDetails, err := pipConfig.ServerDetails()
	if err != nil {
		return err
	}

	orgArgs := cliutils.ExtractCommand(c)

	cmdName, filteredArgs := getCommandName(orgArgs)
	if cmdName == "install" {
		return pipInstallCmd(rtDetails, pipConfig, filteredArgs)
	}
	return pipNativeCmd(cmdName, rtDetails, pipConfig, filteredArgs)
}

func pipInstallCmd(rtDetails *coreConfig.ServerDetails, pipConfig *utils.RepositoryConfig, args []string) error {
	pipCmd := pip.NewPipInstallCommand()
	pipCmd.SetServerDetails(rtDetails).SetRepo(pipConfig.TargetRepo()).SetArgs(args)
	return commands.Exec(pipCmd)
}

func pipNativeCmd(cmdName string, rtDetails *coreConfig.ServerDetails, pipConfig *utils.RepositoryConfig, args []string) error {
	pipCmd := pip.NewPipNativeCommand(cmdName)
	pipCmd.SetServerDetails(rtDetails).SetRepo(pipConfig.TargetRepo()).SetArgs(args)
	return commands.Exec(pipCmd)
}

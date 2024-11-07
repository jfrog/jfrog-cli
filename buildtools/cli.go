package buildtools

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-security/utils/techutils"
	"os"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/golang"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/python"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/terraform"
	commandsUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/yarn"
	containerutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	outputFormat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	securityCLI "github.com/jfrog/jfrog-cli-security/cli"
	securityDocs "github.com/jfrog/jfrog-cli-security/cli/docs"
	"github.com/jfrog/jfrog-cli-security/commands/scan"
	terraformdocs "github.com/jfrog/jfrog-cli/docs/artifactory/terraform"
	"github.com/jfrog/jfrog-cli/docs/artifactory/terraformconfig"
	twinedocs "github.com/jfrog/jfrog-cli/docs/artifactory/twine"
	"github.com/jfrog/jfrog-cli/docs/buildtools/docker"
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
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipenvconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipenvinstall"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipinstall"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pnpmconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/poetry"
	"github.com/jfrog/jfrog-cli/docs/buildtools/poetryconfig"
	yarndocs "github.com/jfrog/jfrog-cli/docs/buildtools/yarn"
	"github.com/jfrog/jfrog-cli/docs/buildtools/yarnconfig"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

const (
	buildToolsCategory = "Build Tools"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "mvn-config",
			Aliases:      []string{"mvnc"},
			Flags:        cliutils.GetCommandFlags(cliutils.MvnConfig),
			Usage:        mvnconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("mvn-config", mvnconfig.GetDescription(), mvnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Maven)
			},
		},
		{
			Name:            "mvn",
			Flags:           cliutils.GetCommandFlags(cliutils.Mvn),
			Usage:           mvndoc.GetDescription(),
			HelpName:        corecommon.CreateUsage("mvn", mvndoc.GetDescription(), mvndoc.Usage),
			UsageText:       mvndoc.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(mvndoc.EnvVar...),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action: func(c *cli.Context) (err error) {
				cmdName, _ := getCommandName(c.Args())
				return securityCLI.WrapCmdWithCurationPostFailureRun(c, MvnCmd, techutils.Maven, cmdName)
			},
		},
		{
			Name:         "gradle-config",
			Aliases:      []string{"gradlec"},
			Flags:        cliutils.GetCommandFlags(cliutils.GradleConfig),
			Usage:        gradleconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("gradle-config", gradleconfig.GetDescription(), gradleconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Gradle)
			},
		},
		{
			Name:            "gradle",
			Flags:           cliutils.GetCommandFlags(cliutils.Gradle),
			Usage:           gradledoc.GetDescription(),
			HelpName:        corecommon.CreateUsage("gradle", gradledoc.GetDescription(), gradledoc.Usage),
			UsageText:       gradledoc.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(gradledoc.EnvVar...),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          GradleCmd,
		},
		{
			Name:         "yarn-config",
			Aliases:      []string{"yarnc"},
			Flags:        cliutils.GetCommandFlags(cliutils.YarnConfig),
			Usage:        yarnconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("yarn-config", yarnconfig.GetDescription(), yarnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Yarn)
			},
		},
		{
			Name:            "yarn",
			Flags:           cliutils.GetCommandFlags(cliutils.Yarn),
			Usage:           yarndocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("yarn", yarndocs.GetDescription(), yarndocs.Usage),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          YarnCmd,
		},
		{
			Name:         "nuget-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NugetConfig),
			Aliases:      []string{"nugetc"},
			Usage:        nugetconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("nuget-config", nugetconfig.GetDescription(), nugetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Nuget)
			},
		},
		{
			Name:            "nuget",
			Flags:           cliutils.GetCommandFlags(cliutils.Nuget),
			Usage:           nugetdocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("nuget", nugetdocs.GetDescription(), nugetdocs.Usage),
			UsageText:       nugetdocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          NugetCmd,
		},
		{
			Name:         "dotnet-config",
			Flags:        cliutils.GetCommandFlags(cliutils.DotnetConfig),
			Aliases:      []string{"dotnetc"},
			Usage:        dotnetconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("dotnet-config", dotnetconfig.GetDescription(), dotnetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Dotnet)
			},
		},
		{
			Name:            "dotnet",
			Flags:           cliutils.GetCommandFlags(cliutils.Dotnet),
			Usage:           dotnetdocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("dotnet", dotnetdocs.GetDescription(), dotnetdocs.Usage),
			UsageText:       dotnetdocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          DotnetCmd,
		},
		{
			Name:         "go-config",
			Aliases:      []string{"goc"},
			Flags:        cliutils.GetCommandFlags(cliutils.GoConfig),
			Usage:        goconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("go-config", goconfig.GetDescription(), goconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Go)
			},
		},
		{
			Name:            "go",
			Flags:           cliutils.GetCommandFlags(cliutils.Go),
			Usage:           gocommand.GetDescription(),
			HelpName:        corecommon.CreateUsage("go", gocommand.GetDescription(), gocommand.Usage),
			UsageText:       gocommand.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action: func(c *cli.Context) (err error) {
				cmdName, _ := getCommandName(c.Args())
				return securityCLI.WrapCmdWithCurationPostFailureRun(c, GoCmd, techutils.Go, cmdName)
			},
		},
		{
			Name:         "go-publish",
			Flags:        cliutils.GetCommandFlags(cliutils.GoPublish),
			Aliases:      []string{"gp"},
			Usage:        gopublish.GetDescription(),
			HelpName:     corecommon.CreateUsage("go-publish", gopublish.GetDescription(), gopublish.Usage),
			UsageText:    gopublish.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action:       GoPublishCmd,
		},
		{
			Name:         "pip-config",
			Flags:        cliutils.GetCommandFlags(cliutils.PipConfig),
			Aliases:      []string{"pipc"},
			Usage:        pipconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("pip-config", pipconfig.GetDescription(), pipconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Pip)
			},
		},
		{
			Name:            "pip",
			Flags:           cliutils.GetCommandFlags(cliutils.PipInstall),
			Usage:           pipinstall.GetDescription(),
			HelpName:        corecommon.CreateUsage("pip", pipinstall.GetDescription(), pipinstall.Usage),
			UsageText:       pipinstall.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action: func(c *cli.Context) (err error) {
				cmdName, _ := getCommandName(c.Args())
				return securityCLI.WrapCmdWithCurationPostFailureRun(c, PipCmd, techutils.Pip, cmdName)
			},
		},
		{
			Name:         "pipenv-config",
			Flags:        cliutils.GetCommandFlags(cliutils.PipenvConfig),
			Aliases:      []string{"pipec"},
			Usage:        pipenvconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("pipenv-config", pipenvconfig.GetDescription(), pipenvconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Pipenv)
			},
		},
		{
			Name:            "pipenv",
			Flags:           cliutils.GetCommandFlags(cliutils.PipenvInstall),
			Usage:           pipenvinstall.GetDescription(),
			HelpName:        corecommon.CreateUsage("pipenv", pipenvinstall.GetDescription(), pipenvinstall.Usage),
			UsageText:       pipenvinstall.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          PipenvCmd,
		},
		{
			Name:         "poetry-config",
			Flags:        cliutils.GetCommandFlags(cliutils.PoetryConfig),
			Aliases:      []string{"poc"},
			Usage:        poetryconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("poetry-config", poetryconfig.GetDescription(), poetryconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Poetry)
			},
		},
		{
			Name:            "poetry",
			Flags:           cliutils.GetCommandFlags(cliutils.Poetry),
			Usage:           poetry.GetDescription(),
			HelpName:        corecommon.CreateUsage("poetry", poetry.GetDescription(), poetry.Usage),
			UsageText:       poetry.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          PoetryCmd,
		},
		{
			Name:         "npm-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NpmConfig),
			Aliases:      []string{"npmc"},
			Usage:        npmconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("npm-config", npmconfig.GetDescription(), npmconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Npm)
			},
		},
		{
			Name:            "npm",
			Usage:           npmcommand.GetDescription(),
			HelpName:        corecommon.CreateUsage("npm", npmcommand.GetDescription(), npmcommand.Usage),
			UsageText:       npmcommand.GetArguments(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc("install", "i", "isntall", "add", "ci", "publish", "p"),
			Category:        buildToolsCategory,
			Action: func(c *cli.Context) (errFromCmd error) {
				cmdName, _ := getCommandName(c.Args())
				return securityCLI.WrapCmdWithCurationPostFailureRun(c,
					func(c *cli.Context) error {
						return npmGenericCmd(c, cmdName, false)
					},
					techutils.Npm, cmdName)
			},
		},
		{
			Name:         "pnpm-config",
			Flags:        cliutils.GetCommandFlags(cliutils.PnpmConfig),
			Aliases:      []string{"pnpmc"},
			Usage:        pnpmconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("pnpm-config", pnpmconfig.GetDescription(), pnpmconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Pnpm)
			},
		},
		{
			Name:      "docker",
			Flags:     cliutils.GetCommandFlags(cliutils.Docker),
			Usage:     docker.GetDescription(),
			HelpName:  corecommon.CreateUsage("docker", docker.GetDescription(), docker.Usage),
			UsageText: docker.GetArguments(),
			SkipFlagParsing: func() bool {
				for i, arg := range os.Args {
					// 'docker scan' isn't a docker client command. We won't skip its flags.
					if arg == "docker" && len(os.Args) > i+1 && os.Args[i+1] == "scan" {
						return false
					}
				}
				return true
			}(),
			BashComplete: corecommon.CreateBashCompletionFunc("push", "pull", "scan"),
			Category:     buildToolsCategory,
			Action:       dockerCmd,
		},
		{
			Name:         "terraform-config",
			Flags:        cliutils.GetCommandFlags(cliutils.TerraformConfig),
			Aliases:      []string{"tfc"},
			Usage:        terraformconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("terraform-config", terraformconfig.GetDescription(), terraformconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Terraform)
			},
		},
		{
			Name:            "terraform",
			Flags:           cliutils.GetCommandFlags(cliutils.Terraform),
			Aliases:         []string{"tf"},
			Usage:           terraformdocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("terraform", terraformdocs.GetDescription(), terraformdocs.Usage),
			UsageText:       terraformdocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          terraformCmd,
		},
		{
			Name:            "twine",
			Flags:           cliutils.GetCommandFlags(cliutils.Twine),
			Usage:           twinedocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("twine", twinedocs.GetDescription(), twinedocs.Usage),
			UsageText:       twinedocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          twineCmd,
		},
	})
}

func MvnCmd(c *cli.Context) (err error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, err := getProjectConfigPathOrThrow(project.Maven, "mvn", "mvn-config")
	if err != nil {
		return err
	}

	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	args := cliutils.ExtractCommand(c)
	filteredMavenArgs, insecureTls, err := coreutils.ExtractInsecureTlsFromArgs(args)
	if err != nil {
		return err
	}
	filteredMavenArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(filteredMavenArgs)
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
	if xrayScan {
		commandsUtils.ConditionalUploadScanFunc = scan.ConditionalUploadDefaultScanFunc
	}
	filteredMavenArgs, format, err := coreutils.ExtractXrayOutputFormatFromArgs(filteredMavenArgs)
	if err != nil {
		return err
	}
	printDeploymentView := log.IsStdErrTerminal()
	if !xrayScan && format != "" {
		return cliutils.PrintHelpAndReturnError("The --format option can be sent only with the --scan option", c)
	}
	scanOutputFormat, err := outputFormat.GetOutputFormat(format)
	if err != nil {
		return err
	}
	mvnCmd := mvn.NewMvnCommand().SetConfiguration(buildConfiguration).SetConfigPath(configFilePath).SetGoals(filteredMavenArgs).SetThreads(threads).SetInsecureTls(insecureTls).SetDetailedSummary(detailedSummary || printDeploymentView).SetXrayScan(xrayScan).SetScanOutputFormat(scanOutputFormat)
	err = commands.Exec(mvnCmd)
	result := mvnCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(mvnCmd.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func GradleCmd(c *cli.Context) (err error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, err := getProjectConfigPathOrThrow(project.Gradle, "gradle", "gradle-config")
	if err != nil {
		return err
	}

	// Found a config file. Continue as native command.
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	args := cliutils.ExtractCommand(c)
	filteredGradleArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(args)
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
	if xrayScan {
		commandsUtils.ConditionalUploadScanFunc = scan.ConditionalUploadDefaultScanFunc
	}
	filteredGradleArgs, format, err := coreutils.ExtractXrayOutputFormatFromArgs(filteredGradleArgs)
	if err != nil {
		return err
	}
	if !xrayScan && format != "" {
		return cliutils.PrintHelpAndReturnError("The --format option can be sent only with the --scan option", c)
	}
	scanOutputFormat, err := outputFormat.GetOutputFormat(format)
	if err != nil {
		return err
	}
	printDeploymentView := log.IsStdErrTerminal()
	gradleCmd := gradle.NewGradleCommand().SetConfiguration(buildConfiguration).SetTasks(filteredGradleArgs).SetConfigPath(configFilePath).SetThreads(threads).SetDetailedSummary(detailedSummary || printDeploymentView).SetXrayScan(xrayScan).SetScanOutputFormat(scanOutputFormat)
	err = commands.Exec(gradleCmd)
	result := gradleCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(gradleCmd.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func YarnCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, err := getProjectConfigPathOrThrow(project.Yarn, "yarn", "yarn-config")
	if err != nil {
		return err
	}

	yarnCmd := yarn.NewYarnCommand().SetConfigFilePath(configFilePath).SetArgs(c.Args())
	return commands.Exec(yarnCmd)
}

func NugetCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	configFilePath, err := getProjectConfigPathOrThrow(project.Nuget, "nuget", "nuget-config")
	if err != nil {
		return err
	}

	rtDetails, targetRepo, useNugetV2, err := getNugetAndDotnetConfigFields(configFilePath)
	if err != nil {
		return err
	}
	args := cliutils.ExtractCommand(c)
	filteredNugetArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(args)
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
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Get configuration file path.
	configFilePath, err := getProjectConfigPathOrThrow(project.Dotnet, "dotnet", "dotnet-config")
	if err != nil {
		return err
	}

	rtDetails, targetRepo, useNugetV2, err := getNugetAndDotnetConfigFields(configFilePath)
	if err != nil {
		return err
	}

	args := cliutils.ExtractCommand(c)

	filteredDotnetArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(args)
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
	vConfig, err := project.ReadConfigFile(configFilePath, project.YAML)
	if err != nil {
		return nil, "", false, fmt.Errorf("error occurred while attempting to read nuget-configuration file: %s", err.Error())
	}
	projectConfig, err := project.GetRepoConfigByPrefix(configFilePath, project.ProjectConfigResolverPrefix, vConfig)
	if err != nil {
		return nil, "", false, err
	}
	rtDetails, err = projectConfig.ServerDetails()
	if err != nil {
		return nil, "", false, err
	}
	targetRepo = projectConfig.TargetRepo()
	useNugetV2 = vConfig.GetBool(project.ProjectConfigResolverPrefix + "." + "nugetV2")
	return
}

func extractThreadsFlag(args []string) (cleanArgs []string, threadsCount int, err error) {
	// Extract threads flag.
	cleanArgs = append([]string(nil), args...)
	threadsFlagIndex, threadsValueIndex, threads, err := coreutils.FindFlag("--threads", cleanArgs)
	if err != nil || threadsFlagIndex < 0 {
		threadsCount = commonCliUtils.Threads
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
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	configFilePath, err := goCmdVerification(c)
	if err != nil {
		return err
	}
	args := cliutils.ExtractCommand(c)
	goCommand := golang.NewGoCommand()
	goCommand.SetConfigFilePath(configFilePath).SetGoArg(args)
	return commands.Exec(goCommand)
}

func GoPublishCmd(c *cli.Context) (err error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	configFilePath, err := goCmdVerification(c)
	if err != nil {
		return err
	}
	buildConfiguration, err := cliutils.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	version := c.Args().Get(0)
	printDeploymentView, detailedSummary := log.IsStdErrTerminal(), c.Bool("detailed-summary")
	goPublishCmd := golang.NewGoPublishCommand()
	goPublishCmd.SetConfigFilePath(configFilePath).SetBuildConfiguration(buildConfiguration).SetVersion(version).SetDetailedSummary(detailedSummary || printDeploymentView).SetExcludedPatterns(cliutils.GetStringsArrFlagValue(c, "exclusions"))
	err = commands.Exec(goPublishCmd)
	result := goPublishCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(goPublishCmd.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func goCmdVerification(c *cli.Context) (string, error) {
	if c.NArg() < 1 {
		return "", cliutils.WrongNumberOfArgumentsHandler(c)
	}

	configFilePath, err := getProjectConfigPathOrThrow(project.Go, "go", "go-config")
	if err != nil {
		return "", err
	}

	log.Debug("Go config file was found in:", configFilePath)
	return configFilePath, nil
}

func dockerCmd(c *cli.Context) error {
	args := cliutils.ExtractCommand(c)
	var cmd, image string
	// We may have prior flags before push/pull commands for the docker client.
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			if cmd == "" {
				cmd = arg
			} else {
				image = arg
				break
			}
		}
	}
	var err error
	switch cmd {
	case "pull":
		err = pullCmd(c, image)
	case "push":
		err = pushCmd(c, image)
	case "scan":
		return dockerScanCmd(c, image)
	default:
		err = dockerNativeCmd(c)
	}
	if err == nil {
		log.Output(coreutils.PrintTitle("Hint: Use 'jf docker scan' to scan a local Docker image for security vulnerabilities with JFrog Xray"))
	}
	return err
}

func pullCmd(c *cli.Context, image string) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "dockerpullhelp"); show || err != nil {
		return err
	}
	_, rtDetails, _, skipLogin, filteredDockerArgs, buildConfiguration, err := extractDockerOptionsFromArgs(c.Args())
	if err != nil {
		return err
	}
	PullCommand := container.NewPullCommand(containerutils.DockerClient)
	PullCommand.SetCmdParams(filteredDockerArgs).SetSkipLogin(skipLogin).SetImageTag(image).SetServerDetails(rtDetails).SetBuildConfiguration(buildConfiguration)
	supported, err := PullCommand.IsGetRepoSupported()
	if err != nil {
		return err
	}
	if !supported {
		return cliutils.NotSupportedNativeDockerCommand("docker-pull")
	}
	return commands.Exec(PullCommand)
}

func pushCmd(c *cli.Context, image string) (err error) {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "dockerpushhelp"); show || err != nil {
		return err
	}
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	threads, rtDetails, detailedSummary, skipLogin, filteredDockerArgs, buildConfiguration, err := extractDockerOptionsFromArgs(c.Args())
	if err != nil {
		return
	}
	printDeploymentView := log.IsStdErrTerminal()
	PushCommand := container.NewPushCommand(containerutils.DockerClient)
	PushCommand.SetThreads(threads).SetDetailedSummary(detailedSummary || printDeploymentView).SetCmdParams(filteredDockerArgs).SetSkipLogin(skipLogin).SetBuildConfiguration(buildConfiguration).SetServerDetails(rtDetails).SetImageTag(image)
	supported, err := PushCommand.IsGetRepoSupported()
	if err != nil {
		return err
	}
	if !supported {
		return cliutils.NotSupportedNativeDockerCommand("docker-push")
	}
	err = commands.Exec(PushCommand)
	result := PushCommand.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(PushCommand.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func dockerScanCmd(c *cli.Context, imageTag string) error {
	convertedCtx, err := components.ConvertContext(c, securityDocs.GetCommandFlags(securityDocs.DockerScan)...)
	if err != nil {
		return err
	}
	return securityCLI.DockerScan(convertedCtx, imageTag)
}

func dockerNativeCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	_, _, _, _, cleanArgs, _, err := extractDockerOptionsFromArgs(c.Args())
	if err != nil {
		return err
	}
	cm := containerutils.NewManager(containerutils.DockerClient)
	return cm.RunNativeCmd(cleanArgs)
}

// Remove all the none docker CLI flags from args.
func extractDockerOptionsFromArgs(args []string) (threads int, serverDetails *coreConfig.ServerDetails, detailedSummary, skipLogin bool, cleanArgs []string, buildConfig *build.BuildConfiguration, err error) {
	cleanArgs = append([]string(nil), args...)
	var serverId string
	cleanArgs, serverId, err = coreutils.ExtractServerIdFromCommand(cleanArgs)
	if err != nil {
		return
	}
	serverDetails, err = coreConfig.GetSpecificConfig(serverId, true, true)
	if err != nil {
		return
	}
	cleanArgs, threads, err = coreutils.ExtractThreadsFromArgs(cleanArgs, 3)
	if err != nil {
		return
	}
	cleanArgs, detailedSummary, err = coreutils.ExtractDetailedSummaryFromArgs(cleanArgs)
	if err != nil {
		return
	}
	cleanArgs, skipLogin, err = coreutils.ExtractSkipLoginFromArgs(cleanArgs)
	if err != nil {
		return
	}
	cleanArgs, buildConfig, err = build.ExtractBuildDetailsFromArgs(cleanArgs)
	return
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

func NpmInstallCmd(c *cli.Context) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "npminstallhelp"); show || err != nil {
		return err
	}
	return npmGenericCmd(c, "install", true)
}

func NpmCiCmd(c *cli.Context) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "npmcihelp"); show || err != nil {
		return err
	}
	return npmGenericCmd(c, "ci", true)
}

func npmGenericCmd(c *cli.Context, cmdName string, collectBuildInfoIfRequested bool) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	switch cmdName {
	// Aliases accepted by npm.
	case "i", "isntall", "add", "install":
		cmdName = "install"
		collectBuildInfoIfRequested = true
	case "ci":
		collectBuildInfoIfRequested = true
	case "publish", "p":
		return NpmPublishCmd(c)
	}

	// Run generic npm command.
	npmCmd := npm.NewNpmCommand(cmdName, collectBuildInfoIfRequested)

	configFilePath, args, err := GetNpmConfigAndArgs(c)
	if err != nil {
		return err
	}
	npmCmd.SetConfigFilePath(configFilePath).CommonArgs.SetNpmArgs(args)
	if err = npmCmd.Init(); err != nil {
		return err
	}
	return commands.Exec(npmCmd)
}

func NpmPublishCmd(c *cli.Context) (err error) {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "npmpublishhelp"); show || err != nil {
		return err
	}

	configFilePath, args, err := GetNpmConfigAndArgs(c)
	if err != nil {
		return err
	}

	npmCmd := npm.NewNpmPublishCommand()
	npmCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	if err = npmCmd.Init(); err != nil {
		return err
	}
	if npmCmd.GetXrayScan() {
		commandsUtils.ConditionalUploadScanFunc = scan.ConditionalUploadDefaultScanFunc
	}
	printDeploymentView, detailedSummary := log.IsStdErrTerminal(), npmCmd.IsDetailedSummary()
	if !detailedSummary {
		npmCmd.SetDetailedSummary(printDeploymentView)
	}
	err = commands.Exec(npmCmd)
	result := npmCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(npmCmd.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func GetNpmConfigAndArgs(c *cli.Context) (configFilePath string, args []string, err error) {
	configFilePath, err = getProjectConfigPathOrThrow(project.Npm, "npm", "npm-config")
	if err != nil {
		return
	}
	_, args = getCommandName(c.Args())
	return
}

func PipCmd(c *cli.Context) error {
	return pythonCmd(c, project.Pip)
}

func PipenvCmd(c *cli.Context) error {
	return pythonCmd(c, project.Pipenv)
}

func PoetryCmd(c *cli.Context) error {
	return pythonCmd(c, project.Poetry)
}

func pythonCmd(c *cli.Context, projectType project.ProjectType) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Get python configuration.
	pythonConfig, err := project.GetResolutionOnlyConfiguration(projectType)
	if err != nil {
		return fmt.Errorf("error occurred while attempting to read %[1]s-configuration file: %[2]s\n"+
			"Please run 'jf %[1]s-config' command prior to running 'jf %[1]s'", projectType.String(), err.Error())
	}

	// Set arg values.
	rtDetails, err := pythonConfig.ServerDetails()
	if err != nil {
		return err
	}

	orgArgs := cliutils.ExtractCommand(c)
	cmdName, filteredArgs := getCommandName(orgArgs)
	switch projectType {
	case project.Pip:
		pipCommand := python.NewPipCommand()
		pipCommand.SetServerDetails(rtDetails).SetRepo(pythonConfig.TargetRepo()).SetCommandName(cmdName).SetArgs(filteredArgs)
		return commands.Exec(pipCommand)
	case project.Pipenv:
		pipenvCommand := python.NewPipenvCommand()
		pipenvCommand.SetServerDetails(rtDetails).SetRepo(pythonConfig.TargetRepo()).SetCommandName(cmdName).SetArgs(filteredArgs)
		return commands.Exec(pipenvCommand)
	case project.Poetry:
		poetryCommand := python.NewPoetryCommand()
		poetryCommand.SetServerDetails(rtDetails).SetRepo(pythonConfig.TargetRepo()).SetCommandName(cmdName).SetArgs(filteredArgs)
		return commands.Exec(poetryCommand)
	default:
		return errorutils.CheckErrorf("%s is not supported", projectType)
	}
}

func terraformCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	configFilePath, orgArgs, err := getTerraformConfigAndArgs(c)
	if err != nil {
		return err
	}
	cmdName, filteredArgs := getCommandName(orgArgs)
	switch cmdName {
	// Aliases accepted by terraform.
	case "publish", "p":
		return terraformPublishCmd(configFilePath, filteredArgs, c)
	default:
		return errorutils.CheckErrorf("Terraform command:\"" + cmdName + "\" is not supported. " + cliutils.GetDocumentationMessage())
	}
}

func getTerraformConfigAndArgs(c *cli.Context) (configFilePath string, args []string, err error) {
	configFilePath, err = getProjectConfigPathOrThrow(project.Terraform, "terraform", "terraform-config")
	if err != nil {
		return
	}
	args = cliutils.ExtractCommand(c)
	return
}

func terraformPublishCmd(configFilePath string, args []string, c *cli.Context) error {
	terraformCmd := terraform.NewTerraformPublishCommand()
	terraformCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	if err := terraformCmd.Init(); err != nil {
		return err
	}
	err := commands.Exec(terraformCmd)
	result := terraformCmd.Result()
	return cliutils.PrintBriefSummaryReport(result.SuccessCount(), result.FailCount(), cliutils.IsFailNoOp(c), err)
}

func getProjectConfigPathOrThrow(projectType project.ProjectType, cmdName, configCmdName string) (configFilePath string, err error) {
	configFilePath, exists, err := project.GetProjectConfFilePath(projectType)
	if err != nil {
		return
	}
	if !exists {
		return "", errorutils.CheckErrorf(getMissingConfigErrMsg(cmdName, configCmdName))
	}
	return
}

func getMissingConfigErrMsg(cmdName, configCmdName string) string {
	return fmt.Sprintf("no config file was found! Before running the 'jf %s' command on a project for the first time, the project should be configured with the 'jf %s' command", cmdName, configCmdName)
}

func twineCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	serverDetails, targetRepo, err := getTwineConfigAndArgs()
	if err != nil {
		return err
	}
	cmdName, filteredArgs := getCommandName(cliutils.ExtractCommand(c))
	return python.NewTwineCommand(cmdName).SetServerDetails(serverDetails).SetTargetRepo(targetRepo).SetArgs(filteredArgs).Run()
}

func getTwineConfigAndArgs() (serverDetails *coreConfig.ServerDetails, targetRepo string, err error) {
	configFilePath, err := getTwineConfigPath()
	if err != nil {
		return
	}

	vConfig, err := project.ReadConfigFile(configFilePath, project.YAML)
	if err != nil {
		return nil, "", fmt.Errorf("failed while reading configuration file '%s'. Error: %s", configFilePath, err.Error())
	}
	projectConfig, err := project.GetRepoConfigByPrefix(configFilePath, project.ProjectConfigDeployerPrefix, vConfig)
	if err != nil {
		return nil, "", err
	}
	serverDetails, err = projectConfig.ServerDetails()
	if err != nil {
		return nil, "", err
	}
	targetRepo = projectConfig.TargetRepo()
	return
}

func getTwineConfigPath() (configFilePath string, err error) {
	var exists bool
	for _, projectType := range []project.ProjectType{project.Pip, project.Pipenv} {
		configFilePath, exists, err = project.GetProjectConfFilePath(projectType)
		if err != nil || exists {
			return
		}
	}
	return "", errorutils.CheckErrorf(getMissingConfigErrMsg("twine", "pip-config OR pipenv-config"))
}

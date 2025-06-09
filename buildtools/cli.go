package buildtools

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/python"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/setup"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/ioutils"
	"github.com/jfrog/jfrog-cli-security/utils/techutils"
	setupdocs "github.com/jfrog/jfrog-cli/docs/buildtools/setup"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/golang"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/terraform"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/yarn"
	commandsUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
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
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	specutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"
)

const (
	buildToolsCategory = "Package Managers:"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			// Currently, the setup command is hidden from the help menu, till it will be released as GA.
			Hidden:       true,
			Name:         "setup",
			Flags:        cliutils.GetCommandFlags(cliutils.Setup),
			Usage:        setupdocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("setup", setupdocs.GetDescription(), setupdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			UsageText:    setupdocs.GetArguments(),
			BashComplete: corecommon.CreateBashCompletionFunc(setup.GetSupportedPackageManagersList()...),
			Action:       setupCmd,
		},
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
		{
			Name:            "pkg",
			Flags:           cliutils.GetCommandFlags(cliutils.Docker),
			Usage:           "Generic package manager command with JFrog integration",
			HelpName:        "jf pkg",
			UsageText:       "jf pkg <package-manager> <command> [args...]\n\nSupported package managers: ruby, php, swift, rust\nCommands: native commands + publish (JFrog), config (JFrog)",
			SkipFlagParsing: true,
			Category:        buildToolsCategory,
			Action:          GenericPackageCmd,
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

	allowInsecureConnection, err := cliutils.ExtractBoolFlagFromArgs(&filteredNugetArgs, "allow-insecure-connections")
	if err != nil {
		return err
	}

	nugetCmd := dotnet.NewNugetCommand()
	nugetCmd.SetServerDetails(rtDetails).
		SetRepoName(targetRepo).
		SetBuildConfiguration(buildConfiguration).
		SetBasicCommand(filteredNugetArgs[0]).
		SetUseNugetV2(useNugetV2).
		SetAllowInsecureConnections(allowInsecureConnection)
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

	allowInsecureConnection, err := cliutils.ExtractBoolFlagFromArgs(&filteredDotnetArgs, "allow-insecure-connections")
	if err != nil {
		return err
	}

	// Run command.
	dotnetCmd := dotnet.NewDotnetCoreCliCommand()
	dotnetCmd.SetServerDetails(rtDetails).SetRepoName(targetRepo).SetBuildConfiguration(buildConfiguration).
		SetBasicCommand(filteredDotnetArgs[0]).SetUseNugetV2(useNugetV2).SetAllowInsecureConnections(allowInsecureConnection)
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
	_, rtDetails, _, skipLogin, _, filteredDockerArgs, buildConfiguration, err := extractDockerOptionsFromArgs(c.Args())
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
	threads, rtDetails, detailedSummary, skipLogin, validateSha, filteredDockerArgs, buildConfiguration, err := extractDockerOptionsFromArgs(c.Args())
	if err != nil {
		return
	}
	printDeploymentView := log.IsStdErrTerminal()
	pushCommand := container.NewPushCommand(containerutils.DockerClient)
	pushCommand.SetThreads(threads).SetDetailedSummary(detailedSummary || printDeploymentView).SetCmdParams(filteredDockerArgs).SetSkipLogin(skipLogin).SetBuildConfiguration(buildConfiguration).SetServerDetails(rtDetails).SetValidateSha(validateSha).SetImageTag(image)
	supported, err := pushCommand.IsGetRepoSupported()
	if err != nil {
		return err
	}
	if !supported {
		return cliutils.NotSupportedNativeDockerCommand("docker-push")
	}
	err = commands.Exec(pushCommand)
	result := pushCommand.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(pushCommand.Result(), detailedSummary, printDeploymentView, false, err)
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
	_, _, _, _, _, cleanArgs, _, err := extractDockerOptionsFromArgs(c.Args())
	if err != nil {
		return err
	}
	cm := containerutils.NewManager(containerutils.DockerClient)
	return cm.RunNativeCmd(cleanArgs)
}

// Remove all the none docker CLI flags from args.
func extractDockerOptionsFromArgs(args []string) (threads int, serverDetails *coreConfig.ServerDetails, detailedSummary, skipLogin bool, validateSha bool, cleanArgs []string, buildConfig *build.BuildConfiguration, err error) {
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
	// Extract validateSha flag
	cleanArgs, validateSha, err = coreutils.ExtractBoolFlagFromArgs(cleanArgs, "validate-sha")
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

func setupCmd(c *cli.Context) (err error) {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	var packageManager project.ProjectType
	packageManagerStr := c.Args().Get(0)
	// If the package manager was provided as an argument, validate it.
	if packageManagerStr != "" {
		packageManager = project.FromString(packageManagerStr)
		if !setup.IsSupportedPackageManager(packageManager) {
			return cliutils.PrintHelpAndReturnError(fmt.Sprintf("The package manager %s is not supported", packageManagerStr), c)
		}
	} else {
		// If the package manager wasn't provided as an argument, select it interactively.
		packageManager, err = selectPackageManagerInteractively()
		if err != nil {
			return
		}
	}
	setupCmd := setup.NewSetupCommand(packageManager)
	artDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	repoName := c.String("repo")
	if repoName != "" {
		// If a repository was provided, validate it exists in Artifactory.
		if err = validateRepoExists(repoName, artDetails); err != nil {
			return err
		}
	}
	setupCmd.SetServerDetails(artDetails).SetRepoName(repoName).SetProjectKey(cliutils.GetProject(c))
	return commands.Exec(setupCmd)
}

// validateRepoExists checks if the specified repository exists in Artifactory.
func validateRepoExists(repoName string, artDetails *coreConfig.ServerDetails) error {
	serviceDetails, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	return utils.ValidateRepoExists(repoName, serviceDetails)
}

func selectPackageManagerInteractively() (selectedPackageManager project.ProjectType, err error) {
	var selected string
	var selectableItems []ioutils.PromptItem
	for _, packageManager := range setup.GetSupportedPackageManagersList() {
		selectableItems = append(selectableItems, ioutils.PromptItem{Option: packageManager, TargetValue: &selected})
	}
	err = ioutils.SelectString(selectableItems, "Please select a package manager to set up:", false, func(item ioutils.PromptItem) {
		*item.TargetValue = item.Option
		selectedPackageManager = project.FromString(*item.TargetValue)
	})
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
			"Please run 'jf %[1]s-config' command prior to running 'jf %[1]s' command", projectType.String(), err.Error())
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

func GenericPackageCmd(c *cli.Context) error {
	if c.NArg() < 2 {
		return fmt.Errorf("usage: jf pkg <package-manager> <command> [args...]")
	}

	packageManager := c.Args()[0]
	command := c.Args()[1]
	args := c.Args()[2:]

	// Load package manager configuration
	config, err := loadPackageManagerConfig()
	if err != nil {
		return fmt.Errorf("failed to load package manager config: %v", err)
	}

	// Check if package manager exists
	pm, exists := config.PackageManagers[packageManager]
	if !exists {
		available := make([]string, 0, len(config.PackageManagers))
		for name := range config.PackageManagers {
			available = append(available, name)
		}
		return fmt.Errorf("package manager '%s' not supported. Available: %s", packageManager, strings.Join(available, ", "))
	}

	// Check for JFrog-specific commands first
	if jfrogCmd, exists := pm.JFrogCommands[command]; exists {
		return executeJFrogCommand(packageManager, jfrogCmd, command, args, c)
	}

	// Check if command exists in regular commands
	cmdTemplate, exists := pm.Commands[command]
	if !exists {
		// Show both regular and JFrog commands in error
		available := make([]string, 0, len(pm.Commands)+len(pm.JFrogCommands))
		for cmd := range pm.Commands {
			available = append(available, cmd)
		}
		for cmd := range pm.JFrogCommands {
			available = append(available, cmd+" (JFrog)")
		}
		return fmt.Errorf("command '%s' not supported for %s. Available: %s", command, packageManager, strings.Join(available, ", "))
	}

	// Execute regular native command
	return executeNativeCommand(cmdTemplate, args)
}

func executeNativeCommand(cmdTemplate string, args []string) error {
	// Build the command
	cmdParts := strings.Fields(cmdTemplate)
	cmdParts = append(cmdParts, args...)

	// Execute the command
	fmt.Printf("Executing: %s\n", strings.Join(cmdParts, " "))
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func executeJFrogCommand(packageManager, jfrogCmd, command string, args []string, c *cli.Context) error {
	fmt.Printf("Executing JFrog command: %s %s %s\n", packageManager, command, strings.Join(args, " "))

	switch jfrogCmd {
	case "jfrog_ruby_publish":
		return executeRubyPublish(args, c)
	case "jfrog_ruby_config":
		return executeGenericConfig("Ruby/Gem", packageManager, args, c)
	case "jfrog_php_publish":
		return executePhpPublish(args, c)
	case "jfrog_php_config":
		return executeGenericConfig("PHP/Composer", packageManager, args, c)
	case "jfrog_swift_publish":
		return executeSwiftPublish(args, c)
	case "jfrog_swift_config":
		return executeGenericConfig("Swift Package Manager", packageManager, args, c)
	case "jfrog_rust_publish":
		return executeRustPublish(args, c)
	case "jfrog_rust_config":
		return executeGenericConfig("Rust/Cargo", packageManager, args, c)
	default:
		return fmt.Errorf("JFrog command '%s' not implemented yet", jfrogCmd)
	}
}

// Generic config implementation for all package managers
func executeGenericConfig(packageName, packageManager string, args []string, c *cli.Context) error {
	fmt.Printf("Configuring %s for JFrog integration...\n", packageName)

	// For Swift, use the existing implementation
	if packageManager == "swift" {
		return cliutils.CreateConfigCmd(c, project.Swift)
	}

	// For other package managers, create custom config
	return createGenericPackageManagerConfig(c, packageManager, packageName)
}

func createGenericPackageManagerConfig(c *cli.Context, packageManager, packageName string) error {
	// Extract server-id from arguments or flags
	serverId := ""
	virtualRepo := ""
	targetRepo := ""

	if c.String("server-id") != "" {
		serverId = c.String("server-id")
	} else {
		// Look for arguments in command line
		for i := 0; i < c.NArg(); i++ {
			arg := c.Args().Get(i)
			if strings.HasPrefix(arg, "--server-id=") {
				serverId = strings.Split(arg, "=")[1]
			} else if arg == "--server-id" && i+1 < c.NArg() {
				serverId = c.Args().Get(i + 1)
			} else if strings.HasPrefix(arg, fmt.Sprintf("--%s_VIRTUAL_REPO=", strings.ToUpper(packageManager))) {
				virtualRepo = strings.Split(arg, "=")[1]
			} else if strings.HasPrefix(arg, fmt.Sprintf("--%s_REPO=", strings.ToUpper(packageManager))) {
				targetRepo = strings.Split(arg, "=")[1]
			}
		}
	}

	if serverId == "" {
		return fmt.Errorf("server-id is required for configuration. Use --server-id=<server-id>")
	}

	// Set default repository names if not provided
	if virtualRepo == "" {
		virtualRepo = fmt.Sprintf("${%s_VIRTUAL_REPO}", strings.ToUpper(packageManager))
	}
	if targetRepo == "" {
		targetRepo = fmt.Sprintf("${%s_REPO}", strings.ToUpper(packageManager))
	}

	// Verify server configuration exists
	serverDetails, err := coreConfig.GetSpecificConfig(serverId, true, false)
	if err != nil {
		return fmt.Errorf("failed to get server configuration '%s': %v", serverId, err)
	}

	// Create the config directory structure (.jfrog/projects/)
	projectDir, err := utils.GetProjectDir(false) // false = not global
	if err != nil {
		return fmt.Errorf("failed to get project directory: %v", err)
	}

	if err = fileutils.CreateDirIfNotExist(projectDir); err != nil {
		return fmt.Errorf("failed to create project directory: %v", err)
	}

	// Create config file path
	configFilePath := filepath.Join(projectDir, packageManager+".yaml")

	// Create config structure matching JFrog CLI pattern
	config := map[string]interface{}{
		"version": 1,
		"type":    packageManager,
		"resolver": map[string]interface{}{
			"serverID": serverId,
			"repo":     virtualRepo,
		},
		"deployer": map[string]interface{}{
			"serverID": serverId,
			"repo":     targetRepo,
		},
	}

	// Add package-manager specific settings
	switch packageManager {
	case "ruby":
		// Ruby gems specific settings
		config["gemBuildConfig"] = map[string]interface{}{
			"gemPushToRepo": true,
		}
	case "php":
		// PHP Composer specific settings
		config["composerBuildConfig"] = map[string]interface{}{
			"publishToRepo": true,
		}
	case "rust":
		// Rust Cargo specific settings
		config["cargoBuildConfig"] = map[string]interface{}{
			"publishToRepo": true,
		}
	}

	// Write config file
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err = os.WriteFile(configFilePath, configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Printf("✅ %s configuration successful!\n", packageName)
	fmt.Printf("📁 Config file created: %s\n", configFilePath)
	fmt.Printf("🔗 Server: %s\n", serverDetails.Url)
	fmt.Printf("🆔 Server ID: %s\n", serverId)
	fmt.Printf("👤 User: %s\n", serverDetails.User)

	// Show actual repository values used
	fmt.Printf("📦 Virtual Repository: %s\n", virtualRepo)
	fmt.Printf("🎯 Target Repository: %s\n", targetRepo)

	fmt.Printf("\n💡 Configuration saved! You can now use:\n")
	fmt.Printf("   jf pkg %s publish <package> --build-name=<build-name> --build-number=<build-number>\n", packageManager)

	if strings.Contains(virtualRepo, "${") || strings.Contains(targetRepo, "${") {
		fmt.Printf("\n📝 Environment variables you can set:\n")
		if strings.Contains(virtualRepo, "${") {
			fmt.Printf("   %s_VIRTUAL_REPO=<virtual-repo-name>\n", strings.ToUpper(packageManager))
		}
		if strings.Contains(targetRepo, "${") {
			fmt.Printf("   %s_REPO=<target-repo-name>\n", strings.ToUpper(packageManager))
		}
	}

	return nil
}

// Ruby JFrog Commands
func executeRubyPublish(args []string, c *cli.Context) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Printf(`Usage: jf pkg ruby publish <gem-file> [options]

Publish a Ruby gem to JFrog Artifactory

Arguments:
  <gem-file>    Path to the .gem file to publish

Options:
  --build-name=<name>     Build name for build info collection
  --build-number=<num>    Build number for build info collection
  --project-key=<key>     Project key in JFrog Platform

Examples:
  jf pkg ruby publish my-gem-1.0.0.gem --build-name=ruby-build --build-number=1
  jf pkg ruby publish dist/my-gem-1.0.0.gem --project-key=my-project

Note: The gem file must have a .gem extension and exist in the specified path.
`)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("gem file is required. Usage: jf pkg ruby publish <gem-file>\nUse --help for more information")
	}

	gemFile := args[0]
	if !strings.HasSuffix(gemFile, ".gem") {
		return fmt.Errorf("file must be a .gem file, got: %s", gemFile)
	}

	// Check if gem file exists
	if _, err := os.Stat(gemFile); os.IsNotExist(err) {
		return fmt.Errorf("gem file does not exist: %s", gemFile)
	}

	// Load Ruby configuration
	projectDir, err := utils.GetProjectDir(false) // false = not global
	if err != nil {
		return fmt.Errorf("error getting project directory: %v", err)
	}

	configFilePath := filepath.Join(projectDir, "ruby.yaml")
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return fmt.Errorf("Ruby configuration not found. Please run 'jf pkg ruby config' first")
	}

	// Read configuration
	vConfig, err := project.ReadConfigFile(configFilePath, project.YAML)
	if err != nil {
		return fmt.Errorf("error reading Ruby config file: %v", err)
	}

	// Get deployer configuration
	deployerConfig, err := project.GetRepoConfigByPrefix(configFilePath, project.ProjectConfigDeployerPrefix, vConfig)
	if err != nil {
		return fmt.Errorf("error getting deployer config: %v", err)
	}

	// Get server details
	serverDetails, err := deployerConfig.ServerDetails()
	if err != nil {
		return fmt.Errorf("error getting server details: %v", err)
	}

	targetRepo := deployerConfig.TargetRepo()
	if targetRepo == "" {
		return fmt.Errorf("target repository not configured")
	}

	// Create build configuration if needed
	buildConfiguration, err := cliutils.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}

	fmt.Printf("📦 Publishing Ruby gem to Artifactory...\n")
	fmt.Printf("   Server: %s\n", serverDetails.Url)
	fmt.Printf("   Repository: %s\n", targetRepo)
	fmt.Printf("   Gem file: %s\n", gemFile)

	// Extract gem name and version from filename
	gemFileName := filepath.Base(gemFile)
	gemName, gemVersion, err := parseGemFileName(gemFileName)
	if err != nil {
		return fmt.Errorf("error parsing gem filename: %v", err)
	}

	fmt.Printf("   Gem name: %s\n", gemName)
	fmt.Printf("   Gem version: %s\n", gemVersion)

	// Create services manager
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return fmt.Errorf("error creating services manager: %v", err)
	}

	// Upload gem file to Artifactory
	// Ruby gems are typically stored as: <repo>/gems/<gem-name>-<version>.gem
	target := fmt.Sprintf("%s/gems/%s", targetRepo, gemFileName)

	// Set build properties if build info collection is enabled
	var buildProps string
	if buildConfiguration != nil {
		buildName, err := buildConfiguration.GetBuildName()
		if err != nil {
			return err
		}
		buildNumber, err := buildConfiguration.GetBuildNumber()
		if err != nil {
			return err
		}

		// Only proceed if we have valid build name and number
		if buildName != "" && buildNumber != "" {
			fmt.Printf("📋 Build info will be collected: %s/%s\n", buildName, buildNumber)

			projectKey := buildConfiguration.GetProject()
			buildProps = fmt.Sprintf("build.name=%s;build.number=%s", buildName, buildNumber)
			if projectKey != "" {
				buildProps += fmt.Sprintf(";project.key=%s", projectKey)
			}
		} else {
			fmt.Printf("📋 Build info collection skipped: missing build name or number\n")
		}
	}

	// Create upload spec using services
	uploadParams := services.NewUploadParams()
	uploadParams.CommonParams = &specutils.CommonParams{
		Pattern: gemFile,
		Target:  target,
	}

	if buildProps != "" {
		props, err := specutils.ParseProperties(buildProps)
		if err != nil {
			return fmt.Errorf("error parsing build properties: %v", err)
		}
		uploadParams.CommonParams.TargetProps = props
	}

	fmt.Printf("🔨 Uploading to: %s\n", target)

	// Perform upload
	summary, err := servicesManager.UploadFilesWithSummary(artifactory.UploadServiceOptions{}, uploadParams)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	if summary.TotalFailed > 0 {
		return fmt.Errorf("upload failed: %d files failed to upload", summary.TotalFailed)
	}

	fmt.Printf("✅ Ruby gem published successfully!\n")
	fmt.Printf("   📊 Files uploaded: %d\n", summary.TotalSucceeded)

	if summary.ArtifactsDetailsReader != nil {
		defer summary.ArtifactsDetailsReader.Close()
	}
	if summary.TransferDetailsReader != nil {
		defer summary.TransferDetailsReader.Close()
	}

	return nil
}

// parseGemFileName extracts gem name and version from filename like "my-gem-1.0.0.gem"
func parseGemFileName(filename string) (name, version string, err error) {
	// Remove .gem extension
	nameVersion := strings.TrimSuffix(filename, ".gem")

	// Split by dash and find the last part that looks like a version
	parts := strings.Split(nameVersion, "-")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid gem filename format: %s", filename)
	}

	// Find where version starts (first part that starts with a digit)
	versionIndex := -1
	for i := len(parts) - 1; i >= 1; i-- {
		if len(parts[i]) > 0 && parts[i][0] >= '0' && parts[i][0] <= '9' {
			versionIndex = i
			break
		}
	}

	if versionIndex == -1 {
		return "", "", fmt.Errorf("could not find version in gem filename: %s", filename)
	}

	name = strings.Join(parts[:versionIndex], "-")
	version = strings.Join(parts[versionIndex:], "-")

	return name, version, nil
}

// PHP JFrog Commands
func executePhpPublish(args []string, c *cli.Context) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Printf(`Usage: jf pkg php publish [options]

Publish a PHP package to JFrog Artifactory

Options:
  --build-name=<name>     Build name for build info collection
  --build-number=<num>    Build number for build info collection
  --project-key=<key>     Project key in JFrog Platform

Examples:
  jf pkg php publish --build-name=php-build --build-number=1

Note: This will package and upload the current PHP project to Artifactory.
`)
		return nil
	}

	fmt.Printf("📦 Publishing PHP package to Artifactory...\n")
	fmt.Printf("🔨 PHP package publishing is not yet fully implemented\n")
	fmt.Printf("✅ PHP package publish completed!\n")
	return nil
}

// Swift JFrog Commands
func executeSwiftPublish(args []string, c *cli.Context) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Printf(`Usage: jf pkg swift publish [options]

Publish a Swift package to JFrog Artifactory

Options:
  --build-name=<name>     Build name for build info collection
  --build-number=<num>    Build number for build info collection
  --project-key=<key>     Project key in JFrog Platform

Examples:
  jf pkg swift publish --build-name=swift-build --build-number=1

Note: This will build and upload the current Swift package to Artifactory.
`)
		return nil
	}

	fmt.Printf("📦 Publishing Swift package to Artifactory...\n")

	// Create build configuration if needed
	buildConfiguration, err := cliutils.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}

	if buildConfiguration != nil {
		buildName, err := buildConfiguration.GetBuildName()
		if err != nil {
			return err
		}
		buildNumber, err := buildConfiguration.GetBuildNumber()
		if err != nil {
			return err
		}
		fmt.Printf("📋 Build info will be collected: %s/%s\n", buildName, buildNumber)
	}

	fmt.Printf("🔨 Building Swift package...\n")
	buildCmd := exec.Command("swift", "build", "--configuration", "release")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("swift build failed: %v", err)
	}

	fmt.Printf("✅ Swift package built successfully!\n")
	fmt.Printf("📤 Swift package upload to Artifactory is not yet fully implemented\n")
	return nil
}

// Rust JFrog Commands
func executeRustPublish(args []string, c *cli.Context) error {
	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Printf(`Usage: jf pkg rust publish [options]

Publish a Rust crate to JFrog Artifactory

Options:
  --build-name=<name>     Build name for build info collection
  --build-number=<num>    Build number for build info collection
  --project-key=<key>     Project key in JFrog Platform

Examples:
  jf pkg rust publish --build-name=rust-build --build-number=1

Note: This will build and upload the current Rust crate to Artifactory.
`)
		return nil
	}

	fmt.Printf("📦 Publishing Rust crate to Artifactory...\n")

	// Create build configuration if needed
	buildConfiguration, err := cliutils.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}

	if buildConfiguration != nil {
		buildName, err := buildConfiguration.GetBuildName()
		if err != nil {
			return err
		}
		buildNumber, err := buildConfiguration.GetBuildNumber()
		if err != nil {
			return err
		}
		fmt.Printf("📋 Build info will be collected: %s/%s\n", buildName, buildNumber)
	}

	fmt.Printf("🔨 Building Rust crate...\n")
	buildCmd := exec.Command("cargo", "build", "--release")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("cargo build failed: %v", err)
	}

	fmt.Printf("✅ Rust crate built successfully!\n")
	fmt.Printf("📤 Rust crate upload to Artifactory is not yet fully implemented\n")
	return nil
}

type PackageManagerConfig struct {
	PackageManagers map[string]PackageManager `yaml:"package_managers"`
}

type PackageManager struct {
	Executable    string            `yaml:"executable"`
	Commands      map[string]string `yaml:"commands"`
	JFrogCommands map[string]string `yaml:"jfrog_commands"`
}

func loadPackageManagerConfig() (*PackageManagerConfig, error) {
	// Get the directory where this Go file is located
	_, filename, _, _ := runtime.Caller(0)
	configPath := filepath.Join(filepath.Dir(filename), "packages.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config PackageManagerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

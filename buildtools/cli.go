package buildtools

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jfrog/build-info-go/utils/pythonutils"

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
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	containerutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	terraformdocs "github.com/jfrog/jfrog-cli/docs/artifactory/terraform"
	"github.com/jfrog/jfrog-cli/docs/artifactory/terraformconfig"
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
	yarndocs "github.com/jfrog/jfrog-cli/docs/buildtools/yarn"
	"github.com/jfrog/jfrog-cli/docs/buildtools/yarnconfig"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/scan"
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
				return cliutils.CreateConfigCmd(c, utils.Maven)
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
			Action: func(c *cli.Context) error {
				return MvnCmd(c)
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
				return cliutils.CreateConfigCmd(c, utils.Gradle)
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
			Action: func(c *cli.Context) error {
				return GradleCmd(c)
			},
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
				return cliutils.CreateConfigCmd(c, utils.Yarn)
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
			Action: func(c *cli.Context) error {
				return YarnCmd(c)
			},
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
				return cliutils.CreateConfigCmd(c, utils.Nuget)
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
			Action: func(c *cli.Context) error {
				return NugetCmd(c)
			},
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
				return cliutils.CreateConfigCmd(c, utils.Dotnet)
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
			Action: func(c *cli.Context) error {
				return DotnetCmd(c)
			},
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
				return cliutils.CreateConfigCmd(c, utils.Go)
			},
		},
		{
			Name:            "go",
			Flags:           cliutils.GetCommandFlags(cliutils.Go),
			Aliases:         []string{"go"},
			Usage:           gocommand.GetDescription(),
			HelpName:        corecommon.CreateUsage("go", gocommand.GetDescription(), gocommand.Usage),
			UsageText:       gocommand.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action: func(c *cli.Context) error {
				return GoCmd(c)
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
			Action: func(c *cli.Context) error {
				return GoPublishCmd(c)
			},
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
				return cliutils.CreateConfigCmd(c, utils.Pip)
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
			Action: func(c *cli.Context) error {
				return PipCmd(c)
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
				return cliutils.CreateConfigCmd(c, utils.Pipenv)
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
			Action: func(c *cli.Context) error {
				return PipenvCmd(c)
			},
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
				return cliutils.CreateConfigCmd(c, utils.Npm)
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
			Action: func(c *cli.Context) error {
				return npmGenericCmd(c)
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
					if arg == "docker" && os.Args[i+1] == "scan" {
						return false
					}
				}
				return true
			}(),
			BashComplete: corecommon.CreateBashCompletionFunc("push", "pull", "scan"),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return dockerCmd(c)
			},
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
				return cliutils.CreateConfigCmd(c, utils.Terraform)
			},
		},
		{
			Name:         "terraform",
			Flags:        cliutils.GetCommandFlags(cliutils.Terraform),
			Aliases:      []string{"tf"},
			Usage:        terraformdocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("terraform", terraformdocs.GetDescription(), terraformdocs.Usage),
			UsageText:    terraformdocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return terraformCmd(c)
			},
		},
	})
}

func MvnCmd(c *cli.Context) (err error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Maven)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("no config file was found! Before running the mvn command on a project for the first time, the project should be configured with the mvn-config command")
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
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
	printDeploymentView := log.IsStdErrTerminal()
	if !xrayScan && format != "" {
		return cliutils.PrintHelpAndReturnError("The --format option can be sent only with the --scan option", c)
	}
	scanOutputFormat, err := commandsUtils.GetXrayOutputFormat(format)
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

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Gradle)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("no config file was found! Before running the gradle command on a project for the first time, the project should be configured with the gradle-config command")
	}
	// Found a config file. Continue as native command.
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
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
	scanOutputFormat, err := commandsUtils.GetXrayOutputFormat(format)
	if err != nil {
		return err
	}
	printDeploymentView := log.IsStdErrTerminal()
	gradleCmd := gradle.NewGradleCommand().SetConfiguration(buildConfiguration).SetTasks(strings.Join(filteredGradleArgs, " ")).SetConfigPath(configFilePath).SetThreads(threads).SetDetailedSummary(detailedSummary || printDeploymentView).SetXrayScan(xrayScan).SetScanOutputFormat(scanOutputFormat)
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

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Yarn)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no config file was found! Before running the yarn command on a project for the first time, the project should be configured using the yarn-config command")
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
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Nuget)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("no config file was found! Before running the nuget command on a project for the first time, the project should be configured using the nuget-config command")
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
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Get configuration file path.
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Dotnet)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no config file was found! Before running the dotnet command on a project for the first time, the project should be configured using the dotnet-config command")
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
		return nil, "", false, fmt.Errorf("error occurred while attempting to read nuget-configuration file: %s", err.Error())
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

func GoPublishCmd(c *cli.Context) (err error) {
	configFilePath, err := goCmdVerification(c)
	if err != nil {
		return err
	}
	buildConfiguration, err := CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	version := c.Args().Get(0)
	printDeploymentView, detailedSummary := log.IsStdErrTerminal(), c.Bool("detailed-summary")
	goPublishCmd := golang.NewGoPublishCommand()
	goPublishCmd.SetConfigFilePath(configFilePath).SetBuildConfiguration(buildConfiguration).SetVersion(version).SetDetailedSummary(detailedSummary || printDeploymentView)
	err = commands.Exec(goPublishCmd)
	result := goPublishCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(goPublishCmd.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func goCmdVerification(c *cli.Context) (string, error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return "", err
	}
	if c.NArg() < 1 {
		return "", cliutils.WrongNumberOfArgumentsHandler(c)
	}
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Go)
	if err != nil {
		return "", err
	}
	// Verify config file is found.
	if !exists {
		return "", fmt.Errorf("no config file was found! Before running the go command on a project for the first time, the project should be configured using the go-config command")
	}
	log.Debug("Go config file was found in:", configFilePath)
	return configFilePath, nil
}

func CreateBuildConfigurationWithModule(c *cli.Context) (buildConfigConfiguration *utils.BuildConfiguration, err error) {
	buildConfigConfiguration = new(utils.BuildConfiguration)
	err = buildConfigConfiguration.SetBuildName(c.String("build-name")).SetBuildNumber(c.String("build-number")).SetProject(c.String("project")).SetModule(c.String("module")).ValidateBuildAndModuleParams()
	return
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
	switch cmd {
	case "pull":
		return pullCmd(c, image)
	case "push":
		return pushCmd(c, image)
	case "scan":
		return scan.DockerScan(c, image)
	default:
		return dockerNativeCmd(c)
	}
}

func pullCmd(c *cli.Context, image string) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "dockerpullhelp"); show || err != nil {
		return err
	}
	_, rtDetails, _, skipLogin, filteredDockerArgs, buildConfiguration, err := commandsUtils.ExtractDockerOptionsFromArgs(c.Args())
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
	threads, rtDetails, detailedSummary, skipLogin, filteredDockerArgs, buildConfiguration, err := commandsUtils.ExtractDockerOptionsFromArgs(c.Args())
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

func dockerNativeCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	_, _, _, _, cleanArgs, _, err := commandsUtils.ExtractDockerOptionsFromArgs(c.Args())
	if err != nil {
		return err
	}
	cm := containerutils.NewManager(containerutils.DockerClient)
	return cm.RunNativeCmd(cleanArgs)
}

func npmGenericCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	orgArgs := c.Args()
	cmdName, _ := getCommandName(orgArgs)
	switch cmdName {
	// Aliases accepted by npm.
	case "install", "i", "isntall", "add":
		return NpmInstallCmd(c)
	case "ci":
		return NpmCiCmd(c)
	case "publish", "p":
		return NpmPublishCmd(c)
	}

	// Run generic npm command.
	npmCmd := npm.NewNpmGenericCommand(cmdName)
	npmCmd.SetNpmArgs(orgArgs)
	return commands.Exec(npmCmd)
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
	return npmInstallCiCmd(c, npm.NewNpmInstallCommand())
}

func NpmCiCmd(c *cli.Context) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "npmcihelp"); show || err != nil {
		return err
	}
	return npmInstallCiCmd(c, npm.NewNpmCiCommand())
}

func npmInstallCiCmd(c *cli.Context, npmCmd *npm.NpmInstallOrCiCommand) error {
	configFilePath, args, err := GetNpmConfigAndArgs(c)
	if err != nil {
		return err
	}

	npmCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	err = npmCmd.Init()
	if err != nil {
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
	err = npmCmd.Init()
	if err != nil {
		return err
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
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Npm)
	if err != nil {
		return "", nil, err
	}

	if !exists {
		return "", nil, errorutils.CheckError(errors.New("no config file was found! Before running the npm command on a project for the first time, the project should be configured using the npm-config command"))
	}
	_, args = getCommandName(c.Args())
	return
}

func PipCmd(c *cli.Context) error {
	return pythonCmd(c, utils.Pip)
}

func PipenvCmd(c *cli.Context) error {
	return pythonCmd(c, utils.Pipenv)
}

func pythonCmd(c *cli.Context, projectType utils.ProjectType) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Get python configuration.
	pythonConfig, err := utils.GetResolutionOnlyConfiguration(projectType)
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
	var pythonTool pythonutils.PythonTool
	if projectType == utils.Pip {
		pythonTool = pythonutils.Pip
	} else if projectType == utils.Pipenv {
		pythonTool = pythonutils.Pipenv
	} else {
		return errorutils.CheckErrorf("%s command is not supported", projectType.String())
	}
	pythonCommand := python.NewPythonCommand(pythonTool)
	pythonCommand.SetServerDetails(rtDetails).SetRepo(pythonConfig.TargetRepo()).SetCommandName(cmdName).SetArgs(filteredArgs)
	return commands.Exec(pythonCommand)
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
		return errorutils.CheckError(errors.New("Terraform command:\"" + cmdName + "\" is not supported. " + cliutils.GetDocumentationMessage()))
	}
}

func getTerraformConfigAndArgs(c *cli.Context) (configFilePath string, args []string, err error) {
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Terraform)
	if err != nil {
		return "", nil, err
	}

	if !exists {
		return "", nil, errors.New("no config file was found! Before running the terraform command on a project for the first time, the project should be configured using the terraform-config command")
	}
	args = cliutils.ExtractCommand(c)
	return
}

func terraformPublishCmd(configFilePath string, args []string, c *cli.Context) error {
	artDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	terraformCmd := terraform.NewTerraformPublishCommand()
	terraformCmd.SetConfigFilePath(configFilePath).SetArgs(args).SetServerDetails(artDetails)
	err = commands.Exec(terraformCmd)
	result := terraformCmd.Result()
	return cliutils.PrintBriefSummaryReport(result.SuccessCount(), result.FailCount(), cliutils.IsFailNoOp(c), err)
}

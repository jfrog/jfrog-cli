package buildtools

import (
	"errors"
	"fmt"
	conancommand "github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/conan"
	nixcommand "github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/nix"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/container/strategies"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/python"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/setup"
	artutils "github.com/jfrog/jfrog-cli-artifactory/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/ioutils"
	"github.com/jfrog/jfrog-cli-security/utils/techutils"
	"github.com/jfrog/jfrog-cli/docs/buildtools/helmcommand"
	"github.com/jfrog/jfrog-cli/docs/buildtools/rubyconfig"
	setupdocs "github.com/jfrog/jfrog-cli/docs/buildtools/setup"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/golang"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/gradle"
	helmcmd "github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/helm"
	huggingfaceCommands "github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/huggingface"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/npm"
	containerutils "github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/ocicontainer"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/pnpm"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/terraform"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/yarn"
	commandsUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
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
	"github.com/jfrog/jfrog-cli/docs/buildtools/conan"
	"github.com/jfrog/jfrog-cli/docs/buildtools/conanconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/docker"
	dotnetdocs "github.com/jfrog/jfrog-cli/docs/buildtools/dotnet"
	"github.com/jfrog/jfrog-cli/docs/buildtools/dotnetconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/gocommand"
	"github.com/jfrog/jfrog-cli/docs/buildtools/goconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/gopublish"
	gradledoc "github.com/jfrog/jfrog-cli/docs/buildtools/gradle"
	"github.com/jfrog/jfrog-cli/docs/buildtools/gradleconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/huggingface"
	huggingfacedownloaddocs "github.com/jfrog/jfrog-cli/docs/buildtools/huggingfacedownload"
	huggingfaceuploaddocs "github.com/jfrog/jfrog-cli/docs/buildtools/huggingfaceupload"
	mvndoc "github.com/jfrog/jfrog-cli/docs/buildtools/mvn"
	"github.com/jfrog/jfrog-cli/docs/buildtools/mvnconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/nix"
	"github.com/jfrog/jfrog-cli/docs/buildtools/npmcommand"
	"github.com/jfrog/jfrog-cli/docs/buildtools/npmconfig"
	nugetdocs "github.com/jfrog/jfrog-cli/docs/buildtools/nuget"
	"github.com/jfrog/jfrog-cli/docs/buildtools/nugetconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipenvconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipenvinstall"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pipinstall"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pnpmcommand"
	"github.com/jfrog/jfrog-cli/docs/buildtools/pnpmconfig"
	"github.com/jfrog/jfrog-cli/docs/buildtools/poetry"
	"github.com/jfrog/jfrog-cli/docs/buildtools/poetryconfig"
	uvcommand "github.com/jfrog/jfrog-cli/docs/buildtools/uvcommand"
	yarndocs "github.com/jfrog/jfrog-cli/docs/buildtools/yarn"
	"github.com/jfrog/jfrog-cli/docs/buildtools/yarnconfig"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/utils/buildinfo"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

const (
	buildToolsCategory      = "Package Managers:"
	huggingfaceAPI          = "api/huggingfaceml"
	HF_ENDPOINT             = "HF_ENDPOINT"
	HF_TOKEN                = "HF_TOKEN"
	HF_HUB_ETAG_TIMEOUT     = "HF_HUB_ETAG_TIMEOUT"
	HF_HUB_DOWNLOAD_TIMEOUT = "HF_HUB_DOWNLOAD_TIMEOUT"
)

func GetCommands() []cli.Command {
	cmds := cliutils.GetSortedCommands(cli.CommandsByName{
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
			Usage:        corecommon.ResolveDescription(mvnconfig.GetDescription(), mvnconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("mvn-config", corecommon.ResolveDescription(mvnconfig.GetDescription(), mvnconfig.GetAIDescription()), mvnconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(mvndoc.GetDescription(), mvndoc.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("mvn", corecommon.ResolveDescription(mvndoc.GetDescription(), mvndoc.GetAIDescription()), mvndoc.Usage),
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
			Usage:        corecommon.ResolveDescription(gradleconfig.GetDescription(), gradleconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("gradle-config", corecommon.ResolveDescription(gradleconfig.GetDescription(), gradleconfig.GetAIDescription()), gradleconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(gradledoc.GetDescription(), gradledoc.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("gradle", corecommon.ResolveDescription(gradledoc.GetDescription(), gradledoc.GetAIDescription()), gradledoc.Usage),
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
			Usage:        corecommon.ResolveDescription(yarnconfig.GetDescription(), yarnconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("yarn-config", corecommon.ResolveDescription(yarnconfig.GetDescription(), yarnconfig.GetAIDescription()), yarnconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(yarndocs.GetDescription(), yarndocs.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("yarn", corecommon.ResolveDescription(yarndocs.GetDescription(), yarndocs.GetAIDescription()), yarndocs.Usage),
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
			Usage:        corecommon.ResolveDescription(nugetconfig.GetDescription(), nugetconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("nuget-config", corecommon.ResolveDescription(nugetconfig.GetDescription(), nugetconfig.GetAIDescription()), nugetconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(nugetdocs.GetDescription(), nugetdocs.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("nuget", corecommon.ResolveDescription(nugetdocs.GetDescription(), nugetdocs.GetAIDescription()), nugetdocs.Usage),
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
			Usage:        corecommon.ResolveDescription(dotnetconfig.GetDescription(), dotnetconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("dotnet-config", corecommon.ResolveDescription(dotnetconfig.GetDescription(), dotnetconfig.GetAIDescription()), dotnetconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(dotnetdocs.GetDescription(), dotnetdocs.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("dotnet", corecommon.ResolveDescription(dotnetdocs.GetDescription(), dotnetdocs.GetAIDescription()), dotnetdocs.Usage),
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
			Usage:        corecommon.ResolveDescription(goconfig.GetDescription(), goconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("go-config", corecommon.ResolveDescription(goconfig.GetDescription(), goconfig.GetAIDescription()), goconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(gocommand.GetDescription(), gocommand.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("go", corecommon.ResolveDescription(gocommand.GetDescription(), gocommand.GetAIDescription()), gocommand.Usage),
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
			Usage:        corecommon.ResolveDescription(gopublish.GetDescription(), gopublish.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("go-publish", corecommon.ResolveDescription(gopublish.GetDescription(), gopublish.GetAIDescription()), gopublish.Usage),
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
			Usage:        corecommon.ResolveDescription(pipconfig.GetDescription(), pipconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("pip-config", corecommon.ResolveDescription(pipconfig.GetDescription(), pipconfig.GetAIDescription()), pipconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(pipinstall.GetDescription(), pipinstall.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("pip", corecommon.ResolveDescription(pipinstall.GetDescription(), pipinstall.GetAIDescription()), pipinstall.Usage),
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
			Usage:        corecommon.ResolveDescription(pipenvconfig.GetDescription(), pipenvconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("pipenv-config", corecommon.ResolveDescription(pipenvconfig.GetDescription(), pipenvconfig.GetAIDescription()), pipenvconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(pipenvinstall.GetDescription(), pipenvinstall.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("pipenv", corecommon.ResolveDescription(pipenvinstall.GetDescription(), pipenvinstall.GetAIDescription()), pipenvinstall.Usage),
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
			Usage:        corecommon.ResolveDescription(poetryconfig.GetDescription(), poetryconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("poetry-config", corecommon.ResolveDescription(poetryconfig.GetDescription(), poetryconfig.GetAIDescription()), poetryconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(poetry.GetDescription(), poetry.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("poetry", corecommon.ResolveDescription(poetry.GetDescription(), poetry.GetAIDescription()), poetry.Usage),
			UsageText:       poetry.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          PoetryCmd,
		},
		{
			Name:            "uv",
			Flags:           cliutils.GetCommandFlags(cliutils.Uv),
			Usage:           uvcommand.GetDescription(),
			HelpName:        corecommon.CreateUsage("uv", uvcommand.GetDescription(), uvcommand.Usage),
			UsageText:       uvcommand.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          UvCmd,
		},
		{
			Name:            "helm",
			Flags:           cliutils.GetCommandFlags(cliutils.Helm),
			Usage:           corecommon.ResolveDescription(helmcommand.GetDescription(), helmcommand.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("helm", corecommon.ResolveDescription(helmcommand.GetDescription(), helmcommand.GetAIDescription()), helmcommand.Usage),
			UsageText:       helmcommand.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			Hidden:          false,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          HelmCmd,
		},
		{
			Name:         "conan-config",
			Flags:        cliutils.GetCommandFlags(cliutils.ConanConfig),
			Aliases:      []string{"conanc"},
			Usage:        corecommon.ResolveDescription(conanconfig.GetDescription(), conanconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("conan-config", corecommon.ResolveDescription(conanconfig.GetDescription(), conanconfig.GetAIDescription()), conanconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Conan)
			},
		},
		{
			Name:            "conan",
			Hidden:          false,
			Flags:           cliutils.GetCommandFlags(cliutils.Conan),
			Usage:           corecommon.ResolveDescription(conan.GetDescription(), conan.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("conan", corecommon.ResolveDescription(conan.GetDescription(), conan.GetAIDescription()), conan.Usage),
			UsageText:       conan.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          ConanCmd,
		},
		{
			Name:            "nix",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.Nix),
			Usage:           nix.GetDescription(),
			HelpName:        corecommon.CreateUsage("nix", nix.GetDescription(), nix.Usage),
			UsageText:       nix.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          NixCmd,
		},
		{
			Name:         "ruby-config",
			Flags:        cliutils.GetCommandFlags(cliutils.RubyConfig),
			Aliases:      []string{"rubyc"},
			Usage:        corecommon.ResolveDescription(rubyconfig.GetDescription(), rubyconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("ruby-config", corecommon.ResolveDescription(rubyconfig.GetDescription(), rubyconfig.GetAIDescription()), rubyconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Ruby)
			},
		},
		{
			Name:         "npm-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NpmConfig),
			Aliases:      []string{"npmc"},
			Usage:        corecommon.ResolveDescription(npmconfig.GetDescription(), npmconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("npm-config", corecommon.ResolveDescription(npmconfig.GetDescription(), npmconfig.GetAIDescription()), npmconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Npm)
			},
		},
		{
			Name:            "npm",
			Usage:           corecommon.ResolveDescription(npmcommand.GetDescription(), npmcommand.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("npm", corecommon.ResolveDescription(npmcommand.GetDescription(), npmcommand.GetAIDescription()), npmcommand.Usage),
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
			Usage:        corecommon.ResolveDescription(pnpmconfig.GetDescription(), pnpmconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("pnpm-config", corecommon.ResolveDescription(pnpmconfig.GetDescription(), pnpmconfig.GetAIDescription()), pnpmconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Category:     buildToolsCategory,
			Action: func(c *cli.Context) error {
				return cliutils.CreateConfigCmd(c, project.Pnpm)
			},
		},
		{
			Name:            "pnpm",
			Flags:           cliutils.GetCommandFlags(cliutils.Pnpm),
			Usage:           corecommon.ResolveDescription(pnpmcommand.GetDescription(), pnpmcommand.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("pnpm", corecommon.ResolveDescription(pnpmcommand.GetDescription(), pnpmcommand.GetAIDescription()), pnpmcommand.Usage),
			UsageText:       pnpmcommand.GetArguments(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc("install", "i", "publish", "p"),
			Category:        buildToolsCategory,
			Action:          pnpmCmd,
		},
		{
			Name:            "docker",
			Flags:           cliutils.GetCommandFlags(cliutils.Docker),
			Usage:           corecommon.ResolveDescription(docker.GetDescription(), docker.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("docker", corecommon.ResolveDescription(docker.GetDescription(), docker.GetAIDescription()), docker.Usage),
			UsageText:       docker.GetArguments(),
			SkipFlagParsing: skipFlagParsingForDockerCmd(),
			BashComplete:    corecommon.CreateBashCompletionFunc("login", "push", "pull", "scan"),
			Category:        buildToolsCategory,
			Action:          dockerCmd,
		},
		{
			Name:         "terraform-config",
			Flags:        cliutils.GetCommandFlags(cliutils.TerraformConfig),
			Aliases:      []string{"tfc"},
			Usage:        corecommon.ResolveDescription(terraformconfig.GetDescription(), terraformconfig.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("terraform-config", corecommon.ResolveDescription(terraformconfig.GetDescription(), terraformconfig.GetAIDescription()), terraformconfig.Usage),
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
			Usage:           corecommon.ResolveDescription(terraformdocs.GetDescription(), terraformdocs.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("terraform", corecommon.ResolveDescription(terraformdocs.GetDescription(), terraformdocs.GetAIDescription()), terraformdocs.Usage),
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
			Usage:           corecommon.ResolveDescription(twinedocs.GetDescription(), twinedocs.GetAIDescription()),
			HelpName:        corecommon.CreateUsage("twine", corecommon.ResolveDescription(twinedocs.GetDescription(), twinedocs.GetAIDescription()), twinedocs.Usage),
			UsageText:       twinedocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Category:        buildToolsCategory,
			Action:          twineCmd,
		},
		{
			Name:        "hugging-face",
			Aliases:     []string{"hf"},
			HelpName:    corecommon.CreateUsage("hugging-face", corecommon.ResolveDescription(huggingface.GetDescription(), huggingface.GetAIDescription()), huggingface.Usage),
			Description: corecommon.ResolveDescription(huggingface.GetDescription(), huggingface.GetAIDescription()),
			Hidden:      false,
			Category:    buildToolsCategory,
			Action: func(c *cli.Context) error {
				if c.Args().Present() {
					return fmt.Errorf("'%s %s' is not a valid subcommand. Run 'jf hf --help' for usage", c.App.Name, c.Args().First())
				}
				return cli.ShowSubcommandHelp(c)
			},
			Subcommands: []cli.Command{
				{
					Name:      "upload",
					Aliases:   []string{"u"},
					Flags:     cliutils.GetCommandFlags(cliutils.HuggingFaceUpload),
					HelpName:  corecommon.CreateUsage("hf upload", corecommon.ResolveDescription(huggingfaceuploaddocs.GetDescription(), huggingfaceuploaddocs.GetAIDescription()), huggingfaceuploaddocs.Usage),
					Usage:     corecommon.ResolveDescription(huggingfaceuploaddocs.GetDescription(), huggingfaceuploaddocs.GetAIDescription()),
					UsageText: huggingfaceuploaddocs.GetArguments(),
					Action:    huggingFaceUploadCmd,
				},
				{
					Name:      "download",
					Aliases:   []string{"d"},
					Flags:     cliutils.GetCommandFlags(cliutils.HuggingFaceDownload),
					HelpName:  corecommon.CreateUsage("hf download", corecommon.ResolveDescription(huggingfacedownloaddocs.GetDescription(), huggingfacedownloaddocs.GetAIDescription()), huggingfacedownloaddocs.Usage),
					Usage:     corecommon.ResolveDescription(huggingfacedownloaddocs.GetDescription(), huggingfacedownloaddocs.GetAIDescription()),
					UsageText: huggingfacedownloaddocs.GetArguments(),
					Action:    huggingFaceDownloadCmd,
				},
			},
		},
	})
	return decorateWithFlagCapture(cmds)
}

func skipFlagParsingForDockerCmd() bool {
	isDockerScan := false
	hasHelpFlag := false
	for i, arg := range os.Args {
		if arg == "docker" && len(os.Args) > i+1 && os.Args[i+1] == "scan" {
			isDockerScan = true
		}
		if arg == "--help" || arg == "-h" {
			hasHelpFlag = true
		}
	}
	// 'docker scan' isn't a docker client command. We won't skip its flags.
	if isDockerScan {
		return hasHelpFlag
	}
	return true
}

// decorateWithFlagCapture injects a Before hook into every command returned from this package,
// so we can capture user-provided flags consistently in one place for all build commands.
func decorateWithFlagCapture(cmds []cli.Command) []cli.Command {
	for i := range cmds {
		skipFlagParsing := cmds[i].SkipFlagParsing
		origBefore := cmds[i].Before
		cmds[i].Before = func(c *cli.Context) error {
			captureUserFlagsForMetrics(c, skipFlagParsing)
			if origBefore != nil {
				return origBefore(c)
			}
			return nil
		}
	}
	return cmds
}

// captureUserFlagsForMetrics extracts flag names as provided by the end-user for the given command
// and records them for usage metrics. Works even when SkipFlagParsing is true by scanning os.Args.
func captureUserFlagsForMetrics(c *cli.Context, skipFlagParsing bool) {
	flagSet := map[string]struct{}{}

	if !skipFlagParsing {
		for _, fn := range c.FlagNames() {
			flagSet[fn] = struct{}{}
		}
	} else {
		for _, next := range c.Args() {
			if !strings.HasPrefix(next, "-") {
				continue
			}
			if strings.HasPrefix(next, "--") {
				name := strings.TrimPrefix(next, "--")
				if eq := strings.Index(name, "="); eq >= 0 {
					name = name[:eq]
				}
				if name != "" {
					flagSet[name] = struct{}{}
				}
			} else {
				trimmed := strings.TrimLeft(next, "-")
				for _, ch := range trimmed {
					flagSet[string(ch)] = struct{}{}
				}
			}
		}
	}

	if len(flagSet) == 0 {
		return
	}
	flags := make([]string, 0, len(flagSet))
	for f := range flagSet {
		flags = append(flags, f)
	}
	sort.Strings(flags)
	commands.SetContextFlags(flags)
}

func MvnCmd(c *cli.Context) (err error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, configExists, err := project.GetProjectConfFilePath(project.Maven)
	if err != nil {
		return err
	}

	// FlexPack bypasses all config file requirements (only when no config exists)
	if artutils.ShouldRunNative(configFilePath) && !configExists {
		log.Debug("Routing to Maven native implementation")
		// Extract build configuration for FlexPack
		args := cliutils.ExtractCommand(c)
		filteredMavenArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(args)
		if err != nil {
			return err
		}
		// Maven does not accept --server-id; use the default configured server for usage reporting.
		serverDetails, err := coreConfig.GetDefaultServerConf()
		if err != nil {
			return err
		}
		mvnCmd := mvn.NewMvnCommand().SetConfigPath("").SetGoals(filteredMavenArgs).SetConfiguration(buildConfiguration).SetServerDetails(serverDetails)
		return commands.ExecWithPackageManager(mvnCmd, project.Maven.String())
	}

	// If config file is missing and not in native mode, return the standard missing-config error.
	if !configExists {
		if configFilePath, err = getProjectConfigPathOrThrow(project.Maven, "mvn", "mvn-config"); err != nil {
			return err
		}
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
	scanOutputFormat := outputFormat.Table
	if format != "" {
		scanOutputFormat, err = outputFormat.ParseOutputFormat(format, outputFormat.All)
		if err != nil {
			return err
		}
	}
	mvnCmd := mvn.NewMvnCommand().SetConfiguration(buildConfiguration).SetConfigPath(configFilePath).SetGoals(filteredMavenArgs).SetThreads(threads).SetInsecureTls(insecureTls).SetDetailedSummary(detailedSummary || printDeploymentView).SetXrayScan(xrayScan).SetScanOutputFormat(scanOutputFormat)
	err = commands.ExecWithPackageManager(mvnCmd, project.Maven.String())
	result := mvnCmd.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(mvnCmd.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func GradleCmd(c *cli.Context) (err error) {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	resolveServer := func(args []string) ([]string, *coreConfig.ServerDetails, error) {
		cleanedArgs, serverID, err := coreutils.ExtractServerIdFromCommand(args)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to extract server ID: %w", err)
		}

		if serverID == "" {
			serverDetails, err := coreConfig.GetDefaultServerConf()
			if err != nil {
				return cleanedArgs, nil, err
			}
			if serverDetails == nil {
				return cleanedArgs, nil, fmt.Errorf("no default server configuration found. Please configure a server using 'jfrog config add' or specify a server using --server-id")
			}
			return cleanedArgs, serverDetails, nil
		}

		serverDetails, err := coreConfig.GetSpecificConfig(serverID, true, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get server configuration for ID '%s': %w", serverID, err)
		}
		return cleanedArgs, serverDetails, nil
	}

	configFilePath, configExists, err := project.GetProjectConfFilePath(project.Gradle)
	if err != nil {
		return err
	}
	nativeMode := artutils.ShouldRunNative(configFilePath)

	// FlexPack native mode for Gradle (bypasses config file requirements)
	if nativeMode && !configExists {
		log.Debug("Routing to Gradle FlexPack implementation")
		if c.NArg() < 1 {
			return cliutils.WrongNumberOfArgumentsHandler(c)
		}
		args := cliutils.ExtractCommand(c)
		args, serverDetails, err := resolveServer(args)
		if err != nil {
			return err
		}
		filteredGradleArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(args)
		if err != nil {
			return err
		}

		// Create Gradle command with FlexPack (no config file needed)
		gradleCmd := gradle.NewGradleCommand().SetConfiguration(buildConfiguration).SetTasks(filteredGradleArgs).SetConfigPath("").SetServerDetails(serverDetails)
		return commands.ExecWithPackageManager(gradleCmd, project.Gradle.String())
	}

	// If config file is missing and not in native mode, return the standard missing-config error.
	if !configExists {
		if configFilePath, err = getProjectConfigPathOrThrow(project.Gradle, "gradle", "gradle-config"); err != nil {
			return err
		}
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
	var scanOutputFormat outputFormat.OutputFormat
	if format == "" {
		scanOutputFormat = outputFormat.Table
	} else {
		scanOutputFormat, err = outputFormat.ParseOutputFormat(format, outputFormat.All)
		if err != nil {
			return err
		}
	}
	printDeploymentView := log.IsStdErrTerminal()
	gradleCmd := gradle.NewGradleCommand().SetConfiguration(buildConfiguration).SetTasks(filteredGradleArgs).SetConfigPath(configFilePath).SetThreads(threads).SetDetailedSummary(detailedSummary || printDeploymentView).SetXrayScan(xrayScan).SetScanOutputFormat(scanOutputFormat)
	err = commands.ExecWithPackageManager(gradleCmd, project.Gradle.String())
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
		// In native/FlexPack mode, a missing config file is expected —
		// the yarn command resolves server details internally.
		if !artutils.ShouldRunNative("") {
			return err
		}
	}

	yarnCmd := yarn.NewYarnCommand().SetConfigFilePath(configFilePath).SetArgs(c.Args())
	return commands.ExecWithPackageManager(yarnCmd, project.Yarn.String())
}

func pnpmCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	args := cliutils.ExtractCommand(c)
	cmdName, filteredArgs := getCommandName(args)

	// Extract JFrog-specific flags (--build-name, --build-number, --project, --module, --server-id)
	// once, so both supported commands and native pass-through use cleaned args.
	serverDetails, cleanArgs, buildConfiguration, err := extractPnpmOptionsFromArgs(filteredArgs)
	if err != nil {
		return err
	}

	switch cmdName {
	case "install", "i", "publish":
		pnpmCommand, err := pnpm.NewCommand(cmdName, cleanArgs, buildConfiguration, serverDetails)
		if err != nil {
			return err
		}
		return commands.ExecWithPackageManager(pnpmCommand, project.Pnpm.String())
	default:
		return runNativePackageManagerCmd("pnpm", append([]string{cmdName}, cleanArgs...))
	}
}

// runNativePackageManagerCmd runs a package manager command directly, passing through stdio.
func runNativePackageManagerCmd(binary string, args []string) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// extractPnpmOptionsFromArgs extracts all JFrog CLI options from pnpm command args.
// Returns server details, cleaned args (with JFrog flags removed), and build configuration.
func extractPnpmOptionsFromArgs(args []string) (serverDetails *coreConfig.ServerDetails, cleanArgs []string, buildConfig *build.BuildConfiguration, err error) {
	cleanArgs = append([]string(nil), args...)
	var serverID string
	cleanArgs, serverID, err = coreutils.ExtractServerIdFromCommand(cleanArgs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to extract server ID: %w", err)
	}
	serverDetails, err = coreConfig.GetSpecificConfig(serverID, true, true)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get server configuration for ID '%s': %w", serverID, err)
	}
	cleanArgs, buildConfig, err = build.ExtractBuildDetailsFromArgs(cleanArgs)
	if err != nil {
		return nil, nil, nil, err
	}
	return serverDetails, cleanArgs, buildConfig, nil
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
	return commands.ExecWithPackageManager(nugetCmd, project.Nuget.String())
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
	return commands.ExecWithPackageManager(dotnetCmd, project.Dotnet.String())
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
	return commands.ExecWithPackageManager(goCommand, project.Go.String())
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
	err = commands.ExecWithPackageManager(goPublishCmd, project.Go.String())
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

// containerManagerEnvVar lets users force the container manager used by 'jf docker'
// subcommands, bypassing auto-detection. Accepted values (case-insensitive): "docker", "podman".
const containerManagerEnvVar = "JFROG_CLI_CONTAINER_MANAGER"

// podmanDetector is indirected through a package-level variable so tests can
// replace the real 'docker version' probe with a deterministic stub.
var podmanDetector = dockerIsPodman

// resolveContainerManagerType returns the container manager to use when running 'jf docker' subcommands.
//
// Resolution order:
//  1. Explicit override via the JFROG_CLI_CONTAINER_MANAGER env var ("docker" or "podman").
//  2. Auto-detection: if the local 'docker' binary reports Podman in its version output
//     (i.e. the podman-docker shim or native podman aliased as docker), treat it as Podman
//     so 'jf docker ...' works transparently for Podman users without daemon-socket access.
//  3. Default: Docker.
//
// Detection is intentionally conservative: only a positive "Podman" signal from 'docker version'
// switches behavior. Real Docker installations are unaffected.
func resolveContainerManagerType() containerutils.ContainerManagerType {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(containerManagerEnvVar))) {
	case "podman":
		log.Debug(containerManagerEnvVar + "=podman. Routing 'jf docker' subcommands through Podman.")
		return containerutils.Podman
	case "docker":
		log.Debug(containerManagerEnvVar + "=docker. Routing 'jf docker' subcommands through Docker.")
		return containerutils.DockerClient
	}
	if podmanDetector() {
		log.Debug("Detected Podman-backed 'docker' CLI. Routing 'jf docker' subcommands through Podman.")
		return containerutils.Podman
	}
	return containerutils.DockerClient
}

// dockerIsPodman returns true if the local 'docker' binary is actually Podman
// (either via the podman-docker shim or an alias). Any error or missing binary returns false.
func dockerIsPodman() bool {
	cmd := exec.Command("docker", "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "podman")
}

func dockerCmd(c *cli.Context) error {
	args := cliutils.ExtractCommand(c)
	var cmd, cmdArg string
	// We may have prior flags before push/pull commands for the docker client.
	// cmdArg is the second non-flag argument: image name for pull/push/scan, subcommand for buildx
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			if cmd == "" {
				cmd = arg
			} else {
				cmdArg = arg
				break
			}
		}
	}
	var err error
	switch cmd {
	case "pull":
		err = pullCmd(c, cmdArg)
	case "push":
		err = pushCmd(c, cmdArg)
	case "login":
		err = loginCmd(c, true)
	case "scan":
		return dockerScanCmd(c, cmdArg)
	case "build":
		err = buildCmd(c)
	case "buildx":
		// Only intercept "buildx build", pass through other buildx subcommands (create, ls, rm, etc.)
		if cmdArg == "build" {
			err = buildCmd(c)
		} else {
			err = dockerNativeCmd(c)
		}
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
	PullCommand := container.NewPullCommand(resolveContainerManagerType())
	PullCommand.SetCmdParams(filteredDockerArgs).SetSkipLogin(skipLogin).SetImageTag(image).SetServerDetails(rtDetails).SetBuildConfiguration(buildConfiguration)
	supported, err := PullCommand.IsGetRepoSupported()
	if err != nil {
		return err
	}
	if !supported {
		return cliutils.NotSupportedNativeDockerCommand("docker-pull")
	}
	return commands.ExecWithPackageManager(PullCommand, project.Docker.String())
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
	pushCommand := container.NewPushCommand(resolveContainerManagerType())
	pushCommand.SetThreads(threads).SetDetailedSummary(detailedSummary || printDeploymentView).SetCmdParams(filteredDockerArgs).SetSkipLogin(skipLogin).SetBuildConfiguration(buildConfiguration).SetServerDetails(rtDetails).SetValidateSha(validateSha).SetImageTag(image)
	supported, err := pushCommand.IsGetRepoSupported()
	if err != nil {
		return err
	}
	if !supported {
		return cliutils.NotSupportedNativeDockerCommand("docker-push")
	}
	err = commands.ExecWithPackageManager(pushCommand, project.Docker.String())
	result := pushCommand.Result()
	defer cliutils.CleanupResult(result, &err)
	err = cliutils.PrintCommandSummary(pushCommand.Result(), detailedSummary, printDeploymentView, false, err)
	return
}

func buildCmd(c *cli.Context) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "dockerbuildhelp"); show || err != nil {
		return err
	}

	// Extract build configuration and arguments
	_, rtDetails, _, skipLogin, _, cleanArgs, buildConfiguration, err := extractDockerOptionsFromArgs(c.Args())
	if err != nil {
		return err
	}
	pushOption, dockerFilePath, imageTag, err := extractDockerBuildOptionsFromArgs(cleanArgs)
	if err != nil {
		return err
	}

	if !skipLogin {
		err = loginCmd(c, false)
		if err != nil {
			return err
		}
	}

	dockerOptions := strategies.DockerBuildOptions{
		DockerFilePath: dockerFilePath,
		PushExpected:   pushOption,
		ImageTag:       imageTag,
	}

	buildCommand := container.NewBuildCommand(cleanArgs).SetDockerBuildOptions(dockerOptions).SetBuildConfiguration(buildConfiguration)
	buildCommand.SetServerDetails(rtDetails)

	return commands.ExecWithPackageManager(buildCommand, project.Docker.String())
}

func loginCmd(c *cli.Context, reportMetrics bool) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "dockerloginhelp"); show || err != nil {
		return err
	}

	// extract all options
	_, rtDetails, _, _, _, _, _, err := extractDockerOptionsFromArgs(c.Args())
	if err != nil {
		return err
	}

	// Report usage metrics only when invoked as a standalone command.
	// When called from buildCmd, metrics are already reported for the build itself.
	if reportMetrics {
		metricCmdName := "rt_docker_login"
		commands.SetPackageManagerContext(project.Docker.String())
		commands.CollectMetrics(metricCmdName, nil)
		defer func() {
			if rtDetails != nil {
				commands.ReportUsage(metricCmdName, rtDetails, nil)
			}
		}()
	}

	// extract login specific options
	user, password, err := extractDockerLoginOptionsFromArgs(c.Args())
	if err != nil {
		return err
	}

	// check if registry is provided by user then use that
	// else use the default from the server details
	// below code checks if the arg after login is not a flag and considers that as the image registry
	var registry string
	for i, arg := range c.Args() {
		if arg == "login" {
			if len(c.Args()) > i+1 && !strings.HasPrefix(c.Args()[i+1], "-") {
				registry = c.Args()[i+1]
				break
			}
			break
		}
	}

	// check if username and password are provided by user then use those to login
	if user != "" && password != "" {
		// registry is mandatory when using username and password
		if registry == "" {
			return errors.New("you need to specify a registry for login using username and password")
		}
		cmd := exec.Command("docker", "login", registry, "-u", user, "-p", password)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return errorutils.CheckErrorf("%s, %s", output, err)
		}
		log.Info(string(output))
		return nil
	}

	// if registry is not provided use the default from the server details
	if registry == "" {
		registry = rtDetails.GetUrl()
	}

	loginCommand := container.NewContainerManagerCommand(containerutils.DockerClient)
	loginCommand.SetPrintConsoleError(true)
	loginCommand.SetServerDetails(rtDetails).SetLoginRegistry(registry)
	// Perform login
	if err := loginCommand.PerformLogin(rtDetails, containerutils.DockerClient); err != nil {
		return err
	}
	// here docker itself returns the login success message, so no need to print it again
	return nil
}

func huggingFaceUploadCmd(c *cli.Context) error {
	if c.NArg() < 2 {
		return cliutils.PrintHelpAndReturnError("Folder path and repository ID are required.", c)
	}
	folderPath := c.Args().Get(0)
	if folderPath == "" {
		return cliutils.PrintHelpAndReturnError("Folder path cannot be empty.", c)
	}
	absPath, err := filepath.Abs(folderPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	folderPath = absPath
	if err = validateFolderHasUploadableFiles(folderPath); err != nil {
		return err
	}
	repoID := c.Args().Get(1)
	if repoID == "" {
		return cliutils.PrintHelpAndReturnError("Repository ID cannot be empty.", c)
	}
	serverDetails, err := getHuggingFaceServerDetails(c)
	if err != nil {
		return err
	}
	err = updateHuggingFaceEnv(c, serverDetails)
	if err != nil {
		return err
	}
	buildConfiguration, err := cliutils.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	revision := c.String("revision")
	if revision == "" {
		revision = "main"
	}
	repoType := c.String("repo-type")
	if repoType == "" {
		repoType = "model"
	}
	if repoType != "model" && repoType != "dataset" {
		return fmt.Errorf("wrong repo type provided, allowed repo-type are : model and dataset")
	}
	cmd := huggingfaceCommands.NewHuggingFaceUpload().
		SetCommandName("upload").
		SetFolderPath(folderPath).
		SetRepoId(repoID).
		SetRepoType(repoType).
		SetRevision(revision).
		SetServerDetails(serverDetails).
		SetBuildConfiguration(buildConfiguration)
	return commands.ExecWithPackageManager(cmd, "huggingface")
}

func huggingFaceDownloadCmd(c *cli.Context) error {
	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Model/Dataset name is required.", c)
	}
	const defaultETagTimeout = 86400
	repoID := c.Args().Get(0)
	if repoID == "" {
		return cliutils.PrintHelpAndReturnError("Model/Dataset name cannot be empty.", c)
	}
	serverDetails, err := getHuggingFaceServerDetails(c)
	if err != nil {
		return err
	}
	err = updateHuggingFaceEnv(c, serverDetails)
	if err != nil {
		return err
	}
	buildConfiguration, err := cliutils.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	etagTimeout := defaultETagTimeout
	if c.String("etag-timeout") != "" {
		etagTimeout, err = strconv.Atoi(c.String("etag-timeout"))
		if err != nil {
			return errorutils.CheckErrorf("invalid etag-timeout value: %s", c.String("etag-timeout"))
		}
	}
	repoType := c.String("repo-type")
	if repoType == "" {
		repoType = "model"
	}
	if repoType != "model" && repoType != "dataset" {
		return fmt.Errorf("wrong repo type provided, allowed repo-type are : model and dataset")
	}
	revision := c.String("revision")
	if revision == "" {
		revision = "main"
	}
	cmd := huggingfaceCommands.NewHuggingFaceDownload().
		SetCommandName("download").
		SetRepoId(repoID).
		SetRepoType(repoType).
		SetRevision(revision).
		SetEtagTimeout(etagTimeout).
		SetServerDetails(serverDetails).
		SetBuildConfiguration(buildConfiguration)
	return commands.ExecWithPackageManager(cmd, "huggingface")
}

// validateFolderHasUploadableFiles walks the folder recursively and returns an error
// if no visible (non-hidden) regular files are found. Hidden entries — anything whose
// name starts with '.' (e.g. .git, .DS_Store) — are skipped entirely.
func validateFolderHasUploadableFiles(folderPath string) error {
	found := false
	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		// Skip hidden files and directories (e.g. .git, .DS_Store).
		if len(name) > 0 && name[0] == '.' {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// A visible regular file counts as uploadable content.
		if !d.IsDir() {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to read folder '%s': %w", folderPath, err)
	}
	if !found {
		return fmt.Errorf("folder '%s' contains no uploadable files (only hidden files or empty directories were found)", folderPath)
	}
	return nil
}

func getHuggingFaceServerDetails(c *cli.Context) (*coreConfig.ServerDetails, error) {
	serverID := c.String("server-id")
	if serverID == "" {
		serverDetails, err := coreConfig.GetDefaultServerConf()
		if err != nil {
			return nil, err
		}
		if serverDetails == nil {
			return nil, fmt.Errorf("no default server configuration found. Please configure a server using 'jfrog config add' or specify a server using --server-id")
		}
		return serverDetails, nil
	}
	serverDetails, err := coreConfig.GetSpecificConfig(serverID, true, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get server configuration for ID '%s': %w", serverID, err)
	}
	return serverDetails, nil
}

func updateHuggingFaceEnv(c *cli.Context, serverDetails *coreConfig.ServerDetails) error {
	repoKey := c.String("repo-key")
	if repoKey != "" {
		hfEndpoint := serverDetails.GetArtifactoryUrl() + huggingfaceAPI + "/" + repoKey
		err := os.Setenv(HF_ENDPOINT, hfEndpoint)
		if err != nil {
			return err
		}
	}
	accessToken := serverDetails.GetAccessToken()
	if accessToken == "" {
		return cliutils.PrintHelpAndReturnError("Access token is expired or missing, please either use rt ping command or update access token.", c)
	}
	err := os.Setenv(HF_TOKEN, accessToken)
	if err != nil {
		return err
	}
	etagTimeout := c.Int("hf-hub-etag-timeout")
	if etagTimeout == 0 {
		etagTimeout = 86400
	}
	err = os.Setenv(HF_HUB_ETAG_TIMEOUT, strconv.Itoa(etagTimeout))
	if err != nil {
		return err
	}
	downloadTimeout := c.Int("hf-hub-download-timeout")
	if downloadTimeout == 0 {
		downloadTimeout = 86400
	}
	err = os.Setenv(HF_HUB_DOWNLOAD_TIMEOUT, strconv.Itoa(downloadTimeout))
	if err != nil {
		return err
	}
	return nil
}

func dockerScanCmd(c *cli.Context, imageTag string) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), securityCLI.DockerScanCmdHiddenName); show || err != nil {
		return err
	}
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
	cm := containerutils.NewManager(resolveContainerManagerType())
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

func extractDockerLoginOptionsFromArgs(args []string) (user, password string, err error) {
	_, _, user, err = coreutils.FindFlag("-u", args)
	if err != nil {
		return
	}
	if user == "" {
		_, user, err = coreutils.ExtractStringOptionFromArgs(args, "username")
		if err != nil {
			return
		}
	}

	_, _, password, err = coreutils.FindFlag("-p", args)
	if err != nil {
		return
	}
	if password == "" {
		_, password, err = coreutils.ExtractStringOptionFromArgs(args, "password")
		if err != nil {
			return
		}
	}
	return
}

func extractDockerBuildOptionsFromArgs(args []string) (pushOption bool, dockerfilePath string, imageTag string, err error) {
	// check for --push flag or output=type=registry or output=push=true flag, first is the shorthand operator of the later
	_, pushOption, err = coreutils.FindBooleanFlag("--push", args)
	if err != nil {
		return
	}
	_, _, outputOption, err := coreutils.FindFlag("--output", args)
	if err != nil {
		return
	}
	if !pushOption && outputOption != "" &&
		(strings.Contains(outputOption, "type=registry") ||
			(strings.Contains(outputOption, "push=true") && strings.Contains(outputOption, "type=image"))) {
		pushOption = true
	}

	// Check for -f or --file flag
	_, _, dockerfilePath, err = coreutils.FindFlag("-f", args)
	if err != nil || dockerfilePath == "" {
		_, _, dockerfilePath, _ = coreutils.FindFlag("--file", args)
	}
	if dockerfilePath == "" {
		// Default to Dockerfile in current directory
		dockerfilePath = "Dockerfile"
	}

	// Extract image tag from command
	_, _, imageTag, err = coreutils.FindFlag("-t", args)
	if err != nil || imageTag == "" {
		// Try --tag flag as alternative
		_, _, imageTag, _ = coreutils.FindFlag("--tag", args)
	}
	if imageTag == "" {
		err = errors.New("could not find image tag in the command arguments. Please provide an image tag using the '-t' or '--tag' flag")
	}
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

// getPrimarySourceFromToml reads pyproject.toml and returns the name of the primary source
// Returns (sourceName, isPrimary) where isPrimary indicates if source has priority='primary'
func getPrimarySourceFromToml() (string, bool) {
	type PyProjectSource struct {
		Name     string `toml:"name"`
		URL      string `toml:"url"`
		Priority string `toml:"priority"`
	}

	type PyProject struct {
		Tool struct {
			Poetry struct {
				Source []PyProjectSource `toml:"source"`
			} `toml:"poetry"`
		} `toml:"tool"`
	}

	tomlPath := filepath.Join(".", "pyproject.toml")
	tomlData, err := os.ReadFile(tomlPath)
	if err != nil {
		return "", false
	}

	var pyproject PyProject
	if err := toml.Unmarshal(tomlData, &pyproject); err != nil {
		return "", false
	}

	// Look for primary source
	for _, source := range pyproject.Tool.Poetry.Source {
		if source.Priority == "primary" {
			return source.Name, true
		}
	}

	// If no primary, use first source
	if len(pyproject.Tool.Poetry.Source) > 0 {
		return pyproject.Tool.Poetry.Source[0].Name, false
	}

	return "", false
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

	configFilePath, args, useNative, err := GetNpmConfigAndArgs(c)
	if err != nil {
		return err
	}
	npmCmd.SetConfigFilePath(configFilePath).SetNpmArgs(args).SetUseNative(useNative)
	if err = npmCmd.Init(); err != nil {
		return err
	}
	return commands.ExecWithPackageManager(npmCmd, project.Npm.String())
}

func NpmPublishCmd(c *cli.Context) (err error) {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "npmpublishhelp"); show || err != nil {
		return err
	}

	configFilePath, args, useNative, err := GetNpmConfigAndArgs(c)
	if err != nil {
		return err
	}

	npmCmd := npm.NewNpmPublishCommand()
	npmCmd.SetConfigFilePath(configFilePath).SetArgs(args).SetUseNative(useNative)
	if err = npmCmd.Init(); err != nil {
		return err
	}
	if npmCmd.GetXrayScan() {
		commandsUtils.ConditionalUploadScanFunc = scan.ConditionalUploadDefaultScanFunc
	}
	// deployment view are not available for native npm commands
	printDeploymentView, detailedSummary := log.IsStdErrTerminal() && !npmCmd.UseNative(), npmCmd.IsDetailedSummary()
	if !detailedSummary {
		npmCmd.SetDetailedSummary(printDeploymentView)
	}
	err = commands.ExecWithPackageManager(npmCmd, project.Npm.String())
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
	return commands.ExecWithPackageManager(setupCmd, packageManager.String())
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

func GetNpmConfigAndArgs(c *cli.Context) (configFilePath string, args []string, useNative bool, err error) {
	var configExists bool
	configFilePath, configExists, err = project.GetProjectConfFilePath(project.Npm)
	if err != nil {
		return
	}
	_, args = getCommandName(c.Args())
	useNative, args, err = npm.CheckIsNativeAndFetchFilteredArgs(args)
	if err != nil {
		return
	}
	if !configExists && !useNative {
		configFilePath, err = getProjectConfigPathOrThrow(project.Npm, "npm", "npm-config")
	} else if !configExists {
		configFilePath = ""
	}
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

func UvCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	args := cliutils.ExtractCommand(c)
	filteredArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(args)
	if err != nil {
		return err
	}
	var serverID string
	filteredArgs, serverID, err = coreutils.ExtractServerIdFromCommand(filteredArgs)
	if err != nil {
		return fmt.Errorf("failed to extract server ID: %w", err)
	}
	cmdName, uvArgs := getCommandName(filteredArgs)
	// Peek at --publish-url to populate DeployerRepo for build-info enrichment.
	// The flag is NOT consumed here — it is forwarded to uv as-is.
	deployerRepo := ""
	for i, arg := range uvArgs {
		if strings.HasPrefix(arg, "--publish-url=") {
			deployerRepo = strings.TrimPrefix(arg, "--publish-url=")
		} else if arg == "--publish-url" && i+1 < len(uvArgs) {
			deployerRepo = uvArgs[i+1]
		}
	}
	uvCommand := python.NewNativeUVCommand().
		SetCommandName(cmdName).
		SetArgs(uvArgs).
		SetServerID(serverID).
		SetDeployerRepo(deployerRepo).
		SetBuildConfiguration(buildConfiguration)
	// For help requests, bypass commands.Exec() to skip the concurrent usage-reporting
	// goroutine that would otherwise make Artifactory version calls unnecessarily.
	for _, a := range uvArgs {
		if a == "-h" || a == "--help" {
			return uvCommand.Run()
		}
	}
	if cmdName == "help" || cmdName == "" {
		return uvCommand.Run()
	}
	return commands.ExecWithPackageManager(uvCommand, project.UV.String())
}

// HelmCmd executes Helm commands with build info collection support
func HelmCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	args := cliutils.ExtractCommand(c)
	cmdName, helmArgs := getCommandName(args)

	helmArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(helmArgs)
	if err != nil {
		return err
	}

	helmArgs, serverDetails, err := extractHelmServerDetails(helmArgs)
	if err != nil {
		return err
	}

	helmArgs, repositoryCachePath := extractRepositoryCacheFromArgs(helmArgs)

	restoreEnv, err := setHelmRepositoryCache(repositoryCachePath)
	if err != nil {
		return err
	}
	defer restoreEnv()

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	helmCmd := helmcmd.NewHelmCommand().
		SetHelmArgs(helmArgs).
		SetBuildConfiguration(buildConfiguration).
		SetServerDetails(serverDetails).
		SetWorkingDirectory(workingDir).
		SetHelmCmdName(cmdName)

	return commands.ExecWithPackageManager(helmCmd, project.Helm.String())
}

// extractRepositoryCacheFromArgs extracts the --repository-cache flag value from Helm command arguments
func extractRepositoryCacheFromArgs(args []string) ([]string, string) {
	cleanedArgs, repositoryCachePath, err := coreutils.ExtractStringOptionFromArgs(args, "repository-cache")
	if err != nil {
		return args, ""
	}
	return cleanedArgs, repositoryCachePath
}

// extractHelmServerDetails extracts server ID from arguments and retrieves server details.
func extractHelmServerDetails(args []string) ([]string, *coreConfig.ServerDetails, error) {
	cleanedArgs, serverID, err := coreutils.ExtractServerIdFromCommand(args)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract server ID: %w", err)
	}

	if serverID == "" {
		serverDetails, err := coreConfig.GetDefaultServerConf()
		if err != nil {
			return cleanedArgs, nil, err
		}
		if serverDetails == nil {
			return cleanedArgs, nil, fmt.Errorf("no default server configuration found. Please configure a server using 'jfrog config add' or specify a server using --server-id")
		}
		return cleanedArgs, serverDetails, nil
	}

	serverDetails, err := coreConfig.GetSpecificConfig(serverID, true, true)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get server configuration for ID '%s': %w", serverID, err)
	}

	return cleanedArgs, serverDetails, nil
}

// setHelmRepositoryCache sets or unsets HELM_REPOSITORY_CACHE environment variable.
func setHelmRepositoryCache(cachePath string) (func(), error) {
	const envVarName = "HELM_REPOSITORY_CACHE"
	originalValue := os.Getenv(envVarName)

	if cachePath != "" {
		if err := os.Setenv(envVarName, cachePath); err != nil {
			return nil, fmt.Errorf("failed to set %s environment variable: %w", envVarName, err)
		}
	} else {
		if err := os.Unsetenv(envVarName); err != nil {
			return nil, fmt.Errorf("failed to unset %s environment variable: %w", envVarName, err)
		}
	}
	restoreFunc := func() {
		if originalValue != "" {
			_ = os.Setenv(envVarName, originalValue)
		} else {
			_ = os.Unsetenv(envVarName)
		}
	}

	return restoreFunc, nil
}

func ConanCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	args := cliutils.ExtractCommand(c)

	// Extract --server-id (or fall back to default) so usage metrics can be reported.
	args, serverID, err := coreutils.ExtractServerIdFromCommand(args)
	if err != nil {
		return fmt.Errorf("failed to extract server ID: %w", err)
	}
	serverDetails, err := coreConfig.GetSpecificConfig(serverID, true, false)
	if err != nil {
		return err
	}

	// Extract build flags (--build-name, --build-number) before passing to Conan
	filteredArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(args)
	if err != nil {
		return err
	}

	cmdName, conanArgs := getCommandName(filteredArgs)

	// Use jfrog-cli-artifactory Conan command with build info support
	conanCommand := conancommand.NewConanCommand().SetCommandName(cmdName).SetArgs(conanArgs).SetBuildConfiguration(buildConfiguration).SetServerDetails(serverDetails)

	return commands.ExecWithPackageManager(conanCommand, project.Conan.String())
}

func NixCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	args := cliutils.ExtractCommand(c)

	// The first arg determines which native tool to run:
	// jf nix nix-channel --add ... → nativeTool="nix-channel", args=["--add", ...]
	// jf nix nix-env -iA ...       → nativeTool="nix-env", args=["-iA", ...]
	// jf nix nix-build '<nixpkgs>' → nativeTool="nix-build", args=["'<nixpkgs>'", ...]
	// jf nix copy --to ...         → nativeTool="copy" (runs as "nix copy"), args=["--to", ...]
	nativeTool, remainingArgs := getCommandName(args)

	// Extract --server-id flag before passing to native tool
	var serverID string
	var err error
	remainingArgs, serverID, err = coreutils.ExtractServerIdFromCommand(remainingArgs)
	if err != nil {
		return fmt.Errorf("failed to extract server ID: %w", err)
	}

	// Extract build flags (--build-name, --build-number, --module, --project)
	filteredArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(remainingArgs)
	if err != nil {
		return err
	}

	cmd := nixcommand.NewNixCommand().SetNativeTool(nativeTool).SetArgs(filteredArgs).SetBuildConfiguration(buildConfiguration)

	// Pass server details — use --server-id if provided, otherwise default
	var serverDetails *coreConfig.ServerDetails
	if serverID != "" {
		serverDetails, err = coreConfig.GetSpecificConfig(serverID, false, false)
	} else {
		serverDetails, err = coreConfig.GetDefaultServerConf()
	}
	if err == nil && serverDetails != nil {
		cmd.SetServerDetails(serverDetails)
	}

	return commands.ExecWithPackageManager(cmd, "nix")
}

func pythonCmd(c *cli.Context, projectType project.ProjectType) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// FlexPack native mode for Poetry (bypasses config file requirements)
	if artutils.ShouldRunNative("") && projectType == project.Poetry {
		log.Debug("Routing to Poetry native implementation")
		args := cliutils.ExtractCommand(c)
		// Extract --server-id so usage metrics can be reported in native mode.
		args, serverID, err := coreutils.ExtractServerIdFromCommand(args)
		if err != nil {
			return fmt.Errorf("failed to extract server ID: %w", err)
		}
		serverDetails, err := coreConfig.GetSpecificConfig(serverID, true, false)
		if err != nil {
			log.Debug("Failed to resolve server for usage reporting:", err.Error())
		}
		filteredArgs, buildConfiguration, err := build.ExtractBuildDetailsFromArgs(args)
		if err != nil {
			return err
		}
		cmdName, poetryArgs := getCommandName(filteredArgs)

		// Tag usage metric with package_manager=poetry and collect.
		// Defer reporting so metrics are sent even when poetry fails.
		metricCmdName := "rt_poetry_" + cmdName
		commands.SetPackageManagerContext(project.Poetry.String())
		commands.CollectMetrics(metricCmdName, nil)
		defer func() {
			if serverDetails != nil {
				commands.ReportUsage(metricCmdName, serverDetails, nil)
			}
		}()

		// Extract --repository flag for artifact collection (if publishing)
		deployerRepo := ""
		for i, arg := range poetryArgs {
			if strings.HasPrefix(arg, "--repository=") {
				deployerRepo = strings.TrimPrefix(arg, "--repository=")
			} else if (arg == "--repository" || arg == "-r") && i+1 < len(poetryArgs) {
				deployerRepo = poetryArgs[i+1]
			}
		}

		// Auto-add repository flag if not provided and we're publishing
		if cmdName == "publish" && deployerRepo == "" {
			// Try to get primary source from pyproject.toml (same as native Poetry behavior)
			if repoName, isPrimary := getPrimarySourceFromToml(); repoName != "" {
				if isPrimary {
					log.Info(fmt.Sprintf("No --repository flag specified. Using '%s' from pyproject.toml (priority='primary')", repoName))
				} else {
					log.Info(fmt.Sprintf("No --repository flag specified. Using '%s' from pyproject.toml (first source)", repoName))
				}
				poetryArgs = append([]string{"-r", repoName}, poetryArgs...)
				deployerRepo = repoName
			} else {
				log.Warn("No repository specified and no sources found in pyproject.toml. Poetry will attempt to publish to PyPI.")
			}
		} else if cmdName == "publish" && deployerRepo != "" {
			log.Info(fmt.Sprintf("Publishing to repository: %s (from --repository flag)", deployerRepo))
		}

		log.Info(fmt.Sprintf("Running Poetry %s.", cmdName))
		if err := runNativePackageManagerCmd("poetry", append([]string{cmdName}, poetryArgs...)); err != nil {
			return fmt.Errorf("poetry %s failed: %w", cmdName, err)
		}

		// Collect build info if build parameters provided
		if buildConfiguration != nil {
			buildName, err := buildConfiguration.GetBuildName()
			if err == nil && buildName != "" {
				workingDir, err := os.Getwd()
				if err != nil {
					log.Warn("Failed to get working directory, skipping build info collection: " + err.Error())
				} else if err := buildinfo.GetPoetryBuildInfo(workingDir, buildConfiguration, deployerRepo, cmdName, poetryArgs); err != nil {
					log.Warn("Failed to collect Poetry build info: " + err.Error())
				} else {
					buildNumber, err := buildConfiguration.GetBuildNumber()
					if err != nil {
						log.Warn("Failed to get build number: " + err.Error())
					} else {
						log.Info(fmt.Sprintf("Poetry build info collected. Use 'jf rt bp %s %s' to publish it to Artifactory.", buildName, buildNumber))
					}
				}
			}
		}

		return nil
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
		return commands.ExecWithPackageManager(pipCommand, project.Pip.String())
	case project.Pipenv:
		pipenvCommand := python.NewPipenvCommand()
		pipenvCommand.SetServerDetails(rtDetails).SetRepo(pythonConfig.TargetRepo()).SetCommandName(cmdName).SetArgs(filteredArgs)
		return commands.ExecWithPackageManager(pipenvCommand, project.Pipenv.String())
	case project.Poetry:
		poetryCommand := python.NewPoetryCommand()
		poetryCommand.SetServerDetails(rtDetails).SetRepo(pythonConfig.TargetRepo()).SetCommandName(cmdName).SetArgs(filteredArgs)
		return commands.ExecWithPackageManager(poetryCommand, project.Poetry.String())
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
		return errorutils.CheckErrorf("Terraform command: '%s' is not supported. %s", cmdName, cliutils.GetDocumentationMessage())
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
	err := commands.ExecWithPackageManager(terraformCmd, project.Terraform.String())
	result := terraformCmd.Result()
	return cliutils.PrintBriefSummaryReport(result.SuccessCount(), result.FailCount(), cliutils.IsFailNoOp(c), err)
}

func getProjectConfigPathOrThrow(projectType project.ProjectType, cmdName, configCmdName string) (configFilePath string, err error) {
	configFilePath, exists, err := project.GetProjectConfFilePath(projectType)
	if err != nil {
		return
	}
	if !exists {
		return "", errorutils.CheckErrorf("%s", getMissingConfigErrMsg(cmdName, configCmdName))
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
	return "", errorutils.CheckErrorf("%s", getMissingConfigErrMsg("twine", "pip-config OR pipenv-config"))
}

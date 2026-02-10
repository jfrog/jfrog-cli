package artifactory

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/python"
	"github.com/jfrog/jfrog-cli/docs/artifactory/cocoapodsconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/swiftconfig"
	"os"
	"strings"

	"github.com/jfrog/jfrog-cli/utils/accesstoken"

	"github.com/jfrog/gofrog/version"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferinstall"
	"github.com/jfrog/jfrog-cli/docs/artifactory/transferplugininstall"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/permissiontarget"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transfer"
	transferconfigcore "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferconfig"
	transferfilescore "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferfiles"

	transferconfigmergecore "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/transferconfigmerge"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/usersmanagement"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/buildtools"
	"github.com/jfrog/jfrog-cli/docs/artifactory/accesstokencreate"
	dotnetdocs "github.com/jfrog/jfrog-cli/docs/artifactory/dotnet"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dotnetconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gocommand"
	"github.com/jfrog/jfrog-cli/docs/artifactory/goconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gopublish"
	gradledoc "github.com/jfrog/jfrog-cli/docs/artifactory/gradle"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gradleconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupaddusers"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupcreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupdelete"
	mvndoc "github.com/jfrog/jfrog-cli/docs/artifactory/mvn"
	"github.com/jfrog/jfrog-cli/docs/artifactory/mvnconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npmci"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npmconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npminstall"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npmpublish"
	nugetdocs "github.com/jfrog/jfrog-cli/docs/artifactory/nuget"
	"github.com/jfrog/jfrog-cli/docs/artifactory/nugetconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetcreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetdelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargettemplate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetupdate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/pipconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/pipinstall"
	"github.com/jfrog/jfrog-cli/docs/artifactory/transferconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/transferconfigmerge"
	"github.com/jfrog/jfrog-cli/docs/artifactory/transferfiles"
	"github.com/jfrog/jfrog-cli/docs/artifactory/transfersettings"
	"github.com/jfrog/jfrog-cli/docs/artifactory/usercreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/userscreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/usersdelete"
	yarndocs "github.com/jfrog/jfrog-cli/docs/artifactory/yarn"
	"github.com/jfrog/jfrog-cli/docs/artifactory/yarnconfig"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jszwec/csvutil"
	"github.com/urfave/cli"
)

const (
	permCategory     = "Permission Targets"
	userCategory     = "User Management"
	transferCategory = "Transfer Between Artifactory Instances"
	releaseBundlesV2 = "release-bundles-v2"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "mvn-config",
			Hidden:       true,
			Aliases:      []string{"mvnc"},
			Flags:        cliutils.GetCommandFlags(cliutils.MvnConfig),
			Usage:        mvnconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt mvn-config", mvnconfig.GetDescription(), mvnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("mvnc", "rt", project.Maven, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "mvn",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.Mvn),
			Usage:           mvndoc.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt mvn", mvndoc.GetDescription(), mvndoc.Usage),
			UsageText:       mvndoc.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(mvndoc.EnvVar...),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("mvn", project.Maven, c, buildtools.MvnCmd)
			},
		},
		{
			Name:         "gradle-config",
			Hidden:       true,
			Aliases:      []string{"gradlec"},
			Flags:        cliutils.GetCommandFlags(cliutils.GradleConfig),
			Usage:        gradleconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt gradle-config", gradleconfig.GetDescription(), gradleconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("gradlec", "rt", project.Gradle, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "gradle",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.Gradle),
			Usage:           gradledoc.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt gradle", gradledoc.GetDescription(), gradledoc.Usage),
			UsageText:       gradledoc.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(gradledoc.EnvVar...),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("gradle", project.Gradle, c, buildtools.GradleCmd)
			},
		},
		{
			Name:         "cocoapods-config",
			Hidden:       true,
			Aliases:      []string{"cocoapodsc"},
			Flags:        cliutils.GetCommandFlags(cliutils.CocoapodsConfig),
			Usage:        gradleconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt cocoapods-config", cocoapodsconfig.GetDescription(), cocoapodsconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("cocoapodsc", "rt", project.Cocoapods, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:         "swift-config",
			Hidden:       true,
			Aliases:      []string{"swiftc"},
			Flags:        cliutils.GetCommandFlags(cliutils.SwiftConfig),
			Usage:        gradleconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt swift-config", swiftconfig.GetDescription(), swiftconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("swiftc", "rt", project.Swift, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:         "npm-config",
			Hidden:       true,
			Flags:        cliutils.GetCommandFlags(cliutils.NpmConfig),
			Aliases:      []string{"npmc"},
			Usage:        npmconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt npm-config", npmconfig.GetDescription(), npmconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("npmc", "rt", project.Npm, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "npm-install",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.NpmInstallCi),
			Aliases:         []string{"npmi"},
			Usage:           npminstall.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt npm-install", npminstall.GetDescription(), npminstall.Usage),
			UsageText:       npminstall.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("npm install", project.Npm, c, buildtools.NpmInstallCmd)
			},
		},
		{
			Name:            "npm-ci",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.NpmInstallCi),
			Aliases:         []string{"npmci"},
			Usage:           npmci.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt npm-ci", npmci.GetDescription(), npmci.Usage),
			UsageText:       npmci.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("npm ci", project.Npm, c, buildtools.NpmCiCmd)
			},
		},
		{
			Name:            "npm-publish",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.NpmPublish),
			Aliases:         []string{"npmp"},
			Usage:           npmpublish.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt npm-publish", npmpublish.GetDescription(), npmpublish.Usage),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("npm p", project.Npm, c, buildtools.NpmPublishCmd)
			},
		},
		{
			Name:         "yarn-config",
			Hidden:       true,
			Aliases:      []string{"yarnc"},
			Flags:        cliutils.GetCommandFlags(cliutils.YarnConfig),
			Usage:        yarnconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt yarn-config", yarnconfig.GetDescription(), yarnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("yarnc", "rt", project.Yarn, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "yarn",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.Yarn),
			Usage:           yarndocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt yarn", yarndocs.GetDescription(), yarndocs.Usage),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("yarn", project.Yarn, c, buildtools.YarnCmd)
			},
		},
		{
			Name:         "nuget-config",
			Hidden:       true,
			Flags:        cliutils.GetCommandFlags(cliutils.NugetConfig),
			Aliases:      []string{"nugetc"},
			Usage:        nugetconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt nuget-config", nugetconfig.GetDescription(), nugetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("nugetc", "rt", project.Nuget, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "nuget",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.Nuget),
			Usage:           nugetdocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt nuget", nugetdocs.GetDescription(), nugetdocs.Usage),
			UsageText:       nugetdocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("nuget", project.Nuget, c, buildtools.NugetCmd)
			},
		},
		{
			Name:         "dotnet-config",
			Hidden:       true,
			Flags:        cliutils.GetCommandFlags(cliutils.DotnetConfig),
			Aliases:      []string{"dotnetc"},
			Usage:        dotnetconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt dotnet-config", dotnetconfig.GetDescription(), dotnetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("dotnetc", "rt", project.Dotnet, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "dotnet",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.Dotnet),
			Usage:           dotnetdocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt dotnet", dotnetdocs.GetDescription(), dotnetdocs.Usage),
			UsageText:       dotnetdocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("dotnet", project.Dotnet, c, buildtools.DotnetCmd)
			},
		},
		{
			Name:         "go-config",
			Hidden:       true,
			Flags:        cliutils.GetCommandFlags(cliutils.GoConfig),
			Usage:        goconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt go-config", goconfig.GetDescription(), goconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("go-config", "rt", project.Go, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:         "go-publish",
			Hidden:       true,
			Flags:        cliutils.GetCommandFlags(cliutils.GoPublish),
			Aliases:      []string{"gp"},
			Usage:        gopublish.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt go-publish", gopublish.GetDescription(), gopublish.Usage),
			UsageText:    gopublish.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("gp", "rt", c, buildtools.GoPublishCmd)
			},
		},
		{
			Name:            "go",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.Go),
			Usage:           gocommand.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt go", gocommand.GetDescription(), gocommand.Usage),
			UsageText:       gocommand.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("go", project.Go, c, buildtools.GoCmd)
			},
		},
		{
			Name:         "pip-config",
			Hidden:       true,
			Flags:        cliutils.GetCommandFlags(cliutils.PipConfig),
			Aliases:      []string{"pipc"},
			Usage:        pipconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt pipc", pipconfig.GetDescription(), pipconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("pipc", "rt", project.Pip, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "pip-install",
			Hidden:          true,
			Flags:           cliutils.GetCommandFlags(cliutils.PipInstall),
			Aliases:         []string{"pipi"},
			Usage:           pipinstall.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt pipi", pipinstall.GetDescription(), pipinstall.Usage),
			UsageText:       pipinstall.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("pip install", project.Pip, c, pipDeprecatedInstallCmd)
			},
		},
		{
			Name:         "permission-target-template",
			Aliases:      []string{"ptt"},
			Usage:        permissiontargettemplate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ptt", permissiontargettemplate.GetDescription(), permissiontargettemplate.Usage),
			UsageText:    permissiontargettemplate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       permissionTargetTemplateCmd,
			Category:     permCategory,
		},
		{
			Name:         "permission-target-create",
			Aliases:      []string{"ptc"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Usage:        permissiontargetcreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ptc", permissiontargetcreate.GetDescription(), permissiontargetcreate.Usage),
			UsageText:    permissiontargetcreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       permissionTargetCreateCmd,
			Category:     permCategory,
		},
		{
			Name:         "permission-target-update",
			Aliases:      []string{"ptu"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Usage:        permissiontargetupdate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ptu", permissiontargetupdate.GetDescription(), permissiontargetupdate.Usage),
			UsageText:    permissiontargetupdate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       permissionTargetUpdateCmd,
			Category:     permCategory,
		},
		{
			Name:         "permission-target-delete",
			Aliases:      []string{"ptdel"},
			Flags:        cliutils.GetCommandFlags(cliutils.PermissionTargetDelete),
			Usage:        permissiontargetdelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ptdel", permissiontargetdelete.GetDescription(), permissiontargetdelete.Usage),
			UsageText:    permissiontargetdelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       permissionTargetDeleteCmd,
			Category:     permCategory,
		},
		{
			Name:         "user-create",
			Flags:        cliutils.GetCommandFlags(cliutils.UserCreate),
			Usage:        usercreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt user-create", usercreate.GetDescription(), usercreate.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       userCreateCmd,
			Category:     userCategory,
		},
		{
			Name:         "users-create",
			Aliases:      []string{"uc"},
			Flags:        cliutils.GetCommandFlags(cliutils.UsersCreate),
			Usage:        userscreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt uc", userscreate.GetDescription(), userscreate.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       usersCreateCmd,
			Category:     userCategory,
		},
		{
			Name:         "users-delete",
			Aliases:      []string{"udel"},
			Flags:        cliutils.GetCommandFlags(cliutils.UsersDelete),
			Usage:        usersdelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt udel", usersdelete.GetDescription(), usersdelete.Usage),
			UsageText:    usersdelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       usersDeleteCmd,
			Category:     userCategory,
		},
		{
			Name:         "group-create",
			Aliases:      []string{"gc"},
			Flags:        cliutils.GetCommandFlags(cliutils.GroupCreate),
			Usage:        groupcreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt gc", groupcreate.GetDescription(), groupcreate.Usage),
			UsageText:    groupcreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       groupCreateCmd,
			Category:     userCategory,
		},
		{
			Name:         "group-add-users",
			Aliases:      []string{"gau"},
			Flags:        cliutils.GetCommandFlags(cliutils.GroupAddUsers),
			Usage:        groupaddusers.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt gau", groupaddusers.GetDescription(), groupaddusers.Usage),
			UsageText:    groupaddusers.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       groupAddUsersCmd,
			Category:     userCategory,
		},
		{
			Name:         "group-delete",
			Aliases:      []string{"gdel"},
			Flags:        cliutils.GetCommandFlags(cliutils.GroupDelete),
			Usage:        groupdelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt gdel", groupdelete.GetDescription(), groupdelete.Usage),
			UsageText:    groupdelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       groupDeleteCmd,
			Category:     userCategory,
		},
		{
			Name:         "access-token-create",
			Hidden:       true,
			Aliases:      []string{"atc"},
			Flags:        cliutils.GetCommandFlags(cliutils.ArtifactoryAccessTokenCreate),
			Usage:        accesstokencreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt atc", accesstokencreate.GetDescription(), accesstokencreate.Usage),
			UsageText:    accesstokencreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("atc", "rt", c, artifactoryAccessTokenCreateCmd)
			},
		},
		{
			Name:         "transfer-settings",
			Usage:        transfersettings.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt transfer-settings", transfersettings.GetDescription(), transfersettings.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       transferSettingsCmd,
			Category:     transferCategory,
		},
		{
			Name:         "transfer-config",
			Flags:        cliutils.GetCommandFlags(cliutils.TransferConfig),
			Usage:        transferconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt transfer-config", transferconfig.GetDescription(), transferconfig.Usage),
			UsageText:    transferconfig.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       transferConfigCmd,
			Category:     transferCategory,
		},
		{
			Name:         "transfer-config-merge",
			Flags:        cliutils.GetCommandFlags(cliutils.TransferConfigMerge),
			Usage:        transferconfigmerge.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt transfer-config-merge", transferconfigmerge.GetDescription(), transferconfigmerge.Usage),
			UsageText:    transferconfigmerge.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       transferConfigMergeCmd,
			Category:     transferCategory,
		},
		{
			Name:         "transfer-files",
			Flags:        cliutils.GetCommandFlags(cliutils.TransferFiles),
			Usage:        transferfiles.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt transfer-files", transferfiles.GetDescription(), transferfiles.Usage),
			UsageText:    transferfiles.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       transferFilesCmd,
			Category:     transferCategory,
		},
		{
			Name:         "transfer-plugin-install",
			Flags:        cliutils.GetCommandFlags(cliutils.TransferInstall),
			Usage:        transferplugininstall.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt transfer-plugin-install", transferplugininstall.GetDescription(), transferplugininstall.Usage),
			UsageText:    transferplugininstall.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       dataTransferPluginInstallCmd,
			Category:     transferCategory,
		},
	})
}

func prepareDownloadCommand(c *cli.Context) (*spec.SpecFiles, error) {
	if c.NArg() > 0 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if c.NArg() != 1 && c.NArg() != 2 && (c.NArg() != 0 || (!c.IsSet("spec") && !c.IsSet("build") && !c.IsSet("bundle"))) {
		return nil, cliutils.WrongNumberOfArgumentsHandler(c)
	}

	var downloadSpec *spec.SpecFiles
	var err error

	if c.IsSet("spec") {
		downloadSpec, err = cliutils.GetSpec(c, true, true)
	} else {
		downloadSpec, err = createDefaultDownloadSpec(c)
	}

	if err != nil {
		return nil, err
	}

	setTransitiveInDownloadSpec(downloadSpec)
	err = spec.ValidateSpec(downloadSpec.Files, false, true)
	if err != nil {
		return nil, err
	}
	return downloadSpec, nil
}

func checkRbExistenceInV2(c *cli.Context) (bool, error) {
	bundleNameAndVersion := c.String("bundle")
	parts := strings.Split(bundleNameAndVersion, "/")
	rbName := parts[0]
	rbVersion := parts[1]

	lcDetails, err := createLifecycleDetailsByFlags(c)
	if err != nil {
		return false, err
	}

	lcServicesManager, err := utils.CreateLifecycleServiceManager(lcDetails, false)
	if err != nil {
		return false, err
	}

	return lcServicesManager.IsReleaseBundleExist(rbName, rbVersion, c.String("project"))
}

func createLifecycleDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	lcDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return nil, err
	}
	if lcDetails.Url == "" {
		return nil, errors.New("platform URL is mandatory for lifecycle commands. Configure it via the --url flag or by running 'jf config add'.")
	}
	PlatformToLifecycleUrls(lcDetails)
	return lcDetails, nil
}

func PlatformToLifecycleUrls(lcDetails *coreConfig.ServerDetails) {
	// For tests only. in prod - this "if" will always return false
	if strings.Contains(lcDetails.Url, "artifactory/") {
		lcDetails.ArtifactoryUrl = clientutils.AddTrailingSlashIfNeeded(lcDetails.Url)
		lcDetails.LifecycleUrl = strings.Replace(
			clientutils.AddTrailingSlashIfNeeded(lcDetails.Url),
			"artifactory/",
			"lifecycle/",
			1,
		)
	} else {
		lcDetails.ArtifactoryUrl = clientutils.AddTrailingSlashIfNeeded(lcDetails.Url) + "artifactory/"
		lcDetails.LifecycleUrl = clientutils.AddTrailingSlashIfNeeded(lcDetails.Url) + "lifecycle/"
	}
	lcDetails.Url = ""
}

func prepareCopyMoveCommand(c *cli.Context) (*spec.SpecFiles, error) {
	if c.NArg() > 0 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if c.NArg() != 2 && (c.NArg() != 0 || !c.IsSet("spec")) {
		return nil, cliutils.WrongNumberOfArgumentsHandler(c)
	}

	var copyMoveSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		copyMoveSpec, err = cliutils.GetSpec(c, false, true)
	} else {
		copyMoveSpec, err = createDefaultCopyMoveSpec(c)
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(copyMoveSpec.Files, true, true)
	if err != nil {
		return nil, err
	}
	return copyMoveSpec, nil
}

func prepareDeleteCommand(c *cli.Context) (*spec.SpecFiles, error) {
	if c.NArg() > 0 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if c.NArg() != 1 && (c.NArg() != 0 || (!c.IsSet("spec") && !c.IsSet("build") && !c.IsSet("bundle"))) {
		return nil, cliutils.WrongNumberOfArgumentsHandler(c)
	}

	var deleteSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		deleteSpec, err = cliutils.GetSpec(c, false, true)
	} else {
		deleteSpec, err = createDefaultDeleteSpec(c)
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(deleteSpec.Files, false, true)
	if err != nil {
		return nil, err
	}
	return deleteSpec, nil
}

func prepareSearchCommand(c *cli.Context) (*spec.SpecFiles, error) {
	if c.NArg() > 0 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if c.NArg() != 1 && (c.NArg() != 0 || (!c.IsSet("spec") && !c.IsSet("build") && !c.IsSet("bundle"))) {
		return nil, cliutils.WrongNumberOfArgumentsHandler(c)
	}

	var searchSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		searchSpec, err = cliutils.GetSpec(c, false, true)
	} else {
		searchSpec, err = createDefaultSearchSpec(c)
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(searchSpec.Files, false, true)
	if err != nil {
		return nil, err
	}
	return searchSpec, err
}

func preparePropsCmd(c *cli.Context) (*generic.PropsCommand, error) {
	if c.NArg() > 1 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("Only the 'artifact properties' argument should be sent when the spec option is used.", c)
	}
	if c.NArg() != 2 && (c.NArg() != 1 || (!c.IsSet("spec") && !c.IsSet("build") && !c.IsSet("bundle"))) {
		return nil, cliutils.WrongNumberOfArgumentsHandler(c)
	}

	var propsSpec *spec.SpecFiles
	var err error
	var props string
	if c.IsSet("spec") {
		props = c.Args()[0]
		propsSpec, err = cliutils.GetSpec(c, false, true)
	} else {
		propsSpec, err = createDefaultPropertiesSpec(c)
		if c.NArg() == 1 {
			props = c.Args()[0]
			propsSpec.Get(0).Pattern = "*"
		} else {
			props = c.Args()[1]
		}
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(propsSpec.Files, false, true)
	if err != nil {
		return nil, err
	}

	command := generic.NewPropsCommand()
	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return nil, err
	}
	threads, err := cliutils.GetThreadsCount(c)
	if err != nil {
		return nil, err
	}

	cmd := command.SetProps(props)
	cmd.SetThreads(threads).SetSpec(propsSpec).SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails)
	return cmd, nil
}

func pipDeprecatedInstallCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Get python configuration.
	pythonConfig, err := project.GetResolutionOnlyConfiguration(project.Pip)
	if err != nil {
		return fmt.Errorf("error occurred while attempting to read %[1]s-configuration file: %[2]s\n"+
			"Please run 'jf %[1]s-config' command prior to running 'jf %[1]s'", project.Pip.String(), err.Error())
	}

	// Set arg values.
	rtDetails, err := pythonConfig.ServerDetails()
	if err != nil {
		return err
	}

	pythonCommand := python.NewPipCommand()
	pythonCommand.SetServerDetails(rtDetails).SetRepo(pythonConfig.TargetRepo()).SetCommandName("install").SetArgs(cliutils.ExtractCommand(c))
	return commands.Exec(pythonCommand)
}

func permissionTargetTemplateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Run command.
	permissionTargetTemplateCmd := permissiontarget.NewPermissionTargetTemplateCommand()
	permissionTargetTemplateCmd.SetTemplatePath(c.Args().Get(0))
	return commands.Exec(permissionTargetTemplateCmd)
}

func permissionTargetCreateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Run command.
	permissionTargetCreateCmd := permissiontarget.NewPermissionTargetCreateCommand()
	permissionTargetCreateCmd.SetTemplatePath(c.Args().Get(0)).SetServerDetails(rtDetails).SetVars(c.String("vars"))
	return commands.Exec(permissionTargetCreateCmd)
}

func permissionTargetUpdateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Run command.
	permissionTargetUpdateCmd := permissiontarget.NewPermissionTargetUpdateCommand()
	permissionTargetUpdateCmd.SetTemplatePath(c.Args().Get(0)).SetServerDetails(rtDetails).SetVars(c.String("vars"))
	return commands.Exec(permissionTargetUpdateCmd)
}

func permissionTargetDeleteCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	permissionTargetDeleteCmd := permissiontarget.NewPermissionTargetDeleteCommand()
	permissionTargetDeleteCmd.SetPermissionTargetName(c.Args().Get(0)).SetServerDetails(rtDetails).SetQuiet(cliutils.GetQuietValue(c))
	return commands.Exec(permissionTargetDeleteCmd)
}

func userCreateCmd(c *cli.Context) error {
	if c.NArg() != 3 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	usersCreateCmd := usersmanagement.NewUsersCreateCommand()
	userDetails := services.User{}
	userDetails.Name = c.Args().Get(0)
	userDetails.Password = c.Args().Get(1)
	userDetails.Email = c.Args().Get(2)

	user := []services.User{userDetails}
	usersGroups := parseUsersGroupsFlag(c)
	if c.String(cliutils.Admin) != "" {
		admin := c.Bool(cliutils.Admin)
		userDetails.Admin = &admin
	}
	// Run command.
	usersCreateCmd.SetServerDetails(rtDetails).SetUsers(user).SetUsersGroups(usersGroups).SetReplaceIfExists(c.Bool(cliutils.Replace))
	return commands.Exec(usersCreateCmd)
}

func usersCreateCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	usersCreateCmd := usersmanagement.NewUsersCreateCommand()
	csvFilePath := c.String("csv")
	if csvFilePath == "" {
		return cliutils.PrintHelpAndReturnError("missing --csv <File Path>. The CSV file must include a header row with the columns: username, password, email", c)
	}
	usersList, err := parseCSVToUsersList(csvFilePath)
	if err != nil {
		return err
	}
	if len(usersList) < 1 {
		return errorutils.CheckErrorf("an empty input file was provided")
	}
	usersGroups := parseUsersGroupsFlag(c)
	// Run command.
	usersCreateCmd.SetServerDetails(rtDetails).SetUsers(usersList).SetUsersGroups(usersGroups).SetReplaceIfExists(c.Bool(cliutils.Replace))
	return commands.Exec(usersCreateCmd)
}

func parseUsersGroupsFlag(c *cli.Context) *[]string {
	if c.String(cliutils.UsersGroups) != "" {
		usersGroup := strings.Split(c.String(cliutils.UsersGroups), ",")
		return &usersGroup
	}
	return nil
}

func usersDeleteCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	usersDeleteCmd := usersmanagement.NewUsersDeleteCommand()
	var usersNamesList = make([]string, 0)
	csvFilePath := c.String("csv")
	if csvFilePath != "" {
		usersList, err := parseCSVToUsersList(csvFilePath)
		if err != nil {
			return err
		}
		// If --csv <users details file path> provided, parse and append its content to the usersNamesList to be deleted.
		usersNamesList = append(usersNamesList, usersToUsersNamesList(usersList)...)
	}
	// If <users list> provided as arg, append its content to the usersNamesList to be deleted.
	if c.NArg() > 0 {
		usersNamesList = append(usersNamesList, strings.Split(c.Args().Get(0), ",")...)
	}

	if len(usersNamesList) < 1 {
		return cliutils.PrintHelpAndReturnError("missing <users list> OR --csv <users details file path>. Example: jf rt udel 'user1,user2' or jf rt udel --csv users.csv", c)
	}

	if !cliutils.GetQuietValue(c) && !coreutils.AskYesNo("This command will delete users. Are you sure you want to continue?\n"+
		"You can avoid this confirmation message by adding --quiet to the command.", false) {
		return nil
	}

	// Run command.
	usersDeleteCmd.SetServerDetails(rtDetails).SetUsers(usersNamesList)
	return commands.Exec(usersDeleteCmd)
}

func parseCSVToUsersList(csvFilePath string) ([]services.User, error) {
	var usersList []services.User
	csvInput, err := os.ReadFile(csvFilePath)
	if err != nil {
		return usersList, errorutils.CheckError(err)
	}
	if err = csvutil.Unmarshal(csvInput, &usersList); err != nil {
		return usersList, errorutils.CheckError(err)
	}
	return usersList, nil
}

func usersToUsersNamesList(usersList []services.User) (usersNames []string) {
	for _, user := range usersList {
		if user.Name != "" {
			usersNames = append(usersNames, user.Name)
		}
	}
	return
}

func groupCreateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Run command.
	groupCreateCmd := usersmanagement.NewGroupCreateCommand()
	groupCreateCmd.SetName(c.Args().Get(0)).SetServerDetails(rtDetails).SetReplaceIfExists(c.Bool(cliutils.Replace))
	return commands.Exec(groupCreateCmd)
}

func groupAddUsersCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Run command.
	groupAddUsersCmd := usersmanagement.NewGroupUpdateCommand()
	groupAddUsersCmd.SetName(c.Args().Get(0)).SetUsers(strings.Split(c.Args().Get(1), ",")).SetServerDetails(rtDetails)
	return commands.Exec(groupAddUsersCmd)
}

func groupDeleteCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	if !cliutils.GetQuietValue(c) && !coreutils.AskYesNo("This command will delete the group. Are you sure you want to continue?\n"+
		"You can avoid this confirmation message by adding --quiet to the command.", false) {
		return nil
	}

	// Run command.
	groupDeleteCmd := usersmanagement.NewGroupDeleteCommand()
	groupDeleteCmd.SetName(c.Args().Get(0)).SetServerDetails(rtDetails)
	return commands.Exec(groupDeleteCmd)
}

func artifactoryAccessTokenCreateCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	serverDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	username := accesstoken.GetSubjectUsername(c, serverDetails)
	expiry, err := cliutils.GetIntFlagValue(c, cliutils.Expiry, cliutils.ArtifactoryTokenExpiry)
	if err != nil {
		return err
	}
	accessTokenCreateCmd := generic.NewAccessTokenCreateCommand()
	accessTokenCreateCmd.SetUserName(username).SetServerDetails(serverDetails).
		SetRefreshable(c.Bool(cliutils.Refreshable)).SetExpiry(expiry).SetGroups(c.String(cliutils.Groups)).
		SetAudience(c.String(cliutils.Audience)).SetGrantAdmin(c.Bool(cliutils.GrantAdmin))
	err = commands.Exec(accessTokenCreateCmd)
	if err != nil {
		return err
	}
	resString, err := accessTokenCreateCmd.Response()
	if err != nil {
		return err
	}
	log.Output(clientutils.IndentJson(resString))

	return nil
}

func transferConfigCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Get source Artifactory server
	sourceServerDetails, err := coreConfig.GetSpecificConfig(c.Args()[0], false, true)
	if err != nil {
		return err
	}

	// Get target artifactory server
	targetServerDetails, err := coreConfig.GetSpecificConfig(c.Args()[1], false, true)
	if err != nil {
		return err
	}

	// Run transfer config command
	transferConfigCmd := transferconfigcore.NewTransferConfigCommand(sourceServerDetails, targetServerDetails).
		SetForce(c.Bool(cliutils.Force)).SetVerbose(c.Bool(cliutils.Verbose)).SetPreChecks(c.Bool(cliutils.PreChecks)).
		SetSourceWorkingDir(c.String(cliutils.SourceWorkingDir)).
		SetTargetWorkingDir(c.String(cliutils.TargetWorkingDir))
	includeReposPatterns, excludeReposPatterns := getTransferIncludeExcludeRepos(c)
	transferConfigCmd.SetIncludeReposPatterns(includeReposPatterns)
	transferConfigCmd.SetExcludeReposPatterns(excludeReposPatterns)

	return transferConfigCmd.Run()
}

func transferConfigMergeCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Get source Artifactory server
	sourceServerDetails, err := coreConfig.GetSpecificConfig(c.Args()[0], false, true)
	if err != nil {
		return err
	}

	// Get target artifactory server
	targetServerDetails, err := coreConfig.GetSpecificConfig(c.Args()[1], false, true)
	if err != nil {
		return err
	}

	// Run transfer config command
	includeReposPatterns, excludeReposPatterns := getTransferIncludeExcludeRepos(c)
	includeProjectsPatterns, excludeProjectsPatterns := getTransferIncludeExcludeProjects(c)
	transferConfigMergeCmd := transferconfigmergecore.NewTransferConfigMergeCommand(sourceServerDetails, targetServerDetails).
		SetIncludeProjectsPatterns(includeProjectsPatterns).SetExcludeProjectsPatterns(excludeProjectsPatterns)
	transferConfigMergeCmd.SetIncludeReposPatterns(includeReposPatterns).SetExcludeReposPatterns(excludeReposPatterns)
	_, err = transferConfigMergeCmd.Run()
	return err
}

func dataTransferPluginInstallCmd(c *cli.Context) error {
	// Get the Artifactory serverID from the argument or use default if not exists
	serverID := ""
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	} else if c.NArg() == 1 {
		serverID = c.Args()[0]
	}
	serverDetails, err := coreConfig.GetSpecificConfig(serverID, true, true)
	if err != nil {
		return err
	}
	installCmd := transferinstall.NewInstallDataTransferCommand(serverDetails)
	// Optional flags
	versionToInstall := c.String(cliutils.Version)
	if versionToInstall != "" {
		installCmd.SetInstallVersion(version.NewVersion(versionToInstall))
	}
	sourceFilesPath := c.String(cliutils.InstallPluginSrcDir)
	if sourceFilesPath != "" {
		installCmd.SetLocalPluginFiles(sourceFilesPath)
	}

	if versionToInstall != "" && sourceFilesPath != "" {
		return cliutils.PrintHelpAndReturnError("Only version or dir is allowed, not both.", c)
	}

	homePath := c.String(cliutils.InstallPluginHomeDir)
	if homePath != "" {
		installCmd.SetJFrogHomePath(homePath)
	}
	return commands.Exec(installCmd)
}

func transferFilesCmd(c *cli.Context) error {
	if c.Bool(cliutils.Status) || c.Bool(cliutils.Stop) {
		newTransferFilesCmd, err := transferfilescore.NewTransferFilesCommand(nil, nil)
		if err != nil {
			return err
		}
		newTransferFilesCmd.SetStatus(c.Bool(cliutils.Status))
		newTransferFilesCmd.SetStop(c.Bool(cliutils.Stop))
		return newTransferFilesCmd.Run()
	}
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	// Get source Artifactory server
	sourceServerDetails, err := coreConfig.GetSpecificConfig(c.Args()[0], false, true)
	if err != nil {
		return err
	}

	// Get target artifactory server
	targetServerDetails, err := coreConfig.GetSpecificConfig(c.Args()[1], false, true)
	if err != nil {
		return err
	}

	// Run transfer data command
	newTransferFilesCmd, err := transferfilescore.NewTransferFilesCommand(sourceServerDetails, targetServerDetails)
	if err != nil {
		return err
	}
	newTransferFilesCmd.SetPreChecks(c.Bool(cliutils.PreChecks))
	newTransferFilesCmd.SetFilestore(c.Bool(cliutils.Filestore))
	includeReposPatterns, excludeReposPatterns := getTransferIncludeExcludeRepos(c)
	newTransferFilesCmd.SetIncludeReposPatterns(includeReposPatterns)
	newTransferFilesCmd.SetExcludeReposPatterns(excludeReposPatterns)
	newTransferFilesCmd.SetIncludeFilesPatterns(getIncludeFilesPatterns(c))
	newTransferFilesCmd.SetIgnoreState(c.Bool(cliutils.IgnoreState))
	newTransferFilesCmd.SetProxyKey(c.String(cliutils.ProxyKey))
	return newTransferFilesCmd.Run()
}

func getTransferIncludeExcludeRepos(c *cli.Context) (includeReposPatterns, excludeReposPatterns []string) {
	const patternSeparator = ";"
	if c.IsSet(cliutils.IncludeRepos) {
		includeReposPatterns = strings.Split(c.String(cliutils.IncludeRepos), patternSeparator)
	}
	if c.IsSet(cliutils.ExcludeRepos) {
		excludeReposPatterns = strings.Split(c.String(cliutils.ExcludeRepos), patternSeparator)
	}
	return
}

func getIncludeFilesPatterns(c *cli.Context) []string {
	const patternSeparator = ";"
	if c.IsSet(cliutils.IncludeFiles) {
		return strings.Split(c.String(cliutils.IncludeFiles), patternSeparator)
	}
	return nil
}

func getTransferIncludeExcludeProjects(c *cli.Context) (includeProjectsPatterns, excludeProjectsPatterns []string) {
	const patternSeparator = ";"
	if c.IsSet(cliutils.IncludeProjects) {
		includeProjectsPatterns = strings.Split(c.String(cliutils.IncludeProjects), patternSeparator)
	}
	if c.IsSet(cliutils.ExcludeProjects) {
		excludeProjectsPatterns = strings.Split(c.String(cliutils.ExcludeProjects), patternSeparator)
	}
	return
}

func transferSettingsCmd(_ *cli.Context) error {
	transferSettingsCmd := transfer.NewTransferSettingsCommand()
	return commands.Exec(transferSettingsCmd)
}

func createDefaultCopyMoveSpec(c *cli.Context) (*spec.SpecFiles, error) {
	offset, limit, err := getOffsetAndLimitValues(c)
	if err != nil {
		return nil, err
	}
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		ExcludeProps(c.String("exclude-props")).
		Build(c.String("build")).
		Project(cliutils.GetProject(c)).
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Flat(c.Bool("flat")).
		IncludeDirs(true).
		Target(c.Args().Get(1)).
		ArchiveEntries(c.String("archive-entries")).
		BuildSpec(), nil
}

func createDefaultDeleteSpec(c *cli.Context) (*spec.SpecFiles, error) {
	offset, limit, err := getOffsetAndLimitValues(c)
	if err != nil {
		return nil, err
	}
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		ExcludeProps(c.String("exclude-props")).
		Build(c.String("build")).
		Project(cliutils.GetProject(c)).
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		ArchiveEntries(c.String("archive-entries")).
		RepoOnly(c.Bool("repo-only")).
		BuildSpec(), nil
}

func createDefaultSearchSpec(c *cli.Context) (*spec.SpecFiles, error) {
	offset, limit, err := getOffsetAndLimitValues(c)
	if err != nil {
		return nil, err
	}
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		ExcludeProps(c.String("exclude-props")).
		Build(c.String("build")).
		Project(cliutils.GetProject(c)).
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		IncludeDirs(c.Bool("include-dirs")).
		ArchiveEntries(c.String("archive-entries")).
		Transitive(c.Bool("transitive")).
		Include(cliutils.GetStringsArrFlagValue(c, "include")).
		BuildSpec(), nil
}

func createDefaultPropertiesSpec(c *cli.Context) (*spec.SpecFiles, error) {
	offset, limit, err := getOffsetAndLimitValues(c)
	if err != nil {
		return nil, err
	}
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		ExcludeProps(c.String("exclude-props")).
		Build(c.String("build")).
		Project(cliutils.GetProject(c)).
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		IncludeDirs(c.Bool("include-dirs")).
		ArchiveEntries(c.String("archive-entries")).
		BuildSpec(), nil
}

func createDefaultDownloadSpec(c *cli.Context) (*spec.SpecFiles, error) {
	offset, limit, err := getOffsetAndLimitValues(c)
	if err != nil {
		return nil, err
	}

	return spec.NewBuilder().
		Pattern(getSourcePattern(c)).
		Props(c.String("props")).
		ExcludeProps(c.String("exclude-props")).
		Build(c.String("build")).
		Project(cliutils.GetProject(c)).
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		PublicGpgKey(c.String("gpg-key")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Flat(c.Bool("flat")).
		Explode(c.String("explode")).
		BypassArchiveInspection(c.Bool("bypass-archive-inspection")).
		IncludeDirs(c.Bool("include-dirs")).
		Target(c.Args().Get(1)).
		ArchiveEntries(c.String("archive-entries")).
		ValidateSymlinks(c.Bool("validate-symlinks")).
		BuildSpec(), nil
}

func getSourcePattern(c *cli.Context) string {
	var source string
	var isRbv2 bool
	var err error

	if c.IsSet("bundle") {
		// If the bundle flag is set, we need to check if the bundle exists in rbv2
		isRbv2, err = checkRbExistenceInV2(c)
		if err != nil {
			log.Error("Error occurred while checking if the bundle exists in rbv2:", err.Error())
		}
	}

	if isRbv2 {
		// RB2 will be downloaded like a regular artifact, path: projectKey-release-bundles-v2/rbName/rbVersion
		source, err = buildSourceForRbv2(c)
		if err != nil {
			log.Error("Error occurred while building source path for rbv2:", err.Error())
			return ""
		}
	} else {
		source = strings.TrimPrefix(c.Args().Get(0), "/")
	}

	return source
}

func buildSourceForRbv2(c *cli.Context) (string, error) {
	bundleNameAndVersion := c.String("bundle")
	projectKey := c.String("project")
	source := projectKey

	// Reset bundle flag
	err := c.Set("bundle", "")
	if err != nil {
		return "", err
	}

	// If projectKey is not empty, append "-" to it
	if projectKey != "" {
		source += "-"
	}
	// Build RB path: projectKey-release-bundles-v2/rbName/rbVersion/
	source += releaseBundlesV2 + "/" + bundleNameAndVersion + "/"
	return source, nil
}

func setTransitiveInDownloadSpec(downloadSpec *spec.SpecFiles) {
	transitive := os.Getenv(coreutils.TransitiveDownload)
	if transitive == "" {
		if transitive = os.Getenv(coreutils.TransitiveDownloadExperimental); transitive == "" {
			return
		}
	}
	for fileIndex := 0; fileIndex < len(downloadSpec.Files); fileIndex++ {
		downloadSpec.Files[fileIndex].Transitive = transitive
	}
}

func getOffsetAndLimitValues(c *cli.Context) (offset, limit int, err error) {
	offset, err = cliutils.GetIntFlagValue(c, "offset", 0)
	if err != nil {
		return 0, 0, err
	}
	limit, err = cliutils.GetIntFlagValue(c, "limit", 0)
	if err != nil {
		return 0, 0, err
	}

	return
}

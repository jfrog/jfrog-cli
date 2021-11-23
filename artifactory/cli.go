package artifactory

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/buildinfo"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/curl"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/oc"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/permissiontarget"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/pip"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/replication"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/repository"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/usersmanagement"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	containerutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreCommonCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/ioutils"
	"github.com/jfrog/jfrog-cli/buildtools"
	"github.com/jfrog/jfrog-cli/docs/artifactory/accesstokencreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildadddependencies"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildaddgit"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildappend"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildclean"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildcollectenv"
	"github.com/jfrog/jfrog-cli/docs/artifactory/builddiscard"
	"github.com/jfrog/jfrog-cli/docs/artifactory/builddockercreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildpromote"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildpublish"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildscan"
	copydocs "github.com/jfrog/jfrog-cli/docs/artifactory/copy"
	curldocs "github.com/jfrog/jfrog-cli/docs/artifactory/curl"
	"github.com/jfrog/jfrog-cli/docs/artifactory/delete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/deleteprops"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dockerpromote"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dockerpull"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dockerpush"
	dotnetdocs "github.com/jfrog/jfrog-cli/docs/artifactory/dotnet"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dotnetconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/download"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gitlfsclean"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gocommand"
	"github.com/jfrog/jfrog-cli/docs/artifactory/goconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gopublish"
	gradledoc "github.com/jfrog/jfrog-cli/docs/artifactory/gradle"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gradleconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupaddusers"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupcreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupdelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/move"
	mvndoc "github.com/jfrog/jfrog-cli/docs/artifactory/mvn"
	"github.com/jfrog/jfrog-cli/docs/artifactory/mvnconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npmci"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npmconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npminstall"
	"github.com/jfrog/jfrog-cli/docs/artifactory/npmpublish"
	nugetdocs "github.com/jfrog/jfrog-cli/docs/artifactory/nuget"
	"github.com/jfrog/jfrog-cli/docs/artifactory/nugetconfig"
	nugettree "github.com/jfrog/jfrog-cli/docs/artifactory/nugetdepstree"
	"github.com/jfrog/jfrog-cli/docs/artifactory/ocstartbuild"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetcreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetdelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargettemplate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetupdate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/ping"
	"github.com/jfrog/jfrog-cli/docs/artifactory/pipconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/pipinstall"
	"github.com/jfrog/jfrog-cli/docs/artifactory/podmanpull"
	"github.com/jfrog/jfrog-cli/docs/artifactory/podmanpush"
	"github.com/jfrog/jfrog-cli/docs/artifactory/replicationcreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/replicationdelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/replicationtemplate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/repocreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/repodelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/repotemplate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/repoupdate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/search"
	"github.com/jfrog/jfrog-cli/docs/artifactory/setprops"
	"github.com/jfrog/jfrog-cli/docs/artifactory/upload"
	"github.com/jfrog/jfrog-cli/docs/artifactory/usercreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/userscreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/usersdelete"
	yarndocs "github.com/jfrog/jfrog-cli/docs/artifactory/yarn"
	"github.com/jfrog/jfrog-cli/docs/artifactory/yarnconfig"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	logUtils "github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	buildinfocmd "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	ioUtils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jszwec/csvutil"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "upload",
			Flags:        cliutils.GetCommandFlags(cliutils.Upload),
			Aliases:      []string{"u"},
			Description:  upload.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt upload", upload.GetDescription(), upload.Usage),
			UsageText:    upload.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(upload.EnvVar),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return uploadCmd(c)
			},
		},
		{
			Name:         "download",
			Flags:        cliutils.GetCommandFlags(cliutils.Download),
			Aliases:      []string{"dl"},
			Description:  download.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt download", download.GetDescription(), download.Usage),
			UsageText:    download.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(download.EnvVar),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return downloadCmd(c)
			},
		},
		{
			Name:         "move",
			Flags:        cliutils.GetCommandFlags(cliutils.Move),
			Aliases:      []string{"mv"},
			Description:  move.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt move", move.GetDescription(), move.Usage),
			UsageText:    move.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return moveCmd(c)
			},
		},
		{
			Name:         "copy",
			Flags:        cliutils.GetCommandFlags(cliutils.Copy),
			Aliases:      []string{"cp"},
			Description:  copydocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt copy", copydocs.GetDescription(), copydocs.Usage),
			UsageText:    copydocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return copyCmd(c)
			},
		},
		{
			Name:         "delete",
			Flags:        cliutils.GetCommandFlags(cliutils.Delete),
			Aliases:      []string{"del"},
			Description:  delete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt delete", delete.GetDescription(), delete.Usage),
			UsageText:    delete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return deleteCmd(c)
			},
		},
		{
			Name:         "search",
			Flags:        cliutils.GetCommandFlags(cliutils.Search),
			Aliases:      []string{"s"},
			Description:  search.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt search", search.GetDescription(), search.Usage),
			UsageText:    search.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return searchCmd(c)
			},
		},
		{
			Name:         "set-props",
			Flags:        cliutils.GetCommandFlags(cliutils.Properties),
			Aliases:      []string{"sp"},
			Description:  setprops.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt set-props", setprops.GetDescription(), setprops.Usage),
			UsageText:    setprops.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return setPropsCmd(c)
			},
		},
		{
			Name:         "delete-props",
			Flags:        cliutils.GetCommandFlags(cliutils.Properties),
			Aliases:      []string{"delp"},
			Description:  deleteprops.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt delete-props", deleteprops.GetDescription(), deleteprops.Usage),
			UsageText:    deleteprops.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return deletePropsCmd(c)
			},
		},
		{
			Name:         "build-publish",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildPublish),
			Aliases:      []string{"bp"},
			Description:  buildpublish.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-publish", buildpublish.GetDescription(), buildpublish.Usage),
			UsageText:    buildpublish.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildPublishCmd(c)
			},
		},
		{
			Name:         "build-collect-env",
			Aliases:      []string{"bce"},
			Flags:        cliutils.GetCommandFlags(cliutils.BuildCollectEnv),
			Description:  buildcollectenv.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-collect-env", buildcollectenv.GetDescription(), buildcollectenv.Usage),
			UsageText:    buildcollectenv.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildCollectEnvCmd(c)
			},
		},
		{
			Name:         "build-append",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildAppend),
			Aliases:      []string{"ba"},
			Description:  buildappend.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-append", buildappend.GetDescription(), buildappend.Usage),
			UsageText:    buildappend.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildAppendCmd(c)
			},
		},
		{
			Name:         "build-add-dependencies",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildAddDependencies),
			Aliases:      []string{"bad"},
			Description:  buildadddependencies.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-add-dependencies", buildadddependencies.GetDescription(), buildadddependencies.Usage),
			UsageText:    buildadddependencies.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildAddDependenciesCmd(c)
			},
		},
		{
			Name:         "build-add-git",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildAddGit),
			Aliases:      []string{"bag"},
			Description:  buildaddgit.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-add-git", buildaddgit.GetDescription(), buildaddgit.Usage),
			UsageText:    buildaddgit.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildAddGitCmd(c)
			},
		},
		{
			Name:         "build-scan",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildScanLegacy),
			Aliases:      []string{"bs"},
			Description:  buildscan.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-scan", buildscan.GetDescription(), buildscan.Usage),
			UsageText:    buildscan.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildScanLegacyCmd(c)
			},
		},
		{
			Name:         "build-clean",
			Aliases:      []string{"bc"},
			Description:  buildclean.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-clean", buildclean.GetDescription(), buildclean.Usage),
			UsageText:    buildclean.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildCleanCmd(c)
			},
		},
		{
			Name:         "build-promote",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildPromote),
			Aliases:      []string{"bpr"},
			Description:  buildpromote.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-promote", buildpromote.GetDescription(), buildpromote.Usage),
			UsageText:    buildpromote.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildPromoteCmd(c)
			},
		},
		{
			Name:         "build-discard",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildDiscard),
			Aliases:      []string{"bdi"},
			Description:  builddiscard.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-discard", builddiscard.GetDescription(), builddiscard.Usage),
			UsageText:    builddiscard.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildDiscardCmd(c)
			},
		},
		{
			Name:         "git-lfs-clean",
			Flags:        cliutils.GetCommandFlags(cliutils.GitLfsClean),
			Aliases:      []string{"glc"},
			Description:  gitlfsclean.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt git-lfs-clean", gitlfsclean.GetDescription(), gitlfsclean.Usage),
			UsageText:    gitlfsclean.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return gitLfsCleanCmd(c)
			},
		},
		{
			Name:         "mvn-config",
			Aliases:      []string{"mvnc"},
			Flags:        cliutils.GetCommandFlags(cliutils.MvnConfig),
			Description:  mvnconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt mvn-config", mvnconfig.GetDescription(), mvnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("mvnc", utils.Maven, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "mvn",
			Flags:           cliutils.GetCommandFlags(cliutils.Mvn),
			Description:     mvndoc.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt mvn", mvndoc.GetDescription(), mvndoc.Usage),
			UsageText:       mvndoc.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(mvndoc.EnvVar),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("mvn", utils.Maven, c, buildtools.MvnCmd)
			},
		},
		{
			Name:         "gradle-config",
			Aliases:      []string{"gradlec"},
			Flags:        cliutils.GetCommandFlags(cliutils.GradleConfig),
			Description:  gradleconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt gradle-config", gradleconfig.GetDescription(), gradleconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("gradlec", utils.Gradle, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "gradle",
			Flags:           cliutils.GetCommandFlags(cliutils.Gradle),
			Description:     gradledoc.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt gradle", gradledoc.GetDescription(), gradledoc.Usage),
			UsageText:       gradledoc.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(gradledoc.EnvVar),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("gradle", utils.Gradle, c, buildtools.GradleCmd)
			},
		},
		{
			Name:         "docker-promote",
			Flags:        cliutils.GetCommandFlags(cliutils.DockerPromote),
			Aliases:      []string{"dpr"},
			Description:  dockerpromote.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt docker-promote", dockerpromote.GetDescription(), dockerpromote.Usage),
			UsageText:    dockerpromote.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return dockerPromoteCmd(c)
			},
		},
		{
			Name:         "docker-push",
			Flags:        cliutils.GetCommandFlags(cliutils.ContainerPush),
			Aliases:      []string{"dp"},
			Description:  dockerpush.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt docker-push", dockerpush.GetDescription(), dockerpush.Usage),
			UsageText:    dockerpush.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return containerPushCmd(c, containerutils.DockerClient)
			},
		},
		{
			Name:         "docker-pull",
			Flags:        cliutils.GetCommandFlags(cliutils.ContainerPull),
			Aliases:      []string{"dpl"},
			Description:  dockerpull.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt docker-pull", dockerpull.GetDescription(), dockerpull.Usage),
			UsageText:    dockerpull.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return containerPullCmd(c, containerutils.DockerClient)
			},
		},
		{
			Name:         "podman-push",
			Flags:        cliutils.GetCommandFlags(cliutils.ContainerPush),
			Aliases:      []string{"pp"},
			Description:  podmanpush.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt podman-push", podmanpush.GetDescription(), podmanpush.Usage),
			UsageText:    podmanpush.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return containerPushCmd(c, containerutils.Podman)
			},
		},
		{
			Name:         "podman-pull",
			Flags:        cliutils.GetCommandFlags(cliutils.ContainerPull),
			Aliases:      []string{"ppl"},
			Description:  podmanpull.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt podman-pull", podmanpull.GetDescription(), podmanpull.Usage),
			UsageText:    podmanpull.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return containerPullCmd(c, containerutils.Podman)
			},
		},
		{
			Name:         "build-docker-create",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildDockerCreate),
			Aliases:      []string{"bdc"},
			Description:  builddockercreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt build-docker-create", builddockercreate.GetDescription(), builddockercreate.Usage),
			UsageText:    builddockercreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return BuildDockerCreateCmd(c)
			},
		},
		{
			Name:            "oc", // Only 'oc start-build' is supported
			Flags:           cliutils.GetCommandFlags(cliutils.OcStartBuild),
			Aliases:         []string{"osb"},
			Description:     ocstartbuild.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt oc start-build", ocstartbuild.GetDescription(), ocstartbuild.Usage),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return ocStartBuildCmd(c)
			},
		},
		{
			Name:         "npm-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NpmConfig),
			Aliases:      []string{"npmc"},
			Description:  npmconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt npm-config", npmconfig.GetDescription(), npmconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("npmc", utils.Npm, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "npm-install",
			Flags:           cliutils.GetCommandFlags(cliutils.Npm),
			Aliases:         []string{"npmi"},
			Description:     npminstall.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt npm-install", npminstall.GetDescription(), npminstall.Usage),
			UsageText:       npminstall.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("npm install", utils.Npm, c, npmInstallCmd)
			},
		},
		{
			Name:            "npm-ci",
			Flags:           cliutils.GetCommandFlags(cliutils.Npm),
			Aliases:         []string{"npmci"},
			Description:     npmci.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt npm-ci", npmci.GetDescription(), npmci.Usage),
			UsageText:       npmci.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("npm ci", utils.Npm, c, npmCiCmd)
			},
		},
		{
			Name:            "npm-publish",
			Flags:           cliutils.GetCommandFlags(cliutils.NpmPublish),
			Aliases:         []string{"npmp"},
			Description:     npmpublish.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt npm-publish", npmpublish.GetDescription(), npmpublish.Usage),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return npmPublishCmd(c)
			},
		},
		{
			Name:         "yarn-config",
			Aliases:      []string{"yarnc"},
			Flags:        cliutils.GetCommandFlags(cliutils.YarnConfig),
			Description:  yarnconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt yarn-config", yarnconfig.GetDescription(), yarnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("yarnc", utils.Yarn, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "yarn",
			Flags:           cliutils.GetCommandFlags(cliutils.Yarn),
			Description:     yarndocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt yarn", yarndocs.GetDescription(), yarndocs.Usage),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("yarn", utils.Yarn, c, buildtools.YarnCmd)
			},
		},
		{
			Name:         "nuget-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NugetConfig),
			Aliases:      []string{"nugetc"},
			Description:  nugetconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt nuget-config", nugetconfig.GetDescription(), nugetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("nugetc", utils.Nuget, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "nuget",
			Flags:           cliutils.GetCommandFlags(cliutils.Nuget),
			Description:     nugetdocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt nuget", nugetdocs.GetDescription(), nugetdocs.Usage),
			UsageText:       nugetdocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("nuget", utils.Nuget, c, buildtools.NugetCmd)
			},
		},
		{
			Name:         "nuget-deps-tree",
			Aliases:      []string{"ndt"},
			Description:  nugettree.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt nuget-deps-tree", nugettree.GetDescription(), nugettree.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return nugetDepsTreeCmd(c)
			},
		},
		{
			Name:         "dotnet-config",
			Flags:        cliutils.GetCommandFlags(cliutils.DotnetConfig),
			Aliases:      []string{"dotnetc"},
			Description:  dotnetconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt dotnet-config", dotnetconfig.GetDescription(), dotnetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("dotnetc", utils.Dotnet, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "dotnet",
			Flags:           cliutils.GetCommandFlags(cliutils.Dotnet),
			Description:     dotnetdocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt dotnet", dotnetdocs.GetDescription(), dotnetdocs.Usage),
			UsageText:       dotnetdocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("dotnet", utils.Dotnet, c, buildtools.DotnetCmd)
			},
		},
		{
			Name:         "go-config",
			Flags:        cliutils.GetCommandFlags(cliutils.GoConfig),
			Description:  goconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt go-config", goconfig.GetDescription(), goconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("go-config", utils.Go, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:         "go-publish",
			Flags:        cliutils.GetCommandFlags(cliutils.GoPublish),
			Aliases:      []string{"gp"},
			Description:  gopublish.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt go-publish", gopublish.GetDescription(), gopublish.Usage),
			UsageText:    gopublish.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("gp", c, buildtools.GoPublishCmd)
			},
		},
		{
			Name:            "go",
			Flags:           cliutils.GetCommandFlags(cliutils.Go),
			Aliases:         []string{"go"},
			Description:     gocommand.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt go", gocommand.GetDescription(), gocommand.Usage),
			UsageText:       gocommand.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("go", utils.Go, c, buildtools.GoCmd)
			},
		},
		{
			Name:         "ping",
			Flags:        cliutils.GetCommandFlags(cliutils.Ping),
			Aliases:      []string{"p"},
			Description:  ping.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ping", ping.GetDescription(), ping.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return pingCmd(c)
			},
		},
		{
			Name:            "curl",
			Flags:           cliutils.GetCommandFlags(cliutils.RtCurl),
			Aliases:         []string{"cl"},
			Description:     curldocs.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt curl", curldocs.GetDescription(), curldocs.Usage),
			UsageText:       curldocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			SkipFlagParsing: true,
			Action: func(c *cli.Context) error {
				return curlCmd(c)
			},
		},
		{
			Name:         "pip-config",
			Flags:        cliutils.GetCommandFlags(cliutils.PipConfig),
			Aliases:      []string{"pipc"},
			Description:  pipconfig.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt pipc", pipconfig.GetDescription(), pipconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunConfigCmdWithDeprecationWarning("pipc", utils.Pip, c, cliutils.CreateConfigCmd)
			},
		},
		{
			Name:            "pip-install",
			Flags:           cliutils.GetCommandFlags(cliutils.PipInstall),
			Aliases:         []string{"pipi"},
			Description:     pipinstall.GetDescription(),
			HelpName:        corecommon.CreateUsage("rt pipi", pipinstall.GetDescription(), pipinstall.Usage),
			UsageText:       pipinstall.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunNativeCmdWithDeprecationWarning("pip install", utils.Pip, c, pipDeprecatedInstallCmd)
			},
		},
		{
			Name:         "repo-template",
			Aliases:      []string{"rpt"},
			Description:  repotemplate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt rpt", repotemplate.GetDescription(), repotemplate.Usage),
			UsageText:    repotemplate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return repoTemplateCmd(c)
			},
		},
		{
			Name:         "repo-create",
			Aliases:      []string{"rc"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Description:  repocreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt rc", repocreate.GetDescription(), repocreate.Usage),
			UsageText:    repocreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return repoCreateCmd(c)
			},
		},
		{
			Name:         "repo-update",
			Aliases:      []string{"ru"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Description:  repoupdate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ru", repoupdate.GetDescription(), repoupdate.Usage),
			UsageText:    repoupdate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return repoUpdateCmd(c)
			},
		},
		{
			Name:         "repo-delete",
			Aliases:      []string{"rdel"},
			Flags:        cliutils.GetCommandFlags(cliutils.RepoDelete),
			Description:  repodelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt rdel", repodelete.GetDescription(), repodelete.Usage),
			UsageText:    repodelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return repoDeleteCmd(c)
			},
		},
		{
			Name:         "replication-template",
			Aliases:      []string{"rplt"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Description:  replicationtemplate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt rplt", replicationtemplate.GetDescription(), replicationtemplate.Usage),
			UsageText:    replicationtemplate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return replicationTemplateCmd(c)
			},
		},
		{
			Name:         "replication-create",
			Aliases:      []string{"rplc"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Description:  replicationcreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt rplc", replicationcreate.GetDescription(), replicationcreate.Usage),
			UsageText:    replicationcreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return replicationCreateCmd(c)
			},
		},
		{
			Name:         "replication-delete",
			Aliases:      []string{"rpldel"},
			Flags:        cliutils.GetCommandFlags(cliutils.ReplicationDelete),
			Description:  replicationdelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt rpldel", replicationdelete.GetDescription(), replicationdelete.Usage),
			UsageText:    replicationdelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return replicationDeleteCmd(c)
			},
		},
		{
			Name:         "permission-target-template",
			Aliases:      []string{"ptt"},
			Description:  permissiontargettemplate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ptt", permissiontargettemplate.GetDescription(), permissiontargettemplate.Usage),
			UsageText:    permissiontargettemplate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return permissionTargetTemplateCmd(c)
			},
		},
		{
			Name:         "permission-target-create",
			Aliases:      []string{"ptc"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Description:  permissiontargetcreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ptc", permissiontargetcreate.GetDescription(), permissiontargetcreate.Usage),
			UsageText:    permissiontargetcreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return permissionTargetCreateCmd(c)
			},
		},
		{
			Name:         "permission-target-update",
			Aliases:      []string{"ptu"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Description:  permissiontargetupdate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ptu", permissiontargetupdate.GetDescription(), permissiontargetupdate.Usage),
			UsageText:    permissiontargetupdate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return permissionTargetUpdateCmd(c)
			},
		},
		{
			Name:         "permission-target-delete",
			Aliases:      []string{"ptdel"},
			Flags:        cliutils.GetCommandFlags(cliutils.PermissionTargetDelete),
			Description:  permissiontargetdelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt ptdel", permissiontargetdelete.GetDescription(), permissiontargetdelete.Usage),
			UsageText:    permissiontargetdelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return permissionTargetDeleteCmd(c)
			},
		},
		{
			Name:         "user-create",
			Flags:        cliutils.GetCommandFlags(cliutils.UserCreate),
			Description:  usercreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt user-create", usercreate.GetDescription(), usercreate.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return userCreateCmd(c)
			},
		},
		{
			Name:         "users-create",
			Aliases:      []string{"uc"},
			Flags:        cliutils.GetCommandFlags(cliutils.UsersCreate),
			Description:  userscreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt uc", userscreate.GetDescription(), userscreate.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return usersCreateCmd(c)
			},
		},
		{
			Name:         "users-delete",
			Aliases:      []string{"udel"},
			Flags:        cliutils.GetCommandFlags(cliutils.UsersDelete),
			Description:  usersdelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt udel", usersdelete.GetDescription(), usersdelete.Usage),
			UsageText:    usersdelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return usersDeleteCmd(c)
			},
		},
		{
			Name:         "group-create",
			Aliases:      []string{"gc"},
			Flags:        cliutils.GetCommandFlags(cliutils.GroupCreate),
			Description:  groupcreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt gc", groupcreate.GetDescription(), groupcreate.Usage),
			UsageText:    groupcreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return groupCreateCmd(c)
			},
		},
		{
			Name:         "group-add-users",
			Aliases:      []string{"gau"},
			Flags:        cliutils.GetCommandFlags(cliutils.GroupAddUsers),
			Description:  groupaddusers.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt gau", groupaddusers.GetDescription(), groupaddusers.Usage),
			UsageText:    groupaddusers.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return groupAddUsersCmd(c)
			},
		},
		{
			Name:         "group-delete",
			Aliases:      []string{"gdel"},
			Flags:        cliutils.GetCommandFlags(cliutils.GroupDelete),
			Description:  groupdelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt gdel", groupdelete.GetDescription(), groupdelete.Usage),
			UsageText:    groupdelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return groupDeleteCmd(c)
			},
		},
		{
			Name:         "access-token-create",
			Aliases:      []string{"atc"},
			Flags:        cliutils.GetCommandFlags(cliutils.AccessTokenCreate),
			Description:  accesstokencreate.GetDescription(),
			HelpName:     corecommon.CreateUsage("rt atc", accesstokencreate.GetDescription(), accesstokencreate.Usage),
			UsageText:    accesstokencreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return accessTokenCreateCmd(c)
			},
		},
	})
}

func createArtifactoryDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	artDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, false, cliutils.Rt)
	if err != nil {
		return nil, err
	}
	if artDetails.ArtifactoryUrl == "" {
		return nil, errors.New("the --url option is mandatory")
	}
	return artDetails, nil
}

func getSplitCount(c *cli.Context) (splitCount int, err error) {
	splitCount = cliutils.DownloadSplitCount
	err = nil
	if c.String("split-count") != "" {
		splitCount, err = strconv.Atoi(c.String("split-count"))
		if err != nil {
			err = errors.New("The '--split-count' option should have a numeric value. " + cliutils.GetDocumentationMessage())
		}
		if splitCount > cliutils.DownloadMaxSplitCount {
			err = errors.New("The '--split-count' option value is limited to a maximum of " + strconv.Itoa(cliutils.DownloadMaxSplitCount) + ".")
		}
		if splitCount < 0 {
			err = errors.New("the '--split-count' option cannot have a negative value")
		}
	}
	return
}

func getMinSplit(c *cli.Context) (minSplitSize int64, err error) {
	minSplitSize = cliutils.DownloadMinSplitKb
	err = nil
	if c.String("min-split") != "" {
		minSplitSize, err = strconv.ParseInt(c.String("min-split"), 10, 64)
		if err != nil {
			err = errors.New("The '--min-split' option should have a numeric value. " + cliutils.GetDocumentationMessage())
			return 0, err
		}
	}

	return minSplitSize, nil
}

func getRetries(c *cli.Context) (retries int, err error) {
	retries = cliutils.Retries
	err = nil
	if c.String("retries") != "" {
		retries, err = strconv.Atoi(c.String("retries"))
		if err != nil {
			err = errors.New("The '--retries' option should have a numeric value. " + cliutils.GetDocumentationMessage())
			return 0, err
		}
	}

	return retries, nil
}

func dockerPromoteCmd(c *cli.Context) error {
	if c.NArg() != 3 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	artDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	params := services.NewDockerPromoteParams(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2))
	params.TargetDockerImage = c.String("target-docker-image")
	params.SourceTag = c.String("source-tag")
	params.TargetTag = c.String("target-tag")
	params.Copy = c.Bool("copy")
	dockerPromoteCommand := container.NewDockerPromoteCommand()
	dockerPromoteCommand.SetParams(params).SetServerDetails(artDetails)

	return commands.Exec(dockerPromoteCommand)
}

func containerPushCmd(c *cli.Context, containerManagerType containerutils.ContainerManagerType) error {
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	artDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	imageTag := c.Args().Get(0)
	targetRepo := c.Args().Get(1)
	skipLogin := c.Bool("skip-login")

	buildConfiguration, err := buildtools.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	dockerPushCommand := container.NewPushCommand(containerManagerType)
	threads, err := cliutils.GetThreadsCount(c)
	if err != nil {
		return err
	}
	dockerPushCommand.SetThreads(threads).SetDetailedSummary(c.Bool("detailed-summary")).SetBuildConfiguration(buildConfiguration).SetRepo(targetRepo).SetSkipLogin(skipLogin).SetServerDetails(artDetails).SetImageTag(imageTag)

	err = commands.Exec(dockerPushCommand)
	if err != nil {
		return err
	}
	if dockerPushCommand.IsDetailedSummary() {
		result := dockerPushCommand.Result()
		return cliutils.PrintDetailedSummaryReport(result.SuccessCount(), result.FailCount(), result.Reader(), true, false, err)
	}
	return nil
}

func containerPullCmd(c *cli.Context, containerManagerType containerutils.ContainerManagerType) error {
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	artDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	imageTag := c.Args().Get(0)
	sourceRepo := c.Args().Get(1)
	skipLogin := c.Bool("skip-login")
	buildConfiguration, err := buildtools.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	dockerPullCommand := container.NewPullCommand(containerManagerType)
	dockerPullCommand.SetImageTag(imageTag).SetRepo(sourceRepo).SetSkipLogin(skipLogin).SetServerDetails(artDetails).SetBuildConfiguration(buildConfiguration)

	return commands.Exec(dockerPullCommand)
}

func BuildDockerCreateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	artDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	sourceRepo := c.Args().Get(0)
	imageNameWithDigestFile := c.String("image-file")
	if imageNameWithDigestFile == "" {
		return cliutils.PrintHelpAndReturnError("The '--image-file' command option was not provided.", c)
	}
	buildConfiguration, err := buildtools.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	buildDockerCreateCommand := container.NewBuildDockerCreateCommand()
	if err := buildDockerCreateCommand.SetImageNameWithDigest(imageNameWithDigestFile); err != nil {
		return err
	}
	buildDockerCreateCommand.SetRepo(sourceRepo).SetServerDetails(artDetails).SetBuildConfiguration(buildConfiguration)
	return commands.Exec(buildDockerCreateCommand)
}

func ocStartBuildCmd(c *cli.Context) error {
	args := cliutils.ExtractCommand(c)

	// After the 'oc' command, only 'start-build' is allowed
	parentArgs := c.Parent().Args()
	if parentArgs[0] == "oc" {
		if len(parentArgs) < 2 || parentArgs[1] != "start-build" {
			return errorutils.CheckErrorf("invalid command. The only OpenShift CLI command supported by JFrog CLI is 'oc start-build'")
		}
		coreutils.RemoveFlagFromCommand(&args, 0, 0)
	}

	if show, err := cliutils.ShowCmdHelpIfNeeded(c, args); show || err != nil {
		return err
	}
	if len(args) < 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	// Extract build configuration
	filteredOcArgs, buildConfiguration, err := utils.ExtractBuildDetailsFromArgs(args)
	if err != nil {
		return err
	}

	// Extract repo
	flagIndex, valueIndex, repo, err := coreutils.FindFlag("--repo", filteredOcArgs)
	if err != nil {
		return err
	}
	coreutils.RemoveFlagFromCommand(&filteredOcArgs, flagIndex, valueIndex)
	if flagIndex == -1 {
		err = errorutils.CheckErrorf("the --repo option is mandatory")
		return err
	}

	// Extract server-id
	flagIndex, valueIndex, serverId, err := coreutils.FindFlag("--server-id", filteredOcArgs)
	if err != nil {
		return err
	}
	coreutils.RemoveFlagFromCommand(&filteredOcArgs, flagIndex, valueIndex)

	ocCmd := oc.NewOcStartBuildCommand().SetOcArgs(filteredOcArgs).SetRepo(repo).SetServerId(serverId).SetBuildConfiguration(buildConfiguration)
	return commands.Exec(ocCmd)
}

func nugetDepsTreeCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	return dotnet.DependencyTreeCmd()
}

// Deprecated
func npmInstallCmd(c *cli.Context) error {
	return npmDeprecatedInstallCiCmd(c, npm.NewNpmInstallCommand())
}

// Deprecated
func npmCiCmd(c *cli.Context) error {
	return npmDeprecatedInstallCiCmd(c, npm.NewNpmCiCommand())
}

// Deprecated
func npmDeprecatedInstallCiCmd(c *cli.Context, npmCmd *npm.NpmInstallOrCiCommand) error {
	configFilePath, args, err := buildtools.GetNpmConfigAndArgs(c)
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

func npmPublishCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Npm)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("No config file was found! Before running the npm-publish command on a project for the first time, the project should be configured using the npm-config command.\nThis configuration includes the Artifactory server and repository to which the package should deployed. ")
	}
	args := cliutils.ExtractCommand(c)
	npmCmd := npm.NewNpmPublishCommand()
	npmCmd.SetConfigFilePath(configFilePath).SetArgs(args)
	err = commands.Exec(npmCmd)
	if err != nil {
		return err
	}
	if npmCmd.IsDetailedSummary() {
		result := npmCmd.Result()
		return cliutils.PrintDetailedSummaryReport(result.SuccessCount(), result.FailCount(), result.Reader(), true, false, err)
	}
	return nil
}

func pingCmd(c *cli.Context) error {
	if c.NArg() > 0 {
		return cliutils.PrintHelpAndReturnError("No arguments should be sent.", c)
	}
	artDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	pingCmd := generic.NewPingCommand()
	pingCmd.SetServerDetails(artDetails)
	err = commands.Exec(pingCmd)
	resString := clientutils.IndentJson(pingCmd.Response())
	if err != nil {
		return errors.New(err.Error() + "\n" + resString)
	}
	log.Output(resString)

	return err
}

func prepareDownloadCommand(c *cli.Context) (*spec.SpecFiles, error) {
	if c.NArg() > 0 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || c.NArg() == 2 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build") || c.IsSet("bundle")))) {
		return nil, cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var downloadSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		downloadSpec, err = cliutils.GetSpec(c, true)
	} else {
		downloadSpec, err = createDefaultDownloadSpec(c)
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(downloadSpec.Files, false, true, false)
	if err != nil {
		return nil, err
	}
	return downloadSpec, nil
}

func downloadCmd(c *cli.Context) error {
	downloadSpec, err := prepareDownloadCommand(c)
	if err != nil {
		return err
	}
	fixWinPathsForDownloadCmd(downloadSpec, c)
	configuration, err := createDownloadConfiguration(c)
	if err != nil {
		return err
	}
	serverDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	buildConfiguration, err := buildtools.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	downloadCommand := generic.NewDownloadCommand()
	downloadCommand.SetConfiguration(configuration).SetBuildConfiguration(buildConfiguration).SetSpec(downloadSpec).SetServerDetails(serverDetails).SetDryRun(c.Bool("dry-run")).SetSyncDeletesPath(c.String("sync-deletes")).SetQuiet(cliutils.GetQuietValue(c)).SetDetailedSummary(c.Bool("detailed-summary")).SetRetries(retries)

	if downloadCommand.ShouldPrompt() && !coreutils.AskYesNo("Sync-deletes may delete some files in your local file system. Are you sure you want to continue?\n"+
		"You can avoid this confirmation message by adding --quiet to the command.", false) {
		return nil
	}

	err = execWithProgress(downloadCommand)
	result := downloadCommand.Result()
	err = cliutils.PrintDetailedSummaryReport(result.SuccessCount(), result.FailCount(), result.Reader(), false, isFailNoOp(c), err)

	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), isFailNoOp(c))
}

func uploadCmd(c *cli.Context) error {
	if c.NArg() > 0 && c.IsSet("spec") {
		return cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var uploadSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		uploadSpec, err = cliutils.GetFileSystemSpec(c)
	} else {
		uploadSpec, err = createDefaultUploadSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(uploadSpec.Files, true, false, true)
	if err != nil {
		return err
	}
	cliutils.FixWinPathsForFileSystemSourcedCmds(uploadSpec, c)
	configuration, err := createUploadConfiguration(c)
	if err != nil {
		return err
	}
	buildConfiguration, err := buildtools.CreateBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	uploadCmd := generic.NewUploadCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	uploadCmd.SetUploadConfiguration(configuration).SetBuildConfiguration(buildConfiguration).SetSpec(uploadSpec).SetServerDetails(rtDetails).SetDryRun(c.Bool("dry-run")).SetSyncDeletesPath(c.String("sync-deletes")).SetQuiet(cliutils.GetQuietValue(c)).SetDetailedSummary(c.Bool("detailed-summary")).SetRetries(retries)

	if uploadCmd.ShouldPrompt() && !coreutils.AskYesNo("Sync-deletes may delete some artifacts in Artifactory. Are you sure you want to continue?\n"+
		"You can avoid this confirmation message by adding --quiet to the command.", false) {
		return nil
	}
	err = execWithProgress(uploadCmd)
	result := uploadCmd.Result()
	err = cliutils.PrintDetailedSummaryReport(result.SuccessCount(), result.FailCount(), result.Reader(), true, isFailNoOp(c), err)

	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), isFailNoOp(c))
}

type CommandWithProgress interface {
	commands.Command
	SetProgress(ioUtils.ProgressMgr)
}

func execWithProgress(cmd CommandWithProgress) error {
	// Init progress bar.
	progressBar, logFile, err := progressbar.InitProgressBarIfPossible()
	if err != nil {
		return err
	}
	if progressBar != nil {
		cmd.SetProgress(progressBar)
		defer logUtils.CloseLogFile(logFile)
		defer progressBar.Quit()
	}
	return commands.Exec(cmd)
}

func prepareCopyMoveCommand(c *cli.Context) (*spec.SpecFiles, error) {
	if c.NArg() > 0 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && (c.IsSet("spec")))) {
		return nil, cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var copyMoveSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		copyMoveSpec, err = cliutils.GetSpec(c, false)
	} else {
		copyMoveSpec, err = createDefaultCopyMoveSpec(c)
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(copyMoveSpec.Files, true, true, false)
	if err != nil {
		return nil, err
	}
	return copyMoveSpec, nil
}

func moveCmd(c *cli.Context) error {
	moveSpec, err := prepareCopyMoveCommand(c)
	if err != nil {
		return err
	}
	moveCmd := generic.NewMoveCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	threads, err := cliutils.GetThreadsCount(c)
	if err != nil {
		return err
	}
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	moveCmd.SetThreads(threads).SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails).SetSpec(moveSpec).SetRetries(retries)
	err = commands.Exec(moveCmd)
	result := moveCmd.Result()
	return printBriefSummaryAndGetError(result.SuccessCount(), result.FailCount(), isFailNoOp(c), err)
}

func copyCmd(c *cli.Context) error {
	copySpec, err := prepareCopyMoveCommand(c)
	if err != nil {
		return err
	}

	copyCommand := generic.NewCopyCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	threads, err := cliutils.GetThreadsCount(c)
	if err != nil {
		return err
	}
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	copyCommand.SetThreads(threads).SetSpec(copySpec).SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails).SetRetries(retries)
	err = commands.Exec(copyCommand)
	result := copyCommand.Result()
	return printBriefSummaryAndGetError(result.SuccessCount(), result.FailCount(), isFailNoOp(c), err)
}

// Prints a 'brief' (not detailed) summary and returns the appropriate exit error.
func printBriefSummaryAndGetError(succeeded, failed int, failNoOp bool, originalErr error) error {
	err := cliutils.PrintBriefSummaryReport(succeeded, failed, failNoOp, originalErr)
	return cliutils.GetCliError(err, succeeded, failed, failNoOp)
}

func prepareDeleteCommand(c *cli.Context) (*spec.SpecFiles, error) {
	if c.NArg() > 0 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build") || c.IsSet("bundle")))) {
		return nil, cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var deleteSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		deleteSpec, err = cliutils.GetSpec(c, false)
	} else {
		deleteSpec, err = createDefaultDeleteSpec(c)
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(deleteSpec.Files, false, true, false)
	if err != nil {
		return nil, err
	}
	return deleteSpec, nil
}

func deleteCmd(c *cli.Context) error {
	deleteSpec, err := prepareDeleteCommand(c)
	if err != nil {
		return err
	}

	deleteCommand := generic.NewDeleteCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	threads, err := cliutils.GetThreadsCount(c)
	if err != nil {
		return err
	}
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	deleteCommand.SetThreads(threads).SetQuiet(cliutils.GetQuietValue(c)).SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails).SetSpec(deleteSpec).SetRetries(retries)
	err = commands.Exec(deleteCommand)
	result := deleteCommand.Result()
	return printBriefSummaryAndGetError(result.SuccessCount(), result.FailCount(), isFailNoOp(c), err)
}

func prepareSearchCommand(c *cli.Context) (*spec.SpecFiles, error) {
	if c.NArg() > 0 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build") || c.IsSet("bundle")))) {
		return nil, cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var searchSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		searchSpec, err = cliutils.GetSpec(c, false)
	} else {
		searchSpec, err = createDefaultSearchSpec(c)
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(searchSpec.Files, false, true, false)
	if err != nil {
		return nil, err
	}
	return searchSpec, err
}

func searchCmd(c *cli.Context) error {
	searchSpec, err := prepareSearchCommand(c)
	if err != nil {
		return err
	}
	artDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(artDetails).SetSpec(searchSpec).SetRetries(retries)
	err = commands.Exec(searchCmd)
	if err != nil {
		return err
	}
	reader := searchCmd.Result().Reader()
	defer reader.Close()
	length, err := reader.Length()
	if err != nil {
		return err
	}
	err = cliutils.GetCliError(err, length, 0, isFailNoOp(c))
	if err != nil {
		return err
	}
	if !c.Bool("count") {
		return utils.PrintSearchResults(reader)
	}
	log.Output(length)
	return nil
}

func preparePropsCmd(c *cli.Context) (*generic.PropsCommand, error) {
	if c.NArg() > 1 && c.IsSet("spec") {
		return nil, cliutils.PrintHelpAndReturnError("Only the 'artifact properties' argument should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 1 && (c.IsSet("spec") || c.IsSet("build") || c.IsSet("bundle")))) {
		return nil, cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var propsSpec *spec.SpecFiles
	var err error
	var props string
	if c.IsSet("spec") {
		props = c.Args()[0]
		propsSpec, err = cliutils.GetSpec(c, false)
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
	err = spec.ValidateSpec(propsSpec.Files, false, true, false)
	if err != nil {
		return nil, err
	}

	command := generic.NewPropsCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c)
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

func setPropsCmd(c *cli.Context) error {
	cmd, err := preparePropsCmd(c)
	if err != nil {
		return err
	}
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	propsCmd := generic.NewSetPropsCommand().SetPropsCommand(*cmd)
	propsCmd.SetRetries(retries)
	err = commands.Exec(propsCmd)
	result := propsCmd.Result()
	return printBriefSummaryAndGetError(result.SuccessCount(), result.FailCount(), isFailNoOp(c), err)
}

func deletePropsCmd(c *cli.Context) error {
	cmd, err := preparePropsCmd(c)
	if err != nil {
		return err
	}
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	propsCmd := generic.NewDeletePropsCommand().DeletePropsCommand(*cmd)
	propsCmd.SetRetries(retries)
	err = commands.Exec(propsCmd)
	result := propsCmd.Result()
	return printBriefSummaryAndGetError(result.SuccessCount(), result.FailCount(), isFailNoOp(c), err)
}

func buildPublishCmd(c *cli.Context) error {
	if c.NArg() > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
		return err
	}
	buildInfoConfiguration := createBuildInfoConfiguration(c)
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	buildPublishCmd := buildinfo.NewBuildPublishCommand().SetServerDetails(rtDetails).SetBuildConfiguration(buildConfiguration).SetConfig(buildInfoConfiguration).SetDetailedSummary(c.Bool("detailed-summary"))

	err = commands.Exec(buildPublishCmd)
	if buildPublishCmd.IsDetailedSummary() {
		if summary := buildPublishCmd.GetSummary(); summary != nil {
			return cliutils.PrintBuildInfoSummaryReport(summary.IsSucceeded(), summary.GetSha256(), err)
		}
	}
	return err
}

func buildAppendCmd(c *cli.Context) error {
	if c.NArg() != 4 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
		return err
	}
	buildNameToAppend, buildNumberToAppend := c.Args().Get(2), c.Args().Get(3)
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	buildAppendCmd := buildinfo.NewBuildAppendCommand().SetServerDetails(rtDetails).SetBuildConfiguration(buildConfiguration).SetBuildNameToAppend(buildNameToAppend).SetBuildNumberToAppend(buildNumberToAppend)
	return commands.Exec(buildAppendCmd)
}

func buildAddDependenciesCmd(c *cli.Context) error {
	if c.NArg() > 2 && c.IsSet("spec") {
		return cliutils.PrintHelpAndReturnError("Only path or spec is allowed, not both.", c)
	}
	if c.IsSet("regexp") && c.IsSet("from-rt") {
		return cliutils.PrintHelpAndReturnError("The --regexp option is not supported when --from-rt is set to true.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
		return err
	}
	// Odd number of args - Use pattern arg
	// Even number of args - Use spec flag
	if c.NArg() > 3 || !(c.NArg()%2 == 1 || (c.NArg()%2 == 0 && c.IsSet("spec"))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var dependenciesSpec *spec.SpecFiles
	var rtDetails *coreConfig.ServerDetails
	var err error
	if c.IsSet("spec") {
		dependenciesSpec, err = cliutils.GetFileSystemSpec(c)
		if err != nil {
			return err
		}
	} else {
		dependenciesSpec = createDefaultBuildAddDependenciesSpec(c)
	}
	if c.Bool("from-rt") {
		rtDetails, err = createArtifactoryDetailsByFlags(c)
		if err != nil {
			return err
		}
	} else {
		cliutils.FixWinPathsForFileSystemSourcedCmds(dependenciesSpec, c)
	}
	buildAddDependenciesCmd := buildinfo.NewBuildAddDependenciesCommand().SetDryRun(c.Bool("dry-run")).SetBuildConfiguration(buildConfiguration).SetDependenciesSpec(dependenciesSpec).SetServerDetails(rtDetails)
	err = commands.Exec(buildAddDependenciesCmd)
	result := buildAddDependenciesCmd.Result()
	return printBriefSummaryAndGetError(result.SuccessCount(), result.FailCount(), isFailNoOp(c), err)
}

func buildCollectEnvCmd(c *cli.Context) error {
	if c.NArg() > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
		return err
	}
	buildCollectEnvCmd := buildinfo.NewBuildCollectEnvCommand().SetBuildConfiguration(buildConfiguration)

	return commands.Exec(buildCollectEnvCmd)
}

func buildAddGitCmd(c *cli.Context) error {
	if c.NArg() > 3 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
		return err
	}

	buildAddGitConfigurationCmd := buildinfo.NewBuildAddGitCommand().SetBuildConfiguration(buildConfiguration).SetConfigFilePath(c.String("config")).SetServerId(c.String("server-id"))
	if c.NArg() == 3 {
		buildAddGitConfigurationCmd.SetDotGitPath(c.Args().Get(2))
	} else if c.NArg() == 1 {
		buildAddGitConfigurationCmd.SetDotGitPath(c.Args().Get(0))
	}
	return commands.Exec(buildAddGitConfigurationCmd)
}

func buildScanLegacyCmd(c *cli.Context) error {
	if c.NArg() > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
		return err
	}
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	buildScanCmd := buildinfo.NewBuildScanLegacyCommand().SetServerDetails(rtDetails).SetFailBuild(c.BoolT("fail")).SetBuildConfiguration(buildConfiguration)
	err = commands.Exec(buildScanCmd)

	return checkBuildScanError(err)
}

func checkBuildScanError(err error) error {
	// If the build was found vulnerable, exit with ExitCodeVulnerableBuild.
	if err == utils.GetBuildScanError() {
		return coreutils.CliError{ExitCode: coreutils.ExitCodeVulnerableBuild, ErrorMsg: err.Error()}
	}
	// If the scan operation failed, for example due to HTTP timeout, exit with ExitCodeError.
	if err != nil {
		return coreutils.CliError{ExitCode: coreutils.ExitCodeError, ErrorMsg: err.Error()}
	}
	return nil
}

func buildCleanCmd(c *cli.Context) error {
	if c.NArg() > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
		return err
	}
	buildCleanCmd := buildinfo.NewBuildCleanCommand().SetBuildConfiguration(buildConfiguration)

	return commands.Exec(buildCleanCmd)
}

func buildPromoteCmd(c *cli.Context) error {
	if c.NArg() > 3 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	if err := validateBuildConfiguration(c, createBuildConfiguration(c)); err != nil {
		return err
	}
	configuration := createBuildPromoteConfiguration(c)
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	buildPromotionCmd := buildinfo.NewBuildPromotionCommand().SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails).SetPromotionParams(configuration)

	return commands.Exec(buildPromotionCmd)
}

func buildDiscardCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	configuration := createBuildDiscardConfiguration(c)
	if configuration.BuildName == "" {
		return cliutils.PrintHelpAndReturnError("Build name is expected as a command argument or environment variable.", c)
	}
	buildDiscardCmd := buildinfo.NewBuildDiscardCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	buildDiscardCmd.SetServerDetails(rtDetails).SetDiscardBuildsParams(configuration)

	return commands.Exec(buildDiscardCmd)
}

func gitLfsCleanCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	configuration := createGitLfsCleanConfiguration(c)
	retries, err := getRetries(c)
	if err != nil {
		return err
	}
	gitLfsCmd := generic.NewGitLfsCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	gitLfsCmd.SetConfiguration(configuration).SetServerDetails(rtDetails).SetDryRun(c.Bool("dry-run")).SetRetries(retries)

	return commands.Exec(gitLfsCmd)
}

func curlCmd(c *cli.Context) error {
	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	rtCurlCommand, err := newRtCurlCommand(c)
	if err != nil {
		return err
	}
	return commands.Exec(rtCurlCommand)
}

func newRtCurlCommand(c *cli.Context) (*curl.RtCurlCommand, error) {
	curlCommand := coreCommonCommands.NewCurlCommand().SetArguments(cliutils.ExtractCommand(c))
	rtCurlCommand := curl.NewRtCurlCommand(*curlCommand)
	rtDetails, err := rtCurlCommand.GetServerDetails()
	if err != nil {
		return nil, err
	}
	rtCurlCommand.SetServerDetails(rtDetails)
	rtCurlCommand.SetUrl(rtDetails.ArtifactoryUrl)
	return rtCurlCommand, err
}

// Deprecated
func pipDeprecatedInstallCmd(c *cli.Context) error {
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
			"Please run 'jfrog rt pip-config' command prior to running 'jfrog rt %s'.", err.Error(), "pip-install"))
	}

	// Set arg values.
	rtDetails, err := pipConfig.ServerDetails()
	if err != nil {
		return err
	}

	// Run command.
	pipCmd := pip.NewPipInstallCommand()
	pipCmd.SetServerDetails(rtDetails).SetRepo(pipConfig.TargetRepo()).SetArgs(cliutils.ExtractCommand(c))
	return commands.Exec(pipCmd)
}

func repoTemplateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	// Run command.
	repoTemplateCmd := repository.NewRepoTemplateCommand()
	repoTemplateCmd.SetTemplatePath(c.Args().Get(0))
	return commands.Exec(repoTemplateCmd)
}

func repoCreateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Run command.
	repoCreateCmd := repository.NewRepoCreateCommand()
	repoCreateCmd.SetTemplatePath(c.Args().Get(0)).SetServerDetails(rtDetails).SetVars(c.String("vars"))
	return commands.Exec(repoCreateCmd)
}

func repoUpdateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Run command.
	repoUpdateCmd := repository.NewRepoUpdateCommand()
	repoUpdateCmd.SetTemplatePath(c.Args().Get(0)).SetServerDetails(rtDetails).SetVars(c.String("vars"))
	return commands.Exec(repoUpdateCmd)
}

func repoDeleteCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	repoDeleteCmd := repository.NewRepoDeleteCommand()
	repoDeleteCmd.SetRepoPattern(c.Args().Get(0)).SetServerDetails(rtDetails).SetQuiet(cliutils.GetQuietValue(c))
	return commands.Exec(repoDeleteCmd)
}

func replicationTemplateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	replicationTemplateCmd := replication.NewReplicationTemplateCommand()
	replicationTemplateCmd.SetTemplatePath(c.Args().Get(0))
	return commands.Exec(replicationTemplateCmd)
}

func replicationCreateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	replicationCreateCmd := replication.NewReplicationCreateCommand()
	replicationCreateCmd.SetTemplatePath(c.Args().Get(0)).SetServerDetails(rtDetails).SetVars(c.String("vars"))
	return commands.Exec(replicationCreateCmd)
}

func replicationDeleteCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	replicationDeleteCmd := replication.NewReplicationDeleteCommand()
	replicationDeleteCmd.SetRepoKey(c.Args().Get(0)).SetServerDetails(rtDetails).SetQuiet(cliutils.GetQuietValue(c))
	return commands.Exec(replicationDeleteCmd)
}

func permissionTargetTemplateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	// Run command.
	permissionTargetTemplateCmd := permissiontarget.NewPermissionTargetTemplateCommand()
	permissionTargetTemplateCmd.SetTemplatePath(c.Args().Get(0))
	return commands.Exec(permissionTargetTemplateCmd)
}

func permissionTargetCreateCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
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
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
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
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	permissionTargetDeleteCmd := permissiontarget.NewPermissionTargetDeleteCommand()
	permissionTargetDeleteCmd.SetPermissionTargetName(c.Args().Get(0)).SetServerDetails(rtDetails).SetQuiet(cliutils.GetQuietValue(c))
	return commands.Exec(permissionTargetDeleteCmd)
}

func userCreateCmd(c *cli.Context) error {
	if c.NArg() != 3 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	usersCreateCmd := usersmanagement.NewUsersCreateCommand()
	userDetails := services.User{}
	userDetails.Name = c.Args().Get(0)
	userDetails.Password = c.Args().Get(1)
	userDetails.Email = c.Args().Get(2)

	user := []services.User{userDetails}
	var usersGroups []string
	if c.String(cliutils.UsersGroups) != "" {
		usersGroups = strings.Split(c.String(cliutils.UsersGroups), ",")
	}
	if c.String(cliutils.Admin) != "" {
		userDetails.Admin = c.Bool(cliutils.Admin)
	}
	// Run command.
	usersCreateCmd.SetServerDetails(rtDetails).SetUsers(user).SetUsersGroups(usersGroups).SetReplaceIfExists(c.Bool(cliutils.Replace))
	return commands.Exec(usersCreateCmd)
}

func usersCreateCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	usersCreateCmd := usersmanagement.NewUsersCreateCommand()
	csvFilePath := c.String("csv")
	if csvFilePath == "" {
		return cliutils.PrintHelpAndReturnError("missing --csv <File Path>", c)
	}
	usersList, err := parseCSVToUsersList(csvFilePath)
	if err != nil {
		return err
	}
	if len(usersList) < 1 {
		return errorutils.CheckErrorf("an empty input file was provided")
	}
	var usersGroups []string
	if c.String(cliutils.UsersGroups) != "" {
		usersGroups = strings.Split(c.String(cliutils.UsersGroups), ",")
	}
	// Run command.
	usersCreateCmd.SetServerDetails(rtDetails).SetUsers(usersList).SetUsersGroups(usersGroups).SetReplaceIfExists(c.Bool(cliutils.Replace))
	return commands.Exec(usersCreateCmd)
}

func usersDeleteCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
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
		return cliutils.PrintHelpAndReturnError("missing <users list> OR --csv <users details file path>", c)
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
	csvInput, err := ioutil.ReadFile(csvFilePath)
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
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
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
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
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
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	rtDetails, err := createArtifactoryDetailsByFlags(c)
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

func accessTokenCreateCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	serverDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	// If the username is provided as an argument, then it is used when creating the token.
	// If not, then the configured username (or the value of the --user option) is used.
	var userName string
	if c.NArg() > 0 {
		userName = c.Args().Get(0)
	} else {
		userName = serverDetails.GetUser()
	}
	expiry, err := cliutils.GetIntFlagValue(c, "expiry", cliutils.TokenExpiry)
	if err != nil {
		return err
	}
	accessTokenCreateCmd := generic.NewAccessTokenCreateCommand()
	accessTokenCreateCmd.SetUserName(userName).SetServerDetails(serverDetails).SetRefreshable(c.Bool("refreshable")).SetExpiry(expiry).SetGroups(c.String("groups")).SetAudience(c.String("audience")).SetGrantAdmin(c.Bool("grant-admin"))
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

func validateBuildConfiguration(c *cli.Context, buildConfiguration *utils.BuildConfiguration) error {
	if buildConfiguration.BuildName == "" || buildConfiguration.BuildNumber == "" {
		return cliutils.PrintHelpAndReturnError("Build name and build number are expected as command arguments or environment variables.", c)
	}
	return nil
}

func getDebFlag(c *cli.Context) (deb string, err error) {
	deb = c.String("deb")
	slashesCount := strings.Count(deb, "/") - strings.Count(deb, "\\/")
	if deb != "" && slashesCount != 2 {
		return "", errors.New("the --deb option should be in the form of distribution/component/architecture")
	}
	return deb, nil
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
		Project(c.String("project")).
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
		Project(c.String("project")).
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
		Project(c.String("project")).
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
		Project(c.String("project")).
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

func createBuildInfoConfiguration(c *cli.Context) *buildinfocmd.Configuration {
	flags := new(buildinfocmd.Configuration)
	flags.BuildUrl = cliutils.GetBuildUrl(c.String("build-url"))
	flags.DryRun = c.Bool("dry-run")
	flags.EnvInclude = c.String("env-include")
	flags.EnvExclude = cliutils.GetEnvExclude(c.String("env-exclude"))
	if flags.EnvInclude == "" {
		flags.EnvInclude = "*"
	}
	// Allow using `env-exclude=""` and get no filters
	if flags.EnvExclude == "" {
		flags.EnvExclude = "*password*;*psw*;*secret*;*key*;*token*"
	}
	return flags
}

func createBuildPromoteConfiguration(c *cli.Context) services.PromotionParams {
	promotionParamsImpl := services.NewPromotionParams()
	promotionParamsImpl.Comment = c.String("comment")
	promotionParamsImpl.SourceRepo = c.String("source-repo")
	promotionParamsImpl.Status = c.String("status")
	promotionParamsImpl.IncludeDependencies = c.Bool("include-dependencies")
	promotionParamsImpl.Copy = c.Bool("copy")
	promotionParamsImpl.Properties = c.String("props")
	promotionParamsImpl.ProjectKey = utils.GetBuildProject(c.String("project"))

	// If the command received 3 args, read the build name, build number
	// and target repo as ags.
	buildName, buildNumber, targetRepo := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)
	// But if the command received only one arg, the build name and build number
	// are expected as env vars, and only the target repo is received as an arg.
	if len(c.Args()) == 1 {
		buildName, buildNumber, targetRepo = "", "", c.Args().Get(0)
	}

	promotionParamsImpl.BuildName, promotionParamsImpl.BuildNumber = utils.GetBuildNameAndNumber(buildName, buildNumber)
	promotionParamsImpl.TargetRepo = targetRepo
	return promotionParamsImpl
}

func createBuildDiscardConfiguration(c *cli.Context) services.DiscardBuildsParams {
	discardParamsImpl := services.NewDiscardBuildsParams()
	discardParamsImpl.DeleteArtifacts = c.Bool("delete-artifacts")
	discardParamsImpl.MaxBuilds = c.String("max-builds")
	discardParamsImpl.MaxDays = c.String("max-days")
	discardParamsImpl.ExcludeBuilds = c.String("exclude-builds")
	discardParamsImpl.Async = c.Bool("async")
	discardParamsImpl.BuildName = cliutils.GetBuildName(c.Args().Get(0))
	discardParamsImpl.ProjectKey = c.String("project")
	return discardParamsImpl
}

func createGitLfsCleanConfiguration(c *cli.Context) (gitLfsCleanConfiguration *generic.GitLfsCleanConfiguration) {
	gitLfsCleanConfiguration = new(generic.GitLfsCleanConfiguration)

	gitLfsCleanConfiguration.Refs = c.String("refs")
	if len(gitLfsCleanConfiguration.Refs) == 0 {
		gitLfsCleanConfiguration.Refs = "refs/remotes/*"
	}

	gitLfsCleanConfiguration.Repo = c.String("repo")
	gitLfsCleanConfiguration.Quiet = cliutils.GetQuietValue(c)
	dotGitPath := ""
	if c.NArg() == 1 {
		dotGitPath = c.Args().Get(0)
	}
	gitLfsCleanConfiguration.GitPath = dotGitPath
	return
}

func createDefaultDownloadSpec(c *cli.Context) (*spec.SpecFiles, error) {
	offset, limit, err := getOffsetAndLimitValues(c)
	if err != nil {
		return nil, err
	}
	return spec.NewBuilder().
		Pattern(strings.TrimPrefix(c.Args().Get(0), "/")).
		Props(c.String("props")).
		ExcludeProps(c.String("exclude-props")).
		Build(c.String("build")).
		Project(c.String("project")).
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
		IncludeDirs(c.Bool("include-dirs")).
		Target(c.Args().Get(1)).
		ArchiveEntries(c.String("archive-entries")).
		ValidateSymlinks(c.Bool("validate-symlinks")).
		BuildSpec(), nil
}

func createDownloadConfiguration(c *cli.Context) (downloadConfiguration *utils.DownloadConfiguration, err error) {
	downloadConfiguration = new(utils.DownloadConfiguration)
	downloadConfiguration.MinSplitSize, err = getMinSplit(c)
	if err != nil {
		return nil, err
	}
	downloadConfiguration.SplitCount, err = getSplitCount(c)
	if err != nil {
		return nil, err
	}
	downloadConfiguration.Threads, err = cliutils.GetThreadsCount(c)
	if err != nil {
		return nil, err
	}
	downloadConfiguration.Symlink = true
	return
}

func createDefaultUploadSpec(c *cli.Context) (*spec.SpecFiles, error) {
	offset, limit, err := getOffsetAndLimitValues(c)
	if err != nil {
		return nil, err
	}
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		TargetProps(c.String("target-props")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Flat(c.Bool("flat")).
		Explode(c.String("explode")).
		Regexp(c.Bool("regexp")).
		Ant(c.Bool("ant")).
		IncludeDirs(c.Bool("include-dirs")).
		Target(strings.TrimPrefix(c.Args().Get(1), "/")).
		Symlinks(c.Bool("symlinks")).
		Archive(c.String("archive")).
		BuildSpec(), nil
}

func createDefaultBuildAddDependenciesSpec(c *cli.Context) *spec.SpecFiles {
	pattern := c.Args().Get(2)
	if pattern == "" {
		// Build name and build number from env
		pattern = c.Args().Get(0)
	}
	return spec.NewBuilder().
		Pattern(pattern).
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Regexp(c.Bool("regexp")).
		Ant(c.Bool("ant")).
		BuildSpec()
}

func fixWinPathsForDownloadCmd(uploadSpec *spec.SpecFiles, c *cli.Context) {
	if coreutils.IsWindows() {
		for i, file := range uploadSpec.Files {
			uploadSpec.Files[i].Target = fixWinPathBySource(file.Target, c.IsSet("spec"))
		}
	}
}

func fixWinPathBySource(path string, fromSpec bool) string {
	if strings.Count(path, "/") > 0 {
		// Assuming forward slashes - not doubling backslash to allow regexp escaping
		return ioutils.UnixToWinPathSeparator(path)
	}
	if fromSpec {
		// Doubling backslash only for paths from spec files (that aren't forward slashed)
		return ioutils.DoubleWinPathSeparator(path)
	}
	return path
}

func createUploadConfiguration(c *cli.Context) (uploadConfiguration *utils.UploadConfiguration, err error) {
	uploadConfiguration = new(utils.UploadConfiguration)
	uploadConfiguration.Threads, err = cliutils.GetThreadsCount(c)
	if err != nil {
		return nil, err
	}
	uploadConfiguration.Deb, err = getDebFlag(c)
	if err != nil {
		return
	}
	return
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

func isFailNoOp(context *cli.Context) bool {
	if context == nil {
		return false
	}
	return context.Bool("fail-no-op")
}

// Returns build configuration struct using the params provided from the console.
func createBuildConfiguration(c *cli.Context) *utils.BuildConfiguration {
	buildConfiguration := new(utils.BuildConfiguration)
	buildNameArg, buildNumberArg := c.Args().Get(0), c.Args().Get(1)
	if buildNameArg == "" || buildNumberArg == "" {
		buildNameArg = ""
		buildNumberArg = ""
	}
	buildConfiguration.BuildName, buildConfiguration.BuildNumber = utils.GetBuildNameAndNumber(buildNameArg, buildNumberArg)
	buildConfiguration.Project = utils.GetBuildProject(c.String("project"))
	return buildConfiguration
}

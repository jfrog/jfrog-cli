package artifactory

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/dotnet"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/permissiontarget"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/usersmanagement"
	containerutils "github.com/jfrog/jfrog-cli-core/artifactory/utils/container"
	coreCommonCommands "github.com/jfrog/jfrog-cli-core/common/commands"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/utils/ioutils"
	"github.com/jfrog/jfrog-cli/docs/artifactory/accesstokencreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/builddockercreate"
	dotnetdocs "github.com/jfrog/jfrog-cli/docs/artifactory/dotnet"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dotnetconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupaddusers"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupcreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/groupdelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetcreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetdelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargettemplate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/permissiontargetupdate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/podmanpull"
	"github.com/jfrog/jfrog-cli/docs/artifactory/podmanpush"
	"github.com/jfrog/jfrog-cli/docs/artifactory/usercreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/userscreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/usersdelete"
	logUtils "github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	ioUtils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jszwec/csvutil"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/buildinfo"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/curl"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/distribution"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/golang"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/pip"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/replication"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/repository"
	commandUtils "github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	npmUtils "github.com/jfrog/jfrog-cli-core/artifactory/utils/npm"
	"github.com/jfrog/jfrog-cli-core/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/docs/common"
	coreConfig "github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli/config"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildadddependencies"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildaddgit"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildappend"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildclean"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildcollectenv"
	"github.com/jfrog/jfrog-cli/docs/artifactory/builddiscard"
	"github.com/jfrog/jfrog-cli/docs/artifactory/builddistribute"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildpromote"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildpublish"
	"github.com/jfrog/jfrog-cli/docs/artifactory/buildscan"
	configdocs "github.com/jfrog/jfrog-cli/docs/artifactory/config"
	copydocs "github.com/jfrog/jfrog-cli/docs/artifactory/copy"
	curldocs "github.com/jfrog/jfrog-cli/docs/artifactory/curl"
	"github.com/jfrog/jfrog-cli/docs/artifactory/delete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/deleteprops"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dockerpromote"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dockerpull"
	"github.com/jfrog/jfrog-cli/docs/artifactory/dockerpush"
	"github.com/jfrog/jfrog-cli/docs/artifactory/download"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gitlfsclean"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gocommand"
	"github.com/jfrog/jfrog-cli/docs/artifactory/goconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gopublish"
	gradledoc "github.com/jfrog/jfrog-cli/docs/artifactory/gradle"
	"github.com/jfrog/jfrog-cli/docs/artifactory/gradleconfig"
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
	"github.com/jfrog/jfrog-cli/docs/artifactory/ping"
	"github.com/jfrog/jfrog-cli/docs/artifactory/pipconfig"
	"github.com/jfrog/jfrog-cli/docs/artifactory/pipinstall"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundlecreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundledelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundledistribute"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundlesign"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundleupdate"
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
	"github.com/jfrog/jfrog-cli/docs/artifactory/use"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	buildinfocmd "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	distributionServices "github.com/jfrog/jfrog-client-go/distribution/services"
	distributionServicesUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "config",
			Flags:        cliutils.GetCommandFlags(cliutils.RtConfig),
			Aliases:      []string{"c"},
			Description:  configdocs.Description,
			HelpName:     corecommon.CreateUsage("rt config", configdocs.Description, configdocs.Usage),
			UsageText:    configdocs.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc("show", "delete", "clear", "import", "export"),
			Action: func(c *cli.Context) error {
				return configCmd(c)
			},
		},
		{
			Name:         "use",
			Description:  use.Description,
			HelpName:     corecommon.CreateUsage("rt use", use.Description, use.Usage),
			UsageText:    use.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(coreCommonCommands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return useCmd(c)
			},
		},
		{
			Name:         "upload",
			Flags:        cliutils.GetCommandFlags(cliutils.Upload),
			Aliases:      []string{"u"},
			Description:  upload.Description,
			HelpName:     corecommon.CreateUsage("rt upload", upload.Description, upload.Usage),
			UsageText:    upload.Arguments,
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
			Description:  download.Description,
			HelpName:     corecommon.CreateUsage("rt download", download.Description, download.Usage),
			UsageText:    download.Arguments,
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
			Description:  move.Description,
			HelpName:     corecommon.CreateUsage("rt move", move.Description, move.Usage),
			UsageText:    move.Arguments,
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
			Description:  copydocs.Description,
			HelpName:     corecommon.CreateUsage("rt copy", copydocs.Description, copydocs.Usage),
			UsageText:    copydocs.Arguments,
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
			Description:  delete.Description,
			HelpName:     corecommon.CreateUsage("rt delete", delete.Description, delete.Usage),
			UsageText:    delete.Arguments,
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
			Description:  search.Description,
			HelpName:     corecommon.CreateUsage("rt search", search.Description, search.Usage),
			UsageText:    search.Arguments,
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
			Description:  setprops.Description,
			HelpName:     corecommon.CreateUsage("rt set-props", setprops.Description, setprops.Usage),
			UsageText:    setprops.Arguments,
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
			Description:  deleteprops.Description,
			HelpName:     corecommon.CreateUsage("rt delete-props", deleteprops.Description, deleteprops.Usage),
			UsageText:    deleteprops.Arguments,
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
			Description:  buildpublish.Description,
			HelpName:     corecommon.CreateUsage("rt build-publish", buildpublish.Description, buildpublish.Usage),
			UsageText:    buildpublish.Arguments,
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
			Description:  buildcollectenv.Description,
			HelpName:     corecommon.CreateUsage("rt build-collect-env", buildcollectenv.Description, buildcollectenv.Usage),
			UsageText:    buildcollectenv.Arguments,
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
			Description:  buildappend.Description,
			HelpName:     corecommon.CreateUsage("rt build-append", buildappend.Description, buildappend.Usage),
			UsageText:    buildappend.Arguments,
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
			Description:  buildadddependencies.Description,
			HelpName:     corecommon.CreateUsage("rt build-add-dependencies", buildadddependencies.Description, buildadddependencies.Usage),
			UsageText:    buildadddependencies.Arguments,
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
			Description:  buildaddgit.Description,
			HelpName:     corecommon.CreateUsage("rt build-add-git", buildaddgit.Description, buildaddgit.Usage),
			UsageText:    buildaddgit.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildAddGitCmd(c)
			},
		},
		{
			Name:         "build-scan",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildScan),
			Aliases:      []string{"bs"},
			Description:  buildscan.Description,
			HelpName:     corecommon.CreateUsage("rt build-scan", buildscan.Description, buildscan.Usage),
			UsageText:    buildscan.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildScanCmd(c)
			},
		},
		{
			Name:         "build-clean",
			Aliases:      []string{"bc"},
			Description:  buildclean.Description,
			HelpName:     corecommon.CreateUsage("rt build-clean", buildclean.Description, buildclean.Usage),
			UsageText:    buildclean.Arguments,
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
			Description:  buildpromote.Description,
			HelpName:     corecommon.CreateUsage("rt build-promote", buildpromote.Description, buildpromote.Usage),
			UsageText:    buildpromote.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildPromoteCmd(c)
			},
		},
		{
			Name:         "build-distribute",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildDistribute),
			Aliases:      []string{"bd"},
			Description:  builddistribute.Description,
			HelpName:     corecommon.CreateUsage("rt build-distribute", builddistribute.Description, builddistribute.Usage),
			UsageText:    builddistribute.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return buildDistributeCmd(c)
			},
		},
		{
			Name:         "build-discard",
			Flags:        cliutils.GetCommandFlags(cliutils.BuildDiscard),
			Aliases:      []string{"bdi"},
			Description:  builddiscard.Description,
			HelpName:     corecommon.CreateUsage("rt build-discard", builddiscard.Description, builddiscard.Usage),
			UsageText:    builddiscard.Arguments,
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
			Description:  gitlfsclean.Description,
			HelpName:     corecommon.CreateUsage("rt git-lfs-clean", gitlfsclean.Description, gitlfsclean.Usage),
			UsageText:    gitlfsclean.Arguments,
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
			Description:  mvnconfig.Description,
			HelpName:     corecommon.CreateUsage("rt mvn-config", mvnconfig.Description, mvnconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return createMvnConfigCmd(c)
			},
		},
		{
			Name:            "mvn",
			Flags:           cliutils.GetCommandFlags(cliutils.Mvn),
			Description:     mvndoc.Description,
			HelpName:        corecommon.CreateUsage("rt mvn", mvndoc.Description, mvndoc.Usage),
			UsageText:       mvndoc.Arguments,
			ArgsUsage:       common.CreateEnvVars(mvndoc.EnvVar),
			SkipFlagParsing: shouldSkipMavenFlagParsing(),
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return mvnCmd(c)
			},
		},
		{
			Name:         "gradle-config",
			Aliases:      []string{"gradlec"},
			Flags:        cliutils.GetCommandFlags(cliutils.GradleConfig),
			Description:  gradleconfig.Description,
			HelpName:     corecommon.CreateUsage("rt gradle-config", gradleconfig.Description, gradleconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return createGradleConfigCmd(c)
			},
		},
		{
			Name:            "gradle",
			Flags:           cliutils.GetCommandFlags(cliutils.Gradle),
			Description:     gradledoc.Description,
			HelpName:        corecommon.CreateUsage("rt gradle", gradledoc.Description, gradledoc.Usage),
			UsageText:       gradledoc.Arguments,
			ArgsUsage:       common.CreateEnvVars(gradledoc.EnvVar),
			SkipFlagParsing: shouldSkipGradleFlagParsing(),
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return gradleCmd(c)
			},
		},
		{
			Name:         "docker-promote",
			Flags:        cliutils.GetCommandFlags(cliutils.DockerPromote),
			Aliases:      []string{"dpr"},
			Description:  dockerpromote.Description,
			HelpName:     corecommon.CreateUsage("rt docker-promote", dockerpromote.Description, dockerpromote.Usage),
			UsageText:    dockerpromote.Arguments,
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
			Description:  dockerpush.Description,
			HelpName:     corecommon.CreateUsage("rt docker-push", dockerpush.Description, dockerpush.Usage),
			UsageText:    dockerpush.Arguments,
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
			Description:  dockerpull.Description,
			HelpName:     corecommon.CreateUsage("rt docker-pull", dockerpull.Description, dockerpull.Usage),
			UsageText:    dockerpull.Arguments,
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
			Description:  podmanpush.Description,
			HelpName:     corecommon.CreateUsage("rt podman-push", podmanpush.Description, podmanpush.Usage),
			UsageText:    podmanpush.Arguments,
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
			Description:  podmanpull.Description,
			HelpName:     corecommon.CreateUsage("rt podman-pull", podmanpull.Description, podmanpull.Usage),
			UsageText:    podmanpull.Arguments,
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
			Description:  builddockercreate.Description,
			HelpName:     corecommon.CreateUsage("rt build-docker-create", builddockercreate.Description, builddockercreate.Usage),
			UsageText:    builddockercreate.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return BuildDockerCreateCmd(c)
			},
		},
		{
			Name:         "npm-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NpmConfig),
			Aliases:      []string{"npmc"},
			Description:  npmconfig.Description,
			HelpName:     corecommon.CreateUsage("rt npm-config", npmconfig.Description, npmconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return createNpmConfigCmd(c)
			},
		},
		{
			Name:            "npm-install",
			Flags:           cliutils.GetCommandFlags(cliutils.Npm),
			Aliases:         []string{"npmi"},
			Description:     npminstall.Description,
			HelpName:        corecommon.CreateUsage("rt npm-install", npminstall.Description, npminstall.Usage),
			UsageText:       npminstall.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: shouldSkipNpmFlagParsing(),
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return npmInstallOrCiCmd(c, npm.NewNpmInstallCommand(), npmLegacyInstallCmd)
			},
		},
		{
			Name:            "npm-ci",
			Flags:           cliutils.GetCommandFlags(cliutils.Npm),
			Aliases:         []string{"npmci"},
			Description:     npmci.Description,
			HelpName:        corecommon.CreateUsage("rt npm-ci", npmci.Description, npmci.Usage),
			UsageText:       npmci.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: shouldSkipNpmFlagParsing(),
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return npmInstallOrCiCmd(c, npm.NewNpmCiCommand(), npmLegacyCiCmd)
			},
		},
		{
			Name:            "npm-publish",
			Flags:           cliutils.GetCommandFlags(cliutils.NpmPublish),
			Aliases:         []string{"npmp"},
			Description:     npmpublish.Description,
			HelpName:        corecommon.CreateUsage("rt npm-publish", npmpublish.Description, npmpublish.Usage),
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: shouldSkipNpmFlagParsing(),
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return npmPublishCmd(c)
			},
		},
		{
			Name:         "nuget-config",
			Flags:        cliutils.GetCommandFlags(cliutils.NugetConfig),
			Aliases:      []string{"nugetc"},
			Description:  nugetconfig.Description,
			HelpName:     corecommon.CreateUsage("rt nuget-config", nugetconfig.Description, nugetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return createNugetConfigCmd(c)
			},
		},
		{
			Name:            "nuget",
			Flags:           cliutils.GetCommandFlags(cliutils.Nuget),
			Description:     nugetdocs.Description,
			HelpName:        corecommon.CreateUsage("rt nuget", nugetdocs.Description, nugetdocs.Usage),
			UsageText:       nugetdocs.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: shouldSkipNugetFlagParsing(),
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return nugetCmd(c)
			},
		},
		{
			Name:         "nuget-deps-tree",
			Aliases:      []string{"ndt"},
			Description:  nugettree.Description,
			HelpName:     corecommon.CreateUsage("rt nuget-deps-tree", nugettree.Description, nugettree.Usage),
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
			Description:  dotnetconfig.Description,
			HelpName:     corecommon.CreateUsage("rt dotnet-config", dotnetconfig.Description, dotnetconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return createDotnetConfigCmd(c)
			},
		},
		{
			Name:            "dotnet",
			Flags:           cliutils.GetCommandFlags(cliutils.Dotnet),
			Description:     dotnetdocs.Description,
			HelpName:        corecommon.CreateUsage("rt dotnet", dotnetdocs.Description, dotnetdocs.Usage),
			UsageText:       dotnetdocs.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return dotnetCmd(c)
			},
		},
		{
			Name:         "go-config",
			Flags:        cliutils.GetCommandFlags(cliutils.GoConfig),
			Description:  goconfig.Description,
			HelpName:     corecommon.CreateUsage("rt go-config", goconfig.Description, goconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return createGoConfigCmd(c)
			},
		},
		{
			Name:         "go-publish",
			Flags:        cliutils.GetCommandFlags(cliutils.GoPublish),
			Aliases:      []string{"gp"},
			Description:  gopublish.Description,
			HelpName:     corecommon.CreateUsage("rt go-publish", gopublish.Description, gopublish.Usage),
			UsageText:    gopublish.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return goCmd(c, goPublishCmd, goLegacyPublishCmd)
			},
		},
		{
			Name:            "go",
			Flags:           cliutils.GetCommandFlags(cliutils.Go),
			Aliases:         []string{"go"},
			Description:     gocommand.Description,
			HelpName:        corecommon.CreateUsage("rt go", gocommand.Description, gocommand.Usage),
			UsageText:       gocommand.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: shouldSkipGoFlagParsing(),
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return goCmd(c, goNativeCmd, goLegacyCmd)
			},
		},
		{
			Name:         "ping",
			Flags:        cliutils.GetCommandFlags(cliutils.Ping),
			Aliases:      []string{"p"},
			Description:  ping.Description,
			HelpName:     corecommon.CreateUsage("rt ping", ping.Description, ping.Usage),
			UsageText:    ping.Arguments,
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
			Description:     curldocs.Description,
			HelpName:        corecommon.CreateUsage("rt curl", curldocs.Description, curldocs.Usage),
			UsageText:       curldocs.Arguments,
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
			Description:  pipconfig.Description,
			HelpName:     corecommon.CreateUsage("rt pipc", pipconfig.Description, pipconfig.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return createPipConfigCmd(c)
			},
		},
		{
			Name:            "pip-install",
			Flags:           cliutils.GetCommandFlags(cliutils.PipInstall),
			Aliases:         []string{"pipi"},
			Description:     pipinstall.Description,
			HelpName:        corecommon.CreateUsage("rt pipi", pipinstall.Description, pipinstall.Usage),
			UsageText:       pipinstall.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			SkipFlagParsing: true,
			BashComplete:    corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return pipInstallCmd(c)
			},
		},
		{
			Name:         "release-bundle-create",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleCreate),
			Aliases:      []string{"rbc"},
			Description:  releasebundlecreate.Description,
			HelpName:     corecommon.CreateUsage("rt rbc", releasebundlecreate.Description, releasebundlecreate.Usage),
			UsageText:    releasebundlecreate.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return releaseBundleCreateCmd(c)
			},
		},
		{
			Name:         "release-bundle-update",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleUpdate),
			Aliases:      []string{"rbu"},
			Description:  releasebundleupdate.Description,
			HelpName:     corecommon.CreateUsage("rt rbu", releasebundleupdate.Description, releasebundleupdate.Usage),
			UsageText:    releasebundleupdate.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return releaseBundleUpdateCmd(c)
			},
		},
		{
			Name:         "release-bundle-sign",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleSign),
			Aliases:      []string{"rbs"},
			Description:  releasebundlesign.Description,
			HelpName:     corecommon.CreateUsage("rt rbs", releasebundlesign.Description, releasebundlesign.Usage),
			UsageText:    releasebundlesign.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return releaseBundleSignCmd(c)
			},
		},
		{
			Name:         "release-bundle-distribute",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleDistribute),
			Aliases:      []string{"rbd"},
			Description:  releasebundledistribute.Description,
			HelpName:     corecommon.CreateUsage("rt rbd", releasebundledistribute.Description, releasebundledistribute.Usage),
			UsageText:    releasebundledistribute.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return releaseBundleDistributeCmd(c)
			},
		},
		{
			Name:         "release-bundle-delete",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleDelete),
			Aliases:      []string{"rbdel"},
			Description:  releasebundledelete.Description,
			HelpName:     corecommon.CreateUsage("rt rbdel", releasebundledelete.Description, releasebundledelete.Usage),
			UsageText:    releasebundledelete.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return releaseBundleDeleteCmd(c)
			},
		},
		{
			Name:         "repo-template",
			Aliases:      []string{"rpt"},
			Description:  repotemplate.Description,
			HelpName:     corecommon.CreateUsage("rt rpt", repotemplate.Description, repotemplate.Usage),
			UsageText:    repotemplate.Arguments,
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
			Description:  repocreate.Description,
			HelpName:     corecommon.CreateUsage("rt rc", repocreate.Description, repocreate.Usage),
			UsageText:    repocreate.Arguments,
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
			Description:  repoupdate.Description,
			HelpName:     corecommon.CreateUsage("rt ru", repoupdate.Description, repoupdate.Usage),
			UsageText:    repoupdate.Arguments,
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
			Description:  repodelete.Description,
			HelpName:     corecommon.CreateUsage("rt rdel", repodelete.Description, repodelete.Usage),
			UsageText:    repodelete.Arguments,
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
			Description:  replicationtemplate.Description,
			HelpName:     corecommon.CreateUsage("rt rplt", replicationtemplate.Description, replicationtemplate.Usage),
			UsageText:    replicationtemplate.Arguments,
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
			Description:  replicationcreate.Description,
			HelpName:     corecommon.CreateUsage("rt rplc", replicationcreate.Description, replicationcreate.Usage),
			UsageText:    replicationcreate.Arguments,
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
			Description:  replicationdelete.Description,
			HelpName:     corecommon.CreateUsage("rt rpldel", replicationdelete.Description, replicationdelete.Usage),
			UsageText:    replicationdelete.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return replicationDeleteCmd(c)
			},
		},
		{
			Name:         "permission-target-template",
			Aliases:      []string{"ptt"},
			Description:  permissiontargettemplate.Description,
			HelpName:     corecommon.CreateUsage("rt ptt", permissiontargettemplate.Description, permissiontargettemplate.Usage),
			UsageText:    permissiontargettemplate.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return permissionTargrtTemplateCmd(c)
			},
		},
		{
			Name:         "permission-target-create",
			Aliases:      []string{"ptc"},
			Flags:        cliutils.GetCommandFlags(cliutils.TemplateConsumer),
			Description:  permissiontargetcreate.Description,
			HelpName:     corecommon.CreateUsage("rt ptc", permissiontargetcreate.Description, permissiontargetcreate.Usage),
			UsageText:    permissiontargetcreate.Arguments,
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
			Description:  permissiontargetupdate.Description,
			HelpName:     corecommon.CreateUsage("rt ptu", permissiontargetupdate.Description, permissiontargetupdate.Usage),
			UsageText:    permissiontargetupdate.Arguments,
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
			Description:  permissiontargetdelete.Description,
			HelpName:     corecommon.CreateUsage("rt ptdel", permissiontargetdelete.Description, permissiontargetdelete.Usage),
			UsageText:    permissiontargetdelete.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return permissionTargetDeleteCmd(c)
			},
		},
		{
			Name:         "user-create",
			Flags:        cliutils.GetCommandFlags(cliutils.UserCreate),
			Description:  usercreate.Description,
			HelpName:     corecommon.CreateUsage("rt user-create", usercreate.Description, usercreate.Usage),
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
			Description:  userscreate.Description,
			HelpName:     corecommon.CreateUsage("rt uc", userscreate.Description, userscreate.Usage),
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
			Description:  usersdelete.Description,
			HelpName:     corecommon.CreateUsage("rt udel", usersdelete.Description, usersdelete.Usage),
			UsageText:    usersdelete.Arguments,
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
			Description:  groupcreate.Description,
			HelpName:     corecommon.CreateUsage("rt gc", groupcreate.Description, groupcreate.Usage),
			UsageText:    groupcreate.Arguments,
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
			Description:  groupaddusers.Description,
			HelpName:     corecommon.CreateUsage("rt gau", groupaddusers.Description, groupaddusers.Usage),
			UsageText:    groupaddusers.Arguments,
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
			Description:  groupdelete.Description,
			HelpName:     corecommon.CreateUsage("rt gdel", groupdelete.Description, groupdelete.Usage),
			UsageText:    groupdelete.Arguments,
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
			Description:  accesstokencreate.Description,
			HelpName:     corecommon.CreateUsage("rt atc", accesstokencreate.Description, accesstokencreate.Usage),
			UsageText:    accesstokencreate.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return accessTokenCreateCmd(c)
			},
		},
	})
}

func createArtifactoryDetailsByFlags(c *cli.Context, distribution bool) (*coreConfig.ServerDetails, error) {
	artDetails, err := createArtifactoryDetailsWithConfigOffer(c, distribution)
	if err != nil {
		return nil, err
	}
	if distribution {
		if artDetails.DistributionUrl == "" {
			return nil, errors.New("the --dist-url option is mandatory")
		}
	} else {
		if artDetails.ArtifactoryUrl == "" {
			return nil, errors.New("the --url option is mandatory")
		}
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

func getThreadsCount(c *cli.Context) (threads int, err error) {
	threads = cliutils.Threads
	err = nil
	if c.String("threads") != "" {
		threads, err = strconv.Atoi(c.String("threads"))
		if err != nil || threads < 1 {
			err = errors.New("the '--threads' option should have a numeric positive value")
			return 0, err
		}
	}
	return threads, nil
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

func validateCommand(args []string, notAllowedFlags []cli.Flag) error {
	for _, arg := range args {
		for _, flag := range notAllowedFlags {
			// Cli flags are in the format of --key, therefore, the -- need to be added to the name
			if strings.Contains(arg, "--"+flag.GetName()) {
				return errorutils.CheckError(fmt.Errorf("flag --%s can't be used with config file", flag.GetName()))
			}
		}
	}
	return nil
}

func useCmd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	log.Warn(`The "jfrog rt use" command is deprecated. Please use "jfrog config use" instead.`)
	return coreCommonCommands.Use(c.Args()[0])
}

func configCmd(c *cli.Context) error {
	if len(c.Args()) > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	log.Warn(`The "jfrog rt config" command is deprecated. Please use "jfrog config" command instead. You can use it as follows:
	The command includes the following sub-commands - "jfrog config add", "jfrog config edit", "jfrog config show", "jfrog config remove", "jfrog config import" and "jfrog config export".
	Important: When switching to the new command, please replace "--url" with "--artifactory-url".
	For example:
	Old syntax: "jfrog rt config <server-id> --url=<artifactoryUrl>"
	New syntax: "jfrog config add <server-id> --artifactory-url=<artifactory-url>"`)

	var serverId string
	configCommandConfiguration, err := config.CreateConfigCommandConfiguration(c)
	if err != nil {
		return err
	}
	configCommandConfiguration.ServerDetails.ArtifactoryUrl = configCommandConfiguration.ServerDetails.Url
	configCommandConfiguration.ServerDetails.Url = ""
	if len(c.Args()) == 2 {
		if c.Args()[0] == "import" {
			return coreCommonCommands.Import(c.Args()[1])
		}
		serverId = c.Args()[1]
		if err := config.ValidateServerId(serverId); err != nil {
			return err
		}
		artDetails, err := coreConfig.GetSpecificConfig(serverId, true, false)
		if err != nil {
			return err
		}
		if artDetails.IsEmpty() {
			log.Info("\"" + serverId + "\" configuration could not be found.")
			return nil
		}
		if c.Args()[0] == "delete" {
			if configCommandConfiguration.Interactive {
				if !coreutils.AskYesNo("Are you sure you want to delete \""+serverId+"\" configuration?", false) {
					return nil
				}
			}
			return coreCommonCommands.DeleteConfig(serverId)
		}
		if c.Args()[0] == "export" {
			return coreCommonCommands.Export(serverId)
		}
	}
	if len(c.Args()) > 0 {
		if c.Args()[0] == "show" {
			return coreCommonCommands.ShowConfig(serverId)
		}
		if c.Args()[0] == "clear" {
			coreCommonCommands.ClearConfig(configCommandConfiguration.Interactive)
			return nil
		}
		serverId = c.Args()[0]
		err = config.ValidateServerId(serverId)
		if err != nil {
			return err
		}
	}
	err = validateConfigFlags(configCommandConfiguration)
	if err != nil {
		return err
	}
	configCmd := coreCommonCommands.NewConfigCommand().SetDetails(configCommandConfiguration.ServerDetails).SetInteractive(configCommandConfiguration.Interactive).
		SetServerId(serverId).SetEncPassword(configCommandConfiguration.EncPassword).SetUseBasicAuthOnly(configCommandConfiguration.BasicAuthOnly)
	return configCmd.Config()
}

func mvnLegacyCmd(c *cli.Context) error {
	log.Warn(deprecatedWarningWithExample(utils.Maven, os.Args[2], "mvnc"))
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	configuration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	threads, err := getThreadsCount(c)
	if err != nil {
		return err
	}
	// We're splitting the goals, because the legacy maven command accepts the goals as one argument.
	goals := strings.Split(c.Args().Get(0), " ")
	mvnCmd := mvn.NewMvnCommand().SetConfiguration(configuration).SetConfigPath(c.Args().Get(1)).SetGoals(goals).SetThreads(threads)

	return commands.Exec(mvnCmd)
}

func mvnCmd(c *cli.Context) error {
	if show, err := showCmdHelpIfNeeded(c); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Maven)
	if err != nil {
		return err
	}
	if exists {
		// Found a config file. Continue as native command.
		if c.NArg() < 1 {
			return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
		}
		args := cliutils.ExtractCommand(c)
		// Validates the mvn command. If a config file is found, the only flags that can be used are build-name, build-number and module.
		// Otherwise, throw an error.
		if err := validateCommand(args, cliutils.GetBasicBuildToolsFlags()); err != nil {
			return err
		}
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
		mvnCmd := mvn.NewMvnCommand().SetConfiguration(buildConfiguration).SetConfigPath(configFilePath).SetGoals(filteredMavenArgs).SetThreads(threads).SetInsecureTls(insecureTls)
		return commands.Exec(mvnCmd)
	}
	return mvnLegacyCmd(c)
}

func gradleCmd(c *cli.Context) error {
	if show, err := showCmdHelpIfNeeded(c); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Gradle)
	if err != nil {
		return err
	}
	if exists {
		// Found a config file. Continue as native command.
		if c.NArg() < 1 {
			return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
		}
		args := cliutils.ExtractCommand(c)
		// Validates the gradle command. If a config file is found, the only flags that can be used are build-name, build-number and module.
		// Otherwise, throw an error.
		if err := validateCommand(args, cliutils.GetBasicBuildToolsFlags()); err != nil {
			return err
		}
		filteredGradleArgs, buildConfiguration, err := utils.ExtractBuildDetailsFromArgs(args)
		if err != nil {
			return err
		}
		filteredGradleArgs, threads, err := extractThreadsFlag(filteredGradleArgs)
		if err != nil {
			return err
		}
		gradleCmd := gradle.NewGradleCommand().SetConfiguration(buildConfiguration).SetTasks(strings.Join(filteredGradleArgs, " ")).SetConfigPath(configFilePath).SetThreads(threads)
		return commands.Exec(gradleCmd)
	}
	return gradleLegacyCmd(c)
}

func gradleLegacyCmd(c *cli.Context) error {
	log.Warn(deprecatedWarningWithExample(utils.Gradle, os.Args[2], "gradlec"))

	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	configuration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	threads, err := getThreadsCount(c)
	if err != nil {
		return err
	}
	gradleCmd := gradle.NewGradleCommand()
	gradleCmd.SetConfiguration(configuration).SetTasks(c.Args().Get(0)).SetConfigPath(c.Args().Get(1)).SetThreads(threads)

	return commands.Exec(gradleCmd)
}

func dockerPromoteCmd(c *cli.Context) error {
	if c.NArg() != 3 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	artDetails, err := createArtifactoryDetailsByFlags(c, false)
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
	artDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	imageTag := c.Args().Get(0)
	targetRepo := c.Args().Get(1)
	skipLogin := c.Bool("skip-login")

	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	dockerPushCommand := container.NewPushCommand(containerManagerType)
	threads, err := getThreadsCount(c)
	if err != nil {
		return err
	}
	dockerPushCommand.SetThreads(threads).SetBuildConfiguration(buildConfiguration).SetRepo(targetRepo).SetSkipLogin(skipLogin).SetServerDetails(artDetails).SetImageTag(imageTag)

	return commands.Exec(dockerPushCommand)
}

func containerPullCmd(c *cli.Context, containerManagerType containerutils.ContainerManagerType) error {
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	artDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	imageTag := c.Args().Get(0)
	sourceRepo := c.Args().Get(1)
	skipLogin := c.Bool("skip-login")
	buildConfiguration, err := createBuildConfigurationWithModule(c)
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
	artDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	sourceRepo := c.Args().Get(0)
	imageNameWithDigestFile := c.String("image-file")
	if imageNameWithDigestFile == "" {
		return cliutils.PrintHelpAndReturnError("The '--image-file' command option was not provided.", c)
	}
	buildConfiguration, err := createBuildConfigurationWithModule(c)
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

func nugetCmd(c *cli.Context) error {
	if show, err := showCmdHelpIfNeeded(c); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Nuget)
	if err != nil {
		return err
	}

	// A config file was found.
	if exists {
		if c.NArg() < 1 {
			return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
		}

		rtDetails, targetRepo, useNugetV2, err := getNugetAndDotnetConfigFields(configFilePath)
		if err != nil {
			return err
		}

		args := cliutils.ExtractCommand(c)

		// Validates the nuget command. If a config file is found, the only flags that can be used are build-name, build-number and module.
		// Otherwise, throw an error.
		if err := validateCommand(args, cliutils.GetLegacyNugetFlags()); err != nil {
			return err
		}

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
	// If config file not found, use nuget legacy command
	return nugetLegacyCmd(c)
}

func nugetLegacyCmd(c *cli.Context) error {
	log.Warn(deprecatedWarningWithExample(utils.Nuget, os.Args[2], "nugetc"))
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	nugetCmd := dotnet.NewLegacyNugetCommand()
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}

	nugetCmd.SetBasicCommand(c.Args().Get(0)).
		SetArgAndFlags(getLegacyNugetArgsList(c)).
		SetRepoName(c.Args().Get(1)).
		SetBuildConfiguration(buildConfiguration).
		SetSolutionPath(c.String(cliutils.SolutionRoot)).
		SetUseNugetV2(c.Bool(cliutils.LegacyNugetV2)).
		SetServerDetails(rtDetails)

	return commands.Exec(nugetCmd)
}

func getLegacyNugetArgsList(c *cli.Context) []string {
	args := c.String(cliutils.NugetArgs)
	if args == "" {
		return []string{}
	}
	return strings.Split(args, " ")
}

func nugetDepsTreeCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	return dotnet.DependencyTreeCmd()
}

func dotnetCmd(c *cli.Context) error {
	if show, err := showCmdHelpIfNeeded(c); show || err != nil {
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
		return errors.New(fmt.Sprintf("Error occurred while attempting to read dotnet-configuration file.\n" +
			"Please run 'jfrog rt dotnet-config' command prior to running 'jfrog rt dotnet'."))
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

func npmLegacyInstallCmd(c *cli.Context) error {
	log.Warn(deprecatedWarningWithExample(utils.Npm, os.Args[2], "npmc"))
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	npmCmd := npm.NewNpmLegacyInstallCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	threads, err := getThreadsCount(c)
	if err != nil {
		return err
	}
	npmInstallArgs, err := coreutils.ParseArgs(strings.Split(c.String("npm-args"), " "))
	if err != nil {
		return err
	}
	npmCmd.SetThreads(threads).SetBuildConfiguration(buildConfiguration).SetRepo(c.Args().Get(0)).SetNpmArgs(npmInstallArgs).SetServerDetails(rtDetails)

	return commands.Exec(npmCmd)
}

func npmInstallOrCiCmd(c *cli.Context, npmCmd *npm.NpmInstallOrCiCommand, npmLegacyCommand func(*cli.Context) error) error {
	if show, err := showCmdHelpIfNeeded(c); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Npm)
	if err != nil {
		return err
	}

	if exists {
		// Found a config file. Continue as native command.
		args := cliutils.ExtractCommand(c)
		// Validates the npm command. If a config file is found, the only flags that can be used are threads, build-name, build-number and module.
		// Otherwise, throw an error.
		if err := validateCommand(args, cliutils.GetLegacyNpmFlags()); err != nil {
			return err
		}
		npmCmd.SetConfigFilePath(configFilePath).SetArgs(args)
		return commands.Exec(npmCmd)
	}
	// If config file not found, use Npm legacy command
	return npmLegacyCommand(c)
}

func npmLegacyCiCmd(c *cli.Context) error {
	log.Warn(deprecatedWarningWithExample(utils.Npm, os.Args[2], "npmc"))
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	npmCmd := npm.NewNpmLegacyCiCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	threads, err := getThreadsCount(c)
	if err != nil {
		return err
	}
	npmCmd.SetThreads(threads).SetBuildConfiguration(buildConfiguration).SetRepo(c.Args().Get(0)).SetServerDetails(rtDetails)
	return commands.Exec(npmCmd)
}

func npmPublishCmd(c *cli.Context) error {
	if show, err := showCmdHelpIfNeeded(c); show || err != nil {
		return err
	}

	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Npm)
	if err != nil {
		return err
	}
	if exists {
		// Found a config file. Continue as native command.
		args := cliutils.ExtractCommand(c)
		// Validates the npm command. If a config file is found, the only flags that can be used are build-name, build-number and module.
		// Otherwise, throw an error.
		if err := validateCommand(args, cliutils.GetLegacyNpmFlags()); err != nil {
			return err
		}
		npmCmd := npm.NewNpmPublishCommand()
		npmCmd.SetConfigFilePath(configFilePath).SetArgs(args)
		return commands.Exec(npmCmd)
	}
	// If config file not found, use Npm legacy command
	return npmLegacyPublishCmd(c)
}

func npmLegacyPublishCmd(c *cli.Context) error {
	log.Warn(deprecatedWarningWithExample(utils.Npm, os.Args[2], "npmc"))
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	npmPublicCmd := npm.NewNpmPublishCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	npmPublicArgs, err := coreutils.ParseArgs(strings.Split(c.String("npm-args"), " "))
	if err != nil {
		return err
	}
	npmPublicCmd.SetBuildConfiguration(buildConfiguration).SetRepo(c.Args().Get(0)).SetNpmArgs(npmPublicArgs).SetServerDetails(rtDetails)

	return commands.Exec(npmPublicCmd)
}

// This function checks whether the command received --help as a single option.
// If it did, the command's help is shown and true is returned.
// This function should be uesd iff the SkipFlagParsing option is used.
func showCmdHelpIfNeeded(c *cli.Context) (bool, error) {
	if len(c.Args()) != 1 {
		return false, nil
	}
	if c.Args()[0] == "--help" {
		err := cli.ShowCommandHelp(c, c.Command.Name)
		return true, err
	}
	return false, nil
}

func shouldSkipGoFlagParsing() bool {
	// This function is executed by code-gangsta, regardless of the CLI command being executed.
	// There's no need to run the code of this function, if the command is not "jfrog rt go*".
	if len(os.Args) < 3 || os.Args[2] != "go" {
		return false
	}

	_, exists, err := utils.GetProjectConfFilePath(utils.Go)
	if err != nil {
		coreutils.ExitOnErr(err)
	}
	return exists
}

func shouldSkipNpmFlagParsing() bool {
	// This function is executed by code-gangsta, regardless of the CLI command being executed.
	// There's no need to run the code of this function, if the command is not "jfrog rt npm*".
	if len(os.Args) < 3 || !npmUtils.IsNpmCommand(os.Args[2]) {
		return false
	}

	_, exists, err := utils.GetProjectConfFilePath(utils.Npm)
	if err != nil {
		coreutils.ExitOnErr(err)
	}
	return exists
}

func shouldSkipNugetFlagParsing() bool {
	// This function is executed by code-gangsta, regardless of the CLI command being executed.
	// There's no need to run the code of this function, if the command is not "jfrog rt nuget*".
	if len(os.Args) < 3 || os.Args[2] != "nuget" {
		return false
	}

	_, exists, err := utils.GetProjectConfFilePath(utils.Nuget)
	if err != nil {
		coreutils.ExitOnErr(err)
	}
	return exists
}

func shouldSkipMavenFlagParsing() bool {
	// This function is executed by code-gangsta, regardless of the CLI command being executed.
	// There's no need to run the code of this function, if the command is not "jfrog rt mvn*".
	if len(os.Args) < 3 || os.Args[2] != "mvn" {
		return false
	}
	_, exists, err := utils.GetProjectConfFilePath(utils.Maven)
	if err != nil {
		coreutils.ExitOnErr(err)
	}
	return exists
}

func shouldSkipGradleFlagParsing() bool {
	// This function is executed by code-gangsta, regardless of the CLI command being executed.
	// There's no need to run the code of this function, if the command is not "jfrog rt gradle*".
	if len(os.Args) < 3 || os.Args[2] != "gradle" {
		return false
	}
	_, exists, err := utils.GetProjectConfFilePath(utils.Gradle)
	if err != nil {
		coreutils.ExitOnErr(err)
	}
	return exists
}

func goCmd(c *cli.Context, goCmd func(*cli.Context, string) error, legacyGoCmd func(*cli.Context) error) error {
	if show, err := showCmdHelpIfNeeded(c); show || err != nil {
		return err
	}
	configFilePath, exists, err := utils.GetProjectConfFilePath(utils.Go)
	if err != nil {
		return err
	}
	if exists {
		log.Debug("Go config file was found in:", configFilePath)
		return goCmd(c, configFilePath)
	}
	log.Debug("Go config file wasn't found.")
	// If config file not found, use Go legacy command
	return legacyGoCmd(c)
}

func goLegacyCmd(c *cli.Context) error {
	log.Warn(deprecatedWarningWithExample(utils.Go, os.Args[2], "go-config"))
	// When the no-registry set to false (default), two arguments are mandatory: go command and the target repository
	if !c.Bool("no-registry") && c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	// When the no-registry is set to true this means that the resolution will not be done via Artifactory.
	// For automation purposes of users, keeping the possibility to pass the repository although we are not using it.
	if c.Bool("no-registry") && c.NArg() > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	goArg, err := coreutils.ParseArgs(strings.Split(c.Args().Get(0), " "))
	if err != nil {
		err = cliutils.PrintSummaryReport(0, 1, nil, "", err)
	}
	targetRepo := c.Args().Get(1)
	details, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	publishDeps := c.Bool("publish-deps")
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	resolverRepo := &utils.RepositoryConfig{}
	resolverRepo.SetTargetRepo(targetRepo).SetServerDetails(details)
	goCmd := golang.NewGoCommand().SetBuildConfiguration(buildConfiguration).
		SetGoArg(goArg).SetNoRegistry(c.Bool("no-registry")).
		SetPublishDeps(publishDeps).SetResolverParams(resolverRepo)
	if publishDeps {
		goCmd.SetDeployerParams(resolverRepo)
	}
	err = commands.Exec(goCmd)
	if err != nil {
		err = cliutils.PrintSummaryReport(0, 1, nil, "", err)
	}
	return err
}

func goPublishCmd(c *cli.Context, configFilePath string) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	version := c.Args().Get(0)
	goPublishCmd := golang.NewGoPublishCommand()
	goPublishCmd.SetConfigFilePath(configFilePath).SetBuildConfiguration(buildConfiguration).SetVersion(version).SetDependencies(c.String("deps"))
	err = commands.Exec(goPublishCmd)
	result := goPublishCmd.Result()
	return cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), nil, "", err)
}

func goLegacyPublishCmd(c *cli.Context) error {
	log.Warn(deprecatedWarning(os.Args[2], "go-config"))
	// When "self" set to true (default), there must be two arguments passed: target repo and the version
	if c.BoolT("self") && c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	// When "self" set to false, the target repository is mandatory but the version is not.
	// The version is only needed for publishing the project
	// But for automation purposes of users, keeping the possibility to pass the version without failing
	if !c.BoolT("self") && c.NArg() > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	targetRepo := c.Args().Get(0)
	version := c.Args().Get(1)
	details, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	goPublishCmd := golang.NewGoLegacyPublishCommand().SetPublishPackage(c.BoolT("self"))
	goPublishCmd.SetBuildConfiguration(buildConfiguration).SetVersion(version).SetDependencies(c.String("deps")).SetTargetRepo(targetRepo).SetServerDetails(details)
	err = commands.Exec(goPublishCmd)
	result := goPublishCmd.Result()
	return cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), nil, "", err)
}

func goNativeCmd(c *cli.Context, configFilePath string) error {
	// Found a config file. Continue as native command.
	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	args := cliutils.ExtractCommand(c)
	// Validates the go command. If a config file is found, the only flags that can be used are build-name, build-number and module.
	// Otherwise, throw an error.
	if err := validateCommand(args, cliutils.GetLegacyGoFlags()); err != nil {
		return err
	}
	goNative := golang.NewGoNativeCommand()
	goNative.SetConfigFilePath(configFilePath).SetGoArg(args)
	return commands.Exec(goNative)
}

func createGradleConfigCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commandUtils.CreateBuildConfig(c, utils.Gradle)
}

func createMvnConfigCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commandUtils.CreateBuildConfig(c, utils.Maven)
}

func createGoConfigCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commandUtils.CreateBuildConfig(c, utils.Go)
}

func createNpmConfigCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commandUtils.CreateBuildConfig(c, utils.Npm)
}

func createNugetConfigCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commandUtils.CreateBuildConfig(c, utils.Nuget)
}

func createDotnetConfigCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commandUtils.CreateBuildConfig(c, utils.Dotnet)
}

func createPipConfigCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commandUtils.CreateBuildConfig(c, utils.Pip)
}

func pingCmd(c *cli.Context) error {
	if c.NArg() > 0 {
		return cliutils.PrintHelpAndReturnError("No arguments should be sent.", c)
	}
	artDetails, err := createArtifactoryDetailsByFlags(c, false)
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

func downloadCmd(c *cli.Context) error {
	if c.NArg() > 0 && c.IsSet("spec") {
		return cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || c.NArg() == 2 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build") || c.IsSet("bundle")))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var downloadSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		downloadSpec, err = getSpec(c, true)
	} else {
		downloadSpec, err = createDefaultDownloadSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(downloadSpec.Files, false, true, false)
	if err != nil {
		return err
	}
	fixWinPathsForDownloadCmd(downloadSpec, c)
	configuration, err := createDownloadConfiguration(c)
	if err != nil {
		return err
	}
	serverDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	downloadCommand := generic.NewDownloadCommand()
	downloadCommand.SetConfiguration(configuration).SetBuildConfiguration(buildConfiguration).SetSpec(downloadSpec).SetServerDetails(serverDetails).SetDryRun(c.Bool("dry-run")).SetSyncDeletesPath(c.String("sync-deletes")).SetQuiet(cliutils.GetQuietValue(c)).SetDetailedSummary(c.Bool("detailed-summary"))

	if downloadCommand.ShouldPrompt() && !coreutils.AskYesNo("Sync-deletes may delete some files in your local file system. Are you sure you want to continue?\n"+
		"You can avoid this confirmation message by adding --quiet to the command.", false) {
		return nil
	}

	err = execWithProgress(downloadCommand)
	result := downloadCommand.Result()
	err = cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), result.Reader(), serverDetails.ArtifactoryUrl, err)

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
		uploadSpec, err = getFileSystemSpec(c)
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
	fixWinPathsForFileSystemSourcedCmds(uploadSpec, c)
	configuration, err := createUploadConfiguration(c)
	if err != nil {
		return err
	}
	buildConfiguration, err := createBuildConfigurationWithModule(c)
	if err != nil {
		return err
	}
	uploadCmd := generic.NewUploadCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	uploadCmd.SetUploadConfiguration(configuration).SetBuildConfiguration(buildConfiguration).SetSpec(uploadSpec).SetServerDetails(rtDetails).SetDryRun(c.Bool("dry-run")).SetSyncDeletesPath(c.String("sync-deletes")).SetQuiet(cliutils.GetQuietValue(c)).SetDetailedSummary(c.Bool("detailed-summary"))

	if uploadCmd.ShouldPrompt() && !coreutils.AskYesNo("Sync-deletes may delete some artifacts in Artifactory. Are you sure you want to continue?\n"+
		"You can avoid this confirmation message by adding --quiet to the command.", false) {
		return nil
	}
	err = execWithProgress(uploadCmd)
	result := uploadCmd.Result()
	err = cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), result.Reader(), "", err)

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

func moveCmd(c *cli.Context) error {
	if c.NArg() > 0 && c.IsSet("spec") {
		return cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var moveSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		moveSpec, err = getSpec(c, false)
	} else {
		moveSpec, err = createDefaultCopyMoveSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(moveSpec.Files, true, true, false)
	if err != nil {
		return err
	}
	moveCmd := generic.NewMoveCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	threads, err := getThreadsCount(c)
	if err != nil {
		return err
	}
	moveCmd.SetThreads(threads).SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails).SetSpec(moveSpec)
	err = commands.Exec(moveCmd)
	result := moveCmd.Result()
	err = cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), nil, "", err)

	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), isFailNoOp(c))
}

func copyCmd(c *cli.Context) error {
	if c.NArg() > 0 && c.IsSet("spec") {
		return cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var copySpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		copySpec, err = getSpec(c, false)
	} else {
		copySpec, err = createDefaultCopyMoveSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(copySpec.Files, true, true, false)
	if err != nil {
		return err
	}

	copyCommand := generic.NewCopyCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	threads, err := getThreadsCount(c)
	if err != nil {
		return err
	}
	copyCommand.SetThreads(threads).SetSpec(copySpec).SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails)
	err = commands.Exec(copyCommand)
	result := copyCommand.Result()
	err = cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), nil, "", err)

	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), isFailNoOp(c))
}

func deleteCmd(c *cli.Context) error {
	if c.NArg() > 0 && c.IsSet("spec") {
		return cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var deleteSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		deleteSpec, err = getSpec(c, false)
	} else {
		deleteSpec, err = createDefaultDeleteSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(deleteSpec.Files, false, true, false)
	if err != nil {
		return err
	}

	deleteCommand := generic.NewDeleteCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}

	threads, err := getThreadsCount(c)
	if err != nil {
		return err
	}
	deleteCommand.SetThreads(threads).SetQuiet(cliutils.GetQuietValue(c)).SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails).SetSpec(deleteSpec)
	err = commands.Exec(deleteCommand)
	result := deleteCommand.Result()
	err = cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), nil, "", err)

	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), isFailNoOp(c))
}

func searchCmd(c *cli.Context) error {
	if c.NArg() > 0 && c.IsSet("spec") {
		return cliutils.PrintHelpAndReturnError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var searchSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		searchSpec, err = getSpec(c, false)
	} else {
		searchSpec, err = createDefaultSearchSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(searchSpec.Files, false, true, false)
	if err != nil {
		return err
	}
	artDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(artDetails).SetSpec(searchSpec)
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
	if !(c.NArg() == 2 || (c.NArg() == 1 && c.IsSet("spec"))) {
		return nil, cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	var propsSpec *spec.SpecFiles
	var err error
	var props string
	if c.IsSet("spec") {
		props = c.Args()[0]
		propsSpec, err = getSpec(c, false)
	} else {
		props = c.Args()[1]
		propsSpec, err = createDefaultPropertiesSpec(c)
	}
	if err != nil {
		return nil, err
	}
	err = spec.ValidateSpec(propsSpec.Files, false, true, false)
	if err != nil {
		return nil, err
	}

	command := generic.NewPropsCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return nil, err
	}
	threads, err := getThreadsCount(c)
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

	propsCmd := generic.NewSetPropsCommand().SetPropsCommand(*cmd)
	err = commands.Exec(propsCmd)
	result := propsCmd.Result()
	err = cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), nil, "", err)

	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), isFailNoOp(c))
}

func deletePropsCmd(c *cli.Context) error {
	cmd, err := preparePropsCmd(c)
	if err != nil {
		return err
	}

	propsCmd := generic.NewDeletePropsCommand().DeletePropsCommand(*cmd)
	err = commands.Exec(propsCmd)
	result := propsCmd.Result()
	err = cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), nil, "", err)

	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), isFailNoOp(c))
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
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	buildPublishCmd := buildinfo.NewBuildPublishCommand().SetServerDetails(rtDetails).SetBuildConfiguration(buildConfiguration).SetConfig(buildInfoConfiguration)

	return commands.Exec(buildPublishCmd)
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
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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
		dependenciesSpec, err = getFileSystemSpec(c)
		if err != nil {
			return err
		}
	} else {
		dependenciesSpec = createDefaultBuildAddDependenciesSpec(c)
	}
	if c.Bool("from-rt") {
		rtDetails, err = createArtifactoryDetailsByFlags(c, false)
		if err != nil {
			return err
		}
	} else {
		fixWinPathsForFileSystemSourcedCmds(dependenciesSpec, c)
	}
	buildAddDependenciesCmd := buildinfo.NewBuildAddDependenciesCommand().SetDryRun(c.Bool("dry-run")).SetBuildConfiguration(buildConfiguration).SetDependenciesSpec(dependenciesSpec).SetServerDetails(rtDetails)
	err = commands.Exec(buildAddDependenciesCmd)
	result := buildAddDependenciesCmd.Result()
	err = cliutils.PrintSummaryReport(result.SuccessCount(), result.FailCount(), nil, "", err)
	if err != nil {
		return err
	}

	return cliutils.GetCliError(err, result.SuccessCount(), result.FailCount(), isFailNoOp(c))
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

func buildScanCmd(c *cli.Context) error {
	if c.NArg() > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
		return err
	}
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	buildScanCmd := buildinfo.NewBuildScanCommand().SetServerDetails(rtDetails).SetFailBuild(c.BoolT("fail")).SetBuildConfiguration(buildConfiguration)
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
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	buildPromotionCmd := buildinfo.NewBuildPromotionCommand().SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails).SetPromotionParams(configuration)

	return commands.Exec(buildPromotionCmd)
}

func buildDistributeCmd(c *cli.Context) error {
	if c.NArg() > 3 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	if err := validateBuildConfiguration(c, createBuildConfiguration(c)); err != nil {
		return err
	}
	configuration := createBuildDistributionConfiguration(c)
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	buildDistributeCmd := buildinfo.NewBuildDistributeCommnad().SetDryRun(c.Bool("dry-run")).SetServerDetails(rtDetails).SetBuildDistributionParams(configuration)

	return commands.Exec(buildDistributeCmd)
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
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	buildDiscardCmd.SetServerDetails(rtDetails).SetDiscardBuildsParams(configuration)

	return commands.Exec(buildDiscardCmd)
}

func releaseBundleCreateCmd(c *cli.Context) error {
	if !(c.NArg() == 2 && c.IsSet("spec") || (c.NArg() == 3 && !c.IsSet("spec"))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	var releaseBundleCreateSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		releaseBundleCreateSpec, err = getSpec(c, true)
	} else {
		releaseBundleCreateSpec = createDefaultReleaseBundleSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(releaseBundleCreateSpec.Files, false, true, false)
	if err != nil {
		return err
	}

	params, err := createReleaseBundleCreateUpdateParams(c, c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}
	releaseBundleCreateCmd := distribution.NewReleaseBundleCreateCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, true)
	if err != nil {
		return err
	}
	releaseBundleCreateCmd.SetServerDetails(rtDetails).SetReleaseBundleCreateParams(params).SetSpec(releaseBundleCreateSpec).SetDryRun(c.Bool("dry-run"))

	return commands.Exec(releaseBundleCreateCmd)
}

func releaseBundleUpdateCmd(c *cli.Context) error {
	if !(c.NArg() == 2 && c.IsSet("spec") || (c.NArg() == 3 && !c.IsSet("spec"))) {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	var releaseBundleUpdateSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		releaseBundleUpdateSpec, err = getSpec(c, true)
	} else {
		releaseBundleUpdateSpec = createDefaultReleaseBundleSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(releaseBundleUpdateSpec.Files, false, true, false)
	if err != nil {
		return err
	}

	params, err := createReleaseBundleCreateUpdateParams(c, c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}
	releaseBundleUpdateCmd := distribution.NewReleaseBundleUpdateCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, true)
	if err != nil {
		return err
	}
	releaseBundleUpdateCmd.SetServerDetails(rtDetails).SetReleaseBundleUpdateParams(params).SetSpec(releaseBundleUpdateSpec).SetDryRun(c.Bool("dry-run"))

	return commands.Exec(releaseBundleUpdateCmd)
}

func releaseBundleSignCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	params := distributionServices.NewSignBundleParams(c.Args().Get(0), c.Args().Get(1))
	params.StoringRepository = c.String("repo")
	params.GpgPassphrase = c.String("passphrase")
	releaseBundleSignCmd := distribution.NewReleaseBundleSignCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, true)
	if err != nil {
		return err
	}
	releaseBundleSignCmd.SetServerDetails(rtDetails).SetReleaseBundleSignParams(params)
	return commands.Exec(releaseBundleSignCmd)
}

func releaseBundleDistributeCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	if c.IsSet("max-wait-minutes") && !c.IsSet("sync") {
		return cliutils.PrintHelpAndReturnError("The --max-wait-minutes option can't be used without --sync", c)
	}
	var distributionRules *spec.DistributionRules
	if c.IsSet("dist-rules") {
		if c.IsSet("site") || c.IsSet("city") || c.IsSet("country-code") {
			return cliutils.PrintHelpAndReturnError("The --dist-rules option can't be used with --site, --city or --country-code", c)
		}
		var err error
		distributionRules, err = spec.CreateDistributionRulesFromFile(c.String("dist-rules"))
		if err != nil {
			return err
		}
	} else {
		distributionRules = createDefaultDistributionRules(c)
	}

	params := distributionServices.NewDistributeReleaseBundleParams(c.Args().Get(0), c.Args().Get(1))
	releaseBundleDistributeCmd := distribution.NewReleaseBundleDistributeCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, true)
	if err != nil {
		return err
	}
	maxWaitMinutes, err := cliutils.GetIntFlagValue(c, "max-wait-minutes", 0)
	if err != nil {
		return err
	}
	releaseBundleDistributeCmd.SetServerDetails(rtDetails).SetDistributeBundleParams(params).SetDistributionRules(distributionRules).SetDryRun(c.Bool("dry-run")).SetSync(c.Bool("sync")).SetMaxWaitMinutes(maxWaitMinutes)

	return commands.Exec(releaseBundleDistributeCmd)
}

func releaseBundleDeleteCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	var distributionRules *spec.DistributionRules
	if c.IsSet("dist-rules") {
		if c.IsSet("site") || c.IsSet("city") || c.IsSet("country-code") {
			return cliutils.PrintHelpAndReturnError("flag --dist-rules can't be used with --site, --city or --country-code", c)
		}
		var err error
		distributionRules, err = spec.CreateDistributionRulesFromFile(c.String("dist-rules"))
		if err != nil {
			return err
		}
	} else {
		distributionRules = createDefaultDistributionRules(c)
	}

	params := distributionServices.NewDeleteReleaseBundleParams(c.Args().Get(0), c.Args().Get(1))
	params.DeleteFromDistribution = c.BoolT("delete-from-dist")
	distributeBundleCmd := distribution.NewReleaseBundleDeleteParams()
	rtDetails, err := createArtifactoryDetailsByFlags(c, true)
	if err != nil {
		return err
	}
	distributeBundleCmd.SetQuiet(cliutils.GetQuietValue(c)).SetServerDetails(rtDetails).SetDistributeBundleParams(params).SetDistributionRules(distributionRules).SetDryRun(c.Bool("dry-run"))

	return commands.Exec(distributeBundleCmd)
}

func gitLfsCleanCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	configuration := createGitLfsCleanConfiguration(c)
	gitLfsCmd := generic.NewGitLfsCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	gitLfsCmd.SetConfiguration(configuration).SetServerDetails(rtDetails).SetDryRun(c.Bool("dry-run"))

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

func pipInstallCmd(c *cli.Context) error {
	if show, err := showCmdHelpIfNeeded(c); show || err != nil {
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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
	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	replicationDeleteCmd := replication.NewReplicationDeleteCommand()
	replicationDeleteCmd.SetRepoKey(c.Args().Get(0)).SetServerDetails(rtDetails).SetQuiet(cliutils.GetQuietValue(c))
	return commands.Exec(replicationDeleteCmd)
}

func permissionTargrtTemplateCmd(c *cli.Context) error {
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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
		return errorutils.CheckError(errors.New("an empty input file was provided"))
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	rtDetails, err := createArtifactoryDetailsByFlags(c, false)
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

	serverDetails, err := createArtifactoryDetailsByFlags(c, false)
	if err != nil {
		return err
	}
	// If the username is provided as an argument, then it is used when creating the token.
	// If not, then the configured username (or the the value of the --user option) is used.
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

func offerConfig(c *cli.Context) (*coreConfig.ServerDetails, error) {
	confirmed, err := cliutils.ShouldOfferConfig()
	if !confirmed || err != nil {
		return nil, err
	}
	details := createArtifactoryDetailsFromFlags(c)
	configCmd := coreCommonCommands.NewConfigCommand().SetDefaultDetails(details).SetInteractive(true).SetEncPassword(true)
	err = configCmd.Config()
	if err != nil {
		return nil, err
	}

	return configCmd.ServerDetails()
}

func createArtifactoryDetailsWithConfigOffer(c *cli.Context, excludeRefreshableTokens bool) (*coreConfig.ServerDetails, error) {
	createdDetails, err := offerConfig(c)
	if err != nil {
		return nil, err
	}
	if createdDetails != nil {
		return createdDetails, err
	}

	details := createArtifactoryDetailsFromFlags(c)
	// If urls or credentials were passed as options, use options as they are.
	// For security reasons, we'd like to avoid using part of the connection details from command options and the rest from the config.
	// Either use command options only or config only.
	if credentialsChanged(details) {
		return details, nil
	}

	// Else, use details from config for requested serverId, or for default server if empty.
	confDetails, err := coreCommonCommands.GetConfig(details.ServerId, excludeRefreshableTokens)
	if err != nil {
		return nil, err
	}

	// Take InsecureTls value from options since it is not saved in config.
	confDetails.InsecureTls = details.InsecureTls
	confDetails.Url = clientutils.AddTrailingSlashIfNeeded(confDetails.Url)
	confDetails.DistributionUrl = clientutils.AddTrailingSlashIfNeeded(confDetails.DistributionUrl)

	// Create initial access token if needed.
	if !excludeRefreshableTokens {
		err = coreConfig.CreateInitialRefreshableTokensIfNeeded(confDetails)
		if err != nil {
			return nil, err
		}
	}

	return confDetails, nil
}

func createArtifactoryDetailsFromFlags(c *cli.Context) (details *coreConfig.ServerDetails) {
	details = cliutils.CreateServerDetailsFromFlags(c)
	details.ArtifactoryUrl = details.Url
	details.Url = ""
	return
}

func credentialsChanged(details *coreConfig.ServerDetails) bool {
	return details.Url != "" || details.ArtifactoryUrl != "" || details.DistributionUrl != "" || details.User != "" || details.Password != "" ||
		details.ApiKey != "" || details.SshKeyPath != "" || details.SshPassphrase != "" || details.AccessToken != "" ||
		details.ClientCertKeyPath != "" || details.ClientCertPath != ""
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
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Flat(c.Bool("flat")).
		IncludeDirs(true).
		Target(c.Args().Get(1)).
		ArchiveEntries(c.String("archive-entries")).
		BuildSpec(), nil
}

func getSpec(c *cli.Context, isDownload bool) (specFiles *spec.SpecFiles, err error) {
	specFiles, err = spec.CreateSpecFromFile(c.String("spec"), coreutils.SpecVarsStringToMap(c.String("spec-vars")))
	if err != nil {
		return nil, err
	}
	// Override spec with CLI options
	for i := 0; i < len(specFiles.Files); i++ {
		if isDownload {
			specFiles.Get(i).Pattern = strings.TrimPrefix(specFiles.Get(i).Pattern, "/")
		}
		overrideFieldsIfSet(specFiles.Get(i), c)
	}
	return
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
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
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
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
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
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
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
	// Allow to use `env-exclude=""` and get no filters
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
	return discardParamsImpl
}

func createBuildDistributionConfiguration(c *cli.Context) services.BuildDistributionParams {
	distributeParamsImpl := services.NewBuildDistributionParams()
	distributeParamsImpl.Publish = c.BoolT("publish")
	distributeParamsImpl.OverrideExistingFiles = c.Bool("override")
	distributeParamsImpl.GpgPassphrase = c.String("passphrase")
	distributeParamsImpl.Async = c.Bool("async")
	distributeParamsImpl.SourceRepos = c.String("source-repos")
	distributeParamsImpl.BuildName, distributeParamsImpl.BuildNumber = utils.GetBuildNameAndNumber(c.Args().Get(0), c.Args().Get(1))
	distributeParamsImpl.TargetRepo = c.Args().Get(2)
	return distributeParamsImpl
}

func createReleaseBundleCreateUpdateParams(c *cli.Context, bundleName, bundleVersion string) (distributionServicesUtils.ReleaseBundleParams, error) {
	releaseBundleParams := distributionServicesUtils.NewReleaseBundleParams(bundleName, bundleVersion)
	releaseBundleParams.SignImmediately = c.Bool("sign")
	releaseBundleParams.StoringRepository = c.String("repo")
	releaseBundleParams.GpgPassphrase = c.String("passphrase")
	releaseBundleParams.Description = c.String("desc")
	if c.IsSet("release-notes-path") {
		bytes, err := ioutil.ReadFile(c.String("release-notes-path"))
		if err != nil {
			return releaseBundleParams, errorutils.CheckError(err)
		}
		releaseBundleParams.ReleaseNotes = string(bytes)
		releaseBundleParams.ReleaseNotesSyntax, err = populateReleaseNotesSyntax(c)
		if err != nil {
			return releaseBundleParams, err
		}
	}
	return releaseBundleParams, nil
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
		ExcludeArtifacts(c.Bool("exclude-artifacts")).
		IncludeDeps(c.Bool("include-deps")).
		Bundle(c.String("bundle")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
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
	downloadConfiguration.Threads, err = getThreadsCount(c)
	if err != nil {
		return nil, err
	}
	downloadConfiguration.Retries, err = getRetries(c)
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
		Build(c.String("build")).
		Offset(offset).
		Limit(limit).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Flat(c.BoolT("flat")).
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
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Regexp(c.Bool("regexp")).
		Ant(c.Bool("ant")).
		BuildSpec()
}

func createDefaultReleaseBundleSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(2)).
		Target(c.String("target")).
		Props(c.String("props")).
		Build(c.String("build")).
		Bundle(c.String("bundle")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Regexp(c.Bool("regexp")).
		TargetProps(c.String("target-props")).
		Ant(c.Bool("ant")).
		BuildSpec()
}

func createDefaultDistributionRules(c *cli.Context) *spec.DistributionRules {
	return &spec.DistributionRules{
		DistributionRules: []spec.DistributionRule{{
			SiteName:     c.String("site"),
			CityName:     c.String("city"),
			CountryCodes: cliutils.GetStringsArrFlagValue(c, "country-codes"),
		}},
	}
}

func getFileSystemSpec(c *cli.Context) (fsSpec *spec.SpecFiles, err error) {
	fsSpec, err = spec.CreateSpecFromFile(c.String("spec"), coreutils.SpecVarsStringToMap(c.String("spec-vars")))
	if err != nil {
		return
	}
	// Override spec with CLI options
	for i := 0; i < len(fsSpec.Files); i++ {
		fsSpec.Get(i).Target = strings.TrimPrefix(fsSpec.Get(i).Target, "/")
		overrideFieldsIfSet(fsSpec.Get(i), c)
	}
	return
}

func fixWinPathsForFileSystemSourcedCmds(uploadSpec *spec.SpecFiles, c *cli.Context) {
	if coreutils.IsWindows() {
		for i, file := range uploadSpec.Files {
			uploadSpec.Files[i].Pattern = fixWinPathBySource(file.Pattern, c.IsSet("spec"))
			for j, exclusion := range uploadSpec.Files[i].Exclusions {
				// If exclusions are set, they override the spec value
				uploadSpec.Files[i].Exclusions[j] = fixWinPathBySource(exclusion, c.IsSet("spec") && !c.IsSet("exclusions"))
			}
			for j, excludePattern := range uploadSpec.Files[i].ExcludePatterns {
				// If exclude patterns are set, they override the spec value
				uploadSpec.Files[i].ExcludePatterns[j] = fixWinPathBySource(excludePattern, c.IsSet("spec") && !c.IsSet("exclude-patterns"))
			}
		}
	}
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
	uploadConfiguration.Retries, err = getRetries(c)
	if err != nil {
		return nil, err
	}
	uploadConfiguration.Threads, err = getThreadsCount(c)
	if err != nil {
		return nil, err
	}
	uploadConfiguration.Deb, err = getDebFlag(c)
	if err != nil {
		return
	}
	return
}

func createBuildConfigurationWithModule(c *cli.Context) (buildConfigConfiguration *utils.BuildConfiguration, err error) {
	buildConfigConfiguration = new(utils.BuildConfiguration)
	buildConfigConfiguration.BuildName, buildConfigConfiguration.BuildNumber = utils.GetBuildNameAndNumber(c.String("build-name"), c.String("build-number"))
	buildConfigConfiguration.Project = utils.GetBuildProject(c.String("project"))
	buildConfigConfiguration.Module = c.String("module")
	err = utils.ValidateBuildAndModuleParams(buildConfigConfiguration)
	return
}

func validateConfigFlags(configCommandConfiguration *coreCommonCommands.ConfigCommandConfiguration) error {
	if !configCommandConfiguration.Interactive && configCommandConfiguration.ServerDetails.ArtifactoryUrl == "" {
		return errors.New("the --url option is mandatory when the --interactive option is set to false or the CI environment variable is set to true.")
	}
	// Validate the option is not used along with an access token
	if configCommandConfiguration.BasicAuthOnly && configCommandConfiguration.ServerDetails.AccessToken != "" {
		return errors.New("the --basic-auth-only option is only supported when username and password/API key are provided")
	}
	return nil
}

// If `fieldName` exist in the cli args, read it to `field` as a string.
func overrideStringIfSet(field *string, c *cli.Context, fieldName string) {
	if c.IsSet(fieldName) {
		*field = c.String(fieldName)
	}
}

// If `fieldName` exist in the cli args, read it to `field` as an array split by `;`.
func overrideArrayIfSet(field *[]string, c *cli.Context, fieldName string) {
	if c.IsSet(fieldName) {
		*field = nil
		for _, singleValue := range strings.Split(c.String(fieldName), ";") {
			*field = append(*field, singleValue)
		}
	}
}

// If `fieldName` exist in the cli args, read it to `field` as a int.
func overrideIntIfSet(field *int, c *cli.Context, fieldName string) {
	if c.IsSet(fieldName) {
		*field = c.Int(fieldName)
	}
}

func overrideFieldsIfSet(spec *spec.File, c *cli.Context) {
	overrideArrayIfSet(&spec.ExcludePatterns, c, "exclude-patterns")
	overrideArrayIfSet(&spec.Exclusions, c, "exclusions")
	overrideArrayIfSet(&spec.SortBy, c, "sort-by")
	overrideIntIfSet(&spec.Offset, c, "offset")
	overrideIntIfSet(&spec.Limit, c, "limit")
	overrideStringIfSet(&spec.SortOrder, c, "sort-order")
	overrideStringIfSet(&spec.Props, c, "props")
	overrideStringIfSet(&spec.TargetProps, c, "target-props")
	overrideStringIfSet(&spec.ExcludeProps, c, "exclude-props")
	overrideStringIfSet(&spec.Build, c, "build")
	overrideStringIfSet(&spec.ExcludeArtifacts, c, "exclude-artifacts")
	overrideStringIfSet(&spec.IncludeDeps, c, "include-deps")
	overrideStringIfSet(&spec.Bundle, c, "bundle")
	overrideStringIfSet(&spec.Recursive, c, "recursive")
	overrideStringIfSet(&spec.Flat, c, "flat")
	overrideStringIfSet(&spec.Explode, c, "explode")
	overrideStringIfSet(&spec.Regexp, c, "regexp")
	overrideStringIfSet(&spec.IncludeDirs, c, "include-dirs")
	overrideStringIfSet(&spec.ValidateSymlinks, c, "validate-symlinks")
	overrideStringIfSet(&spec.Symlinks, c, "symlinks")
	overrideStringIfSet(&spec.Transitive, c, "transitive")
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

func deprecatedWarning(command, configCommand string) string {
	return `You are using a deprecated syntax of the "` + command + `" command.
	To use the new syntax, the command expects the details of the Artifactory server and repositories to be pre-configured.
	To create this configuration, run the following command from the root directory of the project:
	$ jfrog rt ` + configCommand + `
	This will create the configuration inside the .jfrog directory under the root directory of the project.`
}

func deprecatedWarningWithExample(projectType utils.ProjectType, command, configCommand string) string {
	return deprecatedWarning(command, configCommand) + `The new command syntax looks very similar to the ` + projectType.String() + ` CLI command i.e.:
	$ jfrog rt ` + command + ` [` + projectType.String() + ` args and option] --build-name=*BUILD_NAME* --build-number=*BUILD_NUMBER*`
}

func extractThreadsFlag(args []string) (cleanArgs []string, threadsCount int, err error) {
	// Extract threads flag.
	cleanArgs = append([]string(nil), args...)
	threadsFlagIndex, threadsValueIndex, threads, err := coreutils.FindFlag("--threads", cleanArgs)
	if err != nil || threadsFlagIndex < 0 {
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

func populateReleaseNotesSyntax(c *cli.Context) (distributionServicesUtils.ReleaseNotesSyntax, error) {
	// If release notes syntax is set, use it
	releaseNotexSyntax := c.String("release-notes-syntax")
	if releaseNotexSyntax != "" {
		switch releaseNotexSyntax {
		case "markdown":
			return distributionServicesUtils.Markdown, nil
		case "asciidoc":
			return distributionServicesUtils.Asciidoc, nil
		case "plain_text":
			return distributionServicesUtils.PlainText, nil
		default:
			return distributionServicesUtils.PlainText, errorutils.CheckError(errors.New("--release-notes-syntax must be one of: markdown, asciidoc or plain_text."))
		}
	}
	// If the file extension is ".md" or ".markdown", use the markdonwn syntax
	extension := strings.ToLower(filepath.Ext(c.String("release-notes-path")))
	if extension == ".md" || extension == ".markdown" {
		return distributionServicesUtils.Markdown, nil
	}
	return distributionServicesUtils.PlainText, nil
}

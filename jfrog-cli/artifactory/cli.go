package artifactory

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/buildinfo"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/docker"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/golang"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/gradle"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/mvn"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/commands/nuget"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	npmutils "github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/npm"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/buildadddependencies"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/buildaddgit"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/buildclean"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/buildcollectenv"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/builddiscard"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/builddistribute"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/buildpromote"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/buildpublish"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/buildscan"
	configdocs "github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/copy"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/delete"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/deleteprops"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/dockerpull"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/dockerpush"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/download"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/gitlfsclean"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/gocommand"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/gopublish"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/gorecursivepublish"
	gradledoc "github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/gradle"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/gradleconfig"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/move"
	mvndoc "github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/mvn"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/mvnconfig"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/npminstall"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/npmpublish"
	nugetdocs "github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/nuget"
	nugettree "github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/nugetdepstree"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/ping"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/search"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/setprops"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/upload"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/artifactory/use"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	buildinfocmd "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	rtclientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mattn/go-shellwords"
	"strconv"
	"strings"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "config",
			Flags:     getConfigFlags(),
			Aliases:   []string{"c"},
			Usage:     configdocs.Description,
			HelpName:  common.CreateUsage("rt config", configdocs.Description, configdocs.Usage),
			UsageText: configdocs.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				configCmd(c)
			},
		},
		{
			Name:      "use",
			Usage:     use.Description,
			HelpName:  common.CreateUsage("rt use", use.Description, use.Usage),
			UsageText: use.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				useCmd(c)
			},
		},
		{
			Name:      "upload",
			Flags:     getUploadFlags(),
			Aliases:   []string{"u"},
			Usage:     upload.Description,
			HelpName:  common.CreateUsage("rt upload", upload.Description, upload.Usage),
			UsageText: upload.Arguments,
			ArgsUsage: common.CreateEnvVars(upload.EnvVar),
			Action: func(c *cli.Context) {
				uploadCmd(c)
			},
		},
		{
			Name:      "download",
			Flags:     getDownloadFlags(),
			Aliases:   []string{"dl"},
			Usage:     download.Description,
			HelpName:  common.CreateUsage("rt download", download.Description, download.Usage),
			UsageText: download.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				downloadCmd(c)
			},
		},
		{
			Name:      "move",
			Flags:     getMoveFlags(),
			Aliases:   []string{"mv"},
			Usage:     move.Description,
			HelpName:  common.CreateUsage("rt move", move.Description, move.Usage),
			UsageText: move.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				moveCmd(c)
			},
		},
		{
			Name:      "copy",
			Flags:     getCopyFlags(),
			Aliases:   []string{"cp"},
			Usage:     copy.Description,
			HelpName:  common.CreateUsage("rt copy", copy.Description, copy.Usage),
			UsageText: copy.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				copyCmd(c)
			},
		},
		{
			Name:      "delete",
			Flags:     getDeleteFlags(),
			Aliases:   []string{"del"},
			Usage:     delete.Description,
			HelpName:  common.CreateUsage("rt delete", delete.Description, delete.Usage),
			UsageText: delete.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				deleteCmd(c)
			},
		},
		{
			Name:      "search",
			Flags:     getSearchFlags(),
			Aliases:   []string{"s"},
			Usage:     search.Description,
			HelpName:  common.CreateUsage("rt search", search.Description, search.Usage),
			UsageText: search.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				searchCmd(c)
			},
		},
		{
			Name:      "set-props",
			Flags:     getSetPropertiesFlags(),
			Aliases:   []string{"sp"},
			Usage:     setprops.Description,
			HelpName:  common.CreateUsage("rt set-props", setprops.Description, setprops.Usage),
			UsageText: setprops.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				setPropsCmd(c)
			},
		},
		{
			Name:      "delete-props",
			Flags:     getDeletePropertiesFlags(),
			Aliases:   []string{"delp"},
			Usage:     deleteprops.Description,
			HelpName:  common.CreateUsage("rt delete-props", deleteprops.Description, deleteprops.Usage),
			UsageText: deleteprops.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				deletePropsCmd(c)
			},
		},
		{
			Name:      "build-publish",
			Flags:     getBuildPublishFlags(),
			Aliases:   []string{"bp"},
			Usage:     buildpublish.Description,
			HelpName:  common.CreateUsage("rt build-publish", buildpublish.Description, buildpublish.Usage),
			UsageText: buildpublish.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildPublishCmd(c)
			},
		},
		{
			Name:      "build-collect-env",
			Flags:     []cli.Flag{},
			Aliases:   []string{"bce"},
			Usage:     buildcollectenv.Description,
			HelpName:  common.CreateUsage("rt build-collect-env", buildcollectenv.Description, buildcollectenv.Usage),
			UsageText: buildcollectenv.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildCollectEnvCmd(c)
			},
		},
		{
			Name:      "build-add-dependencies",
			Flags:     getBuildAddDependenciesFlags(),
			Aliases:   []string{"bad"},
			Usage:     buildadddependencies.Description,
			HelpName:  common.CreateUsage("rt build-add-dependencies", buildadddependencies.Description, buildadddependencies.Usage),
			UsageText: buildadddependencies.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildAddDependenciesCmd(c)
			},
		},
		{
			Name:      "build-add-git",
			Flags:     getBuildAddGitFlags(),
			Aliases:   []string{"bag"},
			Usage:     buildaddgit.Description,
			HelpName:  common.CreateUsage("rt build-add-git", buildaddgit.Description, buildaddgit.Usage),
			UsageText: buildaddgit.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildAddGitCmd(c)
			},
		},
		{
			Name:      "build-scan",
			Flags:     getBuildScanFlags(),
			Aliases:   []string{"bs"},
			Usage:     buildscan.Description,
			HelpName:  common.CreateUsage("rt build-scan", buildscan.Description, buildscan.Usage),
			UsageText: buildscan.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildScanCmd(c)
			},
		},
		{
			Name:      "build-clean",
			Flags:     []cli.Flag{},
			Aliases:   []string{"bc"},
			Usage:     buildclean.Description,
			HelpName:  common.CreateUsage("rt build-clean", buildclean.Description, buildclean.Usage),
			UsageText: buildclean.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildCleanCmd(c)
			},
		},
		{
			Name:      "build-promote",
			Flags:     getBuildPromotionFlags(),
			Aliases:   []string{"bpr"},
			Usage:     buildpromote.Description,
			HelpName:  common.CreateUsage("rt build-promote", buildpromote.Description, buildpromote.Usage),
			UsageText: buildpromote.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildPromoteCmd(c)
			},
		},
		{
			Name:      "build-distribute",
			Flags:     getBuildDistributeFlags(),
			Aliases:   []string{"bd"},
			Usage:     builddistribute.Description,
			HelpName:  common.CreateUsage("rt build-distribute", builddistribute.Description, builddistribute.Usage),
			UsageText: builddistribute.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildDistributeCmd(c)
			},
		},
		{
			Name:      "build-discard",
			Flags:     getBuildDiscardFlags(),
			Aliases:   []string{"bdi"},
			Usage:     builddiscard.Description,
			HelpName:  common.CreateUsage("rt build-discard", builddiscard.Description, builddiscard.Usage),
			UsageText: builddiscard.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				buildDiscardCmd(c)
			},
		},
		{
			Name:      "git-lfs-clean",
			Flags:     getGitLfsCleanFlags(),
			Aliases:   []string{"glc"},
			Usage:     gitlfsclean.Description,
			HelpName:  common.CreateUsage("rt git-lfs-clean", gitlfsclean.Description, gitlfsclean.Usage),
			UsageText: gitlfsclean.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				gitLfsCleanCmd(c)
			},
		},
		{
			Name:      "mvn",
			Flags:     getBuildToolFlags(),
			Usage:     mvndoc.Description,
			HelpName:  common.CreateUsage("rt mvn", mvndoc.Description, mvndoc.Usage),
			UsageText: mvndoc.Arguments,
			ArgsUsage: common.CreateEnvVars(mvndoc.EnvVar),
			Action: func(c *cli.Context) {
				mvnCmd(c)
			},
		},
		{
			Name:      "mvn-config",
			Aliases:   []string{"mvnc"},
			Usage:     mvnconfig.Description,
			HelpName:  common.CreateUsage("rt mvn-config", mvnconfig.Description, mvnconfig.Usage),
			UsageText: mvnconfig.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				createMvnConfigCmd(c)
			},
		},
		{
			Name:      "gradle",
			Flags:     getBuildToolFlags(),
			Usage:     gradledoc.Description,
			HelpName:  common.CreateUsage("rt gradle", gradledoc.Description, gradledoc.Usage),
			UsageText: gradledoc.Arguments,
			ArgsUsage: common.CreateEnvVars(gradledoc.EnvVar),
			Action: func(c *cli.Context) {
				gradleCmd(c)
			},
		},
		{
			Name:      "gradle-config",
			Aliases:   []string{"gradlec"},
			Usage:     gradleconfig.Description,
			HelpName:  common.CreateUsage("rt gradle-config", gradleconfig.Description, gradleconfig.Usage),
			UsageText: gradleconfig.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				createGradleConfigCmd(c)
			},
		},
		{
			Name:      "docker-push",
			Flags:     getDockerPushFlags(),
			Aliases:   []string{"dp"},
			Usage:     dockerpush.Description,
			HelpName:  common.CreateUsage("rt docker-push", dockerpush.Description, dockerpush.Usage),
			UsageText: dockerpush.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				dockerPushCmd(c)
			},
		},
		{
			Name:      "docker-pull",
			Flags:     getDockerPullFlags(),
			Aliases:   []string{"dpl"},
			Usage:     dockerpull.Description,
			HelpName:  common.CreateUsage("rt docker-pull", dockerpull.Description, dockerpull.Usage),
			UsageText: dockerpull.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				dockerPullCmd(c)
			},
		},
		{
			Name:      "npm-install",
			Flags:     getNpmFlags(),
			Aliases:   []string{"npmi"},
			Usage:     npminstall.Description,
			HelpName:  common.CreateUsage("rt npm-install", npminstall.Description, npminstall.Usage),
			UsageText: npminstall.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				npmInstallCmd(c)
			},
		},
		{
			Name:      "npm-publish",
			Flags:     getNpmFlags(),
			Aliases:   []string{"npmp"},
			Usage:     npmpublish.Description,
			HelpName:  common.CreateUsage("rt npm-publish", npmpublish.Description, npmpublish.Usage),
			UsageText: npmpublish.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				npmPublishCmd(c)
			},
		},
		{
			Name:      "nuget",
			Flags:     getNugetFlags(),
			Usage:     nugetdocs.Description,
			HelpName:  common.CreateUsage("rt nuget", nugetdocs.Description, nugetdocs.Usage),
			UsageText: nugetdocs.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				nugetCmd(c)
			},
		},
		{
			Name:      "nuget-deps-tree",
			Aliases:   []string{"ndt"},
			Usage:     nugettree.Description,
			HelpName:  common.CreateUsage("rt nuget-deps-tree", nugettree.Description, nugettree.Usage),
			UsageText: nugettree.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				nugetDepsTreeCmd(c)
			},
		},
		{
			Name:      "go-publish",
			Flags:     getGoPublishFlags(),
			Aliases:   []string{"gp"},
			Usage:     gopublish.Description,
			HelpName:  common.CreateUsage("rt go-publish", gopublish.Description, gopublish.Usage),
			UsageText: gopublish.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				goPublishCmd(c)
			},
		},
		{
			Name:      "go",
			Flags:     getGoFlags(),
			Aliases:   []string{"go"},
			Usage:     gocommand.Description,
			HelpName:  common.CreateUsage("rt go", gocommand.Description, gocommand.Usage),
			UsageText: gocommand.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) error {
				return goCmd(c)
			},
		},
		{
			Name:      "go-recursive-publish",
			Flags:     getGoRecursivePublishFlags(),
			Aliases:   []string{"grp"},
			Usage:     gorecursivepublish.Description,
			HelpName:  common.CreateUsage("rt grp", gorecursivepublish.Description, gorecursivepublish.Usage),
			UsageText: gorecursivepublish.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				goRecursivePublishCmd(c)
			},
		},
		{
			Name:      "ping",
			Flags:     getServerFlags(),
			Aliases:   []string{"p"},
			Usage:     ping.Description,
			HelpName:  common.CreateUsage("rt ping", ping.Description, ping.Usage),
			UsageText: ping.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				pingCmd(c)
			},
		},
	}
}

func getUrlFlag() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Usage: "[Optional] Artifactory URL.` `",
		},
	}
}

func getBaseFlags() []cli.Flag {
	return append(getUrlFlag(),
		cli.StringFlag{
			Name:  "user",
			Usage: "[Optional] Artifactory username.` `",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "[Optional] Artifactory password.` `",
		},
		cli.StringFlag{
			Name:  "apikey",
			Usage: "[Optional] Artifactory API key.` `",
		},
		cli.StringFlag{
			Name:  "access-token",
			Usage: "[Optional] Artifactory access token.` `",
		})
}

func getCommonFlags() []cli.Flag {
	flags := append(getBaseFlags(),
		cli.StringFlag{
			Name:  "ssh-passphrase",
			Usage: "[Optional] SSH key passphrase.` `",
		})
	return append(flags, getSshKeyPathFlag()...)
}

func getServerFlags() []cli.Flag {
	return append(getCommonFlags(), getServerIdFlag())
}

func getSortLimitFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "sort-by",
			Usage: "[Optional] A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information, see https://www.jfrog.com/confluence/display/RTF/Artifactory+Query+Language#ArtifactoryQueryLanguage-EntitiesandFields` `",
		},
		cli.StringFlag{
			Name:  "sort-order",
			Usage: "[Default: asc] The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.` `",
		},
		cli.StringFlag{
			Name:  "limit",
			Usage: "[Optional] The maximum number of items to fetch. Usually used with the 'sort-by' option.` `",
		},
		cli.StringFlag{
			Name:  "offset",
			Usage: "[Optional] The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.` `",
		},
	}
}

func getUploadFlags() []cli.Flag {
	uploadFlags := append(getServerFlags(), getSpecFlags()...)
	return append(uploadFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "build-name",
			Usage: "[Optional] Build name. Providing this option will record all uploaded artifacts for later build info publication.` `",
		},
		cli.StringFlag{
			Name:  "build-number",
			Usage: "[Optional] Build number. Providing this option will record all uploaded artifacts for later build info publication.` `",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\" to be attached to the uploaded artifacts.` `",
		},
		cli.StringFlag{
			Name:  "deb",
			Usage: "[Optional] Used for Debian packages in the form of distribution/component/architecture. If the the value for distribution, component or architecture include a slash, the slash should be escaped with a back-slash.` `",
		},
		cli.BoolTFlag{
			Name:  "recursive",
			Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be uploaded to Artifactory.` `",
		},
		cli.BoolTFlag{
			Name:  "flat",
			Usage: "[Default: true] If set to false, files are uploaded according to their file system hierarchy.` `",
		},
		cli.BoolFlag{
			Name:  "regexp",
			Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to upload.` `",
		},
		cli.StringFlag{
			Name:  "retries",
			Usage: "[Default: " + strconv.Itoa(cliutils.Retries) + "] Number of upload retries.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "explode",
			Usage: "[Default: false] Set to true to extract an archive after it is deployed to Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "symlinks",
			Usage: "[Default: false] Set to true to preserve symbolic links structure in Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "include-dirs",
			Usage: "[Default: false] Set to true if you'd like to also apply the source path pattern for directories and not just for files.` `",
		},
		getFailNoOpFlag(),
		getExcludePatternsFlag(),
		getThreadsFlag(),
	}...)
}

func getDownloadFlags() []cli.Flag {
	downloadFlags := append(getServerFlags(), getSortLimitFlags()...)
	downloadFlags = append(downloadFlags, getSpecFlags()...)
	return append(downloadFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "build-name",
			Usage: "[Optional] Build name. Providing this option will record all downloaded artifacts for later build info publication.` `",
		},
		cli.StringFlag{
			Name:  "build-number",
			Usage: "[Optional] Build number. Providing this option will record all downloaded artifacts for later build info publication.` `",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be downloaded.` `",
		},
		cli.BoolTFlag{
			Name:  "recursive",
			Usage: "[Default: true] Set to false if you do not wish to include the download of artifacts inside sub-folders in Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "flat",
			Usage: "[Default: false] Set to true if you do not wish to have the Artifactory repository path structure created locally for your downloaded files.` `",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.` `",
		},
		cli.StringFlag{
			Name:  "min-split",
			Value: "",
			Usage: "[Default: " + strconv.Itoa(cliutils.DownloadMinSplitKb) + "] Minimum file size in KB to split into ranges when downloading. Set to -1 for no splits.` `",
		},
		cli.StringFlag{
			Name:  "split-count",
			Value: "",
			Usage: "[Default: " + strconv.Itoa(cliutils.DownloadSplitCount) + "] Number of parts to split a file when downloading. Set to 0 for no splits.` `",
		},
		cli.StringFlag{
			Name:  "retries",
			Usage: "[Default: " + strconv.Itoa(cliutils.Retries) + "] Number of download retries.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "explode",
			Usage: "[Default: false] Set to true to extract an archive after it is downloaded from Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "validate-symlinks",
			Usage: "[Default: false] Set to true to perform a checksum validation when downloading symbolic links.` `",
		},
		cli.BoolFlag{
			Name:  "include-dirs",
			Usage: "[Default: false] Set to true if you'd like to also apply the target path pattern for folders and not just for files in Artifactory.` `",
		},
		getFailNoOpFlag(),
		getExcludePatternsFlag(),
		getThreadsFlag(),
		getArchiveEntriesFlag(),
	}...)
}

func getBuildToolFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "build-name",
			Usage: "[Optional] Providing this option will collect and record build info for this build name.` `",
		},
		cli.StringFlag{
			Name:  "build-number",
			Usage: "[Optional] Providing this option will collect and record build info for this build number. If you provide a build name (using the --build-name option) and do not provide a build number, a build number will be automatically generated.` `",
		},
	}
}

func getServerIdFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "server-id",
		Usage: "[Optional] Artifactory server ID configured using the config command.` `",
	}
}

func getFailNoOpFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  "fail-no-op",
		Usage: "[Default: false] Set to true if you'd like the command to return exit code 2 in case of no files are affected.` `",
	}
}

func getExcludePatternsFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "exclude-patterns",
		Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.` `",
	}
}

func getSpecFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a File Spec.` `",
		},
		cli.StringFlag{
			Name:  "spec-vars",
			Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.` `",
		},
	}
}

func getDockerPushFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, getBuildToolFlags()...)
	flags = append(flags, getServerFlags()...)
	flags = append(flags, getThreadsFlag())
	return flags
}

func getDockerPullFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, getBuildToolFlags()...)
	flags = append(flags, getServerFlags()...)
	return flags
}

func getNpmFlags() []cli.Flag {
	npmFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "npm-args",
			Usage: "[Optional] A list of npm arguments and options in the form of \"--arg1=value1 --arg2=value2\"` `",
		},
	}
	npmFlags = append(npmFlags, getBaseFlags()...)
	npmFlags = append(npmFlags, getServerIdFlag())
	return append(npmFlags, getBuildToolFlags()...)
}

func getNugetFlags() []cli.Flag {
	nugetFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "nuget-args",
			Usage: "[Optional] A list of NuGet arguments and options in the form of \"arg1 value1 arg2 value2\"` `",
		},
		cli.StringFlag{
			Name:  "solution-root",
			Usage: "[Optional] Path to the solution.` `",
		},
	}
	nugetFlags = append(nugetFlags, getBaseFlags()...)
	nugetFlags = append(nugetFlags, getServerIdFlag())
	return append(nugetFlags, getBuildToolFlags()...)
}

func getGoFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolFlag{
			Name:  "no-registry",
			Usage: "[Default: false] Set to true if you don't want to use Artifactory as your proxy` `",
		},
	}
	flags = append(flags, getBaseFlags()...)
	flags = append(flags, getServerIdFlag())
	flags = append(flags, getBuildToolFlags()...)
	return flags
}

func getGoRecursivePublishFlags() []cli.Flag {
	return append(getBaseFlags(), getServerIdFlag())
}

func getGoPublishFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "deps",
			Value: "",
			Usage: "[Optional] List of project dependencies in the form of \"dep1-name:version,dep2-name:version...\" to be published to Artifactory. Use \"ALL\" to publish all dependencies.` `",
		},
		cli.BoolTFlag{
			Name:  "self",
			Usage: "[Default: true] Set false to skip publishing the project package zip file to Artifactory..` `",
		},
	}
	flags = append(flags, getBaseFlags()...)
	flags = append(flags, getServerIdFlag())
	flags = append(flags, getBuildToolFlags()...)
	return flags
}

func getMoveFlags() []cli.Flag {
	moveFlags := append(getServerFlags(), getSortLimitFlags()...)
	moveFlags = append(moveFlags, getSpecFlags()...)
	return append(moveFlags, []cli.Flag{
		cli.BoolTFlag{
			Name:  "recursive",
			Usage: "[Default: true] Set to false if you do not wish to move artifacts inside sub-folders in Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "flat",
			Usage: "[Default: false] If set to false, files are moved according to their file system hierarchy.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.` `",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be moved.` `",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.` `",
		},
		getFailNoOpFlag(),
		getExcludePatternsFlag(),
		getArchiveEntriesFlag(),
	}...)

}

func getCopyFlags() []cli.Flag {
	copyFlags := append(getServerFlags(), getSortLimitFlags()...)
	copyFlags = append(copyFlags, getSpecFlags()...)
	return append(copyFlags, []cli.Flag{
		cli.BoolTFlag{
			Name:  "recursive",
			Usage: "[Default: true] Set to false if you do not wish to copy artifacts inside sub-folders in Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "flat",
			Usage: "[Default: false] If set to false, files are copied according to their file system hierarchy.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.` `",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be copied.` `",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.` `",
		},
		getFailNoOpFlag(),
		getExcludePatternsFlag(),
		getArchiveEntriesFlag(),
	}...)
}

func getDeleteFlags() []cli.Flag {
	deleteFlags := append(getServerFlags(), getSortLimitFlags()...)
	deleteFlags = append(deleteFlags, getSpecFlags()...)
	return append(deleteFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be deleted.` `",
		},
		cli.BoolTFlag{
			Name:  "recursive",
			Usage: "[Default: true] Set to false if you do not wish to delete artifacts inside sub-folders in Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "[Default: false] Set to true to skip the delete confirmation message.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.` `",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.` `",
		},
		getFailNoOpFlag(),
		getExcludePatternsFlag(),
		getArchiveEntriesFlag(),
	}...)
}

func getSearchFlags() []cli.Flag {
	searchFlags := append(getServerFlags(), getSortLimitFlags()...)
	searchFlags = append(searchFlags, getSpecFlags()...)
	return append(searchFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be returned.` `",
		},
		cli.BoolTFlag{
			Name:  "recursive",
			Usage: "[Default: true] Set to false if you do not wish to search artifacts inside sub-folders in Artifactory.` `",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.` `",
		},
		getFailNoOpFlag(),
		getExcludePatternsFlag(),
		getArchiveEntriesFlag(),
	}...)
}

func getSetPropertiesFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties are affected.` `",
		},
	}
	return append(flags, getPropertiesFlags()...)
}

func getDeletePropertiesFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1,key2,...\". Only artifacts with these properties are affected.` `",
		},
	}
	return append(flags, getPropertiesFlags()...)
}

func getPropertiesFlags() []cli.Flag {
	propsFlags := append(getServerFlags(), getSortLimitFlags()...)
	return append(propsFlags, []cli.Flag{
		cli.BoolTFlag{
			Name:  "recursive",
			Usage: "[Default: true] When false, artifacts inside sub-folders in Artifactory will not be affected.` `",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.` `",
		},
		cli.BoolFlag{
			Name:  "include-dirs",
			Usage: "[Default: false] When true, the properties will also be set on folders (and not just files) in Artifactory.` `",
		},
		getFailNoOpFlag(),
		getExcludePatternsFlag(),
		getThreadsFlag(),
		getArchiveEntriesFlag(),
	}...)
}

func getArchiveEntriesFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "archive-entries",
		Usage: "[Optional] If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.` `",
	}
}

func getThreadsFlag() cli.Flag {
	return cli.StringFlag{
		Name:  "threads",
		Value: "",
		Usage: "[Default: 3] Number of working threads.` `",
	}
}

func getBuildPublishFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "build-url",
			Usage: "[Optional] Can be used for setting the CI server build URL in the build-info.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.` `",
		},
		cli.StringFlag{
			Name:  "env-include",
			Usage: "[Default: *] List of patterns in the form of \"value1;value2;...\" Only environment variables match those patterns will be included.` `",
		},
		cli.StringFlag{
			Name:  "env-exclude",
			Usage: "[Default: *password*;*secret*;*key*;*token*] List of case insensitive patterns in the form of \"value1;value2;...\". Environment variables match those patterns will be excluded.` `",
		},
	}...)
}

func getBuildAddDependenciesFlags() []cli.Flag {
	return append(getSpecFlags(), []cli.Flag{
		cli.BoolTFlag{
			Name:  "recursive",
			Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be added to the build info.` `",
		},
		cli.BoolFlag{
			Name:  "regexp",
			Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to be added to the build info.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to only get a summery of the dependencies that will be added to the build info.` `",
		},
		getExcludePatternsFlag(),
	}...)
}

func getBuildPromotionFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "status",
			Usage: "[Optional] Build promotion status.` `",
		},
		cli.StringFlag{
			Name:  "comment",
			Usage: "[Optional] Build promotion comment.` `",
		},
		cli.StringFlag{
			Name:  "source-repo",
			Usage: "[Optional] Build promotion source repository.` `",
		},
		cli.BoolFlag{
			Name:  "include-dependencies",
			Usage: "[Default: false] If set to true, the build dependencies are also promoted.` `",
		},
		cli.BoolFlag{
			Name:  "copy",
			Usage: "[Default: false] If set true, the build artifacts and dependencies are copied to the target repository, otherwise they are moved.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] If true, promotion is only simulated. The build is not promoted.` `",
		},
	}...)
}

func getBuildDistributeFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "source-repos",
			Usage: "[Optional] List of local repositories in the form of \"repo1,repo2,...\" from which build artifacts should be deployed.` `",
		},
		cli.StringFlag{
			Name:  "passphrase",
			Usage: "[Optional] If specified, Artifactory will GPG sign the build deployed to Bintray and apply the specified passphrase.` `",
		},
		cli.BoolTFlag{
			Name:  "publish",
			Usage: "[Default: true] If true, builds are published when deployed to Bintray.` `",
		},
		cli.BoolFlag{
			Name:  "override",
			Usage: "[Default: false] If true, Artifactory overwrites builds already existing in the target path in Bintray.` `",
		},
		cli.BoolFlag{
			Name:  "async",
			Usage: "[Default: false] If true, the build will be distributed asynchronously.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] If true, distribution is only simulated. No files are actually moved.` `",
		},
	}...)
}

func getGitLfsCleanFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "refs",
			Usage: "[Default: refs/remotes/*] List of Git references in the form of \"ref1,ref2,...\" which should be preserved.` `",
		},
		cli.StringFlag{
			Name:  "repo",
			Usage: "[Optional] Local Git LFS repository which should be cleaned. If omitted, this is detected from the Git repository.` `",
		},
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "[Default: false] Set to true to skip the delete confirmation message.` `",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] If true, cleanup is only simulated. No files are actually deleted.` `",
		},
	}...)
}

func getConfigFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolTFlag{
			Name:  "interactive",
			Usage: "[Default: true] Set to false if you do not want the config command to be interactive. If true, the --url option becomes optional.` `",
		},
		cli.BoolTFlag{
			Name:  "enc-password",
			Usage: "[Default: true] If set to false then the configured password will not be encrypted using Artifatory's encryption API.` `",
		},
	}
	flags = append(flags, getBaseFlags()...)
	return append(flags,
		getSshKeyPathFlag()...)
}

func getSshKeyPathFlag() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "ssh-key-path",
			Usage: "[Optional] SSH key file path.` `",
		},
	}
}

func getBuildDiscardFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "max-days",
			Usage: "[Optional] The maximum number of days to keep builds in Artifactory.` `",
		},
		cli.StringFlag{
			Name:  "max-builds",
			Usage: "[Optional] The maximum number of builds to store in Artifactory.` `",
		},
		cli.StringFlag{
			Name:  "exclude-builds",
			Usage: "[Optional] List of build numbers in the form of \"value1,value2,...\", that should not be removed from Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "delete-artifacts",
			Usage: "[Default: false] If set to true, automatically removes build artifacts stored in Artifactory.` `",
		},
		cli.BoolFlag{
			Name:  "async",
			Usage: "[Default: false] If set to true, build discard will run asynchronously and will not wait for response.` `",
		},
	}...)
}

func getBuildScanFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.BoolTFlag{
			Name:  "fail",
			Usage: "[Default: true] Set to false if you do not wish the command to return exit code 3, even if the 'Fail Build' rule is matched by Xray.` `",
		},
	}...)
}

func getBuildAddGitFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "[Optional] Path to a configuration file.` `",
		},
	}
}

func createArtifactoryDetailsByFlags(c *cli.Context, includeConfig bool) *config.ArtifactoryDetails {
	artDetails := createArtifactoryDetails(c, includeConfig)
	if artDetails.Url == "" {
		cliutils.ExitOnErr(errors.New("The --url option is mandatory"))
	}
	return artDetails
}

func getSplitCount(c *cli.Context) (splitCount int) {
	splitCount = cliutils.DownloadSplitCount
	var err error
	if c.String("split-count") != "" {
		splitCount, err = strconv.Atoi(c.String("split-count"))
		if err != nil {
			cliutils.ExitOnErr(errors.New("The '--split-count' option should have a numeric value. " + cliutils.GetDocumentationMessage()))
		}
		if splitCount > cliutils.DownloadMaxSplitCount {
			cliutils.ExitOnErr(errors.New("The '--split-count' option value is limited to a maximum of " + strconv.Itoa(cliutils.DownloadMaxSplitCount) + "."))
		}
		if splitCount < 0 {
			cliutils.ExitOnErr(errors.New("The '--split-count' option cannot have a negative value."))
		}
	}
	return
}

func getThreadsCount(c *cli.Context) (threads int) {
	threads = 3
	var err error
	if c.String("threads") != "" {
		threads, err = strconv.Atoi(c.String("threads"))
		if err != nil || threads < 1 {
			cliutils.ExitOnErr(errors.New("The '--threads' option should have a numeric positive value."))
		}
	}
	return
}

func getMinSplit(c *cli.Context) (minSplitSize int64) {
	minSplitSize = cliutils.DownloadMinSplitKb
	var err error
	if c.String("min-split") != "" {
		minSplitSize, err = strconv.ParseInt(c.String("min-split"), 10, 64)
		if err != nil {
			cliutils.ExitOnErr(errors.New("The '--min-split' option should have a numeric value. " + cliutils.GetDocumentationMessage()))
		}
	}
	return
}

func getRetries(c *cli.Context) (retries int) {
	retries = cliutils.Retries
	var err error
	if c.String("retries") != "" {
		retries, err = strconv.Atoi(c.String("retries"))
		if err != nil {
			cliutils.ExitOnErr(errors.New("The '--retries' option should have a numeric value. " + cliutils.GetDocumentationMessage()))
		}
	}
	return
}

func validateServerId(serverId string) {
	reservedIds := []string{"delete", "use", "show", "clear"}
	for _, reservedId := range reservedIds {
		if serverId == reservedId {
			cliutils.ExitOnErr(errors.New(fmt.Sprintf("Server can't have one of the following ID's: %s\n %s", strings.Join(reservedIds, ", "), cliutils.GetDocumentationMessage())))
		}
	}
}

func useCmd(c *cli.Context) {
	var serverId string
	if len(c.Args()) == 1 {
		serverId = c.Args()[0]
		validateServerId(serverId)
		err := commands.Use(serverId)
		cliutils.ExitOnErr(err)
		return
	} else {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
}

func configCmd(c *cli.Context) {
	if len(c.Args()) > 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var serverId string
	configCommandConfiguration := createConfigCommandConfiguration(c)
	if len(c.Args()) == 2 {
		serverId = c.Args()[1]
		validateServerId(serverId)
		if c.Args()[0] == "delete" {
			artDetails, err := config.GetArtifactorySpecificConfig(serverId)
			cliutils.ExitOnErr(err)
			if artDetails.IsEmpty() {
				log.Info("\"" + serverId + "\" configuration could not be found.")
				return
			}
			if !configCommandConfiguration.Interactive {
				cliutils.ExitOnErr(commands.DeleteConfig(serverId))
				return
			}
			var confirmed = cliutils.InteractiveConfirm("Are you sure you want to delete \"" + serverId + "\" configuration?")
			if !confirmed {
				return
			}
			cliutils.ExitOnErr(commands.DeleteConfig(serverId))
			return
		}
	}
	if len(c.Args()) > 0 {
		if c.Args()[0] == "show" {
			err := commands.ShowConfig(serverId)
			cliutils.ExitOnErr(err)
			return
		} else if c.Args()[0] == "clear" {
			commands.ClearConfig(configCommandConfiguration.Interactive)
			return
		} else {
			serverId = c.Args()[0]
			validateServerId(serverId)
		}
	}
	validateConfigFlags(configCommandConfiguration)
	_, err := commands.Config(configCommandConfiguration.ArtDetails, nil, configCommandConfiguration.Interactive, configCommandConfiguration.EncPassword, serverId)
	cliutils.ExitOnErr(err)
}

func mvnCmd(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	configuration := createBuildToolConfiguration(c)
	err := mvn.Mvn(c.Args().Get(0), c.Args().Get(1), configuration)
	cliutils.ExitOnErr(err)
}

func gradleCmd(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	configuration := createBuildToolConfiguration(c)
	err := gradle.Gradle(c.Args().Get(0), c.Args().Get(1), configuration)
	cliutils.ExitOnErr(err)
}

func dockerPushCmd(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	artDetails := createArtifactoryDetailsByFlags(c, true)
	imageTag := c.Args().Get(0)
	targetRepo := c.Args().Get(1)
	buildName := c.String("build-name")
	buildNumber := c.String("build-number")
	validateBuildParams(buildName, buildNumber)
	err := docker.PushDockerImage(imageTag, targetRepo, buildName, buildNumber, artDetails, getThreadsCount(c))
	cliutils.ExitOnErr(err)
}

func dockerPullCmd(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	artDetails := createArtifactoryDetailsByFlags(c, true)
	imageTag := c.Args().Get(0)
	sourceRepo := c.Args().Get(1)
	buildName := c.String("build-name")
	buildNumber := c.String("build-number")
	validateBuildParams(buildName, buildNumber)
	err := docker.PullDockerImage(imageTag, sourceRepo, buildName, buildNumber, artDetails)
	cliutils.ExitOnErr(err)
}

func nugetCmd(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	params := &nuget.Params{}
	params.Args = c.Args().Get(0)
	params.Flags = c.String("nuget-args")
	params.RepoName = c.Args().Get(1)
	params.BuildName = c.String("build-name")
	params.BuildNumber = c.String("build-number")

	path := c.String("solution-root")
	params.ArtifactoryDetails = createArtifactoryDetailsByFlags(c, true)

	err := nuget.ConsumeCmd(params, path)
	cliutils.ExitOnErr(err)
}

func nugetDepsTreeCmd(c *cli.Context) {
	if c.NArg() != 0 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	err := nuget.DependencyTreeCmd()
	cliutils.ExitOnErr(err)
}

func npmInstallCmd(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	configuration := createNpmConfiguration(c)
	err := npm.Install(c.Args().Get(0), configuration)
	cliutils.ExitOnErr(err)
}

func npmPublishCmd(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	configuration := createNpmConfiguration(c)
	err := npm.Publish(c.Args().Get(0), configuration)
	cliutils.ExitOnErr(err)
}

func goPublishCmd(c *cli.Context) {
	// When "self" set to true (default), there must be two arguments passed: target repo and the version
	if c.BoolT("self") && c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	// When "self" set to false, the target repository is mandatory but the version is not.
	// The version is only needed for publishing the project
	// But for automation purposes of users, keeping the possibility to pass the version without failing
	if !c.BoolT("self") && c.NArg() > 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	buildName := c.String("build-name")
	buildNumber := c.String("build-number")
	targetRepo := c.Args().Get(0)
	version := c.Args().Get(1)
	details := createArtifactoryDetailsByFlags(c, true)

	succeeded, failed, err := golang.Publish(c.BoolT("self"), c.String("deps"), targetRepo, version, buildName, buildNumber, details)
	err = cliutils.PrintSummaryReport(succeeded, failed, err)
	cliutils.ExitOnErr(err)
}

func goCmd(c *cli.Context) error {
	// When the no-registry set to false (default), two arguments are mandatory: go command and the target repository
	if !c.Bool("no-registry") && c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	// When the no-registry is set to true this means that the resolution will not be done via Artifactory.
	// For automation purposes of users, keeping the possibility to pass the repository although we are not using it.
	if c.Bool("no-registry") && c.NArg() > 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	buildName := c.String("build-name")
	buildNumber := c.String("build-number")
	goArg, err := shellwords.Parse(c.Args().Get(0))
	if err != nil {
		err = cliutils.PrintSummaryReport(0, 1, err)
	}
	targetRepo := c.Args().Get(1)
	details := createArtifactoryDetailsByFlags(c, true)

	err = golang.ExecuteGo(c.Bool("no-registry"), goArg, targetRepo, buildName, buildNumber, details)
	if err != nil {
		err = cliutils.PrintSummaryReport(0, 1, err)
	}
	return err
}

func goRecursivePublishCmd(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	targetRepo := c.Args().Get(0)
	if targetRepo == "" {
		cliutils.PrintHelpAndExitWithError("Missing target repo.", c)
	}
	details := createArtifactoryDetailsByFlags(c, true)
	err := golang.Execute(targetRepo, details)
	cliutils.ExitOnErr(err)
}

func createGradleConfigCmd(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	err := gradle.CreateBuildConfig(c.Args().Get(0))
	cliutils.ExitOnErr(err)
}

func createMvnConfigCmd(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	err := mvn.CreateBuildConfig(c.Args().Get(0))
	cliutils.ExitOnErr(err)
}

func pingCmd(c *cli.Context) {
	if c.NArg() > 0 {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent.", c)
	}
	artDetails := createArtifactoryDetails(c, true)
	resBody, err := generic.Ping(artDetails)
	resString := string(clientutils.IndentJson(resBody))
	if err != nil {
		cliutils.ExitOnErr(errors.New(err.Error() + "\n" + resString))
	}
	log.Output(resString)
}

func downloadCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || c.NArg() == 2 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var downloadSpec *spec.SpecFiles
	if c.IsSet("spec") {
		downloadSpec = getDownloadSpec(c)
	} else {
		validateCommonContext(c)
		downloadSpec = createDefaultDownloadSpec(c)
	}

	configuration := createDownloadConfiguration(c)
	downloaded, failed, err := generic.Download(downloadSpec, configuration)
	err = cliutils.PrintSummaryReport(downloaded, failed, err)
	cliutils.FailNoOp(err, downloaded, failed, isFailNoOp(c))
}

func uploadCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var uploadSpec *spec.SpecFiles
	if c.IsSet("spec") {
		uploadSpec = getFileSystemSpec(c, true)
	} else {
		uploadSpec = createDefaultUploadSpec(c)
	}
	configuration := createUploadConfiguration(c)
	uploaded, failed, err := generic.Upload(uploadSpec, configuration)
	err = cliutils.PrintSummaryReport(uploaded, failed, err)
	cliutils.FailNoOp(err, uploaded, failed, isFailNoOp(c))
}

func moveCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var moveSpec *spec.SpecFiles
	if c.IsSet("spec") {
		moveSpec = getCopyMoveSpec(c)
	} else {
		validateCommonContext(c)
		moveSpec = createDefaultCopyMoveSpec(c)
	}

	configuration := createMoveConfiguration(c)
	moveCount, failed, err := generic.Move(moveSpec, configuration)
	err = cliutils.PrintSummaryReport(moveCount, failed, err)
	cliutils.FailNoOp(err, moveCount, failed, isFailNoOp(c))
}

func copyCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var copySpec *spec.SpecFiles
	if c.IsSet("spec") {
		copySpec = getCopyMoveSpec(c)
	} else {
		validateCommonContext(c)
		copySpec = createDefaultCopyMoveSpec(c)
	}

	configuration := createCopyConfiguration(c)
	copyCount, failed, err := generic.Copy(copySpec, configuration)
	err = cliutils.PrintSummaryReport(copyCount, failed, err)
	cliutils.FailNoOp(err, copyCount, failed, isFailNoOp(c))
}

func deleteCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var deleteSpec *spec.SpecFiles
	if c.IsSet("spec") {
		deleteSpec = getDeleteSpec(c)
	} else {
		validateCommonContext(c)
		deleteSpec = createDefaultDeleteSpec(c)
	}

	configuration := createDeleteConfiguration(c)
	pathsToDelete, err := generic.GetPathsToDelete(deleteSpec, configuration)
	cliutils.ExitOnErr(err)
	if c.Bool("quiet") || confirmDelete(pathsToDelete) {
		success, failed, err := generic.DeleteFiles(pathsToDelete, configuration)
		err = cliutils.PrintSummaryReport(success, failed, err)
		cliutils.FailNoOp(err, success, failed, isFailNoOp(c))
	}
}

func confirmDelete(pathsToDelete []rtclientutils.ResultItem) bool {
	if len(pathsToDelete) < 1 {
		return false
	}
	for _, v := range pathsToDelete {
		fmt.Println("  " + v.GetItemRelativePath())
	}
	return cliutils.InteractiveConfirm("Are you sure you want to delete the above paths?")
}

func searchCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && (c.IsSet("spec") || c.IsSet("build")))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var searchSpec *spec.SpecFiles
	if c.IsSet("spec") {
		searchSpec = getSearchSpec(c)
	} else {
		validateCommonContext(c)
		searchSpec = createDefaultSearchSpec(c)
	}

	artDetails := createArtifactoryDetails(c, true)
	searchResults, err := generic.Search(searchSpec, artDetails)
	cliutils.ExitOnErr(err)
	result, err := json.Marshal(searchResults)
	cliutils.FailNoOp(err, len(searchResults), 0, isFailNoOp(c))

	log.Output(string(clientutils.IndentJson(result)))
}

func setPropsCmd(c *cli.Context) {
	validatePropsCommand(c)
	propertiesSpec, properties, artDetails := createPropsParams(c)
	success, failed, err := generic.SetProps(propertiesSpec, properties, getThreadsCount(c), artDetails)
	err = cliutils.PrintSummaryReport(success, failed, err)
	cliutils.FailNoOp(err, success, failed, isFailNoOp(c))
}

func deletePropsCmd(c *cli.Context) {
	validatePropsCommand(c)
	propertiesSpec, properties, artDetails := createPropsParams(c)
	success, failed, err := generic.DeleteProps(propertiesSpec, properties, getThreadsCount(c), artDetails)
	err = cliutils.PrintSummaryReport(success, failed, err)
	cliutils.FailNoOp(err, success, failed, isFailNoOp(c))
}

func buildPublishCmd(c *cli.Context) {
	validateBuildInfoArgument(c)
	configuration, artDetails := createBuildInfoConfiguration(c)
	err := buildinfo.Publish(c.Args().Get(0), c.Args().Get(1), configuration, artDetails)
	cliutils.ExitOnErr(err)
}

func buildAddDependenciesCmd(c *cli.Context) error {
	if c.NArg() > 2 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("Only path or spec is allowed, not both.", c)
	}
	if !(c.NArg() == 3 || (c.NArg() == 2 && c.IsSet("spec"))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var dependenciesSpec *spec.SpecFiles
	if c.IsSet("spec") {
		dependenciesSpec = getFileSystemSpec(c, false)
	} else {
		dependenciesSpec = createDefaultBuildAddDependenciesSpec(c)
	}
	configuration := createBuildAddDependenciesConfiguration(c)
	added, failed, err := buildinfo.AddDependencies(dependenciesSpec, configuration)
	err = cliutils.PrintSummaryReport(added, failed, err)
	cliutils.FailNoOp(err, added, failed, isFailNoOp(c))
	return nil
}

func buildCollectEnvCmd(c *cli.Context) {
	validateBuildInfoArgument(c)
	err := buildinfo.CollectEnv(c.Args().Get(0), c.Args().Get(1))
	cliutils.ExitOnErr(err)
}

func buildAddGitCmd(c *cli.Context) {
	if c.NArg() > 3 || c.NArg() < 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	buildAddGitConfiguration := createBuildAddGitConfiguration(c)
	err := buildinfo.AddGit(buildAddGitConfiguration)
	cliutils.ExitOnErr(err)
}

func buildScanCmd(c *cli.Context) {
	validateBuildInfoArgument(c)
	artDetails := createArtifactoryDetailsByFlags(c, true)
	failBuild, err := buildinfo.XrayScan(c.Args().Get(0), c.Args().Get(1), artDetails, c.BoolT("fail"))
	cliutils.ExitBuildScan(failBuild, err)
}

func buildCleanCmd(c *cli.Context) {
	validateBuildInfoArgument(c)
	err := buildinfo.Clean(c.Args().Get(0), c.Args().Get(1))
	cliutils.ExitOnErr(err)
}

func buildPromoteCmd(c *cli.Context) {
	if c.NArg() != 3 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	configuration := createBuildPromoteConfiguration(c)
	err := buildinfo.Promote(configuration)
	cliutils.ExitOnErr(err)
}

func buildDistributeCmd(c *cli.Context) {
	if c.NArg() != 3 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	configuration := createBuildDistributionConfiguration(c)
	err := buildinfo.Distribute(configuration)
	cliutils.ExitOnErr(err)
}

func buildDiscardCmd(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	configuration := createBuildDiscardConfiguration(c)
	err := buildinfo.BuildDiscard(configuration)
	cliutils.ExitOnErr(err)
}

func gitLfsCleanCmd(c *cli.Context) {
	if c.NArg() > 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	configuration := createGitLfsCleanConfiguration(c)
	filesToDelete, err := generic.PrepareGitLfsClean(configuration)
	cliutils.ExitOnErr(err)
	if len(filesToDelete) < 1 {
		return
	}
	if configuration.Quiet {
		err = generic.DeleteLfsFilesFromArtifactory(filesToDelete, configuration)
		cliutils.ExitOnErr(err)
		return
	}
	interactiveDeleteLfsFiles(filesToDelete, configuration)
}

func interactiveDeleteLfsFiles(filesToDelete []rtclientutils.ResultItem, configuration *generic.GitLfsCleanConfiguration) {
	for _, v := range filesToDelete {
		fmt.Println("  " + v.Name)
	}
	confirmed := cliutils.InteractiveConfirm("Are you sure you want to delete the above files?")
	if confirmed {
		err := generic.DeleteLfsFilesFromArtifactory(filesToDelete, configuration)
		cliutils.ExitOnErr(err)
	}
}

func validateBuildInfoArgument(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
}

func offerConfig(c *cli.Context) (details *config.ArtifactoryDetails) {
	var exists bool
	exists, err := config.IsArtifactoryConfExists()
	cliutils.ExitOnErr(err)
	if exists {
		return
	}

	var val bool
	val, err = clientutils.GetBoolEnvValue("JFROG_CLI_OFFER_CONFIG", true)
	cliutils.ExitOnErr(err)

	if !val {
		config.SaveArtifactoryConf(make([]*config.ArtifactoryDetails, 0))
		return
	}
	msg := "The CLI commands require the Artifactory URL and authentication details\n" +
		"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n" +
		"You can also configure these parameters later using the 'config' command.\n" +
		"Configure now?"
	confirmed := cliutils.InteractiveConfirm(msg)
	if !confirmed {
		config.SaveArtifactoryConf(make([]*config.ArtifactoryDetails, 0))
		return
	}
	details = createArtifactoryDetails(c, false)
	encPassword := c.BoolT("enc-password")
	details, err = commands.Config(nil, details, true, encPassword, "")
	cliutils.ExitOnErr(err)
	return
}

func createArtifactoryDetails(c *cli.Context, includeConfig bool) (details *config.ArtifactoryDetails) {
	if includeConfig {
		details := offerConfig(c)
		if details != nil {
			return details
		}
	}
	details = new(config.ArtifactoryDetails)
	details.Url = c.String("url")
	details.ApiKey = c.String("apikey")
	details.User = c.String("user")
	details.Password = c.String("password")
	details.SshKeyPath = c.String("ssh-key-path")
	details.SshPassphrase = c.String("ssh-passphrase")
	details.AccessToken = c.String("access-token")
	details.ServerId = c.String("server-id")

	if details.ApiKey != "" && details.User != "" && details.Password == "" {
		// The API Key is deprecated, use password option instead.
		details.Password = details.ApiKey
		details.ApiKey = ""
	}

	if includeConfig && !credentialsChanged(details) {
		confDetails, err := commands.GetConfig(details.ServerId)
		cliutils.ExitOnErr(err)

		if details.Url == "" {
			details.Url = confDetails.Url
		}

		if !isAuthMethodSet(details) {
			if details.ApiKey == "" {
				details.ApiKey = confDetails.ApiKey
			}
			if details.User == "" {
				details.User = confDetails.User
			}
			if details.Password == "" {
				details.Password = confDetails.Password
			}
			if details.SshKeyPath == "" {
				details.SshKeyPath = confDetails.SshKeyPath
			}
			if details.AccessToken == "" {
				details.AccessToken = confDetails.AccessToken
			}
		}
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	return
}

func credentialsChanged(details *config.ArtifactoryDetails) bool {
	return details.Url != "" || details.User != "" || details.Password != "" ||
		details.ApiKey != "" || details.SshKeyPath != "" || details.SshAuthHeaderSet() ||
		details.AccessToken != ""
}

func isAuthMethodSet(details *config.ArtifactoryDetails) bool {
	return (details.User != "" && details.Password != "") || details.SshKeyPath != "" || details.ApiKey != "" || details.AccessToken != ""
}

func getDebFlag(c *cli.Context) (deb string) {
	deb = c.String("deb")
	slashesCount := strings.Count(deb, "/") - strings.Count(deb, "\\/")
	if deb != "" && slashesCount != 2 {
		cliutils.ExitOnErr(errors.New("The --deb option should be in the form of distribution/component/architecture"))
	}
	return deb
}

func createDefaultCopyMoveSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		Build(c.String("build")).
		Offset(getIntValue("offset", c)).
		Limit(getIntValue("limit", c)).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Flat(c.Bool("flat")).
		IncludeDirs(true).
		Target(c.Args().Get(1)).
		ArchiveEntries(c.String("archive-entries")).
		BuildSpec()
}

func getCopyMoveSpec(c *cli.Context) (copyMoveSpec *spec.SpecFiles) {
	copyMoveSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)

	//Override spec with CLI options
	for i := 0; i < len(copyMoveSpec.Files); i++ {
		overrideFieldsIfSet(copyMoveSpec.Get(i), c)
	}
	err = spec.ValidateSpec(copyMoveSpec.Files, true, true)
	cliutils.ExitOnErr(err)
	return
}

func createDefaultDeleteSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		Build(c.String("build")).
		Offset(getIntValue("offset", c)).
		Limit(getIntValue("limit", c)).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		ArchiveEntries(c.String("archive-entries")).
		BuildSpec()
}

func getDeleteSpec(c *cli.Context) (deleteSpec *spec.SpecFiles) {
	deleteSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)

	//Override spec with CLI options
	for i := 0; i < len(deleteSpec.Files); i++ {
		overrideFieldsIfSet(deleteSpec.Get(i), c)
	}
	err = spec.ValidateSpec(deleteSpec.Files, false, true)
	cliutils.ExitOnErr(err)
	return
}

func createDeleteConfiguration(c *cli.Context) (deleteConfiguration *generic.DeleteConfiguration) {
	deleteConfiguration = new(generic.DeleteConfiguration)
	deleteConfiguration.DryRun = c.Bool("dry-run")
	deleteConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createMoveConfiguration(c *cli.Context) (moveConfiguration *generic.MoveConfiguration) {
	moveConfiguration = new(generic.MoveConfiguration)
	moveConfiguration.DryRun = c.Bool("dry-run")
	moveConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createCopyConfiguration(c *cli.Context) (copyConfiguration *generic.CopyConfiguration) {
	copyConfiguration = new(generic.CopyConfiguration)
	copyConfiguration.DryRun = c.Bool("dry-run")
	copyConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createDefaultSearchSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		Build(c.String("build")).
		Offset(getIntValue("offset", c)).
		Limit(getIntValue("limit", c)).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		ArchiveEntries(c.String("archive-entries")).
		BuildSpec()
}

func createDefaultPropertiesSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		Build(c.String("build")).
		Offset(getIntValue("offset", c)).
		Limit(getIntValue("limit", c)).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		IncludeDirs(c.Bool("include-dirs")).
		ArchiveEntries(c.String("archive-entries")).
		BuildSpec()
}

func getSearchSpec(c *cli.Context) (searchSpec *spec.SpecFiles) {
	searchSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)
	//Override spec with CLI options
	for i := 0; i < len(searchSpec.Files); i++ {
		overrideFieldsIfSet(searchSpec.Get(i), c)
	}
	err = spec.ValidateSpec(searchSpec.Files, false, true)
	cliutils.ExitOnErr(err)
	return
}

func createBuildInfoConfiguration(c *cli.Context) (flags *buildinfocmd.Configuration, artDetails *config.ArtifactoryDetails) {
	flags = new(buildinfocmd.Configuration)
	artDetails = createArtifactoryDetailsByFlags(c, true)
	flags.BuildUrl = c.String("build-url")
	flags.DryRun = c.Bool("dry-run")
	flags.EnvInclude = c.String("env-include")
	flags.EnvExclude = c.String("env-exclude")
	if len(flags.EnvInclude) == 0 {
		flags.EnvInclude = "*"
	}
	// Allow to use `env-exclude=""` and get no filters
	if !c.IsSet("env-exclude") {
		flags.EnvExclude = "*password*;*secret*;*key*;*token*"
	}
	return
}

func createBuildPromoteConfiguration(c *cli.Context) (promoteConfiguration *buildinfo.BuildPromotionConfiguration) {
	promotionParamsImpl := services.NewPromotionParams()
	promotionParamsImpl.Comment = c.String("comment")
	promotionParamsImpl.SourceRepo = c.String("source-repo")
	promotionParamsImpl.Status = c.String("status")
	promotionParamsImpl.IncludeDependencies = c.Bool("include-dependencies")
	promotionParamsImpl.Copy = c.Bool("copy")
	promoteConfiguration = new(buildinfo.BuildPromotionConfiguration)
	promoteConfiguration.DryRun = c.Bool("dry-run")
	promoteConfiguration.PromotionParams = promotionParamsImpl
	promoteConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	promoteConfiguration.BuildName = c.Args().Get(0)
	promoteConfiguration.BuildNumber = c.Args().Get(1)
	promoteConfiguration.TargetRepo = c.Args().Get(2)
	return
}

func createBuildDiscardConfiguration(c *cli.Context) (discardConfiguration *buildinfo.BuildDiscardConfiguration) {
	discardParamsImpl := services.NewDiscardBuildsParams()
	discardParamsImpl.DeleteArtifacts = c.Bool("delete-artifacts")
	discardParamsImpl.MaxBuilds = c.String("max-builds")
	discardParamsImpl.MaxDays = c.String("max-days")
	discardParamsImpl.ExcludeBuilds = c.String("exclude-builds")
	discardParamsImpl.Async = c.Bool("async")
	discardParamsImpl.BuildName = c.Args().Get(0)
	discardConfiguration = new(buildinfo.BuildDiscardConfiguration)
	discardConfiguration.DiscardBuildsParams = discardParamsImpl
	discardConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createBuildDistributionConfiguration(c *cli.Context) (distributeConfiguration *buildinfo.BuildDistributionConfiguration) {
	distributeParamsImpl := services.NewBuildDistributionParams()
	distributeParamsImpl.Publish = c.BoolT("publish")
	distributeParamsImpl.OverrideExistingFiles = c.Bool("override")
	distributeParamsImpl.GpgPassphrase = c.String("passphrase")
	distributeParamsImpl.Async = c.Bool("async")
	distributeParamsImpl.SourceRepos = c.String("source-repos")
	distributeConfiguration = new(buildinfo.BuildDistributionConfiguration)
	distributeConfiguration.DryRun = c.Bool("dry-run")
	distributeConfiguration.BuildDistributionParams = distributeParamsImpl
	distributeConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	distributeConfiguration.BuildName = c.Args().Get(0)
	distributeConfiguration.BuildNumber = c.Args().Get(1)
	distributeConfiguration.TargetRepo = c.Args().Get(2)
	return
}

func createGitLfsCleanConfiguration(c *cli.Context) (gitLfsCleanConfiguration *generic.GitLfsCleanConfiguration) {
	gitLfsCleanConfiguration = new(generic.GitLfsCleanConfiguration)

	gitLfsCleanConfiguration.Refs = c.String("refs")
	if len(gitLfsCleanConfiguration.Refs) == 0 {
		gitLfsCleanConfiguration.Refs = "refs/remotes/*"
	}

	gitLfsCleanConfiguration.Repo = c.String("repo")
	gitLfsCleanConfiguration.Quiet = c.Bool("quiet")
	gitLfsCleanConfiguration.DryRun = c.Bool("dry-run")
	gitLfsCleanConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	dotGitPath := ""
	if c.NArg() == 1 {
		dotGitPath = c.Args().Get(0)
	}
	gitLfsCleanConfiguration.GitPath = dotGitPath
	return
}

func createDefaultDownloadSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(strings.TrimPrefix(c.Args().Get(0), "/")).
		Props(c.String("props")).
		Build(c.String("build")).
		Offset(getIntValue("offset", c)).
		Limit(getIntValue("limit", c)).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Flat(c.Bool("flat")).
		Explode(c.String("explode")).
		IncludeDirs(c.Bool("include-dirs")).
		Target(c.Args().Get(1)).
		ArchiveEntries(c.String("archive-entries")).
		BuildSpec()
}

func getDownloadSpec(c *cli.Context) (downloadSpec *spec.SpecFiles) {
	downloadSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)

	fixWinDownloadFilesPath(downloadSpec)
	//Override spec with CLI options
	for i := 0; i < len(downloadSpec.Files); i++ {
		downloadSpec.Get(i).Pattern = strings.TrimPrefix(downloadSpec.Get(i).Pattern, "/")
		overrideFieldsIfSet(downloadSpec.Get(i), c)
	}
	err = spec.ValidateSpec(downloadSpec.Files, false, true)
	cliutils.ExitOnErr(err)
	return
}

func createDownloadConfiguration(c *cli.Context) (downloadConfiguration *utils.DownloadConfiguration) {
	downloadConfiguration = new(utils.DownloadConfiguration)
	downloadConfiguration.DryRun = c.Bool("dry-run")
	downloadConfiguration.ValidateSymlink = c.Bool("validate-symlinks")
	downloadConfiguration.MinSplitSize = getMinSplit(c)
	downloadConfiguration.SplitCount = getSplitCount(c)
	downloadConfiguration.Threads = getThreadsCount(c)
	downloadConfiguration.BuildName = c.String("build-name")
	downloadConfiguration.BuildNumber = c.String("build-number")
	downloadConfiguration.Retries = getRetries(c)
	downloadConfiguration.Symlink = true
	validateBuildParams(downloadConfiguration.BuildName, downloadConfiguration.BuildNumber)
	downloadConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createBuildAddDependenciesConfiguration(c *cli.Context) (buildAddDependenciesConfiguration *buildinfo.AddDependenciesConfiguration) {
	buildAddDependenciesConfiguration = new(buildinfo.AddDependenciesConfiguration)
	buildAddDependenciesConfiguration.DryRun = c.Bool("dry-run")
	buildAddDependenciesConfiguration.BuildName = c.Args().Get(0)
	buildAddDependenciesConfiguration.BuildNumber = c.Args().Get(1)
	return
}

func createDefaultUploadSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		Build(c.String("build")).
		Offset(getIntValue("offset", c)).
		Limit(getIntValue("limit", c)).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Flat(c.BoolT("flat")).
		Explode(c.String("explode")).
		Regexp(c.Bool("regexp")).
		IncludeDirs(c.Bool("include-dirs")).
		Target(strings.TrimPrefix(c.Args().Get(1), "/")).
		BuildSpec()
}

func createDefaultBuildAddDependenciesSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(2)).
		Recursive(c.BoolT("recursive")).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Regexp(c.Bool("regexp")).
		BuildSpec()
}

func getFileSystemSpec(c *cli.Context, isTargetMandatory bool) *spec.SpecFiles {
	fsSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)
	//Override spec with CLI options
	for i := 0; i < len(fsSpec.Files); i++ {
		fsSpec.Get(i).Target = strings.TrimPrefix(fsSpec.Get(i).Target, "/")
		overrideFieldsIfSet(fsSpec.Get(i), c)
	}
	fixWinUploadFilesPath(fsSpec)
	err = spec.ValidateSpec(fsSpec.Files, isTargetMandatory, false)
	cliutils.ExitOnErr(err)
	return fsSpec
}

func fixWinUploadFilesPath(uploadSpec *spec.SpecFiles) {
	if cliutils.IsWindows() {
		for i, file := range uploadSpec.Files {
			uploadSpec.Files[i].Pattern = strings.Replace(file.Pattern, "\\", "\\\\", -1)
			for j, excludePattern := range uploadSpec.Files[i].ExcludePatterns {
				uploadSpec.Files[i].ExcludePatterns[j] = strings.Replace(excludePattern, "\\", "\\\\", -1)
			}
		}
	}
}

func fixWinDownloadFilesPath(uploadSpec *spec.SpecFiles) {
	if cliutils.IsWindows() {
		for i, file := range uploadSpec.Files {
			uploadSpec.Files[i].Target = strings.Replace(file.Target, "\\", "\\\\", -1)
		}
	}
}

func createUploadConfiguration(c *cli.Context) (uploadConfiguration *utils.UploadConfiguration) {
	uploadConfiguration = new(utils.UploadConfiguration)
	buildName := c.String("build-name")
	buildNumber := c.String("build-number")
	validateBuildParams(buildName, buildNumber)
	uploadConfiguration.BuildName = buildName
	uploadConfiguration.BuildNumber = buildNumber
	uploadConfiguration.DryRun = c.Bool("dry-run")
	uploadConfiguration.Symlink = c.Bool("symlinks")
	uploadConfiguration.Retries = getRetries(c)
	uploadConfiguration.Threads = getThreadsCount(c)
	uploadConfiguration.Deb = getDebFlag(c)
	uploadConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createBuildToolConfiguration(c *cli.Context) (buildConfigConfiguration *utils.BuildConfiguration) {
	buildConfigConfiguration = new(utils.BuildConfiguration)
	buildConfigConfiguration.BuildName = c.String("build-name")
	buildConfigConfiguration.BuildNumber = c.String("build-number")
	validateBuildParams(buildConfigConfiguration.BuildName, buildConfigConfiguration.BuildNumber)
	return
}

func createNpmConfiguration(c *cli.Context) (npmConfiguration *npmutils.CliConfiguration) {
	npmConfiguration = new(npmutils.CliConfiguration)
	npmConfiguration.BuildName = c.String("build-name")
	npmConfiguration.BuildNumber = c.String("build-number")
	validateBuildParams(npmConfiguration.BuildName, npmConfiguration.BuildNumber)
	npmConfiguration.NpmArgs = c.String("npm-args")
	if c.String("npm-args") != "" {
	}
	npmConfiguration.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createConfigCommandConfiguration(c *cli.Context) (configCommandConfiguration *commands.ConfigCommandConfiguration) {
	configCommandConfiguration = new(commands.ConfigCommandConfiguration)
	configCommandConfiguration.ArtDetails = createArtifactoryDetails(c, false)
	configCommandConfiguration.EncPassword = c.BoolT("enc-password")
	configCommandConfiguration.Interactive = c.BoolT("interactive")
	return
}

func validateConfigFlags(configCommandConfiguration *commands.ConfigCommandConfiguration) {
	if !configCommandConfiguration.Interactive && configCommandConfiguration.ArtDetails.Url == "" {
		cliutils.ExitOnErr(errors.New("The --url option is mandatory when the --interactive option is set to false"))
	}
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

func validateCommonContext(c *cli.Context) {
	// Validate build
	if c.IsSet("build") {
		if c.IsSet("offset") {
			cliutils.ExitOnErr(errors.New("Cannot use 'offset' together with 'build'"))
		}
		if c.IsSet("limit") {
			cliutils.ExitOnErr(errors.New("Cannot use 'limit' together with 'build'"))
		}

		// Allow pattern as '*' or empty only for build search without more filters
		pattern := c.Args().Get(0)
		if pattern == "*" || pattern == "" {
			if c.IsSet("props") {
				cliutils.ExitOnErr(errors.New("Cannot use 'props' together with 'build' if 'pattern' is empty or '*'"))
			}
			if c.IsSet("exclude-patterns") {
				cliutils.ExitOnErr(errors.New("Cannot use 'exclude-patterns' together with 'build' if 'pattern' is empty or '*'"))
			}
			if c.IsSet("archive-entries") {
				cliutils.ExitOnErr(errors.New("Cannot use 'archive-entries' together with 'build' if 'pattern' is empty or '*'"))
			}
			if c.IsSet("recursive") {
				cliutils.ExitOnErr(errors.New("Cannot use 'recursive' together with 'build' if 'pattern' is empty or '*'"))
			}
			if c.IsSet("include-dirs") {
				cliutils.ExitOnErr(errors.New("Cannot use 'include-dirs' together with 'build' if 'pattern' is empty or '*'"))
			}
		}
	}

	// Validate sort-order
	if c.IsSet("sort-order") {
		if !c.IsSet("sort-by") {
			cliutils.ExitOnErr(errors.New("Cannot use 'sort-order' without the 'sort-by' option"))
		}
		if !(c.String("sort-order") == "asc" || c.String("sort-order") == "desc") {
			cliutils.ExitOnErr(errors.New("The 'sort-order' option can only accept 'asc' or 'desc' as values"))
		}
	}
}

func validateBuildParams(buildName, buildNumber string) {
	if (buildName == "" && buildNumber != "") || (buildName != "" && buildNumber == "") {
		cliutils.ExitOnErr(errors.New("The build-name and build-number options cannot be sent separately."))
	}
}

func overrideFieldsIfSet(spec *spec.File, c *cli.Context) {
	overrideArrayIfSet(&spec.ExcludePatterns, c, "exclude-patterns")
	overrideArrayIfSet(&spec.SortBy, c, "sort-by")
	overrideIntIfSet(&spec.Offset, c, "offset")
	overrideIntIfSet(&spec.Limit, c, "limit")
	overrideStringIfSet(&spec.SortOrder, c, "sort-order")
	overrideStringIfSet(&spec.Props, c, "props")
	overrideStringIfSet(&spec.Build, c, "build")
	overrideStringIfSet(&spec.Recursive, c, "recursive")
	overrideStringIfSet(&spec.Flat, c, "flat")
	overrideStringIfSet(&spec.Explode, c, "explode")
	overrideStringIfSet(&spec.Regexp, c, "regexp")
	overrideStringIfSet(&spec.IncludeDirs, c, "include-dirs")
}

func getIntValue(key string, c *cli.Context) int {
	value, err := cliutils.GetIntFlagValue(c, key, 0)
	cliutils.ExitOnErr(err)
	return value
}

func isFailNoOp(context *cli.Context) bool {
	if context == nil {
		return false
	}
	return context.Bool("fail-no-op")
}

func createPropsParams(c *cli.Context) (propertiesSpec *spec.SpecFiles, properties string, artDetails *config.ArtifactoryDetails) {
	propertiesSpec = createDefaultPropertiesSpec(c)
	properties = c.Args()[1]
	artDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func validatePropsCommand(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	validateCommonContext(c)
}

func createBuildAddGitConfiguration(c *cli.Context) (buildAddGitConfiguration *buildinfo.BuildAddGitConfiguration) {
	buildAddGitConfiguration = new(buildinfo.BuildAddGitConfiguration)
	buildAddGitConfiguration.BuildName = c.Args().Get(0)
	buildAddGitConfiguration.BuildNumber = c.Args().Get(1)
	dotGitPath := ""
	if c.NArg() == 3 {
		dotGitPath = c.Args().Get(2)
	}
	buildAddGitConfiguration.DotGitPath = dotGitPath
	buildAddGitConfiguration.ConfigFilePath = c.String("config")
	return
}

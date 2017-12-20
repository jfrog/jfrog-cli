package artifactory

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/commands"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"strconv"
	"strings"
	"encoding/json"
	"runtime"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/common"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/copy"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/delete"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/upload"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/use"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/download"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/move"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/search"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/buildpublish"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/buildcollectenv"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/buildaddgit"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/buildclean"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/buildpromote"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/builddistribute"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/gitlfsclean"
	configdocs "github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/setprops"
	rtclientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/mvn"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/gradle"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/gradleconfig"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/mvnconfig"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/artifactory/buildscan"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/spec"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
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
			Name:      "build-add-git",
			Flags:     []cli.Flag{},
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
			Flags:     getServerFlags(),
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
			Usage:     mvn.Description,
			HelpName:  common.CreateUsage("rt mvn", mvn.Description, mvn.Usage),
			UsageText: mvn.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				mvnCmd(c)
			},
		},
		{
			Name:      "mvn-config",
			Flags:     getBuildToolFlags(),
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
			Usage:     gradle.Description,
			HelpName:  common.CreateUsage("rt gradle", gradle.Description, gradle.Usage),
			UsageText: gradle.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				gradleCmd(c)
			},
		},
		{
			Name:      "gradle-config",
			Flags:     getBuildToolFlags(),
			Aliases:   []string{"gradlec"},
			Usage:     gradleconfig.Description,
			HelpName:  common.CreateUsage("rt gradle-config", gradleconfig.Description, gradleconfig.Usage),
			UsageText: gradleconfig.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				createGradleConfigCmd(c)
			},
		},
	}
}

func getCommonFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Usage: "[Optional] Artifactory URL.",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "[Optional] Artifactory username.",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "[Optional] Artifactory password.",
		},
		cli.StringFlag{
			Name:  "apikey",
			Usage: "[Optional] Artifactory API key.",
		},
		cli.StringFlag{
			Name:  "ssh-key-path",
			Usage: "[Optional] SSH key file path.",
		},
		cli.StringFlag{
			Name:  "ssh-passphrase",
			Usage: "[Optional] SSH key passphrase.",
		},
	}
}

func getServerFlags() []cli.Flag {
	return append(getCommonFlags(), cli.StringFlag{
		Name:  "server-id",
		Usage: "[Optional] Artifactory server ID configured using the config command.",
	})
}

func getSortLimitFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "sort-by",
			Usage: "[Optional] A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information, see https://www.jfrog.com/confluence/display/RTF/Artifactory+Query+Language#ArtifactoryQueryLanguage-EntitiesandFields",
		},
		cli.StringFlag{
			Name:  "sort-order",
			Usage: "[Default: asc] The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.",
		},
		cli.StringFlag{
			Name:  "limit",
			Usage: "[Optional] The maximum number of items to fetch. Usually used with the 'sort-by' option.",
		},
		cli.StringFlag{
			Name:  "offset",
			Usage: "[Optional] The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.",
		},
	}
}

func getUploadFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a File Spec.",
		},
		cli.StringFlag{
			Name:  "spec-vars",
			Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.",
		},
		cli.StringFlag{
			Name:  "build-name",
			Usage: "[Optional] Build name. Providing this option will record all uploaded artifacts for later build info publication.",
		},
		cli.StringFlag{
			Name:  "build-number",
			Usage: "[Optional] Build number. Providing this option will record all uploaded artifacts for later build info publication.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\" to be attached to the uploaded artifacts.",
		},
		cli.StringFlag{
			Name:  "deb",
			Usage: "[Optional] Used for Debian packages in the form of distribution/component/architecture.",
		},
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be uploaded to Artifactory.",
		},
		cli.StringFlag{
			Name:  "flat",
			Value: "",
			Usage: "[Default: true] If set to false, files are uploaded according to their file system hierarchy.",
		},
		cli.BoolFlag{
			Name:  "regexp",
			Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to upload.",
		},
		cli.StringFlag{
			Name:  "threads",
			Value: "",
			Usage: "[Default: 3] Number of artifacts to upload in parallel.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.",
		},
		cli.BoolFlag{
			Name:  "explode",
			Usage: "[Default: false] Set to true to extract an archive after it is deployed to Artifactory.",
		},
		cli.BoolFlag{
			Name:  "symlinks",
			Usage: "[Default: false] Set to true to preserve symbolic links structure in Artifactory.",
		},
		cli.BoolFlag{
			Name:  "include-dirs",
			Usage: "[Default: false] Set to true if you'd like to also apply the source path pattern for directories and not just for files.",
		},
		cli.StringFlag{
			Name:  "exclude-patterns",
			Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.",
		},
	}...)
}

func getDownloadFlags() []cli.Flag {
	downloadFlags := append(getServerFlags(), getSortLimitFlags()...)
	return append(downloadFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a File Spec.",
		},
		cli.StringFlag{
			Name:  "spec-vars",
			Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.",
		},
		cli.StringFlag{
			Name:  "build-name",
			Usage: "[Optional] Build name. Providing this option will record all downloaded artifacts for later build info publication.",
		},
		cli.StringFlag{
			Name:  "build-number",
			Usage: "[Optional] Build number. Providing this option will record all downloaded artifacts for later build info publication.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be downloaded.",
		},
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not wish to include the download of artifacts inside sub-folders in Artifactory.",
		},
		cli.StringFlag{
			Name:  "flat",
			Value: "",
			Usage: "[Default: false] Set to true if you do not wish to have the Artifactory repository path structure created locally for your downloaded files.",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are downloaded. The property format is build-name/build-number.",
		},
		cli.StringFlag{
			Name:  "min-split",
			Value: "",
			Usage: "[Default: 5120] Minimum file size in KB to split into ranges when downloading. Set to -1 for no splits.",
		},
		cli.StringFlag{
			Name:  "split-count",
			Value: "",
			Usage: "[Default: 3] Number of parts to split a file when downloading. Set to 0 for no splits.",
		},
		cli.StringFlag{
			Name:  "threads",
			Value: "",
			Usage: "[Default: 3] Number of artifacts to download in parallel.",
		},
		cli.StringFlag{
			Name:  "retries",
			Usage: "[Default: 3] Number of download retries.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.",
		},
		cli.BoolFlag{
			Name:  "validate-symlinks",
			Usage: "[Default: false] Set to true to perform a checksum validation when downloading symbolic links.",
		},
		cli.BoolFlag{
			Name:  "include-dirs",
			Usage: "[Default: false] Set to true if you'd like to also apply the target path pattern for folders and not just for files in Artifactory.",
		},
		cli.StringFlag{
			Name:  "exclude-patterns",
			Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.",
		},
	}...)
}

func getBuildToolFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "build-name",
			Usage: "[Optional] Providing this option will collect and record build info for this build name.",
		},
		cli.StringFlag{
			Name:  "build-number",
			Usage: "[Optional] Providing this option will collect and record build info for this build number. If you provide a build name (using the --build-name option) and do not provide a build number, a build number will be automatically generated.",
		},
	}
}

func getMoveFlags() []cli.Flag {
	moveFlags := append(getServerFlags(), getSortLimitFlags()...)
	return append(moveFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a File Spec.",
		},
		cli.StringFlag{
			Name:  "spec-vars",
			Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.",
		},
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not wish to move artifacts inside sub-folders in Artifactory.",
		},
		cli.StringFlag{
			Name:  "flat",
			Value: "",
			Usage: "[Default: false] If set to false, files are moved according to their file system hierarchy.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be moved.",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are moved. The property format is build-name/build-number.",
		},
		cli.StringFlag{
			Name:  "exclude-patterns",
			Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.",
		},
	}...)

}

func getCopyFlags() []cli.Flag {
	copyFlags := append(getServerFlags(), getSortLimitFlags()...)
	return append(copyFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a File Spec.",
		},
		cli.StringFlag{
			Name:  "spec-vars",
			Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.",
		},
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not wish to copy artifacts inside sub-folders in Artifactory.",
		},
		cli.StringFlag{
			Name:  "flat",
			Value: "",
			Usage: "[Default: false] If set to false, files are copied according to their file system hierarchy.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be copied.",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are copied. The property format is build-name/build-number.",
		},
		cli.StringFlag{
			Name:  "exclude-patterns",
			Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.",
		},
	}...)
}

func getDeleteFlags() []cli.Flag {
	deleteFlags := append(getServerFlags(), getSortLimitFlags()...)
	return append(deleteFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a File Spec.",
		},
		cli.StringFlag{
			Name:  "spec-vars",
			Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be deleted.",
		},
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not wish to delete artifacts inside sub-folders in Artifactory.",
		},
		cli.StringFlag{
			Name:  "quiet",
			Value: "",
			Usage: "[Default: false] Set to true to skip the delete confirmation message.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are deleted. The property format is build-name/build-number.",
		},
		cli.StringFlag{
			Name:  "exclude-patterns",
			Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.",
		},
	}...)
}

func getSearchFlags() []cli.Flag {
	searchFlags := append(getServerFlags(), getSortLimitFlags()...)
	return append(searchFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a File Spec.",
		},
		cli.StringFlag{
			Name:  "spec-vars",
			Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be returned.",
		},
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not wish to search artifacts inside sub-folders in Artifactory.",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number.",
		},
		cli.StringFlag{
			Name:  "exclude-patterns",
			Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.",
		},
	}...)
}

func getSetPropertiesFlags() []cli.Flag {
	propsFlags := append(getServerFlags(), getSortLimitFlags()...)
	return append(propsFlags, []cli.Flag{
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties are affected.",
		},
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] When false, artifacts inside sub-folders in Artifactory will not be affected.",
		},
		cli.StringFlag{
			Name:  "build",
			Usage: "[Optional] If specified, only artifacts of the specified build are affected. The property format is build-name/build-number.",
		},
		cli.BoolFlag{
			Name:  "include-dirs",
			Usage: "[Default: false] When true, the properties will also be set on folders (and not just files) in Artifactory.",
		},
		cli.StringFlag{
			Name:  "exclude-patterns",
			Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.",
		},
	}...)
}

func getBuildPublishFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.",
		},
		cli.StringFlag{
			Name:  "env-include",
			Usage: "[Default: *] List of patterns in the form of \"value1;value2;...\" Only environment variables match those patterns will be included.",
		},
		cli.StringFlag{
			Name:  "env-exclude",
			Usage: "[Default: *password*;*secret*;*key*;*token*] List of patterns in the form of \"value1;value2;...\"  environment variables match those patterns will be excluded.",
		},
	}...)
}

func getBuildPromotionFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "status",
			Usage: "[Optional] Build promotion status.",
		},
		cli.StringFlag{
			Name:  "comment",
			Usage: "[Optional] Build promotion comment.",
		},
		cli.StringFlag{
			Name:  "source-repo",
			Usage: "[Optional] Build promotion source repository.",
		},
		cli.BoolFlag{
			Name:  "include-dependencies",
			Usage: "[Default: false] If set to true, the build dependencies are also promoted.",
		},
		cli.BoolFlag{
			Name:  "copy",
			Usage: "[Default: false] If set true, the build are artifacts and dependencies are copied to the target repository, otherwise they are moved.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] If true, promotion is only simulated. The build is not promoted.",
		},
	}...)
}

func getBuildDistributeFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "source-repos",
			Usage: "[Optional] List of local repositories in the form of \"repo1,repo2,...\" from which build artifacts should be deployed.",
		},
		cli.StringFlag{
			Name:  "passphrase",
			Usage: "[Optional] If specified, Artifactory will GPG sign the build deployed to Bintray and apply the specified passphrase.",
		},
		cli.BoolFlag{
			Name:  "publish",
			Usage: "[Default: true] If true, builds are published when deployed to Bintray.",
		},
		cli.BoolFlag{
			Name:  "override",
			Usage: "[Default: false] If true, Artifactory overwrites builds already existing in the target path in Bintray.",
		},
		cli.BoolFlag{
			Name:  "async",
			Usage: "[Default: false] If true, the build will be distributed asynchronously.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] If true, distribution is only simulated. No files are actually moved.",
		},
	}...)
}

func getGitLfsCleanFlags() []cli.Flag {
	return append(getServerFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "refs",
			Usage: "[Default: refs/remotes/*] List of Git references in the form of \"ref1,ref2,...\" which should be preserved.",
		},
		cli.StringFlag{
			Name:  "repo",
			Usage: "[Optional] Local Git LFS repository which should be cleaned. If omitted, this is detected from the Git repository.",
		},
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "[Default: false] Set to true to skip the delete confirmation message.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] If true, cleanup is only simulated. No files are actually deleted.",
		},
	}...)
}

func getConfigFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "interactive",
			Usage: "[Default: true] Set to false if you do not want the config command to be interactive. If true, the --url option becomes optional.",
		},
		cli.StringFlag{
			Name:  "enc-password",
			Usage: "[Default: true] If set to false then the configured password will not be encrypted using Artifatory's encryption API.",
		},
	}
	return append(flags, getCommonFlags()...)
}

func createArtifactoryDetailsByFlags(c *cli.Context, includeConfig bool) *config.ArtifactoryDetails {
	artDetails := createArtifactoryDetails(c, includeConfig)
	if artDetails.Url == "" {
		cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory")
	}
	return artDetails
}

func getSplitCount(c *cli.Context) (splitCount int) {
	splitCount = 3
	var err error
	if c.String("split-count") != "" {
		splitCount, err = strconv.Atoi(c.String("split-count"))
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option should have a numeric value. "+cliutils.GetDocumentationMessage())
		}
		if splitCount > 15 {
			cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option value is limitted to a maximum of 15.")
		}
		if splitCount < 0 {
			cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option cannot have a negative value.")
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
			cliutils.Exit(cliutils.ExitCodeError, "The '--threads' option should have a numeric positive value.")
		}
	}
	return
}

func getBuildName(c *cli.Context) (buildName string) {
	buildName = ""
	if c.String("build-name") != "" {
		buildName = c.String("build-name")
	}
	return
}

func getBuildNumber(c *cli.Context) (buildNumber string) {
	buildNumber = ""
	if c.String("build-number") != "" {
		buildNumber = c.String("build-number")
	}
	return
}

func getMinSplit(c *cli.Context) (minSplitSize int64) {
	minSplitSize = 5120
	var err error
	if c.String("min-split") != "" {
		minSplitSize, err = strconv.ParseInt(c.String("min-split"), 10, 64)
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, "The '--min-split' option should have a numeric value. "+cliutils.GetDocumentationMessage())
		}
	}
	return
}

func getRetries(c *cli.Context) (retries int) {
	retries = 3
	var err error
	if c.String("retries") != "" {
		retries, err = strconv.Atoi(c.String("retries"))
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, "The '--retries' option should have a numeric value. "+cliutils.GetDocumentationMessage())
		}
	}
	return
}

func validateServerId(serverId string) {
	reservedIds := []string{"delete", "use", "show", "clear"}
	for _, reservedId := range reservedIds {
		if serverId == reservedId {
			cliutils.Exit(cliutils.ExitCodeError, fmt.Sprintf("Server can't have one of the following ID's: %s\n %s", strings.Join(reservedIds, ", "), cliutils.GetDocumentationMessage()))
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
	configFlags := createConfigFlags(c)
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
			if !configFlags.Interactive {
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
			commands.ClearConfig(configFlags.Interactive)
			return
		} else {
			serverId = c.Args()[0]
			validateServerId(serverId)
		}
	}
	validateConfigFlags(configFlags)
	_, err := commands.Config(configFlags.ArtDetails, nil, configFlags.Interactive, configFlags.EncPassword, serverId)
	cliutils.ExitOnErr(err)
}

func mvnCmd(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	flags := createBuildToolFlags(c)
	err := commands.Mvn(c.Args().Get(0), c.Args().Get(1), flags)
	cliutils.ExitOnErr(err)
}

func gradleCmd(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	flags := createBuildToolFlags(c)
	err := commands.Gradle(c.Args().Get(0), c.Args().Get(1), flags)
	cliutils.ExitOnErr(err)
}

func createGradleConfigCmd(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	err := commands.CreateGradleBuildConfig(c.Args().Get(0))
	cliutils.ExitOnErr(err)
}

func createMvnConfigCmd(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	err := commands.CreateMvnBuildConfig(c.Args().Get(0))
	cliutils.ExitOnErr(err)
}

func downloadCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var downloadSpec *spec.SpecFiles
	if c.IsSet("spec") {
		downloadSpec = getDownloadSpec(c)
	} else {
		validateCommonContext(c)
		downloadSpec = createDefaultDownloadSpec(c)
	}

	flags := createDownloadFlags(c)
	downloaded, failed, err := commands.Download(downloadSpec, flags)
	err = cliutils.PrintSummaryReport(downloaded, failed, err)
	cliutils.ExitOnErr(err)
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
		uploadSpec = getUploadSpec(c)
	} else {
		uploadSpec = createDefaultUploadSpec(c)
	}
	flags := createUploadFlags(c)
	uploaded, failed, err := commands.Upload(uploadSpec, flags)
	err = cliutils.PrintSummaryReport(uploaded, failed, err)
	cliutils.ExitOnErr(err)
	if failed > 0 {
		if uploaded > 0 {
			cliutils.Exit(cliutils.ExitCodeWarning, "")
		}
		cliutils.Exit(cliutils.ExitCodeError, "")
	}
}

func moveCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var moveSpec *spec.SpecFiles
	if c.IsSet("spec") {
		moveSpec = getCopyMoveSpec(c)
	} else {
		validateCommonContext(c)
		moveSpec = createDefaultCopyMoveSpec(c)
	}

	artDetails := createArtifactoryDetails(c, true)
	moveCount, failed, err := commands.Move(moveSpec, artDetails)
	err = cliutils.PrintSummaryReport(moveCount, failed, err)
	cliutils.ExitOnErr(err)
}

func copyCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var copySpec *spec.SpecFiles
	if c.IsSet("spec") {
		copySpec = getCopyMoveSpec(c)
	} else {
		validateCommonContext(c)
		copySpec = createDefaultCopyMoveSpec(c)
	}

	artDetails := createArtifactoryDetails(c, true)
	copyCount, failed, err := commands.Copy(copySpec, artDetails)
	err = cliutils.PrintSummaryReport(copyCount, failed, err)
	cliutils.ExitOnErr(err)
}

func deleteCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.PrintHelpAndExitWithError("No arguments should be sent when the spec option is used.", c)
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	var deleteSpec *spec.SpecFiles
	if c.IsSet("spec") {
		deleteSpec = getDeleteSpec(c)
	} else {
		validateCommonContext(c)
		deleteSpec = createDefaultDeleteSpec(c)
	}

	flags := createDeleteFlags(c)
	pathsToDelete, err := commands.GetPathsToDelete(deleteSpec, flags)
	cliutils.ExitOnErr(err)
	if c.Bool("quiet") || confirmDelete(pathsToDelete) {
		success, failed, err := commands.DeleteFiles(pathsToDelete, flags)
		err = cliutils.PrintSummaryReport(success, failed, err)
		cliutils.ExitOnErr(err)
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
	if !(c.NArg() == 1 || (c.NArg() == 0 && c.IsSet("spec"))) {
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
	SearchResult, err := commands.Search(searchSpec, artDetails)
	cliutils.ExitOnErr(err)
	result, err := json.Marshal(SearchResult)
	cliutils.ExitOnErr(err)

	log.Output(string(clientutils.IndentJson(result)))
}

func setPropsCmd(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	validateCommonContext(c)
	setPropertiesSpec := createDefaultSetPropertiesSpec(c)
	properties := c.Args()[1]
	artDetails := createArtifactoryDetailsByFlags(c, true)
	success, failed, err := commands.SetProps(setPropertiesSpec, properties, artDetails)
	err = cliutils.PrintSummaryReport(success, failed, err)
	cliutils.ExitOnErr(err)
}

func buildPublishCmd(c *cli.Context) {
	validateBuildInfoArgument(c)
	buildInfoFlags, artDetails := createBuildInfoFlags(c)
	err := commands.BuildPublish(c.Args().Get(0), c.Args().Get(1), buildInfoFlags, artDetails)
	cliutils.ExitOnErr(err)
}

func buildCollectEnvCmd(c *cli.Context) {
	validateBuildInfoArgument(c)
	err := commands.BuildCollectEnv(c.Args().Get(0), c.Args().Get(1))
	cliutils.ExitOnErr(err)
}

func buildAddGitCmd(c *cli.Context) {
	if c.NArg() > 3 || c.NArg() < 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	dotGitPath := ""
	if c.NArg() == 3 {
		dotGitPath = c.Args().Get(2)
	}
	err := commands.BuildAddGit(c.Args().Get(0), c.Args().Get(1), dotGitPath)
	cliutils.ExitOnErr(err)
}

func buildScanCmd(c *cli.Context) {
	validateBuildInfoArgument(c)

	artDetails := createArtifactoryDetailsByFlags(c, true)
	err := commands.BuildScan(c.Args().Get(0), c.Args().Get(1), artDetails)
	cliutils.ExitOnErr(err)
}

func buildCleanCmd(c *cli.Context) {
	validateBuildInfoArgument(c)
	err := commands.BuildClean(c.Args().Get(0), c.Args().Get(1))
	cliutils.ExitOnErr(err)
}

func buildPromoteCmd(c *cli.Context) {
	if c.NArg() != 3 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	buildPromoteFlags := createBuildPromoteFlags(c)
	buildPromoteFlags.BuildName = c.Args().Get(0)
	buildPromoteFlags.BuildNumber = c.Args().Get(1)
	buildPromoteFlags.TargetRepo = c.Args().Get(2)
	err := commands.BuildPromote(buildPromoteFlags)
	cliutils.ExitOnErr(err)
}

func buildDistributeCmd(c *cli.Context) {
	if c.NArg() != 3 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	buildDistributeFlags := createBuildDistributionFlags(c)
	buildDistributeFlags.BuildName = c.Args().Get(0)
	buildDistributeFlags.BuildNumber = c.Args().Get(1)
	buildDistributeFlags.TargetRepo = c.Args().Get(2)
	err := commands.BuildDistribute(buildDistributeFlags)
	cliutils.ExitOnErr(err)
}

func gitLfsCleanCmd(c *cli.Context) {
	if c.NArg() > 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	dotGitPath := ""
	if c.NArg() == 1 {
		dotGitPath = c.Args().Get(0)
	}
	gitLfsCleanFlags := createGitLfsCleanFlags(c)
	gitLfsCleanFlags.GitLfsCleanParamsImpl.GitPath = dotGitPath
	filesToDelete, err := commands.PrepareGitLfsClean(gitLfsCleanFlags)
	cliutils.ExitOnErr(err)
	if len(filesToDelete) < 1 {
		return
	}
	if gitLfsCleanFlags.Quiet {
		err = commands.DeleteLfsFilesFromArtifactory(filesToDelete, gitLfsCleanFlags)
		cliutils.ExitOnErr(err)
		return
	}
	interactiveDeleteLfsFiles(filesToDelete, gitLfsCleanFlags)
}

func interactiveDeleteLfsFiles(filesToDelete []rtclientutils.ResultItem, flags *commands.GitLfsCleanConfiguration) {
	for _, v := range filesToDelete {
		fmt.Println("  " + v.Name)
	}
	confirmed := cliutils.InteractiveConfirm("Are you sure you want to delete the above files?")
	if confirmed {
		err := commands.DeleteLfsFilesFromArtifactory(filesToDelete, flags)
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
	val, err = cliutils.GetBoolEnvValue("JFROG_CLI_OFFER_CONFIG", true)
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
	encPassword := cliutils.GetBoolFlagValue(c, "enc-password", true)
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
	details.ServerId = c.String("server-id")

	if includeConfig {
		confDetails, err := commands.GetConfig(c.String("server-id"))
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
		}
	}
	details.Url = cliutils.AddTrailingSlashIfNeeded(details.Url)
	return
}

func isAuthMethodSet(details *config.ArtifactoryDetails) bool {
	return (details.User != "" && details.Password != "") || details.SshKeyPath != "" || details.ApiKey != ""
}

func getDebFlag(c *cli.Context) (deb string) {
	deb = c.String("deb")
	if deb != "" && len(strings.Split(deb, "/")) != 3 {
		cliutils.Exit(cliutils.ExitCodeError, "The --deb option should be in the form of distribution/component/architecture")
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
		Recursive(cliutils.GetBoolFlagValue(c, "recursive", true)).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Flat(cliutils.GetBoolFlagValue(c, "flat", false)).
		IncludeDirs(true).
		Target(c.Args().Get(1)).
		BuildSpec()
}

func getCopyMoveSpec(c *cli.Context) (searchSpec *spec.SpecFiles) {
	searchSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)

	//Override spec with CLI options
	for i := 0; i < len(searchSpec.Files); i++ {
		overrideFieldsIfSet(searchSpec.Get(i), c)
	}
	err = spec.ValidateSpec(searchSpec.Files, true)
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
		Recursive(cliutils.GetBoolFlagValue(c, "recursive", true)).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		BuildSpec()
}

func getDeleteSpec(c *cli.Context) (deleteSpec *spec.SpecFiles) {
	deleteSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)

	//Override spec with CLI options
	for i := 0; i < len(deleteSpec.Files); i++ {
		overrideFieldsIfSet(deleteSpec.Get(i), c)
	}
	err = spec.ValidateSpec(deleteSpec.Files, false)
	cliutils.ExitOnErr(err)
	return
}

func createDeleteFlags(c *cli.Context) (deleteFlags *commands.DeleteConfiguration) {
	deleteFlags = new(commands.DeleteConfiguration)
	deleteFlags.DryRun = c.Bool("dry-run")
	deleteFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
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
		Recursive(cliutils.GetBoolFlagValue(c, "recursive", true)).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		BuildSpec()
}

func createDefaultSetPropertiesSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Props(c.String("props")).
		Build(c.String("build")).
		Offset(getIntValue("offset", c)).
		Limit(getIntValue("limit", c)).
		SortOrder(c.String("sort-order")).
		SortBy(cliutils.GetStringsArrFlagValue(c, "sort-by")).
		Recursive(cliutils.GetBoolFlagValue(c, "recursive", true)).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		IncludeDirs(cliutils.GetBoolFlagValue(c, "include-dirs", false)).
		BuildSpec()
}

func getSearchSpec(c *cli.Context) (searchSpec *spec.SpecFiles) {
	searchSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)
	//Override spec with CLI options
	for i := 0; i < len(searchSpec.Files); i++ {
		overrideFieldsIfSet(searchSpec.Get(i), c)
	}
	return
}

func createBuildInfoFlags(c *cli.Context) (flags *buildinfo.Flags, artDetails *config.ArtifactoryDetails) {
	flags = new(buildinfo.Flags)
	artDetails = createArtifactoryDetailsByFlags(c, true)
	flags.DryRun = c.Bool("dry-run")
	flags.EnvInclude = c.String("env-include")
	flags.EnvExclude = c.String("env-exclude")
	if len(flags.EnvInclude) == 0 {
		flags.EnvInclude = "*"
	}
	if len(flags.EnvExclude) == 0 {
		flags.EnvExclude = "*password*;*secret*;*key*;*token*"
	}
	return
}

func createBuildPromoteFlags(c *cli.Context) (promoteFlags *commands.BuildPromotionConfiguration) {
	promotionParamsImpl := new(services.PromotionParamsImpl)
	promotionParamsImpl.Comment = c.String("comment")
	promotionParamsImpl.SourceRepo = c.String("source-repo")
	promotionParamsImpl.Status = c.String("status")
	promotionParamsImpl.IncludeDependencies = c.Bool("include-dependencies")
	promotionParamsImpl.Copy = c.Bool("copy")
	promoteFlags = new(commands.BuildPromotionConfiguration)
	promoteFlags.DryRun = c.Bool("dry-run")
	promoteFlags.PromotionParamsImpl = promotionParamsImpl
	promoteFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createBuildDistributionFlags(c *cli.Context) (distributeFlags *commands.BuildDistributionConfiguration) {
	distributeParamsImpl := new(services.BuildDistributionParamsImpl)
	distributeParamsImpl.Publish = cliutils.GetBoolFlagValue(c, "publish", true)
	distributeParamsImpl.OverrideExistingFiles = c.Bool("override")
	distributeParamsImpl.GpgPassphrase = c.String("passphrase")
	distributeParamsImpl.Async = c.Bool("async")
	distributeParamsImpl.SourceRepos = c.String("source-repos")
	distributeFlags = new(commands.BuildDistributionConfiguration)
	distributeFlags.DryRun = c.Bool("dry-run")
	distributeFlags.BuildDistributionParamsImpl = distributeParamsImpl
	distributeFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createGitLfsCleanFlags(c *cli.Context) (gitLfsCleanFlags *commands.GitLfsCleanConfiguration) {
	gitLfsCleanFlags = new(commands.GitLfsCleanConfiguration)
	refs := c.String("refs")
	if len(refs) == 0 {
		refs = "refs/remotes/*"
	}
	repo := c.String("repo")
	gitLfsCleanFlags.GitLfsCleanParamsImpl = &services.GitLfsCleanParamsImpl{Repo: repo, Refs: refs}
	gitLfsCleanFlags.Quiet = c.Bool("quiet")
	gitLfsCleanFlags.DryRun = c.Bool("dry-run")
	gitLfsCleanFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
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
		Recursive(cliutils.GetBoolFlagValue(c, "recursive", true)).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Flat(cliutils.GetBoolFlagValue(c, "flat", false)).
		IncludeDirs(cliutils.GetBoolFlagValue(c, "include-dirs", false)).
		Target(c.Args().Get(1)).
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
	err = spec.ValidateSpec(downloadSpec.Files, false)
	cliutils.ExitOnErr(err)
	return
}

func createDownloadFlags(c *cli.Context) (downloadFlags *commands.DownloadConfiguration) {
	downloadFlags = new(commands.DownloadConfiguration)
	downloadFlags.DryRun = c.Bool("dry-run")
	downloadFlags.ValidateSymlink = c.Bool("validate-symlinks")
	downloadFlags.MinSplitSize = getMinSplit(c)
	downloadFlags.SplitCount = getSplitCount(c)
	downloadFlags.Threads = getThreadsCount(c)
	downloadFlags.BuildName = getBuildName(c)
	downloadFlags.BuildNumber = getBuildNumber(c)
	downloadFlags.Retries = getRetries(c)
	downloadFlags.Symlink = true
	if (downloadFlags.BuildName == "" && downloadFlags.BuildNumber != "") || (downloadFlags.BuildName != "" && downloadFlags.BuildNumber == "") {
		cliutils.Exit(cliutils.ExitCodeError, "The build-name and build-number options cannot be sent separately.")
	}
	downloadFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
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
		Recursive(cliutils.GetBoolFlagValue(c, "recursive", true)).
		ExcludePatterns(cliutils.GetStringsArrFlagValue(c, "exclude-patterns")).
		Flat(cliutils.GetBoolFlagValue(c, "flat", true)).
		Regexp(c.Bool("regexp")).
		IncludeDirs(c.Bool("include-dirs")).
		Target(strings.TrimPrefix(c.Args().Get(1), "/")).
		BuildSpec()
}

func getUploadSpec(c *cli.Context) (uploadSpec *spec.SpecFiles) {
	uploadSpec, err := spec.CreateSpecFromFile(c.String("spec"), cliutils.SpecVarsStringToMap(c.String("spec-vars")))
	cliutils.ExitOnErr(err)
	//Override spec with CLI options
	for i := 0; i < len(uploadSpec.Files); i++ {
		uploadSpec.Get(i).Target = strings.TrimPrefix(uploadSpec.Get(i).Target, "/")
		overrideFieldsIfSet(uploadSpec.Get(i), c)
	}
	fixWinUploadFilesPath(uploadSpec)
	err = spec.ValidateSpec(uploadSpec.Files, true)
	cliutils.ExitOnErr(err)
	return
}

func fixWinUploadFilesPath(uploadSpec *spec.SpecFiles) {
	if runtime.GOOS == "windows" {
		for i, file := range uploadSpec.Files {
			uploadSpec.Files[i].Pattern = strings.Replace(file.Pattern, "\\", "\\\\", -1)
			for j, excludePattern := range uploadSpec.Files[i].ExcludePatterns {
				uploadSpec.Files[i].ExcludePatterns[j] = strings.Replace(excludePattern, "\\", "\\\\", -1)
			}
		}
	}
}

func fixWinDownloadFilesPath(uploadSpec *spec.SpecFiles) {
	if runtime.GOOS == "windows" {
		for i, file := range uploadSpec.Files {
			uploadSpec.Files[i].Target = strings.Replace(file.Target, "\\", "\\\\", -1)
		}
	}
}

func createUploadFlags(c *cli.Context) (uploadFlags *commands.UploadConfiguration) {
	uploadFlags = new(commands.UploadConfiguration)
	buildName := getBuildName(c)
	buildNumber := getBuildNumber(c)
	if (buildName == "" && buildNumber != "") || (buildName != "" && buildNumber == "") {
		cliutils.Exit(cliutils.ExitCodeError, "The build-name and build-number options cannot be sent separately.")
	}
	uploadFlags.BuildName = buildName
	uploadFlags.BuildNumber = buildNumber
	uploadFlags.DryRun = c.Bool("dry-run")
	uploadFlags.ExplodeArchive = c.Bool("explode")
	uploadFlags.Symlink = c.Bool("symlinks")
	uploadFlags.Threads = getThreadsCount(c)
	uploadFlags.Deb = getDebFlag(c)
	uploadFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	return
}

func createBuildToolFlags(c *cli.Context) (buildConfigFlags *utils.BuildConfigFlags) {
	buildConfigFlags = new(utils.BuildConfigFlags)
	buildConfigFlags.BuildName = getBuildName(c)
	buildConfigFlags.BuildNumber = getBuildNumber(c)
	if (buildConfigFlags.BuildName == "" && buildConfigFlags.BuildNumber != "") || (buildConfigFlags.BuildName != "" && buildConfigFlags.BuildNumber == "") {
		cliutils.Exit(cliutils.ExitCodeError, "The build-name and build-number options cannot be sent separately.")
	}
	return
}

func createConfigFlags(c *cli.Context) (configFlag *commands.ConfigFlags) {
	configFlag = new(commands.ConfigFlags)
	configFlag.ArtDetails = createArtifactoryDetails(c, false)
	configFlag.EncPassword = cliutils.GetBoolFlagValue(c, "enc-password", true)
	configFlag.Interactive = cliutils.GetBoolFlagValue(c, "interactive", true)
	return
}

func validateConfigFlags(configFlag *commands.ConfigFlags) {
	if !configFlag.Interactive && configFlag.ArtDetails.Url == "" {
		cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory when the --interactive option is set to false")
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
	if c.IsSet("build") && c.IsSet("offset") {
		cliutils.Exit(cliutils.ExitCodeError, "The 'offset' option cannot be used together with the 'build' option")
	}
	if c.IsSet("build") && c.IsSet("limit") {
		cliutils.Exit(cliutils.ExitCodeError, "The 'limit' option cannot be used together with the 'build' option")
	}
	if c.IsSet("sort-order") && !c.IsSet("sort-by") {
		cliutils.Exit(cliutils.ExitCodeError, "The 'sort-order' option cannot be used without the 'sort-by' option")
	}
	if c.IsSet("sort-order") && !(c.String("sort-order") == "asc" || c.String("sort-order") == "desc") {
		cliutils.Exit(cliutils.ExitCodeError, "The 'sort-order' option can only accept 'asc' or 'desc' as values")
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
	overrideStringIfSet(&spec.Regexp, c, "regexp")
	overrideStringIfSet(&spec.IncludeDirs, c, "include-dirs")
}

func getIntValue(key string, c *cli.Context) int {
	value, err := cliutils.GetIntFlagValue(c, key, 0)
	cliutils.ExitOnErr(err)
	return value
}

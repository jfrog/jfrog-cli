package artifactory

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/commands"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"strconv"
	"strings"
	"encoding/json"
	"runtime"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "config",
			Flags:   getConfigFlags(),
			Aliases: []string{"c"},
			Usage:   "Configure Artifactory details.",
			Action: func(c *cli.Context) {
				configCmd(c)
			},
		},
		{
			Name:    "upload",
			Flags:   getUploadFlags(),
			Aliases: []string{"u"},
			Usage:   "Upload files.",
			Action: func(c *cli.Context) {
				uploadCmd(c)
			},
		},
		{
			Name:    "download",
			Flags:   getDownloadFlags(),
			Aliases: []string{"dl"},
			Usage:   "Download files.",
			Action: func(c *cli.Context) {
				downloadCmd(c)
			},
		},
		{
			Name:    "move",
			Flags:   getMoveFlags(),
			Aliases: []string{"mv"},
			Usage:   "Move files.",
			Action: func(c *cli.Context) {
				moveCmd(c)
			},
		},
		{
			Name:    "copy",
			Flags:   getCopyFlags(),
			Aliases: []string{"cp"},
			Usage:   "Copy files.",
			Action: func(c *cli.Context) {
				copyCmd(c)
			},
		},
		{
			Name:    "delete",
			Flags:   getDeleteFlags(),
			Aliases: []string{"del"},
			Usage:   "Delete files.",
			Action: func(c *cli.Context) {
				deleteCmd(c)
			},
		},
		{
			Name:    "search",
			Flags:   getSearchFlags(),
			Aliases: []string{"s"},
			Usage:   "Search files.",
			Action: func(c *cli.Context) {
				searchCmd(c)
			},
		},
		{
			Name:    "build-publish",
			Flags:   getBuildPublishFlags(),
			Aliases: []string{"bp"},
			Usage:   "Publish build info.",
			Action: func(c *cli.Context) {
				buildPublishCmd(c)
			},
		},
		{
			Name:    "build-collect-env",
			Flags:    []cli.Flag{},
			Aliases: []string{"bce"},
			Usage:   "Capture environment varaibles.",
			Action: func(c *cli.Context) {
				buildCollectEnvCmd(c)
			},
		},
		{
			Name:    "build-clean",
			Flags:    []cli.Flag{},
			Aliases: []string{"bc"},
			Usage:   "Clean build.",
			Action: func(c *cli.Context) {
				buildCleanCmd(c)
			},
		},
		{
			Name:    "build-promote",
			Flags:   getBuildPromotionFlags(),
			Aliases: []string{"bpr"},
			Usage:   "Promote build.",
			Action: func(c *cli.Context) {
				buildPromoteCmd(c)
			},
		},
		{
			Name:    "build-distribute",
			Flags:   getBuildDistributeFlags(),
			Aliases: []string{"bd"},
			Usage:   "Distribute build.",
			Action: func(c *cli.Context) {
				buildDistributeCmd(c)
			},
		},
	}
}

func getFlags() []cli.Flag {
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
	}
}

func getUploadFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a spec file.",
		},
		cli.StringFlag{
			Name:  "build-name",
			Usage: "[Optional] Build name.",
		},
		cli.StringFlag{
			Name:  "build-number",
			Usage: "[Optional] Build number.",
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
	}...)
}

func getDownloadFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a spec file.",
		},
		cli.StringFlag{
			Name:  "build-name",
			Usage: "[Optional] Build name.",
		},
		cli.StringFlag{
			Name:  "build-number",
			Usage: "[Optional] Build number.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\" Only artifacts with these properties will be downloaded.",
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
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Artifactory.",
		},
		cli.BoolFlag{
			Name:  "symlinks",
			Usage: "[Default: false] Set to true create symlinks as represented in Artifactory.",
		},
		cli.BoolFlag{
			Name:  "validate-symlinks",
			Usage: "[Default: false] Set to true to perform a checksum validation when downloading symbolic links.",
		},
	}...)
}

func getMoveFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a spec file.",
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
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\" Only artifacts with these properties will be moved.",
		},
	}...)

}

func getCopyFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a spec file.",
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
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\" Only artifacts with these properties will be copied.",
		},
	}...)
}

func getDeleteFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a spec file.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\" Only artifacts with these properties will be deleted.",
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
	}...)
}

func getSearchFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "spec",
			Usage: "[Optional] Path to a spec file.",
		},
		cli.StringFlag{
			Name:  "props",
			Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\" Only artifacts with these properties will be returned.",
		},
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not wish to search artifacts inside sub-folders in Artifactory.",
		},
	}...)
}

func getBuildPublishFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
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
			Usage: "[Default: *password*;*secret*;*key*] List of patterns in the form of \"value1;value2;...\"  environment variables match those patterns will be eccluded.",
		},
	}...)
}

func getBuildPromotionFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
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
	return append(getFlags(), []cli.Flag{
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
	return append(flags, getFlags()...)
}

func createArtifactoryDetailsByFlags(c *cli.Context, includeConfig bool) (*config.ArtifactoryDetails, error) {
	artDetails, err := createArtifactoryDetails(c, includeConfig)
	if err != nil {
		return nil, err
	}
	if artDetails.Url == "" {
		cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory")
	}
	return artDetails, nil
}

func getSplitCount(c *cli.Context) (splitCount int) {
	var err error
	splitCount = 3
	if c.String("split-count") != "" {
		splitCount, err = strconv.Atoi(c.String("split-count"))
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option should have a numeric value. " + cliutils.GetDocumentationMessage())
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
			cliutils.Exit(cliutils.ExitCodeError, "The '--min-split' option should have a numeric value. " + cliutils.GetDocumentationMessage())
		}
	}
	return
}

func configCmd(c *cli.Context) {
	if len(c.Args()) > 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	} else if len(c.Args()) == 1 {
		if c.Args()[0] == "show" {
			commands.ShowConfig()
		} else if c.Args()[0] == "clear" {
			commands.ClearConfig()
		} else {
			cliutils.Exit(cliutils.ExitCodeError, "Unknown argument '" + c.Args()[0] + "'. Available arguments are 'show' and 'clear'.")
		}
	} else {
		configFlags, err := createConfigFlags(c)
		cliutils.ExitOnErr(err)
		_, err = commands.Config(configFlags.ArtDetails, nil, configFlags.Interactive, configFlags.EncPassword)
		cliutils.ExitOnErr(err)
	}
}

func downloadCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.Exit(cliutils.ExitCodeError, "No arguments should be sent when the spec option is used. " + cliutils.GetDocumentationMessage())
	}
	if !(c.NArg() == 1 || c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}

	var downloadSpec *utils.SpecFiles
	if c.IsSet("spec") {
		var err error
		downloadSpec, err = getDownloadSpec(c)
		cliutils.ExitOnErr(err)
	} else {
		downloadSpec = createDefaultDownloadSpec(c)
	}

	flags, err := createDownloadFlags(c)
	cliutils.ExitOnErr(err)
	err = commands.Download(downloadSpec, flags)
	cliutils.ExitOnErr(err)
}

func uploadCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.Exit(cliutils.ExitCodeError, "No arguments should be sent when the spec option is used. " + cliutils.GetDocumentationMessage())
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}

	var uploadSpec *utils.SpecFiles
	if c.IsSet("spec") {
		var err error
		uploadSpec, err = getUploadSpec(c)
		cliutils.ExitOnErr(err)
	} else {
		uploadSpec = createDefaultUploadSpec(c)
	}

	flags, err := createUploadFlags(c)
	cliutils.ExitOnErr(err)
	uploaded, failed, err := commands.Upload(uploadSpec, flags)
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
		cliutils.Exit(cliutils.ExitCodeError, "No arguments should be sent when the spec option is used. " + cliutils.GetDocumentationMessage())
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}

	var moveSpec *utils.SpecFiles
	if c.IsSet("spec") {
		var err error
		moveSpec, err = getMoveSpec(c)
		cliutils.ExitOnErr(err)
	} else {
		moveSpec = createDefaultMoveSpec(c)
	}

	flags, err := createMoveFlags(c)
	cliutils.ExitOnErr(err)
	err = commands.Move(moveSpec, flags)
	cliutils.ExitOnErr(err)
}

func copyCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.Exit(cliutils.ExitCodeError, "No arguments should be sent when the spec option is used. " + cliutils.GetDocumentationMessage())
	}
	if !(c.NArg() == 2 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}

	var copySpec *utils.SpecFiles
	if c.IsSet("spec") {
		var err error
		copySpec, err = getMoveSpec(c)
		cliutils.ExitOnErr(err)
	} else {
		copySpec = createDefaultMoveSpec(c)
	}

	flags, err := createMoveFlags(c)
	cliutils.ExitOnErr(err)
	err = commands.Copy(copySpec, flags)
	cliutils.ExitOnErr(err)
}

func deleteCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.Exit(cliutils.ExitCodeError, "No arguments should be sent when the spec option is used. " + cliutils.GetDocumentationMessage())
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}

	var deleteSpec *utils.SpecFiles
	if c.IsSet("spec") {
		var err error
		deleteSpec, err = getDeleteSpec(c)
		cliutils.ExitOnErr(err)
	} else {
		deleteSpec = createDefaultDeleteSpec(c)
	}

	flags, err := createDeleteFlags(c)
	cliutils.ExitOnErr(err)
	if !c.Bool("quiet") {
		pathsToDelete, err := commands.GetPathsToDelete(deleteSpec, flags)
		cliutils.ExitOnErr(err)
		if len(pathsToDelete) < 1 {
			return
		}
		for _, v := range pathsToDelete {
			fmt.Println("  " + v.GetFullUrl())
		}
		var confirm string
		fmt.Print("Are you sure you want to delete the above paths? (y/n): ")
		fmt.Scanln(&confirm)
		if !cliutils.ConfirmAnswer(confirm) {
			return
		}
		err = commands.DeleteFiles(pathsToDelete, flags)
		cliutils.ExitOnErr(err)
	} else {
		err = commands.Delete(deleteSpec, flags)
		cliutils.ExitOnErr(err)
	}
}

func searchCmd(c *cli.Context) {
	if c.NArg() > 0 && c.IsSet("spec") {
		cliutils.Exit(cliutils.ExitCodeError, "No arguments should be sent when the spec option is used. " + cliutils.GetDocumentationMessage())
	}
	if !(c.NArg() == 1 || (c.NArg() == 0 && c.IsSet("spec"))) {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}

	var searchSpec *utils.SpecFiles
	if c.IsSet("spec") {
		var err error
		searchSpec, err = getSearchSpec(c)
		cliutils.ExitOnErr(err)
	} else {
		searchSpec = createDefaultSearchSpec(c)
	}

	flags, err := createSearchFlags(c)
	cliutils.ExitOnErr(err)
	SearchResult, err := commands.Search(searchSpec, flags)
	cliutils.ExitOnErr(err)
	result, err := json.Marshal(SearchResult)
	cliutils.ExitOnErr(err)

	fmt.Println(string(cliutils.IndentJson(result)))
}

func buildPublishCmd(c *cli.Context) {
	vlidateBuildInfoArgument(c)
	buildInfoFlags, err := createBuildInfoFlags(c)
	cliutils.ExitOnErrWithMsg(err)
	err = commands.BuildPublish(c.Args().Get(0), c.Args().Get(1), buildInfoFlags)
	cliutils.ExitOnErr(err)
}

func buildCollectEnvCmd(c *cli.Context) {
	vlidateBuildInfoArgument(c)
	err := commands.BuildCollectEnv(c.Args().Get(0), c.Args().Get(1))
	cliutils.ExitOnErr(err)
}

func buildCleanCmd(c *cli.Context) {
	vlidateBuildInfoArgument(c)
	err := commands.BuildClean(c.Args().Get(0), c.Args().Get(1))
	cliutils.ExitOnErr(err)
}

func buildPromoteCmd(c *cli.Context) {
	if c.NArg() != 3 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	buildPromoteFlags, err := createBuildPromoteFlags(c)
	if err != nil {
		cliutils.ExitOnErr(err)
	}
	err = commands.BuildPromote(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), buildPromoteFlags)
	cliutils.ExitOnErr(err)
}

func buildDistributeCmd(c *cli.Context) {
	if c.NArg() != 3 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	buildDistributeFlags, err := createBuildDistributionFlags(c)
	if err != nil {
		cliutils.ExitOnErr(err)
	}
	err = commands.BuildDistribute(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), buildDistributeFlags)
	cliutils.ExitOnErr(err)
}

func vlidateBuildInfoArgument(c *cli.Context) {
	if c.NArg() != 2 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
}

func offerConfig(c *cli.Context) (details *config.ArtifactoryDetails, err error) {
	var exists bool
	exists, err = config.IsArtifactoryConfExists()
	if err != nil {
		return
	}
	if exists {
		return
	}
	var val bool
	val, err = cliutils.GetBoolEnvValue("JFROG_CLI_OFFER_CONFIG", true)
	if err != nil {
		return
	}
	if !val {
		config.SaveArtifactoryConf(new(config.ArtifactoryDetails))
		return
	}
	msg := "The CLI commands require the Artifactory URL and authentication details\n" +
			"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n" +
			"You can also configure these parameters later using the 'config' command.\n" +
			"Configure now? (y/n): "
	fmt.Print(msg)
	var confirm string
	fmt.Scanln(&confirm)
	if !cliutils.ConfirmAnswer(confirm) {
		config.SaveArtifactoryConf(new(config.ArtifactoryDetails))
		return
	}
	details, err = createArtifactoryDetails(c, false)
	if err != nil {
		return
	}
	encPassword := cliutils.GetBoolFlagValue(c, "enc-password", true)
	details, err = commands.Config(nil, details, true, encPassword)
	return
}

func createArtifactoryDetails(c *cli.Context, includeConfig bool) (*config.ArtifactoryDetails, error) {
	if includeConfig {
		details, err := offerConfig(c)
		if err != nil {
			return nil, err
		}
		if details != nil {
			return details, nil
		}
	}
	details := new(config.ArtifactoryDetails)
	details.Url = c.String("url")
	details.ApiKey = c.String("apikey")
	details.User = c.String("user")
	details.Password = c.String("password")
	details.SshKeyPath = c.String("ssh-key-path")

	if includeConfig {
		confDetails, err := commands.GetConfig()
		if err != nil {
			return nil, err
		}
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
	return details, nil
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

func createDefaultMoveSpec(c *cli.Context) *utils.SpecFiles {
	pattern := c.Args().Get(0)
	target := c.Args().Get(1)
	props := c.String("props")
	recursive := cliutils.GetBoolFlagValue(c, "recursive", true)
	flat := cliutils.GetBoolFlagValue(c, "flat", false)

	return utils.CreateSpec(pattern, target, props, recursive, flat, false)
}

func getMoveSpec(c *cli.Context) (searchSpec *utils.SpecFiles, err error) {
	searchSpec, err = utils.CreateSpecFromFile(c.String("spec"))
	if err != nil {
		return
	}
	//Override spec with CLI options
	for i := 0; i < len(searchSpec.Files); i++ {
		overrideStringIfSet(&searchSpec.Get(i).Props, c, "props")
		overrideStringIfSet(&searchSpec.Get(i).Recursive, c, "recursive")
		overrideStringIfSet(&searchSpec.Get(i).Flat, c, "flat")
	}
	return
}

func createMoveFlags(c *cli.Context) (moveFlags *utils.MoveFlags, err error) {
	moveFlags = new(utils.MoveFlags)
	moveFlags.DryRun = c.Bool("dry-run")
	moveFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	return
}

func createDefaultDeleteSpec(c *cli.Context) *utils.SpecFiles {
	pattern := c.Args().Get(0)
	props := c.String("props")
	recursive := cliutils.GetBoolFlagValue(c, "recursive", true)

	return utils.CreateSpec(pattern, "", props, recursive, false, false)
}

func getDeleteSpec(c *cli.Context) (searchSpec *utils.SpecFiles, err error) {
	searchSpec, err = utils.CreateSpecFromFile(c.String("spec"))
	if err != nil {
		return
	}

	//Override spec with CLI options
	for i := 0; i < len(searchSpec.Files); i++ {
		overrideStringIfSet(&searchSpec.Get(i).Props, c, "props")
		overrideStringIfSet(&searchSpec.Get(i).Recursive, c, "recursive")
	}
	return
}

func createDeleteFlags(c *cli.Context) (deleteFlags *commands.DeleteFlags, err error) {
	deleteFlags = new(commands.DeleteFlags)
	deleteFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	if err != nil {
		return
	}
	deleteFlags.DryRun = c.Bool("dry-run")
	return
}

func createDefaultSearchSpec(c *cli.Context) *utils.SpecFiles {
	pattern := c.Args().Get(0)
	props := c.String("props")
	recursive := cliutils.GetBoolFlagValue(c, "recursive", true)

	return utils.CreateSpec(pattern, "", props, recursive, false, false)
}

func getSearchSpec(c *cli.Context) (searchSpec *utils.SpecFiles, err error) {
	searchSpec, err = utils.CreateSpecFromFile(c.String("spec"))
	if err != nil {
		return
	}
	//Override spec with CLI options
	for i := 0; i < len(searchSpec.Files); i++ {
		overrideStringIfSet(&searchSpec.Get(i).Props, c, "props")
		overrideStringIfSet(&searchSpec.Get(i).Recursive, c, "recursive")
	}
	return
}

func createSearchFlags(c *cli.Context) (searchFlags *commands.SearchFlags, err error) {
	searchFlags = new(commands.SearchFlags)
	searchFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	return
}

func createBuildInfoFlags(c *cli.Context) (flags *utils.BuildInfoFlags, err error) {
	flags = new(utils.BuildInfoFlags)
	flags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	flags.DryRun = c.Bool("dry-run")
	flags.EnvInclude = c.String("env-include")
	flags.EnvExclude = c.String("env-exclude")
	if len(flags.EnvInclude) == 0 {
		flags.EnvInclude = "*"
	}
	if len(flags.EnvExclude) == 0 {
		flags.EnvExclude = "*password*;*secret*;*key*"
	}
	return
}

func createBuildPromoteFlags(c *cli.Context) (promoteFlags *commands.BuildPromotionFlags, err error) {
	promoteFlags = new(commands.BuildPromotionFlags)
	promoteFlags.Comment = c.String("comment")
	promoteFlags.SourceRepo = c.String("source-repo")
	promoteFlags.Status = c.String("status")
	promoteFlags.IncludeDependencies = c.Bool("include-dependencies")
	promoteFlags.Copy = c.Bool("copy")
	promoteFlags.DryRun = c.Bool("dry-run")

	promoteFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	return
}

func createBuildDistributionFlags(c *cli.Context) (distributeFlags *commands.BuildDistributionFlags, err error) {
	distributeFlags = new(commands.BuildDistributionFlags)
	distributeFlags.Publish = cliutils.GetBoolFlagValue(c, "publish", true)
	distributeFlags.OverrideExistingFiles = c.Bool("override")
	distributeFlags.GpgPassphrase = c.String("passphrase")
	distributeFlags.Async = c.Bool("async")
	distributeFlags.SourceRepos = c.String("source-repos")
	distributeFlags.DryRun = c.Bool("dry-run")

	distributeFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	return
}

func createDefaultDownloadSpec(c *cli.Context) *utils.SpecFiles {
	pattern := strings.TrimPrefix(c.Args().Get(0), "/")
	target := c.Args().Get(1)
	props := c.String("props")
	recursive := cliutils.GetBoolFlagValue(c, "recursive", true)
	flat := cliutils.GetBoolFlagValue(c, "flat", false)

	return utils.CreateSpec(pattern, target, props, recursive, flat, false)
}

func getDownloadSpec(c *cli.Context) (downloadSpec *utils.SpecFiles, err error) {
	downloadSpec, err = utils.CreateSpecFromFile(c.String("spec"))
	if err != nil {
		return
	}
	fixWinDownloadFilesPath(downloadSpec)
	//Override spec with CLI options
	for i := 0; i < len(downloadSpec.Files); i++ {
		downloadSpec.Get(i).Pattern = strings.TrimPrefix(downloadSpec.Get(i).Pattern, "/")
		overrideStringIfSet(&downloadSpec.Get(i).Props, c, "props")
		overrideStringIfSet(&downloadSpec.Get(i).Flat, c, "flat")
		overrideStringIfSet(&downloadSpec.Get(i).Recursive, c, "recursive")
	}
	return
}

func createDownloadFlags(c *cli.Context) (downloadFlags *commands.DownloadFlags, err error) {
	downloadFlags = new(commands.DownloadFlags)
	downloadFlags.DryRun = c.Bool("dry-run")
	downloadFlags.Symlink = c.Bool("symlinks")
	downloadFlags.ValidateSymlink = c.Bool("validate-symlinks")
	downloadFlags.MinSplitSize = getMinSplit(c)
	downloadFlags.SplitCount = getSplitCount(c)
	downloadFlags.Threads = getThreadsCount(c);
	downloadFlags.BuildName = getBuildName(c);
	downloadFlags.BuildNumber = getBuildNumber(c);
	if (downloadFlags.BuildName == "" && downloadFlags.BuildNumber != "") || (downloadFlags.BuildName != "" && downloadFlags.BuildNumber == "") {
		cliutils.Exit(cliutils.ExitCodeError, "The build-name and build-number options cannot be sent separately.")
	}
	downloadFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	return
}

func createDefaultUploadSpec(c *cli.Context) *utils.SpecFiles {
	pattern := c.Args().Get(0)
	target := strings.TrimPrefix(c.Args().Get(1), "/")
	props := c.String("props")
	recursive := cliutils.GetBoolFlagValue(c, "recursive", true)
	flat := cliutils.GetBoolFlagValue(c, "flat", true)
	regexp := c.Bool("regexp")

	return utils.CreateSpec(pattern, target, props, recursive, flat, regexp)
}

func getUploadSpec(c *cli.Context) (uploadSpec *utils.SpecFiles, err error) {
	uploadSpec, err = utils.CreateSpecFromFile(c.String("spec"))
	if err != nil {
		return
	}
	fixWinUploadFilesPath(uploadSpec)
	//Override spec with CLI options
	for i := 0; i < len(uploadSpec.Files); i++ {
		uploadSpec.Get(i).Target = strings.TrimPrefix(uploadSpec.Get(i).Target, "/")
		overrideStringIfSet(&uploadSpec.Get(i).Props, c, "props")
		overrideStringIfSet(&uploadSpec.Get(i).Flat, c, "flat")
		overrideStringIfSet(&uploadSpec.Get(i).Recursive, c, "recursive")
		overrideStringIfSet(&uploadSpec.Get(i).Regexp, c, "regexp")
	}
	return
}

func fixWinUploadFilesPath(uploadSpec *utils.SpecFiles) {
	if runtime.GOOS == "windows" {
		for i, file := range uploadSpec.Files {
			uploadSpec.Files[i].Pattern = strings.Replace(file.Pattern, "\\", "\\\\", -1)
		}
	}
}

func fixWinDownloadFilesPath(uploadSpec *utils.SpecFiles) {
	if runtime.GOOS == "windows" {
		for i, file := range uploadSpec.Files {
			uploadSpec.Files[i].Target = strings.Replace(file.Target, "\\", "\\\\", -1)
		}
	}
}

func createUploadFlags(c *cli.Context) (uploadFlags *commands.UploadFlags, err error) {
	uploadFlags = new(commands.UploadFlags)
	uploadFlags.DryRun = c.Bool("dry-run")
	uploadFlags.ExplodeArchive = c.Bool("explode")
	uploadFlags.Symlink = c.Bool("symlinks")
	uploadFlags.Threads = getThreadsCount(c);
	uploadFlags.BuildName = getBuildName(c);
	uploadFlags.BuildNumber = getBuildNumber(c);
	if (uploadFlags.BuildName == "" && uploadFlags.BuildNumber != "") || (uploadFlags.BuildName != "" && uploadFlags.BuildNumber == "") {
		cliutils.Exit(cliutils.ExitCodeError, "The build-name and build-number options cannot be sent separately.")
	}
	uploadFlags.Deb = getDebFlag(c);
	uploadFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	return
}

func createConfigFlags(c *cli.Context) (configFlag *commands.ConfigFlags, err error) {
	configFlag = new(commands.ConfigFlags)
	configFlag.ArtDetails, err = createArtifactoryDetails(c, false)
	if err != nil {
		return
	}
	configFlag.EncPassword = cliutils.GetBoolFlagValue(c, "enc-password", true)
	configFlag.Interactive = cliutils.GetBoolFlagValue(c, "interactive", true)
	if !configFlag.Interactive && configFlag.ArtDetails.Url == "" {
		cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory when the --interactive option is set to false")
	}
	return
}

func overrideStringIfSet(field *string, c *cli.Context, fieldName string) {
	if c.IsSet(fieldName) {
		*field = c.String(fieldName)
	}
}

func overrideBoolIfSet(field *bool, c *cli.Context, fieldName string) {
	if c.IsSet(fieldName) {
		*field = c.Bool(fieldName)
	}
}


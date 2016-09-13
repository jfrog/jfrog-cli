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
	"runtime"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "config",
			Flags:   getConfigFlags(),
			Aliases: []string{"c"},
			Usage:   "Configure Artifactory details",
			Action: func(c *cli.Context) {
				configCmd(c)
			},
		},
		{
			Name:    "upload",
			Flags:   getUploadFlags(),
			Aliases: []string{"u"},
			Usage:   "Upload files",
			Action: func(c *cli.Context) {
				uploadCmd(c)
			},
		},
		{
			Name:    "download",
			Flags:   getDownloadFlags(),
			Aliases: []string{"dl"},
			Usage:   "Download files",
			Action: func(c *cli.Context) {
				downloadCmd(c)
			},
		},
		{
			Name:    "move",
			Flags:   getMoveFlags(),
			Aliases: []string{"mv"},
			Usage:   "Move files",
			Action: func(c *cli.Context) {
				moveCmd(c)
			},
		},
		{
			Name:    "copy",
			Flags:   getCopyFlags(),
			Aliases: []string{"cp"},
			Usage:   "Copy files",
			Action: func(c *cli.Context) {
				copyCmd(c)
			},
		},
		{
			Name:    "delete",
			Flags:   getDeleteFlags(),
			Aliases: []string{"del"},
			Usage:   "Delete files",
			Action: func(c *cli.Context) {
				deleteCmd(c)
			},
		},
		{
			Name:    "search",
			Flags:   getSearchFlags(),
			Aliases: []string{"s"},
			Usage:   "Search files",
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
	}
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Usage: "[Optional] Artifactory URL",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "[Optional] Artifactory username",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "[Optional] Artifactory password",
		},
		cli.StringFlag{
			Name:  "apikey",
			Usage: "[Optional] Artifactory API key",
		},
		cli.StringFlag{
			Name:  "ssh-key-path",
			Usage: "[Optional] SSH key file path",
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
		exitOnErr(err)
		_, err = commands.Config(configFlags.ArtDetails, nil, configFlags.Interactive, configFlags.EncPassword)
		exitOnErr(err)
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
		exitOnErr(err)
	} else {
		downloadSpec = createDefaultDownloadSpec(c)
	}

	flags, err := createDownloadFlags(c)
	exitOnErr(err)
	err = commands.Download(downloadSpec, flags)
	exitOnErr(err)
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
		exitOnErr(err)
	} else {
		uploadSpec = createDefaultUploadSpec(c)
	}

	flags, err := createUploadFlags(c)
	exitOnErr(err)
	uploaded, failed, err := commands.Upload(uploadSpec, flags)
	exitOnErr(err)
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
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, err.Error())
		}
	} else {
		moveSpec = createDefaultMoveSpec(c)
	}

	flags, err := createMoveFlags(c)
	exitOnErr(err)
	err = commands.Move(moveSpec, flags)
	exitOnErr(err)
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
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, err.Error())
		}
	} else {
		copySpec = createDefaultMoveSpec(c)
	}

	flags, err := createMoveFlags(c)
	exitOnErr(err)
	err = commands.Copy(copySpec, flags)
	exitOnErr(err)
}

func deleteCmd(c *cli.Context) {
	if len(c.Args()) != 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	path := c.Args()[0]
	if !c.Bool("quiet") {
		var confirm string
		fmt.Print("Delete path " + path + "? (y/n): ")
		fmt.Scanln(&confirm)
		if !cliutils.ConfirmAnswer(confirm) {
			return
		}
	}
	flags, err := createDeleteFlags(c)
	exitOnErr(err)
	err = commands.Delete(path, flags)
	exitOnErr(err)
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
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, err.Error())
		}
	} else {
		searchSpec = createDefaultSearchSpec(c)
	}

	flags, err := createSearchFlags(c)
	exitOnErr(err)
	err = commands.Search(searchSpec, flags)
	exitOnErr(err)
}

func buildPublishCmd(c *cli.Context) {
	vlidateBuildInfoArgument(c)
	buildInfoFlags, err := createBuildInfoFlags(c)
	exitOnErrWithMsg(err)
	err = commands.BuildPublish(c.Args().Get(0), c.Args().Get(1), buildInfoFlags)
	exitOnErr(err)
}

func buildCollectEnvCmd(c *cli.Context) {
	vlidateBuildInfoArgument(c)
	err := commands.BuildCollectEnv(c.Args().Get(0), c.Args().Get(1))
	exitOnErr(err)
}

func buildCleanCmd(c *cli.Context) {
	vlidateBuildInfoArgument(c)
	err := commands.BuildClean(c.Args().Get(0), c.Args().Get(1))
	exitOnErr(err)
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
	// Override options from user
	for i := 0; i < len(searchSpec.Files); i++ {
		overrideStringIfSet(&searchSpec.Get(i).Props, c, "props")
		overrideBoolIfSet(&searchSpec.Get(i).Recursive, c, "recursive")
		overrideBoolIfSet(&searchSpec.Get(i).Flat, c, "flat")
	}
	return
}

func createMoveFlags(c *cli.Context) (moveFlags *utils.MoveFlags, err error) {
	moveFlags = new(utils.MoveFlags)
	moveFlags.DryRun = c.Bool("dry-run")
	moveFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	return
}

func createDeleteFlags(c *cli.Context) (deleteFlags *commands.DeleteFlags, err error) {
	deleteFlags = new(commands.DeleteFlags)
	deleteFlags.ArtDetails, err = createArtifactoryDetailsByFlags(c, true)
	if err != nil {
		return
	}
	deleteFlags.Recursive = cliutils.GetBoolFlagValue(c, "recursive", true)
	deleteFlags.DryRun = c.Bool("dry-run")
	deleteFlags.Props = c.String("props")
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
	// Override options from user
	for i := 0; i < len(searchSpec.Files); i++ {
		overrideStringIfSet(&searchSpec.Get(i).Props, c, "props")
		overrideBoolIfSet(&searchSpec.Get(i).Recursive, c, "recursive")
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
	// Override options from user
	for i := 0; i < len(downloadSpec.Files); i++ {
		downloadSpec.Get(i).Pattern = strings.TrimPrefix(downloadSpec.Get(i).Pattern, "/")
		overrideStringIfSet(&downloadSpec.Get(i).Props, c, "props")
		overrideBoolIfSet(&downloadSpec.Get(i).Flat, c, "flat")
		overrideBoolIfSet(&downloadSpec.Get(i).Recursive, c, "recursive")
	}
	return
}

func createDownloadFlags(c *cli.Context) (downloadFlags *commands.DownloadFlags, err error) {
	downloadFlags = new(commands.DownloadFlags)
	downloadFlags.DryRun = c.Bool("dry-run")
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
	// Override options from user
	for i := 0; i < len(uploadSpec.Files); i++ {
		uploadSpec.Get(i).Target = strings.TrimPrefix(uploadSpec.Get(i).Target, "/")
		overrideStringIfSet(&uploadSpec.Get(i).Props, c, "props")
		overrideBoolIfSet(&uploadSpec.Get(i).Flat, c, "flat")
		overrideBoolIfSet(&uploadSpec.Get(i).Recursive, c, "recursive")
		overrideBoolIfSet(&uploadSpec.Get(i).Regexp, c, "regexp")
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

func exitOnErr(err error) {
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, "")
	}
}

func exitOnErrWithMsg(err error) {
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
}
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

func createArtifactoryDetailsByFlags(c *cli.Context, includeConfig bool) *config.ArtifactoryDetails {
	artDetails := createArtifactoryDetails(c, includeConfig)
	if artDetails.Url == "" {
		cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory")
	}
	return artDetails
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
		configFlags := createConfigFlags(c)
		commands.Config(configFlags.ArtDetails, nil, configFlags.Interactive, configFlags.EncPassword)
	}
}

func downloadCmd(c *cli.Context) {
	if len(c.Args()) != 1 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	pattern := strings.TrimPrefix(c.Args()[0], "/")
	commands.Download(pattern, createDownloadFlags(c))
}

func uploadCmd(c *cli.Context) {
	size := len(c.Args())
	if size != 2 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	localPath := c.Args()[0]
	targetPath := c.Args()[1]
	uploaded, failed := commands.Upload(localPath, targetPath, createUploadFlags(c))
	if failed > 0 {
		if uploaded > 0 {
			cliutils.Exit(cliutils.ExitCodeWarning, "")
		}
		cliutils.Exit(cliutils.ExitCodeError, "")
	}
}

func moveCmd(c *cli.Context) {
	if len(c.Args()) != 2 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	sourcePattern := c.Args()[0]
	targetPath := c.Args()[1]
	commands.Move(sourcePattern, targetPath, createMoveFlags(c))
}

func copyCmd(c *cli.Context) {
	if len(c.Args()) != 2 {
		cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. " + cliutils.GetDocumentationMessage())
	}
	sourcePattern := c.Args()[0]
	targetPath := c.Args()[1]
	commands.Copy(sourcePattern, targetPath, createMoveFlags(c))
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
	commands.Delete(path, createDeleteFlags(c))
}

func offerConfig(c *cli.Context) *config.ArtifactoryDetails {
    if config.IsArtifactoryConfExists() {
        return nil
    }
	if !cliutils.GetBoolEnvValue("JFROG_CLI_OFFER_CONFIG", true) {
		config.SaveArtifactoryConf(new(config.ArtifactoryDetails))
		return nil
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
        return nil
    }
    details := createArtifactoryDetails(c, false)

    encPassword := cliutils.GetBoolFlagValue(c, "enc-password", true)
    return commands.Config(nil, details, true, encPassword)
}

func createArtifactoryDetails(c *cli.Context, includeConfig bool) *config.ArtifactoryDetails {
	if includeConfig {
        details := offerConfig(c)
        if details != nil {
            return details
        }
	}
	details := new(config.ArtifactoryDetails)
	details.Url = c.String("url")
	details.ApiKey = c.String("apikey")
	details.User = c.String("user")
	details.Password = c.String("password")
	details.SshKeyPath = c.String("ssh-key-path")

	if includeConfig {
		confDetails := commands.GetConfig()
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
	return details
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

func createMoveFlags(c *cli.Context) (moveFlags *utils.MoveFlags) {
	moveFlags = new(utils.MoveFlags)
	moveFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	moveFlags.Recursive = cliutils.GetBoolFlagValue(c, "recursive", true)
	moveFlags.Flat = cliutils.GetBoolFlagValue(c, "flat", false)
	moveFlags.DryRun = c.Bool("dry-run")
	moveFlags.Props = c.String("props")
	return
}

func createDeleteFlags(c *cli.Context) (deleteFlags *commands.DeleteFlags) {
	deleteFlags = new(commands.DeleteFlags)
	deleteFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	deleteFlags.Recursive = cliutils.GetBoolFlagValue(c, "recursive", true)
	deleteFlags.DryRun = c.Bool("dry-run")
	deleteFlags.Props = c.String("props")
	return
}

func createDownloadFlags(c *cli.Context) (downloadFlags *commands.DownloadFlags) {
	downloadFlags = new(commands.DownloadFlags)
	downloadFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	downloadFlags.Recursive = cliutils.GetBoolFlagValue(c, "recursive", true)
	downloadFlags.Flat = cliutils.GetBoolFlagValue(c, "flat", false)
	downloadFlags.Props = c.String("props")
	downloadFlags.DryRun = c.Bool("dry-run")
	downloadFlags.MinSplitSize = getMinSplit(c)
	downloadFlags.SplitCount = getSplitCount(c)
	downloadFlags.Threads = getThreadsCount(c);
	return
}

func createUploadFlags(c *cli.Context) (uploadFlags *commands.UploadFlags) {
	uploadFlags = new(commands.UploadFlags)
	uploadFlags.ArtDetails = createArtifactoryDetailsByFlags(c, true)
	uploadFlags.Recursive = cliutils.GetBoolFlagValue(c, "recursive", true)
	uploadFlags.Flat = cliutils.GetBoolFlagValue(c, "flat", true)
	uploadFlags.Props = c.String("props")
	uploadFlags.DryRun = c.Bool("dry-run")
	uploadFlags.Threads = getThreadsCount(c);
	uploadFlags.Deb = getDebFlag(c);
	uploadFlags.UseRegExp = c.Bool("regexp")
	return
}

func createConfigFlags(c *cli.Context) (configFlag *commands.ConfigFlags) {
	configFlag = new(commands.ConfigFlags)
	configFlag.ArtDetails = createArtifactoryDetails(c, false)
	configFlag.EncPassword = cliutils.GetBoolFlagValue(c, "enc-password", true)
	configFlag.Interactive = cliutils.GetBoolFlagValue(c, "interactive", true)
	if !configFlag.Interactive && configFlag.ArtDetails.Url == "" {
		cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory when the --interactive option is set to false")
	}
	return
}
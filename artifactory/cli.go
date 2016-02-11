package artifactory

import (
	"strings"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/commands"
	"github.com/codegangsta/cli"
)

var flags = new(utils.Flags)

func GetCommands() []cli.Command {
    return []cli.Command{
        {
            Name: "config",
            Flags: getConfigFlags(),
            Aliases: []string{"c"},
            Usage: "config",
            Action: func(c *cli.Context) {
                config(c)
            },
        },
        {
            Name: "upload",
            Flags: getUploadFlags(),
            Aliases: []string{"u"},
            Usage: "upload <local path> <repo path>",
            Action: func(c *cli.Context) {
                upload(c)
            },
        },
        {
            Name: "download",
            Flags: getDownloadFlags(),
            Aliases: []string{"d"},
            Usage: "download <repo path>",
            Action: func(c *cli.Context) {
                download(c)
            },
        },
    }
}

func getFlags() []cli.Flag {
    return []cli.Flag{
        cli.StringFlag{
         Name:  "url",
         Usage: "[Mandatory] Artifactory URL",
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
            Value:  "",
            Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be uploaded to Artifactory.",
        },
        cli.StringFlag{
            Name:  "flat",
            Value:  "",
            Usage: "[Default: true] If not set to true, and the upload path ends with a slash, files are uploaded according to their file system hierarchy.",
        },
        cli.BoolFlag{
             Name:  "regexp",
             Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to upload.",
        },
        cli.StringFlag{
             Name:  "threads",
             Value:  "",
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
            Value:  "",
            Usage: "[Default: true] Set to false if you do not wish to include the download of artifacts inside sub-folders in Artifactory.",
        },
        cli.StringFlag{
            Name:  "flat",
            Value:  "",
            Usage: "[Default: false] Set to true if you do not wish to have the Artifactory repository path structure created locally for your downloaded files.",
        },
        cli.StringFlag{
            Name:  "min-split",
            Value:  "",
            Usage: "[Default: 5120] Minimum file size in KB to split into ranges when downloading. Set to -1 for no splits.",
        },
        cli.StringFlag{
            Name:  "split-count",
            Value:  "",
            Usage: "[Default: 3] Number of parts to split a file when downloading. Set to 0 for no splits.",
        },
        cli.StringFlag{
             Name:  "threads",
             Value:  "",
             Usage: "[Default: 3] Number of artifacts to download in parallel.",
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
            Name: "enc-password",
            Usage: "[Default: true] If set to false then the configured password will not be encrypted using Artifatory's encryption API.",
        },
    }
    return append(flags, getFlags()...)
}

func initFlags(c *cli.Context, cmd string) {
    if c.String("recursive") == "" {
        flags.Recursive = true
    } else {
        flags.Recursive = c.Bool("recursive")
    }
    if c.String("interactive") == "" {
        flags.Interactive = true
    } else {
        flags.Interactive = c.Bool("interactive")
    }
    if c.String("enc-password") == "" {
        flags.EncPassword = true
    } else {
        flags.EncPassword = c.Bool("enc-password")
    }

    if cmd == "config" {
        flags.ArtDetails = getArtifactoryDetails(c, false)
        if !flags.Interactive && flags.ArtDetails.Url == "" {
            cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory when the --interactive option is set to false")
        }
    } else {
        flags.ArtDetails = getArtifactoryDetails(c, true)
        if flags.ArtDetails.Url == "" {
            cliutils.Exit(cliutils.ExitCodeError, "The --url option is mandatory")
        }
    }

    strFlat := c.String("flat")
    if cmd == "upload" {
        if strFlat == "" {
            flags.Flat = true
        } else {
            flags.Flat, _ = strconv.ParseBool(strFlat)
        }
    } else {
        if strFlat == "" {
            flags.Flat = false
        } else {
             flags.Flat, _ = strconv.ParseBool(strFlat)
         }
    }

    flags.Deb = c.String("deb")
    if flags.Deb != "" && len(strings.Split(flags.Deb, "/")) != 3 {
        cliutils.Exit(cliutils.ExitCodeError, "The --deb option should be in the form of distribution/component/architecture")
    }
    flags.Props = c.String("props")
    flags.DryRun = c.Bool("dry-run")
    flags.UseRegExp = c.Bool("regexp")
    var err error
    if c.String("threads") == "" {
        flags.Threads = 3
    } else {
        flags.Threads, err = strconv.Atoi(c.String("threads"))
        if err != nil || flags.Threads < 1 {
            cliutils.Exit(cliutils.ExitCodeError, "The '--threads' option should have a numeric positive value.")
        }
    }
    if c.String("min-split") == "" {
        flags.MinSplitSize = 5120
    } else {
        flags.MinSplitSize, err = strconv.ParseInt(c.String("min-split"), 10, 64)
        if err != nil {
            cliutils.Exit(cliutils.ExitCodeError, "The '--min-split' option should have a numeric value. Try 'art download --help'.")
        }
    }
    if c.String("split-count") == "" {
        flags.SplitCount = 3
    } else {
        flags.SplitCount, err = strconv.Atoi(c.String("split-count"))
        if err != nil {
            cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option should have a numeric value. Try 'art download --help'.")
        }
        if flags.SplitCount > 15 {
            cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option value is limitted to a maximum of 15.")
        }
        if flags.SplitCount < 0 {
            cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option cannot have a negative value.")
        }
    }
}

func config(c *cli.Context) {
    if len(c.Args()) > 1 {
        cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. Try 'art config --help'.")
    } else
    if len(c.Args()) == 1 {
        if c.Args()[0] == "show" {
            commands.ShowConfig()
        } else
        if c.Args()[0] == "clear" {
            commands.ClearConfig()
        } else {
            cliutils.Exit(cliutils.ExitCodeError, "Unknown argument '" + c.Args()[0] + "'. Available arguments are 'show' and 'clear'.")
        }
    } else {
        initFlags(c, "config")
        commands.Config(flags.ArtDetails, flags.Interactive, flags.EncPassword)
    }
}

func download(c *cli.Context) {
    initFlags(c, "download")
    if len(c.Args()) != 1 {
        cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. Try 'art download --help'.")
    }
    pattern := c.Args()[0]
    commands.Download(pattern, flags)
}

func upload(c *cli.Context) {
    initFlags(c, "upload")
    size := len(c.Args())
    if size != 2 {
        cliutils.Exit(cliutils.ExitCodeError, "Wrong number of arguments. Try 'art upload --help'.")
    }
    localPath := c.Args()[0]
    targetPath := c.Args()[1]
    uploaded , failed := commands.Upload(localPath, targetPath, flags)
    if failed > 0 {
        if uploaded > 0 {
            cliutils.Exit(cliutils.ExitCodeWarning, "")
        }
        cliutils.Exit(cliutils.ExitCodeError, "")
    }
}

func getArtifactoryDetails(c *cli.Context, includeConfig bool) *utils.ArtifactoryDetails {
    details := new(utils.ArtifactoryDetails)
    details.Url = c.String("url")
    details.User = c.String("user")
    details.Password = c.String("password")
    details.SshKeyPath = c.String("ssh-key-path")

    if includeConfig {
        if details.Url == "" ||
            ((details.User == "" || details.Password == "") && details.SshKeyPath == "") {

            confDetails := commands.GetConfig()
            if details.Url == "" {
                details.Url = confDetails.Url
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
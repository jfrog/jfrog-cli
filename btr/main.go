package main

import (
    "os"
    "github.com/codegangsta/cli"
    "github.com/JFrogDev/bintray-cli-go/utils"
    "github.com/JFrogDev/bintray-cli-go/commands"
)

func main() {
    app := cli.NewApp()
    app.Name = "btr"
    app.Usage = "See https://github.com/JFrogDev/bintray-cli-go for usage instructions."
    app.Version = "0.0.1"

    app.Commands = []cli.Command{
        {
            Name: "download-ver",
            Usage: "Download version files",
            Aliases: []string{"dv"},
            Flags: getDownloadVersionFlags(),
            Action: func(c *cli.Context) {
                downloadVersion(c)
            },
        },
        {
            Name: "download-file",
            Usage: "Download file",
            Aliases: []string{"df"},
            Flags: getDownloadFileFlags(),
            Action: func(c *cli.Context) {
                downloadFile(c)
            },
        },
    }
    app.Run(os.Args)
}

func getFlags() []cli.Flag {
    return []cli.Flag{
        cli.StringFlag{
            Name:  "user",
            EnvVar: "BINTRAY_USER",
            Usage: "[Mandatory] Bintray username",
        },
        cli.StringFlag{
            Name:  "key",
            EnvVar: "BINTRAY_KEY",
            Usage: "[Mandatory] Bintray API key",
        },
        cli.StringFlag{
            Name:  "org",
            EnvVar: "BINTRAY_ORG",
            Usage: "[Optional] Bintray organization",
        },
        cli.StringFlag{
            Name: "api-url",
            EnvVar: "BINTRAY_API_URL",
            Usage: "[Default: https://api.bintray.com] Bintray API URL",
        },
        cli.StringFlag{
            Name: "download-url",
            EnvVar: "BINTRAY_DOWNLOAD_URL",
            Usage: "[Default: https://dl.bintray.com] Bintray download server URL",
        },
    }
}

func getDownloadVersionFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:5], getFlags())
    flags[5] = cli.StringFlag{
         Name:  "repo",
         Usage: "[Mandatory] Bintray repository",
    }
    flags[6] = cli.StringFlag{
         Name:  "package",
         Usage: "[Mandatory] Bintray package",
    }
    return flags
}

func getDownloadFileFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:5], getFlags())
    flags[5] = cli.StringFlag{
         Name:  "repo",
         Usage: "[Mandatory] Bintray repository",
    }
    return flags
}

func downloadVersion(c *cli.Context) {
    if len(c.Args()) != 1 {
        utils.Exit("Wrong number of arguments. Try 'btr download-ver --help'.")
    }
    version := c.Args()[0]
    flags := createDownloadVersionFlags(c)
    commands.DownloadVersion(version, flags)
}

func downloadFile(c *cli.Context) {
    if len(c.Args()) != 1 {
        utils.Exit("Wrong number of arguments. Try 'btr download-ver --help'.")
    }
    path := c.Args()[0]
    flags := createDownloadFileFlags(c)
    commands.DownloadFile(path, flags)
}

func createDownloadVersionFlags(c *cli.Context) *commands.DownloadVersionFlags {
    repo := c.String("repo")
    pkg := c.String("package")
    if repo == "" {
        utils.Exit("The --repo option is mandatory")
    }
    if pkg == "" {
        utils.Exit("The --package option is mandatory")
    }
    return &commands.DownloadVersionFlags {
        Repo: repo,
        Package: pkg,
        BintrayDetails: createBintrayDetails(c)}
}

func createDownloadFileFlags(c *cli.Context) *commands.DownloadFileFlags {
    repo := c.String("repo")
    if repo == "" {
        utils.Exit("The --repo option is mandatory")
    }
    return &commands.DownloadFileFlags {
        Repo: repo,
        BintrayDetails: createBintrayDetails(c)}
}

func createBintrayDetails(c *cli.Context) *utils.BintrayDetails {
    if c.String("user") == "" {
        utils.Exit("Please use the --user option or set the BINTRAY_USER envrionemt variable")
    }
    if c.String("key") == "" {
        utils.Exit("Please use the --key option or set the BINTRAY_KEY envrionemt variable")
    }
    apiUrl := c.String("api-url")
    if apiUrl == "" {
        apiUrl = "https://api.bintray.com/"
    }
    downloadServerUrl := c.String("download-url")
    if downloadServerUrl == "" {
        downloadServerUrl = "https://dl.bintray.com/"
    }

    apiUrl = utils.AddTrailingSlashIfNeeded(apiUrl)
    downloadServerUrl = utils.AddTrailingSlashIfNeeded(downloadServerUrl)
    org := c.String("org")
    if org == "" {
        org = c.String("user")
    }
    return &utils.BintrayDetails {
        ApiUrl: apiUrl,
        DownloadServerUrl: downloadServerUrl,
        Org: org,
        User: c.String("user"),
        Key: c.String("key") }
}
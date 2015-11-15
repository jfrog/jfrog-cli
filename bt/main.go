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
            Aliases: []string{"dlv"},
            Flags: getDownloadVersionFlags(),
            Action: func(c *cli.Context) {
                downloadVersion(c)
            },
        },
        {
            Name: "download-file",
            Usage: "Download file",
            Aliases: []string{"dlf"},
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
            Usage: "[Optional] Bintray username. If not set, the subject sent as part of the command argument is used for authentication.",
        },
        cli.StringFlag{
            Name:  "key",
            EnvVar: "BINTRAY_KEY",
            Usage: "[Mandatory] Bintray API key",
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
    return getFlags()
}

func getDownloadFileFlags() []cli.Flag {
    return getFlags()
}

func downloadVersion(c *cli.Context) {
    if len(c.Args()) != 1 {
        utils.Exit("Wrong number of arguments. Try 'btr download-ver --help'.")
    }
    versionDetails := utils.CreateVersionDetails(c.Args()[0])
    bintrayDetails := createBintrayDetails(c)
    if bintrayDetails.User == "" {
        bintrayDetails.User = versionDetails.Subject
    }
    commands.DownloadVersion(versionDetails, bintrayDetails)
}

func downloadFile(c *cli.Context) {
    if len(c.Args()) != 1 {
        utils.Exit("Wrong number of arguments. Try 'btr download-ver --help'.")
    }
    versionDetails, path := utils.CreateVersionDetailsAndPath(c.Args()[0])
    bintrayDetails := createBintrayDetails(c)
    if bintrayDetails.User == "" {
        bintrayDetails.User = versionDetails.Subject
    }
    commands.DownloadFile(versionDetails, path, bintrayDetails)
}

func createBintrayDetails(c *cli.Context) *utils.BintrayDetails {
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
    return &utils.BintrayDetails {
        ApiUrl: apiUrl,
        DownloadServerUrl: downloadServerUrl,
        User: c.String("user"),
        Key: c.String("key") }
}
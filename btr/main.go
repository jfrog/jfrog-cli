package main

import (
    "os"
    "github.com/codegangsta/cli"
    "github.com/JFrogDev/bintray-cli-go/utils"
    "github.com/JFrogDev/bintray-cli-go/commands"
)

func main() {
    app := cli.NewApp()
    app.Name = "btray"
    app.Usage = "task list on the command line"

    app.Commands = []cli.Command{
        {
            Name: "download-ver",
            Usage: "download-ver",
            Aliases: []string{"dv"},
            Flags: getDownloadVersionFlags(),
            Action: func(c *cli.Context) {
                downloadVersion(c)
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
            Usage: "Bintray username",
        },
        cli.StringFlag{
            Name:  "key",
            EnvVar: "BINTRAY_KEY",
            Usage: "Bintray API key",
        },
        cli.StringFlag{
            Name:  "org",
            EnvVar: "BINTRAY_ORG",
            Usage: "Bintray organization",
        },
        cli.StringFlag{
            Name: "url",
            EnvVar: "BINTRAY_API_URL",
            Usage: "Bintray API URL",
        },
    }
}

func getDownloadVersionFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    flags[4] = cli.StringFlag{
         Name:  "repo",
         Usage: "Bintray repository",
    }
    flags[5] = cli.StringFlag{
         Name:  "package",
         Usage: "Bintray package",
    }
    flags[6] = cli.StringFlag{
         Name:  "version",
         Usage: "Package version",
    }
    return flags
}

func downloadVersion(c *cli.Context) {
    flags := createDownloadVersionFlags(c)
    commands.DownloadVersion(flags)
}

func createDownloadVersionFlags(c *cli.Context) *commands.DownloadVersionFlags {
    repo := c.String("repo")
    pkg := c.String("package")
    version := c.String("version")
    if repo == "" {
        utils.Exit("The --repo option is mandatory")
    }
    if pkg == "" {
        utils.Exit("The --package option is mandatory")
    }
    if version == "" {
        utils.Exit("The --version option is mandatory")
    }
    return &commands.DownloadVersionFlags {
        Repo: repo,
        Package: pkg,
        Version: version,
        BintrayDetails: createBintrayDetails(c)}
}

func createBintrayDetails(c *cli.Context) *utils.BintrayDetails {
    if c.String("user") == "" {
        utils.Exit("Please use the --user option or set the BINTRAY_USER envrionemt variable")
    }
    if c.String("key") == "" {
        utils.Exit("Please use the --key option or set the BINTRAY_KEY envrionemt variable")
    }
    url := c.String("url")
    if url == "" {
        url = "https://api.bintray.com"
    }
    org := c.String("org")
    if org == "" {
        org = c.String("user")
    }
    return &utils.BintrayDetails {
        Url: url,
        Org: org,
        User: c.String("user"),
        Key: c.String("key") }
}
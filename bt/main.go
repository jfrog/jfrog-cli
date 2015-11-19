package main

import (
    "os"
    "strconv"
    "github.com/codegangsta/cli"
    "github.com/JFrogDev/bintray-cli-go/utils"
    "github.com/JFrogDev/bintray-cli-go/commands"
)

func main() {
    app := cli.NewApp()
    app.Name = "bt"
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
        {
            Name: "entitlements",
            Usage: "Entitlements",
            Aliases: []string{"ent"},
            Flags: getEntitlementsFlags(),
            Action: func(c *cli.Context) {
                entitlements(c)
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

func getEntitlementsFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    flags[4] = cli.StringFlag{
         Name:  "org",
         Usage: "[Optional] Bintray organization",
    }
    flags[5] = cli.StringFlag{
         Name:  "key-id",
         Usage: "[Optional] Download Key ID (required for 'bt entitlements key show/create/update/delete'",
    }
    flags[6] = cli.StringFlag{
         Name:  "key-expiry",
         Usage: "[Optional] Download Key expiry (required for 'bt entitlements key show/create/update/delete'",
    }
    flags[7] = cli.StringFlag{
         Name:  "key-ex-check-url",
         Usage: "[Optional] Used for Download Key creation and update. You can optionally provide an existence check directive, in the form of a callback URL, to verify whether the source identity of the Download Key still exists.",
    }
    flags[8] = cli.StringFlag{
         Name:  "key-ex-check-cache",
         Usage: "[Optional] Used for Download Key creation and update. You can optionally provide the period in seconds for the callback URK results cache.",
    }
    flags[8] = cli.StringFlag{
         Name:  "key-ex-check-cache",
         Usage: "[Optional] Used for Download Key creation and update. You can optionally provide the period in seconds for the callback URK results cache.",
    }
    flags[9] = cli.StringFlag{
         Name:  "key-white-cidrs",
         Usage: "[Optional] Used for Download Key creation and update. Specifying white CIDRs in the form of \"127.0.0.1/22\", \"193.5.0.1/92\" will allow access only for those IPs that exist in that address range.",
    }
    flags[10] = cli.StringFlag{
         Name:  "key-black-cidrs",
         Usage: "[Optional] Used for Download Key creation and update. Specifying black CIDRs in the foem of \"127.0.0.1/22\",\"193.5.0.1/92\" will block access for all IPs that exist in the specified range.",
    }
    return flags
}

func downloadVersion(c *cli.Context) {
    if len(c.Args()) != 1 {
        utils.Exit("Wrong number of arguments. Try 'bt download-ver --help'.")
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
        utils.Exit("Wrong number of arguments. Try 'bt download-ver --help'.")
    }
    versionDetails, path := utils.CreateVersionDetailsAndPath(c.Args()[0])
    bintrayDetails := createBintrayDetails(c)
    if bintrayDetails.User == "" {
        bintrayDetails.User = versionDetails.Subject
    }
    commands.DownloadFile(versionDetails, path, bintrayDetails)
}

func entitlements(c *cli.Context) {
    org := c.String("org")
    argsSize := len(c.Args())
    if argsSize == 0 {
        utils.Exit("Wrong number of arguments. Try 'bt entitlements --help'.")
    }
    bintrayDetails := createBintrayDetails(c)
    if c.Args()[0] == "keys" {
        commands.ShowDownloadKeys(bintrayDetails, org)
        return
    }
    if argsSize != 2 {
        utils.Exit("Wrong number of arguments. Try 'bt entitlements --help'.")
    }
    if c.Args()[0] == "key" {
        if c.Args()[1] == "show" {
            commands.ShowDownloadKey(createDownloadKeyForShowAndDelete(c), org)
        } else
        if c.Args()[1] == "create" {
            commands.CreateDownloadKey(createDownloadKeyForCreateAndUpdate(c), org)
        } else
        if c.Args()[1] == "update" {
            commands.UpdateDownloadKey(createDownloadKeyForCreateAndUpdate(c), org)
        } else
        if c.Args()[1] == "delete" {
            commands.DeleteDownloadKey(createDownloadKeyForShowAndDelete(c), org)
        } else {
            utils.Exit("Expecting show, create, update or delete after the key argument. Got " + c.Args()[1])
        }
    }
}

func createDownloadKeyForShowAndDelete(c *cli.Context) *commands.DownloadKeyFlags {
    if c.String("key-id") == "" {
        utils.Exit("Please add the --key-id option")
    }
    return &commands.DownloadKeyFlags {
        BintrayDetails: createBintrayDetails(c),
        Id: c.String("key-id") }
}

func createDownloadKeyForCreateAndUpdate(c *cli.Context) *commands.DownloadKeyFlags {
    if c.String("key-id") == "" {
        utils.Exit("Please add the --key-id option")
    }
    var cachePeriod int
    if c.String("key-ex-check-cache") != "" {
        var err error
        cachePeriod, err = strconv.Atoi(c.String("key-ex-check-cache"))
        if err != nil {
            utils.Exit("The --key-ex-check-cache option should have a numeric value.")
        }
    }
    return &commands.DownloadKeyFlags {
        BintrayDetails: createBintrayDetails(c),
        Id: c.String("key-id"),
        Expiry: c.String("key-expiry"),
        ExistenceCheckUrl: c.String("key-ex-check-url"),
        ExistenceCheckCache: cachePeriod,
        WhiteCidrs: c.String("key-white-cidrs"),
        BlackCidrs: c.String("key-black-cidrs") }
}

func createBintrayDetails(c *cli.Context) *utils.BintrayDetails {
    if c.String("key") == "" {
        utils.Exit("Please add the --key option or set the BINTRAY_KEY envrionemt variable")
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
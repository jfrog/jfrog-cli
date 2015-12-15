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
        {
            Name: "entitlement-keys",
            Usage: "Entitlement keys",
            Aliases: []string{"ent-keys"},
            Flags: getEntitlementKeysFlags(),
            Action: func(c *cli.Context) {
                entitlementKeys(c)
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
        nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    flags[4] = cli.StringFlag{
         Name:  "id",
         Usage: "[Optional] Entitlement ID. Used for Entitlements update.",
    }
    flags[5] = cli.StringFlag{
         Name:  "access",
         Usage: "[Optional] Entitlement access. Used for Entitlements creation and update.",
    }
    flags[6] = cli.StringFlag{
         Name:  "keys",
         Usage: "[Optional] Used for Entitlements creation and update. List of Download Keys in the form of \"key1\",\"key2\"...",
    }
    flags[7] = cli.StringFlag{
         Name:  "path",
         Usage: "[Optional] Entitlement path. Used for Entitlements creating and update.",
    }
    return flags
}

func getEntitlementKeysFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    flags[4] = cli.StringFlag{
         Name:  "org",
         Usage: "[Optional] Bintray organization",
    }
    flags[5] = cli.StringFlag{
         Name:  "id",
         Usage: "[Optional] Download Key ID (required for 'bt ent-keys show/create/update/delete'",
    }
    flags[6] = cli.StringFlag{
         Name:  "expiry",
         Usage: "[Optional] Download Key expiry (required for 'bt ent-keys show/create/update/delete'",
    }
    flags[7] = cli.StringFlag{
         Name:  "ex-check-url",
         Usage: "[Optional] Used for Download Key creation and update. You can optionally provide an existence check directive, in the form of a callback URL, to verify whether the source identity of the Download Key still exists.",
    }
    flags[8] = cli.StringFlag{
         Name:  "ex-check-cache",
         Usage: "[Optional] Used for Download Key creation and update. You can optionally provide the period in seconds for the callback URK results cache.",
    }
    flags[9] = cli.StringFlag{
         Name:  "white-cidrs",
         Usage: "[Optional] Used for Download Key creation and update. Specifying white CIDRs in the form of \"127.0.0.1/22\",\"193.5.0.1/92\" will allow access only for those IPs that exist in that address range.",
    }
    flags[10] = cli.StringFlag{
         Name:  "black-cidrs",
         Usage: "[Optional] Used for Download Key creation and update. Specifying black CIDRs in the foem of \"127.0.0.1/22\",\"193.5.0.1/92\" will block access for all IPs that exist in the specified range.",
    }
    return flags
}

func downloadVersion(c *cli.Context) {
    if len(c.Args()) != 1 {
        utils.Exit("Wrong number of arguments. Try 'bt download-ver --help'.")
    }
    versionDetails := commands.CreateVersionDetailsForDownloadVersion(c.Args()[0])
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

func entitlementKeys(c *cli.Context) {
    org := c.String("org")
    argsSize := len(c.Args())
    if argsSize == 0 {
        bintrayDetails := createBintrayDetails(c)
        commands.ShowDownloadKeys(bintrayDetails, org)
        return
    }
    if argsSize != 2 {
        utils.Exit("Wrong number of arguments. Try 'bt ent-keys --help'.")
    }
    keyId := c.Args()[1]
    if c.Args()[0] == "show" {
        commands.ShowDownloadKey(createDownloadKeyForShowAndDelete(keyId, c), org)
    } else
    if c.Args()[0] == "create" {
        commands.CreateDownloadKey(createDownloadKeyForCreateAndUpdate(keyId, c), org)
    } else
    if c.Args()[0] == "update" {
        commands.UpdateDownloadKey(createDownloadKeyForCreateAndUpdate(keyId, c), org)
    } else
    if c.Args()[0] == "delete" {
        commands.DeleteDownloadKey(createDownloadKeyForShowAndDelete(keyId, c), org)
    } else {
        utils.Exit("Expecting show, create, update or delete after the key argument. Got " + c.Args()[0])
    }
}

func entitlements(c *cli.Context) {
    argsSize := len(c.Args())
    if argsSize == 0 {
        utils.Exit("Wrong number of arguments. Try 'bt ent --help'.")
    }
    if argsSize == 1 {
        bintrayDetails := createBintrayDetails(c)
        details := commands.CreateVersionDetailsForEntitlements(c.Args()[0])
        commands.ShowEntitlements(bintrayDetails, details)
        return
    }
    if argsSize != 2 {
        utils.Exit("Wrong number of arguments. Try 'bt ent --help'.")
    }
    details := commands.CreateVersionDetailsForEntitlements(c.Args()[1])
    if c.Args()[0] == "show" {
        println("show")
    } else
    if c.Args()[0] == "create" {
        commands.CreateEntitlement(createEntitlementForCreateAndUpdate(c), details)
    } else
    if c.Args()[0] == "update" {
        println("update")
    } else
    if c.Args()[0] == "delete" {
        println("delete")
    } else {
        utils.Exit("Expecting show, create, update or delete before " + c.Args()[1] + ". Got " + c.Args()[0])
    }
}

func createEntitlementForCreateAndUpdate(c *cli.Context) *commands.EntitlementFlags {
    if c.String("keys") == "" {
        utils.Exit("Please add the --keys option")
    }
    if c.String("access") == "" {
        utils.Exit("Please add the --access option")
    }
    return &commands.EntitlementFlags {
        BintrayDetails: createBintrayDetails(c),
        Id: c.String("id"),
        Path: c.String("path"),
        Access: c.String("access"),
        Keys: c.String("keys") }
}

func createDownloadKeyForShowAndDelete(keyId string, c *cli.Context) *commands.DownloadKeyFlags {
    return &commands.DownloadKeyFlags {
        BintrayDetails: createBintrayDetails(c),
        Id: keyId }
}

func createDownloadKeyForCreateAndUpdate(keyId string, c *cli.Context) *commands.DownloadKeyFlags {
    var cachePeriod int
    if c.String("ex-check-cache") != "" {
        var err error
        cachePeriod, err = strconv.Atoi(c.String("ex-check-cache"))
        if err != nil {
            utils.Exit("The --ex-check-cache option should have a numeric value.")
        }
    }
    return &commands.DownloadKeyFlags {
        BintrayDetails: createBintrayDetails(c),
        Id: keyId,
        Expiry: c.String("expiry"),
        ExistenceCheckUrl: c.String("ex-check-url"),
        ExistenceCheckCache: cachePeriod,
        WhiteCidrs: c.String("white-cidrs"),
        BlackCidrs: c.String("black-cidrs") }
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
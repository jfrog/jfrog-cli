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
            Name: "upload",
            Usage: "Upload files",
            Aliases: []string{"u"},
            Flags: getUploadFlags(),
            Action: func(c *cli.Context) {
                upload(c)
            },
        },
        {
            Name: "download-file",
            Usage: "Download file",
            Aliases: []string{"dlf"},
            Flags: getFlags(),
            Action: func(c *cli.Context) {
                downloadFile(c)
            },
        },
        {
            Name: "download-ver",
            Usage: "Download version files",
            Aliases: []string{"dlv"},
            Flags: getFlags(),
            Action: func(c *cli.Context) {
                downloadVersion(c)
            },
        },
        {
            Name: "package-create",
            Usage: "Create a package",
            Aliases: []string{"pc"},
            Flags: getCreatePackageFlags(),
            Action: func(c *cli.Context) {
                createPackage(c)
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

func getCreatePackageFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    flags[4] = cli.StringFlag{
        Name:  "desc",
        Value:  "",
        Usage: "[Optional] Package description.",
    }
    flags[5] = cli.StringFlag{
        Name:  "labels",
        Value:  "",
        Usage: "[Optional] Package lables in the form of \"lable11\",\"lable2\"...",
    }
    flags[6] = cli.StringFlag{
         Name:  "licenses",
         Value:  "",
         Usage: "[Mandatory] Package licenses in the form of \"Apache-2.0\",\"GPL-3.0\"...",
    }
    flags[7] = cli.StringFlag{
         Name:  "cust-licenses",
         Value:  "",
         Usage: "[Optional] Package custom licenses in the form of \"my-license-1\",\"my-license-2\"...",
    }
    flags[8] = cli.StringFlag{
         Name:  "vcs-url",
         Value:  "",
         Usage: "[Mandatory] Package VCS URL.",
    }
    flags[9] = cli.StringFlag{
         Name:  "website-url",
         Value:  "",
         Usage: "[Optional] Package web site URL.",
    }
    flags[10] = cli.StringFlag{
         Name:  "i-tracker-url",
         Value:  "",
         Usage: "[Optional] Package Issues Tracker URL.",
    }
    flags[11] = cli.StringFlag{
         Name:  "github-repo",
         Value:  "",
         Usage: "[Optional] Package Github repository.",
    }
    flags[12] = cli.StringFlag{
         Name:  "github-rel-notes",
         Value:  "",
         Usage: "[Optional] Github release notes file.",
    }
    flags[13] = cli.BoolFlag{
         Name:  "pub-dn",
         Usage: "[Default: false] Public download numbers.",
    }
    flags[14] = cli.BoolFlag{
         Name:  "pub-stats",
         Usage: "[Default: false] Public statistics",
    }

    return flags
}

func getUploadFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    flags[4] = cli.StringFlag{
        Name:  "recursive",
        Value:  "",
        Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be uploaded to Artifactory.",
    }
    flags[5] = cli.StringFlag{
        Name:  "flat",
        Value:  "",
        Usage: "[Default: true] If not set to true, and the upload path ends with a slash, files are uploaded according to their file system hierarchy.",
    }
    flags[6] = cli.BoolFlag{
         Name:  "regexp",
         Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to upload.",
    }
    flags[7] = cli.BoolFlag{
         Name:  "dry-run",
         Usage: "[Default: false] Set to true to disable communication with Artifactory.",
    }
    return flags
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

func createPackage(c *cli.Context) {
    if len(c.Args()) != 1 {
        utils.Exit("Wrong number of arguments. Try 'bt create-package --help'.")
    }
    packageDetails := utils.CreatePackageDetails(c.Args()[0])
    packageFlags := createPackageFlags(c)
    if packageFlags.BintrayDetails.User == "" {
        packageFlags.BintrayDetails.User = packageDetails.Subject
    }
    commands.CreatePackage(packageDetails, packageFlags)
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

func upload(c *cli.Context) {
    if len(c.Args()) != 2 {
        utils.Exit("Wrong number of arguments. Try 'bt upload --help'.")
    }
    localPath := c.Args()[0]
    versionDetails, uploadPath := utils.CreateVersionDetailsAndPath(c.Args()[1])
    uploadFlags := createUploadFlags(c)
    if uploadFlags.BintrayDetails.User == "" {
        uploadFlags.BintrayDetails.User = versionDetails.Subject
    }
    commands.Upload(versionDetails, localPath, uploadPath, uploadFlags)
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
        commands.ShowDownloadKey(createDownloadKeyFlagsForShowAndDelete(keyId, c), org)
    } else
    if c.Args()[0] == "create" {
        commands.CreateDownloadKey(createDownloadKeyFlagsForCreateAndUpdate(keyId, c), org)
    } else
    if c.Args()[0] == "update" {
        commands.UpdateDownloadKey(createDownloadKeyFlagsForCreateAndUpdate(keyId, c), org)
    } else
    if c.Args()[0] == "delete" {
        commands.DeleteDownloadKey(createDownloadKeyFlagsForShowAndDelete(keyId, c), org)
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
        commands.ShowEntitlement(createEntitlementFlagsForShowAndDelete(c), details)
    } else
    if c.Args()[0] == "create" {
        commands.CreateEntitlement(createEntitlementFlagsForCreate(c), details)
    } else
    if c.Args()[0] == "update" {
        commands.UpdateEntitlement(createEntitlementFlagsForUpdate(c), details)
    } else
    if c.Args()[0] == "delete" {
        commands.DeleteEntitlement(createEntitlementFlagsForShowAndDelete(c), details)
    } else {
        utils.Exit("Expecting show, create, update or delete before " + c.Args()[1] + ". Got " + c.Args()[0])
    }
}

func createPackageFlags(c *cli.Context) *commands.PackageFlags {
    return &commands.PackageFlags {
        BintrayDetails: createBintrayDetails(c),
        Desc: c.String("desc"),
        Labels: c.String("labels"),
        Licenses: c.String("licenses"),
        CustomLicenses: c.String("cust-licenses"),
        VcsUrl: c.String("vcs-url"),
        WebsiteUrl: c.String("website-url"),
        IssueTrackerUrl: c.String("i-tracker-url"),
        GithubRepo: c.String("github-repo"),
        GithubReleaseNotesFile: c.String("github-rel-notes"),
        PublicDownloadNumbers: c.Bool("pub-dn"),
        PublicStats: c.Bool("pub-stats") }
}

func createUploadFlags(c *cli.Context) *commands.UploadFlags {
    var recursive bool
    var flat bool

    if c.String("recursive") == "" {
        recursive = true
    } else {
        recursive = c.Bool("recursive")
    }
    if c.String("flat") == "" {
        flat = true
    } else {
        flat = c.Bool("flat")
    }

    return &commands.UploadFlags {
        BintrayDetails: createBintrayDetails(c),
        Recursive: recursive,
        Flat: flat,
        UseRegExp: c.Bool("regexp"),
        DryRun: c.Bool("dry-run") }
}

func createEntitlementFlagsForShowAndDelete(c *cli.Context) *commands.EntitlementFlags {
    if c.String("id") == "" {
        utils.Exit("Please add the --id option")
    }
    return &commands.EntitlementFlags {
        BintrayDetails: createBintrayDetails(c),
        Id: c.String("id") }
}

func createEntitlementFlagsForCreate(c *cli.Context) *commands.EntitlementFlags {
    if c.String("access") == "" {
        utils.Exit("Please add the --access option")
    }
    return &commands.EntitlementFlags {
        BintrayDetails: createBintrayDetails(c),
        Path: c.String("path"),
        Access: c.String("access"),
        Keys: c.String("keys") }
}

func createEntitlementFlagsForUpdate(c *cli.Context) *commands.EntitlementFlags {
    if c.String("id") == "" {
        utils.Exit("Please add the --id option")
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

func createDownloadKeyFlagsForShowAndDelete(keyId string, c *cli.Context) *commands.DownloadKeyFlags {
    return &commands.DownloadKeyFlags {
        BintrayDetails: createBintrayDetails(c),
        Id: keyId }
}

func createDownloadKeyFlagsForCreateAndUpdate(keyId string, c *cli.Context) *commands.DownloadKeyFlags {
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
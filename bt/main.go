package main

import (
    "os"
    "fmt"
    "strings"
    "strconv"
    "github.com/codegangsta/cli"
    "github.com/JFrogDev/bintray-cli-go/bintray/commands"
    "github.com/JFrogDev/bintray-cli-go/cliutils"
    "github.com/JFrogDev/bintray-cli-go/bintray/utils"
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
            Usage: "Download Version files",
            Aliases: []string{"dlv"},
            Flags: getFlags(),
            Action: func(c *cli.Context) {
                downloadVersion(c)
            },
        },
        {
            Name: "package-show",
            Usage: "Show Package details",
            Aliases: []string{"ps"},
            Flags: getFlags(),
            Action: func(c *cli.Context) {
                showPackage(c)
            },
        },
        {
            Name: "package-create",
            Usage: "Create Package",
            Aliases: []string{"pc"},
            Flags: getCreateAndUpdatePackageFlags(),
            Action: func(c *cli.Context) {
                createPackage(c)
            },
        },
        {
            Name: "package-update",
            Usage: "Update Package",
            Aliases: []string{"pu"},
            Flags: getCreateAndUpdatePackageFlags(),
            Action: func(c *cli.Context) {
                updatePackage(c)
            },
        },
        {
            Name: "package-delete",
            Usage: "Delete Package",
            Aliases: []string{"pd"},
            Flags: getDeletePackageAndVersionFlags(),
            Action: func(c *cli.Context) {
                deletePackage(c)
            },
        },
        {
            Name: "version-show",
            Usage: "Show Version",
            Aliases: []string{"vs"},
            Flags: getFlags(),
            Action: func(c *cli.Context) {
                showVersion(c)
            },
        },
        {
            Name: "version-create",
            Usage: "Create Version",
            Aliases: []string{"vc"},
            Flags: getCreateAndUpdateVersionFlags(),
            Action: func(c *cli.Context) {
                createVersion(c)
            },
        },
        {
            Name: "version-update",
            Usage: "Update Version",
            Aliases: []string{"vu"},
            Flags: getCreateAndUpdateVersionFlags(),
            Action: func(c *cli.Context) {
                updateVersion(c)
            },
        },
        {
            Name: "version-delete",
            Usage: "Delete Version",
            Aliases: []string{"vd"},
            Flags: getDeletePackageAndVersionFlags(),
            Action: func(c *cli.Context) {
                deleteVersion(c)
            },
        },
        {
            Name: "version-piblish",
            Usage: "Publish Version",
            Aliases: []string{"vp"},
            Flags: getFlags(),
            Action: func(c *cli.Context) {
                publishVersion(c)
            },
        },
        {
            Name: "entitlements",
            Usage: "Manage Entitlements",
            Aliases: []string{"ent"},
            Flags: getEntitlementsFlags(),
            Action: func(c *cli.Context) {
                entitlements(c)
            },
        },
        {
            Name: "entitlement-keys",
            Usage: "Manage Entitlement Keys",
            Aliases: []string{"ent-keys"},
            Flags: getEntitlementKeysFlags(),
            Action: func(c *cli.Context) {
                entitlementKeys(c)
            },
        },
        {
            Name: "url-sign",
            Usage: "Create Signed Download URL",
            Aliases: []string{"us"},
            Flags: getUrlSigningFlags(),
            Action: func(c *cli.Context) {
                signUrl(c)
            },
        },
        {
            Name: "gpg-sign-file",
            Usage: "GPF Sign file",
            Aliases: []string{"gsf"},
            Flags: getGpgSigningFlags(),
            Action: func(c *cli.Context) {
                gpgSignFile(c)
            },
        },
        {
            Name: "gpg-sign-ver",
            Usage: "GPF Sign Version",
            Aliases: []string{"gsv"},
            Flags: getGpgSigningFlags(),
            Action: func(c *cli.Context) {
                gpgSignVersion(c)
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

func getPackageFlags(prefix string) []cli.Flag {
    return []cli.Flag{
        cli.StringFlag{
            Name:  prefix + "desc",
            Value:  "",
            Usage: "[Optional] Package description.",
        },
        cli.StringFlag{
            Name:  prefix + "labels",
            Value:  "",
            Usage: "[Optional] Package lables in the form of \"lable11\",\"lable2\"...",
        },
        cli.StringFlag{
             Name:  prefix + "licenses",
             Value:  "",
             Usage: "[Mandatory] Package licenses in the form of \"Apache-2.0\",\"GPL-3.0\"...",
        },
        cli.StringFlag{
             Name:  prefix + "cust-licenses",
             Value:  "",
             Usage: "[Optional] Package custom licenses in the form of \"my-license-1\",\"my-license-2\"...",
        },
        cli.StringFlag{
             Name:  prefix + "vcs-url",
             Value:  "",
             Usage: "[Mandatory] Package VCS URL.",
        },
        cli.StringFlag{
             Name:  prefix + "website-url",
             Value:  "",
             Usage: "[Optional] Package web site URL.",
        },
        cli.StringFlag{
             Name:  prefix + "i-tracker-url",
             Value:  "",
             Usage: "[Optional] Package Issues Tracker URL.",
        },
        cli.StringFlag{
             Name:  prefix + "github-repo",
             Value:  "",
             Usage: "[Optional] Package Github repository.",
        },
        cli.StringFlag{
             Name:  prefix + "github-rel-notes",
             Value:  "",
             Usage: "[Optional] Github release notes file.",
        },
        cli.StringFlag{
             Name:  prefix + "pub-dn",
             Value:  "",
             Usage: "[Default: false] Public download numbers.",
        },
        cli.StringFlag{
             Name:  prefix + "pub-stats",
             Value:  "",
             Usage: "[Default: false] Public statistics",
        },
    }
}

func getVersionFlags(prefix string) []cli.Flag {
   return []cli.Flag{
       cli.StringFlag{
           Name:  prefix + "desc",
           Value:  "",
           Usage: "[Optional] Version description.",
       },
       cli.StringFlag{
           Name:  prefix + "released",
           Value:  "",
           Usage: "[Optional] Release date in ISO8601 format (yyyy-MM-dd'T'HH:mm:ss.SSSZ)",
       },
       cli.StringFlag{
            Name:  prefix + "github-rel-notes",
            Value:  "",
            Usage: "[Optional] Github release notes file.",
       },
       cli.StringFlag{
            Name:  prefix + "github-tag-rel-notes",
            Value:  "",
            Usage: "[Default: false] Set to true if you wish to use a Github tag release notes.",
       },
       cli.StringFlag{
            Name:  prefix + "vcs-tag",
            Value:  "",
            Usage: "[Optional] VCS tag.",
       },
   }
}

func getCreateAndUpdatePackageFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    copy(flags[4:15], getPackageFlags(""))
    return flags
}

func getCreateAndUpdateVersionFlags() []cli.Flag {
    flags := []cli.Flag{
       nil,nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    copy(flags[4:9], getVersionFlags(""))
    return flags
}

func getDeletePackageAndVersionFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    flags[4] = cli.StringFlag{
        Name:  "q",
        Value:  "",
        Usage: "[Default: false] Set to true to skip the delete confirmation message.",
    }
    return flags
}

func getUploadFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,
        nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,
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
    copy(flags[8:19], getPackageFlags("pkg-"))
    copy(flags[19:24], getVersionFlags("ver-"))
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

func getUrlSigningFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:4], getFlags())
    flags[4] = cli.StringFlag{
         Name:  "expiry",
         Usage: "[Optional] An expiry date for the URL, in Unix epoch time in milliseconds, after which the URL will be invalid. By default, expiry date will be 24 hours.",
    }
    flags[5] = cli.StringFlag{
         Name:  "valid-for",
         Usage: "[Optional] The number of seconds since generation before the URL expires. Mutually exclusive with the --expiry option.",
    }
    flags[6] = cli.StringFlag{
         Name:  "callback-id",
         Usage: "[Optional] An applicative identifier for the request. This identifier appears in download logs and is used in email and download webhook notifications.",
    }
    flags[7] = cli.StringFlag{
         Name:  "callback-email",
         Usage: "[Optional] An email address to send mail to when a user has used the download URL. This requiers a callback_id. The callback-id will be included in the mail message.",
    }
    flags[8] = cli.StringFlag{
         Name:  "callback-url",
         Usage: "[Optional] A webhook URL to call when a user has used the download URL.",
    }
    flags[9] = cli.StringFlag{
         Name:  "callback-method",
         Usage: "[Optional] HTTP method to use for making the callback. Will use POST by default. Supported methods are: GET, POST, PUT and HEAD.",
    }
    return flags
}

func getGpgSigningFlags() []cli.Flag {
   flags := []cli.Flag{
       nil,nil,nil,nil,nil,
   }
   copy(flags[0:4], getFlags())
   flags[4] = cli.StringFlag{
        Name:  "passphrase",
        Usage: "[Optional] GPG passphrase.",
   }
   return flags
}

func showPackage(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt package-show --help'.")
    }
    packageDetails := utils.CreatePackageDetails(c.Args()[0])
    bintrayDetails := createBintrayDetails(c)
    commands.ShowPackage(packageDetails, bintrayDetails)
}

func showVersion(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt version-show --help'.")
    }
    versionDetails := utils.CreateVersionDetails(c.Args()[0])
    bintrayDetails := createBintrayDetails(c)
    commands.ShowVersion(versionDetails, bintrayDetails)
}

func createPackage(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt package-create --help'.")
    }
    packageDetails := utils.CreatePackageDetails(c.Args()[0])
    packageFlags := createPackageFlags(c, "")
    commands.CreatePackage(packageDetails, packageFlags)
}

func createVersion(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt version-create --help'.")
    }
    versionDetails := utils.CreateVersionDetails(c.Args()[0])
    versionFlags := createVersionFlags(c, "")
    commands.CreateVersion(versionDetails, versionFlags)
}

func updateVersion(c *cli.Context) {
  if len(c.Args()) != 1 {
      cliutils.Exit("Wrong number of arguments. Try 'bt version-update --help'.")
  }
  versionDetails := utils.CreateVersionDetails(c.Args()[0])
  versionFlags := createVersionFlags(c, "")
  commands.UpdateVersion(versionDetails, versionFlags)
}

func updatePackage(c *cli.Context) {
  if len(c.Args()) != 1 {
      cliutils.Exit("Wrong number of arguments. Try 'bt package-update --help'.")
  }
  packageDetails := utils.CreatePackageDetails(c.Args()[0])
  packageFlags := createPackageFlags(c, "")
  commands.UpdatePackage(packageDetails, packageFlags)
}

func deletePackage(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt package-delete --help'.")
    }
    packageDetails := utils.CreatePackageDetails(c.Args()[0])
    bintrayDetails := createBintrayDetails(c)

    if !c.Bool("q") {
        var confirm string
        fmt.Print("Delete package " + packageDetails.Package + "? (y/n): ")
        fmt.Scanln(&confirm)
        if !cliutils.ConfirmAnswer(confirm) {
            return
        }
    }
    commands.DeletePackage(packageDetails, bintrayDetails)
}

func deleteVersion(c *cli.Context) {
  if len(c.Args()) != 1 {
      cliutils.Exit("Wrong number of arguments. Try 'bt version-delete --help'.")
  }
  versionDetails := utils.CreateVersionDetails(c.Args()[0])
  bintrayDetails := createBintrayDetails(c)

  if !c.Bool("q") {
      var confirm string
      fmt.Print("Delete version " + versionDetails.Package + "? (y/n): ")
      fmt.Scanln(&confirm)
      if !cliutils.ConfirmAnswer(confirm) {
          return
      }
  }
  commands.DeleteVersion(versionDetails, bintrayDetails)
}

func publishVersion(c *cli.Context) {
     if len(c.Args()) != 1 {
         cliutils.Exit("Wrong number of arguments. Try 'bt version-publish --help'.")
     }
     versionDetails := utils.CreateVersionDetails(c.Args()[0])
     bintrayDetails := createBintrayDetails(c)
     commands.PublishVersion(versionDetails, bintrayDetails)
}

func downloadVersion(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt download-ver --help'.")
    }
    versionDetails := commands.CreateVersionDetailsForDownloadVersion(c.Args()[0])
    bintrayDetails := createBintrayDetails(c)
    commands.DownloadVersion(versionDetails, bintrayDetails)
}

func upload(c *cli.Context) {
    if len(c.Args()) != 2 {
        cliutils.Exit("Wrong number of arguments. Try 'bt upload --help'.")
    }
    localPath := c.Args()[0]
    versionDetails, uploadPath := utils.CreateVersionDetailsAndPath(c.Args()[1])
    uploadFlags := createUploadFlags(c)
    packageFlags := createPackageFlags(c, "pkg-")
    versionFlags := createVersionFlags(c, "ver-")
    commands.Upload(versionDetails, localPath, uploadPath, uploadFlags, packageFlags, versionFlags)
}

func downloadFile(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt download-file --help'.")
    }
    versionDetails, path := utils.CreateVersionDetailsAndPath(c.Args()[0])
    bintrayDetails := createBintrayDetails(c)
    commands.DownloadFile(versionDetails, path, bintrayDetails)
}

func signUrl(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt url-sign --help'.")
    }
    urlSigningDetails := utils.CreatePathDetails(c.Args()[0])
    urlSigningFlags := createUrlSigningFlags(c)
    commands.SignVersion(urlSigningDetails, urlSigningFlags)
}

func gpgSignFile(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt url-sign --help'.")
    }
    pathDetails := utils.CreatePathDetails(c.Args()[0])
    commands.GpgSignFile(pathDetails, c.String("passphrase"), createBintrayDetails(c))
}

func gpgSignVersion(c *cli.Context) {
    if len(c.Args()) != 1 {
        cliutils.Exit("Wrong number of arguments. Try 'bt url-sign --help'.")
    }
    versionDetails := utils.CreateVersionDetails(c.Args()[0])
    commands.GpgSignVersion(versionDetails, c.String("passphrase"), createBintrayDetails(c))
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
        cliutils.Exit("Wrong number of arguments. Try 'bt ent-keys --help'.")
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
        cliutils.Exit("Expecting show, create, update or delete after the key argument. Got " + c.Args()[0])
    }
}

func entitlements(c *cli.Context) {
    argsSize := len(c.Args())
    if argsSize == 0 {
        cliutils.Exit("Wrong number of arguments. Try 'bt ent --help'.")
    }
    if argsSize == 1 {
        bintrayDetails := createBintrayDetails(c)
        details := commands.CreateVersionDetailsForEntitlements(c.Args()[0])
        commands.ShowEntitlements(bintrayDetails, details)
        return
    }
    if argsSize != 2 {
        cliutils.Exit("Wrong number of arguments. Try 'bt ent --help'.")
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
        cliutils.Exit("Expecting show, create, update or delete before " + c.Args()[1] + ". Got " + c.Args()[0])
    }
}

func createPackageFlags(c *cli.Context, prefix string) *utils.PackageFlags {
    var publicDownloadNumbers string
    var publicStats string
    if c.String(prefix + "pub-dn") != "" {
        publicDownloadNumbers = c.String(prefix + "pub-dn")
        publicDownloadNumbers = strings.ToLower(publicDownloadNumbers)
        if publicDownloadNumbers != "true" && publicDownloadNumbers != "false" {
            cliutils.Exit("The --" + prefix + "pub-dn option should have a boolean value.")
        }
    }
    if c.String(prefix + "pub-stats") != "" {
        publicStats = c.String(prefix + "pub-stats")
        publicStats = strings.ToLower(publicStats)
        if publicStats != "true" && publicStats != "false" {
            cliutils.Exit("The --" + prefix + "pub-stats option should have a boolean value.")
        }
    }

    return &utils.PackageFlags {
        BintrayDetails: createBintrayDetails(c),
        Desc: c.String(prefix + "desc"),
        Labels: c.String(prefix + "labels"),
        Licenses: c.String(prefix + "licenses"),
        CustomLicenses: c.String(prefix + "cust-licenses"),
        VcsUrl: c.String(prefix + "vcs-url"),
        WebsiteUrl: c.String(prefix + "website-url"),
        IssueTrackerUrl: c.String(prefix + "i-tracker-url"),
        GithubRepo: c.String(prefix + "github-repo"),
        GithubReleaseNotesFile: c.String(prefix + "github-rel-notes"),
        PublicDownloadNumbers: publicDownloadNumbers,
        PublicStats: publicStats }
}

func createVersionFlags(c *cli.Context, prefix string) *utils.VersionFlags {
    var githubTagReleaseNotes string
    if c.String(prefix + "github-tag-rel-notes") != "" {
        githubTagReleaseNotes = c.String(prefix + "github-tag-rel-notes")
        githubTagReleaseNotes = strings.ToLower(githubTagReleaseNotes)
        if githubTagReleaseNotes != "true" && githubTagReleaseNotes != "false" {
            cliutils.Exit("The --github-tag-rel-notes option should have a boolean value.")
        }
    }
    return &utils.VersionFlags {
       BintrayDetails: createBintrayDetails(c),
       Desc: c.String(prefix + "desc"),
       VcsTag: c.String(prefix + "vcs-tag"),
       Released: c.String(prefix + "released"),
       GithubReleaseNotesFile: c.String(prefix + "github-rel-notes"),
       GithubUseTagReleaseNotes: githubTagReleaseNotes }
}

func createUrlSigningFlags(c *cli.Context) *commands.UrlSigningFlags {
    if c.String("valid-for") != "" {
        _, err := strconv.ParseInt(c.String("valid-for"), 10, 64)
        if err != nil {
            cliutils.Exit("The '--valid-for' option should have a numeric value.")
        }
    }

    return &commands.UrlSigningFlags {
        BintrayDetails: createBintrayDetails(c),
        Expiry: c.String("expiry"),
        ValidFor: c.String("valid-for"),
        CallbackId: c.String("callback-id"),
        CallbackEmail: c.String("callback-email"),
        CallbackUrl: c.String("callback-url"),
        CallbackMethod: c.String("callback-method") }
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
        cliutils.Exit("Please add the --id option")
    }
    return &commands.EntitlementFlags {
        BintrayDetails: createBintrayDetails(c),
        Id: c.String("id") }
}

func createEntitlementFlagsForCreate(c *cli.Context) *commands.EntitlementFlags {
    if c.String("access") == "" {
        cliutils.Exit("Please add the --access option")
    }
    return &commands.EntitlementFlags {
        BintrayDetails: createBintrayDetails(c),
        Path: c.String("path"),
        Access: c.String("access"),
        Keys: c.String("keys") }
}

func createEntitlementFlagsForUpdate(c *cli.Context) *commands.EntitlementFlags {
    if c.String("id") == "" {
        cliutils.Exit("Please add the --id option")
    }
    if c.String("access") == "" {
        cliutils.Exit("Please add the --access option")
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
            cliutils.Exit("The --ex-check-cache option should have a numeric value.")
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
        cliutils.Exit("Please add the --key option or set the BINTRAY_KEY envrionemt variable")
    }
    apiUrl := c.String("api-url")
    if apiUrl == "" {
        apiUrl = "https://api.bintray.com/"
    }
    downloadServerUrl := c.String("download-url")
    if downloadServerUrl == "" {
        downloadServerUrl = "https://dl.bintray.com/"
    }

    apiUrl = cliutils.AddTrailingSlashIfNeeded(apiUrl)
    downloadServerUrl = cliutils.AddTrailingSlashIfNeeded(downloadServerUrl)
    return &utils.BintrayDetails {
        ApiUrl: apiUrl,
        DownloadServerUrl: downloadServerUrl,
        User: c.String("user"),
        Key: c.String("key") }
}
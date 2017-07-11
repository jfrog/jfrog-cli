package bintray

import (
	"github.com/codegangsta/cli"
	"github.com/jfrogdev/jfrog-cli-go/bintray/commands"
	"github.com/jfrogdev/jfrog-cli-go/bintray/commands/entitlements"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"os"
	"strconv"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/docs/common"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/downloadfile"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/downloadver"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/packageshow"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/packagecreate"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/packageupdate"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/packagedelete"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/versionshow"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/versioncreate"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/versionupdate"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/versiondelete"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/versionpublish"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/accesskeys"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/urlsign"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/gpgsignfile"
	"github.com/jfrogdev/jfrog-cli-go/docs/bintray/gpgsignver"
	entitlementsdocs "github.com/jfrogdev/jfrog-cli-go/docs/bintray/entitlements"
	logsdocs "github.com/jfrogdev/jfrog-cli-go/docs/bintray/logs"
	configdocs "github.com/jfrogdev/jfrog-cli-go/docs/bintray/config"
	uploaddocs "github.com/jfrogdev/jfrog-cli-go/docs/bintray/upload"
	streamdocs "github.com/jfrogdev/jfrog-cli-go/docs/bintray/stream"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "config",
			Flags:     getConfigFlags(),
			Aliases:   []string{"c"},
			Usage:     configdocs.Description,
			HelpName:  common.CreateUsage("bt config", configdocs.Description, configdocs.Usage),
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				configure(c)
			},
		},
		{
			Name:      "upload",
			Flags:     getUploadFlags(),
			Aliases:   []string{"u"},
			Usage:     uploaddocs.Description,
			HelpName:  common.CreateUsage("bt upload", uploaddocs.Description, uploaddocs.Usage),
			UsageText: uploaddocs.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				upload(c)
			},
		},
		{
			Name:      "download-file",
			Flags:     getDownloadFileFlags(),
			Aliases:   []string{"dlf"},
			Usage:     downloadfile.Description,
			HelpName:  common.CreateUsage("bt download-file", downloadfile.Description, downloadfile.Usage),
			UsageText: downloadfile.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				downloadFile(c)
			},
		},
		{
			Name:      "download-ver",
			Flags:     getDownloadVersionFlags(),
			Aliases:   []string{"dlv"},
			Usage:     downloadver.Description,
			HelpName:  common.CreateUsage("bt download-ver", downloadver.Description, downloadver.Usage),
			UsageText: downloadver.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				downloadVersion(c)
			},
		},
		{
			Name:      "package-show",
			Flags:     getFlags(),
			Aliases:   []string{"ps"},
			Usage:     packageshow.Description,
			HelpName:  common.CreateUsage("bt package-show", packageshow.Description, packageshow.Usage),
			UsageText: packageshow.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				showPackage(c)
			},
		},
		{
			Name:      "package-create",
			Flags:     getCreateAndUpdatePackageFlags(),
			Aliases:   []string{"pc"},
			Usage:     packagecreate.Description,
			HelpName:  common.CreateUsage("bt package-create", packagecreate.Description, packagecreate.Usage),
			UsageText: packagecreate.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				createPackage(c)
			},
		},
		{
			Name:      "package-update",
			Flags:     getCreateAndUpdatePackageFlags(),
			Aliases:   []string{"pu"},
			Usage:     packageupdate.Description,
			HelpName:  common.CreateUsage("bt package-update", packageupdate.Description, packageupdate.Usage),
			UsageText: packageupdate.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				updatePackage(c)
			},
		},
		{
			Name:      "package-delete",
			Flags:     getDeletePackageAndVersionFlags(),
			Aliases:   []string{"pd"},
			Usage:     packagedelete.Description,
			HelpName:  common.CreateUsage("bt package-delete", packagedelete.Description, packagedelete.Usage),
			UsageText: packagedelete.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				deletePackage(c)
			},
		},
		{
			Name:      "version-show",
			Flags:     getFlags(),
			Aliases:   []string{"vs"},
			Usage:     versionshow.Description,
			HelpName:  common.CreateUsage("bt version-show", versionshow.Description, versionshow.Usage),
			UsageText: versionshow.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				showVersion(c)
			},
		},
		{
			Name:      "version-create",
			Flags:     getCreateAndUpdateVersionFlags(),
			Aliases:   []string{"vc"},
			Usage:     versioncreate.Description,
			HelpName:  common.CreateUsage("bt version-create", versioncreate.Description, versioncreate.Usage),
			UsageText: versioncreate.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				createVersion(c)
			},
		},
		{
			Name:      "version-update",
			Flags:     getCreateAndUpdateVersionFlags(),
			Aliases:   []string{"vu"},
			Usage:     versionupdate.Description,
			HelpName:  common.CreateUsage("bt version-update", versionupdate.Description, versionupdate.Usage),
			UsageText: versionupdate.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				updateVersion(c)
			},
		},
		{
			Name:      "version-delete",
			Flags:     getDeletePackageAndVersionFlags(),
			Aliases:   []string{"vd"},
			Usage:     versiondelete.Description,
			HelpName:  common.CreateUsage("bt version-delete", versiondelete.Description, versiondelete.Usage),
			UsageText: versiondelete.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				deleteVersion(c)
			},
		},
		{
			Name:      "version-publish",
			Flags:     getFlags(),
			Aliases:   []string{"vp"},
			Usage:     versionpublish.Description,
			HelpName:  common.CreateUsage("bt version-publish", versionpublish.Description, versionpublish.Usage),
			UsageText: versionpublish.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				publishVersion(c)
			},
		},
		{
			Name:      "entitlements",
			Flags:     getEntitlementsFlags(),
			Aliases:   []string{"ent"},
			Usage:     entitlementsdocs.Description,
			HelpName:  common.CreateUsage("bt entitlements", entitlementsdocs.Description, entitlementsdocs.Usage),
			UsageText: entitlementsdocs.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				handleEntitlements(c)
			},
		},
		{
			Name:      "access-keys",
			Flags:     getAccessKeysFlags(),
			Aliases:   []string{"acc-keys"},
			Usage:     accesskeys.Description,
			HelpName:  common.CreateUsage("bt access-keys", accesskeys.Description, accesskeys.Usage),
			UsageText: accesskeys.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				accessKeys(c)
			},
		},
		{
			Name:      "url-sign",
			Flags:     getUrlSigningFlags(),
			Aliases:   []string{"us"},
			Usage:     urlsign.Description,
			HelpName:  common.CreateUsage("bt url-sign", urlsign.Description, urlsign.Usage),
			UsageText: urlsign.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				signUrl(c)
			},
		},
		{
			Name:      "gpg-sign-file",
			Flags:     getGpgSigningFlags(),
			Aliases:   []string{"gsf"},
			Usage:     gpgsignfile.Description,
			HelpName:  common.CreateUsage("bt gpg-sign-file", gpgsignfile.Description, gpgsignfile.Usage),
			UsageText: gpgsignfile.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				gpgSignFile(c)
			},
		},
		{
			Name:      "gpg-sign-ver",
			Flags:     getGpgSigningFlags(),
			Aliases:   []string{"gsv"},
			Usage:     gpgsignver.Description,
			HelpName:  common.CreateUsage("bt gpg-sign-ver", gpgsignver.Description, gpgsignver.Usage),
			UsageText: gpgsignver.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				gpgSignVersion(c)
			},
		},
		{
			Name:      "logs",
			Flags:     getFlags(),
			Aliases:   []string{"l"},
			Usage:     logsdocs.Description,
			HelpName:  common.CreateUsage("bt logs", logsdocs.Description, logsdocs.Usage),
			UsageText: logsdocs.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				logs(c)
			},
		},
		{
			Name:      "stream",
			Flags:     getStreamFlags(),
			Aliases:   []string{"st"},
			Usage:     streamdocs.Description,
			HelpName:  common.CreateUsage("bt stream", streamdocs.Description, streamdocs.Usage),
			UsageText: streamdocs.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action: func(c *cli.Context) {
				stream(c)
			},
		},
	}
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "user",
			Value: "",
			Usage:  "[Optional] Bintray username. If not set, the subject sent as part of the command argument is used for authentication.",
		},
		cli.StringFlag{
			Name:   "key",
			Value: "",
			Usage:  "[Mandatory] Bintray API key",
		},
	}
}

func getStreamFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "include",
			Value: "",
			Usage: "[Optional] List of events type in the form of \"value1;value2;...\" leave empty to include all.",
		},
	}...)
}

func getConfigFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "interactive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not want the config command to be interactive.",
		},
	}
	flags = append(flags, getFlags()...)
	return append(flags, cli.StringFlag{
		Name:  "licenses",
		Value: "",
		Usage: "[Optional] Default package licenses in the form of Apache-2.0,GPL-3.0...",
	})
}

func getPackageFlags(prefix string) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "licenses",
			Value: "",
			Usage: "[Mandatory for OSS] Package licenses in the form of Apache-2.0,GPL-3.0...",
		},
		cli.StringFlag{
			Name:  "vcs-url",
			Value: "",
			Usage: "[Mandatory for OSS] Package VCS URL.",
		},
		cli.StringFlag{
			Name:  "pub-dn",
			Value: "",
			Usage: "[Default: false] Public download numbers.",
		},
		cli.StringFlag{
			Name:  "pub-stats",
			Value: "",
			Usage: "[Default: true] Public statistics.",
		},
		cli.StringFlag{
			Name:  "desc",
			Value: "",
			Usage: "[Optional] Package description.",
		},
		cli.StringFlag{
			Name:  "labels",
			Value: "",
			Usage: "[Optional] Package lables in the form of \"lable11\",\"lable2\"...",
		},
		cli.StringFlag{
			Name:  "cust-licenses",
			Value: "",
			Usage: "[Optional] Package custom licenses in the form of \"my-license-1\",\"my-license-2\"...",
		},
		cli.StringFlag{
			Name:  "website-url",
			Value: "",
			Usage: "[Optional] Package web site URL.",
		},
		cli.StringFlag{
			Name:  "issuetracker-url",
			Value: "",
			Usage: "[Optional] Package Issues Tracker URL.",
		},
		cli.StringFlag{
			Name:  "github-repo",
			Value: "",
			Usage: "[Optional] Package Github repository.",
		},
		cli.StringFlag{
			Name:  "github-rel-notes",
			Value: "",
			Usage: "[Optional] Github release notes file.",
		},
	}
}

func getVersionFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "github-tag-rel-notes",
			Value: "",
			Usage: "[Default: false] Set to true if you wish to use a Github tag release notes.",
		},
		cli.StringFlag{
			Name:  "desc",
			Value: "",
			Usage: "[Optional] Version description.",
		},
		cli.StringFlag{
			Name:  "released",
			Value: "",
			Usage: "[Optional] Release date in ISO8601 format (yyyy-MM-dd'T'HH:mm:ss.SSSZ)",
		},
		cli.StringFlag{
			Name:  "github-rel-notes",
			Value: "",
			Usage: "[Optional] Github release notes file.",
		},
		cli.StringFlag{
			Name:  "vcs-tag",
			Value: "",
			Usage: "[Optional] VCS tag.",
		},
	}
}

func getCreateAndUpdatePackageFlags() []cli.Flag {
	return append(getFlags(), getPackageFlags("")...)
}

func getCreateAndUpdateVersionFlags() []cli.Flag {
	return append(getFlags(), getVersionFlags()...)
}

func getDeletePackageAndVersionFlags() []cli.Flag {
	return append(getFlags(), cli.StringFlag{
		Name:  "quiet",
		Value: "",
		Usage: "[Default: false] Set to true to skip the delete confirmation message.",
	})
}

func getDownloadFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "flat",
			Value: "",
			Usage: "[Default: false] Set to true if you do not wish to have the Bintray path structure created locally for your downloaded files.",
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
			Name:  "unpublished",
			Value: "",
			Usage: "[Default: false] Download both published and unpublished files.",
		},
	}
}

func getDownloadFileFlags() []cli.Flag {
	return append(getFlags(), getDownloadFlags()...)
}

func getDownloadVersionFlags() []cli.Flag {
	flags := append(getFlags(), cli.StringFlag{
		Name:  "threads",
		Value: "",
		Usage: "[Default: 3] Number of artifacts to download in parallel.",
	})
	return append(flags, getDownloadFlags()...)
}

func getUploadFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "recursive",
			Value: "",
			Usage: "[Default: true] Set to false if you do not wish to collect files in sub-folders to be uploaded to Bintray.",
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
			Name:  "publish",
			Value: "",
			Usage: "[Default: false] Set to true to publish the uploaded files.",
		},
		cli.StringFlag{
			Name:  "override",
			Value: "",
			Usage: "[Default: false] Set to true to enable overriding existing published files.",
		},
		cli.StringFlag{
			Name:  "explode",
			Value: "",
			Usage: "[Default: false] Set to true to explode archived files after upload.",
		},
		cli.StringFlag{
			Name:  "threads",
			Value: "",
			Usage: "[Default: 3] Number of artifacts to upload in parallel.",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "[Default: false] Set to true to disable communication with Bintray.",
		},
		cli.StringFlag{
			Name:  "deb",
			Value: "",
			Usage: "[Optional] Used for Debian packages in the form of distribution/component/architecture.",
		},
	}...)
}

func getEntitlementsFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Usage: "[Optional] Entitlement ID. Used for entitlements update.",
		},
		cli.StringFlag{
			Name:  "access",
			Usage: "[Optional] Entitlement access. Used for entitlements creation and update.",
		},
		cli.StringFlag{
			Name:  "keys",
			Usage: "[Optional] Used for entitlements creation and update. List of Access Keys in the form of \"key1\",\"key2\"...",
		},
		cli.StringFlag{
			Name:  "path",
			Usage: "[Optional] Entitlement path. Used for entitlements creating and update.",
		},
	}...)
}

func getAccessKeysFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "org",
			Usage: "[Optional] Bintray organization",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "[Optional] Access Key password.",
		},
		cli.StringFlag{
			Name:  "expiry",
			Usage: "[Optional] Access Key expiry (required for 'jfrog bt acc-keys show/create/update/delete'",
		},
		cli.StringFlag{
			Name:  "ex-check-url",
			Usage: "[Optional] You can optionally provide an existence check directive, in the form of a callback URL, to verify whether the source identity of the Access Key still exists.",
		},
		cli.StringFlag{
			Name:  "ex-check-cache",
			Usage: "[Optional] You can optionally provide the period in seconds for the callback URL results cache.",
		},
		cli.StringFlag{
			Name:  "white-cidrs",
			Usage: "[Optional] Specifying white CIDRs in the form of 127.0.0.1/22,193.5.0.1/92 will allow access only for those IPs that exist in that address range.",
		},
		cli.StringFlag{
			Name:  "black-cidrs",
			Usage: "[Optional] Specifying black CIDRs in the form of 127.0.0.1/22,193.5.0.1/92 will block access for all IPs that exist in the specified range.",
		},
		cli.StringFlag{
			Name:  "api-only",
			Usage: "[Default: true] You can set api_only to false to allow access keys access to Bintray UI as well as to the API.",
		},
	}...)
}

func getUrlSigningFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "expiry",
			Usage: "[Optional] An expiry date for the URL, in Unix epoch time in milliseconds, after which the URL will be invalid. By default, expiry date will be 24 hours.",
		},
		cli.StringFlag{
			Name:  "valid-for",
			Usage: "[Optional] The number of seconds since generation before the URL expires. Mutually exclusive with the --expiry option.",
		},
		cli.StringFlag{
			Name:  "callback-id",
			Usage: "[Optional] An applicative identifier for the request. This identifier appears in download logs and is used in email and download webhook notifications.",
		},
		cli.StringFlag{
			Name:  "callback-email",
			Usage: "[Optional] An email address to send mail to when a user has used the download URL. This requiers a callback_id. The callback-id will be included in the mail message.",
		},
		cli.StringFlag{
			Name:  "callback-url",
			Usage: "[Optional] A webhook URL to call when a user has used the download URL.",
		},
		cli.StringFlag{
			Name:  "callback-method",
			Usage: "[Optional] HTTP method to use for making the callback. Will use POST by default. Supported methods are: GET, POST, PUT and HEAD.",
		},
	}...)
}

func getGpgSigningFlags() []cli.Flag {
	return append(getFlags(), cli.StringFlag{
		Name:  "passphrase",
		Usage: "[Optional] GPG passphrase.",
	})
}

func configure(c *cli.Context) {
	if c.NArg() > 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	} else if c.NArg() == 1 {
		if c.Args().Get(0) == "show" {
			commands.ShowConfig()
		} else if c.Args().Get(0) == "clear" {
			commands.ClearConfig()
		} else {
			cliutils.Exit(cliutils.ExitCodeError, "Unknown argument '" + c.Args().Get(0) + "'. Available arguments are 'show' and 'clear'.")
		}
	} else {
		interactive := cliutils.GetBoolFlagValue(c, "interactive", true)
		if !interactive {
			if c.String("user") == "" || c.String("key") == "" {
				cliutils.Exit(cliutils.ExitCodeError, "The --user and --key options are mandatory when the --interactive option is set to false")
			}
		}
		bintrayDetails, err := createBintrayDetails(c, false)
		cliutils.ExitOnErr(err)
		commands.Config(bintrayDetails, nil, interactive)
	}
}

func showPackage(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	packageDetails, err := utils.CreatePackageDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	bintrayDetails, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.ShowPackage(packageDetails, bintrayDetails)
	cliutils.ExitOnErr(err)
}

func showVersion(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	versionDetails, err := utils.CreateVersionDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	bintrayDetails, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.ShowVersion(versionDetails, bintrayDetails)
	cliutils.ExitOnErr(err)
}

func createPackage(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	packageDetails, err := utils.CreatePackageDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	packageFlags, err := createPackageFlags(c)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.CreatePackage(packageDetails, packageFlags)
	cliutils.ExitOnErr(err)
}

func createVersion(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	versionDetails, err := utils.CreateVersionDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	versionFlags, err := createVersionFlags(c, "")
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.CreateVersion(versionDetails, versionFlags)
	cliutils.ExitOnErr(err)
}

func updateVersion(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	versionDetails, err := utils.CreateVersionDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	versionFlags, err := createVersionFlags(c, "")
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.UpdateVersion(versionDetails, versionFlags)
	cliutils.ExitOnErr(err)
}

func updatePackage(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	packageDetails, err := utils.CreatePackageDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	packageFlags, err := createPackageFlags(c)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.UpdatePackage(packageDetails, packageFlags)
	cliutils.ExitOnErr(err)
}

func deletePackage(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	packageDetails, err := utils.CreatePackageDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	bintrayDetails, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}

	if !c.Bool("quiet") {
		confirmed := cliutils.InteractiveConfirm("Delete package " + packageDetails.Package + "?")
		if !confirmed {
			return
		}
	}
	err = commands.DeletePackage(packageDetails, bintrayDetails)
	cliutils.ExitOnErr(err)
}

func deleteVersion(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	versionDetails, err := utils.CreateVersionDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	bintrayDetails, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}

	if !c.Bool("quiet") {
		confirmed := cliutils.InteractiveConfirm("Delete version " + versionDetails.Version +
			" of package " + versionDetails.Package + "?")
		if !confirmed {
			return
		}
	}
	err = commands.DeleteVersion(versionDetails, bintrayDetails)
	cliutils.ExitOnErr(err)
}

func publishVersion(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	versionDetails, err := utils.CreateVersionDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	bintrayDetails, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.PublishVersion(versionDetails, bintrayDetails)
	cliutils.ExitOnErr(err)
}

func downloadVersion(c *cli.Context) {
	if c.NArg() < 1 || c.NArg() > 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	versionDetails, err := commands.CreateVersionDetailsForDownloadVersion(c.Args().Get(0))
	cliutils.ExitOnErr(err)
	targetPath := c.Args().Get(1)
	if strings.HasPrefix(targetPath, "/") {
		targetPath = targetPath[1:]
	}
	flags, err := createDownloadFlags(c)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	downloaded, failed, err := commands.DownloadVersion(versionDetails, targetPath, flags)
	cliutils.ExitOnErr(err)
	if failed > 0 {
		if downloaded > 0 {
			cliutils.Exit(cliutils.ExitCodeWarning, "")
		}
		cliutils.Exit(cliutils.ExitCodeError, "")
	}
}

func upload(c *cli.Context) {
	if c.NArg() < 2 || c.NArg() > 3 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	localPath := c.Args().Get(0)
	versionDetails, err := utils.CreateVersionDetails(c.Args().Get(1))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	uploadPath := c.Args().Get(2)
	if strings.HasPrefix(uploadPath, "/") {
		uploadPath = uploadPath[1:]
	}

	uploadFlags, err := createUploadFlags(c)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	uploaded, failed, err := commands.Upload(versionDetails, localPath, uploadPath, uploadFlags)
	cliutils.ExitOnErr(err)
	if failed > 0 {
		if uploaded > 0 {
			cliutils.Exit(cliutils.ExitCodeWarning, "")
		}
		cliutils.Exit(cliutils.ExitCodeError, "")
	}
}

func downloadFile(c *cli.Context) {
	if c.NArg() < 1 || c.NArg() > 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	pathDetails, err := utils.CreatePathDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	targetPath := c.Args().Get(1)
	if strings.HasPrefix(targetPath, "/") {
		targetPath = targetPath[1:]
	}

	flags, err := createDownloadFlags(c)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.DownloadFile(pathDetails, targetPath, flags)
	cliutils.ExitOnErr(err)
}

func signUrl(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	urlSigningDetails, err := utils.CreatePathDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	urlSigningFlags, err := createUrlSigningFlags(c)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.SignVersion(urlSigningDetails, urlSigningFlags)
	cliutils.ExitOnErr(err)
}

func gpgSignFile(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	pathDetails, err := utils.CreatePathDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	flags, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.GpgSignFile(pathDetails, c.String("passphrase"), flags)
	cliutils.ExitOnErr(err)
}

func logs(c *cli.Context) {
	bintrayDetails, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	if c.NArg() == 1 {
		packageDetails, err := utils.CreatePackageDetails(c.Args().Get(0))
		cliutils.ExitOnErr(err)
		err = commands.LogsList(packageDetails, bintrayDetails)
		cliutils.ExitOnErr(err)
	} else if c.NArg() == 3 {
		if c.Args().Get(0) == "download" {
			packageDetails, err := utils.CreatePackageDetails(c.Args().Get(1))
			if err != nil {
				cliutils.Exit(cliutils.ExitCodeError, err.Error())
			}
			err = commands.DownloadLog(packageDetails, c.Args().Get(2), bintrayDetails)
			cliutils.ExitOnErr(err)
		} else {
			cliutils.Exit(cliutils.ExitCodeError, "Unkown argument " + c.Args().Get(0) + ". " + cliutils.GetDocumentationMessage())
		}
	} else {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
}

func stream(c *cli.Context) {
	bintrayDetails, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}

	streamDetails := &commands.StreamDetails{
		BintrayDetails: bintrayDetails,
		Subject: c.Args().Get(0),
		Include: c.String("include"),
	}
	err = commands.Stream(streamDetails, os.Stdout)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, "")
	}
}

func gpgSignVersion(c *cli.Context) {
	if c.NArg() != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	versionDetails, err := utils.CreateVersionDetails(c.Args().Get(0))
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	flags, err := createBintrayDetails(c, true)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, err.Error())
	}
	err = commands.GpgSignVersion(versionDetails, c.String("passphrase"), flags)
	cliutils.ExitOnErr(err)
}

func accessKeys(c *cli.Context) {
	org := c.String("org")
	if c.NArg() == 0 {
		bintrayDetails, err := createBintrayDetails(c, true)
		cliutils.ExitOnErr(err)
		err = commands.ShowAccessKeys(bintrayDetails, org)
		cliutils.ExitOnErr(err)
		return
	}
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	keyId := c.Args().Get(1)
	var flags *commands.AccessKeyFlags
	var err error
	switch c.Args().Get(0) {
	case "show":
		flags, err = createAccessKeyFlagsForShowAndDelete(keyId, c)
		cliutils.ExitOnErr(err)
		err = commands.ShowAccessKey(flags, org)
	case "create":
		flags, err = createAccessKeyFlagsForCreateAndUpdate(keyId, c)
		cliutils.ExitOnErr(err)
		err = commands.CreateAccessKey(flags, org)
	case "update":
		flags, err = createAccessKeyFlagsForCreateAndUpdate(keyId, c)
		cliutils.ExitOnErr(err)
		err = commands.UpdateAccessKey(flags, org)
	case "delete":
		flags, err = createAccessKeyFlagsForShowAndDelete(keyId, c)
		cliutils.ExitOnErr(err)
		err = commands.DeleteAccessKey(flags, org)
	default:
		cliutils.Exit(cliutils.ExitCodeError, "Expecting show, create, update or delete before the key argument. Got " + c.Args().Get(0))
	}
	cliutils.ExitOnErr(err)
}

func handleEntitlements(c *cli.Context) {
	if c.NArg() == 0 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	if c.NArg() == 1 {
		bintrayDetails, err := createBintrayDetails(c, true)
		cliutils.ExitOnErr(err)
		details, err := entitlements.CreateVersionDetails(c.Args().Get(0))
		cliutils.ExitOnErr(err)
		err = entitlements.ShowEntitlements(bintrayDetails, details)
		cliutils.ExitOnErr(err)
		return
	}
	if c.NArg() != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	details, err := entitlements.CreateVersionDetails(c.Args().Get(1))
	cliutils.ExitOnErr(err)

	var flags *entitlements.EntitlementFlags
	switch c.Args().Get(0) {
	case "show":
		flags, err = createEntitlementFlagsForShowAndDelete(c)
		cliutils.ExitOnErr(err)
		err = entitlements.ShowEntitlement(flags, details)
	case "create":
		flags, err = createEntitlementFlagsForCreate(c)
		cliutils.ExitOnErr(err)
		err = entitlements.CreateEntitlement(flags, details)
	case "update":
		flags, err = createEntitlementFlagsForUpdate(c)
		cliutils.ExitOnErr(err)
		err = entitlements.UpdateEntitlement(flags, details)
	case "delete":
		flags, err = createEntitlementFlagsForShowAndDelete(c)
		cliutils.ExitOnErr(err)
		err = entitlements.DeleteEntitlement(flags, details)
	default:
		cliutils.Exit(cliutils.ExitCodeError, "Expecting show, create, update or delete before " + c.Args().Get(1) + ". Got " + c.Args().Get(0))
	}
	cliutils.ExitOnErr(err)
}

func createPackageFlags(c *cli.Context) (*utils.PackageFlags, error) {
	var publicDownloadNumbers string
	var publicStats string
	if c.String("pub-dn") != "" {
		publicDownloadNumbers = c.String("pub-dn")
		publicDownloadNumbers = strings.ToLower(publicDownloadNumbers)
		if publicDownloadNumbers != "true" && publicDownloadNumbers != "false" {
			cliutils.Exit(cliutils.ExitCodeError, "The --pub-dn option should have a boolean value.")
		}
	}
	if c.String("pub-stats") != "" {
		publicStats = c.String("pub-stats")
		publicStats = strings.ToLower(publicStats)
		if publicStats != "true" && publicStats != "false" {
			cliutils.Exit(cliutils.ExitCodeError, "The --pub-stats option should have a boolean value.")
		}
	}
	licenses := c.String("licenses")
	if licenses == "" {
		confDetails, err := commands.GetConfig()
		if err != nil {
			return nil, err
		}
		licenses = confDetails.DefPackageLicenses
	}
	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &utils.PackageFlags{
		BintrayDetails:         details,
		Desc:                   c.String("desc"),
		Labels:                 c.String("labels"),
		Licenses:               licenses,
		CustomLicenses:         c.String("cust-licenses"),
		VcsUrl:                 c.String("vcs-url"),
		WebsiteUrl:             c.String("website-url"),
		IssueTrackerUrl:        c.String("issuetracker-url"),
		GithubRepo:             c.String("github-repo"),
		GithubReleaseNotesFile: c.String("github-rel-notes"),
		PublicDownloadNumbers:  publicDownloadNumbers,
		PublicStats:            publicStats}, nil
}

func createVersionFlags(c *cli.Context, prefix string) (*utils.VersionFlags, error) {
	var githubTagReleaseNotes string
	if c.String("github-tag-rel-notes") != "" {
		githubTagReleaseNotes = c.String("github-tag-rel-notes")
		githubTagReleaseNotes = strings.ToLower(githubTagReleaseNotes)
		if githubTagReleaseNotes != "true" && githubTagReleaseNotes != "false" {
			cliutils.Exit(cliutils.ExitCodeError, "The --github-tag-rel-notes option should have a boolean value.")
		}
	}
	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &utils.VersionFlags{
		BintrayDetails:           details,
		Desc:                     c.String("desc"),
		VcsTag:                   c.String("vcs-tag"),
		Released:                 c.String("released"),
		GithubReleaseNotesFile:   c.String("github-rel-notes"),
		GithubUseTagReleaseNotes: githubTagReleaseNotes}, nil
}

func createUrlSigningFlags(c *cli.Context) (*commands.UrlSigningFlags, error) {
	if c.String("valid-for") != "" {
		_, err := strconv.ParseInt(c.String("valid-for"), 10, 64)
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, "The '--valid-for' option should have a numeric value.")
		}
	}

	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &commands.UrlSigningFlags{
		BintrayDetails: details,
		Expiry:         c.String("expiry"),
		ValidFor:       c.String("valid-for"),
		CallbackId:     c.String("callback-id"),
		CallbackEmail:  c.String("callback-email"),
		CallbackUrl:    c.String("callback-url"),
		CallbackMethod: c.String("callback-method")}, nil
}

func createUploadFlags(c *cli.Context) (*commands.UploadFlags, error) {
	deb := c.String("deb")
	if deb != "" && len(strings.Split(deb, "/")) != 3 {
		cliutils.Exit(cliutils.ExitCodeError, "The --deb option should be in the form of distribution/component/architecture")
	}
	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &commands.UploadFlags{
		BintrayDetails: details,
		Recursive:      cliutils.GetBoolFlagValue(c, "recursive", true),
		Flat:           cliutils.GetBoolFlagValue(c, "flat", true),
		Publish:        cliutils.GetBoolFlagValue(c, "publish", false),
		Override:       cliutils.GetBoolFlagValue(c, "override", false),
		Explode:        cliutils.GetBoolFlagValue(c, "explode", false),
		UseRegExp:      c.Bool("regexp"),
		Threads:        getThreadsOptionValue(c),
		Deb:            deb,
		DryRun:         c.Bool("dry-run")}, nil
}

func getThreadsOptionValue(c *cli.Context) (threads int) {
	if c.String("threads") == "" {
		threads = 3
	} else {
		var err error
		threads, err = strconv.Atoi(c.String("threads"))
		if err != nil || threads < 1 {
			cliutils.Exit(cliutils.ExitCodeError, "The '--threads' option should have a numeric positive value.")
		}
	}
	return
}

func createDownloadFlags(c *cli.Context) (*utils.DownloadFlags, error) {
	flat := false
	if c.String("flat") != "" {
		flat = c.Bool("flat")
	}
	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &utils.DownloadFlags{
		BintrayDetails: details,
		Threads:            getThreadsOptionValue(c),
		MinSplitSize:       getMinSplitFlag(c),
		SplitCount:         getSplitCountFlag(c),
		IncludeUnpublished: cliutils.GetBoolFlagValue(c, "unpublished", false),
		Flat:               flat}, nil
}

func createEntitlementFlagsForShowAndDelete(c *cli.Context) (*entitlements.EntitlementFlags, error) {
	if c.String("id") == "" {
		cliutils.Exit(cliutils.ExitCodeError, "Please add the --id option")
	}
	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &entitlements.EntitlementFlags{
		BintrayDetails: details,
		Id:             c.String("id")}, nil
}

func createEntitlementFlagsForCreate(c *cli.Context) (*entitlements.EntitlementFlags, error) {
	if c.String("access") == "" {
		cliutils.Exit(cliutils.ExitCodeError, "Please add the --access option")
	}
	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &entitlements.EntitlementFlags{
		BintrayDetails: details,
		Path:           c.String("path"),
		Access:         c.String("access"),
		Keys:           c.String("keys")}, nil
}

func createEntitlementFlagsForUpdate(c *cli.Context) (*entitlements.EntitlementFlags, error) {
	if c.String("id") == "" {
		cliutils.Exit(cliutils.ExitCodeError, "Please add the --id option")
	}
	if c.String("access") == "" {
		cliutils.Exit(cliutils.ExitCodeError, "Please add the --access option")
	}
	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &entitlements.EntitlementFlags{
		BintrayDetails: details,
		Id:             c.String("id"),
		Path:           c.String("path"),
		Access:         c.String("access"),
		Keys:           c.String("keys")}, nil
}

func createAccessKeyFlagsForShowAndDelete(keyId string, c *cli.Context) (*commands.AccessKeyFlags, error) {
	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &commands.AccessKeyFlags{
		BintrayDetails: details,
		Id:             keyId}, nil
}

func createAccessKeyFlagsForCreateAndUpdate(keyId string, c *cli.Context) (*commands.AccessKeyFlags, error) {
	var cachePeriod int
	if c.String("ex-check-cache") != "" {
		var err error
		cachePeriod, err = strconv.Atoi(c.String("ex-check-cache"))
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, "The --ex-check-cache option should have a numeric value.")
		}
	}

	var expiry int64
	if c.String("expiry") != "" {
		var err error
		expiry, err = strconv.ParseInt(c.String("expiry"), 10, 64)
		if err != nil {
			cliutils.Exit(cliutils.ExitCodeError, "The --expiry option should have a numeric value.")
		}
	}

	details, err := createBintrayDetails(c, true)
	if err != nil {
		return nil, err
	}
	return &commands.AccessKeyFlags{
		BintrayDetails:      details,
		Id:                  keyId,
		Password:            c.String("password"),
		Expiry:              expiry,
		ExistenceCheckUrl:   c.String("ex-check-url"),
		ExistenceCheckCache: cachePeriod,
		WhiteCidrs:          c.String("white-cidrs"),
		BlackCidrs:          c.String("black-cidrs"),
		ApiOnly:             c.String("api-only")}, nil
}

func offerConfig(c *cli.Context) (*config.BintrayDetails, error) {
	exists, err := config.IsBintrayConfExists()
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, nil
	}
	val, err := cliutils.GetBoolEnvValue("JFROG_CLI_OFFER_CONFIG", true)
	if err != nil {
		return nil, err
	}
	if !val {
		config.SaveBintrayConf(new(config.BintrayDetails))
		return nil, nil
	}
	msg := "Some CLI commands require the following common options:\n" +
			"- User\n" +
			"- API Key\n" +
			"- Default Package Licenses\n" +
			"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n" +
			"You can also configure these parameters later using the 'config' command.\n" +
			"Configure now?"
	confirmed := cliutils.InteractiveConfirm(msg)
	if !confirmed {
		config.SaveBintrayConf(new(config.BintrayDetails))
		return nil, nil
	}
	bintrayDetails, err := createBintrayDetails(c, false)
	if err != nil {
		return nil, err
	}
	details, err := commands.Config(nil, bintrayDetails, true)
	cliutils.ExitOnErr(err)
	details.ApiUrl = bintrayDetails.ApiUrl
	details.DownloadServerUrl = bintrayDetails.DownloadServerUrl
	return details, nil
}

func createBintrayDetails(c *cli.Context, includeConfig bool) (*config.BintrayDetails, error) {
	if includeConfig {
		bintrayDetails, err := offerConfig(c)
		if err != nil {
			return nil, err
		}
		if bintrayDetails != nil {
			return bintrayDetails, nil
		}
	}
	user := c.String("user")
	key := c.String("key")
	defaultPackageLicenses := c.String("licenses")
	if includeConfig && (user == "" || key == "" || defaultPackageLicenses == "") {
		confDetails, err := commands.GetConfig()
		if err != nil {
			return nil, err
		}
		if user == "" {
			user = confDetails.User
		}
		if key == "" {
			key = confDetails.Key
		}
		if key == "" {
			cliutils.Exit(cliutils.ExitCodeError, "Please set your Bintray API key using the config command or send it as the --key option.")
		}
		if defaultPackageLicenses == "" {
			defaultPackageLicenses = confDetails.DefPackageLicenses
		}
	}
	apiUrl := os.Getenv("JFROG_CLI_BINTRAY_API_URL")
	if apiUrl == "" {
		apiUrl = "https://bintray.com/api/v1/"
	}
	downloadServerUrl := os.Getenv("JFROG_CLI_BINTRAY_DOWNLOAD_URL")
	if downloadServerUrl == "" {
		downloadServerUrl = "https://dl.bintray.com/"
	}
	apiUrl = cliutils.AddTrailingSlashIfNeeded(apiUrl)
	downloadServerUrl = cliutils.AddTrailingSlashIfNeeded(downloadServerUrl)
	return &config.BintrayDetails{
		ApiUrl:             apiUrl,
		DownloadServerUrl:  downloadServerUrl,
		User:               user,
		Key:                key,
		DefPackageLicenses: defaultPackageLicenses}, nil
}

func getMinSplitFlag(c *cli.Context) int64 {
	if c.String("min-split") == "" {
		return 5120
	}
	minSplit, err := strconv.ParseInt(c.String("min-split"), 10, 64)
	if err != nil {
		cliutils.PrintHelpAndExitWithError("The '--min-split' option should have a numeric value.", c)
	}
	return minSplit
}

func getSplitCountFlag(c *cli.Context) int {
	if c.String("split-count") == "" {
		return 3
	}
	splitCount, err := strconv.Atoi(c.String("split-count"))
	if err != nil {
		cliutils.PrintHelpAndExitWithError("The '--split-count' option should have a numeric value.", c)
	}
	if splitCount > 15 {
		cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option value is limitted to a maximum of 15.")
	}
	if splitCount < 0 {
		cliutils.Exit(cliutils.ExitCodeError, "The '--split-count' option cannot have a negative value.")
	}
	return splitCount
}
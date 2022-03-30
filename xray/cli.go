package xray

import (
	"errors"
	"time"

	corecommon "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommondocs "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/curl"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/offlineupdate"
	"github.com/jfrog/jfrog-cli/docs/common"
	auditgodocs "github.com/jfrog/jfrog-cli/docs/xray/auditgo"
	"github.com/jfrog/jfrog-cli/docs/xray/auditgradle"
	"github.com/jfrog/jfrog-cli/docs/xray/auditmvn"
	auditnpmdocs "github.com/jfrog/jfrog-cli/docs/xray/auditnpm"
	auditpipdocs "github.com/jfrog/jfrog-cli/docs/xray/auditpip"
	curldocs "github.com/jfrog/jfrog-cli/docs/xray/curl"
	offlineupdatedocs "github.com/jfrog/jfrog-cli/docs/xray/offlineupdate"
	scandocs "github.com/jfrog/jfrog-cli/docs/xray/scan"
	"github.com/jfrog/jfrog-cli/scan"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/urfave/cli"
)

const DateFormat = "2006-01-02"

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:            "curl",
			Flags:           cliutils.GetCommandFlags(cliutils.XrCurl),
			Aliases:         []string{"cl"},
			Usage:           curldocs.GetDescription(),
			HelpName:        corecommondocs.CreateUsage("xr curl", curldocs.GetDescription(), curldocs.Usage),
			UsageText:       curldocs.GetArguments(),
			ArgsUsage:       common.CreateEnvVars(),
			BashComplete:    corecommondocs.CreateBashCompletionFunc(),
			SkipFlagParsing: true,
			Action:          curlCmd,
		},
		{
			Name:         "audit-mvn",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditMvn),
			Aliases:      []string{"am"},
			Usage:        auditmvn.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("xr audit-mvn", auditmvn.GetDescription(), auditmvn.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       scan.AuditMvnCmd,
		},
		{
			Name:         "audit-gradle",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditGradle),
			Aliases:      []string{"ag"},
			Usage:        auditgradle.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("xr audit-gradle", auditgradle.GetDescription(), auditgradle.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       scan.AuditGradleCmd,
		},
		{
			Name:         "audit-npm",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditNpm),
			Aliases:      []string{"an"},
			Usage:        auditnpmdocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("xr audit-npm", auditnpmdocs.GetDescription(), auditnpmdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       scan.AuditNpmCmd,
		},
		{
			Name:         "audit-go",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditGo),
			Aliases:      []string{"ago"},
			Usage:        auditgodocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("xr audit-go", auditgodocs.GetDescription(), auditgodocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       scan.AuditGoCmd,
		},
		{
			Name:         "audit-pip",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditPip),
			Aliases:      []string{"ap"},
			Usage:        auditpipdocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("xr audit-pip", auditpipdocs.GetDescription(), auditpipdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       scan.AuditPipCmd,
		},
		{
			Name:         "scan",
			Flags:        cliutils.GetCommandFlags(cliutils.XrScan),
			Aliases:      []string{"s"},
			Usage:        scandocs.GetDescription(),
			UsageText:    scandocs.GetArguments(),
			HelpName:     corecommondocs.CreateUsage("xr scan", scandocs.GetDescription(), scandocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("scan", "xr", c, scan.ScanCmd)
			},
		},
		{
			Name:         "offline-update",
			Usage:        offlineupdatedocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("xr offline-update", offlineupdatedocs.GetDescription(), offlineupdatedocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			Flags:        cliutils.GetCommandFlags(cliutils.OfflineUpdate),
			Aliases:      []string{"ou"},
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       offlineUpdates,
		},
	})
}

func getOfflineUpdatesFlag(c *cli.Context) (flags *offlineupdate.OfflineUpdatesFlags, err error) {
	flags = new(offlineupdate.OfflineUpdatesFlags)
	flags.Version = c.String("version")
	flags.License = c.String("license-id")
	flags.Target = c.String("target")
	if len(flags.License) < 1 {
		return nil, errors.New("the --license-id option is mandatory")
	}
	from := c.String("from")
	to := c.String("to")
	if len(to) > 0 && len(from) < 1 {
		return nil, errors.New("the --from option is mandatory, when the --to option is sent")
	}
	if len(from) > 0 && len(to) < 1 {
		return nil, errors.New("the --to option is mandatory, when the --from option is sent")
	}
	if len(from) > 0 && len(to) > 0 {
		flags.From, err = dateToMilliseconds(from)
		err = errorutils.CheckError(err)
		if err != nil {
			return
		}
		flags.To, err = dateToMilliseconds(to)
		err = errorutils.CheckError(err)
	}
	return
}

func dateToMilliseconds(date string) (dateInMillisecond int64, err error) {
	t, err := time.Parse(DateFormat, date)
	if err != nil {
		err = errorutils.CheckError(err)
		return
	}
	dateInMillisecond = t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
	return
}

func offlineUpdates(c *cli.Context) error {
	offlineUpdateFlags, err := getOfflineUpdatesFlag(c)
	if err != nil {
		return err
	}

	return offlineupdate.OfflineUpdate(offlineUpdateFlags)
}

func curlCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	xrCurlCmd, err := newXrCurlCommand(c)
	if err != nil {
		return err
	}
	return corecommon.Exec(xrCurlCmd)
}

func newXrCurlCommand(c *cli.Context) (*curl.XrCurlCommand, error) {
	curlCommand := corecommon.NewCurlCommand().SetArguments(cliutils.ExtractCommand(c))
	xrCurlCommand := curl.NewXrCurlCommand(*curlCommand)
	xrDetails, err := xrCurlCommand.GetServerDetails()
	if err != nil {
		return nil, err
	}
	xrCurlCommand.SetServerDetails(xrDetails)
	xrCurlCommand.SetUrl(xrDetails.XrayUrl)
	return xrCurlCommand, err
}

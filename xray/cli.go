package xray

import (
	"errors"
	"github.com/jfrog/jfrog-cli/auditscan"
	auditgodocs "github.com/jfrog/jfrog-cli/docs/xray/auditgo"
	"github.com/jfrog/jfrog-cli/docs/xray/auditgradle"
	"github.com/jfrog/jfrog-cli/docs/xray/auditmvn"
	auditnpmdocs "github.com/jfrog/jfrog-cli/docs/xray/auditnpm"
	auditpipdocs "github.com/jfrog/jfrog-cli/docs/xray/auditpip"
	scandocs "github.com/jfrog/jfrog-cli/docs/xray/scan"
	"time"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommondocs "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/curl"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/offlineupdate"
	"github.com/jfrog/jfrog-cli/docs/common"
	curldocs "github.com/jfrog/jfrog-cli/docs/xray/curl"
	offlineupdatedocs "github.com/jfrog/jfrog-cli/docs/xray/offlineupdate"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const DateFormat = "2006-01-02"

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:            "curl",
			Flags:           cliutils.GetCommandFlags(cliutils.XrCurl),
			Aliases:         []string{"cl"},
			Description:     curldocs.Description,
			HelpName:        corecommondocs.CreateUsage("xr curl", curldocs.Description, curldocs.Usage),
			UsageText:       curldocs.Arguments,
			ArgsUsage:       common.CreateEnvVars(),
			BashComplete:    corecommondocs.CreateBashCompletionFunc(),
			SkipFlagParsing: true,
			Action:          curlCmd,
		},
		{
			Name:         "audit-mvn",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditMvn),
			Aliases:      []string{"am"},
			Description:  auditmvn.Description,
			HelpName:     corecommondocs.CreateUsage("xr audit-mvn", auditmvn.Description, auditmvn.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("audit-mvn", c, auditscan.AuditMvnCmd)
			},
		},
		{
			Name:         "audit-gradle",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditGradle),
			Aliases:      []string{"ag"},
			Description:  auditgradle.Description,
			HelpName:     corecommondocs.CreateUsage("xr audit-gradle", auditgradle.Description, auditgradle.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("audit-gradle", c, auditscan.AuditGradleCmd)
			},
		},
		{
			Name:         "audit-npm",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditNpm),
			Aliases:      []string{"an"},
			Description:  auditnpmdocs.Description,
			HelpName:     corecommondocs.CreateUsage("xr audit-npm", auditnpmdocs.Description, auditnpmdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("audit-npm", c, auditscan.AuditNpmCmd)
			},
		},
		{
			Name:         "audit-go",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditGo),
			Aliases:      []string{"ago"},
			Description:  auditgodocs.Description,
			HelpName:     corecommondocs.CreateUsage("xr audit-go", auditgodocs.Description, auditgodocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("audit-go", c, auditscan.AuditGoCmd)
			},
		},
		{
			Name:         "audit-pip",
			Flags:        cliutils.GetCommandFlags(cliutils.AuditPip),
			Aliases:      []string{"ap"},
			Description:  auditpipdocs.Description,
			HelpName:     corecommondocs.CreateUsage("xr audit-pip", auditpipdocs.Description, auditpipdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("audit-pip", c, auditscan.AuditPipCmd)
			},
		},
		{
			Name:         "scan",
			Flags:        cliutils.GetCommandFlags(cliutils.XrScan),
			Aliases:      []string{"s"},
			Description:  scandocs.Description,
			HelpName:     corecommondocs.CreateUsage("xr scan", scandocs.Description, scandocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return cliutils.RunCmdWithDeprecationWarning("scan", c, auditscan.ScanCmd)
			},
		},
		{
			Name:         "offline-update",
			Description:  offlineupdatedocs.Description,
			HelpName:     corecommondocs.CreateUsage("xr offline-update", offlineupdatedocs.Description, offlineupdatedocs.Usage),
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
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	xrCurlCmd, err := newXrCurlCommand(c)
	if err != nil {
		return err
	}
	return commands.Exec(xrCurlCmd)
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

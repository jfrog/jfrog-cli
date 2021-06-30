package xray

import (
	"errors"
	"time"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/npm"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/common/commands"
	corecommondocs "github.com/jfrog/jfrog-cli-core/docs/common"
	coreconfig "github.com/jfrog/jfrog-cli-core/utils/config"
	scan "github.com/jfrog/jfrog-cli-core/xray/commands/audit"
	"github.com/jfrog/jfrog-cli-core/xray/commands/curl"
	"github.com/jfrog/jfrog-cli-core/xray/commands/offlineupdate"
	"github.com/jfrog/jfrog-cli/docs/common"
	auditnpmdocs "github.com/jfrog/jfrog-cli/docs/xray/auditnpm"
	curldocs "github.com/jfrog/jfrog-cli/docs/xray/curl"
	offlineupdatedocs "github.com/jfrog/jfrog-cli/docs/xray/offlineupdate"
	scandocs "github.com/jfrog/jfrog-cli/docs/xray/scan"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const DATE_FORMAT = "2006-01-02"

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
			Name:         "audit-npm",
			Flags:        cliutils.GetCommandFlags(cliutils.XrAuditNpm),
			Aliases:      []string{"an"},
			Description:  auditnpmdocs.Description,
			HelpName:     corecommondocs.CreateUsage("xr audit-npm", auditnpmdocs.Description, auditnpmdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       auditNpmCmd,
		},
		{
			Name:         "scan",
			Flags:        cliutils.GetCommandFlags(cliutils.XrScan),
			Aliases:      []string{"s"},
			Description:  scandocs.Description,
			HelpName:     corecommondocs.CreateUsage("xr scan", scandocs.Description, scandocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       scanCmd,
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
		return nil, errors.New("The --license-id option is mandatory.")
	}
	from := c.String("from")
	to := c.String("to")
	if len(to) > 0 && len(from) < 1 {
		return nil, errors.New("The --from option is mandatory, when the --to option is sent.")
	}
	if len(from) > 0 && len(to) < 1 {
		return nil, errors.New("The --to option is mandatory, when the --from option is sent.")
	}
	if len(from) > 0 && len(to) > 0 {
		flags.From, err = dateToMilliseconds(from)
		errorutils.CheckError(err)
		if err != nil {
			return
		}
		flags.To, err = dateToMilliseconds(to)
		errorutils.CheckError(err)
	}
	return
}

func dateToMilliseconds(date string) (dateInMillisecond int64, err error) {
	t, err := time.Parse(DATE_FORMAT, date)
	if err != nil {
		errorutils.CheckError(err)
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
	if show, err := cliutils.ShowCmdHelpIfNeeded(c); show || err != nil {
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

func auditNpmCmd(c *cli.Context) error {
	serverDetailes, err := CreateServerDetailsWithConfigOffer(c)
	if err != nil {
		return err
	}
	var typeRestriction = npm.All
	switch c.String("type-restriction") {
	case "devOnly":
		typeRestriction = npm.DevOnly
	case "prodOnly":
		typeRestriction = npm.ProdOnly
	}
	xrAuditNpmCmd := scan.NewXrAuditNpmCommand().SetServerDetails(serverDetailes).SetNpmTypeRestriction(typeRestriction)
	return commands.Exec(xrAuditNpmCmd)
}

func scanCmd(c *cli.Context) error {
	serverDetailes, err := CreateServerDetailsWithConfigOffer(c)
	if err != nil {
		return err
	}
	var specFile *spec.SpecFiles
	if c.IsSet("spec") {
		specFile, err = cliutils.GetFileSystemSpec(c)
	} else {
		specFile, err = createDefaultScanSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(specFile.Files, false, false, false)
	if err != nil {
		return err
	}
	threads, err := cliutils.GetThreadsCount(c)
	if err != nil {
		return err
	}
	cliutils.FixWinPathsForFileSystemSourcedCmds(specFile, c)
	xrScanCmd := scan.NewXrBinariesScanCommand().SetServerDetails(serverDetailes).SetThreads(threads).SetSpec(specFile).SetPrintResults(true)
	return commands.Exec(xrScanCmd)
}

func createDefaultScanSpec(c *cli.Context) (*spec.SpecFiles, error) {
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Target("").
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Regexp(c.Bool("regexp")).
		Ant(c.Bool("ant")).
		IncludeDirs(c.Bool("include-dirs")).
		BuildSpec(), nil
}

func CreateServerDetailsWithConfigOffer(c *cli.Context) (*coreconfig.ServerDetails, error) {
	return cliutils.CreateServerDetailsWithConfigOffer(c, true)
}

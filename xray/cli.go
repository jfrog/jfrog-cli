package xray

import (
	"errors"
	offlineupdatedocs "github.com/jfrog/jfrog-cli/docs/xray/offlineupdate"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	corecommondocs "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	npmutils "github.com/jfrog/jfrog-cli-core/v2/utils/npm"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/audit"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/curl"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/offlineupdate"
	"github.com/jfrog/jfrog-cli/docs/common"
	curldocs "github.com/jfrog/jfrog-cli/docs/xray/curl"
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
		// Temporarily disable new Xray commands
		//{
		//	Name:         "audit-npm",
		//	Flags:        cliutils.GetCommandFlags(cliutils.AuditNpm),
		//	Aliases:      []string{"an"},
		//	Description:  auditnpmdocs.Description,
		//	HelpName:     corecommondocs.CreateUsage("xr audit-npm", auditnpmdocs.Description, auditnpmdocs.Usage),
		//	ArgsUsage:    common.CreateEnvVars(),
		//	BashComplete: corecommondocs.CreateBashCompletionFunc(),
		//	Action:       auditNpmCmd,
		//},
		//{
		//	Name:         "scan",
		//	Flags:        cliutils.GetCommandFlags(cliutils.XrScan),
		//	Aliases:      []string{"s"},
		//	Description:  scandocs.Description,
		//	HelpName:     corecommondocs.CreateUsage("xr scan", scandocs.Description, scandocs.Usage),
		//	ArgsUsage:    common.CreateEnvVars(),
		//	BashComplete: corecommondocs.CreateBashCompletionFunc(),
		//	Action:       scanCmd,
		//},
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
	var typeRestriction = npmutils.All
	switch c.String("type-restriction") {
	case "devOnly":
		typeRestriction = npmutils.DevOnly
	case "prodOnly":
		typeRestriction = npmutils.ProdOnly
	}
	auditNpmCmd := audit.NewAuditNpmCommand().SetServerDetails(serverDetailes).SetNpmTypeRestriction(typeRestriction).SetTargetRepoPath(c.String("repo-path"))
	// Handle Xray's contex flags
	if c.String("watches") != "" {
		auditNpmCmd.SetWatches(strings.Split(c.String("watches"), ","))
	} else if c.String("project") != "" {
		auditNpmCmd.SetProject(c.String("project"))
	}
	auditNpmCmd.SetIncludeVulnerabilities(shouldIncludeVulnerabilities(c))
	return commands.Exec(auditNpmCmd)
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
		specFile, err = createDefaultScanSpec(c, c.String("repo-path"))
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
	scanCmd := audit.NewScanCommand().SetServerDetails(serverDetailes).SetThreads(threads).SetSpec(specFile).SetPrintResults(true)
	// Handle Xray's contex flags
	if c.String("watches") != "" {
		scanCmd.SetWatches(strings.Split(c.String("watches"), ","))
	} else if c.String("project") != "" {
		scanCmd.SetProject(c.String("project"))
	}
	scanCmd.SetIncludeVulnerabilities(shouldIncludeVulnerabilities(c))
	return commands.Exec(scanCmd)
}

func createDefaultScanSpec(c *cli.Context, defaultTarget string) (*spec.SpecFiles, error) {
	return spec.NewBuilder().
		Pattern(c.Args().Get(0)).
		Target(defaultTarget).
		Recursive(c.BoolT("recursive")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Regexp(c.Bool("regexp")).
		Ant(c.Bool("ant")).
		IncludeDirs(c.Bool("include-dirs")).
		BuildSpec(), nil
}

func CreateServerDetailsWithConfigOffer(c *cli.Context) (*coreconfig.ServerDetails, error) {
	return cliutils.CreateServerDetailsWithConfigOffer(c, true, "xr")
}

func shouldIncludeVulnerabilities(c *cli.Context) bool {
	// If no contex was provided by the user, no Violations will be triggered by Xray, so include general vulnerabilities in the command output
	return c.String("watches") == "" && c.String("project") == "" && c.String("repo-path") == ""
}

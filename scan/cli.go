package scan

import (
	"os"
	"strings"

	"github.com/jfrog/jfrog-cli/utils/progressbar"

	commandsutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	corecommondocs "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/audit/generic"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/scan"
	"github.com/jfrog/jfrog-cli/docs/common"
	auditdocs "github.com/jfrog/jfrog-cli/docs/scan/audit"
	auditgodocs "github.com/jfrog/jfrog-cli/docs/scan/auditgo"
	auditgradledocs "github.com/jfrog/jfrog-cli/docs/scan/auditgradle"
	"github.com/jfrog/jfrog-cli/docs/scan/auditmvn"
	auditnpmdocs "github.com/jfrog/jfrog-cli/docs/scan/auditnpm"
	auditpipdocs "github.com/jfrog/jfrog-cli/docs/scan/auditpip"
	auditpipenvdocs "github.com/jfrog/jfrog-cli/docs/scan/auditpipenv"
	buildscandocs "github.com/jfrog/jfrog-cli/docs/scan/buildscan"
	scandocs "github.com/jfrog/jfrog-cli/docs/scan/scan"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const auditScanCategory = "Audit & Scan"

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "audit",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.Audit),
			Aliases:      []string{"aud"},
			Usage:        auditdocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit", auditdocs.GetDescription(), auditdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       AuditCmd,
		},
		{
			Name:         "audit-mvn",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditMvn),
			Aliases:      []string{"am"},
			Usage:        auditmvn.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-mvn", auditmvn.GetDescription(), auditmvn.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return AuditSpecificCmd(c, coreutils.Maven)
			},
			Hidden: true,
		},
		{
			Name:         "audit-gradle",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditGradle),
			Aliases:      []string{"ag"},
			Usage:        auditgradledocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-gradle", auditgradledocs.GetDescription(), auditgradledocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return AuditSpecificCmd(c, coreutils.Gradle)
			},
			Hidden: true,
		},
		{
			Name:         "audit-npm",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditNpm),
			Aliases:      []string{"an"},
			Usage:        auditnpmdocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-npm", auditnpmdocs.GetDescription(), auditnpmdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return AuditSpecificCmd(c, coreutils.Npm)
			},
			Hidden: true,
		},
		{
			Name:         "audit-go",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditGo),
			Aliases:      []string{"ago"},
			Usage:        auditgodocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-go", auditgodocs.GetDescription(), auditgodocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return AuditSpecificCmd(c, coreutils.Go)
			},
			Hidden: true,
		},
		{
			Name:         "audit-pip",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditPip),
			Aliases:      []string{"ap"},
			Usage:        auditpipdocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-pip", auditpipdocs.GetDescription(), auditpipdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return AuditSpecificCmd(c, coreutils.Pip)
			},
			Hidden: true,
		},
		{
			Name:         "audit-pipenv",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditPip),
			Aliases:      []string{"ape"},
			Usage:        auditpipenvdocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-pipenv", auditpipenvdocs.GetDescription(), auditpipenvdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return AuditSpecificCmd(c, coreutils.Pipenv)
			},
			Hidden: true,
		},
		{
			Name:         "scan",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.XrScan),
			Aliases:      []string{"s"},
			Usage:        scandocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("scan", scandocs.GetDescription(), scandocs.Usage),
			UsageText:    scandocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       ScanCmd,
		},
		{
			Name:         "build-scan",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.BuildScan),
			Aliases:      []string{"bs"},
			Usage:        buildscandocs.GetDescription(),
			UsageText:    buildscandocs.GetArguments(),
			HelpName:     corecommondocs.CreateUsage("build-scan", buildscandocs.GetDescription(), buildscandocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       BuildScan,
		},
	})
}

func AuditCmd(c *cli.Context) error {
	auditCmd, err := createGenericAuditCmd(c)
	if err != nil {
		return err
	}

	// Check if user used specific technologies flags
	allTechnologies := coreutils.GetAllTechnologiesList()
	technologies := []string{}
	for _, tech := range allTechnologies {
		techExists := false
		switch tech {
		case coreutils.Maven:
			// On Maven we use '--mvn' flag
			techExists = c.Bool("mvn")
		default:
			techExists = c.Bool(tech.ToString())
		}
		if techExists {
			technologies = append(technologies, tech.ToString())
		}
	}
	auditCmd.SetTechnologies(technologies)
	return progressbar.ExecWithProgress(auditCmd)
}

func AuditSpecificCmd(c *cli.Context, technology coreutils.Technology) error {
	cliutils.LogNonGenericAuditCommandDeprecation(c.Command.Name)
	auditCmd, err := createGenericAuditCmd(c)
	if err != nil {
		return err
	}
	technologies := []string{string(technology)}
	auditCmd.SetTechnologies(technologies)
	return progressbar.ExecWithProgress(auditCmd)
}

func createGenericAuditCmd(c *cli.Context) (*audit.GenericAuditCommand, error) {
	auditCmd := audit.NewGenericAuditCommand()
	err := validateXrayContext(c)
	if err != nil {
		return nil, err
	}
	serverDetails, err := createServerDetailsWithConfigOffer(c)
	if err != nil {
		return nil, err
	}
	format, err := commandsutils.GetXrayOutputFormat(c.String("format"))
	if err != nil {
		return nil, err
	}

	auditCmd.SetServerDetails(serverDetails).
		SetOutputFormat(format).
		SetTargetRepoPath(addTrailingSlashToRepoPathIfNeeded(c)).
		SetProject(c.String("project")).
		SetIncludeVulnerabilities(shouldIncludeVulnerabilities(c)).
		SetIncludeLicenses(c.Bool("licenses")).
		SetFail(c.BoolT("fail")).
		SetPrintExtendedTable(c.Bool(cliutils.ExtendedTable))

	if c.String("watches") != "" {
		auditCmd.SetWatches(strings.Split(c.String("watches"), ","))
	}

	return auditCmd.SetExcludeTestDependencies(c.Bool(cliutils.ExcludeTestDeps)).
			SetUseWrapper(c.Bool(cliutils.UseWrapper)).
			SetInsecureTls(c.Bool(cliutils.InsecureTls)).
			SetNpmScope(c.String(cliutils.DepType)).
			SetPipRequirementsFile(c.String(cliutils.RequirementsFile)),
		err
}

func ScanCmd(c *cli.Context) error {
	if c.NArg() == 0 && !c.IsSet("spec") {
		return cliutils.PrintHelpAndReturnError("providing either a <source pattern> argument or the 'spec' option is mandatory", c)
	}
	err := validateXrayContext(c)
	if err != nil {
		return err
	}
	serverDetails, err := createServerDetailsWithConfigOffer(c)
	if err != nil {
		return err
	}
	var specFile *spec.SpecFiles
	if c.IsSet("spec") {
		specFile, err = cliutils.GetFileSystemSpec(c)
	} else {
		specFile, err = createDefaultScanSpec(c, addTrailingSlashToRepoPathIfNeeded(c))
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(specFile.Files, false, false)
	if err != nil {
		return err
	}
	threads, err := cliutils.GetThreadsCount(c)
	if err != nil {
		return err
	}
	format, err := commandsutils.GetXrayOutputFormat(c.String("format"))
	if err != nil {
		return err
	}
	cliutils.FixWinPathsForFileSystemSourcedCmds(specFile, c)
	scanCmd := scan.NewScanCommand().SetServerDetails(serverDetails).SetThreads(threads).SetSpec(specFile).SetOutputFormat(format).
		SetProject(c.String("project")).SetIncludeVulnerabilities(shouldIncludeVulnerabilities(c)).
		SetIncludeLicenses(c.Bool("licenses")).SetFail(c.BoolT("fail")).SetPrintExtendedTable(c.Bool(cliutils.ExtendedTable))
	if c.String("watches") != "" {
		scanCmd.SetWatches(strings.Split(c.String("watches"), ","))
	}
	return commands.Exec(scanCmd)
}

// Scan published builds with Xray
func BuildScan(c *cli.Context) error {
	if c.NArg() > 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	buildConfiguration := cliutils.CreateBuildConfiguration(c)
	if err := buildConfiguration.ValidateBuildParams(); err != nil {
		return err
	}

	serverDetails, err := createServerDetailsWithConfigOffer(c)
	if err != nil {
		return err
	}
	err = validateXrayContext(c)
	if err != nil {
		return err
	}
	format, err := commandsutils.GetXrayOutputFormat(c.String("format"))
	if err != nil {
		return err
	}
	buildScanCmd := scan.NewBuildScanCommand().
		SetServerDetails(serverDetails).
		SetFailBuild(c.BoolT("fail")).
		SetBuildConfiguration(buildConfiguration).
		SetIncludeVulnerabilities(c.Bool("vuln")).
		SetOutputFormat(format).
		SetPrintExtendedTable(c.Bool(cliutils.ExtendedTable)).
		SetRescan(c.Bool("rescan"))
	return commands.Exec(buildScanCmd)
}

func DockerScan(c *cli.Context, image string) error {
	if show, err := cliutils.ShowGenericCmdHelpIfNeeded(c, c.Args(), "dockerscanhelp"); show || err != nil {
		return err
	}
	if image == "" {
		return cli.ShowCommandHelp(c, "dockerscanhelp")
	}
	serverDetails, err := createServerDetailsWithConfigOffer(c)
	if err != nil {
		return err
	}
	containerScanCommand := scan.NewDockerScanCommand()
	format, err := commandsutils.GetXrayOutputFormat(c.String("format"))
	if err != nil {
		return err
	}
	containerScanCommand.SetServerDetails(serverDetails).SetOutputFormat(format).SetProject(c.String("project")).
		SetIncludeVulnerabilities(shouldIncludeVulnerabilities(c)).SetIncludeLicenses(c.Bool("licenses")).
		SetFail(c.BoolT("fail")).SetPrintExtendedTable(c.Bool(cliutils.ExtendedTable))
	if c.String("watches") != "" {
		containerScanCommand.SetWatches(strings.Split(c.String("watches"), ","))
	}
	containerScanCommand.SetImageTag(c.Args().Get(1))
	return progressbar.ExecWithProgress(containerScanCommand)
}

func addTrailingSlashToRepoPathIfNeeded(c *cli.Context) string {
	repoPath := c.String("repo-path")
	if repoPath != "" && !strings.Contains(repoPath, "/") {
		// In case a only repo name was provided (no path) we are adding a trailing slash.
		repoPath += "/"
	}
	return repoPath
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

func createServerDetailsWithConfigOffer(c *cli.Context) (*coreconfig.ServerDetails, error) {
	return cliutils.CreateServerDetailsWithConfigOffer(c, true, "xr")
}

func shouldIncludeVulnerabilities(c *cli.Context) bool {
	// If no context was provided by the user, no Violations will be triggered by Xray, so include general vulnerabilities in the command output
	return c.String("watches") == "" && !isProjectProvided(c) && c.String("repo-path") == ""
}

func validateXrayContext(c *cli.Context) error {
	contextFlag := 0
	if c.String("watches") != "" {
		contextFlag++
	}
	if isProjectProvided(c) {
		contextFlag++
	}
	if c.String("repo-path") != "" {
		contextFlag++
	}
	if contextFlag > 1 {
		return errorutils.CheckErrorf("only one of the following flags can be supplied: --watches, --project or --repo-path")
	}
	return nil
}

func isProjectProvided(c *cli.Context) bool {
	return c.String("project") != "" || os.Getenv(coreutils.Project) != ""
}

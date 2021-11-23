package scan

import (
	"github.com/codegangsta/cli"
	commandsutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	rtutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	containerutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	corecommondocs "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	npmutils "github.com/jfrog/jfrog-cli-core/v2/utils/npm"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/audit"
	scancommands "github.com/jfrog/jfrog-cli-core/v2/xray/commands/scan"
	"github.com/jfrog/jfrog-cli/docs/common"
	auditgodocs "github.com/jfrog/jfrog-cli/docs/scan/auditgo"
	auditgradledocs "github.com/jfrog/jfrog-cli/docs/scan/auditgradle"
	"github.com/jfrog/jfrog-cli/docs/scan/auditmvn"
	auditnpmdocs "github.com/jfrog/jfrog-cli/docs/scan/auditnpm"
	auditpipdocs "github.com/jfrog/jfrog-cli/docs/scan/auditpip"
	scandocs "github.com/jfrog/jfrog-cli/docs/scan/scan"
	buildscandocs "github.com/jfrog/jfrog-cli/docs/xray/buildscan"
	"github.com/jfrog/jfrog-cli/docs/xray/dockerscan"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"strings"
)

const auditScanCategory = "Audit & Scan"

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "audit-mvn",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditMvn),
			Aliases:      []string{"am"},
			Description:  auditmvn.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-mvn", auditmvn.GetDescription(), auditmvn.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       AuditMvnCmd,
		},
		{
			Name:         "audit-gradle",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditGradle),
			Aliases:      []string{"ag"},
			Description:  auditgradledocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-gradle", auditgradledocs.GetDescription(), auditgradledocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       AuditGradleCmd,
		},
		{
			Name:         "audit-npm",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditNpm),
			Aliases:      []string{"an"},
			Description:  auditnpmdocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-npm", auditnpmdocs.GetDescription(), auditnpmdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       AuditNpmCmd,
		},
		{
			Name:         "audit-go",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditGo),
			Aliases:      []string{"ago"},
			Description:  auditgodocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-go", auditgodocs.GetDescription(), auditgodocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       AuditGoCmd,
		},
		{
			Name:         "audit-pip",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.AuditPip),
			Aliases:      []string{"ap"},
			Description:  auditpipdocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("audit-pip", auditpipdocs.GetDescription(), auditpipdocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       AuditPipCmd,
		},
		{
			Name:         "scan",
			Category:     auditScanCategory,
			Flags:        cliutils.GetCommandFlags(cliutils.XrScan),
			Aliases:      []string{"s"},
			Description:  scandocs.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("scan", scandocs.GetDescription(), scandocs.Usage),
			UsageText:    scandocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       ScanCmd,
		},
		{
			Name:         "build-scan",
			Flags:        cliutils.GetCommandFlags(cliutils.XrBuildScan),
			Aliases:      []string{"bs"},
			Description:  buildscandocs.GetDescription(),
			UsageText:    buildscandocs.GetArguments(),
			HelpName:     corecommondocs.CreateUsage("build-scan", buildscandocs.GetDescription(), buildscandocs.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       BuildScan,
		},
		{
			Name:         "docker",
			Flags:        cliutils.GetCommandFlags(cliutils.DockerScan),
			Aliases:      []string{"docker"},
			Description:  dockerscan.GetDescription(),
			HelpName:     corecommondocs.CreateUsage("docker scan", dockerscan.GetDescription(), dockerscan.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommondocs.CreateBashCompletionFunc(),
			Action:       DockerScan,
		},
	})
}

func AuditMvnCmd(c *cli.Context) error {
	genericAuditCmd, err := createGenericAuditCmd(c)
	if err != nil {
		return err
	}
	xrAuditMvnCmd := audit.NewAuditMavenCommand(*genericAuditCmd).SetInsecureTls(c.Bool(cliutils.InsecureTls))
	return commands.Exec(xrAuditMvnCmd)
}

func AuditGradleCmd(c *cli.Context) error {
	genericAuditCmd, err := createGenericAuditCmd(c)
	if err != nil {
		return err
	}
	xrAuditGradleCmd := audit.NewAuditGradleCommand(*genericAuditCmd).SetExcludeTestDeps(c.Bool(cliutils.ExcludeTestDeps)).SetUseWrapper(c.Bool(cliutils.UseWrapper))
	return commands.Exec(xrAuditGradleCmd)
}

func AuditNpmCmd(c *cli.Context) error {
	genericAuditCmd, err := createGenericAuditCmd(c)
	if err != nil {
		return err
	}
	var typeRestriction = npmutils.All
	switch c.String("dep-type") {
	case "devOnly":
		typeRestriction = npmutils.DevOnly
	case "prodOnly":
		typeRestriction = npmutils.ProdOnly
	}
	auditNpmCmd := audit.NewAuditNpmCommand(*genericAuditCmd).SetNpmTypeRestriction(typeRestriction)
	return commands.Exec(auditNpmCmd)
}

func AuditGoCmd(c *cli.Context) error {
	genericAuditCmd, err := createGenericAuditCmd(c)
	if err != nil {
		return err
	}
	auditGoCmd := audit.NewAuditGoCommand(*genericAuditCmd)
	return commands.Exec(auditGoCmd)
}

func AuditPipCmd(c *cli.Context) error {
	genericAuditCmd, err := createGenericAuditCmd(c)
	if err != nil {
		return err
	}
	auditPipCmd := audit.NewAuditPipCommand(*genericAuditCmd)
	return commands.Exec(auditPipCmd)
}

func createGenericAuditCmd(c *cli.Context) (*audit.AuditCommand, error) {
	auditCmd := audit.NewAuditCommand()
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
		SetIncludeLicenses(c.Bool("licenses"))

	if c.String("watches") != "" {
		auditCmd.SetWatches(strings.Split(c.String("watches"), ","))
	}
	return auditCmd, err
}

func ScanCmd(c *cli.Context) error {
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
	err = spec.ValidateSpec(specFile.Files, false, false, false)
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
	scanCmd := scancommands.NewScanCommand().SetServerDetails(serverDetails).SetThreads(threads).SetSpec(specFile).SetOutputFormat(format).
		SetProject(c.String("project")).
		SetIncludeVulnerabilities(shouldIncludeVulnerabilities(c)).SetIncludeLicenses(c.Bool("licenses"))
	if c.String("watches") != "" {
		scanCmd.SetWatches(strings.Split(c.String("watches"), ","))
	}
	return commands.Exec(scanCmd)
}

// New Build-Scan command that works directly with Xray
func BuildScan(c *cli.Context) error {
	if c.NArg() > 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	buildConfiguration := createBuildConfiguration(c)
	if err := validateBuildConfiguration(c, buildConfiguration); err != nil {
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
	buildScanCmd := scancommands.NewBuildScanCommand().SetServerDetails(serverDetails).SetFailBuild(c.BoolT("fail")).SetBuildConfiguration(buildConfiguration).
		SetIncludeVulnerabilities(c.Bool("vulnerabilities")).SetOutputFormat(format)
	return commands.Exec(buildScanCmd)
}

func DockerScan(c *cli.Context) error {
	args := cliutils.ExtractCommand(c)
	cmdName := ""
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			cmdName = arg
			break
		}
	}
	switch cmdName {
	// Aliases accepted by npm.
	case "scan":
		return containerScan(c, containerutils.DockerClient)
	default:
		return errorutils.CheckErrorf("'jf docker %s' command is currently not supported by JFrog CLI", cmdName)
	}
}

func containerScan(c *cli.Context, containerManagerType containerutils.ContainerManagerType) error {
	if c.NArg() != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	serverDetails, err := createServerDetailsWithConfigOffer(c)
	if err != nil {
		return err
	}
	format, err := commandsutils.GetXrayOutputFormat(c.String("format"))
	if err != nil {
		return err
	}
	containerScanCommand := scancommands.NewEmptyContainerScanCommand()
	containerScanCommand.SetServerDetails(serverDetails).SetOutputFormat(format).SetProject(c.String("project")).
		SetIncludeVulnerabilities(shouldIncludeVulnerabilities(c)).SetIncludeLicenses(c.Bool("licenses"))
	if c.String("watches") != "" {
		containerScanCommand.SetWatches(strings.Split(c.String("watches"), ","))
	}
	containerScanCommand.SetImageTag(c.Args().Get(1)).SetContainerManagerType(containerManagerType)
	return commands.Exec(containerScanCommand)
}

// Returns build configuration struct using the params provided from the console.
func createBuildConfiguration(c *cli.Context) *rtutils.BuildConfiguration {
	buildConfiguration := new(rtutils.BuildConfiguration)
	buildNameArg, buildNumberArg := c.Args().Get(0), c.Args().Get(1)
	if buildNameArg == "" || buildNumberArg == "" {
		buildNameArg = ""
		buildNumberArg = ""
	}
	buildConfiguration.BuildName, buildConfiguration.BuildNumber = rtutils.GetBuildNameAndNumber(buildNameArg, buildNumberArg)
	buildConfiguration.Project = rtutils.GetBuildProject(c.String("project"))
	return buildConfiguration
}

func validateBuildConfiguration(c *cli.Context, buildConfiguration *rtutils.BuildConfiguration) error {
	if buildConfiguration.BuildName == "" || buildConfiguration.BuildNumber == "" {
		return cliutils.PrintHelpAndReturnError("Build name and build number are expected as command arguments or environment variables.", c)
	}
	return nil
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
	return c.String("watches") == "" && c.String("project") == "" && c.String("repo-path") == ""
}

func validateXrayContext(c *cli.Context) error {
	contextFlag := 0
	if c.String("watches") != "" {
		contextFlag++
	}
	if c.String("project") != "" {
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

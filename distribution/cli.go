package distribution

import (
	"errors"
	distributionCommands "github.com/jfrog/jfrog-cli-artifactory/distribution/commands"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	coreCommonDocs "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundlecreate"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundledelete"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundledistribute"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundlesign"
	"github.com/jfrog/jfrog-cli/docs/artifactory/releasebundleupdate"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/distribution"
	distributionServices "github.com/jfrog/jfrog-client-go/distribution/services"
	distributionServicesUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"strings"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "release-bundle-create",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleV1Create),
			Aliases:      []string{"rbc"},
			Usage:        releasebundlecreate.GetDescription(),
			HelpName:     coreCommonDocs.CreateUsage("ds rbc", releasebundlecreate.GetDescription(), releasebundlecreate.Usage),
			UsageText:    releasebundlecreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: coreCommonDocs.CreateBashCompletionFunc(),
			Action:       releaseBundleCreateCmd,
		},
		{
			Name:         "release-bundle-update",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleV1Update),
			Aliases:      []string{"rbu"},
			Usage:        releasebundleupdate.GetDescription(),
			HelpName:     coreCommonDocs.CreateUsage("ds rbu", releasebundleupdate.GetDescription(), releasebundleupdate.Usage),
			UsageText:    releasebundleupdate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: coreCommonDocs.CreateBashCompletionFunc(),
			Action:       releaseBundleUpdateCmd,
		},
		{
			Name:         "release-bundle-sign",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleV1Sign),
			Aliases:      []string{"rbs"},
			Usage:        releasebundlesign.GetDescription(),
			HelpName:     coreCommonDocs.CreateUsage("ds rbs", releasebundlesign.GetDescription(), releasebundlesign.Usage),
			UsageText:    releasebundlesign.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: coreCommonDocs.CreateBashCompletionFunc(),
			Action:       releaseBundleSignCmd,
		},
		{
			Name:         "release-bundle-distribute",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleV1Distribute),
			Aliases:      []string{"rbd"},
			Usage:        releasebundledistribute.GetDescription(),
			HelpName:     coreCommonDocs.CreateUsage("ds rbd", releasebundledistribute.GetDescription(), releasebundledistribute.Usage),
			UsageText:    releasebundledistribute.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: coreCommonDocs.CreateBashCompletionFunc(),
			Action:       releaseBundleDistributeCmd,
		},
		{
			Name:         "release-bundle-delete",
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleV1Delete),
			Aliases:      []string{"rbdel"},
			Usage:        releasebundledelete.GetDescription(),
			HelpName:     coreCommonDocs.CreateUsage("ds rbdel", releasebundledelete.GetDescription(), releasebundledelete.Usage),
			UsageText:    releasebundledelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: coreCommonDocs.CreateBashCompletionFunc(),
			Action:       releaseBundleDeleteCmd,
		},
	})
}

func releaseBundleCreateCmd(c *cli.Context) error {
	if !(c.NArg() == 2 && c.IsSet("spec") || (c.NArg() == 3 && !c.IsSet("spec"))) {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	if c.IsSet("detailed-summary") && !c.IsSet("sign") {
		return cliutils.PrintHelpAndReturnError("The --detailed-summary option can't be used without --sign", c)
	}
	var releaseBundleCreateSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		releaseBundleCreateSpec, err = cliutils.GetSpec(c, true, true)
	} else {
		releaseBundleCreateSpec = createDefaultReleaseBundleSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(releaseBundleCreateSpec.Files, false, true)
	if err != nil {
		return err
	}

	params, err := createReleaseBundleCreateUpdateParams(c, c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}
	releaseBundleCreateCmd := distributionCommands.NewReleaseBundleCreateCommand()
	dsDetails, err := createDistributionDetailsByFlags(c)
	if err != nil {
		return err
	}
	releaseBundleCreateCmd.SetServerDetails(dsDetails).SetReleaseBundleCreateParams(params).SetSpec(releaseBundleCreateSpec).SetDryRun(c.Bool("dry-run")).SetDetailedSummary(c.Bool("detailed-summary"))

	err = commands.Exec(releaseBundleCreateCmd)
	if releaseBundleCreateCmd.IsDetailedSummary() {
		if summary := releaseBundleCreateCmd.GetSummary(); summary != nil {
			return cliutils.PrintBuildInfoSummaryReport(summary.IsSucceeded(), summary.GetSha256(), err)
		}
	}
	return err
}

func releaseBundleUpdateCmd(c *cli.Context) error {
	if !(c.NArg() == 2 && c.IsSet("spec") || (c.NArg() == 3 && !c.IsSet("spec"))) {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	if c.IsSet("detailed-summary") && !c.IsSet("sign") {
		return cliutils.PrintHelpAndReturnError("The --detailed-summary option can't be used without --sign", c)
	}
	var releaseBundleUpdateSpec *spec.SpecFiles
	var err error
	if c.IsSet("spec") {
		releaseBundleUpdateSpec, err = cliutils.GetSpec(c, true, true)
	} else {
		releaseBundleUpdateSpec = createDefaultReleaseBundleSpec(c)
	}
	if err != nil {
		return err
	}
	err = spec.ValidateSpec(releaseBundleUpdateSpec.Files, false, true)
	if err != nil {
		return err
	}

	params, err := createReleaseBundleCreateUpdateParams(c, c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}
	releaseBundleUpdateCmd := distributionCommands.NewReleaseBundleUpdateCommand()
	dsDetails, err := createDistributionDetailsByFlags(c)
	if err != nil {
		return err
	}
	releaseBundleUpdateCmd.SetServerDetails(dsDetails).SetReleaseBundleUpdateParams(params).SetSpec(releaseBundleUpdateSpec).SetDryRun(c.Bool("dry-run")).SetDetailedSummary(c.Bool("detailed-summary"))

	err = commands.Exec(releaseBundleUpdateCmd)
	if releaseBundleUpdateCmd.IsDetailedSummary() {
		if summary := releaseBundleUpdateCmd.GetSummary(); summary != nil {
			return cliutils.PrintBuildInfoSummaryReport(summary.IsSucceeded(), summary.GetSha256(), err)
		}
	}
	return err
}

func releaseBundleSignCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	params := distributionServices.NewSignBundleParams(c.Args().Get(0), c.Args().Get(1))
	params.StoringRepository = c.String("repo")
	params.GpgPassphrase = c.String("passphrase")
	releaseBundleSignCmd := distributionCommands.NewReleaseBundleSignCommand()
	dsDetails, err := createDistributionDetailsByFlags(c)
	if err != nil {
		return err
	}
	releaseBundleSignCmd.SetServerDetails(dsDetails).SetReleaseBundleSignParams(params).SetDetailedSummary(c.Bool("detailed-summary"))
	err = commands.Exec(releaseBundleSignCmd)
	if releaseBundleSignCmd.IsDetailedSummary() {
		if summary := releaseBundleSignCmd.GetSummary(); summary != nil {
			return cliutils.PrintBuildInfoSummaryReport(summary.IsSucceeded(), summary.GetSha256(), err)
		}
	}
	return err
}

func releaseBundleDistributeCmd(c *cli.Context) error {
	if err := distribution.ValidateReleaseBundleDistributeCmd(c); err != nil {
		return err
	}

	dsDetails, err := createDistributionDetailsByFlags(c)
	if err != nil {
		return err
	}
	distributionRules, maxWaitMinutes, params, err := distribution.InitReleaseBundleDistributeCmd(c)
	if err != nil {
		return err
	}

	distributeCmd := distributionCommands.NewReleaseBundleDistributeV1Command()
	distributeCmd.SetServerDetails(dsDetails).
		SetDistributeBundleParams(params).
		SetDistributionRules(distributionRules).
		SetDryRun(c.Bool("dry-run")).
		SetSync(c.Bool("sync")).
		SetMaxWaitMinutes(maxWaitMinutes).
		SetAutoCreateRepo(c.Bool("create-repo"))

	return commands.Exec(distributeCmd)
}

func releaseBundleDeleteCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	var distributionRules *spec.DistributionRules
	if c.IsSet("dist-rules") {
		if c.IsSet("site") || c.IsSet("city") || c.IsSet("country-code") {
			return cliutils.PrintHelpAndReturnError("flag --dist-rules can't be used with --site, --city or --country-code", c)
		}
		var err error
		distributionRules, err = spec.CreateDistributionRulesFromFile(c.String("dist-rules"))
		if err != nil {
			return err
		}
	} else {
		distributionRules = distribution.CreateDefaultDistributionRules(c)
	}

	params := distributionServices.NewDeleteReleaseBundleParams(c.Args().Get(0), c.Args().Get(1))
	params.DeleteFromDistribution = c.BoolT("delete-from-dist")
	params.Sync = c.Bool("sync")
	maxWaitMinutes, err := cliutils.GetIntFlagValue(c, "max-wait-minutes", 60)
	if err != nil {
		return err
	}
	params.MaxWaitMinutes = maxWaitMinutes
	distributeBundleCmd := distributionCommands.NewReleaseBundleDeleteParams()
	dsDetails, err := createDistributionDetailsByFlags(c)
	if err != nil {
		return err
	}
	distributeBundleCmd.SetQuiet(cliutils.GetQuietValue(c)).SetServerDetails(dsDetails).SetDistributeBundleParams(params).SetDistributionRules(distributionRules).SetDryRun(c.Bool("dry-run"))

	return commands.Exec(distributeBundleCmd)
}

func createDefaultReleaseBundleSpec(c *cli.Context) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(c.Args().Get(2)).
		Target(c.String("target")).
		Props(c.String("props")).
		Build(c.String("build")).
		Bundle(c.String("bundle")).
		Exclusions(cliutils.GetStringsArrFlagValue(c, "exclusions")).
		Regexp(c.Bool("regexp")).
		TargetProps(c.String("target-props")).
		Ant(c.Bool("ant")).
		BuildSpec()
}

func createReleaseBundleCreateUpdateParams(c *cli.Context, bundleName, bundleVersion string) (distributionServicesUtils.ReleaseBundleParams, error) {
	releaseBundleParams := distributionServicesUtils.NewReleaseBundleParams(bundleName, bundleVersion)
	releaseBundleParams.SignImmediately = c.Bool("sign")
	releaseBundleParams.StoringRepository = c.String("repo")
	releaseBundleParams.GpgPassphrase = c.String("passphrase")
	releaseBundleParams.Description = c.String("desc")
	if c.IsSet("release-notes-path") {
		bytes, err := os.ReadFile(c.String("release-notes-path"))
		if err != nil {
			return releaseBundleParams, errorutils.CheckError(err)
		}
		releaseBundleParams.ReleaseNotes = string(bytes)
		releaseBundleParams.ReleaseNotesSyntax, err = populateReleaseNotesSyntax(c)
		if err != nil {
			return releaseBundleParams, err
		}
	}
	return releaseBundleParams, nil
}

func populateReleaseNotesSyntax(c *cli.Context) (distributionServicesUtils.ReleaseNotesSyntax, error) {
	// If release notes syntax is set, use it
	releaseNotesSyntax := c.String("release-notes-syntax")
	if releaseNotesSyntax != "" {
		switch releaseNotesSyntax {
		case "markdown":
			return distributionServicesUtils.Markdown, nil
		case "asciidoc":
			return distributionServicesUtils.Asciidoc, nil
		case "plain_text":
			return distributionServicesUtils.PlainText, nil
		default:
			return distributionServicesUtils.PlainText, errorutils.CheckErrorf("--release-notes-syntax must be one of: markdown, asciidoc or plain_text.")
		}
	}
	// If the file extension is ".md" or ".markdown", use the Markdown syntax
	extension := strings.ToLower(filepath.Ext(c.String("release-notes-path")))
	if extension == ".md" || extension == ".markdown" {
		return distributionServicesUtils.Markdown, nil
	}
	return distributionServicesUtils.PlainText, nil
}

func createDistributionDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	dsDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Ds)
	if err != nil {
		return nil, err
	}
	if dsDetails.DistributionUrl == "" {
		return nil, errors.New("no JFrog Distribution URL specified, either via the --url flag or as part of the server configuration")
	}
	return dsDetails, nil
}

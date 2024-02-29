package lifecycle

import (
	"errors"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	coreCommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/lifecycle"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/docs/common"
	rbCreate "github.com/jfrog/jfrog-cli/docs/lifecycle/create"
	rbDistribute "github.com/jfrog/jfrog-cli/docs/lifecycle/distribute"
	rbPromote "github.com/jfrog/jfrog-cli/docs/lifecycle/promote"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/distribution"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/urfave/cli"
	"strings"
)

const lcCategory = "Lifecycle"

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "release-bundle-create",
			Aliases:      []string{"rbc"},
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleCreate),
			Usage:        rbCreate.GetDescription(),
			HelpName:     coreCommon.CreateUsage("release-bundle-create", rbCreate.GetDescription(), rbCreate.Usage),
			UsageText:    rbCreate.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: coreCommon.CreateBashCompletionFunc(),
			Category:     lcCategory,
			Action:       create,
		},
		{
			Name:         "release-bundle-promote",
			Aliases:      []string{"rbp"},
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundlePromote),
			Usage:        rbPromote.GetDescription(),
			HelpName:     coreCommon.CreateUsage("release-bundle-promote", rbPromote.GetDescription(), rbPromote.Usage),
			UsageText:    rbPromote.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: coreCommon.CreateBashCompletionFunc(),
			Category:     lcCategory,
			Action:       promote,
		},
		{
			Name:         "release-bundle-distribute",
			Aliases:      []string{"rbd"},
			Flags:        cliutils.GetCommandFlags(cliutils.ReleaseBundleDistribute),
			Usage:        rbDistribute.GetDescription(),
			HelpName:     coreCommon.CreateUsage("rbd", rbDistribute.GetDescription(), rbDistribute.Usage),
			UsageText:    rbDistribute.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: coreCommon.CreateBashCompletionFunc(),
			Category:     lcCategory,
			Action:       distribute,
		},
	})
}

func validateCreateReleaseBundleContext(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	if err := assertSigningKeyProvided(c); err != nil {
		return err
	}

	return assertValidCreationMethod(c)
}

func assertValidCreationMethod(c *cli.Context) error {
	methods := []bool{
		c.IsSet("spec"), c.IsSet(cliutils.Builds), c.IsSet(cliutils.ReleaseBundles)}
	if coreutils.SumTrueValues(methods) > 1 {
		return errorutils.CheckErrorf("exactly one creation source must be supplied: --%s, --%s or --%s.\n"+
			"The spec option is the recommended approach.", "spec", cliutils.Builds, cliutils.ReleaseBundles)
	}
	// If the user did not provide a source, we suggest only the recommended spec approach.
	if coreutils.SumTrueValues(methods) == 0 {
		return errorutils.CheckErrorf("the --spec option is mandatory")
	}
	return nil
}

func create(c *cli.Context) (err error) {
	if err = validateCreateReleaseBundleContext(c); err != nil {
		return err
	}

	var creationSpec *spec.SpecFiles
	if c.IsSet("spec") {
		creationSpec, err = cliutils.GetSpec(c, true)
		if err != nil {
			return
		}
	}

	lcDetails, err := createLifecycleDetailsByFlags(c)
	if err != nil {
		return
	}

	createCmd := lifecycle.NewReleaseBundleCreateCommand().SetServerDetails(lcDetails).SetReleaseBundleName(c.Args().Get(0)).
		SetReleaseBundleVersion(c.Args().Get(1)).SetSigningKeyName(c.String(cliutils.SigningKey)).SetSync(c.Bool(cliutils.Sync)).
		SetReleaseBundleProject(cliutils.GetProject(c)).SetSpec(creationSpec).
		SetBuildsSpecPath(c.String(cliutils.Builds)).SetReleaseBundlesSpecPath(c.String(cliutils.ReleaseBundles))
	return commands.Exec(createCmd)
}

func promote(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	if c.NArg() != 3 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	if err := assertSigningKeyProvided(c); err != nil {
		return err
	}

	lcDetails, err := createLifecycleDetailsByFlags(c)
	if err != nil {
		return err
	}

	createCmd := lifecycle.NewReleaseBundlePromoteCommand().SetServerDetails(lcDetails).SetReleaseBundleName(c.Args().Get(0)).
		SetReleaseBundleVersion(c.Args().Get(1)).SetEnvironment(c.Args().Get(2)).SetSigningKeyName(c.String(cliutils.SigningKey)).
		SetSync(c.Bool(cliutils.Sync)).SetReleaseBundleProject(cliutils.GetProject(c)).
		SetIncludeReposPatterns(splitRepos(c, cliutils.IncludeRepos)).SetExcludeReposPatterns(splitRepos(c, cliutils.ExcludeRepos))
	return commands.Exec(createCmd)
}

func distribute(c *cli.Context) error {
	if err := validateDistributeCommand(c); err != nil {
		return err
	}

	lcDetails, err := createLifecycleDetailsByFlags(c)
	if err != nil {
		return err
	}
	distributionRules, _, params, err := distribution.InitReleaseBundleDistributeCmd(c)
	if err != nil {
		return err
	}

	distributeCmd := lifecycle.NewReleaseBundleDistributeCommand()
	distributeCmd.SetServerDetails(lcDetails).
		SetDistributeBundleParams(params).
		SetDistributionRules(distributionRules).
		SetDryRun(c.Bool("dry-run")).
		SetAutoCreateRepo(c.Bool(cliutils.CreateRepo)).
		SetPathMappingPattern(c.String(cliutils.PathMappingPattern)).
		SetPathMappingTarget(c.String(cliutils.PathMappingTarget))
	return commands.Exec(distributeCmd)
}

func validateDistributeCommand(c *cli.Context) error {
	if err := distribution.ValidateReleaseBundleDistributeCmd(c); err != nil {
		return err
	}

	mappingPatternProvided := c.IsSet(cliutils.PathMappingPattern)
	mappingTargetProvided := c.IsSet(cliutils.PathMappingTarget)
	if (mappingPatternProvided && !mappingTargetProvided) ||
		(!mappingPatternProvided && mappingTargetProvided) {
		return errorutils.CheckErrorf("the options --%s and --%s must be provided together", cliutils.PathMappingPattern, cliutils.PathMappingTarget)
	}
	return nil
}

func assertSigningKeyProvided(c *cli.Context) error {
	if c.String(cliutils.SigningKey) == "" {
		return errorutils.CheckErrorf("the --%s option is mandatory", cliutils.SigningKey)
	}
	return nil
}

func createLifecycleDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	lcDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return nil, err
	}
	if lcDetails.Url == "" {
		return nil, errors.New("platform URL is mandatory for lifecycle commands")
	}
	PlatformToLifecycleUrls(lcDetails)
	return lcDetails, nil
}

func PlatformToLifecycleUrls(lcDetails *coreConfig.ServerDetails) {
	lcDetails.ArtifactoryUrl = utils.AddTrailingSlashIfNeeded(lcDetails.Url) + "artifactory/"
	lcDetails.LifecycleUrl = utils.AddTrailingSlashIfNeeded(lcDetails.Url) + "lifecycle/"
	lcDetails.Url = ""
}

func splitRepos(c *cli.Context, reposOptionKey string) []string {
	if c.IsSet(reposOptionKey) {
		return strings.Split(c.String(reposOptionKey), ";")
	}
	return nil
}

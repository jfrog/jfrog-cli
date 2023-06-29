package lifecycle

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreCommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/lifecycle"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/docs/common"
	rbCreate "github.com/jfrog/jfrog-cli/docs/lifecycle/create"
	rbPromote "github.com/jfrog/jfrog-cli/docs/lifecycle/promote"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/urfave/cli"
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
			Action: func(c *cli.Context) error {
				return create(c)
			},
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
			Action: func(c *cli.Context) error {
				return promote(c)
			},
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

	buildsSourceProvided := c.String(cliutils.Builds) != ""
	releaseBundlesSourceProvided := c.String(cliutils.ReleaseBundles) != ""
	if (buildsSourceProvided && releaseBundlesSourceProvided) ||
		!(buildsSourceProvided || releaseBundlesSourceProvided) {
		return errorutils.CheckErrorf("exactly one of the following options must be supplied: --%s or --%s", cliutils.Builds, cliutils.ReleaseBundles)
	}
	return nil
}

func create(c *cli.Context) (err error) {
	if err = validateCreateReleaseBundleContext(c); err != nil {
		return err
	}

	lcDetails, err := createLifecycleDetailsByFlags(c)
	if err != nil {
		return
	}

	createCmd := lifecycle.NewReleaseBundleCreate().SetServerDetails(lcDetails).SetReleaseBundleName(c.Args().Get(0)).
		SetReleaseBundleVersion(c.Args().Get(1)).SetSigningKeyName(c.String(cliutils.SigningKey)).SetAsync(c.BoolT(cliutils.Async)).
		SetReleaseBundleProject(cliutils.GetProject(c)).SetBuildsSpecPath(c.String(cliutils.Builds)).
		SetReleaseBundlesSpecPath(c.String(cliutils.ReleaseBundles))
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

	createCmd := lifecycle.NewReleaseBundlePromote().SetServerDetails(lcDetails).SetReleaseBundleName(c.Args().Get(0)).
		SetReleaseBundleVersion(c.Args().Get(1)).SetEnvironment(c.Args().Get(2)).SetSigningKeyName(c.String(cliutils.SigningKey)).
		SetAsync(c.BoolT(cliutils.Async)).SetReleaseBundleProject(cliutils.GetProject(c)).SetOverwrite(c.Bool(cliutils.Overwrite))
	return commands.Exec(createCmd)
}

func assertSigningKeyProvided(c *cli.Context) error {
	if c.String(cliutils.SigningKey) == "" {
		return errorutils.CheckErrorf("the --%s option is mandatory", cliutils.SigningKey)
	}
	return nil
}

func createLifecycleDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	lcDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, false, cliutils.Platform)
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

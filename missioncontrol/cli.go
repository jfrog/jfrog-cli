package missioncontrol

import (
	"strconv"

	coreCommonCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/missioncontrol/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/jpdadd"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/jpddelete"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/licenseacquire"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/licensedeploy"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/licenserelease"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "license-acquire",
			Flags:        cliutils.GetCommandFlags(cliutils.LicenseAcquire),
			Usage:        licenseacquire.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc license-acquire", licenseacquire.GetDescription(), licenseacquire.Usage),
			UsageText:    licenseacquire.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"la"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return licenseAcquire(c)
			},
		},
		{
			Name:         "license-deploy",
			Flags:        cliutils.GetCommandFlags(cliutils.LicenseDeploy),
			Usage:        licensedeploy.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc license-deploy", licensedeploy.GetDescription(), licensedeploy.Usage),
			UsageText:    licensedeploy.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"ld"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return licenseDeploy(c)
			},
		},
		{
			Name:         "license-release",
			Flags:        cliutils.GetCommandFlags(cliutils.LicenseRelease),
			Usage:        licenserelease.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc license-release", licenserelease.GetDescription(), licenserelease.Usage),
			UsageText:    licenserelease.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"lr"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return licenseRelease(c)
			},
		},
		{
			Name:         "jpd-add",
			Flags:        cliutils.GetCommandFlags(cliutils.JpdAdd),
			Usage:        jpdadd.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc jpd-add", jpdadd.GetDescription(), jpdadd.Usage),
			UsageText:    jpdadd.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"ja"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return jpdAdd(c)
			},
		},
		{
			Name:         "jpd-delete",
			Flags:        cliutils.GetCommandFlags(cliutils.JpdDelete),
			Usage:        jpddelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc jpd-delete", jpddelete.GetDescription(), jpddelete.Usage),
			UsageText:    jpddelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"jd"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return jpdDelete(c)
			},
		},
	})
}

func jpdAdd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	jpdAddFlags, err := createJpdAddFlags(c)
	if err != nil {
		return err
	}
	return commands.JpdAdd(jpdAddFlags)
}

func jpdDelete(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	mcDetails, err := createMissionControlDetails(c)
	if err != nil {
		return err
	}
	return commands.JpdDelete(c.Args()[0], mcDetails)
}

func licenseAcquire(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	mcDetails, err := createMissionControlDetails(c)
	if err != nil {
		return err
	}

	return commands.LicenseAcquire(c.Args()[0], c.Args()[1], mcDetails)
}

func licenseDeploy(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	flags, err := createLicenseDeployFlags(c)
	if err != nil {
		return err
	}
	return commands.LicenseDeploy(c.Args()[0], c.Args()[1], flags)
}

func licenseRelease(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	mcDetails, err := createMissionControlDetails(c)
	if err != nil {
		return err
	}
	return commands.LicenseRelease(c.Args()[0], c.Args()[1], mcDetails)
}

func offerConfig(c *cli.Context) (*config.ServerDetails, error) {
	confirmed, err := cliutils.ShouldOfferConfig()
	if !confirmed || err != nil {
		return nil, err
	}
	details := createMCDetailsFromFlags(c)
	configCmd := coreCommonCommands.NewConfigCommand(coreCommonCommands.AddOrEdit, details.ServerId).SetDefaultDetails(details).SetInteractive(true)
	err = configCmd.Run()
	if err != nil {
		return nil, err
	}

	return configCmd.ServerDetails()
}

func createLicenseDeployFlags(c *cli.Context) (flags *commands.LicenseDeployFlags, err error) {
	flags = new(commands.LicenseDeployFlags)
	flags.ServerDetails, err = createMissionControlDetails(c)
	if err != nil {
		return
	}
	flags.LicenseCount = cliutils.DefaultLicenseCount
	if c.IsSet("license-count") {
		flags.LicenseCount, err = strconv.Atoi(c.String("license-count"))
		if err != nil {
			return nil, cliutils.PrintHelpAndReturnError("The '--license-count' option must have a numeric value. ", c)
		}
		if flags.LicenseCount < 1 {
			return nil, cliutils.PrintHelpAndReturnError("The --license-count option must be at least "+strconv.Itoa(cliutils.DefaultLicenseCount), c)
		}
	}
	return
}

func createJpdAddFlags(c *cli.Context) (flags *commands.JpdAddFlags, err error) {
	flags = new(commands.JpdAddFlags)
	flags.ServerDetails, err = createMissionControlDetails(c)
	if err != nil {
		return
	}
	flags.JpdConfig, err = fileutils.ReadFile(c.Args()[0])
	if errorutils.CheckError(err) != nil {
		return
	}
	return
}

func createMissionControlDetails(c *cli.Context) (*config.ServerDetails, error) {
	createdDetails, err := offerConfig(c)
	if err != nil {
		return nil, err
	}
	if createdDetails != nil {
		return createdDetails, nil
	}

	details := createMCDetailsFromFlags(c)
	// If urls or credentials were passed as options, use options as they are.
	// For security reasons, we'd like to avoid using part of the connection details from command options and the rest from the config.
	// Either use command options only or config only.
	if credentialsChanged(details) {
		return details, nil
	}

	// Else, use details from config for requested serverId, or for default server if empty.
	confDetails, err := coreCommonCommands.GetConfig(details.ServerId, true)
	if err != nil {
		return nil, err
	}

	confDetails.Url = clientutils.AddTrailingSlashIfNeeded(confDetails.MissionControlUrl)
	return confDetails, nil
}

func createMCDetailsFromFlags(c *cli.Context) (details *config.ServerDetails) {
	details = cliutils.CreateServerDetailsFromFlags(c)
	details.MissionControlUrl = details.Url
	details.Url = ""
	return
}

func credentialsChanged(details *config.ServerDetails) bool {
	return details.MissionControlUrl != "" || details.User != "" || details.Password != "" || details.AccessToken != ""
}

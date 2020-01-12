package missioncontrol

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/docs/common"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/acquirelicense"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/addjpd"
	configdocs "github.com/jfrog/jfrog-cli-go/docs/missioncontrol/config"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/deletejpd"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/deploylicense"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/releaselicense"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/commands"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"strconv"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "config",
			Flags:        getConfigFlags(),
			Usage:        configdocs.Description,
			HelpName:     common.CreateUsage("mc config", configdocs.Description, configdocs.Usage),
			UsageText:    configdocs.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"c"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return configure(c)
			},
		},
		{
			Name:         "acquire-license",
			Flags:        getMcAuthenticationFlags(),
			Usage:        acquirelicense.Description,
			HelpName:     common.CreateUsage("mc acquire-license", acquirelicense.Description, acquirelicense.Usage),
			UsageText:    acquirelicense.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"al"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return acquireLicense(c)
			},
		},
		{
			Name:         "deploy-license",
			Flags:        getDeployLicenseFlags(),
			Usage:        deploylicense.Description,
			HelpName:     common.CreateUsage("mc deploy-license", deploylicense.Description, deploylicense.Usage),
			UsageText:    deploylicense.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"dl"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return deployLicense(c)
			},
		},
		{
			Name:         "release-license",
			Flags:        getMcAuthenticationFlags(),
			Usage:        releaselicense.Description,
			HelpName:     common.CreateUsage("mc release-license", releaselicense.Description, releaselicense.Usage),
			UsageText:    releaselicense.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"rl"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return releaseLicense(c)
			},
		},
		{
			Name:         "add-jpd",
			Flags:        getMcAuthenticationFlags(),
			Usage:        addjpd.Description,
			HelpName:     common.CreateUsage("mc add-jpd", addjpd.Description, addjpd.Usage),
			UsageText:    addjpd.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"aj"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return addJpd(c)
			},
		},
		{
			Name:         "delete-jpd",
			Flags:        getMcAuthenticationFlags(),
			Usage:        deletejpd.Description,
			HelpName:     common.CreateUsage("mc delete-jpd", deletejpd.Description, deletejpd.Usage),
			UsageText:    deletejpd.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"dj"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return deleteJpd(c)
			},
		},
	}
}

func getMcAuthenticationFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Usage: "[Optional] Mission Control URL.` `",
		},
		cli.StringFlag{
			Name:  "access-token",
			Usage: "[Optional] Mission Control Admin token.` `",
		},
	}
}

func getDeployLicenseFlags() []cli.Flag {
	return append(getMcAuthenticationFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "license-count",
			Value: "",
			Usage: "[Default: " + strconv.Itoa(commands.DefaultLicenseCount) + "] The number of licenses to deploy. Minimum value is 1.` `",
		},
	}...)
}

func getConfigFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolTFlag{
			Name:  "interactive",
			Usage: "[Default: true] Set to false if you do not want the config command to be interactive. If true, the other command options become optional.",
		},
	}
	return append(flags, getMcAuthenticationFlags()...)
}

func addJpd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	addJpdFlags, err := createAddJpdFlags(c)
	if err != nil {
		return err
	}
	return commands.AddJpd(addJpdFlags)
}

func deleteJpd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	mcDetails, err := createMissionControlDetails(c, true)
	if err != nil {
		return err
	}
	return commands.DeleteJpd(c.Args()[0], mcDetails)
}

func acquireLicense(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	mcDetails, err := createMissionControlDetails(c, true)
	if err != nil {
		return err
	}

	return commands.AcquireLicense(c.Args()[0], c.Args()[1], mcDetails)
}

func deployLicense(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	flags, err := createDeployLicenseFlags(c)
	if err != nil {
		return err
	}
	return commands.DeployLicense(c.Args()[0], c.Args()[1], flags)
}

func releaseLicense(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	mcDetails, err := createMissionControlDetails(c, true)
	if err != nil {
		return err
	}
	return commands.ReleaseLicense(c.Args()[0], c.Args()[1], mcDetails)
}

func offerConfig(c *cli.Context) (*config.MissionControlDetails, error) {
	exists, err := config.IsMissionControlConfExists()
	if err != nil || exists {
		return nil, err
	}
	val, err := clientutils.GetBoolEnvValue(cliutils.OfferConfig, true)
	if err != nil {
		return nil, err
	}
	if !val {
		_ = config.SaveMissionControlConf(new(config.MissionControlDetails))
		return nil, nil
	}
	msg := fmt.Sprintf("To avoid this message in the future, set the %s environment variable to false.\n"+
		"The CLI commands require the Mission Control URL and authentication details\n"+
		"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n"+
		"You can also configure these parameters later using the 'config' command.\n"+
		"Configure now?", cliutils.OfferConfig)
	confirmed := cliutils.InteractiveConfirm(msg)
	if !confirmed {
		_ = config.SaveMissionControlConf(new(config.MissionControlDetails))
		return nil, nil
	}
	details, err := createMissionControlDetails(c, false)
	if err != nil {
		return nil, err
	}
	return commands.Config(nil, details, true)
}

func configure(c *cli.Context) error {
	if len(c.Args()) > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	if len(c.Args()) == 1 {
		switch c.Args()[0] {
		case "show":
			return commands.ShowConfig()
		case "clear":
			return commands.ClearConfig()
		default:
			return cliutils.PrintHelpAndReturnError("Unknown argument '"+c.Args()[0]+"'. Available arguments are 'show' and 'clear'.", c)
		}
	}
	flags, err := createConfigFlags(c)
	if err != nil {
		return err
	}
	_, err = commands.Config(flags.MissionControlDetails, nil, flags.Interactive)
	return err
}

func createDeployLicenseFlags(c *cli.Context) (flags *commands.DeployLicenseFlags, err error) {
	flags = new(commands.DeployLicenseFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	flags.LicenseCount = commands.DefaultLicenseCount
	if c.String("license-count") != "" {
		flags.LicenseCount, err = strconv.Atoi(c.String("license-count"))
		if err != nil {
			return nil, cliutils.PrintHelpAndReturnError("The '--license-count' option must have a numeric value. ", c)
		}
		if flags.LicenseCount < 1 {
			return nil, cliutils.PrintHelpAndReturnError("The --license-count option must be at least "+strconv.Itoa(commands.DefaultLicenseCount), c)
		}
	}
	return
}

func createConfigFlags(c *cli.Context) (flags *commands.ConfigFlags, err error) {
	flags = new(commands.ConfigFlags)
	flags.Interactive = c.BoolT("interactive")
	flags.MissionControlDetails, err = createMissionControlDetails(c, false)
	if err != nil {
		return
	}
	if !flags.Interactive && (flags.MissionControlDetails.Url == "" || flags.MissionControlDetails.AccessToken == "") {
		return nil, cliutils.PrintHelpAndReturnError("the --url and --access-token options are mandatory when the --interactive option is set to false", c)
	}
	return
}

func createAddJpdFlags(c *cli.Context) (flags *commands.AddJpdFlags, err error) {
	flags = new(commands.AddJpdFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	flags.JpdSpec, err = fileutils.ReadFile(c.Args()[0])
	if errorutils.CheckError(err) != nil {
		return
	}
	return
}

func createMissionControlDetails(c *cli.Context, includeConfig bool) (*config.MissionControlDetails, error) {
	if includeConfig {
		details, err := offerConfig(c)
		if err != nil {
			return nil, err
		}
		if details != nil {
			return details, nil
		}
	}
	details := new(config.MissionControlDetails)
	details.Url = c.String("url")
	details.AccessToken = c.String("access-token")

	if includeConfig {
		if details.Url == "" || details.AccessToken == "" {
			confDetails, err := commands.GetConfig()
			if err != nil {
				return nil, err
			}
			if details.Url == "" {
				details.Url = confDetails.Url
			}
			if details.AccessToken == "" {
				details.SetAccessToken(confDetails.AccessToken)
			}
		}
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	return details, nil
}

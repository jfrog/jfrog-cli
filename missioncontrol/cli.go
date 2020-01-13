package missioncontrol

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/docs/common"
	configdocs "github.com/jfrog/jfrog-cli-go/docs/missioncontrol/config"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/jpdadd"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/jpddelete"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/licenseacquire"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/licensedeploy"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/licenserelease"
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
			Name:         "license-acquire",
			Flags:        getMcAuthenticationFlags(),
			Usage:        licenseacquire.Description,
			HelpName:     common.CreateUsage("mc license-acquire", licenseacquire.Description, licenseacquire.Usage),
			UsageText:    licenseacquire.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"la"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return licenseAcquire(c)
			},
		},
		{
			Name:         "license-deploy",
			Flags:        getLicenseDeployFlags(),
			Usage:        licensedeploy.Description,
			HelpName:     common.CreateUsage("mc license-deploy", licensedeploy.Description, licensedeploy.Usage),
			UsageText:    licensedeploy.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"ld"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return licenseDeploy(c)
			},
		},
		{
			Name:         "license-release",
			Flags:        getMcAuthenticationFlags(),
			Usage:        licenserelease.Description,
			HelpName:     common.CreateUsage("mc license-release", licenserelease.Description, licenserelease.Usage),
			UsageText:    licenserelease.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"lr"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return licenseRelease(c)
			},
		},
		{
			Name:         "jpd-add",
			Flags:        getMcAuthenticationFlags(),
			Usage:        jpdadd.Description,
			HelpName:     common.CreateUsage("mc jpd-add", jpdadd.Description, jpdadd.Usage),
			UsageText:    jpdadd.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"ja"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return jpdAdd(c)
			},
		},
		{
			Name:         "jpd-delete",
			Flags:        getMcAuthenticationFlags(),
			Usage:        jpddelete.Description,
			HelpName:     common.CreateUsage("mc jpd-delete", jpddelete.Description, jpddelete.Usage),
			UsageText:    jpddelete.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"jd"},
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return jpdDelete(c)
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

func getLicenseDeployFlags() []cli.Flag {
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

func jpdAdd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	jpdAddFlags, err := createJpdAddFlags(c)
	if err != nil {
		return err
	}
	return commands.JpdAdd(jpdAddFlags)
}

func jpdDelete(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	mcDetails, err := createMissionControlDetails(c, true)
	if err != nil {
		return err
	}
	return commands.JpdDelete(c.Args()[0], mcDetails)
}

func licenseAcquire(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	mcDetails, err := createMissionControlDetails(c, true)
	if err != nil {
		return err
	}

	return commands.LicenseAcquire(c.Args()[0], c.Args()[1], mcDetails)
}

func licenseDeploy(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
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
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	mcDetails, err := createMissionControlDetails(c, true)
	if err != nil {
		return err
	}
	return commands.LicenseRelease(c.Args()[0], c.Args()[1], mcDetails)
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

func createLicenseDeployFlags(c *cli.Context) (flags *commands.LicenseDeployFlags, err error) {
	flags = new(commands.LicenseDeployFlags)
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

func createJpdAddFlags(c *cli.Context) (flags *commands.JpdAddFlags, err error) {
	flags = new(commands.JpdAddFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	flags.JpdConfig, err = fileutils.ReadFile(c.Args()[0])
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

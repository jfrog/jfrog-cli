package missioncontrol

import (
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/docs/common"
	configdocs "github.com/jfrog/jfrog-cli-go/docs/missioncontrol/config"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/services/add"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/services/attachlic"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/services/detachlic"
	"github.com/jfrog/jfrog-cli-go/docs/missioncontrol/services/remove"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/commands"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/commands/services"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"strings"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "services",
			Aliases:     []string{"s"},
			Usage:       "Services",
			Subcommands: getRtiSubCommands(),
		},
		{
			Name:      "config",
			Flags:     getConfigFlags(),
			Usage:     configdocs.Description,
			HelpName:  common.CreateUsage("mc config", configdocs.Description, configdocs.Usage),
			UsageText: configdocs.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Aliases:   []string{"c"},
			Action:    configure,
		},
	}
}

func getRtiSubCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "add",
			Flags:     getAddServiceFlags(),
			Usage:     add.Description,
			HelpName:  common.CreateUsage("mc services add", add.Description, add.Usage),
			UsageText: add.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action:    addService,
		},
		{
			Name:      "remove",
			Flags:     getRemoveServiceFlags(),
			Usage:     remove.Description,
			HelpName:  common.CreateUsage("mc services remove", remove.Description, remove.Usage),
			UsageText: remove.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action:    removeService,
		},
		{
			Name:      "attach-lic",
			Flags:     getAttachLicenseFlags(),
			Usage:     attachlic.Description,
			HelpName:  common.CreateUsage("mc services attach-lic", attachlic.Description, attachlic.Usage),
			UsageText: attachlic.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action:    attachLicense,
		},
		{
			Name:      "detach-lic",
			Flags:     getDetachLicenseFlags(),
			Usage:     detachlic.Description,
			HelpName:  common.CreateUsage("mc services detach-lic", detachlic.Description, detachlic.Usage),
			UsageText: detachlic.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action:    detachLicense,
		},
	}
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Usage: "[Optional] Mission Control URL` `",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "[Optional] Mission Control username` `",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "[Optional] Mission Control password` `",
		},
	}
}

func getRemoveServiceFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "[Default: false] Set to true to skip the delete confirmation message.",
		},
	}...)
}

func getAddServiceFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "service-url",
			Usage: "[Mandatory] Service URL.` `",
		},
		cli.StringFlag{
			Name:  "service-user",
			Usage: "[Mandatory] Service username.` `",
		},
		cli.StringFlag{
			Name:  "service-password",
			Usage: "[Mandatory] Service password.` `",
		},
		cli.StringFlag{
			Name:  "desc",
			Usage: "[Optional] Service description.` `",
		},
		cli.StringFlag{
			Name:  "site-name",
			Usage: "[Optional] Service site name, e.g. US.` `",
		},
	}...)
}

func getAttachLicenseFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "bucket-id",
			Usage: "[Mandatory] license bucket ID` `",
		},
		cli.StringFlag{
			Name:  "license-path",
			Usage: "[Optional] Full path to the license file` `",
		},
		cli.BoolFlag{
			Name:  "override",
			Usage: "[Default: false] Set to true to override licence file.` `",
		},
		cli.BoolFlag{
			Name:  "deploy",
			Usage: "[Default: false] Set to true to deploy licence to service.` `",
		},
	}...)
}

func getDetachLicenseFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "bucket-id",
			Usage: "[Mandatory] license bucket ID` `",
		},
	}...)
}

func getConfigFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolTFlag{
			Name:  "interactive",
			Usage: "[Default: true] Set to false if you do not want the config command to be interactive. If true, the --url option becomes optional.",
		},
	}
	return append(flags, getFlags()...)
}

func addService(c *cli.Context) {
	if len(c.Args()) != 2 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	addServiceFlags, err := createAddServiceFlag(c)
	cliutils.ExitOnErr(err)
	serviceType := c.Args()[0]
	serviceName := c.Args()[1]
	err = services.AddService(serviceType, serviceName, addServiceFlags)
	cliutils.ExitOnErr(err)
}

func removeService(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	serviceName := c.Args()[0]
	if !c.Bool("quiet") {
		confirmed := cliutils.InteractiveConfirm("Remove Service,  " + serviceName + "?")
		if !confirmed {
			return
		}
	}
	flags, err := createRemoveServiceFlags(c)
	cliutils.ExitOnErr(err)
	err = services.Remove(serviceName, flags)
	cliutils.ExitOnErr(err)
}

func attachLicense(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	flags, err := createAttachLicFlags(c)
	cliutils.ExitOnErr(err)
	err = services.AttachLic(c.Args()[0], flags)
	cliutils.ExitOnErr(err)
}

func detachLicense(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	flags, err := createDetachLicFlags(c)
	cliutils.ExitOnErr(err)
	err = services.DetachLic(c.Args()[0], flags)
	cliutils.ExitOnErr(err)
}

func offerConfig(c *cli.Context) (*config.MissionControlDetails, error) {
	exists, err := config.IsMissionControlConfExists()
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, nil
	}
	val, err := clientutils.GetBoolEnvValue(cliutils.OfferConfig, true)
	if err != nil {
		return nil, err
	}
	if !val {
		config.SaveMissionControlConf(new(config.MissionControlDetails))
		return nil, nil
	}
	msg := fmt.Sprintf("To avoid this message in the future, set the %s environment variable to false.\n"+
		"The CLI commands require the Mission Control URL and authentication details\n"+
		"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n"+
		"You can also configure these parameters later using the 'config' command.\n"+
		"Configure now?", cliutils.OfferConfig)
	confirmed := cliutils.InteractiveConfirm(msg)
	if !confirmed {
		config.SaveMissionControlConf(new(config.MissionControlDetails))
		return nil, nil
	}
	details, err := createMissionControlDetails(c, false)
	if err != nil {
		return nil, err
	}
	return commands.Config(nil, details, true)
}

func configure(c *cli.Context) {
	if len(c.Args()) > 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	} else if len(c.Args()) == 1 {
		if c.Args()[0] == "show" {
			commands.ShowConfig()
		} else if c.Args()[0] == "clear" {
			commands.ClearConfig()
		} else {
			cliutils.ExitOnErr(errors.New("Unknown argument '" + c.Args()[0] + "'. Available arguments are 'show' and 'clear'."))
		}
	} else {
		flags, err := createConfigFlags(c)
		cliutils.ExitOnErr(err)
		commands.Config(flags.MissionControlDetails, nil, flags.Interactive)
	}
}

func createDetachLicFlags(c *cli.Context) (flags *services.DetachLicFlags, err error) {
	flags = new(services.DetachLicFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	if flags.BucketId = c.String("bucket-id"); flags.BucketId == "" {
		cliutils.PrintHelpAndExitWithError("The --bucket-id option is mandatory.", c)
	}
	return
}

func createAttachLicFlags(c *cli.Context) (flags *services.AttachLicFlags, err error) {
	flags = new(services.AttachLicFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	flags.LicensePath = c.String("license-path")
	if strings.HasSuffix(flags.LicensePath, fileutils.GetFileSeparator()) {
		cliutils.ExitOnErr(errors.New("The --license-path option cannot be a directory"))
	}
	if flags.BucketId = c.String("bucket-id"); flags.BucketId == "" {
		cliutils.PrintHelpAndExitWithError("The --bucket-id option is mandatory.", c)
	}
	flags.Override = c.Bool("override")
	flags.Deploy = c.Bool("deploy")
	return
}

func createConfigFlags(c *cli.Context) (flags *commands.ConfigFlags, err error) {
	flags = new(commands.ConfigFlags)
	flags.Interactive = c.BoolT("interactive")
	flags.MissionControlDetails, err = createMissionControlDetails(c, false)
	if err != nil {
		return
	}
	if !flags.Interactive && flags.MissionControlDetails.Url == "" {
		cliutils.ExitOnErr(errors.New("The --url option is mandatory when the --interactive option is set to false"))
	}
	return
}

func createAddServiceFlag(c *cli.Context) (flags *services.AddServiceFlags, err error) {
	flags = new(services.AddServiceFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	flags.ServiceDetails = new(utils.ServiceDetails)

	if flags.ServiceDetails.Url = c.String("service-url"); flags.ServiceDetails.Url == "" {
		cliutils.ExitOnErr(errors.New("The --service-url option is mandatory"))
	}
	if flags.ServiceDetails.User = c.String("service-user"); flags.ServiceDetails.User == "" {
		cliutils.ExitOnErr(errors.New("The --service-user option is mandatory"))
	}
	if flags.ServiceDetails.Password = c.String("service-password"); flags.ServiceDetails.Password == "" {
		cliutils.ExitOnErr(errors.New("The --service-password option is mandatory"))
	}
	flags.Description = c.String("desc")
	flags.SiteName = c.String("site-name")
	return
}

func createRemoveServiceFlags(c *cli.Context) (flags *services.RemoveFlags, err error) {
	details, err := createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	flags = &services.RemoveFlags{
		MissionControlDetails: details,
		Interactive:           c.BoolT("interactive")}

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
	details.User = c.String("user")
	details.Password = c.String("password")

	if includeConfig {
		if details.Url == "" || details.User == "" || details.Password == "" {
			confDetails, err := commands.GetConfig()
			if err != nil {
				return nil, err
			}
			if details.Url == "" {
				details.Url = confDetails.Url
			}
			if details.User == "" {
				details.SetUser(confDetails.User)
			}
			if details.Password == "" {
				details.SetPassword(confDetails.Password)
			}
		}
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	return details, nil
}

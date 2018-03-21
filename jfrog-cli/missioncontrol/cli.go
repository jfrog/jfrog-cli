package missioncontrol

import (
	"github.com/codegangsta/cli"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/common"
	configdocs "github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/missioncontrol/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/missioncontrol/rtinstances/add"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/missioncontrol/rtinstances/attachlic"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/missioncontrol/rtinstances/detachlic"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/docs/missioncontrol/rtinstances/remove"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/missioncontrol/commands"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/missioncontrol/commands/rtinstances"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"strings"
	"errors"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "rt-instances",
			Aliases:     []string{"rti"},
			Usage:       "Artifactory instances",
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
			Flags:     getAddInstanceFlags(),
			Usage:     add.Description,
			HelpName:  common.CreateUsage("mc rt-instances add", add.Description, add.Usage),
			UsageText: add.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action:    addInstance,
		},
		{
			Name:      "remove",
			Flags:     getRemoveInstanceFlags(),
			Usage:     remove.Description,
			HelpName:  common.CreateUsage("mc rt-instances remove", remove.Description, remove.Usage),
			UsageText: remove.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action:    removeInstance,
		},
		{
			Name:      "attach-lic",
			Flags:     getAttachLicenseFlags(),
			Usage:     attachlic.Description,
			HelpName:  common.CreateUsage("mc rt-instances attach-lic", attachlic.Description, attachlic.Usage),
			UsageText: attachlic.Arguments,
			ArgsUsage: common.CreateEnvVars(),
			Action:    attachLicense,
		},
		{
			Name:      "detach-lic",
			Flags:     getDetachLicenseFlags(),
			Usage:     detachlic.Description,
			HelpName:  common.CreateUsage("mc rt-instances detach-lic", detachlic.Description, detachlic.Usage),
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
			Usage: "[Optional] Mission Control URL",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "[Optional] Mission Control username",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "[Optional] Mission Control password",
		},
	}
}

func getRemoveInstanceFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "[Default: false] Set to true to skip the delete confirmation message.",
		},
	}...)
}

func getAddInstanceFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "rt-url",
			Usage: "[Mandatory] Artifactory URL.",
		},
		cli.StringFlag{
			Name:  "rt-user",
			Usage: "[Mandatory] Artifactory admin username.",
		},
		cli.StringFlag{
			Name:  "rt-password",
			Usage: "[Mandatory] Artifactory admin password - optionally encrypted.",
		},
		cli.StringFlag{
			Name:  "desc",
			Usage: "[Optional] Artifactory instance description.",
		},
		cli.StringFlag{
			Name:  "location",
			Usage: "[Optional] Artifactory instance location, e.g. US.",
		},
	}...)
}

func getAttachLicenseFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "bucket-id",
			Usage: "[Mandatory] license bucket ID",
		},
		cli.StringFlag{
			Name:  "node-id",
			Usage: "[Optional] Unique HA node identifier",
		},
		cli.StringFlag{
			Name:  "license-path",
			Usage: "[Optional] Full path to the license file",
		},
		cli.BoolFlag{
			Name:  "override",
			Usage: "[Default: false] Set to true to override licence file.",
		},
		cli.BoolFlag{
			Name:  "deploy",
			Usage: "[Default: false] Set to true to deploy licence to instace.",
		},
	}...)
}

func getDetachLicenseFlags() []cli.Flag {
	return append(getFlags(), []cli.Flag{
		cli.StringFlag{
			Name:  "bucket-id",
			Usage: "[Mandatory] license bucket ID",
		},
		cli.StringFlag{
			Name:  "node-id",
			Usage: "[Optional] Unique HA node identifier",
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

func addInstance(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	addInstanceFlags, err := createAddInstanceFlag(c)
	cliutils.ExitOnErr(err)
	rtinstances.AddInstance(c.Args()[0], addInstanceFlags)
}

func removeInstance(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	instanceName := c.Args()[0]
	if !c.Bool("quiet") {
		confirmed := cliutils.InteractiveConfirm("Remove Instance,  " + instanceName + "?")
		if !confirmed {
			return
		}
	}
	flags, err := createRemoveInstanceFlags(c)
	cliutils.ExitOnErr(err)
	rtinstances.Remove(instanceName, flags)
}

func attachLicense(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	flags, err := createAttachLicFlags(c)
	cliutils.ExitOnErr(err)
	rtinstances.AttachLic(c.Args()[0], flags)
}

func detachLicense(c *cli.Context) {
	size := len(c.Args())
	if size != 1 {
		cliutils.PrintHelpAndExitWithError("Wrong number of arguments.", c)
	}
	flags, err := createDetachLicFlags(c)
	cliutils.ExitOnErr(err)
	rtinstances.DetachLic(c.Args()[0], flags)
}

func offerConfig(c *cli.Context) (*config.MissionControlDetails, error) {
	exists, err := config.IsMissionControlConfExists()
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, nil
	}
	val, err := cliutils.GetBoolEnvValue("JFROG_CLI_OFFER_CONFIG", true)
	if err != nil {
		return nil, err
	}
	if !val {
		config.SaveMissionControlConf(new(config.MissionControlDetails))
		return nil, nil
	}
	msg := "The CLI commands require the Mission Control URL and authentication details\n" +
		"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n" +
		"You can also configure these parameters later using the 'config' command.\n" +
		"Configure now?"
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
			cliutils.ExitOnErr(errors.New("Unknown argument '"+c.Args()[0]+"'. Available arguments are 'show' and 'clear'."))
		}
	} else {
		flags, err := createConfigFlags(c)
		cliutils.ExitOnErr(err)
		commands.Config(flags.MissionControlDetails, nil, flags.Interactive)
	}
}

func createDetachLicFlags(c *cli.Context) (flags *rtinstances.DetachLicFlags, err error) {
	flags = new(rtinstances.DetachLicFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	if flags.BucketId = c.String("bucket-id"); flags.BucketId == "" {
		cliutils.PrintHelpAndExitWithError("The --bucket-id option is mandatory.", c)
	}
	flags.NodeId = c.String("node-id")
	return
}

func createAttachLicFlags(c *cli.Context) (flags *rtinstances.AttachLicFlags, err error) {
	flags = new(rtinstances.AttachLicFlags)
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
	flags.NodeId = c.String("node-id")
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

func createAddInstanceFlag(c *cli.Context) (flags *rtinstances.AddInstanceFlags, err error) {
	flags = new(rtinstances.AddInstanceFlags)
	flags.MissionControlDetails, err = createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	flags.ArtifactoryInstanceDetails = new(utils.ArtifactoryInstanceDetails)
	if flags.ArtifactoryInstanceDetails.Url = c.String("rt-url"); flags.ArtifactoryInstanceDetails.Url == "" {
		cliutils.ExitOnErr(errors.New("The --rt-url option is mandatory"))
	}
	if flags.ArtifactoryInstanceDetails.User = c.String("rt-user"); flags.ArtifactoryInstanceDetails.User == "" {
		cliutils.ExitOnErr(errors.New("The --rt-user option is mandatory"))
	}
	if flags.ArtifactoryInstanceDetails.Password = c.String("rt-password"); flags.ArtifactoryInstanceDetails.Password == "" {
		cliutils.ExitOnErr(errors.New("The --rt-password option is mandatory"))
	}
	flags.Description = c.String("desc")
	flags.Location = c.String("location")
	return
}

func createRemoveInstanceFlags(c *cli.Context) (flags *rtinstances.RemoveFlags, err error) {
	details, err := createMissionControlDetails(c, true)
	if err != nil {
		return
	}
	flags = &rtinstances.RemoveFlags{
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

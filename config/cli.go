package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/docs/common"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/docs/config/delete"
	"github.com/jfrog/jfrog-cli/docs/config/use"

	"github.com/jfrog/jfrog-cli/docs/config/exportcmd"
	"github.com/jfrog/jfrog-cli/docs/config/importcmd"
	"github.com/jfrog/jfrog-cli/docs/config/show"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "show",
			Aliases:      []string{"s"},
			Description:  show.Description,
			HelpName:     corecommon.CreateUsage("c show", show.Description, show.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return showCmd(c)
			},
		},
		{
			Name:         "delete",
			Aliases:      []string{"del"},
			Description:  delete.Description,
			Flags:        cliutils.GetCommandFlags(cliutils.DeleteConfig),
			HelpName:     corecommon.CreateUsage("c del", delete.Description, delete.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return deleteCmd(c)
			},
		},
		{
			Name:         "import",
			Aliases:      []string{"im"},
			Description:  importcmd.Description,
			HelpName:     corecommon.CreateUsage("c import", importcmd.Description, importcmd.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return importCmd(c)
			},
		},
		{
			Name:         "export",
			Aliases:      []string{"ex"},
			Description:  exportcmd.Description,
			HelpName:     corecommon.CreateUsage("c export", exportcmd.Description, exportcmd.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return exportCmd(c)
			},
		},
		{
			Name:         "use",
			Description:  use.Description,
			HelpName:     corecommon.CreateUsage("c use", use.Description, use.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return useCmd(c)
			},
		},
	}
}

func ConfigCmd(c *cli.Context) error {
	if len(c.Args()) > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}

	configCommandConfiguration, err := createConfigCommandConfiguration(c)
	if err != nil {
		return err
	}

	var serverId string
	if len(c.Args()) > 0 {
		serverId = c.Args()[0]
		if err := ValidateServerId(serverId); err != nil {
			return err
		}
	}
	err = validateConfigFlags(configCommandConfiguration)
	if err != nil {
		return err
	}
	configCmd := commands.NewConfigCommand().SetDetails(configCommandConfiguration.ServerDetails).SetInteractive(configCommandConfiguration.Interactive).
		SetServerId(serverId).SetEncPassword(configCommandConfiguration.EncPassword).SetUseBasicAuthOnly(configCommandConfiguration.BasicAuthOnly)
	return configCmd.Config()
}

func showCmd(c *cli.Context) error {
	if len(c.Args()) > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	var serverId string
	if c.NArg() == 1 {
		serverId = c.Args()[0]
	}
	return commands.ShowConfig(serverId)
}

func deleteCmd(c *cli.Context) error {
	if len(c.Args()) > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	quiet := cliutils.GetQuietValue(c)

	// Clear all configurations
	if c.NArg() == 0 {
		commands.ClearConfig(!quiet)
		return nil
	}

	// Delete single configuration
	serverId := c.Args()[0]
	if !quiet && !coreutils.AskYesNo("Are you sure you want to delete \""+serverId+"\" configuration?", false) {
		return nil
	}
	return commands.DeleteConfig(serverId)
}

func importCmd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commands.Import(c.Args()[0])
}

func exportCmd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commands.Export(c.Args()[0])
}

func useCmd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commands.Use(c.Args()[0])
}

func createConfigCommandConfiguration(c *cli.Context) (configCommandConfiguration *commands.ConfigCommandConfiguration, err error) {
	configCommandConfiguration = new(commands.ConfigCommandConfiguration)
	configCommandConfiguration.ServerDetails = cliutils.CreateServerDetailsFromFlags(c)
	configCommandConfiguration.EncPassword = c.BoolT("enc-password")
	configCommandConfiguration.Interactive = cliutils.GetInteractiveValue(c)
	configCommandConfiguration.BasicAuthOnly = c.Bool("basic-auth-only")
	return
}

func ValidateServerId(serverId string) error {
	reservedIds := []string{"delete", "use", "show", "clear"}
	for _, reservedId := range reservedIds {
		if serverId == reservedId {
			return errors.New(fmt.Sprintf("Server can't have one of the following ID's: %s\n %s", strings.Join(reservedIds, ", "), cliutils.GetDocumentationMessage()))
		}
	}
	return nil
}

func validateConfigFlags(configCommandConfiguration *commands.ConfigCommandConfiguration) error {
	if !configCommandConfiguration.Interactive && configCommandConfiguration.ServerDetails.ArtifactoryUrl == "" {
		return errors.New("the --artifactory-url option is mandatory when the --interactive option is set to false or the CI environment variable is set to true.")
	}
	// Validate the option is not used along an access token
	if configCommandConfiguration.BasicAuthOnly && configCommandConfiguration.ServerDetails.AccessToken != "" {
		return errors.New("the --basic-auth-only option is only supported when username and password/API key are provided")
	}
	return nil
}

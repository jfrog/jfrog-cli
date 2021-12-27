package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/docs/config/add"
	"github.com/jfrog/jfrog-cli/docs/config/edit"
	"github.com/jfrog/jfrog-cli/docs/config/remove"
	"github.com/jfrog/jfrog-cli/docs/config/use"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/urfave/cli"

	"github.com/jfrog/jfrog-cli/docs/config/exportcmd"
	"github.com/jfrog/jfrog-cli/docs/config/importcmd"
	"github.com/jfrog/jfrog-cli/docs/config/show"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "add",
			Description:  add.GetDescription(),
			Flags:        cliutils.GetCommandFlags(cliutils.AddConfig),
			HelpName:     corecommon.CreateUsage("c add", add.GetDescription(), add.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return addCmd(c)
			},
		},
		{
			Name:         "edit",
			Description:  edit.GetDescription(),
			Flags:        cliutils.GetCommandFlags(cliutils.EditConfig),
			HelpName:     corecommon.CreateUsage("c edit", edit.GetDescription(), edit.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return editCmd(c)
			},
		},
		{
			Name:         "show",
			Aliases:      []string{"s"},
			Description:  show.GetDescription(),
			HelpName:     corecommon.CreateUsage("c show", show.GetDescription(), show.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return showCmd(c)
			},
		},
		{
			Name:         "remove",
			Aliases:      []string{"rm"},
			Description:  remove.GetDescription(),
			Flags:        cliutils.GetCommandFlags(cliutils.DeleteConfig),
			HelpName:     corecommon.CreateUsage("c rm", remove.GetDescription(), remove.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return deleteCmd(c)
			},
		},
		{
			Name:         "import",
			Aliases:      []string{"im"},
			Description:  importcmd.GetDescription(),
			HelpName:     corecommon.CreateUsage("c import", importcmd.GetDescription(), importcmd.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return importCmd(c)
			},
		},
		{
			Name:         "export",
			Aliases:      []string{"ex"},
			Description:  exportcmd.GetDescription(),
			HelpName:     corecommon.CreateUsage("c export", exportcmd.GetDescription(), exportcmd.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return exportCmd(c)
			},
		},
		{
			Name:         "use",
			Description:  use.GetDescription(),
			HelpName:     corecommon.CreateUsage("c use", use.GetDescription(), use.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action: func(c *cli.Context) error {
				return useCmd(c)
			},
		},
	})
}

func addCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	if c.Bool("overwrite") {
		return addOrEdit(c, overwriteOperation)
	}
	return addOrEdit(c, addOperation)
}

func editCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return addOrEdit(c, editOperation)
}

type configOperation int

const (
	// "config add" comment
	addOperation configOperation = iota
	// "config edit" comment
	editOperation
	// "config add with --overwrite" comment
	overwriteOperation
)

func addOrEdit(c *cli.Context, operation configOperation) error {
	configCommandConfiguration, err := CreateConfigCommandConfiguration(c)
	if err != nil {
		return err
	}

	var serverId string
	if c.NArg() > 0 {
		serverId = c.Args()[0]
		if err := ValidateServerId(serverId); err != nil {
			return err
		}
		if operation != overwriteOperation {
			if err := validateServerExistence(serverId, operation); err != nil {
				return err
			}
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
	if c.NArg() > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	var serverId string
	if c.NArg() == 1 {
		serverId = c.Args()[0]
	}
	return commands.ShowConfig(serverId)
}

func deleteCmd(c *cli.Context) error {
	if c.NArg() > 1 {
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
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commands.Import(c.Args()[0])
}

func exportCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	// If no server Id was given, export the default server.
	serverId := ""
	if c.NArg() == 1 {
		serverId = c.Args()[0]
	}
	return commands.Export(serverId)
}

func useCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return commands.Use(c.Args()[0])
}

func CreateConfigCommandConfiguration(c *cli.Context) (configCommandConfiguration *commands.ConfigCommandConfiguration, err error) {
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
			return fmt.Errorf("Server can't have one of the following ID's: %s\n %s", strings.Join(reservedIds, ", "), cliutils.GetDocumentationMessage())
		}
	}
	return nil
}

func validateServerExistence(serverId string, operation configOperation) error {
	config, err := commands.GetConfig(serverId, false)
	serverExist := err == nil && config.ServerId != ""
	if operation == editOperation && !serverExist {
		return errorutils.CheckErrorf("Server ID '%s' doesn't exist.", serverId)
	} else if operation == addOperation && serverExist {
		return errorutils.CheckErrorf("Server ID '%s' already exists.", serverId)
	}
	return nil
}

func validateConfigFlags(configCommandConfiguration *commands.ConfigCommandConfiguration) error {
	// Validate the option is not used along with an access token
	if configCommandConfiguration.BasicAuthOnly && configCommandConfiguration.ServerDetails.AccessToken != "" {
		return errors.New("the --basic-auth-only option is only supported when username and password/API key are provided")
	}
	return nil
}

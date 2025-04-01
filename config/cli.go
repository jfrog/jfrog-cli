package config

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/general/token"
	"os"
	"strings"

	"github.com/jfrog/jfrog-client-go/auth/cert"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/docs/config/add"
	"github.com/jfrog/jfrog-cli/docs/config/edit"
	"github.com/jfrog/jfrog-cli/docs/config/remove"
	"github.com/jfrog/jfrog-cli/docs/config/use"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
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
			Usage:        add.GetDescription(),
			Flags:        cliutils.GetCommandFlags(cliutils.AddConfig),
			HelpName:     corecommon.CreateUsage("c add", add.GetDescription(), add.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       addCmd,
		},
		{
			Name:         "edit",
			Usage:        edit.GetDescription(),
			Flags:        cliutils.GetCommandFlags(cliutils.EditConfig),
			HelpName:     corecommon.CreateUsage("c edit", edit.GetDescription(), edit.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       editCmd,
		},
		{
			Name:         "show",
			Aliases:      []string{"s"},
			Usage:        show.GetDescription(),
			HelpName:     corecommon.CreateUsage("c show", show.GetDescription(), show.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       showCmd,
		},
		{
			Name:         "remove",
			Aliases:      []string{"rm"},
			Usage:        remove.GetDescription(),
			Flags:        cliutils.GetCommandFlags(cliutils.DeleteConfig),
			HelpName:     corecommon.CreateUsage("c rm", remove.GetDescription(), remove.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       deleteCmd,
		},
		{
			Name:         "import",
			Aliases:      []string{"im"},
			Usage:        importcmd.GetDescription(),
			HelpName:     corecommon.CreateUsage("c import", importcmd.GetDescription(), importcmd.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       importCmd,
		},
		{
			Name:         "export",
			Aliases:      []string{"ex"},
			Usage:        exportcmd.GetDescription(),
			HelpName:     corecommon.CreateUsage("c export", exportcmd.GetDescription(), exportcmd.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       exportCmd,
		},
		{
			Name:         "use",
			Usage:        use.GetDescription(),
			HelpName:     corecommon.CreateUsage("c use", use.GetDescription(), use.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       useCmd,
		},
	})
}

func addCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	if c.Bool(cliutils.Overwrite) {
		return addOrEdit(c, overwriteOperation)
	}
	return addOrEdit(c, addOperation)
}

func editCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
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

	oidcParams, err := createOidcParamsFromFlags(c)
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

	configCmd := commands.NewConfigCommand(commands.AddOrEdit, serverId).
		SetDetails(configCommandConfiguration.ServerDetails).
		SetInteractive(configCommandConfiguration.Interactive).
		SetEncPassword(configCommandConfiguration.EncPassword).
		SetUseBasicAuthOnly(configCommandConfiguration.BasicAuthOnly).
		SetOIDCParams(oidcParams)

	return configCmd.Run()
}

func createOidcParamsFromFlags(c *cli.Context) (*token.OidcTokenParams, error) {
	providerType, err := token.OidcProviderTypeFromString(cliutils.GetFlagOrEnvValue(c, cliutils.OidcProviderType, coreutils.OidcProviderType))
	if err != nil {
		return nil, err
	}
	return &token.OidcTokenParams{
		ProviderType: providerType,
		ProviderName: c.String(cliutils.OidcProviderName),
		Audience:     c.String(cliutils.OidcAudience),
		// Values can be set by the user or injected by the CI wrapper plugin
		TokenId:        cliutils.GetFlagOrEnvValue(c, cliutils.OidcTokenID, coreutils.OidcExchangeTokenId),
		ProjectKey:     cliutils.GetFlagOrEnvValue(c, cliutils.Project, coreutils.Project),
		ApplicationKey: cliutils.GetFlagOrEnvValue(c, cliutils.ApplicationKey, coreutils.ApplicationKey),
		// Values from the CI environment
		JobId:      os.Getenv(coreutils.CIJobID),
		RunId:      os.Getenv(coreutils.CIRunID),
		Repository: os.Getenv(coreutils.SourceCodeRepository),
	}, nil
}

func showCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	var serverId string
	if c.NArg() == 1 {
		serverId = c.Args()[0]
	}
	return commands.ShowConfig(serverId)
}

func deleteCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	quiet := cliutils.GetQuietValue(c)

	// Clear all configurations
	if c.NArg() == 0 {
		return commands.NewConfigCommand(commands.Clear, "").SetInteractive(!quiet).Run()
	}

	// Delete single configuration
	serverId := c.Args()[0]
	if !quiet && !coreutils.AskYesNo("Are you sure you want to delete \""+serverId+"\" configuration?", false) {
		return nil
	}
	return commands.NewConfigCommand(commands.Delete, serverId).Run()
}

func importCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	return commands.Import(c.Args()[0])
}

func exportCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
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
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	serverId := c.Args()[0]
	return commands.NewConfigCommand(commands.Use, serverId).Run()
}

func CreateConfigCommandConfiguration(c *cli.Context) (configCommandConfiguration *commands.ConfigCommandConfiguration, err error) {
	configCommandConfiguration = new(commands.ConfigCommandConfiguration)
	if configCommandConfiguration.ServerDetails, err = cliutils.CreateServerDetailsFromFlags(c); err != nil {
		return
	}
	if configCommandConfiguration.OidcParams, err = createOidcParamsFromFlags(c); err != nil {
		return
	}
	configCommandConfiguration.EncPassword = c.BoolT(cliutils.EncPassword)
	configCommandConfiguration.Interactive = cliutils.GetInteractiveValue(c)
	configCommandConfiguration.BasicAuthOnly = c.Bool(cliutils.BasicAuthOnly)
	return
}

func ValidateServerId(serverId string) error {
	reservedIds := []string{"delete", "use", "show", "clear"}
	for _, reservedId := range reservedIds {
		if serverId == reservedId {
			return fmt.Errorf("server can't have one of the following ID's: %s\n%s", strings.Join(reservedIds, ", "), cliutils.GetDocumentationMessage())
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
	// Validate the option is not used along with access token
	if configCommandConfiguration.BasicAuthOnly && configCommandConfiguration.ServerDetails.AccessToken != "" {
		return errorutils.CheckErrorf("the --%s option is only supported when username and password/API key are provided", cliutils.BasicAuthOnly)
	}
	if err := validatePathsExist(configCommandConfiguration.ServerDetails.SshKeyPath, configCommandConfiguration.ServerDetails.ClientCertPath, configCommandConfiguration.ServerDetails.ClientCertKeyPath); err != nil {
		return err
	}
	if configCommandConfiguration.ServerDetails.ClientCertPath != "" {
		_, err := cert.LoadCertificate(configCommandConfiguration.ServerDetails.ClientCertPath, configCommandConfiguration.ServerDetails.ClientCertKeyPath)
		if err != nil {
			return err
		}
	}

	// OIDC validation logic
	if configCommandConfiguration.OidcParams.ProviderName != "" {
		// Exchange token ID is injected by the CI wrapper plugin or provided manually by the user
		if os.Getenv(coreutils.OidcExchangeTokenId) == "" && configCommandConfiguration.OidcParams.TokenId == "" {
			return errorutils.CheckErrorf("the --oidc-token-id flag must be provided when --oidc-provider is used. Ensure the flag is set or the environment variable is exported. If running on a CI server, verify the token is correctly injected.")
		}
		if configCommandConfiguration.ServerDetails.Url == "" {
			return errorutils.CheckErrorf("the --url flag must be provided when --oidc-provider is used")
		}

	}

	return nil
}

func validatePathsExist(paths ...string) error {
	for _, path := range paths {
		if path != "" {
			exists, err := fileutils.IsFileExists(path, true)
			if err != nil {
				return err
			}
			if !exists {
				return errorutils.CheckErrorf("file does not exit at " + path)
			}
		}
	}
	return nil
}

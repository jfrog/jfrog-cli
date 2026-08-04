package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/jfrog/jfrog-cli-core/v2/general/token"

	"github.com/jfrog/jfrog-client-go/auth/cert"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/docs/config/add"
	"github.com/jfrog/jfrog-cli/docs/config/edit"
	"github.com/jfrog/jfrog-cli/docs/config/remove"
	"github.com/jfrog/jfrog-cli/docs/config/use"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
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
			Usage:        corecommon.ResolveDescription(add.GetDescription(), add.GetAIDescription()),
			Flags:        cliutils.GetCommandFlags(cliutils.AddConfig),
			HelpName:     corecommon.CreateUsage("c add", corecommon.ResolveDescription(add.GetDescription(), add.GetAIDescription()), add.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       addCmd,
		},
		{
			Name:         "edit",
			Usage:        corecommon.ResolveDescription(edit.GetDescription(), edit.GetAIDescription()),
			Flags:        cliutils.GetCommandFlags(cliutils.EditConfig),
			HelpName:     corecommon.CreateUsage("c edit", corecommon.ResolveDescription(edit.GetDescription(), edit.GetAIDescription()), edit.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       editCmd,
		},
		{
			Name:         "show",
			Aliases:      []string{"s"},
			Usage:        corecommon.ResolveDescription(show.GetDescription(), show.GetAIDescription()),
			Flags:        cliutils.GetCommandFlags(cliutils.ConfigShow),
			HelpName:     corecommon.CreateUsage("c show", corecommon.ResolveDescription(show.GetDescription(), show.GetAIDescription()), show.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       showCmd,
		},
		{
			Name:         "remove",
			Aliases:      []string{"rm"},
			Usage:        corecommon.ResolveDescription(remove.GetDescription(), remove.GetAIDescription()),
			Flags:        cliutils.GetCommandFlags(cliutils.DeleteConfig),
			HelpName:     corecommon.CreateUsage("c rm", corecommon.ResolveDescription(remove.GetDescription(), remove.GetAIDescription()), remove.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       deleteCmd,
		},
		{
			Name:         "import",
			Aliases:      []string{"im"},
			Usage:        corecommon.ResolveDescription(importcmd.GetDescription(), importcmd.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("c import", corecommon.ResolveDescription(importcmd.GetDescription(), importcmd.GetAIDescription()), importcmd.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       importCmd,
		},
		{
			Name:         "export",
			Aliases:      []string{"ex"},
			Usage:        corecommon.ResolveDescription(exportcmd.GetDescription(), exportcmd.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("c export", corecommon.ResolveDescription(exportcmd.GetDescription(), exportcmd.GetAIDescription()), exportcmd.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(commands.GetAllServerIds()...),
			Action:       exportCmd,
		},
		{
			Name:         "use",
			Usage:        corecommon.ResolveDescription(use.GetDescription(), use.GetAIDescription()),
			HelpName:     corecommon.CreateUsage("c use", corecommon.ResolveDescription(use.GetDescription(), use.GetAIDescription()), use.Usage),
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
		SetOIDCParams(configCommandConfiguration.OidcParams)

	return configCmd.Run()
}

func createOidcParamsFromFlags(c *cli.Context) (*token.OidcParams, error) {
	providerType, err := token.OidcProviderTypeFromString(cliutils.GetFlagOrEnvValue(c, cliutils.OidcProviderType, coreutils.OidcProviderType))
	if err != nil {
		return nil, err
	}
	return &token.OidcParams{
		ProviderType: providerType,
		ProviderName: c.String(cliutils.OidcProviderName),
		Audience:     c.String(cliutils.OidcAudience),
		// Values can be set by the user or injected by the CI wrapper plugin
		TokenId:        cliutils.GetFlagOrEnvValue(c, cliutils.OidcTokenID, coreutils.OidcExchangeTokenId),
		ProjectKey:     cliutils.GetFlagOrEnvValue(c, cliutils.Project, coreutils.Project),
		ApplicationKey: cliutils.GetFlagOrEnvValue(c, cliutils.ApplicationKey, coreutils.ApplicationKey),
		VcsUrl:         os.Getenv(coreutils.CIVcsUrl),
		VcsBranch:      os.Getenv(coreutils.CIVcsBranch),
		VcsRevision:    os.Getenv(coreutils.CIVcsRevision),
		// Values from the CI environment
		JobId: os.Getenv(coreutils.CIJobID),
		RunId: os.Getenv(coreutils.CIRunID),
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

	if !c.IsSet(cliutils.Format) {
		// No format flag — use existing behavior unchanged.
		return commands.ShowConfig(serverId)
	}

	outputFormat, err := commonCliUtils.GetOutputFormat(c, coreformat.Table)
	if err != nil {
		return err
	}

	// Read config directly so we can render it.
	var configs []*coreconfig.ServerDetails
	if serverId != "" {
		single, err := coreconfig.GetSpecificConfig(serverId, false, false)
		if err != nil {
			return err
		}
		configs = []*coreconfig.ServerDetails{single}
	} else {
		configs, err = coreconfig.GetAllServersConfigs()
		if err != nil {
			return err
		}
	}
	return printConfigShowResponse(configs, outputFormat, os.Stdout)
}

func printConfigShowResponse(configs []*coreconfig.ServerDetails, outputFormat coreformat.OutputFormat, w io.Writer) error {
	switch outputFormat {
	case coreformat.Json:
		return printConfigShowJSON(configs)
	case coreformat.Table:
		return printConfigShowTable(configs, w)
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for config show. Accepted values: table, json", outputFormat)
	}
}

// sanitizeServerDetails returns a copy with sensitive fields masked.
func sanitizeServerDetails(d *coreconfig.ServerDetails) coreconfig.ServerDetails {
	s := *d
	if s.Password != "" {
		s.Password = "***"
	}
	if s.SshPassphrase != "" {
		s.SshPassphrase = "***"
	}
	if s.AccessToken != "" {
		s.AccessToken = "***"
	}
	if s.RefreshToken != "" {
		s.RefreshToken = "***"
	}
	if s.ArtifactoryRefreshToken != "" {
		s.ArtifactoryRefreshToken = "***"
	}
	return s
}

func printConfigShowJSON(configs []*coreconfig.ServerDetails) error {
	sanitized := make([]coreconfig.ServerDetails, len(configs))
	for i, c := range configs {
		sanitized[i] = sanitizeServerDetails(c)
	}
	data, err := json.Marshal(sanitized) // #nosec G117
	if err != nil {
		return errorutils.CheckErrorf("failed to marshal config: %s", err.Error())
	}
	log.Output(clientUtils.IndentJson(data))
	return nil
}

func printConfigShowTable(configs []*coreconfig.ServerDetails, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, d := range configs {
		if i > 0 {
			_, _ = fmt.Fprintln(tw, "---")
		}
		_, _ = fmt.Fprintln(tw, "FIELD\tVALUE")
		printConfigRow := func(field, value string) {
			if value != "" {
				_, _ = fmt.Fprintf(tw, "%s\t%s\n", field, value)
			}
		}
		printConfigRow("server_id", d.ServerId)
		printConfigRow("url", d.Url)
		printConfigRow("artifactory_url", d.ArtifactoryUrl)
		printConfigRow("distribution_url", d.DistributionUrl)
		printConfigRow("xray_url", d.XrayUrl)
		printConfigRow("mission_control_url", d.MissionControlUrl)
		printConfigRow("pipelines_url", d.PipelinesUrl)
		printConfigRow("user", d.User)
		if d.Password != "" {
			printConfigRow("password", "***")
		}
		if d.AccessToken != "" {
			printConfigRow("access_token", "***")
		}
		if d.RefreshToken != "" {
			printConfigRow("refresh_token", "***")
		}
		printConfigRow("ssh_key_path", d.SshKeyPath)
		if d.SshPassphrase != "" {
			printConfigRow("ssh_passphrase", "***")
		}
		printConfigRow("client_cert_path", d.ClientCertPath)
		printConfigRow("client_cert_key_path", d.ClientCertKeyPath)
		if d.IsDefault {
			printConfigRow("is_default", "true")
		}
	}
	return tw.Flush()
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
	if slices.Contains(reservedIds, serverId) {
		return fmt.Errorf("server can't have one of the following IDs: %s\n%s", strings.Join(reservedIds, ", "), cliutils.GetDocumentationMessage())
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
			return errorutils.CheckErrorf("the --oidc-token-id flag or the JFROG_CLI_OIDC_EXCHANGE_TOKEN_ID environment variable must be provided when using --oidc-provider.")
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
				return errorutils.CheckErrorf("File doesn't exist at %s", path)
			}
		}
	}
	return nil
}

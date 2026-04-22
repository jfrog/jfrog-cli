package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	generic "github.com/jfrog/jfrog-cli-core/v2/general/token"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/accesstoken"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

func AccessTokenCreateCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	serverDetails, err := createPlatformDetailsByFlags(c)
	if err != nil {
		return err
	}

	if err = assertAccessTokenAvailable(serverDetails); err != nil {
		return err
	}

	if err = assertScopeOptions(c); err != nil {
		return err
	}

	username := accesstoken.GetSubjectUsername(c, serverDetails)

	expiry, err := getExpiry(c)
	if err != nil {
		return err
	}

	accessTokenCreateCmd := generic.NewAccessTokenCreateCommand()
	accessTokenCreateCmd.
		SetServerDetails(serverDetails).
		SetUsername(username).
		SetProjectKey(c.String(cliutils.Project)).
		SetGroups(c.String(cliutils.Groups)).
		SetScope(c.String(cliutils.Scope)).
		SetGrantAdmin(c.Bool(cliutils.GrantAdmin)).
		SetExpiry(expiry).
		SetRefreshable(c.Bool(cliutils.Refreshable)).
		SetDescription(c.String(cliutils.Description)).
		SetAudience(c.String(cliutils.Audience)).
		SetIncludeReferenceToken(c.Bool(cliutils.Reference))
	err = commands.Exec(accessTokenCreateCmd)
	if err != nil {
		return err
	}
	resBytes, err := accessTokenCreateCmd.Response()
	if err != nil {
		return err
	}
	outputFormat, err := getTokenOutputFormat(c)
	if err != nil {
		return err
	}
	return printTokenResponse(resBytes, outputFormat, os.Stdout)
}

// getTokenOutputFormat returns the requested output format, defaulting to json
// to preserve backward-compatible behaviour (the command always emitted JSON).
func getTokenOutputFormat(c *cli.Context) (coreformat.OutputFormat, error) {
	if !c.IsSet(cliutils.Format) {
		return coreformat.Json, nil
	}
	return coreformat.GetOutputFormat(c.String(cliutils.Format))
}

// printTokenResponse writes the token response in the requested format to w.
func printTokenResponse(data []byte, outputFormat coreformat.OutputFormat, w io.Writer) error {
	switch outputFormat {
	case coreformat.Json:
		log.Output(clientUtils.IndentJson(data))
		return nil
	case coreformat.Table:
		return printTokenTable(data, w)
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for access-token-create. Accepted values: table, json", outputFormat)
	}
}

// printTokenTable renders the token fields as a plain two-column table.
// The access_token value is truncated to avoid flooding the terminal.
func printTokenTable(data []byte, w io.Writer) error {
	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return errorutils.CheckErrorf("failed to parse token response: %s", err.Error())
	}

	// Print fields in a stable, human-friendly order; skip absent/nil ones.
	orderedKeys := []string{
		"access_token", "token_id", "expires_in", "scope",
		"token_type", "refreshable", "refresh_token", "reference_token",
		"grant_type", "audience",
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "FIELD\tVALUE")
	for _, key := range orderedKeys {
		val, ok := fields[key]
		if !ok || val == nil {
			continue
		}
		strVal := fmt.Sprintf("%v", val)
		if key == "access_token" && len(strVal) > 40 {
			strVal = strVal[:40] + "..."
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", key, strVal)
	}
	return tw.Flush()
}

func ExchangeOidcTokenCmd(c *cli.Context) error {
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	serverDetails, err := createPlatformDetailsByFlags(c)
	if err != nil {
		return err
	}

	oidcAccessTokenCreateCmd, err := CreateOidcTokenExchangeCommand(c, serverDetails)
	if err != nil {
		return err
	}

	if err = commands.ExecAndThenReportUsage(oidcAccessTokenCreateCmd); err != nil {
		return err
	}

	outputFormat, err := getOidcTokenOutputFormat(c)
	if err != nil {
		return err
	}
	return printOidcTokenResponse(oidcAccessTokenCreateCmd.Response(), outputFormat, os.Stdout)
}

// getOidcTokenOutputFormat returns the requested output format, defaulting to json
// to preserve backward-compatible behaviour (the command always emitted JSON).
func getOidcTokenOutputFormat(c *cli.Context) (coreformat.OutputFormat, error) {
	if !c.IsSet(cliutils.Format) {
		return coreformat.Json, nil
	}
	return coreformat.GetOutputFormat(c.String(cliutils.Format))
}

// printOidcTokenResponse writes the OIDC exchange result in the requested format to w.
func printOidcTokenResponse(response *auth.OidcTokenResponseData, outputFormat coreformat.OutputFormat, w io.Writer) error {
	data, err := json.Marshal(response)
	if err != nil {
		return errorutils.CheckErrorf("failed to marshal OIDC token response: %s", err.Error())
	}
	switch outputFormat {
	case coreformat.Json:
		log.Output(clientUtils.IndentJson(data))
		return nil
	case coreformat.Table:
		return printOidcTokenTable(response, w)
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for exchange-oidc-token. Accepted values: table, json", outputFormat)
	}
}

// printOidcTokenTable renders the OIDC token fields as a plain two-column table.
// The access token value is truncated to avoid flooding the terminal.
func printOidcTokenTable(response *auth.OidcTokenResponseData, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "FIELD\tVALUE")
	if response.AccessToken != "" {
		val := response.AccessToken
		if len(val) > 40 {
			val = val[:40] + "..."
		}
		_, _ = fmt.Fprintf(tw, "access_token\t%s\n", val)
	}
	if response.Username != "" {
		_, _ = fmt.Fprintf(tw, "username\t%s\n", response.Username)
	}
	if response.IssuedTokenType != "" {
		_, _ = fmt.Fprintf(tw, "issued_token_type\t%s\n", response.IssuedTokenType)
	}
	if response.Scope != "" {
		_, _ = fmt.Fprintf(tw, "scope\t%s\n", response.Scope)
	}
	if response.TokenType != "" {
		_, _ = fmt.Fprintf(tw, "token_type\t%s\n", response.TokenType)
	}
	return tw.Flush()
}

func CreateOidcTokenExchangeCommand(c *cli.Context, serverDetails *coreConfig.ServerDetails) (*generic.OidcTokenExchangeCommand, error) {
	oidcAccessTokenCreateCmd := generic.NewOidcTokenExchangeCommand()
	// Validate supported oidc provider type
	if err := oidcAccessTokenCreateCmd.SetProviderTypeAsString(cliutils.GetFlagOrEnvValue(c, cliutils.OidcProviderType, coreutils.OidcProviderType)); err != nil {
		return nil, err
	}

	oidcAccessTokenCreateCmd.
		SetServerDetails(serverDetails).
		// Mandatory flags
		SetProviderName(c.Args().Get(0)).
		SetOidcTokenID(getOidcTokenIdInput(c)).
		SetAudience(c.String(cliutils.OidcAudience)).
		// Optional values exported by CI servers
		SetJobId(os.Getenv(coreutils.CIJobID)).
		SetRunId(os.Getenv(coreutils.CIRunID)).
		SetVcsRevision(os.Getenv(coreutils.CIVcsRevision)).
		SetVcsUrl(os.Getenv(coreutils.CIVcsUrl)).
		SetVcsBranch(os.Getenv(coreutils.CIVcsBranch)).
		// Values which can both be exported or explicitly set
		SetProjectKey(cliutils.GetFlagOrEnvValue(c, cliutils.Project, coreutils.Project)).
		SetApplicationKey(cliutils.GetJFrogApplicationKey(c))

	return oidcAccessTokenCreateCmd, nil
}

func createPlatformDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	platformDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return nil, err
	}
	if platformDetails.Url == "" {
		return nil, errors.New("no JFrog Platform URL specified, either via the --url flag or as part of the server configuration")
	}
	return platformDetails, nil
}

func getExpiry(c *cli.Context) (*uint, error) {
	if !c.IsSet(cliutils.Expiry) {
		return nil, nil
	}

	expiryInt, err := strconv.Atoi(c.String(cliutils.Expiry))
	if err != nil {
		return nil, cliutils.PrintHelpAndReturnError(
			fmt.Sprintf("The '--%s' option must have a numeric value. ", cliutils.Expiry), c)
	}
	if expiryInt < 0 {
		return nil, cliutils.PrintHelpAndReturnError(
			fmt.Sprintf("The '--%s' option must be non-negative. ", cliutils.Expiry), c)
	}
	expiry := uint(expiryInt)
	return &expiry, nil
}

func assertScopeOptions(c *cli.Context) error {
	if c.IsSet(cliutils.Scope) && (c.IsSet(cliutils.GrantAdmin) || c.IsSet(cliutils.Groups)) {
		return cliutils.PrintHelpAndReturnError(
			fmt.Sprintf("Scope can either be provided explicitly with '--%s', or implicitly with '--%s' and '--%s'. ",
				cliutils.Scope, cliutils.GrantAdmin, cliutils.Groups), c)
	}
	return nil
}

func assertAccessTokenAvailable(serverDetails *coreConfig.ServerDetails) error {
	if serverDetails.AccessToken == "" {
		return errorutils.CheckErrorf("authenticating with access token is currently mandatory for creating access tokens")
	}
	return nil
}

// The OIDC token ID can be provided as a command line argument or as a flag or environment variable.
// Depends on the origin of the command.
// For example when used in CI/CD, the token ID is provided as a environment variable.
func getOidcTokenIdInput(c *cli.Context) string {
	oidcTokenId := c.Args().Get(1)
	if oidcTokenId == "" {
		oidcTokenId = cliutils.GetFlagOrEnvValue(c, cliutils.OidcTokenID, coreutils.OidcExchangeTokenId)
	}
	return oidcTokenId
}

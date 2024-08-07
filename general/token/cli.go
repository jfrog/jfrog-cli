package token

import (
	"errors"
	"fmt"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	generic "github.com/jfrog/jfrog-cli-core/v2/general/token"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/accesstoken"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"strconv"
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
	resString, err := accessTokenCreateCmd.Response()
	if err != nil {
		return err
	}
	log.Output(clientUtils.IndentJson(resString))

	return nil
}

func createPlatformDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	platformDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return nil, err
	}
	if platformDetails.Url == "" {
		return nil, errors.New("JFrog Platform URL is mandatory for access token creation")
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

package token

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type RefreshTokenCommand struct {
	CreateTokenCommand
	refreshToken string
	accessToken  string
}

func NewRefreshTokenCommand() *RefreshTokenCommand {
	return &RefreshTokenCommand{CreateTokenCommand: *NewCreateTokenCommand()}
}

func (rt *RefreshTokenCommand) CommandName() string {
	return "rt_refresh_token"
}

func (rt *RefreshTokenCommand) Result() CreateTokenResult {
	return rt.result
}

func (rt *RefreshTokenCommand) SetAccessToken(accessToken string) *RefreshTokenCommand {
	rt.accessToken = accessToken
	return rt
}

func (rt *RefreshTokenCommand) SetRefreshToken(refreshToken string) *RefreshTokenCommand {
	rt.refreshToken = refreshToken
	return rt
}

func (rt *RefreshTokenCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(rt.rtDetails, false)
	if err != nil {
		return err
	}
	params := services.NewRefreshTokenParams()
	params.Token.Audience = rt.audience
	params.Token.Scope = rt.scope
	params.Token.Refreshable = rt.refreshable
	params.Token.ExpiresIn = rt.expiresIn
	params.Token.Username = rt.username
	params.AccessToken = rt.accessToken
	params.RefreshToken = rt.refreshToken
	result, err := servicesManager.RefreshToken(params)
	if err != nil {
		return err
	}
	rt.result = CreateTokenResult{
		Scope:        result.Scope,
		AccessToken:  result.AccessToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
		RefreshToken: result.RefreshToken,
	}
	return err
}

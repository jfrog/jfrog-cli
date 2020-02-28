package token

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type RevokeTokenCommand struct {
	TokenCommand
	token   string
	tokenID string
	result  string
}

func NewRevokeTokenCommand() *RevokeTokenCommand {
	return &RevokeTokenCommand{TokenCommand: *NewTokenCommand()}
}

func (rt *RevokeTokenCommand) CommandName() string {
	return "rt_revoke_token"
}

func (rt *RevokeTokenCommand) Result() string {
	return rt.result
}

func (rt *RevokeTokenCommand) SetToken(token string) *RevokeTokenCommand {
	rt.token = token
	return rt
}

func (rt *RevokeTokenCommand) SetTokenID(tokenID string) *RevokeTokenCommand {
	rt.tokenID = tokenID
	return rt
}

func (rt *RevokeTokenCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(rt.rtDetails, false)
	if err != nil {
		return err
	}
	params := services.NewRevokeTokenParams()
	params.TokenId = rt.tokenID
	params.Token = rt.token
	result, err := servicesManager.RevokeToken(params)
	if err != nil {
		return err
	}
	rt.result = result
	return err
}

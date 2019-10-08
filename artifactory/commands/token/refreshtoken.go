package token

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type RefreshTokenCommand struct {
	TokenCommand
	params services.RefreshTokenParams
	result services.CreateTokenResponseData
}

func NewRefreshTokenCommand() *RefreshTokenCommand {
	return &RefreshTokenCommand{TokenCommand: *NewTokenCommand()}
}

func (rt *RefreshTokenCommand) CommandName() string {
	return "rt_refresh_token"
}

func (rt *RefreshTokenCommand) Result() services.CreateTokenResponseData {
	return rt.result
}

func (rt *RefreshTokenCommand) SetParams(params services.RefreshTokenParams) {
	rt.params = params
}

func (rt *RefreshTokenCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(rt.rtDetails, false)
	if err != nil {
		return err
	}
	result, err := servicesManager.RefreshToken(rt.params)
	if err != nil {
		return err
	}
	rt.result = result
	return err
}

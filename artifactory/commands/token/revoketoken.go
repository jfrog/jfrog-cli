package token

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type RevokeTokenCommand struct {
	TokenCommand
	params    services.RevokeTokenParams
	result   string
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

func (rt *RevokeTokenCommand) SetParams(params services.RevokeTokenParams) {
	rt.params = params
}

func (rt *RevokeTokenCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(rt.rtDetails, false)
	if err != nil {
		return err
	}
	result, err := servicesManager.RevokeToken(rt.params)
	if err != nil {
		return err
	}
	rt.result = result
	return err
}

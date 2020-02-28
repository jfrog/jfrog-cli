package token

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type CreateTokenCommand struct {
	TokenCommand
	params services.CreateTokenParams
	result services.CreateTokenResponseData
}

func NewCreateTokenCommand() *CreateTokenCommand {
	return &CreateTokenCommand{TokenCommand: *NewTokenCommand()}
}

func (ct *CreateTokenCommand) Result() services.CreateTokenResponseData {
	return ct.result
}

func (ct *CreateTokenCommand) CommandName() string {
	return "rt_create_token"
}

func (ct *CreateTokenCommand) SetParams(params services.CreateTokenParams) {
	ct.params = params
}

func (ct *CreateTokenCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(ct.rtDetails, false)
	if err != nil {
		return err
	}
	// If username is not provided, use the current user from configuration or the `--user` flag
	if ct.params.Username == "" {
		ct.params.Username = servicesManager.GetConfig().GetArtDetails().GetUser()
	}
	result, err := servicesManager.CreateToken(ct.params)
	if err != nil {
		return err
	}
	ct.result = result
	return err
}

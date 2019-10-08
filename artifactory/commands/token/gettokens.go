package token

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type GetTokensCommand struct {
	TokenCommand
	result services.GetTokensResponseData
}

func NewGetTokensCommand() *GetTokensCommand {
	return &GetTokensCommand{TokenCommand: *NewTokenCommand()}
}

func (gt *GetTokensCommand) Result() services.GetTokensResponseData {
	return gt.result
}

func (gt *GetTokensCommand) CommandName() string {
	return "rt_get_tokens"
}

func (gt *GetTokensCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(gt.rtDetails, false)
	if err != nil {
		return err
	}
	resultItems, err := servicesManager.GetTokens()
	if err != nil {
		return err
	}
	gt.result = resultItems
	return err
}

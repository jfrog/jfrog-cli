package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type GetTokensCommand struct {
	GenericCommand
	results services.GetTokensResponseData
}

func NewGetTokensCommand() *GetTokensCommand {
	return &GetTokensCommand{GenericCommand: *NewGenericCommand()}
}

func (gt *GetTokensCommand) Results() services.GetTokensResponseData {
	return gt.results
}

func (gt *GetTokensCommand) CommandName() string {
	return "rt_get_tokens"
}

func (gt *GetTokensCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(gt.rtDetails, gt.dryRun)
	if err != nil {
		return err
	}
	resultItems, err := servicesManager.GetTokens()
	if err != nil {
		return err
	}
	gt.results = resultItems
	return err
}

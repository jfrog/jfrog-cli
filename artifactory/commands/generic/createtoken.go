package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type CreateTokenCommand struct {
	GenericCommand
	params  services.CreateTokenParams
	results services.CreateTokenResponseData
}

func NewCreateTokenCommand() *CreateTokenCommand {
	return &CreateTokenCommand{GenericCommand: *NewGenericCommand()}
}

func (ct *CreateTokenCommand) Results() services.CreateTokenResponseData {
	return ct.results
}

func (ct *CreateTokenCommand) CommandName() string {
	return "rt_create_token"
}

func (ct *CreateTokenCommand) SetParams(params services.CreateTokenParams) {
	ct.params = params
}

func (ct *CreateTokenCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(ct.rtDetails, ct.dryRun)
	if err != nil {
		return err
	}
	results, err := servicesManager.CreateToken(ct.params)
	if err != nil {
		return err
	}
	ct.results = results
	return err
}

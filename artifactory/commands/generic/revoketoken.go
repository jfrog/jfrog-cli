package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type RevokeTokenCommand struct {
	GenericCommand
	params  services.RevokeTokenParams
	results string
}

func NewRevokeTokenCommand() *RevokeTokenCommand {
	return &RevokeTokenCommand{GenericCommand: *NewGenericCommand()}
}

func (ct *RevokeTokenCommand) CommandName() string {
	return "rt_revoke_token"
}

func (ct *RevokeTokenCommand) Results() string {
	return ct.results
}

func (ct *RevokeTokenCommand) SetParams(params services.RevokeTokenParams) {
	ct.params = params
}

func (ct *RevokeTokenCommand) Run() error {
	var responseText string
	servicesManager, err := utils.CreateServiceManager(ct.rtDetails, ct.dryRun)
	if err != nil {
		return err
	}
	responseText, err = servicesManager.RevokeToken(ct.params)
	if err != nil {
		return err
	}
	ct.results = responseText
	return err
}

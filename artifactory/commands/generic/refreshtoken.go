package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type RefreshTokenCommand struct {
	GenericCommand
	params  services.RefreshTokenParams
	results services.CreateTokenResponseData
}

func NewRefreshTokenCommand() *RefreshTokenCommand {
	return &RefreshTokenCommand{GenericCommand: *NewGenericCommand()}
}

func (ct *RefreshTokenCommand) CommandName() string {
	return "rt_refresh_token"
}

func (ct *RefreshTokenCommand) Results() services.CreateTokenResponseData {
	return ct.results
}

func (ct *RefreshTokenCommand) SetParams(params services.RefreshTokenParams) {
	ct.params = params
}

func (ct *RefreshTokenCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(ct.rtDetails, ct.dryRun)
	if err != nil {
		return err
	}
	results, err := servicesManager.RefreshToken(ct.params)
	if err != nil {
		return err
	}
	ct.results = results
	return err
}

package token

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type CreateTokenCommand struct {
	TokenCommand
	CreateTokenParams
	result      CreateTokenResult

}

type CreateTokenParams struct {
	scope       string
	username    string
	expiresIn   int
	refreshable bool
	audience    string
}

type CreateTokenResult struct {
	Scope        string `json:"scope,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (ct *CreateTokenCommand) SetScope(scope string) *CreateTokenCommand {
	ct.scope = scope
	return ct
}

func (ct *CreateTokenCommand) SetUsername(username string) *CreateTokenCommand {
	ct.username = username
	return ct
}

func (ct *CreateTokenCommand) SetExpiresIn(expiresIn int) *CreateTokenCommand {
	ct.expiresIn = expiresIn
	return ct
}

func (ct *CreateTokenCommand) SetRefreshable(refreshable bool) *CreateTokenCommand {
	ct.refreshable = refreshable
	return ct
}

func (ct *CreateTokenCommand) SetAudience(audience string) *CreateTokenCommand {
	ct.audience = audience
	return ct
}

func NewCreateTokenCommand() *CreateTokenCommand {
	return &CreateTokenCommand{TokenCommand: *NewTokenCommand()}
}

func (ct *CreateTokenCommand) Result() CreateTokenResult {
	return ct.result
}

func (ct *CreateTokenCommand) CommandName() string {
	return "rt_create_token"
}

func (ct *CreateTokenCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(ct.rtDetails, false)
	if err != nil {
		return err
	}
	params := services.NewCreateTokenParams()
	params.Audience = ct.audience
	params.Scope = ct.scope
	params.Refreshable = ct.refreshable
	params.ExpiresIn = ct.expiresIn
	params.Username = ct.username
	result, err := servicesManager.CreateToken(params)
	if err != nil {
		return err
	}
	ct.result = CreateTokenResult{
		Scope:        result.Scope,
		AccessToken:  result.AccessToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
		RefreshToken: result.RefreshToken,
	}
	return err
}

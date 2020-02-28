package token

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
)

type GetTokensCommand struct {
	TokenCommand
	result GetTokensResult
}

type GetTokensResult struct {
	Tokens []struct {
		Issuer      string `json:"issuer,omitempty"`
		Subject     string `json:"subject,omitempty"`
		Refreshable bool   `json:"refreshable,omitempty"`
		Expiry      int    `json:"expiry,omitempty"`
		TokenId     string `json:"token_id,omitempty"`
		IssuedAt    int    `json:"issued_at,omitempty"`
	}
}

func NewGetTokensCommand() *GetTokensCommand {
	return &GetTokensCommand{TokenCommand: *NewTokenCommand()}
}

func (gt *GetTokensCommand) Result() GetTokensResult {
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
	gt.result = GetTokensResult{Tokens: resultItems.Tokens}
	return err
}

package token

import "github.com/jfrog/jfrog-cli-go/utils/config"

type TokenCommand struct {
	rtDetails *config.ArtifactoryDetails
}

func NewTokenCommand() *TokenCommand {
	return &TokenCommand{}
}

func (tc *TokenCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return tc.rtDetails, nil
}

func (tc *TokenCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *TokenCommand {
	tc.rtDetails = rtDetails
	return tc
}

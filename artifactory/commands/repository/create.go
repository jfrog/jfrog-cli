package repository

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
)

type CreateRepoCommand struct {
	RepoCommand
}

func NewCreateRepoCommand() *CreateRepoCommand {
	return &CreateRepoCommand{}
}

func (rcc *CreateRepoCommand) SetTemplatePath(path string) *CreateRepoCommand {
	rcc.templatePath = path
	return rcc
}

func (rcc *CreateRepoCommand) SetVars(vars string) *CreateRepoCommand {
	rcc.vars = vars
	return rcc
}

func (rcc *CreateRepoCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *CreateRepoCommand {
	rcc.rtDetails = rtDetails
	return rcc
}

func (rcc *CreateRepoCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return rcc.rtDetails, nil
}

func (rcc *CreateRepoCommand) CommandName() string {
	return "rt_repo_create"
}

func (rcc *CreateRepoCommand) Run() (err error) {
	return rcc.PerformRepoCmd(false)
}

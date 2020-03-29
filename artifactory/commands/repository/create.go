package repository

import (
	"github.com/jfrog/jfrog-cli/utils/config"
)

type RepoCreateCommand struct {
	RepoCommand
}

func NewRepoCreateCommand() *RepoCreateCommand {
	return &RepoCreateCommand{}
}

func (rcc *RepoCreateCommand) SetTemplatePath(path string) *RepoCreateCommand {
	rcc.templatePath = path
	return rcc
}

func (rcc *RepoCreateCommand) SetVars(vars string) *RepoCreateCommand {
	rcc.vars = vars
	return rcc
}

func (rcc *RepoCreateCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *RepoCreateCommand {
	rcc.rtDetails = rtDetails
	return rcc
}

func (rcc *RepoCreateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return rcc.rtDetails, nil
}

func (rcc *RepoCreateCommand) CommandName() string {
	return "rt_repo_create"
}

func (rcc *RepoCreateCommand) Run() (err error) {
	return rcc.PerformRepoCmd(false)
}

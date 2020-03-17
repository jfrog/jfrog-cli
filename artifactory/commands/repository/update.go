package repository

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
)

type UpdateRepoCommand struct {
	RepoCommand
}

func NewUpdateRepoCommand() *UpdateRepoCommand {
	return &UpdateRepoCommand{}
}

func (ruc *UpdateRepoCommand) SetTemplatePath(path string) *UpdateRepoCommand {
	ruc.templatePath = path
	return ruc
}

func (ruc *UpdateRepoCommand) SetVars(vars string) *UpdateRepoCommand {
	ruc.vars = vars
	return ruc
}

func (ruc *UpdateRepoCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *UpdateRepoCommand {
	ruc.rtDetails = rtDetails
	return ruc
}

func (ruc *UpdateRepoCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return ruc.rtDetails, nil
}

func (ruc *UpdateRepoCommand) CommandName() string {
	return "rt_repo_update"
}

func (ruc *UpdateRepoCommand) Run() (err error) {
	return ruc.PerformRepoCmd(true)
}

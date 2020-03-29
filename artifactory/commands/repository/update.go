package repository

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
)

type RepoUpdateCommand struct {
	RepoCommand
}

func NewRepoUpdateCommand() *RepoUpdateCommand {
	return &RepoUpdateCommand{}
}

func (ruc *RepoUpdateCommand) SetTemplatePath(path string) *RepoUpdateCommand {
	ruc.templatePath = path
	return ruc
}

func (ruc *RepoUpdateCommand) SetVars(vars string) *RepoUpdateCommand {
	ruc.vars = vars
	return ruc
}

func (ruc *RepoUpdateCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *RepoUpdateCommand {
	ruc.rtDetails = rtDetails
	return ruc
}

func (ruc *RepoUpdateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return ruc.rtDetails, nil
}

func (ruc *RepoUpdateCommand) CommandName() string {
	return "rt_repo_update"
}

func (ruc *RepoUpdateCommand) Run() (err error) {
	return ruc.PerformRepoCmd(true)
}

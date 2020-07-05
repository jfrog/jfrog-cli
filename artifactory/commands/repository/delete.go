package repository

import (
	rtUtils "github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
)

type RepoDeleteCommand struct {
	rtDetails *config.ArtifactoryDetails
	repoKey   string
	quiet     bool
}

func NewRepoDeleteCommand() *RepoDeleteCommand {
	return &RepoDeleteCommand{}
}

func (rdc *RepoDeleteCommand) SetRepoKey(repoKey string) *RepoDeleteCommand {
	rdc.repoKey = repoKey
	return rdc
}

func (rdc *RepoDeleteCommand) SetQuiet(quiet bool) *RepoDeleteCommand {
	rdc.quiet = quiet
	return rdc
}

func (rdc *RepoDeleteCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *RepoDeleteCommand {
	rdc.rtDetails = rtDetails
	return rdc
}

func (rdc *RepoDeleteCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return rdc.rtDetails, nil
}

func (rdc *RepoDeleteCommand) CommandName() string {
	return "rt_repo_delete"
}

func (rdc *RepoDeleteCommand) Run() (err error) {
	if !rdc.quiet && !cliutils.InteractiveConfirm("Are you sure you want to permanently delete the repository "+rdc.repoKey+" including all of it content?", false) {
		return nil
	}
	servicesManager, err := rtUtils.CreateServiceManager(rdc.rtDetails, false)
	if err != nil {
		return err
	}
	return servicesManager.DeleteRepository(rdc.repoKey)
}

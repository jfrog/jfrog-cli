package replication

import (
	rtUtils "github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
)

type ReplicationDeleteCommand struct {
	rtDetails *config.ArtifactoryDetails
	repoKey   string
	quiet     bool
}

func NewReplicationDeleteCommand() *ReplicationDeleteCommand {
	return &ReplicationDeleteCommand{}
}

func (rdc *ReplicationDeleteCommand) SetRepoKey(repoKey string) *ReplicationDeleteCommand {
	rdc.repoKey = repoKey
	return rdc
}

func (rdc *ReplicationDeleteCommand) SetQuiet(quiet bool) *ReplicationDeleteCommand {
	rdc.quiet = quiet
	return rdc
}

func (rdc *ReplicationDeleteCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *ReplicationDeleteCommand {
	rdc.rtDetails = rtDetails
	return rdc
}

func (rdc *ReplicationDeleteCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return rdc.rtDetails, nil
}

func (rdc *ReplicationDeleteCommand) CommandName() string {
	return "rt_replication_delete"
}

func (rdc *ReplicationDeleteCommand) Run() (err error) {
	if !rdc.quiet && !cliutils.AskYesNo("Are you sure you want to delete the replication for  "+rdc.repoKey+" ?", false) {
		return nil
	}
	servicesManager, err := rtUtils.CreateServiceManager(rdc.rtDetails, false)
	return servicesManager.DeleteReplication(rdc.repoKey)
}

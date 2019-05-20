package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type BuildDiscardCommand struct {
	rtDetails *config.ArtifactoryDetails
	services.DiscardBuildsParams
}

func NewBuildDiscardCommand() *BuildDiscardCommand {
	return &BuildDiscardCommand{}
}

func (buildDiscard *BuildDiscardCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *BuildDiscardCommand {
	buildDiscard.rtDetails = rtDetails
	return buildDiscard
}

func (buildDiscard *BuildDiscardCommand) SetDiscardBuildsParams(params services.DiscardBuildsParams) *BuildDiscardCommand {
	buildDiscard.DiscardBuildsParams = params
	return buildDiscard
}

func (buildDiscard *BuildDiscardCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(buildDiscard.rtDetails, false)
	if err != nil {
		return err
	}
	return servicesManager.DiscardBuilds(buildDiscard.DiscardBuildsParams)
}

func (buildDiscard *BuildDiscardCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return buildDiscard.rtDetails, nil
}

func (buildDiscard *BuildDiscardCommand) CommandName() string {
	return "rt_build_discard"
}

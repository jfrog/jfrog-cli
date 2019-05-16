package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type BuildDiscardConfigurationCommand struct {
	rtDetails *config.ArtifactoryDetails
	services.DiscardBuildsParams
}

func (bdc *BuildDiscardConfigurationCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *BuildDiscardConfigurationCommand {
	bdc.rtDetails = rtDetails
	return bdc
}

func (bdc *BuildDiscardConfigurationCommand) SetDiscardBuildsParams(params services.DiscardBuildsParams) *BuildDiscardConfigurationCommand {
	bdc.DiscardBuildsParams = params
	return bdc
}

func (bdc *BuildDiscardConfigurationCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(bdc.rtDetails, false)
	if err != nil {
		return err
	}
	return servicesManager.DiscardBuilds(bdc.DiscardBuildsParams)
}

func (bdc *BuildDiscardConfigurationCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return bdc.rtDetails, nil
}

func (bdc *BuildDiscardConfigurationCommand) CommandName() string {
	return "rt_build_discard"
}

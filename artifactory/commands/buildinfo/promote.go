package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type BuildPromotionCommand struct {
	services.PromotionParams
	rtDetails *config.ArtifactoryDetails
	dryRun    bool
}

func (bpc *BuildPromotionCommand) SetDryRun(dryRun bool) *BuildPromotionCommand {
	bpc.dryRun = dryRun
	return bpc
}

func (bpc *BuildPromotionCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *BuildPromotionCommand {
	bpc.rtDetails = rtDetails
	return bpc
}

func (bpc *BuildPromotionCommand) SetPromotionParams(params services.PromotionParams) *BuildPromotionCommand {
	bpc.PromotionParams = params
	return bpc
}

func (bpc *BuildPromotionCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(bpc.rtDetails, bpc.dryRun)
	if err != nil {
		return err
	}
	return servicesManager.PromoteBuild(bpc.PromotionParams)
}

func (bpc *BuildPromotionCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return bpc.rtDetails, nil
}

func (bpc *BuildPromotionCommand) CommandName() string {
	return "rt_build_promote"
}

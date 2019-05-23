package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

type BuildDistributeCommnad struct {
	rtDetails *config.ArtifactoryDetails
	services.BuildDistributionParams
	dryRun bool
}

func NewBuildDistributeCommnad() *BuildDistributeCommnad {
	return &BuildDistributeCommnad{}
}

func (bdc *BuildDistributeCommnad) SetRtDetails(rtDetails *config.ArtifactoryDetails) *BuildDistributeCommnad {
	bdc.rtDetails = rtDetails
	return bdc
}

func (bdc *BuildDistributeCommnad) SetDryRun(dryRun bool) *BuildDistributeCommnad {
	bdc.dryRun = dryRun
	return bdc
}

func (bdc *BuildDistributeCommnad) SetBuildDistributionParams(buildDistributeParams services.BuildDistributionParams) *BuildDistributeCommnad {
	bdc.BuildDistributionParams = buildDistributeParams
	return bdc
}

func (bdc *BuildDistributeCommnad) Run() error {
	servicesManager, err := utils.CreateServiceManager(bdc.rtDetails, bdc.dryRun)
	if err != nil {
		return err
	}
	return servicesManager.DistributeBuild(bdc.BuildDistributionParams)
}

func (bdc *BuildDistributeCommnad) RtDetails() (*config.ArtifactoryDetails, error) {
	return bdc.rtDetails, nil
}

func (bdc *BuildDistributeCommnad) CommandName() string {
	return "rt_build_distribute"
}

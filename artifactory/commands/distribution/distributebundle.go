package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
)

type DistributeBundleCommand struct {
	rtDetails               *config.ArtifactoryDetails
	distributeBundlesParams services.DistributionParams
	spec                    *spec.DistributionSpecs
	dryRun                  bool
}

func NewDistributeBundleCommand() *DistributeBundleCommand {
	return &DistributeBundleCommand{}
}

func (distributeBundle *DistributeBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DistributeBundleCommand {
	distributeBundle.rtDetails = rtDetails
	return distributeBundle
}

func (distributeBundle *DistributeBundleCommand) SetDistributeBundleParams(params services.DistributionParams) *DistributeBundleCommand {
	distributeBundle.distributeBundlesParams = params
	return distributeBundle
}

func (distributeBundle *DistributeBundleCommand) SetSpec(spec *spec.DistributionSpecs) *DistributeBundleCommand {
	distributeBundle.spec = spec
	return distributeBundle
}

func (distributeBundle *DistributeBundleCommand) SetDryRun(dryRun bool) *DistributeBundleCommand {
	distributeBundle.dryRun = dryRun
	return distributeBundle
}

func (distributeBundle *DistributeBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(distributeBundle.rtDetails, distributeBundle.dryRun)
	if err != nil {
		return err
	}

	for _, spec := range distributeBundle.spec.Specs {
		distributeBundle.distributeBundlesParams.DistributionSpecs = append(distributeBundle.distributeBundlesParams.DistributionSpecs, spec.ToDistributionCommonParams())
	}

	return servicesManager.DistributeReleaseBundle(distributeBundle.distributeBundlesParams)
}

func (distributeBundle *DistributeBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return distributeBundle.rtDetails, nil
}

func (distributeBundle *DistributeBundleCommand) CommandName() string {
	return "rt_distribute_bundle"
}

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
	distributionRules       *spec.DistributionRules
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

func (distributeBundle *DistributeBundleCommand) SetDistributionRules(distributionRules *spec.DistributionRules) *DistributeBundleCommand {
	distributeBundle.distributionRules = distributionRules
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

	for _, rule := range distributeBundle.distributionRules.DistributionRules {
		distributeBundle.distributeBundlesParams.DistributionRules = append(distributeBundle.distributeBundlesParams.DistributionRules, rule.ToDistributionCommonParams())
	}

	return servicesManager.DistributeReleaseBundle(distributeBundle.distributeBundlesParams)
}

func (distributeBundle *DistributeBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return distributeBundle.rtDetails, nil
}

func (distributeBundle *DistributeBundleCommand) CommandName() string {
	return "rt_distribute_bundle"
}

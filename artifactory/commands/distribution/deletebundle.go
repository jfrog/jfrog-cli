package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
)

type DeleteBundleCommand struct {
	rtDetails               *config.ArtifactoryDetails
	deleteBundlesParams services.DeleteDistributionParams
	distributionRules       *spec.DistributionRules
	dryRun                  bool
}

func NewDeleteBundleCommand() *DeleteBundleCommand {
	return &DeleteBundleCommand{}
}

func (distributeBundle *DeleteBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DeleteBundleCommand {
	distributeBundle.rtDetails = rtDetails
	return distributeBundle
}

func (distributeBundle *DeleteBundleCommand) SetDistributeBundleParams(params services.DeleteDistributionParams) *DeleteBundleCommand {
	distributeBundle.deleteBundlesParams = params
	return distributeBundle
}

func (distributeBundle *DeleteBundleCommand) SetDistributionRules(distributionRules *spec.DistributionRules) *DeleteBundleCommand {
	distributeBundle.distributionRules = distributionRules
	return distributeBundle
}

func (distributeBundle *DeleteBundleCommand) SetDryRun(dryRun bool) *DeleteBundleCommand {
	distributeBundle.dryRun = dryRun
	return distributeBundle
}

func (distributeBundle *DeleteBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(distributeBundle.rtDetails, distributeBundle.dryRun)
	if err != nil {
		return err
	}

	for _, spec := range distributeBundle.distributionRules.DistributionRules {
		distributeBundle.deleteBundlesParams.DistributionRules = append(distributeBundle.deleteBundlesParams.DistributionRules, spec.ToDistributionCommonParams())
	}

	return servicesManager.DeleteReleaseBundle(distributeBundle.deleteBundlesParams)
}

func (distributeBundle *DeleteBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return distributeBundle.rtDetails, nil
}

func (distributeBundle *DeleteBundleCommand) CommandName() string {
	return "rt_delete_bundle"
}

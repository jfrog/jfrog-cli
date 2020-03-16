package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
	distributionServicesUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
)

type UpdateBundleCommand struct {
	rtDetails            *config.ArtifactoryDetails
	releaseBundlesParams distributionServicesUtils.ReleaseBundleParams
	spec                 *spec.SpecFiles
	dryRun               bool
}

func NewReleaseBundleUpdateCommand() *UpdateBundleCommand {
	return &UpdateBundleCommand{}
}

func (cb *UpdateBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *UpdateBundleCommand {
	cb.rtDetails = rtDetails
	return cb
}

func (cb *UpdateBundleCommand) SetReleaseBundleUpdateParams(params distributionServicesUtils.ReleaseBundleParams) *UpdateBundleCommand {
	cb.releaseBundlesParams = params
	return cb
}

func (cb *UpdateBundleCommand) SetSpec(spec *spec.SpecFiles) *UpdateBundleCommand {
	cb.spec = spec
	return cb
}

func (cb *UpdateBundleCommand) SetDryRun(dryRun bool) *UpdateBundleCommand {
	cb.dryRun = dryRun
	return cb
}

func (cb *UpdateBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(cb.rtDetails, cb.dryRun)
	if err != nil {
		return err
	}

	for _, spec := range cb.spec.Files {
		cb.releaseBundlesParams.SpecFiles = append(cb.releaseBundlesParams.SpecFiles, spec.ToArtifactoryCommonParams())
	}

	params := services.UpdateReleaseBundleParams{ReleaseBundleParams: cb.releaseBundlesParams}
	return servicesManager.UpdateReleaseBundle(params)
}

func (cb *UpdateBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return cb.rtDetails, nil
}

func (cb *UpdateBundleCommand) CommandName() string {
	return "rt_bundle_update"
}

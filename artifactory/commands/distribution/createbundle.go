package distribution

import (
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
	distributionServicesUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
)

type CreateBundleCommand struct {
	rtDetails            *config.ArtifactoryDetails
	releaseBundlesParams distributionServicesUtils.ReleaseBundleParams
	spec                 *spec.SpecFiles
	dryRun               bool
}

func NewReleaseBundleCreateCommand() *CreateBundleCommand {
	return &CreateBundleCommand{}
}

func (cb *CreateBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *CreateBundleCommand {
	cb.rtDetails = rtDetails
	return cb
}

func (cb *CreateBundleCommand) SetReleaseBundleCreateParams(params distributionServicesUtils.ReleaseBundleParams) *CreateBundleCommand {
	cb.releaseBundlesParams = params
	return cb
}

func (cb *CreateBundleCommand) SetSpec(spec *spec.SpecFiles) *CreateBundleCommand {
	cb.spec = spec
	return cb
}

func (cb *CreateBundleCommand) SetDryRun(dryRun bool) *CreateBundleCommand {
	cb.dryRun = dryRun
	return cb
}

func (cb *CreateBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(cb.rtDetails, cb.dryRun)
	if err != nil {
		return err
	}

	for _, spec := range cb.spec.Files {
		cb.releaseBundlesParams.SpecFiles = append(cb.releaseBundlesParams.SpecFiles, spec.ToArtifactoryCommonParams())
	}

	params := services.CreateReleaseBundleParams{ReleaseBundleParams: cb.releaseBundlesParams}
	return servicesManager.CreateReleaseBundle(params)
}

func (cb *CreateBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return cb.rtDetails, nil
}

func (cb *CreateBundleCommand) CommandName() string {
	return "rt_bundle_create"
}

package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
	distributionServicesUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
)

type CommandType string

const (
	Create CommandType = "Create"
	Update             = "Update"
)

type CreateBundleCommand struct {
	commandType          CommandType
	rtDetails            *config.ArtifactoryDetails
	releaseBundlesParams distributionServicesUtils.ReleaseBundleParams
	spec                 *spec.SpecFiles
	dryRun               bool
}

func NewReleaseBundleCreateUpdateCommand(commandType CommandType) *CreateBundleCommand {
	return &CreateBundleCommand{commandType: commandType}
}

func (cb *CreateBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *CreateBundleCommand {
	cb.rtDetails = rtDetails
	return cb
}

func (cb *CreateBundleCommand) SetReleaseBundleCreateUpdateParams(params distributionServicesUtils.ReleaseBundleParams) *CreateBundleCommand {
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

	if cb.commandType == Create {
		params := services.CreateReleaseBundleParams{
			ReleaseBundleParams: cb.releaseBundlesParams,
		}
		return servicesManager.CreateReleaseBundle(params)
	}
	params := services.UpdateReleaseBundleParams{
		ReleaseBundleParams: cb.releaseBundlesParams,
	}
	return servicesManager.UpdateReleaseBundle(params)
}

func (cb *CreateBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return cb.rtDetails, nil
}

func (cb *CreateBundleCommand) CommandName() string {
	return "rt_create_bundle"
}

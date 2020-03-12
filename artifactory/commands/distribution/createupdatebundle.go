package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
)

type CommandType string

const (
	Create CommandType = "Create"
	Update             = "Update"
)

type CreateBundleCommand struct {
	commandType         CommandType
	rtDetails           *config.ArtifactoryDetails
	createBundlesParams services.CreateUpdateReleaseBundleParams
	spec                *spec.SpecFiles
	dryRun              bool
}

func NewReleaseBundleCreateUpdateCommand(commandType CommandType) *CreateBundleCommand {
	return &CreateBundleCommand{commandType: commandType}
}

func (cb *CreateBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *CreateBundleCommand {
	cb.rtDetails = rtDetails
	return cb
}

func (cb *CreateBundleCommand) SetReleaseBundleCreateUpdateParams(params services.CreateUpdateReleaseBundleParams) *CreateBundleCommand {
	cb.createBundlesParams = params
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
		cb.createBundlesParams.SpecFiles = append(cb.createBundlesParams.SpecFiles, spec.ToArtifactoryCommonParams())
	}

	if cb.commandType == Create {
		return servicesManager.CreateReleaseBundle(cb.createBundlesParams)
	}
	return servicesManager.UpdateReleaseBundle(cb.createBundlesParams)
}

func (cb *CreateBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return cb.rtDetails, nil
}

func (cb *CreateBundleCommand) CommandName() string {
	return "rt_create_bundle"
}

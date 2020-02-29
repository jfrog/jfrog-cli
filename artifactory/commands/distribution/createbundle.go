package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
)

type CreateBundleCommand struct {
	rtDetails           *config.ArtifactoryDetails
	createBundlesParams services.CreateBundleParams
	spec                *spec.SpecFiles
	dryRun              bool
}

func NewCreateBundleCommand() *CreateBundleCommand {
	return &CreateBundleCommand{}
}

func (createBundle *CreateBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *CreateBundleCommand {
	createBundle.rtDetails = rtDetails
	return createBundle
}

func (createBundle *CreateBundleCommand) SetCreateBundleParams(params services.CreateBundleParams) *CreateBundleCommand {
	createBundle.createBundlesParams = params
	return createBundle
}

func (createBundle *CreateBundleCommand) SetSpec(spec *spec.SpecFiles) *CreateBundleCommand {
	createBundle.spec = spec
	return createBundle
}

func (createBundle *CreateBundleCommand) SetDryRun(dryRun bool) *CreateBundleCommand {
	createBundle.dryRun = dryRun
	return createBundle
}

func (createBundle *CreateBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(createBundle.rtDetails, createBundle.dryRun)
	if err != nil {
		return err
	}

	for _, spec := range createBundle.spec.Files {
		createBundle.createBundlesParams.SpecFiles = append(createBundle.createBundlesParams.SpecFiles, spec.ToArtifactoryCommonParams())
	}

	return servicesManager.CreateReleaseBundle(createBundle.createBundlesParams)
}

func (createBundle *CreateBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return createBundle.rtDetails, nil
}

func (createBundle *CreateBundleCommand) CommandName() string {
	return "rt_create_bundle"
}

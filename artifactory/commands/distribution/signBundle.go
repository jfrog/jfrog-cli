package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
)

type SignBundleCommand struct {
	rtDetails         *config.ArtifactoryDetails
	signBundlesParams services.SignBundleParams
}

func NewSignBundleCommand() *SignBundleCommand {
	return &SignBundleCommand{}
}

func (signBundle *SignBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *SignBundleCommand {
	signBundle.rtDetails = rtDetails
	return signBundle
}

func (signBundle *SignBundleCommand) SetSignBundleParams(params services.SignBundleParams) *SignBundleCommand {
	signBundle.signBundlesParams = params
	return signBundle
}

func (signBundle *SignBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(signBundle.rtDetails, false)
	if err != nil {
		return err
	}

	return servicesManager.SignReleaseBundle(signBundle.signBundlesParams)
}

func (signBundle *SignBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return signBundle.rtDetails, nil
}

func (signBundle *SignBundleCommand) CommandName() string {
	return "rt_sign_bundle"
}

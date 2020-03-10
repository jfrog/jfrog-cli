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

func (sb *SignBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *SignBundleCommand {
	sb.rtDetails = rtDetails
	return sb
}

func (sb *SignBundleCommand) SetSignBundleParams(params services.SignBundleParams) *SignBundleCommand {
	sb.signBundlesParams = params
	return sb
}

func (sb *SignBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(sb.rtDetails, false)
	if err != nil {
		return err
	}

	return servicesManager.SignReleaseBundle(sb.signBundlesParams)
}

func (sb *SignBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return sb.rtDetails, nil
}

func (sb *SignBundleCommand) CommandName() string {
	return "rt_sign_bundle"
}

package bundle

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func UpgradeBundle(artDetails *config.ArtifactoryDetails, bundleName string) (err error) {

	//Create Service Manager:
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return
	}

	versions, err := servicesManager.GetBundleVersions(bundleName)
	log.Output(versions)

	return err
}

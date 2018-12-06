package bundle

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mcuadros/go-version"
)

func UpgradeBundle(artDetails *config.ArtifactoryDetails, bundleConfigDetails *config.BundleDetails) error {

	//Create Service Manager:
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return err
	}

	constraints := version.NewConstrainGroupFromString(bundleConfigDetails.Version)
	cc := constraints.GetConstraints()

	versions, err := servicesManager.GetBundleVersions(bundleConfigDetails.Name)
	log.Output(versions)
	for _, versionToCheck := range versions {
		log.Output("Version", versionToCheck.Version, "is", constraints.Match(versionToCheck.Version), "with", bundleConfigDetails.Version)
	}

	return err
}

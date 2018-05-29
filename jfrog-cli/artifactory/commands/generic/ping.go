package generic

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
)

func Ping(artDetails *config.ArtifactoryDetails) ([]byte, error) {
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return nil, err
	}
	return servicesManager.Ping()
}

package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
)

func Ping(artDetails *config.ArtifactoryDetails) ([]byte, error) {
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return nil, err
	}
	return servicesManager.Ping()
}

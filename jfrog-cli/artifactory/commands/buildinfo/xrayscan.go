package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func XrayScan(buildName, buildNumber string, artDetails *config.ArtifactoryDetails) error {
	log.Info("Performing Xray build scan, this operation might take few minutes...")
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return err
	}

	params := new(services.XrayScanParamsImpl)
	params.BuildName = buildName
	params.BuildNumber = buildNumber
	result, err := servicesManager.XrayScanBuild(params)
	if err != nil {
		return err
	}

	log.Info("Xray scan completed.")
	log.Output(clientutils.IndentJson(result))
	return err
}

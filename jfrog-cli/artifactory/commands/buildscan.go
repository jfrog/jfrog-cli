package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

func BuildScan(buildName, buildNumber string, artDetails *config.ArtifactoryDetails) error {
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

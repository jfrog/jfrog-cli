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

	xrayScanParams := GetXrayScanParams(buildName, buildNumber)
	result, err := servicesManager.XrayScanBuild(xrayScanParams)
	if err != nil {
		return err
	}

	log.Info("Xray scan completed.")
	log.Output(clientutils.IndentJson(result))
	return err
}

func GetXrayScanParams(buildName, buildNumber string) (xrayScanParams services.XrayScanParams) {

	xrayScanParams = services.NewXrayScanParams()
	xrayScanParams.BuildName = buildName
	xrayScanParams.BuildNumber = buildNumber

	return
}

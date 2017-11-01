package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"fmt"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

func BuildScan(buildName, buildNumber string, artDetails *config.ArtifactoryDetails) (err error) {
	cliutils.CliLogger.Info("Performing Xray build scan, this operation might take few minutes...")
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

	cliutils.CliLogger.Info("Scan result:")
	fmt.Println(clientutils.IndentJson(result))
	return err
}

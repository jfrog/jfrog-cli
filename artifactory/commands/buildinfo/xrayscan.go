package buildinfo

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func XrayScan(buildName, buildNumber string, artDetails *config.ArtifactoryDetails, failBuild bool) (buildFailed bool, err error) {
	log.Info("Triggered Xray build scan... The scan may take a few minutes.")
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return false, err
	}

	xrayScanParams := getXrayScanParams(buildName, buildNumber)
	result, err := servicesManager.XrayScanBuild(xrayScanParams)
	if err != nil {
		return false, err
	}

	var scanResults scanResult
	err = json.Unmarshal(result, &scanResults)
	if errorutils.CheckError(err) != nil {
		return false, err
	}

	log.Info("Xray scan completed.")
	log.Output(clientutils.IndentJson(result))

	// Check if should fail build
	if failBuild && scanResults.Summary.FailBuild {
		return true, errorutils.CheckError(errors.New(scanResults.Summary.Message))
	}

	return false, err
}

// To unmarshal xray scan summary result
type scanResult struct {
	Summary scanSummary `json:"summary,omitempty"`
}

type scanSummary struct {
	TotalAlerts int    `json:"total_alerts,omitempty"`
	FailBuild   bool   `json:"fail_build,omitempty"`
	Message     string `json:"message,omitempty"`
	Url         string `json:"more_details_url,omitempty"`
}

func getXrayScanParams(buildName, buildNumber string) services.XrayScanParams {
	xrayScanParams := services.NewXrayScanParams()
	xrayScanParams.BuildName = buildName
	xrayScanParams.BuildNumber = buildNumber

	return xrayScanParams
}

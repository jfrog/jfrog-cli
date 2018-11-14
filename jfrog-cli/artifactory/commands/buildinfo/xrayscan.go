package buildinfo

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func XrayScan(buildName, buildNumber string, artDetails *config.ArtifactoryDetails, failBuild bool) (buildFailed bool, err error) {
	log.Info("Performing Xray build scan, this operation might take few minutes...")
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return false, err
	}

	params := new(services.XrayScanParamsImpl)
	params.BuildName = buildName
	params.BuildNumber = buildNumber
	result, err := servicesManager.XrayScanBuild(params)
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

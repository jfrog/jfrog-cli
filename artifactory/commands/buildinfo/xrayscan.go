package buildinfo

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type BuildScanCommand struct {
	buildConfiguration *utils.BuildConfiguration
	failBuild          bool
	rtDetails          *config.ArtifactoryDetails
}

func NewBuildScanCommand() *BuildScanCommand {
	return &BuildScanCommand{}
}

func (bsc *BuildScanCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *BuildScanCommand {
	bsc.rtDetails = rtDetails
	return bsc
}

func (bsc *BuildScanCommand) SetFailBuild(failBuild bool) *BuildScanCommand {
	bsc.failBuild = failBuild
	return bsc
}

func (bsc *BuildScanCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *BuildScanCommand {
	bsc.buildConfiguration = buildConfiguration
	return bsc
}

func (bsc *BuildScanCommand) CommandName() string {
	return "rt_build_scan"
}

func (bsc *BuildScanCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return bsc.rtDetails, nil
}

func (bsc *BuildScanCommand) Run() error {
	log.Info("Triggered Xray build scan... The scan may take a few minutes.")
	servicesManager, err := utils.CreateServiceManager(bsc.rtDetails, false)
	if err != nil {
		return err
	}

	xrayScanParams := getXrayScanParams(bsc.buildConfiguration.BuildName, bsc.buildConfiguration.BuildNumber)
	result, err := servicesManager.XrayScanBuild(xrayScanParams)
	if err != nil {
		return err
	}

	var scanResults scanResult
	err = json.Unmarshal(result, &scanResults)
	if errorutils.CheckError(err) != nil {
		return err
	}

	log.Info("Xray scan completed.")
	log.Output(clientutils.IndentJson(result))

	// Check if should fail build
	if bsc.failBuild && scanResults.Summary.FailBuild {
		// We're specifically returning the 'buildScanError' and not a regular error
		// to indicate that Xray indeed scanned the build, and the failure is not due to
		// networking connectivity or other issues.
		return errorutils.CheckError(utils.GetBuildScanError())
	}

	return err
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

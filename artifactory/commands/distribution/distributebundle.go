package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
)

type DistributeBundleCommand struct {
	rtDetails               *config.ArtifactoryDetails
	distributeBundlesParams services.DistributionParams
	distributionRules       *spec.DistributionRules
	dryRun                  bool
}

func NewReleaseBundleDistributeCommand() *DistributeBundleCommand {
	return &DistributeBundleCommand{}
}

func (db *DistributeBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DistributeBundleCommand {
	db.rtDetails = rtDetails
	return db
}

func (db *DistributeBundleCommand) SetDistributeBundleParams(params services.DistributionParams) *DistributeBundleCommand {
	db.distributeBundlesParams = params
	return db
}

func (db *DistributeBundleCommand) SetDistributionRules(distributionRules *spec.DistributionRules) *DistributeBundleCommand {
	db.distributionRules = distributionRules
	return db
}

func (db *DistributeBundleCommand) SetDryRun(dryRun bool) *DistributeBundleCommand {
	db.dryRun = dryRun
	return db
}

func (db *DistributeBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(db.rtDetails, db.dryRun)
	if err != nil {
		return err
	}

	for _, rule := range db.distributionRules.DistributionRules {
		db.distributeBundlesParams.DistributionRules = append(db.distributeBundlesParams.DistributionRules, rule.ToDistributionCommonParams())
	}

	return servicesManager.DistributeReleaseBundle(db.distributeBundlesParams)
}

func (db *DistributeBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return db.rtDetails, nil
}

func (db *DistributeBundleCommand) CommandName() string {
	return "rt_distribute_bundle"
}

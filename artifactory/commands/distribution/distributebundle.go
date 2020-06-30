package distribution

import (
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
)

type DistributeReleaseBundleCommand struct {
	rtDetails               *config.ArtifactoryDetails
	distributeBundlesParams services.DistributionParams
	distributionRules       *spec.DistributionRules
	sync                    bool
	maxWaitMinutes          int
	dryRun                  bool
}

func NewReleaseBundleDistributeCommand() *DistributeReleaseBundleCommand {
	return &DistributeReleaseBundleCommand{}
}

func (db *DistributeReleaseBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DistributeReleaseBundleCommand {
	db.rtDetails = rtDetails
	return db
}

func (db *DistributeReleaseBundleCommand) SetDistributeBundleParams(params services.DistributionParams) *DistributeReleaseBundleCommand {
	db.distributeBundlesParams = params
	return db
}

func (db *DistributeReleaseBundleCommand) SetDistributionRules(distributionRules *spec.DistributionRules) *DistributeReleaseBundleCommand {
	db.distributionRules = distributionRules
	return db
}

func (db *DistributeReleaseBundleCommand) SetSync(sync bool) *DistributeReleaseBundleCommand {
	db.sync = sync
	return db
}

func (db *DistributeReleaseBundleCommand) SetMaxWaitMinutes(maxWaitMinutes int) *DistributeReleaseBundleCommand {
	db.maxWaitMinutes = maxWaitMinutes
	return db
}

func (db *DistributeReleaseBundleCommand) SetDryRun(dryRun bool) *DistributeReleaseBundleCommand {
	db.dryRun = dryRun
	return db
}

func (db *DistributeReleaseBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(db.rtDetails, db.dryRun)
	if err != nil {
		return err
	}

	for _, rule := range db.distributionRules.DistributionRules {
		db.distributeBundlesParams.DistributionRules = append(db.distributeBundlesParams.DistributionRules, rule.ToDistributionCommonParams())
	}

	if db.sync {
		return servicesManager.DistributeReleaseBundleSync(db.distributeBundlesParams, db.maxWaitMinutes)
	}
	return servicesManager.DistributeReleaseBundle(db.distributeBundlesParams)
}

func (db *DistributeReleaseBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return db.rtDetails, nil
}

func (db *DistributeReleaseBundleCommand) CommandName() string {
	return "rt_distribute_bundle"
}

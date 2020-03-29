package distribution

import (
	"encoding/json"
	"fmt"

	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

var (
	defaultDeleteRules = spec.DistributionRules{
		DistributionRules: []spec.DistributionRule{{
			SiteName: "*",
		}},
	}
)

type DeleteReleaseBundleCommand struct {
	rtDetails           *config.ArtifactoryDetails
	deleteBundlesParams services.DeleteDistributionParams
	distributionRules   *spec.DistributionRules
	dryRun              bool
	quiet               bool
}

func NewReleaseBundleDeleteParams() *DeleteReleaseBundleCommand {
	return &DeleteReleaseBundleCommand{}
}

func (db *DeleteReleaseBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DeleteReleaseBundleCommand {
	db.rtDetails = rtDetails
	return db
}

func (db *DeleteReleaseBundleCommand) SetDistributeBundleParams(params services.DeleteDistributionParams) *DeleteReleaseBundleCommand {
	db.deleteBundlesParams = params
	return db
}

func (db *DeleteReleaseBundleCommand) SetDistributionRules(distributionRules *spec.DistributionRules) *DeleteReleaseBundleCommand {
	db.distributionRules = distributionRules
	return db
}

func (db *DeleteReleaseBundleCommand) SetDryRun(dryRun bool) *DeleteReleaseBundleCommand {
	db.dryRun = dryRun
	return db
}

func (db *DeleteReleaseBundleCommand) SetQuiet(quiet bool) *DeleteReleaseBundleCommand {
	db.quiet = quiet
	return db
}

func (db *DeleteReleaseBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(db.rtDetails, db.dryRun)
	if err != nil {
		return err
	}

	for _, spec := range db.distributionRules.DistributionRules {
		db.deleteBundlesParams.DistributionRules = append(db.deleteBundlesParams.DistributionRules, spec.ToDistributionCommonParams())
	}

	distributionRulesEmpty := db.distributionRulesEmpty()
	if !db.quiet {
		confirm, err := db.confirmDelete(distributionRulesEmpty)
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	if distributionRulesEmpty && db.deleteBundlesParams.DeleteFromDistribution {
		return servicesManager.DeleteLocalReleaseBundle(db.deleteBundlesParams)
	}
	return servicesManager.DeleteReleaseBundle(db.deleteBundlesParams)
}

func (db *DeleteReleaseBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return db.rtDetails, nil
}

func (db *DeleteReleaseBundleCommand) CommandName() string {
	return "rt_bundle_delete"
}

// Return true iff there are no distribution rules
func (db *DeleteReleaseBundleCommand) distributionRulesEmpty() bool {
	return db.distributionRules == nil ||
		len(db.distributionRules.DistributionRules) == 0 ||
		len(db.distributionRules.DistributionRules) == 1 && db.distributionRules.DistributionRules[0].IsEmpty()
}

func (db *DeleteReleaseBundleCommand) confirmDelete(distributionRulesEmpty bool) (bool, error) {
	message := fmt.Sprintf("Are you sure you want to delete the release bundle \"%s\"/\"%s\" ", db.deleteBundlesParams.Name, db.deleteBundlesParams.Version)
	if distributionRulesEmpty && db.deleteBundlesParams.DeleteFromDistribution {
		return cliutils.InteractiveConfirm(message + "locally from distribution?\n" +
			"You can avoid this confirmation message by adding --quiet to the command."), nil
	}

	var distributionRulesBodies []services.DistributionRulesBody
	for _, rule := range db.deleteBundlesParams.DistributionRules {
		distributionRulesBodies = append(distributionRulesBodies, services.DistributionRulesBody{
			SiteName:     rule.GetSiteName(),
			CityName:     rule.GetCityName(),
			CountryCodes: rule.GetCountryCodes(),
		})
	}
	bytes, err := json.Marshal(distributionRulesBodies)
	if err != nil {
		return false, errorutils.CheckError(err)
	}

	fmt.Println(clientutils.IndentJson(bytes))
	if db.deleteBundlesParams.DeleteFromDistribution {
		fmt.Println("This command will also delete the release bundle locally from distribution.")
	}
	return cliutils.InteractiveConfirm(message + "with the above distribution rules?\n" +
		"You can avoid this confirmation message by adding --quiet to the command."), nil
}

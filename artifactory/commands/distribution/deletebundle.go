package distribution

import (
	"encoding/json"
	"fmt"

	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/distribution/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type DeleteBundleCommand struct {
	rtDetails           *config.ArtifactoryDetails
	deleteBundlesParams services.DeleteDistributionParams
	distributionRules   *spec.DistributionRules
	dryRun              bool
	quiet               bool
}

func NewDeleteBundleCommand() *DeleteBundleCommand {
	return &DeleteBundleCommand{}
}

func (db *DeleteBundleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DeleteBundleCommand {
	db.rtDetails = rtDetails
	return db
}

func (db *DeleteBundleCommand) SetDistributeBundleParams(params services.DeleteDistributionParams) *DeleteBundleCommand {
	db.deleteBundlesParams = params
	return db
}

func (db *DeleteBundleCommand) SetDistributionRules(distributionRules *spec.DistributionRules) *DeleteBundleCommand {
	db.distributionRules = distributionRules
	return db
}

func (db *DeleteBundleCommand) SetDryRun(dryRun bool) *DeleteBundleCommand {
	db.dryRun = dryRun
	return db
}

func (db *DeleteBundleCommand) SetQuiet(quiet bool) *DeleteBundleCommand {
	db.quiet = quiet
	return db
}

func (db *DeleteBundleCommand) Run() error {
	servicesManager, err := utils.CreateDistributionServiceManager(db.rtDetails, db.dryRun)
	if err != nil {
		return err
	}

	for _, spec := range db.distributionRules.DistributionRules {
		db.deleteBundlesParams.DistributionRules = append(db.deleteBundlesParams.DistributionRules, spec.ToDistributionCommonParams())
	}

	if !db.quiet {
		confirm, err := db.confirmDelete()
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	return servicesManager.DeleteReleaseBundle(db.deleteBundlesParams)
}

func (db *DeleteBundleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return db.rtDetails, nil
}

func (db *DeleteBundleCommand) CommandName() string {
	return "rt_delete_bundle"
}

func (db *DeleteBundleCommand) confirmDelete() (bool, error) {
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
	return cliutils.InteractiveConfirm("Are you sure you want to delete the release bundle \"" +
		db.deleteBundlesParams.Name + "/" + db.deleteBundlesParams.Version + "\" with the above distribution rules?\n" +
		"You can avoid this confirmation message by adding --quiet to the command."), nil
}

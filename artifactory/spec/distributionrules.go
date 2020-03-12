package spec

import (
	"encoding/json"

	"github.com/jfrog/jfrog-client-go/distribution/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type DistributionRules struct {
	DistributionRules []DistributionRule `json:"distribution_rules,omitempty"`
}

type DistributionRule struct {
	SiteName     string   `json:"site_name,omitempty"`
	CityName     string   `json:"city_name,omitempty"`
	CountryCodes []string `json:"country_codes,omitempty"`
}

func (distributionRules *DistributionRules) Get(index int) *DistributionRule {
	if index < len(distributionRules.DistributionRules) {
		return &distributionRules.DistributionRules[index]
	}
	return new(DistributionRule)
}

func (distributionRule *DistributionRule) ToDistributionCommonParams() *utils.DistributionCommonParams {
	return &utils.DistributionCommonParams{
		SiteName:     distributionRule.SiteName,
		CityName:     distributionRule.CityName,
		CountryCodes: distributionRule.CountryCodes,
	}
}

func (distributionRule *DistributionRule) IsEmpty() bool {
	return distributionRule.SiteName == "" && distributionRule.CityName == "" && len(distributionRule.CountryCodes) == 0
}

func CreateDistributionRulesFromFile(distributionSpecPath string) (*DistributionRules, error) {
	content, err := fileutils.ReadFile(distributionSpecPath)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	distributionRules := new(DistributionRules)
	err = json.Unmarshal(content, distributionRules)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return distributionRules, nil
}

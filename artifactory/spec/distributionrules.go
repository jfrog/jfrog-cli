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

func (spec *DistributionRules) Get(index int) *DistributionRule {
	if index < len(spec.DistributionRules) {
		return &spec.DistributionRules[index]
	}
	return new(DistributionRule)
}

func (ds *DistributionRule) ToDistributionCommonParams() *utils.DistributionCommonParams {
	return &utils.DistributionCommonParams{
		SiteName:     ds.SiteName,
		CityName:     ds.CityName,
		CountryCodes: ds.CountryCodes,
	}
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

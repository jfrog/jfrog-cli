package spec

import (
	"encoding/json"

	"github.com/jfrog/jfrog-client-go/distribution/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type DistributionRules struct {
	DistributionRules []DistributionRule
}

func (spec *DistributionRules) Get(index int) *DistributionRule {
	if index < len(spec.DistributionRules) {
		return &spec.DistributionRules[index]
	}
	return new(DistributionRule)
}

type DistributionRule struct {
	SiteName     string
	CityName     string
	CountryCodes []string
}

func (ds *DistributionRule) ToDistributionCommonParams() *utils.DistributionCommonParams {
	return &utils.DistributionCommonParams{
		SiteName:     ds.SiteName,
		CityName:     ds.CityName,
		CountryCodes: ds.CountryCodes,
	}
}

func CreateDistributionRulesFromFile(distributionSpecPath string) (rule *DistributionRules, err error) {
	rule = new(DistributionRules)
	content, err := fileutils.ReadFile(distributionSpecPath)
	if errorutils.CheckError(err) != nil {
		return
	}

	err = json.Unmarshal(content, rule)
	if errorutils.CheckError(err) != nil {
		return
	}
	return
}

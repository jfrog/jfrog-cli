package spec

import (
	"encoding/json"

	"github.com/jfrog/jfrog-client-go/distribution/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type DistributionSpecs struct {
	Specs []DistributionSpec
}

func (spec *DistributionSpecs) Get(index int) *DistributionSpec {
	if index < len(spec.Specs) {
		return &spec.Specs[index]
	}
	return new(DistributionSpec)
}

type DistributionSpec struct {
	SiteName     string
	CityName     string
	CountryCodes []string
}

func (ds *DistributionSpec) ToDistributionCommonParams() *utils.DistributionCommonParams {
	return &utils.DistributionCommonParams{
		SiteName:     ds.SiteName,
		CityName:     ds.CityName,
		CountryCodes: ds.CountryCodes,
	}
}

func CreateDistributionSpecFromFile(distributionSpecPath string) (spec *DistributionSpecs, err error) {
	spec = new(DistributionSpecs)
	content, err := fileutils.ReadFile(distributionSpecPath)
	if errorutils.CheckError(err) != nil {
		return
	}

	err = json.Unmarshal(content, spec)
	if errorutils.CheckError(err) != nil {
		return
	}
	return
}

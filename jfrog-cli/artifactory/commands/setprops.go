package commands

import (
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/spec"
)

func SetProps(spec *spec.SpecFiles, props string, artDetails *config.ArtifactoryDetails) error {
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return err
	}
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(spec.Files); i++ {
		params, err := spec.Get(i).ToArtifatorySetPropsParams()
		if err != nil {
			return err
		}
		currentResultItems, err := servicesManager.Search(&clientutils.SearchParamsImpl{ArtifactoryCommonParams: params})
		if err != nil {
			return err
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	return servicesManager.SetProps(&services.SetPropsParamsImpl{Items:resultItems, Props:props})
}
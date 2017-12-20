package commands

import (
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/spec"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

func SetProps(spec *spec.SpecFiles, props string, artDetails *config.ArtifactoryDetails) (int, int, error) {
	servicesManager, err := utils.CreateServiceManager(artDetails, false)
	if err != nil {
		return 0, 0, err
	}
	var resultItems []clientutils.ResultItem
	for i := 0; i < len(spec.Files); i++ {
		params, err := spec.Get(i).ToArtifatorySetPropsParams()
		if err != nil {
			log.Error(err)
			continue
		}
		currentResultItems, err := servicesManager.Search(&clientutils.SearchParamsImpl{ArtifactoryCommonParams: params})
		if err != nil {
			log.Error(err)
			continue
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	success, err := servicesManager.SetProps(&services.SetPropsParamsImpl{Items: resultItems, Props: props})
	return success, len(resultItems) - success, err
}
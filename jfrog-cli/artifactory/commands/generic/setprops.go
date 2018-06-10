package generic

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory"
)

func SetProps(spec *spec.SpecFiles, props string, threads int, artDetails *config.ArtifactoryDetails) (successCount, failCount int, err error) {
	servicesManager, err := createSetPropsServiceManager(threads, artDetails)
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

func createSetPropsServiceManager(threads int, artDetails *config.ArtifactoryDetails) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := utils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := artifactory.NewConfigBuilder().
			SetArtDetails(artAuth).
			SetCertificatesPath(certPath).
			SetLogger(log.Logger).
			SetThreads(threads).
			Build()

	return artifactory.New(serviceConfig)
}
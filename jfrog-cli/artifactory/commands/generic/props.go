package generic

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func SetProps(spec *spec.SpecFiles, props string, threads int, artDetails *config.ArtifactoryDetails) (successCount, failCount int, err error) {
	servicesManager, err := createPropsServiceManager(threads, artDetails)
	if err != nil {
		return 0, 0, err
	}

	resultItems := searchItems(spec, servicesManager)

	propsParams := GetPropsParams(resultItems, props)

	success, err := servicesManager.SetProps(propsParams)
	return success, len(resultItems) - success, err
}

func DeleteProps(spec *spec.SpecFiles, props string, threads int, artDetails *config.ArtifactoryDetails) (successCount, failCount int, err error) {
	servicesManager, err := createPropsServiceManager(threads, artDetails)
	if err != nil {
		return 0, 0, err
	}

	resultItems := searchItems(spec, servicesManager)

	propsParams := GetPropsParams(resultItems, props)

	success, err := servicesManager.DeleteProps(propsParams)
	return success, len(resultItems) - success, err
}

func createPropsServiceManager(threads int, artDetails *config.ArtifactoryDetails) (*artifactory.ArtifactoryServicesManager, error) {
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
		SetInsecureTls(artDetails.InsecureTls).
		SetLogger(log.Logger).
		SetThreads(threads).
		Build()

	return artifactory.New(&artAuth, serviceConfig)
}

func searchItems(spec *spec.SpecFiles, servicesManager *artifactory.ArtifactoryServicesManager) (resultItems []clientutils.ResultItem) {
	for i := 0; i < len(spec.Files); i++ {
		searchParams, err := getSearchParamsForProps(spec.Get(i))
		if err != nil {
			log.Error(err)
			continue
		}

		currentResultItems, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			log.Error(err)
			continue
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	return
}

func GetPropsParams(resultItems []clientutils.ResultItem, properties string) (propsParams services.PropsParams) {
	propsParams = services.NewPropsParams()
	propsParams.Items = resultItems
	propsParams.Props = properties
	return
}

func getSearchParamsForProps(f *spec.File) (searchParams services.SearchParams, err error) {
	searchParams = services.NewSearchParams()
	searchParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	searchParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	searchParams.IncludeDirs, err = f.IsIncludeDirs(false)
	if err != nil {
		return
	}
	return
}

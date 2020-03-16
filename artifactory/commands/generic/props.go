package generic

import (
	"errors"

	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type PropsCommand struct {
	props   string
	threads int
	GenericCommand
}

func NewPropsCommand() *PropsCommand {
	return &PropsCommand{GenericCommand: *NewGenericCommand()}
}

func (pc *PropsCommand) Threads() int {
	return pc.threads
}

func (pc *PropsCommand) SetThreads(threads int) *PropsCommand {
	pc.threads = threads
	return pc
}

func (pc *PropsCommand) Props() string {
	return pc.props
}

func (pc *PropsCommand) SetProps(props string) *PropsCommand {
	pc.props = props
	return pc
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
	serviceConfig, err := clientConfig.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetCertificatesPath(certPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetThreads(threads).
		Build()

	return artifactory.New(&artAuth, serviceConfig)
}

func searchItems(spec *spec.SpecFiles, servicesManager *artifactory.ArtifactoryServicesManager) (resultItems []clientutils.ResultItem, err error) {
	var errorOccurred = false
	for i := 0; i < len(spec.Files); i++ {
		searchParams, err := getSearchParamsForProps(spec.Get(i))
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}

		currentResultItems, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	if errorOccurred {
		err = errors.New("Operation finished with errors, please review the logs.")
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

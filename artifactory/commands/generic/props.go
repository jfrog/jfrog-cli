package generic

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	SetProperties    CommandName = "rt_set_properties"
	DeleteProperties CommandName = "rt_delete_properties"
)

type CommandName string

type PropsCommand struct {
	props       string
	commandName CommandName
	threads     int
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

func (pc *PropsCommand) SetCommandName(commandName CommandName) *PropsCommand {
	pc.commandName = commandName
	return pc
}

func (pc *PropsCommand) Run() error {
	servicesManager, err := createPropsServiceManager(pc.threads, pc.RtDetails())
	if err != nil {
		return err
	}

	resultItems := searchItems(pc.Spec(), servicesManager)

	propsParams := GetPropsParams(resultItems, pc.props)
	var success int
	switch pc.commandName {
	case SetProperties:
		{
			success, err = servicesManager.SetProps(propsParams)
		}
	case DeleteProperties:
		{
			success, err = servicesManager.DeleteProps(propsParams)
		}
	default:
		return fmt.Errorf("Received unknown command name: %s", pc.commandName)
	}
	result := pc.Result()
	result.SetSuccessCount(success)
	result.SetFailCount(len(resultItems) - success)
	return err
}

func (pc *PropsCommand) CommandName() string {
	return string(pc.commandName)
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

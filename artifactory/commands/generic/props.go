package generic

import (
	"errors"

	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
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

func createPropsServiceManager(threads int, artDetails *config.ArtifactoryDetails) (artifactory.ArtifactoryServicesManager, error) {
	certsPath, err := cliutils.GetJfrogCertsDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := clientConfig.NewConfigBuilder().
		SetServiceDetails(artAuth).
		SetCertificatesPath(certsPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetThreads(threads).
		Build()

	return artifactory.New(&artAuth, serviceConfig)
}

func searchItems(spec *spec.SpecFiles, servicesManager artifactory.ArtifactoryServicesManager) (resultReader *content.ContentReader, err error) {
	var errorOccurred = false
	temp := []*content.ContentReader{}
	writer, err := content.NewContentWriter("results", true, false)
	if err != nil {
		return
	}
	defer writer.Close()
	for i := 0; i < len(spec.Files); i++ {
		searchParams, err := getSearchParamsForProps(spec.Get(i))
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		reader, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		temp = append(temp, reader)
		if i == 0 {
			defer func() {
				for _, reader := range temp {
					reader.Close()
				}
			}()
		}
	}
	resultReader, err = content.MergeReaders(temp, content.DefaultKey)
	if err != nil {
		return
	}
	if errorOccurred {
		err = errorutils.CheckError(errors.New("Operation finished with errors, please review the logs."))
	}
	return
}

func GetPropsParams(reader *content.ContentReader, properties string) (propsParams services.PropsParams) {
	propsParams = services.NewPropsParams()
	propsParams.Reader = reader
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

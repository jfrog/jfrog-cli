package generic

import (
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	clientartutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type SearchCommand struct {
	GenericCommand
}

func NewSearchCommand() *SearchCommand {
	return &SearchCommand{GenericCommand: *NewGenericCommand()}
}

func (sc *SearchCommand) CommandName() string {
	return "rt_search"
}

func (sc *SearchCommand) Run() error {
	reader, err := sc.Search()
	sc.Result().SetReader(reader)
	return err
}

func (sc *SearchCommand) Search() (*content.ContentReader, error) {
	// Service Manager
	rtDetails, err := sc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	servicesManager, err := utils.CreateServiceManager(rtDetails, false)
	if err != nil {
		return nil, err
	}

	// Search Loop
	log.Info("Searching artifacts...")
	var searchResults []*content.ContentReader
	for i := 0; i < len(sc.Spec().Files); i++ {
		searchParams, err := utils.GetSearchParams(sc.Spec().Get(i))
		if err != nil {
			log.Error(err)
			return nil, err
		}

		reader, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		searchResults = append(searchResults, reader)
		if i == 0 {
			defer func() {
				for _, reader := range searchResults {
					reader.Close()
				}
			}()
		}
	}
	reader, err := utils.AqlResultToSearchResult(searchResults)
	if err != nil {
		return nil, err
	}
	length, err := reader.Length()
	clientartutils.LogSearchResults(length)
	return reader, err
}

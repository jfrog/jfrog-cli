package inttestutils

import (
	"testing"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
)

// Verify the input slice exist in Artifactory
// expected - The slice to check
// specFile - File spec for the search command
// serverDetails - Target Artifactory server details
// t - Tests object
func VerifyExistInArtifactory(expected []string, specFile string, serverDetails *config.ServerDetails, t *testing.T) {
	results, _ := SearchInArtifactory(specFile, serverDetails, t)
	tests.CompareExpectedVsActual(expected, results, t)
}

func SearchInArtifactory(specFile string, serverDetails *config.ServerDetails, t *testing.T) ([]utils.SearchResult, error) {
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	assert.NoError(t, err)
	for searchResult := new(utils.SearchResult); readerNoDate.NextRecord(searchResult) == nil; searchResult = new(utils.SearchResult) {
		resultItems = append(resultItems, *searchResult)
	}
	assert.NoError(t, reader.Close(), "Couldn't close reader")
	assert.NoError(t, reader.GetError(), "Couldn't get reader error")
	return resultItems, err
}

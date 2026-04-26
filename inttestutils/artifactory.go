package inttestutils

import (
	"testing"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
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
	// When Search() fails (e.g. a transient network error on the AQL POST),
	// reader is nil. Returning early prevents a nil-pointer panic from
	// tearing down the whole test binary and cascading into every
	// subsequent test in the shard.
	if err != nil || reader == nil {
		assert.NoError(t, err)
		return nil, err
	}
	defer func() {
		assert.NoError(t, reader.Close(), "Couldn't close reader")
		assert.NoError(t, reader.GetError(), "Couldn't get reader error")
	}()
	readerNoDate, err := utils.SearchResultNoDate(reader)
	if err != nil || readerNoDate == nil {
		assert.NoError(t, err)
		return nil, err
	}
	var resultItems []utils.SearchResult
	for searchResult := new(utils.SearchResult); readerNoDate.NextRecord(searchResult) == nil; searchResult = new(utils.SearchResult) {
		resultItems = append(resultItems, *searchResult)
	}
	return resultItems, nil
}

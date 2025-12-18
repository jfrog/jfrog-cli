package inttestutils

import (
	"testing"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verify the input slice exist in Artifactory
// expected - The slice to check
// specFile - File spec for the search command
// serverDetails - Target Artifactory server details
// t - Tests object
func VerifyExistInArtifactory(expected []string, specFile string, serverDetails *config.ServerDetails, t *testing.T) {
	results, err := SearchInArtifactory(specFile, serverDetails, t)
	require.NoError(t, err)
	tests.CompareExpectedVsActual(expected, results, t)
}

func SearchInArtifactory(specFile string, serverDetails *config.ServerDetails, t *testing.T) ([]utils.SearchResult, error) {
	searchSpec, err := spec.CreateSpecFromFile(specFile, nil)
	if err != nil {
		return nil, err
	}
	if searchSpec == nil {
		return nil, assert.AnError
	}
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	if err != nil {
		return nil, err
	}
	var resultItems []utils.SearchResult
	readerNoDate, err := utils.SearchResultNoDate(reader)
	if err != nil {
		_ = reader.Close()
		return nil, err
	}
	for searchResult := new(utils.SearchResult); readerNoDate.NextRecord(searchResult) == nil; searchResult = new(utils.SearchResult) {
		resultItems = append(resultItems, *searchResult)
	}
	if cerr := reader.Close(); cerr != nil {
		return resultItems, cerr
	}
	if rerr := reader.GetError(); rerr != nil {
		return resultItems, rerr
	}
	return resultItems, nil
}

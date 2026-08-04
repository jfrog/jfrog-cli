package inttestutils

import (
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
)

// VerifyExistInArtifactory verifies that the expected artifacts exist in Artifactory.
// It retries up to 5 times with 3-second intervals to handle Artifactory's async indexing delay.
func VerifyExistInArtifactory(expected []string, specFile string, serverDetails *config.ServerDetails, t *testing.T) {
	const maxRetries = 5
	const retryInterval = 3 * time.Second
	var results []utils.SearchResult
	for i := 0; i < maxRetries; i++ {
		var err error
		results, err = SearchInArtifactory(specFile, serverDetails, t)
		if err != nil {
			return
		}
		if len(results) >= len(expected) {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}
	tests.CompareExpectedVsActual(expected, results, t)
}

func SearchInArtifactory(specFile string, serverDetails *config.ServerDetails, t *testing.T) ([]utils.SearchResult, error) {
	searchSpec, _ := spec.CreateSpecFromFile(specFile, nil)
	return searchBySpec(searchSpec, serverDetails, t)
}

// SearchPathsByPattern returns the repo-relative paths of all items matching the
// given pattern (e.g. "my-repo/*"), searched recursively.
func SearchPathsByPattern(pattern string, serverDetails *config.ServerDetails, t *testing.T) []string {
	results, _ := searchBySpec(spec.NewBuilder().Pattern(pattern).Recursive(true).BuildSpec(), serverDetails, t)
	paths := make([]string, 0, len(results))
	for _, r := range results {
		paths = append(paths, r.Path)
	}
	return paths
}

func searchBySpec(searchSpec *spec.SpecFiles, serverDetails *config.ServerDetails, t *testing.T) ([]utils.SearchResult, error) {
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

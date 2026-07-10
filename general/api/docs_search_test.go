package api

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"

	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func stubOps(t *testing.T) []apispec.Operation {
	t.Helper()
	ops, err := apispec.Operations()
	require.NoError(t, err)
	require.NotEmpty(t, ops)
	return ops
}

func hasPath(matches []match, path string) bool {
	return slices.ContainsFunc(matches, func(m match) bool { return m.Path == path })
}

func TestFilterAndScore_ContainsMatchOnKnownStubOps(t *testing.T) {
	matches := filterAndScore(stubOps(t), "user", "", "")
	require.NotEmpty(t, matches, "query 'user' should match the stub's user operations")
	assert.True(t, hasPath(matches, "/access/api/v2/users"), "expected /access/api/v2/users (getUserList/createUser) in results")

	for i := 1; i < len(matches); i++ {
		assert.GreaterOrEqual(t, matches[i-1].Score, matches[i].Score, "matches must be sorted by score desc")
	}
}

func TestFilterAndScore_TagFilter(t *testing.T) {
	matches := filterAndScore(stubOps(t), "", "Users", "")
	require.NotEmpty(t, matches)
	for _, m := range matches {
		assert.True(t, slices.Contains(m.Tags, "Users"), "match %s %s should carry the Users tag", m.Method, m.Path)
	}
}

func TestFilterAndScore_MethodFilter(t *testing.T) {
	matches := filterAndScore(stubOps(t), "", "", "DELETE")
	require.NotEmpty(t, matches)
	for _, m := range matches {
		assert.Equal(t, "DELETE", m.Method)
	}
	assert.True(t, hasPath(matches, "/worker/api/v1/workers/{workerKey}"), "expected deleteWorker in DELETE-filtered results")
}

func TestFilterAndScore_MethodFilterCaseInsensitive(t *testing.T) {
	matches := filterAndScore(stubOps(t), "", "", "delete")
	require.NotEmpty(t, matches)
	for _, m := range matches {
		assert.Equal(t, "DELETE", m.Method)
	}
}

func TestFilterAndScore_CombinedFiltersNarrowResults(t *testing.T) {
	all := filterAndScore(stubOps(t), "", "", "")
	narrowed := filterAndScore(stubOps(t), "", "Users", "GET")
	assert.Less(t, len(narrowed), len(all))
	for _, m := range narrowed {
		assert.Equal(t, "GET", m.Method)
	}
}

func TestFilterAndScore_NoMatchIsEmpty(t *testing.T) {
	matches := filterAndScore(stubOps(t), "zzzznotreal", "", "")
	assert.Empty(t, matches, "a nonsense query should not fuzzy-match any stub operation")
}

func TestFilterAndScore_EmptyQueryMatchesEverything(t *testing.T) {
	ops := stubOps(t)
	matches := filterAndScore(ops, "", "", "")
	assert.Len(t, matches, len(ops))
}

func TestScoreOperation_MultiFieldContainsBeatsSingleField(t *testing.T) {
	// "user" contains-matches all four fields of multi; "widget" contains-matches
	// only the Summary of single -- weights must sum, not just take the best field.
	multi := apispec.Operation{OperationId: "getUserList", Path: "/access/api/v2/users", Summary: "Get User List", Tags: []string{"Users"}}
	single := apispec.Operation{OperationId: "createThing", Path: "/api/v2/things", Summary: "Create a Widget", Tags: []string{"Things"}}

	multiScore := scoreOperation(multi, "user")
	singleScore := scoreOperation(single, "widget")
	assert.Equal(t, 100, multiScore)
	assert.Equal(t, weightSummary, singleScore)
	assert.Greater(t, multiScore, singleScore, "matching operationId+path+summary+tag should outscore a single-field match")
}

func TestScoreOperation_ContainsAlwaysBeatsFuzzy(t *testing.T) {
	weakContains := apispec.Operation{OperationId: "x", Path: "/a", Summary: "s", Tags: []string{"usery"}}
	// "usrr" doesn't contain "user" as a substring (order differs), but is a
	// one-substitution Levenshtein neighbor of it -- a pure fuzzy match.
	closeFuzzy := apispec.Operation{OperationId: "usrr", Path: "/b", Summary: "s", Tags: []string{"z"}}

	containsScore := scoreOperation(weakContains, "user")
	fuzzyScoreVal := scoreOperation(closeFuzzy, "user")
	require.Greater(t, containsScore, 0)
	require.Greater(t, fuzzyScoreVal, 0)
	assert.Greater(t, containsScore, fuzzyScoreVal, "any contains-match must outrank a pure fuzzy-match, regardless of field")
}

func TestScoreOperation_DissimilarQueryScoresZero(t *testing.T) {
	op := apispec.Operation{OperationId: "getUserList", Path: "/access/api/v2/users", Summary: "Get User List", Tags: []string{"Users"}}
	assert.Equal(t, 0, scoreOperation(op, "zzzznotreal"))
}

// TestScoreOperation_UnrelatedSimilarLengthWordDoesNotFuzzyMatch guards a real
// false positive found against the full bundle: "evidence" has no contains-match
// anywhere, but its whole-string Levenshtein similarity to "Environments" (0.42)
// used to clear the old, too-low fuzzy threshold, flooding "jf api docs search
// evidence" with unrelated Environments-tagged operations instead of correctly
// reporting no match.
func TestScoreOperation_UnrelatedSimilarLengthWordDoesNotFuzzyMatch(t *testing.T) {
	op := apispec.Operation{OperationId: "getGlobalEnvironments", Path: "/access/api/v1/environments", Summary: "Get Global Environments", Tags: []string{"Environments"}}
	assert.Equal(t, 0, scoreOperation(op, "evidence"))
}

// TestScoreOperation_RealTypoStillFuzzyMatches is the flip side of the above:
// the raised fuzzy threshold must still tolerate genuine near-typos.
func TestScoreOperation_RealTypoStillFuzzyMatches(t *testing.T) {
	op := apispec.Operation{OperationId: "getWorkers", Path: "/worker/api/v1/workers", Summary: "Get Workers", Tags: []string{"Workers"}}
	assert.Greater(t, scoreOperation(op, "workrs"), 0, "a one-letter-dropped typo of 'workers' should still fuzzy-match")
}

func TestFuzzySimilarityThreshold_DefaultAndOverride(t *testing.T) {
	t.Run("unset uses default", func(t *testing.T) {
		t.Setenv(envFuzzySimilarityMin, "")
		assert.Equal(t, fuzzySimilarityMin, fuzzySimilarityThreshold())
	})
	t.Run("valid override wins", func(t *testing.T) {
		t.Setenv(envFuzzySimilarityMin, "0.2")
		assert.Equal(t, 0.2, fuzzySimilarityThreshold())
	})
	for _, invalid := range []string{"not-a-number", "-0.1", "1.5"} {
		t.Run("invalid falls back: "+invalid, func(t *testing.T) {
			t.Setenv(envFuzzySimilarityMin, invalid)
			assert.Equal(t, fuzzySimilarityMin, fuzzySimilarityThreshold())
		})
	}
}

func TestFuzzyScore_RespectsEnvOverride(t *testing.T) {
	// "evidence" vs "environments" sits at ~0.42 similarity -- below the 0.6
	// default (excluded) but above a permissive 0.3 override (included).
	op := apispec.Operation{OperationId: "getGlobalEnvironments", Path: "/access/api/v1/environments", Summary: "Get Global Environments", Tags: []string{"Environments"}}

	t.Setenv(envFuzzySimilarityMin, "")
	assert.Equal(t, 0, scoreOperation(op, "evidence"), "default threshold should reject this coincidental match")

	t.Setenv(envFuzzySimilarityMin, "0.3")
	assert.Greater(t, scoreOperation(op, "evidence"), 0, "a lowered override should admit it")
}

func TestJfApiOneLiner(t *testing.T) {
	assert.Equal(t, "jf api /access/api/v2/users", jfApiOneLiner(apispec.Operation{Method: "GET", Path: "/access/api/v2/users"}))
	assert.Equal(t, "jf api /access/api/v2/users -X POST", jfApiOneLiner(apispec.Operation{Method: "POST", Path: "/access/api/v2/users"}))
}

func TestHasTag(t *testing.T) {
	assert.True(t, hasTag([]string{"Users", "Access"}, "users"))
	assert.False(t, hasTag([]string{"Users"}, "workers"))
}

// newSearchApp builds a minimal cli.App exercising runSearchCmd exactly like
// the real "search" subcommand's flag set, without going through main.go's
// full command tree -- same technique as TestResolveRequestBody in
// cli_test.go.
func newSearchApp(stdOut *bytes.Buffer, capturedErr *error) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: flagTag},
		cli.StringFlag{Name: flagMethod},
		cli.IntFlag{Name: flagLimit, Value: defaultLimit},
		cli.StringFlag{Name: "format"},
	}
	app.Action = func(c *cli.Context) error {
		*capturedErr = runSearchCmd(c, stdOut)
		return nil
	}
	return app
}

// TestRunSearchCmd_DefaultsToJSON verifies JSON is the default output format
// when --format is omitted entirely -- this command exists primarily for
// agent consumption, unlike most other jf commands whose JSON default is
// gated on --ai-help/$JFROG_CLI_AI_HELP. Swaps the shared client logger since
// the JSON path writes via its Output channel, same technique as
// TestApiJSONErrorMode_EmitsJSONOnStdout in cli_test.go.
func TestRunSearchCmd_DefaultsToJSON(t *testing.T) {
	var out bytes.Buffer
	prevLogger := clientlog.GetLogger()
	t.Cleanup(func() { clientlog.SetLogger(prevLogger) })
	clientlog.SetLogger(clientlog.NewLoggerWithFlags(clientlog.INFO, &out, 0))

	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "user"}))
	require.NoError(t, runErr)

	var result map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &result), "default output should be parseable JSON")
	assert.Equal(t, "stub", result["spec_bundle"])
	assert.Empty(t, stdOut.String(), "JSON goes through the logger's Output channel, not the stdOut writer")
}

func TestRunSearchCmd_TableOutput(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "user"}))
	require.NoError(t, runErr)
	assert.Contains(t, stdOut.String(), "METHOD")
	assert.Contains(t, stdOut.String(), "/access/api/v2/users")
}

func TestRunSearchCmd_EmptyResultTableStillReportsSpecBundle(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "zzzznotreal"}))
	require.NoError(t, runErr, "empty results must not be treated as a command failure")
	assert.Contains(t, stdOut.String(), "spec_bundle=")
	assert.Contains(t, stdOut.String(), "stub")
}

func TestRunSearchCmd_LimitTruncates(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "--limit", "1", ""}))
	require.NoError(t, runErr)
	// header + exactly one data row
	lineCount := 0
	for _, b := range stdOut.Bytes() {
		if b == '\n' {
			lineCount++
		}
	}
	assert.Equal(t, 2, lineCount, "expected a header row plus exactly one match row")
}

func TestRunSearchCmd_WrongNumberOfArguments(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd"}))
	assert.Error(t, runErr)
}

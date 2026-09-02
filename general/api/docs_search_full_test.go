//go:build full

// Integration tests for `jf api docs search` against the embedded full
// OpenAPI bundle (populated from rdme-admin at release time). These only run
// under `go test -tags full`, same gate as docs/api-spec/parser_full_test.go.

package api

import (
	"bytes"
	"strconv"
	"testing"

	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fullOps(t *testing.T) []apispec.Operation {
	t.Helper()
	ops, err := apispec.Operations()
	require.NoError(t, err)
	require.NotEmpty(t, ops, "full bundle must be populated before running -tags full tests")
	return ops
}

func TestRunSearchCmd_DefaultsToJSON(t *testing.T) {
	result, _ := runSearchJSON(t, "permission")
	assert.Equal(t, "full", result["spec_bundle"])
	assert.NotEmpty(t, result["spec_version"])
	matches, ok := result["matches"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, matches, "full bundle should match 'permission'")
}

func TestRunSearchCmd_TableOutput(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "permission"}))
	require.NoError(t, runErr)
	assert.Contains(t, stdOut.String(), "METHOD")
	assert.NotContains(t, stdOut.String(), "stub")
}

func TestRunSearchCmd_EmptyResultStillReportsFullBundle(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "zzzznotreal"}))
	require.NoError(t, runErr, "empty results must not be treated as a command failure")
	assert.Contains(t, stdOut.String(), "spec_bundle=full")
}

func TestRunSearchCmd_LimitTruncates(t *testing.T) {
	fullOps(t)

	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "--limit", "1", ""}))
	require.NoError(t, runErr)

	lineCount := 0
	for _, b := range stdOut.Bytes() {
		if b == '\n' {
			lineCount++
		}
	}
	assert.Equal(t, 2, lineCount, "expected a header row plus exactly one match row")
}

func TestRunSearchCmd_TruncationFieldsInJSON(t *testing.T) {
	ops := fullOps(t)
	if len(ops) <= 1 {
		t.Skip("truncation requires more than one operation in the full bundle")
	}

	result, logged := runSearchJSON(t, "--limit", "1", "")
	assert.Equal(t, float64(len(ops)), result["total_matches"])
	assert.Equal(t, true, result["truncated"])
	assert.Len(t, result["matches"], 1)
	assert.Contains(t, logged, "of ", "truncation warning should mention the full match count")
}

func TestRunSearchCmd_NoTruncationWhenUnderLimit(t *testing.T) {
	ops := fullOps(t)

	result, logged := runSearchJSON(t, "--limit", strconv.Itoa(len(ops)+10), "")
	assert.Equal(t, float64(len(ops)), result["total_matches"])
	assert.Equal(t, false, result["truncated"])
	assert.Len(t, result["matches"], len(ops))
	assert.Empty(t, logged, "no truncation warning expected when everything fits under the limit")
}

// TestRunSearchCmd_TruncationWarningDoesNotLeakIntoTable guards that the
// truncation warning goes to stderr only, never into the table stdOut writer.
func TestRunSearchCmd_TruncationWarningDoesNotLeakIntoTable(t *testing.T) {
	fullOps(t)

	var stdOut bytes.Buffer
	var runErr error
	app := newSearchApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "--limit", "1", ""}))
	require.NoError(t, runErr)

	lineCount := 0
	for _, b := range stdOut.Bytes() {
		if b == '\n' {
			lineCount++
		}
	}
	assert.Equal(t, 2, lineCount, "the stdOut table must still be exactly header + one row")
	assert.NotContains(t, stdOut.String(), "increase --limit")
}

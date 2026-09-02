//go:build full

// Integration tests for `jf api docs describe` against the embedded full
// OpenAPI bundle (populated from rdme-admin at release time). These only run
// under `go test -tags full`, same gate as docs/api-spec/parser_full_test.go.

package api

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDescribeCmd_KnownOperation(t *testing.T) {
	result := runDescribeJSON(t, "GET", "/access/api/v2/users")
	assert.Equal(t, "GET", result["method"])
	assert.Equal(t, "/access/api/v2/users", result["path"])
	assert.Equal(t, "full", result["spec_bundle"])
	assert.NotEmpty(t, result["spec_version"])
	assert.NotEmpty(t, result["jf_api"])
}

func TestRunDescribeCmd_CaseInsensitiveMethod(t *testing.T) {
	result := runDescribeJSON(t, "get", "/access/api/v2/users")
	assert.Equal(t, "GET", result["method"])
}

func TestRunDescribeCmd_PathWithoutLeadingSlashNormalizes(t *testing.T) {
	result := runDescribeJSON(t, "GET", "access/api/v2/users")
	assert.Equal(t, "/access/api/v2/users", result["path"])
}

func TestRunDescribeCmd_NotFoundReturnsError(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "GET", "/not/a/real/path"}))
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), "no operation found")
	assert.NotContains(t, runErr.Error(), `"stub"`)
}

func TestRunDescribeCmd_TableOutput(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "POST", "/access/api/v2/users"}))
	require.NoError(t, runErr)
	assert.Contains(t, stdOut.String(), "METHOD")
	assert.Contains(t, stdOut.String(), "POST")
	assert.Contains(t, stdOut.String(), "JF API")
}

// TestSearchThenDescribe_EndToEnd guards the intended agent flow on a full
// bundle: search → describe must agree on method/path/jf_api for the top hit.
func TestSearchThenDescribe_EndToEnd(t *testing.T) {
	result, _ := runSearchJSON(t, "user")
	matches, ok := result["matches"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, matches)

	top, ok := matches[0].(map[string]any)
	require.True(t, ok)
	method, _ := top["method"].(string)
	path, _ := top["path"].(string)
	jfApi, _ := top["jf_api"].(string)
	require.NotEmpty(t, method)
	require.NotEmpty(t, path)

	described := runDescribeJSON(t, method, path)
	assert.Equal(t, method, described["method"])
	assert.Equal(t, path, described["path"])
	assert.Equal(t, jfApi, described["jf_api"])
}

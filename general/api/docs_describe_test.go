package api

import (
	"bytes"
	"encoding/json"
	"testing"

	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestNormalizeApiPath(t *testing.T) {
	assert.Equal(t, "/access/api/v2/users", normalizeApiPath("/access/api/v2/users"))
	assert.Equal(t, "/access/api/v2/users", normalizeApiPath("access/api/v2/users"))
	assert.Equal(t, "", normalizeApiPath(""))
	assert.Equal(t, "", normalizeApiPath("   "))
}

func TestFormatResponses(t *testing.T) {
	assert.Equal(t, "-", formatResponses(nil))
	assert.Equal(t, "200:Success, 400", formatResponses([]apispec.Response{
		{Code: "200", Description: "Success"},
		{Code: "400"},
	}))
}

// newDescribeApp builds a minimal cli.App exercising runDescribeCmd exactly
// like the real "describe" subcommand's flag set -- same technique as
// newSearchApp in docs_search_test.go.
func newDescribeApp(stdOut *bytes.Buffer, capturedErr *error) *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "format"},
	}
	app.Action = func(c *cli.Context) error {
		*capturedErr = runDescribeCmd(c, stdOut)
		return nil
	}
	return app
}

func TestRunDescribeCmd_KnownGetOperation(t *testing.T) {
	result, _ := runDescribeJSON(t, "GET", "/access/api/v2/users")
	assert.Equal(t, "GET", result["method"])
	assert.Equal(t, "/access/api/v2/users", result["path"])
	assert.Equal(t, "stub", result["spec_bundle"])
	assert.NotEmpty(t, result["parameters"])
	assert.Nil(t, result["request_body"])
	assert.NotEmpty(t, result["responses"])
	assert.Equal(t, "jf api /access/api/v2/users", result["jf_api"])
}

func TestRunDescribeCmd_KnownPostOperation(t *testing.T) {
	result, _ := runDescribeJSON(t, "POST", "/access/api/v2/users")
	assert.Equal(t, "POST", result["method"])

	requestBody, ok := result["request_body"].(map[string]any)
	require.True(t, ok, "createUser should carry a request_body")
	assert.True(t, requestBody["required"].(bool))
	properties, ok := requestBody["properties"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, properties)

	jfApi, ok := result["jf_api"].(string)
	require.True(t, ok)
	assert.Contains(t, jfApi, "-X POST")
	assert.Contains(t, jfApi, "-d '")
}

func TestRunDescribeCmd_CaseInsensitiveMethod(t *testing.T) {
	result, _ := runDescribeJSON(t, "get", "/access/api/v2/users")
	assert.Equal(t, "GET", result["method"])
}

func TestRunDescribeCmd_PathWithoutLeadingSlashNormalizes(t *testing.T) {
	result, _ := runDescribeJSON(t, "GET", "access/api/v2/users")
	assert.Equal(t, "/access/api/v2/users", result["path"])
}

func TestRunDescribeCmd_NotFoundReturnsError(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "GET", "/not/a/real/path"}))
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), "spec_bundle")
	assert.Contains(t, runErr.Error(), "docs search")
}

func TestRunDescribeCmd_WrongNumberOfArguments(t *testing.T) {
	for _, args := range [][]string{
		{"cmd"},
		{"cmd", "GET"},
		{"cmd", "GET", "/a", "extra"},
	} {
		var stdOut bytes.Buffer
		var runErr error
		app := newDescribeApp(&stdOut, &runErr)
		require.NoError(t, app.Run(args))
		assert.Error(t, runErr, "args %v should be rejected", args)
	}
}

func TestRunDescribeCmd_TableOutput(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "--format", "table", "POST", "/access/api/v2/users"}))
	require.NoError(t, runErr)
	assert.Contains(t, stdOut.String(), "METHOD")
	assert.Contains(t, stdOut.String(), "POST")
	assert.Contains(t, stdOut.String(), "REQUEST BODY")
	assert.Contains(t, stdOut.String(), "RESPONSES")
	assert.Contains(t, stdOut.String(), "JF API")
}

// TestRunDescribeCmd_DefaultsToJSON verifies JSON is the default output format
// when --format is omitted, matching docs search's unconditional-JSON-default
// convention (see TestRunSearchCmd_DefaultsToJSON).
func TestRunDescribeCmd_DefaultsToJSON(t *testing.T) {
	var out bytes.Buffer
	prevLogger := clientlog.GetLogger()
	t.Cleanup(func() { clientlog.SetLogger(prevLogger) })
	clientlog.SetLogger(clientlog.NewLoggerWithFlags(clientlog.INFO, &out, 0))

	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "GET", "/access/api/v2/users"}))
	require.NoError(t, runErr)

	var result map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &result), "default output should be parseable JSON")
	assert.Equal(t, "stub", result["spec_bundle"])
	assert.Empty(t, stdOut.String(), "JSON goes through the logger's Output channel, not the stdOut writer")
}

// TestSearchThenDescribe_EndToEnd guards the intended agent flow: a search
// result's method+path must resolve cleanly through describe, and describe's
// jf_api one-liner must match search's one-liner for the same operation (both
// call the shared jfApiOneLiner helper).
func TestSearchThenDescribe_EndToEnd(t *testing.T) {
	matches := filterAndScore(stubOps(t), "user", "", "")
	require.NotEmpty(t, matches)
	top := matches[0]

	result, _ := runDescribeJSON(t, top.Method, top.Path)
	assert.Equal(t, top.Method, result["method"])
	assert.Equal(t, top.Path, result["path"])
	assert.Equal(t, top.JfApi, result["jf_api"])
}

// runDescribeJSON runs the describe app with JSON output (the default) and
// returns the parsed result body -- same technique as runSearchJSON in
// docs_search_test.go.
func runDescribeJSON(t *testing.T, method, path string) (result map[string]any, logged string) {
	t.Helper()
	var jsonOut, logOut bytes.Buffer
	logger := clientlog.NewLoggerWithFlags(clientlog.INFO, &logOut, 0)
	logger.SetOutputWriter(&jsonOut)
	prevLogger := clientlog.GetLogger()
	t.Cleanup(func() { clientlog.SetLogger(prevLogger) })
	clientlog.SetLogger(logger)

	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", method, path}))
	require.NoError(t, runErr)

	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &result), "output should be parseable JSON")
	return result, logOut.String()
}

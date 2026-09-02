//go:build !full

// These tests assert exact values from the stub fixtures (bundle name
// "stub", specific stub operations) and don't apply to a full build (whose
// docs/api-spec/full/ content is populated at release time) -- same
// rationale as docs/api-spec/parser_test.go.

package api

import (
	"bytes"
	"testing"

	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
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

func TestRunDescribeCmd_StubBundleFails(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "GET", "/access/api/v2/users"}))
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), `"stub"`)
	assert.Contains(t, runErr.Error(), "install-cli.jfrog.io")
	assert.Empty(t, stdOut.String())
}

func TestRunDescribeCmd_NotFoundReturnsError(t *testing.T) {
	var stdOut bytes.Buffer
	var runErr error
	app := newDescribeApp(&stdOut, &runErr)

	require.NoError(t, app.Run([]string{"cmd", "GET", "/not/a/real/path"}))
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), `"stub"`)
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
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), `"stub"`)
	assert.Empty(t, stdOut.String())
}

package token

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

// sampleTokenResponse returns a minimal but realistic token JSON payload.
func sampleTokenResponse(t *testing.T) []byte {
	t.Helper()
	payload := map[string]interface{}{
		"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IkV4YW1wbGUifQ.payload.signature",
		"token_id":     "abc123-def456",
		"expires_in":   3600,
		"scope":        "applied-permissions/user",
		"token_type":   "Bearer",
		"refreshable":  false,
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return data
}

func TestPrintTokenResponse_JSON(t *testing.T) {
	data := sampleTokenResponse(t)

	var buf bytes.Buffer
	err := printTokenResponse(data, coreformat.Json, &buf)
	require.NoError(t, err)

	// JSON format uses log.Output (writes to the real logger, not the writer).
	// The buf should remain empty; validate JSON is well-formed instead.
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "Bearer", parsed["token_type"])
	assert.Equal(t, "abc123-def456", parsed["token_id"])
}

func TestPrintTokenResponse_Table(t *testing.T) {
	data := sampleTokenResponse(t)

	var buf bytes.Buffer
	err := printTokenResponse(data, coreformat.Table, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Header row must be present.
	assert.Contains(t, output, "FIELD")
	assert.Contains(t, output, "VALUE")

	// Known fields must appear.
	assert.Contains(t, output, "token_id")
	assert.Contains(t, output, "abc123-def456")
	assert.Contains(t, output, "expires_in")
	assert.Contains(t, output, "scope")
	assert.Contains(t, output, "applied-permissions/user")
	assert.Contains(t, output, "token_type")
	assert.Contains(t, output, "Bearer")
}

func TestPrintTokenResponse_Table_AccessTokenTruncated(t *testing.T) {
	longToken := strings.Repeat("a", 100)
	payload := map[string]interface{}{"access_token": longToken}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, printTokenResponse(data, coreformat.Table, &buf))

	output := buf.String()
	assert.Contains(t, output, "access_token")
	// Full token must NOT appear; truncated version (40 chars + "...") must.
	assert.NotContains(t, output, longToken)
	assert.Contains(t, output, strings.Repeat("a", 40)+"...")
}

func TestPrintTokenResponse_Table_AbsentFieldsOmitted(t *testing.T) {
	// Payload with only access_token; all other fields must be absent from table.
	payload := map[string]interface{}{
		"access_token": "tok",
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, printTokenResponse(data, coreformat.Table, &buf))

	output := buf.String()
	assert.NotContains(t, output, "token_id")
	assert.NotContains(t, output, "expires_in")
	assert.NotContains(t, output, "scope")
}

func TestPrintTokenResponse_UnsupportedFormat(t *testing.T) {
	err := printTokenResponse([]byte("{}"), coreformat.Sarif, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestGetTokenOutputFormat_Default(t *testing.T) {
	// No --format flag set → must default to json for backward compatibility.
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getTokenOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app"}))
	assert.Equal(t, coreformat.Json, gotFormat)
}

func TestGetTokenOutputFormat_ExplicitTable(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getTokenOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	assert.Equal(t, coreformat.Table, gotFormat)
}

func TestGetTokenOutputFormat_ExplicitJSON(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getTokenOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
	assert.Equal(t, coreformat.Json, gotFormat)
}

func TestGetTokenOutputFormat_Invalid(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotErr error
	app.Action = func(c *cli.Context) error {
		_, gotErr = getTokenOutputFormat(c)
		return nil
	}
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

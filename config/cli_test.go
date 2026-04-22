package config

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newFormatContext creates a *cli.Context with --format set to formatVal.
// If formatVal is empty, the flag is registered but left unset.
func newFormatContext(formatVal string) *cli.Context {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("format", "", "")
	if formatVal != "" {
		_ = fs.Set("format", formatVal)
	}
	return cli.NewContext(app, fs, nil)
}

// ---------------------------------------------------------------------------
// printConfigShowTable
// ---------------------------------------------------------------------------

func TestPrintConfigShowTable_Basic(t *testing.T) {
	configs := []*coreconfig.ServerDetails{
		{ServerId: "my-server", ArtifactoryUrl: "https://acme.jfrog.io/artifactory"},
	}
	var buf bytes.Buffer
	err := printConfigShowTable(configs, &buf)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "FIELD")
	assert.Contains(t, out, "VALUE")
	assert.Contains(t, out, "server_id")
	assert.Contains(t, out, "my-server")
	assert.Contains(t, out, "artifactory_url")
	assert.Contains(t, out, "https://acme.jfrog.io/artifactory")
}

func TestPrintConfigShowTable_MasksSensitiveFields(t *testing.T) {
	configs := []*coreconfig.ServerDetails{
		{
			ServerId:                "masked",
			Password:                "secret-password",
			AccessToken:             "secret-token",
			RefreshToken:            "secret-refresh",
			SshPassphrase:           "secret-passphrase",
			ArtifactoryRefreshToken: "secret-art-refresh",
		},
	}
	var buf bytes.Buffer
	err := printConfigShowTable(configs, &buf)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "***")
	assert.NotContains(t, out, "secret-password")
	assert.NotContains(t, out, "secret-token")
	assert.NotContains(t, out, "secret-refresh")
	assert.NotContains(t, out, "secret-passphrase")
}

func TestPrintConfigShowTable_OmitsEmptyFields(t *testing.T) {
	configs := []*coreconfig.ServerDetails{
		{ServerId: "only-id"},
	}
	var buf bytes.Buffer
	err := printConfigShowTable(configs, &buf)
	require.NoError(t, err)
	out := buf.String()
	// Fields that are not set should not appear.
	assert.NotContains(t, out, "artifactory_url")
	assert.NotContains(t, out, "xray_url")
	assert.NotContains(t, out, "user")
	assert.NotContains(t, out, "password")
	assert.NotContains(t, out, "access_token")
	assert.Contains(t, out, "only-id")
}

func TestPrintConfigShowTable_MultipleConfigs(t *testing.T) {
	configs := []*coreconfig.ServerDetails{
		{ServerId: "server-one"},
		{ServerId: "server-two"},
	}
	var buf bytes.Buffer
	err := printConfigShowTable(configs, &buf)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "---")
	assert.Contains(t, out, "server-one")
	assert.Contains(t, out, "server-two")
}

// ---------------------------------------------------------------------------
// printConfigShowJSON
// ---------------------------------------------------------------------------

func TestPrintConfigShowJSON_NoError(t *testing.T) {
	configs := []*coreconfig.ServerDetails{
		{ServerId: "json-server", ArtifactoryUrl: "https://acme.jfrog.io/artifactory"},
		{ServerId: "json-server-2"},
	}
	// JSON is written to log.Output, not to w — just verify no error is returned.
	err := printConfigShowJSON(configs)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// printConfigShowResponse
// ---------------------------------------------------------------------------

func TestPrintConfigShowResponse_UnsupportedFormat(t *testing.T) {
	configs := []*coreconfig.ServerDetails{{ServerId: "test"}}
	var buf bytes.Buffer
	err := printConfigShowResponse(configs, coreformat.Sarif, &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

// ---------------------------------------------------------------------------
// getConfigShowOutputFormat
// ---------------------------------------------------------------------------

func TestGetConfigShowOutputFormat_Default(t *testing.T) {
	// No --format flag set at all.
	c := newFormatContext("")
	f, err := getConfigShowOutputFormat(c)
	require.NoError(t, err)
	assert.Equal(t, coreformat.Table, f)
}

func TestGetConfigShowOutputFormat_ExplicitJSON(t *testing.T) {
	c := newFormatContext("json")
	f, err := getConfigShowOutputFormat(c)
	require.NoError(t, err)
	assert.Equal(t, coreformat.Json, f)
}

func TestGetConfigShowOutputFormat_ExplicitTable(t *testing.T) {
	c := newFormatContext("table")
	f, err := getConfigShowOutputFormat(c)
	require.NoError(t, err)
	assert.Equal(t, coreformat.Table, f)
}

func TestGetConfigShowOutputFormat_Invalid(t *testing.T) {
	c := newFormatContext("xml")
	_, err := getConfigShowOutputFormat(c)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "only the following output formats are supported"))
}

// ---------------------------------------------------------------------------
// sanitizeServerDetails
// ---------------------------------------------------------------------------

func TestSanitizeServerDetails(t *testing.T) {
	original := &coreconfig.ServerDetails{
		ServerId:                "my-server",
		Password:                "real-password",
		SshPassphrase:           "real-passphrase",
		AccessToken:             "real-access-token",
		RefreshToken:            "real-refresh-token",
		ArtifactoryRefreshToken: "real-art-refresh-token",
		User:                    "admin",
	}
	sanitized := sanitizeServerDetails(original)

	// Sensitive fields must be masked.
	assert.Equal(t, "***", sanitized.Password)
	assert.Equal(t, "***", sanitized.SshPassphrase)
	assert.Equal(t, "***", sanitized.AccessToken)
	assert.Equal(t, "***", sanitized.RefreshToken)
	assert.Equal(t, "***", sanitized.ArtifactoryRefreshToken)

	// Non-sensitive fields must be preserved.
	assert.Equal(t, "my-server", sanitized.ServerId)
	assert.Equal(t, "admin", sanitized.User)

	// Original must be unchanged.
	assert.Equal(t, "real-password", original.Password)
	assert.Equal(t, "real-access-token", original.AccessToken)
}

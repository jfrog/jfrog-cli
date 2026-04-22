package missioncontrol

import (
	"bytes"
	"testing"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestPrintLicenseAcquireResponse_Table(t *testing.T) {
	var buf bytes.Buffer
	err := printLicenseAcquireResponse("ABCD-1234-EFGH-5678", coreformat.Table, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "FIELD")
	assert.Contains(t, output, "VALUE")
	assert.Contains(t, output, "license_key")
	assert.Contains(t, output, "ABCD-1234-EFGH-5678")
}

func TestPrintLicenseAcquireResponse_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := printLicenseAcquireResponse("ABCD-1234-EFGH-5678", coreformat.Json, &buf)
	require.NoError(t, err)
	// JSON goes via log.Output; verify no error and no table output.
	assert.Empty(t, buf.String())
}

func TestPrintLicenseAcquireResponse_UnsupportedFormat(t *testing.T) {
	err := printLicenseAcquireResponse("key", coreformat.Sarif, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestGetLicenseAcquireOutputFormat_Default(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getLicenseAcquireOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app"}))
	assert.Equal(t, coreformat.Table, gotFormat)
}

func TestGetLicenseAcquireOutputFormat_ExplicitJSON(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getLicenseAcquireOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
	assert.Equal(t, coreformat.Json, gotFormat)
}

func TestGetLicenseAcquireOutputFormat_ExplicitTable(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getLicenseAcquireOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	assert.Equal(t, coreformat.Table, gotFormat)
}

func TestGetLicenseAcquireOutputFormat_Invalid(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotErr error
	app.Action = func(c *cli.Context) error {
		_, gotErr = getLicenseAcquireOutputFormat(c)
		return nil
	}
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

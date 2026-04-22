package missioncontrol

import (
	"bytes"
	"testing"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

// ---- jpd-add tests ----

func TestPrintJpdAddResponse_JSON(t *testing.T) {
	var buf bytes.Buffer
	body := []byte(`{"id":"jpd1","name":"My JPD","url":"https://my.jfrog.io"}`)
	err := printJpdAddResponse(body, coreformat.Json, &buf)
	require.NoError(t, err)
	// JSON goes via log.Output; verify no error and no table output.
	assert.Empty(t, buf.String())
}

func TestPrintJpdAddResponse_Table(t *testing.T) {
	var buf bytes.Buffer
	body := []byte(`{"id":"jpd1","name":"My JPD","url":"https://my.jfrog.io","location":"US","type":"EDGE"}`)
	err := printJpdAddResponse(body, coreformat.Table, &buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "FIELD")
	assert.Contains(t, output, "VALUE")
	assert.Contains(t, output, "id")
	assert.Contains(t, output, "jpd1")
	assert.Contains(t, output, "name")
	assert.Contains(t, output, "My JPD")
}

func TestPrintJpdAddResponse_Table_UnknownKeys(t *testing.T) {
	var buf bytes.Buffer
	body := []byte(`{"custom_field":"custom_value","another_field":"another_value"}`)
	err := printJpdAddResponse(body, coreformat.Table, &buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "custom_field")
	assert.Contains(t, output, "custom_value")
	assert.Contains(t, output, "another_field")
	assert.Contains(t, output, "another_value")
}

func TestPrintJpdAddResponse_UnsupportedFormat(t *testing.T) {
	err := printJpdAddResponse([]byte(`{}`), coreformat.Sarif, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestGetJpdAddOutputFormat_Default(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getJpdAddOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app"}))
	assert.Equal(t, coreformat.Json, gotFormat)
}

func TestGetJpdAddOutputFormat_ExplicitTable(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getJpdAddOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	assert.Equal(t, coreformat.Table, gotFormat)
}

func TestGetJpdAddOutputFormat_ExplicitJSON(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getJpdAddOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
	assert.Equal(t, coreformat.Json, gotFormat)
}

func TestGetJpdAddOutputFormat_Invalid(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotErr error
	app.Action = func(c *cli.Context) error {
		_, gotErr = getJpdAddOutputFormat(c)
		return nil
	}
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

// ---- license-deploy tests ----

func TestPrintLicenseDeployResponse_JSON(t *testing.T) {
	var buf bytes.Buffer
	body := []byte(`{"bucket_id":"bucket1","jpd_id":"jpd1","license_count":5}`)
	err := printLicenseDeployResponse(body, coreformat.Json, &buf)
	require.NoError(t, err)
	// JSON goes via log.Output; verify no error and no table output.
	assert.Empty(t, buf.String())
}

func TestPrintLicenseDeployResponse_Table(t *testing.T) {
	var buf bytes.Buffer
	body := []byte(`{"bucket_id":"bucket1","jpd_id":"jpd1","license_count":5,"status":"deployed"}`)
	err := printLicenseDeployResponse(body, coreformat.Table, &buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "FIELD")
	assert.Contains(t, output, "VALUE")
	assert.Contains(t, output, "bucket_id")
	assert.Contains(t, output, "bucket1")
	assert.Contains(t, output, "jpd_id")
	assert.Contains(t, output, "jpd1")
}

func TestPrintLicenseDeployResponse_Table_UnknownKeys(t *testing.T) {
	var buf bytes.Buffer
	body := []byte(`{"extra_field":"extra_value","another_field":"another_value"}`)
	err := printLicenseDeployResponse(body, coreformat.Table, &buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "extra_field")
	assert.Contains(t, output, "extra_value")
	assert.Contains(t, output, "another_field")
	assert.Contains(t, output, "another_value")
}

func TestPrintLicenseDeployResponse_UnsupportedFormat(t *testing.T) {
	err := printLicenseDeployResponse([]byte(`{}`), coreformat.Sarif, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestGetLicenseDeployOutputFormat_Default(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getLicenseDeployOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app"}))
	assert.Equal(t, coreformat.Json, gotFormat)
}

func TestGetLicenseDeployOutputFormat_ExplicitTable(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getLicenseDeployOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	assert.Equal(t, coreformat.Table, gotFormat)
}

func TestGetLicenseDeployOutputFormat_ExplicitJSON(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getLicenseDeployOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
	assert.Equal(t, coreformat.Json, gotFormat)
}

func TestGetLicenseDeployOutputFormat_Invalid(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotErr error
	app.Action = func(c *cli.Context) error {
		_, gotErr = getLicenseDeployOutputFormat(c)
		return nil
	}
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

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

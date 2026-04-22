package pipelines

import (
	"bytes"
	"testing"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/jfrog/jfrog-client-go/pipelines/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func samplePipelineStatusResponse() *services.PipelineRunStatusResponse {
	return &services.PipelineRunStatusResponse{
		TotalCount: 1,
		Pipelines: []services.Pipelines{
			{
				Name:                 "my-pipeline",
				PipelineSourceBranch: "main",
				LatestRunID:          42,
				Run: services.Run{
					RunNumber:       7,
					StatusCode:      4004,
					DurationSeconds: 120,
				},
			},
		},
	}
}

func TestPrintPipelineStatusResponse_Table(t *testing.T) {
	var buf bytes.Buffer
	err := printPipelineStatusResponse(samplePipelineStatusResponse(), coreformat.Table, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "PIPELINE")
	assert.Contains(t, output, "BRANCH")
	assert.Contains(t, output, "RUN")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "my-pipeline")
	assert.Contains(t, output, "main")
	assert.Contains(t, output, "7")
}

func TestPrintPipelineStatusResponse_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := printPipelineStatusResponse(samplePipelineStatusResponse(), coreformat.Json, &buf)
	require.NoError(t, err)
	// JSON goes via log.Output; buf remains empty.
	assert.Empty(t, buf.String())
}

func TestPrintPipelineStatusResponse_Table_SkipsUnrunPipelines(t *testing.T) {
	resp := &services.PipelineRunStatusResponse{
		Pipelines: []services.Pipelines{
			{Name: "never-ran", LatestRunID: 0},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, printPipelineStatusResponse(resp, coreformat.Table, &buf))
	assert.NotContains(t, buf.String(), "never-ran")
}

func TestPrintPipelineStatusResponse_UnsupportedFormat(t *testing.T) {
	err := printPipelineStatusResponse(samplePipelineStatusResponse(), coreformat.Sarif, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestGetPipelineStatusOutputFormat_Default(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getPipelineStatusOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app"}))
	assert.Equal(t, coreformat.Table, gotFormat)
}

func TestGetPipelineStatusOutputFormat_ExplicitJSON(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getPipelineStatusOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "json"}))
	assert.Equal(t, coreformat.Json, gotFormat)
}

func TestGetPipelineStatusOutputFormat_ExplicitTable(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotFormat coreformat.OutputFormat
	app.Action = func(c *cli.Context) error {
		var err error
		gotFormat, err = getPipelineStatusOutputFormat(c)
		return err
	}
	require.NoError(t, app.Run([]string{"app", "--format", "table"}))
	assert.Equal(t, coreformat.Table, gotFormat)
}

func TestGetPipelineStatusOutputFormat_Invalid(t *testing.T) {
	app := cli.NewApp()
	app.Flags = []cli.Flag{cli.StringFlag{Name: "format"}}
	var gotErr error
	app.Action = func(c *cli.Context) error {
		_, gotErr = getPipelineStatusOutputFormat(c)
		return nil
	}
	require.NoError(t, app.Run([]string{"app", "--format", "xml"}))
	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "only the following output formats are supported")
}

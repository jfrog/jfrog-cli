package pipelines

import (
	"bytes"
	"testing"
	"time"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	"github.com/jfrog/jfrog-client-go/pipelines/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// --- pl sync-status tests ---

func sampleSyncStatuses() []services.PipelineSyncStatus {
	isSyncing := true
	return []services.PipelineSyncStatus{
		{
			PipelineSourceBranch: "main",
			LastSyncStatusCode:   4002,
			IsSyncing:            &isSyncing,
			LastSyncStartedAt:    time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC),
			LastSyncEndedAt:      time.Date(2026, 1, 2, 15, 6, 5, 0, time.UTC),
			CommitData: services.CommitData{
				CommitSha: "abc123",
			},
		},
	}
}

func TestPrintSyncStatusResponse_Table(t *testing.T) {
	var buf bytes.Buffer
	err := printSyncStatusResponse(sampleSyncStatuses(), coreformat.Table, &buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "BRANCH")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "main")
	assert.Contains(t, output, "abc123")
}

func TestPrintSyncStatusResponse_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := printSyncStatusResponse(sampleSyncStatuses(), coreformat.Json, &buf)
	require.NoError(t, err)
	// JSON goes via log.Output; buf remains empty.
	assert.Empty(t, buf.String())
}

func TestPrintSyncStatusResponse_UnsupportedFormat(t *testing.T) {
	err := printSyncStatusResponse(sampleSyncStatuses(), coreformat.Sarif, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}


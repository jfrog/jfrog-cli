package cliutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FormatHTTPResponseJSON writes to log.Output, not to an io.Writer, so these
// tests verify the function runs without panicking and handles edge cases.

func TestFormatHTTPResponseJSON_ValidJSON(t *testing.T) {
	// Should not panic on valid JSON body
	assert.NotPanics(t, func() {
		FormatHTTPResponseJSON([]byte(`{"run_id":42,"pipeline":"my-pipeline"}`), 200)
	})
}

func TestFormatHTTPResponseJSON_EmptyBody(t *testing.T) {
	// nil body → synthesizes {"status_code":200,"message":"OK"}
	assert.NotPanics(t, func() {
		FormatHTTPResponseJSON(nil, 200)
	})
}

func TestFormatHTTPResponseJSON_NonJSONBody(t *testing.T) {
	// Plain-text body → synthesizes status object with message = body text
	assert.NotPanics(t, func() {
		FormatHTTPResponseJSON([]byte("Internal Server Error"), 500)
	})
}

func TestFormatHTTPResponseJSON_InvalidJSONFallback(t *testing.T) {
	// Partial/corrupt JSON → falls back to synthetic object
	assert.NotPanics(t, func() {
		FormatHTTPResponseJSON([]byte(`{"broken`), 400)
	})
}

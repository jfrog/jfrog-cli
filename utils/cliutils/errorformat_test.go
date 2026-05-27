package cliutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/stretchr/testify/assert"
)

func newHTTPErr(status int, statusText string, body []byte) error {
	resp := &http.Response{StatusCode: status, Status: statusText}
	return errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK)
}

func TestHandleHTTPErrorAsJSON_DisabledByDefault(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "")
	var buf bytes.Buffer
	err := newHTTPErr(401, "401 Unauthorized", []byte(`{"x":1}`))
	assert.False(t, HandleHTTPErrorAsJSON(&buf, err))
	assert.Empty(t, buf.String())
}

func TestHandleHTTPErrorAsJSON_TextValueDoesNothing(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "text")
	var buf bytes.Buffer
	err := newHTTPErr(500, "500 Internal Server Error", []byte(`{"x":1}`))
	assert.False(t, HandleHTTPErrorAsJSON(&buf, err))
	assert.Empty(t, buf.String())
}

func TestHandleHTTPErrorAsJSON_NilErr(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "json")
	var buf bytes.Buffer
	assert.False(t, HandleHTTPErrorAsJSON(&buf, nil))
	assert.Empty(t, buf.String())
}

func TestHandleHTTPErrorAsJSON_NonHTTPErrorFallsThrough(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "json")
	var buf bytes.Buffer
	assert.False(t, HandleHTTPErrorAsJSON(&buf, errors.New("plain")))
	assert.Empty(t, buf.String())
}

func TestHandleHTTPErrorAsJSON_EmitsParsedBody(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "json")
	var buf bytes.Buffer
	body := []byte(`{"errors":[{"code":"UNAUTHORIZED","message":"bad creds"}]}`)
	err := newHTTPErr(http.StatusUnauthorized, "401 Unauthorized", body)
	assert.True(t, HandleHTTPErrorAsJSON(&buf, err))

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.EqualValues(t, 401, out["status_code"])
	assert.Equal(t, "401 Unauthorized", out["status"])
	assert.NotNil(t, out["body"])
	// body must be embedded as a structured object, not a string.
	_, isObject := out["body"].(map[string]interface{})
	assert.True(t, isObject, "body should be embedded as JSON object, got: %T", out["body"])
}

func TestHandleHTTPErrorAsJSON_NonJSONBodyKeptAsString(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "json")
	var buf bytes.Buffer
	err := newHTTPErr(503, "503 Service Unavailable", []byte("<html>down</html>"))
	assert.True(t, HandleHTTPErrorAsJSON(&buf, err))

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.EqualValues(t, 503, out["status_code"])
	assert.Equal(t, "<html>down</html>", out["body"])
}

func TestHandleHTTPErrorAsJSON_UnwrapsThroughFmtErrorf(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "json")
	var buf bytes.Buffer
	inner := newHTTPErr(http.StatusForbidden, "403 Forbidden", []byte(`{"msg":"nope"}`))
	wrapped := fmt.Errorf("failed to exchange OIDC token: %w", inner)
	assert.True(t, HandleHTTPErrorAsJSON(&buf, wrapped))

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.EqualValues(t, 403, out["status_code"])
}

func TestHandleHTTPErrorAsJSON_CaseInsensitiveValue(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "  JSON  ")
	var buf bytes.Buffer
	err := newHTTPErr(404, "404 Not Found", nil)
	assert.True(t, HandleHTTPErrorAsJSON(&buf, err))
	assert.Contains(t, buf.String(), `"status_code": 404`)
}

// TestHandleHTTPErrorAsJSON_LegacyTextFallback covers the case where intermediate
// code stripped the typed-error wrap chain (e.g. errors.New(err.Error() + ...)).
// The legacy text format is the only signal left, and we parse it back to JSON.
func TestHandleHTTPErrorAsJSON_LegacyTextFallback(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "json")
	var buf bytes.Buffer
	// Mimic a callsite that flattened the typed error into plain text.
	plain := errors.New("server response: 502 Bad Gateway\n{\"err\":\"upstream down\"}")
	assert.True(t, HandleHTTPErrorAsJSON(&buf, plain))

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.EqualValues(t, 502, out["status_code"])
	assert.Equal(t, "502 Bad Gateway", out["status"])
	body, ok := out["body"].(map[string]interface{})
	assert.True(t, ok, "body must reparse as JSON object")
	assert.Equal(t, "upstream down", body["err"])
}

func TestParseLegacyHTTPResponseError_NoPrefixReturnsNil(t *testing.T) {
	assert.Nil(t, parseLegacyHTTPResponseError("totally unrelated message"))
}

func TestParseLegacyHTTPResponseError_StatusOnly(t *testing.T) {
	got := parseLegacyHTTPResponseError("server response: 204 No Content")
	if assert.NotNil(t, got) {
		assert.Equal(t, 204, got.StatusCode)
		assert.Equal(t, "204 No Content", got.Status)
		assert.Empty(t, got.Body)
	}
}

// TestLegacyHTTPErrorCodeDigitLimits exercises the status-code digit-count
// boundary in parseLegacyHTTPResponseError and the end-to-end behavior in
// HandleHTTPErrorAsJSON. Real HTTP statuses are 3 digits; the parser accepts
// up to 9 (overflow protection under int32) and rejects longer inputs as
// malformed so the caller falls back to the text path rather than emitting
// a truncated status_code the server never sent.
func TestLegacyHTTPErrorCodeDigitLimits(t *testing.T) {
	t.Run("9-digit code accepted (upper boundary)", func(t *testing.T) {
		got := parseLegacyHTTPResponseError("server response: 123456789 Pathological")
		if assert.NotNil(t, got) {
			assert.Equal(t, 123456789, got.StatusCode)
			assert.Equal(t, "123456789 Pathological", got.Status)
		}
	})

	t.Run("10-digit code rejected (one over the cap)", func(t *testing.T) {
		assert.Nil(t, parseLegacyHTTPResponseError("server response: 1234567890 Custom"))
	})

	t.Run("very long code rejected", func(t *testing.T) {
		assert.Nil(t, parseLegacyHTTPResponseError("server response: 99999999999999999999 Custom"))
	})

	t.Run("HandleHTTPErrorAsJSON falls back to text on too-long code", func(t *testing.T) {
		t.Setenv(JfrogCliErrorOutputFormat, "json")
		var buf bytes.Buffer
		plain := errors.New("server response: 1234567890 Custom\n{\"err\":\"x\"}")
		assert.False(t, HandleHTTPErrorAsJSON(&buf, plain))
		assert.Empty(t, buf.String())
	})
}

func TestEnableJSONErrorIfFormatJSON_EqualsForm(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "")
	EnableJSONErrorIfFormatJSON([]string{"jf", "rt", "ping", "--format=json"})
	assert.Equal(t, ErrorFormatJSON, os.Getenv(JfrogCliErrorOutputFormat))
}

func TestEnableJSONErrorIfFormatJSON_SpaceForm(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "")
	EnableJSONErrorIfFormatJSON([]string{"jf", "rt", "search", "--format", "json", "repo/*"})
	assert.Equal(t, ErrorFormatJSON, os.Getenv(JfrogCliErrorOutputFormat))
}

func TestEnableJSONErrorIfFormatJSON_SingleDashEquals(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "")
	EnableJSONErrorIfFormatJSON([]string{"jf", "rt", "ping", "-format=json"})
	assert.Equal(t, ErrorFormatJSON, os.Getenv(JfrogCliErrorOutputFormat))
}

func TestEnableJSONErrorIfFormatJSON_CaseInsensitiveValue(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "")
	EnableJSONErrorIfFormatJSON([]string{"jf", "x", "--format=JSON"})
	assert.Equal(t, ErrorFormatJSON, os.Getenv(JfrogCliErrorOutputFormat))
}

func TestEnableJSONErrorIfFormatJSON_NonJSONValueIgnored(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "")
	EnableJSONErrorIfFormatJSON([]string{"jf", "x", "--format=table"})
	assert.Empty(t, os.Getenv(JfrogCliErrorOutputFormat))
}

func TestEnableJSONErrorIfFormatJSON_NoFlag(t *testing.T) {
	t.Setenv(JfrogCliErrorOutputFormat, "")
	EnableJSONErrorIfFormatJSON([]string{"jf", "rt", "ping"})
	assert.Empty(t, os.Getenv(JfrogCliErrorOutputFormat))
}

func TestEnableJSONErrorIfFormatJSON_LooksLikeFormatInOtherFlag(t *testing.T) {
	// Reject false positives: "--format=json" must be a standalone arg, not
	// part of another flag's value.
	t.Setenv(JfrogCliErrorOutputFormat, "")
	EnableJSONErrorIfFormatJSON([]string{"jf", "rt", "ping", "--header=x-format=json"})
	assert.Empty(t, os.Getenv(JfrogCliErrorOutputFormat))
}

func TestEnableJSONErrorIfFormatJSON_DanglingFormatFlag(t *testing.T) {
	// Trailing "--format" with no value: do not panic, do not set.
	t.Setenv(JfrogCliErrorOutputFormat, "")
	EnableJSONErrorIfFormatJSON([]string{"jf", "rt", "ping", "--format"})
	assert.Empty(t, os.Getenv(JfrogCliErrorOutputFormat))
}

func TestEnableJSONErrorIfFormatJSON_ExplicitEnvWins(t *testing.T) {
	// User explicitly set the env var (to anything) — auto-promotion stays out.
	t.Setenv(JfrogCliErrorOutputFormat, "text")
	EnableJSONErrorIfFormatJSON([]string{"jf", "x", "--format=json"})
	assert.Equal(t, "text", os.Getenv(JfrogCliErrorOutputFormat))
}

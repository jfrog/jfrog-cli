package cliutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// legacyHTTPErrPrefix matches the text produced by errorutils.HttpResponseError.Error()
// (and historically by GenerateResponseError). When intermediate code wraps the
// typed error in a way that strips the chain (e.g. errors.New(err.Error() + ...)),
// we fall back to parsing this prefix so that JSON emission still applies.
const legacyHTTPErrPrefix = "server response: "

// EnableJSONErrorIfFormatJSON inspects args for "--format=json" / "--format json"
// (or the single-dash "-format" variant) and, when found, sets
// JFROG_CLI_ERROR_OUTPUT_FORMAT=json for the lifetime of this process.
//
// Rationale: any command (direct or one dispatched via commands.Exec) that
// already accepts --format=json for its success output should produce
// machine-readable JSON for HTTP error responses too, so a single flag covers
// both code paths and scripts can rely on `jq` end-to-end.
//
// Must be called early in main() — the env var is read by HandleHTTPErrorAsJSON,
// which fires both during command execution (e.g. `jf api`) and at process exit.
// An explicit env var setting wins: if JFROG_CLI_ERROR_OUTPUT_FORMAT is already
// set, the function is a no-op.
func EnableJSONErrorIfFormatJSON(args []string) {
	if os.Getenv(JfrogCliErrorOutputFormat) != "" {
		return
	}
	for i, a := range args {
		if val, ok := matchFormatEquals(a); ok && strings.EqualFold(val, ErrorFormatJSON) {
			_ = os.Setenv(JfrogCliErrorOutputFormat, ErrorFormatJSON)
			return
		}
		if (a == "--format" || a == "-format") && i+1 < len(args) &&
			strings.EqualFold(args[i+1], ErrorFormatJSON) {
			_ = os.Setenv(JfrogCliErrorOutputFormat, ErrorFormatJSON)
			return
		}
	}
}

func matchFormatEquals(a string) (string, bool) {
	for _, prefix := range []string{"--format=", "-format="} {
		if strings.HasPrefix(a, prefix) {
			return a[len(prefix):], true
		}
	}
	return "", false
}

// HandleHTTPErrorAsJSON emits err as a JSON object to w when:
//
//  1. JFROG_CLI_ERROR_OUTPUT_FORMAT is set to "json", and
//  2. err wraps a *errorutils.HttpResponseError (the typed error returned by
//     jfrog-client-go's CheckResponseStatus / CheckResponseStatusWithBody),
//     or its legacy "server response: ..." text equivalent.
//
// w is expected to be the caller's stdout sink: the JSON object is *data* a
// script consumer wants to parse alongside successful command output. Stderr
// continues to carry the human-readable logger output (Info/Warn/Error lines)
// untouched, which a downstream `| jq` pipeline can simply ignore.
//
// Returns true when JSON was emitted. The caller is then responsible for setting
// the process exit code and suppressing any subsequent text logging of the same
// error. Returns false in every other case (env var unset/other value, err nil,
// or err is not an HTTP response error), so callers can fall through to the
// default text reporting path (which uses stderr).
//
// This covers all command-paths that use the standard jfrog-client-go HTTP
// helpers, including OIDC token exchange (errors are wrapped with %w so
// errors.As walks through them).
func HandleHTTPErrorAsJSON(w io.Writer, err error) bool {
	if err == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv(JfrogCliErrorOutputFormat)), ErrorFormatJSON) {
		return false
	}
	httpErr := extractHTTPResponseError(err)
	if httpErr == nil {
		return false
	}
	payload := map[string]interface{}{
		"status_code": httpErr.StatusCode,
		"status":      httpErr.Status,
	}
	if len(httpErr.Body) > 0 {
		var parsed interface{}
		if json.Unmarshal(httpErr.Body, &parsed) == nil {
			payload["body"] = parsed
		} else {
			payload["body"] = string(httpErr.Body)
		}
	}
	encoded, marshalErr := json.MarshalIndent(payload, "", "  ")
	if marshalErr != nil {
		return false
	}
	_, _ = fmt.Fprintln(w, string(encoded))
	return true
}

// extractHTTPResponseError returns the underlying *errorutils.HttpResponseError
// if err wraps one (via errors.As). When no typed error is found, it falls back
// to parsing the legacy "server response: <STATUS>\n<BODY>" text format produced
// by errorutils.GenerateResponseError. The fallback handles cases where
// intermediate code stripped the wrap chain (e.g. errors.New(err.Error() + ...))
// while preserving the human-readable message.
func extractHTTPResponseError(err error) *errorutils.HttpResponseError {
	var httpErr *errorutils.HttpResponseError
	if errors.As(err, &httpErr) && httpErr != nil {
		return httpErr
	}
	return parseLegacyHTTPResponseError(err.Error())
}

func parseLegacyHTTPResponseError(msg string) *errorutils.HttpResponseError {
	idx := strings.Index(msg, legacyHTTPErrPrefix)
	if idx < 0 {
		return nil
	}
	rest := msg[idx+len(legacyHTTPErrPrefix):]
	statusLine, bodyStr, _ := strings.Cut(rest, "\n")
	statusLine = strings.TrimSpace(statusLine)

	// First token must be a numeric status code (e.g. "401 Unauthorized").
	// Cap the digit count so Atoi cannot overflow on malformed input — real
	// HTTP statuses are 3 digits; 9 leaves plenty of headroom under int32.
	codeEnd := 0
	for codeEnd < len(statusLine) && codeEnd < 9 && statusLine[codeEnd] >= '0' && statusLine[codeEnd] <= '9' {
		codeEnd++
	}
	if codeEnd == 0 {
		return nil
	}
	code, _ := strconv.Atoi(statusLine[:codeEnd])
	return &errorutils.HttpResponseError{
		StatusCode: code,
		Status:     statusLine,
		Body:       []byte(strings.TrimRight(bodyStr, "\n")),
	}
}

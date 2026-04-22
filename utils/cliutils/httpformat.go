package cliutils

import (
	"encoding/json"
	"net/http"
	"strings"

	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// FormatHTTPResponseJSON formats an HTTP response body as indented JSON output.
// If body is valid JSON it is printed as-is (pretty-printed).
// Otherwise a synthetic {"status_code": N, "message": "..."} is printed, where
// message is the response body text when non-empty, or the HTTP status phrase.
func FormatHTTPResponseJSON(body []byte, statusCode int) {
	if len(body) > 0 {
		var parsed interface{}
		if json.Unmarshal(body, &parsed) == nil {
			log.Output(clientUtils.IndentJson(body))
			return
		}
	}
	msg := http.StatusText(statusCode)
	if len(body) > 0 {
		msg = strings.TrimSpace(string(body))
	}
	synthetic := map[string]interface{}{
		"status_code": statusCode,
		"message":     msg,
	}
	data, _ := json.Marshal(synthetic)
	log.Output(clientUtils.IndentJson(data))
}

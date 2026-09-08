package api

import (
	"os"
	"strings"

	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// envRequireFullBundle controls whether `jf api docs search` and
// `jf api docs describe` fail on binaries that embed the dev "stub" OpenAPI
// bundle. Enabled by default; set to "false" to allow the partial catalog.
const envRequireFullBundle = "JFROG_CLI_API_DOCS_REQUIRE_FULL_BUNDLE"

func apiDocsRequireFullBundle() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(envRequireFullBundle)))
	return v != "false"
}

func maybeRequireFullApiDocsBundle() error {
	if !apiDocsRequireFullBundle() {
		return nil
	}
	return errorutils.CheckError(apispec.RequireFullBundle())
}

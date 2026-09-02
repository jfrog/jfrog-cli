package api

import (
	"os"
	"strings"

	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// envRequireFullBundle, when set to "true", makes `jf api docs search` and
// `jf api docs describe` fail on binaries that embed the dev "stub" OpenAPI
// bundle instead of returning a partial catalog.
const envRequireFullBundle = "JFROG_CLI_API_DOCS_REQUIRE_FULL_BUNDLE"

func apiDocsRequireFullBundle() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(envRequireFullBundle)), "true")
}

func maybeRequireFullApiDocsBundle() error {
	if !apiDocsRequireFullBundle() {
		return nil
	}
	return errorutils.CheckError(apispec.RequireFullBundle())
}

package api

import (
	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

func requireFullApiDocsBundle() error {
	return errorutils.CheckError(apispec.RequireFullBundle())
}

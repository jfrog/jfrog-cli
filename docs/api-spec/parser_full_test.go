//go:build full

// This test only runs when built with -tags full, i.e. against the real
// rdme-admin-sourced bundle populated into docs/api-spec/full/ at release
// time (see embed_full.go). It exists as a release gate for JGC-537: a
// spec change upstream that this package's narrow parser can't handle
// (e.g. an OpenAPI construct decoded into the wrong Go type) must fail this
// test and block the release, rather than surface for the first time as a
// runtime error from `jf api docs search`/`describe`.
//
// The release pipeline that builds the "full" binary must run
// `go test -tags full ./docs/api-spec/...` (or equivalent) after fetching
// the rdme-admin bundle and fail the build on any failure here.

package apispec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperations_Full(t *testing.T) {
	ops, err := Operations()
	require.NoError(t, err, "the embedded full OpenAPI spec bundle directory must be readable")
	assert.NotEmpty(t, ops, "the full bundle should declare at least one operation")

	// Operations() tolerates a bad spec file by skipping it (see parseDir) so
	// that most of the catalog stays searchable at runtime. A release build
	// must not tolerate that silently, though: any skipped file means part of
	// the catalog didn't make it in, which should block the release rather
	// than ship a gap that only surfaces when a user searches for the
	// missing operation.
	assert.Empty(t, Failures(), "every embedded full spec file must parse cleanly for a release build")
}

func TestInfo_Full(t *testing.T) {
	info := Info()
	assert.Equal(t, "full", info.SpecBundle)
	assert.NotEmpty(t, info.SpecVersion, "full builds should report the rdme-admin version they were fetched from")
}

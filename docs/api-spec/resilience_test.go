package apispec

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseDir_SkipsBadFileAndContinues is a regression test for JGC-537: a
// single spec file that fails to parse must not take down every other
// file's operations. It uses a synthetic in-memory filesystem rather than a
// real embedded bundle so the failure case doesn't need to live as a
// deliberately-broken fixture in the shipped stub/full bundle.
func TestParseDir_SkipsBadFileAndContinues(t *testing.T) {
	fsys := fstest.MapFS{
		"spec/good.yaml": &fstest.MapFile{Data: []byte(`
paths:
  /widgets:
    get:
      operationId: listWidgets
      responses:
        '200':
          description: OK
`)},
		// type as a mapping is invalid under any interpretation (unlike
		// flexType's string-or-sequence), so this always fails to parse
		// regardless of what other leniency this package grows over time.
		"spec/bad.yaml": &fstest.MapFile{Data: []byte(`
paths:
  /broken:
    post:
      operationId: broken
      requestBody:
        content:
          application/json:
            schema:
              type: {not: a-valid-type}
      responses:
        '200':
          description: OK
`)},
	}

	ops, failures, err := parseDir(fsys, "spec")
	require.NoError(t, err, "one bad file should not fail the whole bundle")

	require.Len(t, ops, 1, "the good file's operation should still be present")
	assert.Equal(t, "listWidgets", ops[0].OperationId)

	require.Len(t, failures, 1)
	assert.Equal(t, "bad.yaml", failures[0].File)
}

func TestParseDir_NoFailuresWhenAllFilesParse(t *testing.T) {
	fsys := fstest.MapFS{
		"spec/good.yaml": &fstest.MapFile{Data: []byte(`
paths:
  /widgets:
    get:
      operationId: listWidgets
      responses:
        '200':
          description: OK
`)},
	}

	_, failures, err := parseDir(fsys, "spec")
	require.NoError(t, err)
	assert.Empty(t, failures)
}

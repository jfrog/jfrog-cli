package schema

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func TestFileSpecSchema(t *testing.T) {
	// Load File Spec schema
	schema, err := ioutil.ReadFile("filespec-schema.json")
	assert.NoError(t, err)
	schemaLoader := gojsonschema.NewBytesLoader(schema)

	// Validate all specs in ../testdata/filespecs against the filespec-schema.json
	assert.NoError(t,
		filepath.Walk(filepath.Join("..", "testdata", "filespecs"), func(path string, info os.FileInfo, err error) error {
			assert.NoError(t, err)
			if info.IsDir() {
				return nil
			}

			specFileContent, err := ioutil.ReadFile(path)
			assert.NoError(t, err)
			documentLoader := gojsonschema.NewBytesLoader(specFileContent)
			result, err := gojsonschema.Validate(schemaLoader, documentLoader)
			assert.NoError(t, err)
			assert.True(t, result.Valid(), result.Errors())
			return nil
		}))
}

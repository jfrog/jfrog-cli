//go:build full

package apispec

import (
	"embed"
	"strings"
)

//go:embed full/*.yaml full/VERSION
var specFS embed.FS

const rootDir = "full"

// Bundle identifies which OpenAPI spec set is embedded in this binary.
const Bundle = "full"

func specVersion() string {
	data, err := specFS.ReadFile(rootDir + "/VERSION")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

//go:build !full

package apispec

import "embed"

//go:embed stub/*.yaml
var specFS embed.FS

const rootDir = "stub"

// Bundle identifies which OpenAPI spec set is embedded in this binary.
const Bundle = "stub"

func specVersion() string {
	return ""
}

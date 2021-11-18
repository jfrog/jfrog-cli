package npmci

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run npm ci."

var Usage = []string{cliutils.CliExecutableName + " rt npmci [npm ci args] [command options]"}

const Arguments string = `	npm ci args
		The npm ci args to run npm ci.`

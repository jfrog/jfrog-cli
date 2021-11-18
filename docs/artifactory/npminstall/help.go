package npminstall

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run npm install."

var Usage = []string{cliutils.CliExecutableName + " rt npmi [npm install args] [command options]"}

const Arguments string = `	npm install args
		The npm install args to run npm install. For example, --global.`

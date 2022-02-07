package npmci

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"npm ci [npm ci args] [command options]"}

func GetDescription() string {
	return `Run npm ci, using the npm repository, configured by the '` + coreutils.GetCliExecutableName() + ` npmc' command.`
}

func GetArguments() string {
	return `	npm ci args
		The npm ci args to run npm ci.`
}

package npminstall

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"npm install [npm install args] [command options]"}

func GetDescription() string {
	return `Run npm install, using the npm repository, configured by the '` + coreutils.GetCliExecutableName() + ` npmc' command.`
}

func GetArguments() string {
	return `	npm install args
		The npm install args to run npm install. For example, --global.`
}

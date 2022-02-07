package npmpublish

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"npm publish [command options]"}

func GetDescription() string {
	return `Packs and deploys the npm package to the Artifactory npm repository, configured by the '` + coreutils.GetCliExecutableName() + ` npmc' command.`
}

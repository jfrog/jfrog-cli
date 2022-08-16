package exportcmd

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"config export [server ID]"}

func GetDescription() string {
	return `Creates a server configuration token. The generated Config Token can be imported by the "` + coreutils.GetCliExecutableName() + ` config import <Config Token>" command.`
}

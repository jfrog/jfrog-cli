package exportcmd

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"config export [server ID]"}

func GetDescription() string {
	return `Creates a server configuration token. The generated token can be imported by the "` + coreutils.GetCliExecutableName() + ` config import <Server token>" command.`
}

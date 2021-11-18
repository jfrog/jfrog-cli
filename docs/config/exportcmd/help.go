package exportcmd

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description string = `Creates a server configuration token. The generated token can be imported by the "` + cliutils.CliExecutableName + ` config import <Server token>" command.`

var Usage = []string{cliutils.CliExecutableName + " config export [server ID]"}

const Arguments string = `	server ID
		The configured server ID.
		If not specified, the active server will be exported.`

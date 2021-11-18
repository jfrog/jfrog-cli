package add

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description string = `Adds a server configuration.`

var Usage = []string{cliutils.CliExecutableName + " config add",
	cliutils.CliExecutableName + " config add <server ID>"}

const Arguments string = `	server ID
		The configured server ID.`

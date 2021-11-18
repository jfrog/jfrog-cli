package edit

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description string = `Edits a server configuration.`

var Usage = []string{cliutils.CliExecutableName + " config edit <server ID>"}

const Arguments string = `	server ID
		The configured server ID.`

package use

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Set the active server by its ID."

var Usage = []string{cliutils.CliExecutableName + " config use <server ID>"}

const Arguments string = `	server ID
		The configured server ID which will be used by default.`

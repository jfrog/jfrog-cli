package curl

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Execute a cUrl command, using the configured Artifactory details."

var Usage = []string{cliutils.CliExecutableName + " rt curl [command options] <curl command>"}

const Arguments string = `	curl command
		cUrl command to run.`

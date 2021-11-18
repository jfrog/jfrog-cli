package npmcommand

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run npm command."

var Usage = []string{cliutils.CliExecutableName + " npm <go arguments> [command options]"}

const Arguments string = `	npm commands
		Arguments and options for the npm command.`

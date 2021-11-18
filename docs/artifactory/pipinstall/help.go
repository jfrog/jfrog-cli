package pipinstall

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run pip install."

var Usage = []string{cliutils.CliExecutableName + " rt pipi <pip sub-command>"}

const Arguments string = `	pip sub-command
		Arguments and options for the pip command.`

package pipinstall

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run pip install."

var Usage = []string{cliutils.CliExecutableName + " pip <pip arguments> [command options]"}

const Arguments string = `	pip sub-command
		Arguments and options for the pip command.`

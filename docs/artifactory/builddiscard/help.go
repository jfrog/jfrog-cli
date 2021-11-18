package builddiscard

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Discard builds by setting retention parameters."

var Usage = []string{cliutils.CliExecutableName + " rt bdi [command options] <build name>"}

const Arguments string = `	build name
		Build name.`

package buildscan

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "This command is used to perform Xray scan on a build."

var Usage = []string{cliutils.CliExecutableName + " rt bs [command options] <build name> <build number>"}

const Arguments string = `	build name
		Build name.

	build number
		Build number.`

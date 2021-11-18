package releasebundledistribute

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Distribute a release bundle."

var Usage = []string{cliutils.CliExecutableName + " ds rbd [command options] <release bundle name> <release bundle version>"}

const Arguments string = `	release bundle name
		Release bundle name.

	release bundle version
		Release bundle version.`

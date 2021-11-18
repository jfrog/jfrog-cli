package releasebundlesign

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Sign a release bundle."

var Usage = []string{cliutils.CliExecutableName + " ds rbs [command options] <release bundle name> <release bundle version>"}

const Arguments string = `	release bundle name
		Release bundle name.

	release bundle version
		Release bundle version.`

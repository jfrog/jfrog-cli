package buildpromote

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "This command is used to promote build in Artifactory."

var Usage = []string{cliutils.CliExecutableName + " rt bpr [command options] <build name> <build number> <target repository>"}

const Arguments string = `	build name
		Build name.

	build number
		Build number.

	target repository
		Build promotion target repository.`

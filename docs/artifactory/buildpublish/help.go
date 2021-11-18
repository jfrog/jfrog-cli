package buildpublish

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Publish build info."

var Usage = []string{cliutils.CliExecutableName + " rt bp [command options] <build name> <build number>"}

const Arguments string = `	build name
		Build name.

	build number
		Build number.`

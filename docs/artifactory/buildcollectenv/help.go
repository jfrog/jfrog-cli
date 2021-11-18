package buildcollectenv

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Collect environment variables. Environment variables can be excluded using the build-publish command."

var Usage = []string{cliutils.CliExecutableName + " rt bce <build name> <build number>"}

const Arguments string = `	build name
		Build name.

	build number
		Build number.`

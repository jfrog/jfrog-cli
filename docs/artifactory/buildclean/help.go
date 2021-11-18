package buildclean

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "This command is used to clean (remove) build info collected locally."

var Usage = []string{cliutils.CliExecutableName + " rt bc <build name> <build number>"}

const Arguments string = `	build name
		Build name.

	build number
		Build number.`

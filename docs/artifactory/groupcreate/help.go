package groupcreate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Create new users group."

var Usage = []string{cliutils.CliExecutableName + " rt gc <group name>"}

const Arguments string = `	group name
		The name of the new group.`

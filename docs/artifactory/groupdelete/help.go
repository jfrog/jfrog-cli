package groupdelete

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Delete a users group."

var Usage = []string{cliutils.CliExecutableName + " rt gdel <group name>"}

const Arguments string = `	group name
		Group name to be deleted.`

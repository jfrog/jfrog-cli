package groupaddusers

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Add a list of users to a group."

var Usage = []string{cliutils.CliExecutableName + " rt gau <group name> <users list>"}

const Arguments string = `	group name
		The name of the group.

	users list
		Specifies the usernames to add to the specified group.
		The list should be comma-separated. 
	`

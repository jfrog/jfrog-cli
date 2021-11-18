package usersdelete

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Delete users."

var Usage = []string{cliutils.CliExecutableName + " rt udel <users list>", cliutils.CliExecutableName + " rt udel --csv <users details file path>"}

const Arguments string = `	users list
		Comma-separated list of usernames to delete.`

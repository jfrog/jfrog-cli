package accesstokencreate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Creates an access token. By default an user-scoped token will be created, unless the --groups and/or --grant-admin options are specified."

var Usage = []string{cliutils.CliExecutableName + " rt atc", cliutils.CliExecutableName + " rt atc <user name>"}

const Arguments string = `	user name
		The user name for which this token is created. If not specified, the token will be created for the current user.`

package usercreate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Create new user."

var Usage = []string{cliutils.CliExecutableName + " rt user-create <username> <password> <email>"}

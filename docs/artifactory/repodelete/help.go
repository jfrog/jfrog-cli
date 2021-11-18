package repodelete

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Permanently delete repositories with all of their content from Artifactory."

var Usage = []string{cliutils.CliExecutableName + " rt rdel <repository pattern>"}

const Arguments string = `	repository pattern
		Specifies the repositories that should be removed. You can use wildcards to specify multiple repositories.`

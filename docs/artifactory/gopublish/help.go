package gopublish

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Publish go package and/or its dependencies to Artifactory"

var Usage = []string{cliutils.CliExecutableName + " rt gp [command options] <project version>"}

const Arguments string = `	project version
		Package version to be published.`

package repocreate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Create a new repository in Artifactory."

var Usage = []string{cliutils.CliExecutableName + " rt rc <template path>"}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used for the repository creation. The template can be created using the "` + cliutils.CliExecutableName + ` rt rpt" command.`

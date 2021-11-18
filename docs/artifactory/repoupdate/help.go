package repoupdate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Update an exiting repository configuration in Artifactory."

var Usage = []string{cliutils.CliExecutableName + " rt ru <template path>"}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used for the repository update. The template can be created using the "` + cliutils.CliExecutableName + ` rt rpt" command.`

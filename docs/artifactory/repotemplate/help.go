package repotemplate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Create a JSON template for repository creation or update."

var Usage = []string{cliutils.CliExecutableName + " rt rpt <template path>"}

const Arguments string = `	template path
		Specifies the local file system path for the template file.`

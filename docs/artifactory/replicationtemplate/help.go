package replicationtemplate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Create a JSON template for creation replication repository."

var Usage = []string{cliutils.CliExecutableName + " rt rplt <template path>"}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used for the replication create.`

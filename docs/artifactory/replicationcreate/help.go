package replicationcreate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Create a new replication in Artifactory."

var Usage = []string{cliutils.CliExecutableName + " rt rplc <template path>"}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used to create a replication. The template can be created using the “jfrog rt rplt” command.`

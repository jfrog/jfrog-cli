package replicationdelete

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Remove a replication repository from Artifactory."

var Usage = []string{cliutils.CliExecutableName + " rt rpldel <repository key>"}

const Arguments string = `	repository key
		The repository from which the replication will be deleted.`

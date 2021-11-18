package delete

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Delete files."

var Usage = []string{cliutils.CliExecutableName + " rt del [command options] <delete pattern>",
	cliutils.CliExecutableName + " rt del --spec=<File Spec path> [command options]"}

const Arguments string = `	delete pattern
		Specifies the source path in Artifactory, from which the artifacts should be deleted,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`

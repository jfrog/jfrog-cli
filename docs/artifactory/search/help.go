package search

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Search files."

var Usage = []string{cliutils.CliExecutableName + " rt s [command options] <search pattern>",
	cliutils.CliExecutableName + " rt s --spec=<File Spec path> [command options]"}

const Arguments string = `	search pattern
		Specifies the search path in Artifactory, in the following format: <repository name>/<repository path>.
		You can use wildcards to specify multiple artifacts.`

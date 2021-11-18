package buildaddgit

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Collect VCS details from git and add them to a build."

var Usage = []string{cliutils.CliExecutableName + " rt bag [command options] <build name> <build number> [Path To .git]"}

const Arguments string = `	build name
		Build name.

	build number
		Build number.

	path to .git
		Path to a directory containing the .git directory. If not specified, the .git directory is assumed to be in the current directory or in one of the parent directories.`

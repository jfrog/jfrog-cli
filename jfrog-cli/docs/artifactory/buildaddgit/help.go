package buildaddgit

const Description = "Capture git revision and remote url."

var Usage = []string{"jfrog rt bag [command options] <build name> <build number> [Path To .git]"}

const Arguments string =
`	build name
		Build name.

	build number
		Build number.

	path to .git
		Path to a directory containing the .git directory. If not specific, the .git directory is assumed to be in the current directory.`
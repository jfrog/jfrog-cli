package buildaddgit

var Usage = []string{"rt bag [command options] <build name> <build number> [Path To .git]"}

func GetDescription() string {
	return `Collects the Git revision and URL from the local .git directory and adds it to the build-info.`
}

func GetArguments() string {
	return `	build name
		Build name.

	build number
		Build number.

	path to .git
		Path to a directory containing the .git directory. If not specified, the .git directory is assumed to be in the current directory or in one of the parent directories.
		It can also collect the list of tracked project issues (for example, issues stored in JIRA or other bug tracking systems) and add them to the build-info. 
		The issues are collected by reading the git commit messages from the local git log.
		Each commit message is matched against a pre-configured regular expression, which retrieves the issue ID and issue summary.
		The information required for collecting the issues is retrieved from a yaml configuration file provided to the command.`
}

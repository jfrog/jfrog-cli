package gitlfsclean

var Usage = []string{"rt glc [command options] [path to .git]"}

func GetDescription() string {
	return "Clean files from a Git LFS repository. The command deletes all files from a Git LFS repository that are no longer available in a corresponding Git repository."
}

func GetArguments() string {
	return `	path to .git
		Path to a directory containing the .git directory. If not specific, the .git directory is assumed to be in the current directory.`
}

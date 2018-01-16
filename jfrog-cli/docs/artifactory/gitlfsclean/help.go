package gitlfsclean

const Description = "Clean files from a Git LFS repository. The command deletes all files from a Git LFS repository that are no longer available in a corresponding Git repository."

var Usage = []string{"jfrog rt glc [command options] [path to .git]"}

const Arguments string = `	path to .git
		Path to a directory containing the .git directory. If not specific, the .git directory is assumed to be in the current directory.`

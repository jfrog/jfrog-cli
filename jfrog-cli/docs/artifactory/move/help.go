package move

const Description = "Move files."

var Usage = []string{"jfrog rt mv [command options] <source pattern> <target pattern>",
	"jfrog rt mv --spec=<File Spec path> [command options]"}

const Arguments string =
`	source pattern
		Specifies the source path in Artifactory, from which the artifacts should be moved,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.

	target pattern
		Specifies the target path in Artifactory, to which the artifacts should be moved, in the following format: <repository name>/<repository path>.
		If the pattern ends with a slash, the target path is assumed to be a folder. For example, if you specify the target as "repo-name/a/b/",
		then "b" is assumed to be a folder in Artifactory into which files should be moved.
		If there is no terminal slash, the target path is assumed to be a file to which the moved file should be renamed.
		For example, if you specify the target as "repo-name/a/b", the moved file is renamed to "b" in Artifactory.
		For flexibility in specifying the upload path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding
		tokens in the source path that are enclosed in parenthesis.`
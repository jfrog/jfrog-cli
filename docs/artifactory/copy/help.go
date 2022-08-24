package copy

var Usage = []string{"rt cp [command options] <source pattern> <target pattern>",
	"rt cp --spec=<File Spec path> [command options]"}

const EnvVar string = `	JFROG_CLI_FAIL_NO_OP
	[Default: false]
	Set to true if you'd like the command to return exit code 2 in case of no files are affected.`

func GetDescription() string {
	return "Copy files."
}

func GetArguments() string {
	return `	source Pattern
		Specifies the source path in Artifactory, from which the artifacts should be copied,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.

	target Pattern
		Specifies the target path in Artifactory, to which the artifacts should be copied, in the following format: <repository name>/<repository path>.
		If the pattern ends with a slash, the target path is assumed to be a folder. For example, if you specify the target as "repo-name/a/b/",
		then "b" is assumed to be a folder in Artifactory into which files should be copied.
		If there is no terminal slash, the target path is assumed to be a file to which the copied file should be renamed.
		For example, if you specify the target as "repo-name/a/b", the copied file is renamed to "b" in Artifactory.
		For flexibility in specifying the upload path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding
		tokens in the source path that are enclosed in parenthesis.`
}

package download

var Usage = []string{"rt dl [command options] <source pattern> [target pattern]",
	"rt dl --spec=<File Spec path> [command options]"}

const EnvVar string = `	JFROG_CLI_TRANSITIVE_DOWNLOAD_EXPERIMENTAL
		[Default: false]
		Set to true to look for artifacts also in remote repositories. The search will run on the first five remote repositories within the virtual repository. This feature is experimental and available on Artifactory version 7.17.0 or higher.

	JFROG_CLI_FAIL_NO_OP
		[Default: false]
		Set to true if you'd like the command to return exit code 2 in case of no files are affected.`

func GetDescription() string {
	return "Download files."
}

func GetArguments() string {
	return `	source pattern
		Specifies the source path in Artifactory, from which the artifacts should be downloaded,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.

	target pattern
		The second argument is optional and specifies the local file system target path.
		If the target path ends with a slash, the path is assumed to be a directory.
		For example, if you specify the target as "repo-name/a/b/", then "b" is assumed to be a directory into which files should be downloaded.
		If there is no terminal slash, the target path is assumed to be a file to which the downloaded file should be renamed.
		For example, if you specify the target as "a/b", the downloaded file is renamed to "b".
		For flexibility in specifying the target path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding
		tokens in the source path that are enclosed in parenthesis.`
}

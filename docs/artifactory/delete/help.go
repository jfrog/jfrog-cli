package delete

var Usage = []string{"rt del [command options] <delete pattern>",
	"rt del --spec=<File Spec path> [command options]"}

const EnvVar string = `	JFROG_CLI_FAIL_NO_OP
	[Default: false]
	Set to true if you'd like the command to return exit code 2 in case of no files are affected.`
	
func GetDescription() string {
	return "Delete files."
}

func GetArguments() string {
	return `	delete pattern
		Specifies the source path in Artifactory, from which the artifacts should be deleted,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`
}

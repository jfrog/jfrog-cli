package search

var Usage = []string{"rt s [command options] <search pattern>",
	"rt s --spec=<File Spec path> [command options]"}

const EnvVar string = `	JFROG_CLI_FAIL_NO_OP
	[Default: false]
	Set to true if you'd like the command to return exit code 2 in case of no files are affected.`
	
func GetDescription() string {
	return "Search files."
}

func GetArguments() string {
	return `	search pattern
		Specifies the search path in Artifactory, in the following format: <repository name>/<repository path>.
		You can use wildcards to specify multiple artifacts.`
}

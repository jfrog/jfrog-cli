package setprops

var Usage = []string{"rt sp [command options] <artifacts pattern> <artifact properties>",
	"rt sp <artifact properties> --spec=<File Spec path> [command options]"}

const EnvVar string = `	JFROG_CLI_FAIL_NO_OP
	[Default: false]
	Set to true if you'd like the command to return exit code 2 in case of no files are affected.`

func GetDescription() string {
	return "Set properties on existing files in Artifactory."
}

func GetArguments() string {
	return `	artifacts pattern
		Artifacts that match the pattern will be set with the specified properties.

	artifact properties
		The list of properties, in the form of key1=value1;key2=value2,..., to be set on the matching artifacts.`
}

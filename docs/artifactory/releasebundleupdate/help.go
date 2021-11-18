package releasebundleupdate

var Usage = []string{"ds rbu [command options] <release bundle name> <release bundle version> <pattern>",
	"ds rbu --spec=<File Spec path> [command options] <release bundle name> <release bundle version>"}

func GetDescription() string {
	return "Updates an existing unsigned release bundle version."
}

func GetArguments() string {
	return `	release bundle name
		The name of the release bundle.

	release bundle version
		The release bundle version.

	pattern
		Specifies the source path in Artifactory, from which the artifacts should be bundled,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`
}

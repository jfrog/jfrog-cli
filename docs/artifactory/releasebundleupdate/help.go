package releasebundleupdate

const Description = "Updates an existing unsigned release bundle version."

var Usage = []string{"jfrog rt rbu [command options] <bundle name> <bundle version> <pattern>",
	"jfrog rt rbu --spec=<File Spec path> [command options] <bundle name> <bundle version>"}

const Arguments string = `	bundle name
		Build name.

	bundle version
		Bundle version.

	pattern
		Specifies the source path in Artifactory, from which the artifacts should be bundled,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`

package releasebundlecreate

const Description = "Create a release bundle."

var Usage = []string{"jfrog rt rbc [command options] <release bundle name> <release bundle version> <pattern>",
	"jfrog rt rbc --spec=<File Spec path> [command options] <release bundle name> <release bundle version>"}

const Arguments string = `	release bundle name
		The name of the release bundle.

	release bundle version
		The release bundle version.

	pattern
		Specifies the source path in Artifactory, from which the artifacts should be bundled,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`

package releasebundlecreate

const Description = "Create a release bundle."

var Usage = []string{"jfrog rt rbc [command options] <bundle name> <bundle version> <pattern>",
	"jfrog rt rbc --spec=<File Spec path> [command options] <bundle name> <bundle version>"}

const Arguments string = `	bundle name
		Build name.

	bundle version
		Bundle version.

	pattern
		Specifies the source path in Artifactory, from which the artifacts should be bundled,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`

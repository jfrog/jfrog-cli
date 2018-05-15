package buildaddartifact

const Description = "This command is used to add a priorly uploaded artifact to a build."

var Usage = []string{"jfrog rt baa [command options] <build name> <build number> <artifact pattern>",
	"jfrog rt baa --spec=<File Spec path> [command options] <build_name> <build_number>"}


const Arguments string =
`	build name
		Build name.

	build number
		Build number.

	artifact pattern
		Specifies the search path in Artifactory, in the following format: <repository name>/<repository path>.
		You can use wildcards to specify multiple artifacts.`
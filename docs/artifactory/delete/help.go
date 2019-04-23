package delete

const Description = "Delete files."

var Usage = []string{"jfrog rt del [command options] <delete pattern>",
	"jfrog rt del --spec=<File Spec path> [command options]"}

const Arguments string = `	delete pattern
		Specifies the source path in Artifactory, from which the artifacts should be deleted,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`

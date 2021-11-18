package delete

var Usage = []string{"rt del [command options] <delete pattern>",
	"rt del --spec=<File Spec path> [command options]"}

func GetDescription() string {
	return "Delete files."
}

func GetArguments() string {
	return `	delete pattern
		Specifies the source path in Artifactory, from which the artifacts should be deleted,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`
}

package importbundle

var Usage = []string{"rbi [command options] <path to archive>"}

func GetDescription() string {
	return "Import a local release bundle archive to Artifactory"
}

func GetArguments() string {
	return `	path to archive
		Path to the release bundle archive on the filesystem
`
}

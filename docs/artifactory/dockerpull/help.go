package dockerpull

var Usage = []string{"rt docker-pull <image tag> <source repo>"}

func GetDescription() string {
	return "Docker pull."
}

func GetArguments() string {
	return `	image tag
		Docker image tag to pull.
	source repo
		Source repository in Artifactory.
`
}

package dockerpush

var Usage = []string{"rt docker-push <image tag> <target repo>"}

func GetDescription() string {
	return "Docker push."
}

func GetArguments() string {
	return `	image tag
		Docker image tag to push.
	target repo
		Target repository in Artifactory.
`
}

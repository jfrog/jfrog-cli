package podmanpush

var Usage = []string{"rt podman-push <image tag> <target repo>"}

func GetDescription() string {
	return "Podman push."
}

func GetArguments() string {
	return `	image tag
		Docker image tag to push.
	target repo
		Target repository in Artifactory.
`
}

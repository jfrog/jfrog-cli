package podmanpull

var Usage = []string{"rt podman-pull <image tag> <target repo>"}

func GetDescription() string {
	return "Podman pull."
}

func GetArguments() string {
	return `	image tag
		Docker image tag to pull.
	target repo
		Source repository in Artifactory.
`
}

package podmanpull

var Usage = []string{"rt podman-pull <image tag> <source repo>"}

func GetDescription() string {
	return "Podman pull."
}

func GetArguments() string {
	return `	image tag
		Docker image tag to pull.
	source repo
		Source repository in Artifactory.
`
}

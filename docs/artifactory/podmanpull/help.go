package podmanpull

const Description = "Podman pull."

var Usage = []string{"jfrog rt podman-pull <image tag> <target repo>"}

const Arguments string = `	image tag
		Docker image tag to pull.
	target repo
		Source repository in Artifactory.
`

package podmanpull

const Description = "Podman pull."

var Usage = []string{"jfrog rt podman-pull <image tag> <target repo>"}

const Arguments string = `	image tag
		Podman image tag to pull.
	target repo
		Target repository in Artifactory.
`

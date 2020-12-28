package podmanpush

const Description = "Podman push."

var Usage = []string{"jfrog rt podman-push <image tag> <target repo>"}

const Arguments string = `	image tag
		Docker image tag to push.
	target repo
		Target repository in Artifactory.
`

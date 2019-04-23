package dockerpush

const Description = "Docker push."

var Usage = []string{"jfrog rt docker-push <image tag> <target repo>"}

const Arguments string = `	image tag
		Docker image tag to push.
	target repo
		Target repository in Artifactory.
`

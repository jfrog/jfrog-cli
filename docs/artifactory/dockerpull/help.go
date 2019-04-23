package dockerpull

const Description = "Docker pull."

var Usage = []string{"jfrog rt docker-pull <image tag> <source repo>"}

const Arguments string = `	image tag
		Docker image tag to pull.
	source repo
		Source repository in Artifactory.
`

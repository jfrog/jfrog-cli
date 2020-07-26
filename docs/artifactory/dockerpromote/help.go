package dockerpromote

const Description = "Promotes a Docker image from one repository to another. Supported by local repositories only."

var Usage = []string{"jfrog rt docker-promote <source repo> <target repo> <source docker image>"}

const Arguments string = `	source repo
		Source repository in Artifactory.
	target repo
		Target repository in Artifactory.
	source docker image
		The docker image name to promote.
`

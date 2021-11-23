package dockerpromote

var Usage = []string{"rt docker-promote <source docker image> <source repo> <target repo>"}

func GetDescription() string {
	return "Promotes a Docker image from one repository to another. Supported by local repositories only."
}

func GetArguments() string {
	return `	source docker image
		The docker image name to promote.
	source repo
		Source repository in Artifactory.
	target repo
		Target repository in Artifactory.
`
}

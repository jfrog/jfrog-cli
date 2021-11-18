package dockerpush

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Docker push."

var Usage = []string{cliutils.CliExecutableName + " rt docker-push <image tag> <target repo>"}

const Arguments string = `	image tag
		Docker image tag to push.
	target repo
		Target repository in Artifactory.
`

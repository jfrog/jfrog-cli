package podmanpull

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Podman pull."

var Usage = []string{cliutils.CliExecutableName + " rt podman-pull <image tag> <target repo>"}

const Arguments string = `	image tag
		Docker image tag to pull.
	target repo
		Source repository in Artifactory.
`

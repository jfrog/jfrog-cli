package dockerpull

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Docker pull."

var Usage = []string{cliutils.CliExecutableName + " rt docker-pull <image tag> <source repo>"}

const Arguments string = `	image tag
		Docker image tag to pull.
	source repo
		Source repository in Artifactory.
`

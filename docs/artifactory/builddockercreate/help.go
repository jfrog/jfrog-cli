package builddockercreate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Add a published docker image to the build-info."

var Usage = []string{cliutils.CliExecutableName + " rt build-docker-create <target repo> --image-file=<Image file path>"}

const Arguments string = `	target repo
		The repository to which the image was pushed.
`

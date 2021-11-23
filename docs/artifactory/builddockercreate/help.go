package builddockercreate

var Usage = []string{"rt build-docker-create <target repo> --image-file=<Image file path>"}

func GetDescription() string {
	return "Add a published docker image to the build-info."
}

func GetArguments() string {
	return `	target repo
		The repository to which the image was pushed.
`
}

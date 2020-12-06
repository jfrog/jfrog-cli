package builddockercreate

const Description = "Build Docker Create."

var Usage = []string{"jfrog rt build-docker-create <target repo> <--image-name-with-digest-file=*file-path*>"}

const Arguments string = `	image-name-with-digest-file
		File path to the image name and manifest's digest in Artifactory.
	target repo
		The repository to which the image was pushed.
`

package builddockercreate

const Description = "Build Docker Create."

var Usage = []string{"jfrog rt build-docker-create <target repo> <--image-file=*file-path*>"}

const Arguments string = `	image-file
		Path to a file which includes one line in the following format: IMAGE-TAG@sha256:MANIFEST-SHA256.
	target repo
		The repository to which the image was pushed.
`

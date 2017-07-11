package buildclean

const Description = "This command is used to clean (remove) build info collected locally."

var Usage = []string{"jfrog rt bc <build name> <build number>"}

const Arguments string =
`	build name
		Build name.

	build number
		Build number.`
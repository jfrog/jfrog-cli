package buildappend

var Usage = []string{"rt ba <build name> <build number> <build name to append> <build number to append>"}

func GetDescription() string {
	return "Append published build to the build info."
}

func GetArguments() string {
	return `	build name
		The current build name.

	build number
		The current build number.

	build name to append
		The published build name to append to the current build.
		
	build number to append
		The published build number to append to the current build.`
}

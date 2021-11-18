package buildclean

var Usage = []string{"rt bc <build name> <build number>"}

func GetDescription() string {
	return "This command is used to clean (remove) build info collected locally."
}

func GetArguments() string {
	return `	build name
		Build name.

	build number
		Build number.`
}

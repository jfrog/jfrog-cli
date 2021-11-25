package buildcollectenv

var Usage = []string{"rt bce <build name> <build number>"}

func GetDescription() string {
	return "Collect environment variables. Environment variables can be excluded using the build-publish command."
}

func GetArguments() string {
	return `	build name
		Build name.

	build number
		Build number.`
}

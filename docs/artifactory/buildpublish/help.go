package buildpublish

var Usage = []string{"rt bp [command options] <build name> <build number>"}

func GetDescription() string {
	return "Publish build info."
}

func GetArguments() string {
	return `	build name
		Build name.

	build number
		Build number.`
}

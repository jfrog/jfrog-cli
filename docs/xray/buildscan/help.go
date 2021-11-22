package buildscan

var Usage = []string{"bs [command options] <build name> <build number>"}

func GetDescription() string {
	return "This command is used to perform Xray scan on a build."
}

func GetArguments() string {
	return `	build name
		Build name.

	build number
		Build number.`
}

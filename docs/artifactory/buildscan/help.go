package buildscan

var Usage = []string{"rt bs [command options] <build name> <build number>"}

func GetDescription() string {
	return "Scan a published build-info with Xray."
}

func GetArguments() string {
	return `	build name
		Build name.

	build number
		Build number.`
}

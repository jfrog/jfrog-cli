package releasebundledistribute

var Usage = []string{"ds rbd [command options] <release bundle name> <release bundle version>"}

func GetDescription() string {
	return "Distribute a release bundle."
}

func GetArguments() string {
	return `	release bundle name
		Release bundle name.

	release bundle version
		Release bundle version.`
}

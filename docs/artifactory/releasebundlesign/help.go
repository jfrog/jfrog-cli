package releasebundlesign

var Usage = []string{"ds rbs [command options] <release bundle name> <release bundle version>"}

func GetDescription() string {
	return "Sign a release bundle."
}

func GetArguments() string {
	return `	release bundle name
		Release bundle name.

	release bundle version
		Release bundle version.`
}

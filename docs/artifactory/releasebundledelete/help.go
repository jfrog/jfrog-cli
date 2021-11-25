package releasebundledelete

var Usage = []string{"ds rbdel [command options] <release bundle name> <release bundle version>"}

func GetDescription() string {
	return "Delete a release bundle."
}

func GetArguments() string {
	return `	release bundle name
		Release bundle name.

	release bundle version
		Release bundle version.`
}

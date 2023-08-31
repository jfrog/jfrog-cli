package distribute

var Usage = []string{"rbd [command options] <release bundle name> <release bundle version>"}

func GetDescription() string {
	return "Distribute a release bundle."
}

func GetArguments() string {
	return `	release bundle name
		Name of the Release Bundle to distribute.

	release bundle version
		Version of the Release Bundle to distribute.`
}

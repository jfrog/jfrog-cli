package create

var Usage = []string{"rbc [command options] <release bundle name> <release bundle version>"}

func GetDescription() string {
	return "Create a release bundle from builds or from existing release bundles"
}

func GetArguments() string {
	return `	release bundle name
		Name of the newly created Release Bundle.

	release bundle version
		Version of the newly created Release Bundle.`
}

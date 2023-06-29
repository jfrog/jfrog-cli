package create

var Usage = []string{"rbc [command options] <release bundle name> <release bundle version> <signing key name>"}

func GetDescription() string {
	return "Create a release bundle from builds or release bundles"
}

func GetArguments() string {
	return `	release bundle name
		Name of the newly created Release Bundle.

	release bundle version
		Version of the newly created Release Bundle.

	signing key name
		The GPG/RSA key-pair name given in Artifactory.`
}

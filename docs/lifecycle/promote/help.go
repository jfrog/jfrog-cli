package promote

var Usage = []string{"rbp [command options] <release bundle name> <release bundle version> <signing key name>"}

func GetDescription() string {
	return "Promote a release bundle"
}

func GetArguments() string {
	return `	release bundle name
		Name of the Release Bundle to promote.

	release bundle version
		Version of the Release Bundle to promote.

	signing key name
		The GPG/RSA key-pair name given in Artifactory.`
}

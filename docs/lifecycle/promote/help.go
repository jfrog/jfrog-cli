package promote

var Usage = []string{"rbp [command options] <release bundle name> <release bundle version> <environment>"}

func GetDescription() string {
	return "Promote a release bundle"
}

func GetArguments() string {
	return `	release bundle name
		Name of the Release Bundle to promote.

	release bundle version
		Version of the Release Bundle to promote.

	environment
		Name of the target environment for the promotion.`
}

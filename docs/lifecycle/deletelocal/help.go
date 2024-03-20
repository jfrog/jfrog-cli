package deletelocal

var Usage = []string{"rbdell [command options] <release bundle name> <release bundle version>",
	"rbdell [command options] <release bundle name> <release bundle version> <environment>"}

func GetDescription() string {
	return "Delete all release bundle promotions to an environment or delete a release bundle locally altogether."
}

func GetArguments() string {
	return `	release bundle name
		Name of the Release Bundle to delete locally.

	release bundle version
		Version of the Release Bundle to delete locally.

	environment
		If provided, all promotions to this environment are deleted. 
		Otherwise, the release bundle is deleted locally with all its promotions.`
}

package deleteremote

var Usage = []string{"rbdelr [command options] <release bundle name> <release bundle version>"}

func GetDescription() string {
	return "Delete a release bundle remotely."
}

func GetArguments() string {
	return `	release bundle name
		Name of the Release Bundle to delete remotely.

	release bundle version
		Version of the Release Bundle to delete remotely.`
}

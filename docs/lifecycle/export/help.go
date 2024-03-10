package export

var Usage = []string{"rbe [command options] <release bundle name> <release bundle version>"}

func GetDescription() string {
	return "Export and download a Release Bundle"
}

func GetArguments() string {
	return `	release bundle name
		Name of the Release Bundle to export.

	release bundle version
		Version of the Release Bundle to export.`
}

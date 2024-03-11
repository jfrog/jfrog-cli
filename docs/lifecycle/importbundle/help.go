package importbundle

var Usage = []string{"rbi [command options] <path to bundle>"}

func GetDescription() string {
	return "Import a Release Bundle archive"
}

func GetArguments() string {
	return `	path to bundle
		Path for the desired imported bundle
`
}

package builddiscard

var Usage = []string{"rt bdi [command options] <build name>"}

func GetDescription() string {
	return "Discard builds by setting retention parameters."
}

func GetArguments() string {
	return `	build name
		Build name.`
}

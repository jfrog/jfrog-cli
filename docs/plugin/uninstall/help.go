package uninstall

var Usage = []string{"plugin uninstall <plugin name>"}

func GetDescription() string {
	return "Uninstall a JFrog CLI plugin."
}

func GetArguments() string {
	return `	plugin name
		Specifies the name of the JFrog CLI Plugin you wish to uninstall from the local plugins pool.`
}

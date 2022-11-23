package installtransfer

var Usage = []string{"rt transfer-install <server-id> [command options]"}

func GetDescription() string {
	return "Download and install the data-transfer user plugin on the primary node of Artifactory, which is running on this local machine."
}

func GetArguments() string {
	return `	server-id
		The ID of the source server, on which the plugin should be installed.`
}

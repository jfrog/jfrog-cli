package installtransfer

var Usage = []string{"rt transfer-install <server-id> [command options]"}

func GetDescription() string {
	return "Download and install the data-transfer user plugin on the primary node of this server, This command must be run on the source Artifactory machine."
}

func GetArguments() string {
	return `	server-id
		The source server ID that the plugin will be installed on.`
}

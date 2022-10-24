package installtransfer

var Usage = []string{"rt transfer-install <server-id> [command options]"}

func GetDescription() string {
	return "Download and install the data-transfer user plugin."
}

func GetArguments() string {
	return `	server-id
		The source server ID. The plugin will be install on the primary node of this server.`
}

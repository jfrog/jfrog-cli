package configtransfer

var Usage = []string{"rt transfer-config [command options] <source-server-id> <target-server-id>"}

func GetDescription() string {
	return "Copy full Artifactory configuration from source Artifactory server to target Artifactory server. Warning - This action will wipe all Artifactory content in this target server."
}

func GetArguments() string {
	return `	source-server-id
		The source server ID. The configuration will be exported from this server.

	target-server-id
		The target server ID. The configuration will be imported to this server.`
}

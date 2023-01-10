package transferconfigmerge

var Usage = []string{"rt transfer-config-merge [command options] <source-server-id> <target-server-id>"}

func GetDescription() string {
	return "Merge repositories and projects from source Artifactory server to target Artifactory server and print merge conflicts CSV"
}

func GetArguments() string {
	return `	source-server-id
		The source server ID. The configuration will be exported from this server.

	target-server-id
		The target server ID. The configuration will be imported to this server.`
}

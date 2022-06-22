package transferdata

var Usage = []string{"rt transfer-data [command options] <source-server-id> <target-server-id>"}

func GetDescription() string {
	return "Transfer data from one Artifactory to another."
}

func GetArguments() string {
	return `	source-server-id
		Server ID of the Artifactory instance to transfer from.

	target-server-id
		Server ID of the Artifactory instance to transfer to.`
}

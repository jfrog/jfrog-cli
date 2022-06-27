package transferfiles

var Usage = []string{"rt transfer-files [command options] <source-server-id> <target-server-id>"}

func GetDescription() string {
	return "Transfer files from one Artifactory to another."
}

func GetArguments() string {
	return `	source-server-id
		Server ID of the Artifactory instance to transfer from.

	target-server-id
		Server ID of the Artifactory instance to transfer to.`
}

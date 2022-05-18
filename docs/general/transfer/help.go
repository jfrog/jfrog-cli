package transfer

var Usage = []string{"transfer <source-server-id> <target-server-id> <repository-name>"}

func GetDescription() string {
	return "Transfer from one Artifactory to another."
}

func GetArguments() string {
	return `	source-server-id
		Server ID of the Artifactory instance to transfer from.

	target-server-id
		Server ID of the Artifactory instance to transfer to.

	repository-name
	Repository to transfer.`
}

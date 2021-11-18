package replicationdelete

var Usage = []string{"rt rpldel <repository key>"}

func GetDescription() string {
	return "Remove a replication repository from Artifactory."
}

func GetArguments() string {
	return `	repository key
		The repository from which the replication will be deleted.`
}

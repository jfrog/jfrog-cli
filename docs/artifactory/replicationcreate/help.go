package replicationcreate

var Usage = []string{"rt rplc <template path>"}

func GetDescription() string {
	return "Create a new replication in Artifactory."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file to be used to create a replication. The template can be created using the “jfrog rt rplt” command.`
}

package replicationtemplate

var Usage = []string{"rt rplt <template path>"}

func GetDescription() string {
	return "Create a JSON template for creation replication repository."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file to be used for the replication create.`
}

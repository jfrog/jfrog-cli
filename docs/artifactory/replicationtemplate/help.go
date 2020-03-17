package replicationtemplate

const Description = "Create a JSON template for creation replication repository."

var Usage = []string{`jfrog rt rplt <template path>`}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used for the replication create. The template can be created using the “jfrog rt rplt” command.`

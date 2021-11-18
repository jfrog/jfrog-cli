package repotemplate

var Usage = []string{"rt rpt <template path>"}

func GetDescription() string {
	return "Create a JSON template for repository creation or update."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file.`
}

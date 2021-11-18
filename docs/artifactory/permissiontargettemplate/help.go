package permissiontargettemplate

var Usage = []string{"rt ptt <template path>"}

func GetDescription() string {
	return "Create a JSON template for a permission target creation or replacement."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file.`
}

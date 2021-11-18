package groupdelete

var Usage = []string{"rt gdel <group name>"}

func GetDescription() string {
	return "Delete a users group."
}

func GetArguments() string {
	return `	group name
		Group name to be deleted.`
}

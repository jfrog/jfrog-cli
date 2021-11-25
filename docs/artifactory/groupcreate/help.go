package groupcreate

var Usage = []string{"rt gc <group name>"}

func GetDescription() string {
	return "Create new users group."
}

func GetArguments() string {
	return `	group name
		The name of the new group.`
}

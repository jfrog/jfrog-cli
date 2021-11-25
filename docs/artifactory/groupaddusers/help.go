package groupaddusers

var Usage = []string{"rt gau <group name> <users list>"}

func GetDescription() string {
	return "Add a list of users to a group."
}

func GetArguments() string {
	return `	group name
		The name of the group.

	users list
		Specifies the usernames to add to the specified group.
		The list should be comma-separated. 
	`
}

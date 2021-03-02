package groupaddusers

const Description = "Add a list of users to a group."

var Usage = []string{`jfrog rt gau <group name> <users list>`}

const Arguments string = `	group name
		The name of the group.

	users list
		Specifies the usernames to add to the specified group.
		The list should be comma-separated. 
	`

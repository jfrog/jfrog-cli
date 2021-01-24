package groupaddusers

const Description = "Add a list of users to a group."

var Usage = []string{`jfrog rt gau <Group Name> <Users list>`}

const Arguments string = `	group name
		The name of the new group.


	users list
	Specifies the usernames to add to the specified group.
	The list should be comma-seperated. 
	`

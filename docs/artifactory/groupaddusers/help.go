package groupaddusers

const Description = "Add users to a group in the JFrog Platform."

var Usage = []string{`jfrog rt gau <Group Name> <Users list>`}

const Arguments string = `	group name
		Specifies the desired name of the new group.

	users list
		Specifies the users names to add to the specified group.
		The list should be consist of user's names separated by a comma(',') 
		for example, if we desire to add Alice and Bob to the group "secrete-keepers"
		we should use the following command: "jfrog rt gau secrete-keepers Alice,Bob" 
	`

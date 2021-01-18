package usersdelete

const Description = "Deletes users from the JFrog Platform."

var Usage = []string{`jfrog rt udel <Users List>`, `jfrog rt udel --csv <Users Details File Path>`}

const Arguments string = `
users list
	Specifies the users names to be deleted
	The list should be consist of user's names separated by a comma(',').`

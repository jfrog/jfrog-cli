package usersdelete

const Description = "Deletes users."

var Usage = []string{`jfrog rt udel <Users List>`, `jfrog rt udel --csv <Users Details File Path>`}

const Arguments string = `
users list
	Comma-separated list of usernames to delete.`

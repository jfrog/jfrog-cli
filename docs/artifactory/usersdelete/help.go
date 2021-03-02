package usersdelete

const Description = "Delete users."

var Usage = []string{`jfrog rt udel <users list>`, `jfrog rt udel --csv <users details file path>`}

const Arguments string = `	users list
		Comma-separated list of usernames to delete.`

package usersdelete

var Usage = []string{"rt udel <users list>", "rt udel --csv <users details file path>"}

func GetDescription() string {
	return "Delete users."
}

func GetArguments() string {
	return `	users list
		Comma-separated list of usernames to delete.`
}

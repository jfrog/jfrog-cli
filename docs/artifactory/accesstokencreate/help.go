package accesstokencreate

var Usage = []string{"rt atc", "rt atc <username>"}

func GetDescription() string {
	return "Creates an Artifactory access token. By default an user-scoped token will be created, unless the --groups and/or --grant-admin options are specified."
}

func GetArguments() string {
	return `	username
		The username for which this token is created. If not specified, the token will be created for the current user.`
}

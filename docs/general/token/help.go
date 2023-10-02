package token

var Usage = []string{"atc", "atc <username>"}

func GetDescription() string {
	return `Creates an access token.
		By default, an user-scoped token will be created. 
		Administrator may provide the scope explicitly with '--scope', or implicitly with '--groups', '--grant-admin'.`
}

func GetArguments() string {
	return `	username
		The username for which this token is created. If not specified, the token will be created for the current user.`
}

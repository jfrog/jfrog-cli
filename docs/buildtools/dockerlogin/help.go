package dockerlogin

var Usage = []string{"docker login [command options]"}

func GetDescription() string {
	return `Login to a artifactory Docker registry`
}

func GetArguments() string {
	return `	Command accepts optional registry for login. If not provided, the registry from the jfrog config will be used.
	This argument is mandatory when logging in using username and password.`
}

package dockerlogin

var Usage = []string{"docker login [command options]"}

func GetDescription() string {
	return `Login to a artifactory Docker registry`
}

func GetArguments() string {
	return `	command accepts no arguments.`
}

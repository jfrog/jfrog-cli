package dockerpull

var Usage = []string{"docker pull <image tag> [command options]"}

func GetDescription() string {
	return `Run Docker pull command.`
}

func GetArguments() string {
	return `	docker pull args
		The docker pull args to run docker pull.`
}

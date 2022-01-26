package docker

var Usage = []string{"docker <image tag> [command options]"}

func GetDescription() string {
	return `Run Docker command.
		Tip: Use 'docker scan <image tag>' to run vulnerabilities scan on you local docker image.`
}

func GetArguments() string {
	return `	docker commands
		Arguments and options for the npm command.`
}

package dockerscan

var Usage = []string{"docker scan <image tag> [command options]"}

func GetDescription() string {
	return `Run Docker scan command.`
}

func GetArguments() string {
	return `	docker scan args
		The docker scan args to run docker scan.`
}
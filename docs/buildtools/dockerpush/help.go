package dockerpush

var Usage = []string{"docker push <image tag> [command options]"}

func GetDescription() string {
	return `Run Docker push command.`
}

func GetArguments() string {
	return `	docker push args
		The docker push args to run docker push.`
}

package docker

var Usage = []string{"docker <docker arguments> [command options]"}

func GetDescription() string {
	return `Run Docker command.
		Tip: Use 'docker scan <image tag>' to scan a local docker container for security vulnerabilities with JFrog Xray.`
}

func GetArguments() string {
	return `	push                        Run docker push.
	pull                        Run docker pull.
	scan                        Scan a local docker container for security vulnerabilities with JFrog Xray.`
}

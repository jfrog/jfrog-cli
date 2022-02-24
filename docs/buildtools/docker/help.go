package docker

var Usage = []string{"docker <docker arguments> [command options]"}

func GetDescription() string {
	return `Run any docker command, including ‘jf docker scan’ for scanning a local image with Xray.`
}

func GetArguments() string {
	return `	push                        Run docker push.
	pull                        Run docker pull.
	scan                        Scan a local docker image for security vulnerabilities with JFrog Xray.`
}

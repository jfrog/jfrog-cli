package dockerscan

var Usage = []string{"docker scan <image tag>"}

func GetDescription() string {
	return "Scan local docker image using the docker client and Xray."
}

func GetArguments() string {
	return `	docker scan args
		The docker scan args to run docker scan.`
}

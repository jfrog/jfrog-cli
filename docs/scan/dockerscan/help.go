package dockerscan

var Usage = []string{"docker scan <image tag>"}

func GetDescription() string {
	return "Scan local docker image using the docker client and Xray."
}

package dockerscan

var Usage = []string{"docker-scan <image tag>"}

func GetDescription() string {
	return "Scan docker image located on file system with Xray."
}

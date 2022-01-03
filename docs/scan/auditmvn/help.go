package auditmvn

var Usage = []string{"audit-mvn [command options]"}

func GetDescription() string {
	return "Execute an audit Maven command, using the configured Xray details."
}
